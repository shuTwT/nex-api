package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/shuTwT/nex-api/internal/handler/cron"
	"github.com/shuTwT/nex-api/internal/handler/oauth"
	"github.com/shuTwT/nex-api/internal/handler/router"
	"github.com/shuTwT/nex-api/internal/infra/config"
	"github.com/shuTwT/nex-api/internal/infra/logger"
	"github.com/shuTwT/nex-api/internal/infra/schedule"
	"github.com/shuTwT/nex-api/internal/infra/worker"
	servicegateway "github.com/shuTwT/nex-api/internal/service/gateway"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "", "path to a config file (optional; environment variables take precedence per key)")
	flag.Parse()

	options := []config.Option{config.WithDotEnvFile(defaultDotEnvFile())}
	if configFile != "" {
		options = append(options, config.WithConfigFile(configFile))
	}
	cfg, err := config.Load(options...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nex-api: load configuration: %v\n", err)
		os.Exit(1)
	}
	loggerInstance, err := logger.NewLogger(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nex-api: create logger: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := run(ctx, cfg, loggerInstance); err != nil {
		loggerInstance.Error("server exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}

// defaultDotEnvFile selects only the dotenv file in the process working
// directory; it never searches parent or child directories.
func defaultDotEnvFile() string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return ".env"
	}
	return filepath.Join(workingDirectory, ".env")
}

func run(ctx context.Context, cfg config.Config, loggerInstance *slog.Logger) error {
	if ctx == nil {
		return errors.New("server context is nil")
	}

	// The run context owns every long-lived resource, including script-worker
	// subprocesses.  This makes caller cancellation and SIGINT/SIGTERM follow
	// the same graceful shutdown path.
	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	deps, scheduleManager, err := buildServices(runCtx, cfg, loggerInstance)
	if err != nil {
		return err
	}
	defer func() {
		shutdownServerResources(runCtx, cfg.Server.ShutdownTimeout, deps, scheduleManager, loggerInstance)
	}()

	// Redis 未就绪时只告警、不阻止启动;readyz 会持续暴露该状态。
	if err := deps.Redis.Ping(runCtx).Err(); err != nil {
		loggerInstance.Warn("redis is unreachable; usage statistics and rate limiting will be unavailable", slog.Any("err", err))
	}

	pool, err := newWorkerPool(runCtx, loggerInstance)
	if err != nil {
		loggerInstance.Warn("script worker unavailable; API script transforms will be disabled", slog.Any("err", err))
	} else {
		loggerInstance.Info("script worker pool started", slog.Int("workers", 2))
	}
	deps.WorkerPool = pool
	deps.Transforms = pool
	deps.Accountant = servicegateway.NewEntAccountant(deps.Client)

	routerConfig := router.Config{
		Cron: cron.Config{Enabled: cfg.Cron.Enabled, Secret: cfg.Cron.Secret},
		OAuth: oauth.Config{
			AppURL:        cfg.AppURL,
			SessionSecret: []byte(cfg.Auth.SessionSecret),
		},
	}
	handler, err := router.BuildRouter(runCtx, routerConfig, deps.Dependencies)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	server, err := buildHTTPServer(runCtx, cfg, deps, loggerInstance, handler)
	if err != nil {
		return err
	}
	scheduleManager.Start()

	loggerInstance.Info("server starting",
		slog.String("addr", server.Addr()),
		slog.String("environment", cfg.Environment),
	)
	return server.Run(runCtx)
}

// shutdownServerResources releases resources in dependency order after the
// HTTP server has stopped accepting work. Pool.Close is deliberately explicit:
// worker.Pool also observes runCtx, but this covers startup failures before a
// server is running and keeps ownership clear at the composition root.
func shutdownServerResources(ctx context.Context, timeout time.Duration, deps serverDependencies, scheduleManager *schedule.ScheduleManager, loggerInstance *slog.Logger) {
	shutdownCtx, cancel := gracefulShutdownContext(ctx, timeout)
	defer cancel()

	if deps.WorkerPool != nil {
		if err := deps.WorkerPool.Close(); err != nil && loggerInstance != nil {
			loggerInstance.Error("script worker pool shutdown failed", slog.Any("err", err))
		}
	}
	if scheduleManager != nil {
		if err := scheduleManager.Shutdown(shutdownCtx); err != nil && loggerInstance != nil {
			loggerInstance.Error("schedule manager shutdown failed", slog.Any("err", err))
		}
	}
	if deps.Redis != nil {
		if err := deps.Redis.Close(); err != nil && loggerInstance != nil {
			loggerInstance.Error("redis shutdown failed", slog.Any("err", err))
		}
	}
	if deps.Client != nil {
		if err := deps.Client.Close(); err != nil && loggerInstance != nil {
			loggerInstance.Error("database shutdown failed", slog.Any("err", err))
		}
	}
}

func gracefulShutdownContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return context.WithTimeout(context.WithoutCancel(parent), timeout)
}

// newWorkerPool 启动 script-worker 进程池;找不到可执行文件时返回错误,
// 由调用方降级(禁用脚本转换)而不是阻止服务启动。
func newWorkerPool(ctx context.Context, loggerInstance *slog.Logger) (*worker.Pool, error) {
	executable := resolveWorkerExecutable()
	if executable == "" {
		return nil, errors.New("script-worker executable not found (set NEX_WORKER_PATH or run make build)")
	}
	pool, err := worker.NewPool(ctx, worker.PoolOptions{
		Executable:  executable,
		WorkerCount: 2,
		JobTimeout:  30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("start script worker pool: %w", err)
	}
	return pool, nil
}

// resolveWorkerExecutable 按优先级查找 script-worker 可执行文件:
// NEX_WORKER_PATH 环境变量 > 与 server 同目录 > bin/script-worker。
func resolveWorkerExecutable() string {
	if path := strings.TrimSpace(os.Getenv("NEX_WORKER_PATH")); path != "" {
		return path
	}
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "script-worker")
		if fileExists(candidate) {
			return candidate
		}
	}
	for _, candidate := range []string{"bin/script-worker", "./bin/script-worker"} {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// isProduction 判断是否为生产环境(决定 cookie 是否带 Secure 标志)。
func isProduction(environment string) bool {
	env := strings.ToLower(strings.TrimSpace(environment))
	return env == "production" || env == "prod"
}
