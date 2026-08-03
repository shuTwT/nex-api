package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/shuTwT/nex-api/backend/internal/accounts"
	"github.com/shuTwT/nex-api/backend/internal/ads"
	"github.com/shuTwT/nex-api/backend/internal/auth"
	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/catalog"
	"github.com/shuTwT/nex-api/backend/internal/config"
	"github.com/shuTwT/nex-api/backend/internal/cron"
	"github.com/shuTwT/nex-api/backend/internal/dashboard"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/gateway"
	"github.com/shuTwT/nex-api/backend/internal/httpapi/generated"
	"github.com/shuTwT/nex-api/backend/internal/marketplace"
	"github.com/shuTwT/nex-api/backend/internal/mcpgateway"
	"github.com/shuTwT/nex-api/backend/internal/membership"
	"github.com/shuTwT/nex-api/backend/internal/oauth"
	"github.com/shuTwT/nex-api/backend/internal/payment"
	"github.com/shuTwT/nex-api/backend/internal/schedule"
	"github.com/shuTwT/nex-api/backend/internal/settings"
	"github.com/shuTwT/nex-api/backend/internal/stats"
	"github.com/shuTwT/nex-api/backend/internal/system"
	"github.com/shuTwT/nex-api/backend/internal/upload"
	"github.com/shuTwT/nex-api/backend/internal/worker"
)

// dependencies 汇集 buildRouter 所需的共享依赖,由 main 构造。
type dependencies struct {
	client *ent.Client
	redis  *redis.Client
	logger *slog.Logger

	statsStore  *stats.Store
	authService *auth.Service
	apiTokens   *authz.TokenService
	audit       *accounts.AuditService
	apiService  *catalog.APIService
	mcpService  *catalog.MCPService
	payment     *payment.Service
	statsSync   *stats.SyncService
	schedule    *schedule.Service
	workerPool  *worker.Pool // 可为 nil:禁用脚本转换
}

