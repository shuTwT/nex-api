package main

import (
	"context"
	"crypto/tls"
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

	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite"

	"github.com/shuTwT/nex-api/backend/internal/accounts"
	"github.com/shuTwT/nex-api/backend/internal/auth"
	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/catalog"
	"github.com/shuTwT/nex-api/backend/internal/config"
	"github.com/shuTwT/nex-api/backend/internal/payment"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
	"github.com/shuTwT/nex-api/backend/internal/schedule"
	"github.com/shuTwT/nex-api/backend/internal/stats"
	"github.com/shuTwT/nex-api/backend/internal/worker"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "", "path to a config file (optional; environment variables take precedence per key)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var options []config.Option
	if configFile != "" {
		options = append(options, config.WithConfigFile(configFile))
	}
	cfg, err := config.Load(options...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nex-api: load configuration: %v\n", err)
		os.Exit(1)
	}
	logger, err := runtime.NewLogger(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nex-api: create logger: %v\n", err)
		os.Exit(1)
	}

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("server exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	client, err := openDatabase(ctx, cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer client.Close()

	redisClient, err := newRedisClient(cfg.Redis)
	if err != nil {
		return fmt.Errorf("create redis client: %w", err)
	}
	defer redisClient.Close()
	// Redis 未就绪时只告警、不阻止启动;readyz 会持续暴露该状态。
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn("redis is unreachable; usage statistics and rate limiting will be unavailable", slog.Any("err", err))
	}

	statsStore, err := stats.NewStore(redisClient)
	if err != nil {
		return fmt.Errorf("create stats store: %w", err)
	}

	authService, err := auth.NewService(client, cfg.Auth,
		auth.WithSecureCookies(isProduction(cfg.Environment)),
	)
	if err != nil {
		return fmt.Errorf("create auth service: %w", err)
	}
	apiTokenStore, err := authz.NewEntTokenStore(client)
	if err != nil {
		return fmt.Errorf("create api token store: %w", err)
	}
	apiTokens, err := authz.NewTokenService(apiTokenStore)
	if err != nil {
		return fmt.Errorf("create api token service: %w", err)
	}
	audit, err := accounts.NewAuditService(client)
	if err != nil {
		return fmt.Errorf("create audit service: %w", err)
	}
	apiService, err := catalog.NewAPIService(client)
	if err != nil {
		return fmt.Errorf("create api service: %w", err)
	}
	mcpService, err := catalog.NewMCPService(client)
	if err != nil {
		return fmt.Errorf("create mcp service: %w", err)
	}
	paymentService, err := payment.NewService(client, cfg.AppURL)
	if err != nil {
		return fmt.Errorf("create payment service: %w", err)
	}
	statsSync, err := stats.NewSyncService(statsStore, client)
	if err != nil {
		return fmt.Errorf("create stats sync service: %w", err)
	}
	scheduleManager, err := schedule.NewScheduleManager(logger)
	if err != nil {
		return fmt.Errorf("create schedule manager: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		if shutdownErr := scheduleManager.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Error("schedule manager shutdown failed", slog.Any("err", shutdownErr))
		}
	}()
	if err := scheduleManager.RegisterTask(schedule.TaskKeyStatsSync, "将 Redis 中的 API/MCP 调用统计同步到数据库", statsSync.Sync); err != nil {
		return fmt.Errorf("register stats sync job: %w", err)
	}
	if err := scheduleManager.RegisterTask(schedule.TaskKeyExpirePayments, "关闭并过期超过支付时限的 pending 订单", func(ctx context.Context) error {
		count, expireErr := paymentService.ExpirePendingPayments(ctx)
		if count > 0 {
			logger.Info("expired pending payment orders", slog.Int("count", count))
		}
		return expireErr
	}); err != nil {
		return fmt.Errorf("register payment expiration job: %w", err)
	}
	scheduleService, err := schedule.NewService(client, scheduleManager)
	if err != nil {
		return fmt.Errorf("create schedule service: %w", err)
	}
	if err := scheduleService.EnsureDefaults(ctx,
		schedule.DefaultJob{
			Name: "统计数据同步", TaskKey: schedule.TaskKeyStatsSync, ScheduleType: "duration",
			Expression: cfg.Cron.Interval.String(), Enabled: cfg.Cron.Enabled,
			Description: "周期性将 Redis 调用计数同步到数据库",
		},
		schedule.DefaultJob{
			Name: "支付订单过期", TaskKey: schedule.TaskKeyExpirePayments, ScheduleType: "duration",
			Expression: "1m", Enabled: true,
			Description: "关闭 Provider 订单并将超过 expiredAt 的 pending 订单置为 expired",
		},
	); err != nil {
		return fmt.Errorf("bootstrap scheduled jobs: %w", err)
	}
	if err := scheduleService.LoadEnabled(ctx); err != nil {
		return fmt.Errorf("load scheduled jobs: %w", err)
	}

	pool, err := newWorkerPool(ctx, logger)
	if err != nil {
		logger.Warn("script worker unavailable; API script transforms will be disabled", slog.Any("err", err))
	} else {
		logger.Info("script worker pool started", slog.Int("workers", 2))
	}

	deps := dependencies{
		client:      client,
		redis:       redisClient,
		logger:      logger,
		statsStore:  statsStore,
		authService: authService,
		apiTokens:   apiTokens,
		audit:       audit,
		apiService:  apiService,
		mcpService:  mcpService,
		payment:     paymentService,
		statsSync:   statsSync,
		schedule:    scheduleService,
		workerPool:  pool,
	}

	handler, err := buildRouter(ctx, cfg, deps)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	server, err := runtime.NewServer(cfg, runtime.Dependencies{
		Handler: sessionMiddleware(authService, handler),
		Logger:  logger,
		Readiness: []runtime.DependencyCheck{
			{
				Name: "database",
				Check: func(ctx context.Context) error {
					_, err := client.User.Query().Count(ctx)
					return err
				},
			},
			{
				Name:  "redis",
				Check: func(ctx context.Context) error { return redisClient.Ping(ctx).Err() },
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}
	scheduleManager.Start()

	logger.Info("server starting",
		slog.String("addr", server.Addr()),
		slog.String("environment", cfg.Environment),
	)
	return server.RunWithSignals(ctx)
}

// newRedisClient 根据配置创建 Redis 客户端。
func newRedisClient(cfg config.Redis) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	if cfg.TLS {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	// 降低不可用时的阻塞时长,避免 readiness 检查长时间等待。
	opt.DialTimeout = 2 * time.Second
	opt.MaxRetries = 1
	return redis.NewClient(opt), nil
}

// newWorkerPool 启动 script-worker 进程池;找不到可执行文件时返回错误,
// 由调用方降级(禁用脚本转换)而不是阻止服务启动。
func newWorkerPool(ctx context.Context, logger *slog.Logger) (*worker.Pool, error) {
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
