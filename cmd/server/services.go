package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/internal/handler/accounts"
	"github.com/shuTwT/nex-api/internal/handler/auth"
	"github.com/shuTwT/nex-api/internal/handler/router"
	"github.com/shuTwT/nex-api/internal/infra/config"
	"github.com/shuTwT/nex-api/internal/infra/database"
	"github.com/shuTwT/nex-api/internal/infra/schedule"
	"github.com/shuTwT/nex-api/internal/infra/storage"
	"github.com/shuTwT/nex-api/internal/infra/worker"
	"github.com/shuTwT/nex-api/internal/job"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
	serviceads "github.com/shuTwT/nex-api/internal/service/ads"
	serviceauth "github.com/shuTwT/nex-api/internal/service/auth"
	serviceauthz "github.com/shuTwT/nex-api/internal/service/authz"
	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
	servicedashboard "github.com/shuTwT/nex-api/internal/service/dashboard"
	servicemarketplace "github.com/shuTwT/nex-api/internal/service/marketplace"
	servicemcpgateway "github.com/shuTwT/nex-api/internal/service/mcpgateway"
	servicemembership "github.com/shuTwT/nex-api/internal/service/membership"
	serviceoauth "github.com/shuTwT/nex-api/internal/service/oauth"
	servicepayment "github.com/shuTwT/nex-api/internal/service/payment"
	serviceschedule "github.com/shuTwT/nex-api/internal/service/schedule"
	servicesettings "github.com/shuTwT/nex-api/internal/service/settings"
	"github.com/shuTwT/nex-api/internal/service/stats"
	servicesystem "github.com/shuTwT/nex-api/internal/service/system"
	serviceupload "github.com/shuTwT/nex-api/internal/service/upload"
)

// serverDependencies keeps infrastructure lifecycle handles in the
// composition root while exposing only application dependencies to router.
type serverDependencies struct {
	router.Dependencies
	Client     *ent.Client
	Redis      *redis.Client
	WorkerPool *worker.Pool
}