// buildRouter 注册全部业务路由并返回根 handler。
//
// 路由分两套体系:
//   - 手写 ServeMux:accounts/auth/catalog/dashboard/marketplace/membership/
//     oauth/payment/settings/system/ads/upload 等全部业务端点;
//   - OpenAPI 生成路由(chi):/api/v1/* 网关与 /api/cron/*,由
//     gateway/mcpgateway/cron 三个模块实现,其余端点由 Unimplemented 兜底(501)。
func buildRouter(ctx context.Context, cfg config.Config, deps dependencies) (http.Handler, error) {
	if deps.client == nil {
		return nil, errors.New("router: ent client is required")
	}
	mux := http.NewServeMux()

	// --- accounts:users / tokens / personal / audit-logs ---
	users, err := accounts.NewUserService(deps.client, deps.audit)
	if err != nil {
		return nil, fmt.Errorf("router: accounts users: %w", err)
	}
	tokens, err := accounts.NewTokenService(deps.client, deps.audit)
	if err != nil {
		return nil, fmt.Errorf("router: accounts tokens: %w", err)
	}
	profiles, err := accounts.NewProfileService(deps.client, deps.audit)
	if err != nil {
		return nil, fmt.Errorf("router: accounts profiles: %w", err)
	}
	if err := accounts.RegisterRoutes(mux, accounts.Services{
		Users: users, Tokens: tokens, Profiles: profiles, Audits: deps.audit,
	}); err != nil {
		return nil, fmt.Errorf("router: accounts: %w", err)
	}

	// --- auth:csrf / login / me / logout ---
	if err := auth.RegisterRoutes(mux, deps.authService); err != nil {
		return nil, fmt.Errorf("router: auth: %w", err)
	}

	// --- catalog:apis / categories / mcp-services ---
	categories, err := catalog.NewCategoryService(deps.client)
	if err != nil {
		return nil, fmt.Errorf("router: catalog categories: %w", err)
	}
	catalogHandler, err := catalog.NewHandler(deps.apiService, categories, deps.mcpService)
	if err != nil {
		return nil, fmt.Errorf("router: catalog handler: %w", err)
	}
	if err := catalog.RegisterRoutes(mux, catalogHandler); err != nil {
		return nil, fmt.Errorf("router: catalog: %w", err)
	}

	// --- dashboard & marketplace & settings & system & ads ---
	dashboardHandler, err := dashboard.NewHandler(deps.client, deps.statsStore)
	if err != nil {
		return nil, fmt.Errorf("router: dashboard: %w", err)
	}
	if err := dashboard.RegisterRoutes(mux, dashboardHandler); err != nil {
		return nil, fmt.Errorf("router: dashboard: %w", err)
	}
	marketplaceHandler, err := marketplace.NewHandler(deps.client, deps.statsStore)
	if err != nil {
		return nil, fmt.Errorf("router: marketplace: %w", err)
	}
	if err := marketplace.RegisterRoutes(mux, marketplaceHandler); err != nil {
		return nil, fmt.Errorf("router: marketplace: %w", err)
	}
	settingsHandler, err := settings.NewHandler(deps.client)
	if err != nil {
		return nil, fmt.Errorf("router: settings: %w", err)
	}
	if err := settings.RegisterRoutes(mux, settingsHandler); err != nil {
		return nil, fmt.Errorf("router: settings: %w", err)
	}
	if err := schedule.RegisterRoutes(mux, deps.schedule); err != nil {
		return nil, fmt.Errorf("router: schedule: %w", err)
	}
	systemService, err := system.NewService(deps.client)
	if err != nil {
		return nil, fmt.Errorf("router: system: %w", err)
	}
	if err := system.RegisterRoutes(mux, systemService); err != nil {
		return nil, fmt.Errorf("router: system: %w", err)
	}
	adsService, err := ads.NewService(deps.client)
	if err != nil {
		return nil, fmt.Errorf("router: ads service: %w", err)
	}
	adsHandler, err := ads.NewHandler(adsService)
	if err != nil {
		return nil, fmt.Errorf("router: ads handler: %w", err)
	}
	if err := ads.RegisterRoutes(mux, adsHandler); err != nil {
		return nil, fmt.Errorf("router: ads: %w", err)
	}

	// --- upload ---
	storage, err := upload.NewStorage(cfg.Upload)
	if err != nil {
		return nil, fmt.Errorf("router: upload storage: %w", err)
	}
	if err := upload.RegisterRoutes(mux, storage); err != nil {
		return nil, fmt.Errorf("router: upload: %w", err)
	}

	// --- payment & membership & redemption ---
	if err := payment.RegisterServiceRoutes(mux, deps.payment); err != nil {
		return nil, fmt.Errorf("router: payment: %w", err)
	}
	planService, err := membership.NewPlanService(deps.client)
	if err != nil {
		return nil, fmt.Errorf("router: membership plans: %w", err)
	}
	membershipService, err := membership.NewMembershipService(deps.client)
	if err != nil {
		return nil, fmt.Errorf("router: membership: %w", err)
	}
	redemptionService, err := membership.NewRedemptionService(deps.client)
	if err != nil {
		return nil, fmt.Errorf("router: membership redemption: %w", err)
	}
	if err := membership.RegisterServiceRoutes(mux, planService, membershipService, redemptionService); err != nil {
		return nil, fmt.Errorf("router: membership: %w", err)
	}

	// --- oauth(具体前缀先于 auth 的 /api/auth/ 注册,ServeMux 按最具体模式选择) ---
	oauthHandler, err := oauth.New(deps.client, deps.authService, cfg)
	if err != nil {
		return nil, fmt.Errorf("router: oauth: %w", err)
	}
	for _, pattern := range []string{
		"/api/auth/signin/", "/api/auth/callback/",
		"/api/auth/github", "/api/auth/github/",
		"/api/auth/easy1auth", "/api/auth/easy1auth/",
	} {
		mux.Handle(pattern, oauthHandler)
	}

	// --- API 网关 / MCP 网关 / cron(OpenAPI 生成路由) ---
	gatewayHandler, err := gateway.New(gateway.Options{
		APIs:       deps.apiService,
		Tokens:     deps.apiTokens,
		Transforms: deps.workerPool, // nil 时禁用脚本转换
		Usage:      deps.statsStore,
		Audit:      deps.audit,
		Database:   deps.client,
		Logger:     deps.logger,
	})
	if err != nil {
		return nil, fmt.Errorf("router: gateway: %w", err)
	}
	creditLedger, err := mcpgateway.NewEntCreditLedger(deps.client)
	if err != nil {
		return nil, fmt.Errorf("router: mcp gateway ledger: %w", err)
	}
	mcpGatewayHandler, err := mcpgateway.New(mcpgateway.Services{
		Authenticator: deps.apiTokens,
		Catalog:       deps.mcpService,
		Credits:       creditLedger,
		Audits:        deps.audit,
		Stats:         deps.statsStore,
	}, mcpgateway.HandlerOptions{Logger: deps.logger})
	if err != nil {
		return nil, fmt.Errorf("router: mcp gateway: %w", err)
	}
	cronHandler, err := cron.NewSyncStatsHandler(cfg.Cron, deps.statsSync.Sync)
	if err != nil {
		return nil, fmt.Errorf("router: cron: %w", err)
	}

	// 生成路由只通过 /api/v1/ 与 /api/cron/ 两个前缀对外暴露(其余生成
	// 端点由手写 mux 在规范路径提供,不会到达该 chi router)。
	openAPIRouter := chi.NewRouter()
	generated.HandlerFromMux(&openAPIServer{
		Unimplemented: generated.Unimplemented{},
		gateway:       gatewayHandler,
		mcpGateway:    mcpGatewayHandler,
		cron:          cronHandler,
	}, openAPIRouter)
	mux.Handle("/api/v1/", openAPIRouter)
	mux.Handle("/api/cron/", openAPIRouter)

	return mux, nil
}