// buildServices loads the infrastructure, creates every service, registers
// the built-in scheduled jobs and returns the router dependencies.
func buildServices(ctx context.Context, cfg config.Config, logger *slog.Logger) (serverDependencies, *schedule.ScheduleManager, error) {
	var deps serverDependencies
	var scheduleManager *schedule.ScheduleManager
	var cleanup cleanupStack
	complete := false
	defer func() {
		if !complete {
			cleanup.run()
		}
	}()

	client, err := database.Open(ctx, cfg.Database, logger)
	if err != nil {
		return deps, nil, fmt.Errorf("open database: %w", err)
	}
	deps.Client = client
	cleanup.add(func() { _ = client.Close() })

	redisClient, err := database.NewRedis(cfg.Redis)
	if err != nil {
		return deps, nil, fmt.Errorf("create redis client: %w", err)
	}
	deps.Redis = redisClient
	cleanup.add(func() { _ = redisClient.Close() })
	deps.Logger = logger

	statsStore, err := stats.NewStore(redisClient)
	if err != nil {
		return deps, nil, fmt.Errorf("create stats store: %w", err)
	}
	deps.StatsStore = statsStore

	authService, err := serviceauth.NewService(client, cfg.Auth,
		serviceauth.WithSecureCookies(isProduction(cfg.Environment)),
	)
	if err != nil {
		return deps, nil, fmt.Errorf("create auth service: %w", err)
	}
	deps.AuthService = authService
	apiTokenStore, err := serviceauthz.NewEntTokenStore(client)
	if err != nil {
		return deps, nil, fmt.Errorf("create api token store: %w", err)
	}
	apiTokens, err := serviceauthz.NewTokenService(apiTokenStore)
	if err != nil {
		return deps, nil, fmt.Errorf("create api token service: %w", err)
	}
	deps.APITokens = apiTokens
	audit, err := serviceaccounts.NewAuditService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("create audit service: %w", err)
	}
	deps.Audit = audit
	apiService, err := servicecatalog.NewAPIService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("create api service: %w", err)
	}
	deps.APIService = apiService
	mcpService, err := servicecatalog.NewMCPService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("create mcp service: %w", err)
	}
	deps.MCPService = mcpService
	paymentService, err := servicepayment.NewService(client, cfg.AppURL)
	if err != nil {
		return deps, nil, fmt.Errorf("create payment service: %w", err)
	}
	deps.Payment = paymentService
	statsSync, err := stats.NewSyncService(statsStore, client)
	if err != nil {
		return deps, nil, fmt.Errorf("create stats sync service: %w", err)
	}
	deps.StatsSync = statsSync

	// --- scheduled jobs ---
	scheduleManager, err = schedule.NewScheduleManager(logger)
	if err != nil {
		return deps, nil, fmt.Errorf("create schedule manager: %w", err)
	}
	cleanup.add(func() {
		shutdownCtx, cancel := gracefulShutdownContext(ctx, cfg.Server.ShutdownTimeout)
		defer cancel()
		_ = scheduleManager.Shutdown(shutdownCtx)
	})
	if err := job.RegisterAll(scheduleManager, job.Dependencies{
		StatsSync: statsSync.Sync,
		Payments:  paymentService,
		Logger:    logger,
	}); err != nil {
		return deps, nil, err
	}
	scheduleService, err := serviceschedule.NewService(client, scheduleManager)
	if err != nil {
		return deps, nil, fmt.Errorf("create schedule service: %w", err)
	}
	deps.Schedule = scheduleService
	if err := scheduleService.EnsureDefaults(ctx,
		serviceschedule.DefaultJob{
			Name: "统计数据同步", TaskKey: schedule.TaskKeyStatsSync, ScheduleType: "duration",
			Expression: "5m", Enabled: true,
			Description: "周期性将 Redis 调用计数同步到数据库",
		},
		serviceschedule.DefaultJob{
			Name: "支付订单过期", TaskKey: schedule.TaskKeyExpirePayments, ScheduleType: "duration",
			Expression: "10m", Enabled: true,
			Description: "关闭 Provider 订单并将超过 expiredAt 的 pending 订单置为 expired",
		},
	); err != nil {
		return deps, nil, fmt.Errorf("bootstrap scheduled jobs: %w", err)
	}
	if err := scheduleService.LoadEnabled(ctx); err != nil {
		return deps, nil, fmt.Errorf("load scheduled jobs: %w", err)
	}

	// --- domain services used by the router ---
	users, err := serviceaccounts.NewUserService(client, audit)
	if err != nil {
		return deps, nil, fmt.Errorf("accounts users: %w", err)
	}
	tokens, err := serviceaccounts.NewTokenService(client, audit)
	if err != nil {
		return deps, nil, fmt.Errorf("accounts tokens: %w", err)
	}
	profiles, err := serviceaccounts.NewProfileService(client, audit)
	if err != nil {
		return deps, nil, fmt.Errorf("accounts profiles: %w", err)
	}
	deps.Accounts = &accounts.Services{Users: users, Tokens: tokens, Profiles: profiles, Audits: audit}
	categories, err := servicecatalog.NewCategoryService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("catalog categories: %w", err)
	}
	deps.Categories = categories
	dashboardService, err := servicedashboard.NewService(client, statsStore)
	if err != nil {
		return deps, nil, fmt.Errorf("dashboard service: %w", err)
	}
	deps.Dashboard = dashboardService
	marketplaceService, err := servicemarketplace.NewService(client, statsStore)
	if err != nil {
		return deps, nil, fmt.Errorf("marketplace service: %w", err)
	}
	deps.Marketplace = marketplaceService
	settingsService, err := servicesettings.NewService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("settings service: %w", err)
	}
	deps.Settings = settingsService
	systemService, err := servicesystem.NewService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("system service: %w", err)
	}
	deps.System = systemService
	adsService, err := serviceads.NewService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("ads service: %w", err)
	}
	deps.Ads = adsService
	uploadStorage, err := storage.NewStorage(cfg.Upload)
	if err != nil {
		return deps, nil, fmt.Errorf("upload storage: %w", err)
	}
	deps.Upload = serviceupload.NewService(uploadStorage)
	planService, err := servicemembership.NewPlanService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("membership plans: %w", err)
	}
	deps.Plans = planService
	membershipService, err := servicemembership.NewMembershipService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("membership: %w", err)
	}
	deps.Membership = membershipService
	redemptionService, err := servicemembership.NewRedemptionService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("membership redemption: %w", err)
	}
	deps.Redemption = redemptionService
	oauthService, err := serviceoauth.NewService(client)
	if err != nil {
		return deps, nil, fmt.Errorf("oauth service: %w", err)
	}
	deps.OAuthService = oauthService
	deps.SessionIssuer = auth.NewSessionCookieWriter(authService)
	mcpLedger, err := servicemcpgateway.NewEntCreditLedger(client)
	if err != nil {
		return deps, nil, fmt.Errorf("mcp gateway ledger: %w", err)
	}
	deps.McpLedger = mcpLedger

	complete = true
	return deps, scheduleManager, nil
}

// cleanupStack lets service construction roll back resources in the reverse
// order they were acquired. It deliberately accepts closures so individual
// resources retain their native close semantics.
type cleanupStack struct {
	actions []func()
}

func (s *cleanupStack) add(action func()) {
	if action != nil {
		s.actions = append(s.actions, action)
	}
}

func (s *cleanupStack) run() {
	for i := len(s.actions) - 1; i >= 0; i-- {
		s.actions[i]()
	}
}