// openAPIServer 把 OpenAPI 生成路由中已实现的端点组合到生成的
// ServerInterface 上,其余端点由 generated.Unimplemented 兜底(501)。
//
// 注意:不能直接匿名嵌入 generated.Unimplemented 与各 Handler —— Go 的
// 方法提升规则会让深度更浅的 Unimplemented 遮蔽各 Handler 的实现,
// 因此这里显式定义 8 个已实现端点的转发方法。
type openAPIServer struct {
	generated.Unimplemented
	gateway    *gateway.Handler
	mcpGateway *mcpgateway.Handler
	cron       *cron.SyncStatsHandler
}

func (s *openAPIServer) V1AliasRouteGet(w http.ResponseWriter, r *http.Request, alias string, params generated.V1AliasRouteGetParams) {
	s.gateway.V1AliasRouteGet(w, r, alias, params)
}

func (s *openAPIServer) V1AliasRoutePost(w http.ResponseWriter, r *http.Request, alias string) {
	s.gateway.V1AliasRoutePost(w, r, alias)
}

func (s *openAPIServer) V1AliasRoutePut(w http.ResponseWriter, r *http.Request, alias string) {
	s.gateway.V1AliasRoutePut(w, r, alias)
}

func (s *openAPIServer) V1AliasRoutePatch(w http.ResponseWriter, r *http.Request, alias string) {
	s.gateway.V1AliasRoutePatch(w, r, alias)
}

func (s *openAPIServer) V1AliasRouteDelete(w http.ResponseWriter, r *http.Request, alias string) {
	s.gateway.V1AliasRouteDelete(w, r, alias)
}

func (s *openAPIServer) V1McpIdentifierRouteOptions(w http.ResponseWriter, r *http.Request, identifier string) {
	s.mcpGateway.V1McpIdentifierRouteOptions(w, r, identifier)
}

func (s *openAPIServer) V1McpIdentifierRoutePost(w http.ResponseWriter, r *http.Request, identifier string) {
	s.mcpGateway.V1McpIdentifierRoutePost(w, r, identifier)
}

func (s *openAPIServer) CronSyncStatsRoutePost(w http.ResponseWriter, r *http.Request) {
	s.cron.CronSyncStatsRoutePost(w, r)
}
