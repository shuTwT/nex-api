// Package router assembles the root HTTP mux: it registers every business
// route on the passed services. It never creates services; cmd/server owns
// their construction.
package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/shuTwT/nex-api/backend/internal/handler/accounts"
	"github.com/shuTwT/nex-api/backend/internal/handler/ads"
	"github.com/shuTwT/nex-api/backend/internal/handler/auth"
	"github.com/shuTwT/nex-api/backend/internal/handler/catalog"
	"github.com/shuTwT/nex-api/backend/internal/handler/cron"
	"github.com/shuTwT/nex-api/backend/internal/handler/dashboard"
	"github.com/shuTwT/nex-api/backend/internal/handler/gateway"
	"github.com/shuTwT/nex-api/backend/internal/handler/httpapi/generated"
	"github.com/shuTwT/nex-api/backend/internal/handler/marketplace"
	"github.com/shuTwT/nex-api/backend/internal/handler/mcpgateway"
	"github.com/shuTwT/nex-api/backend/internal/handler/membership"
	"github.com/shuTwT/nex-api/backend/internal/handler/oauth"
	"github.com/shuTwT/nex-api/backend/internal/handler/payment"
	"github.com/shuTwT/nex-api/backend/internal/handler/schedule"
	"github.com/shuTwT/nex-api/backend/internal/handler/settings"
	"github.com/shuTwT/nex-api/backend/internal/handler/system"
	"github.com/shuTwT/nex-api/backend/internal/handler/upload"
	serviceaccounts "github.com/shuTwT/nex-api/backend/internal/service/accounts"
	serviceads "github.com/shuTwT/nex-api/backend/internal/service/ads"
	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
	servicedashboard "github.com/shuTwT/nex-api/backend/internal/service/dashboard"
	servicegateway "github.com/shuTwT/nex-api/backend/internal/service/gateway"
	servicemarketplace "github.com/shuTwT/nex-api/backend/internal/service/marketplace"
	servicemcpgateway "github.com/shuTwT/nex-api/backend/internal/service/mcpgateway"
	servicemembership "github.com/shuTwT/nex-api/backend/internal/service/membership"
	serviceoauth "github.com/shuTwT/nex-api/backend/internal/service/oauth"
	servicepayment "github.com/shuTwT/nex-api/backend/internal/service/payment"
	serviceschedule "github.com/shuTwT/nex-api/backend/internal/service/schedule"
	servicesettings "github.com/shuTwT/nex-api/backend/internal/service/settings"
	"github.com/shuTwT/nex-api/backend/internal/service/stats"
	servicesystem "github.com/shuTwT/nex-api/backend/internal/service/system"
	serviceupload "github.com/shuTwT/nex-api/backend/internal/service/upload"
)

// Dependencies carries every shared service constructed by cmd/server.
type Dependencies struct {
	Logger        *slog.Logger
	StatsStore    *stats.Store
	AuthService   *serviceauth.Service
	APITokens     *serviceauthz.TokenService
	Audit         *serviceaccounts.AuditService
	APIService    *servicecatalog.APIService
	MCPService    *servicecatalog.MCPService
	Payment       *servicepayment.Service
	StatsSync     *stats.SyncService
	Schedule      *serviceschedule.Service
	Transforms    servicegateway.Transformer
	Accountant    servicegateway.Accountant
	OAuthService  *serviceoauth.Service
	SessionIssuer oauth.SessionIssuer

	Accounts    *accounts.Services
	Ads         *serviceads.Service
	Categories  *servicecatalog.CategoryService
	Dashboard   *servicedashboard.Service
	Marketplace *servicemarketplace.Service
	Settings    *servicesettings.Service
	System      *servicesystem.Service
	Plans       *servicemembership.PlanService
	Membership  *servicemembership.MembershipService
	Redemption  *servicemembership.RedemptionService
	Upload      *serviceupload.Service
	McpLedger   *servicemcpgateway.EntCreditLedger
}

type Config struct {
	Cron  cron.Config
	OAuth oauth.Config
}

// BuildRouter registers all business routes and returns the root handler.
//
// Routing uses two systems:
//   - handwritten ServeMux for accounts/auth/catalog/dashboard/marketplace/
//     membership/oauth/payment/settings/system/ads/upload endpoints;
//   - OpenAPI generated routes (chi) under /api/v1/* and /api/cron/*,
//     implemented by gateway/mcpgateway/cron and stubbed by Unimplemented.
func BuildRouter(ctx context.Context, cfg Config, deps Dependencies) (http.Handler, error) {
	mux := http.NewServeMux()

	// --- accounts:users / tokens / personal / audit-logs ---
	if err := accounts.RegisterRoutes(mux, *deps.Accounts); err != nil {
		return nil, fmt.Errorf("router: accounts: %w", err)
	}

	// --- auth:csrf / login / me / logout ---
	if err := auth.RegisterRoutes(mux, deps.AuthService); err != nil {
		return nil, fmt.Errorf("router: auth: %w", err)
	}

	// --- catalog:apis / categories / mcp-services ---
	catalogHandler, err := catalog.NewHandler(deps.APIService, deps.Categories, deps.MCPService)
	if err != nil {
		return nil, fmt.Errorf("router: catalog handler: %w", err)
	}
	if err := catalog.RegisterRoutes(mux, catalogHandler); err != nil {
		return nil, fmt.Errorf("router: catalog: %w", err)
	}

	// --- dashboard & marketplace & settings & system & ads ---
	dashboardHandler, err := dashboard.NewHandler(deps.Dashboard)
	if err != nil {
		return nil, fmt.Errorf("router: dashboard: %w", err)
	}
	if err := dashboard.RegisterRoutes(mux, dashboardHandler); err != nil {
		return nil, fmt.Errorf("router: dashboard: %w", err)
	}
	marketplaceHandler, err := marketplace.NewHandler(deps.Marketplace)
	if err != nil {
		return nil, fmt.Errorf("router: marketplace: %w", err)
	}
	if err := marketplace.RegisterRoutes(mux, marketplaceHandler); err != nil {
		return nil, fmt.Errorf("router: marketplace: %w", err)
	}
	settingsHandler, err := settings.NewHandler(deps.Settings)
	if err != nil {
		return nil, fmt.Errorf("router: settings: %w", err)
	}
	if err := settings.RegisterRoutes(mux, settingsHandler); err != nil {
		return nil, fmt.Errorf("router: settings: %w", err)
	}
	if err := schedule.RegisterRoutes(mux, deps.Schedule); err != nil {
		return nil, fmt.Errorf("router: schedule: %w", err)
	}
	if err := system.RegisterRoutes(mux, deps.System); err != nil {
		return nil, fmt.Errorf("router: system: %w", err)
	}
	adsHandler, err := ads.NewHandler(deps.Ads)
	if err != nil {
		return nil, fmt.Errorf("router: ads handler: %w", err)
	}
	if err := ads.RegisterRoutes(mux, adsHandler); err != nil {
		return nil, fmt.Errorf("router: ads: %w", err)
	}

	// --- upload ---
	if err := upload.RegisterRoutes(mux, deps.Upload); err != nil {
		return nil, fmt.Errorf("router: upload: %w", err)
	}

	// --- payment & membership & redemption ---
	if err := payment.RegisterServiceRoutes(mux, deps.Payment); err != nil {
		return nil, fmt.Errorf("router: payment: %w", err)
	}
	if err := membership.RegisterServiceRoutes(mux, deps.Plans, deps.Membership, deps.Redemption); err != nil {
		return nil, fmt.Errorf("router: membership: %w", err)
	}

	// --- oauth(具体前缀先于 auth 的 /api/auth/ 注册,ServeMux 按最具体模式选择) ---
	oauthHandler, err := oauth.New(deps.OAuthService, deps.SessionIssuer, cfg.OAuth)
	if err != nil {
		return nil, fmt.Errorf("router: oauth: %w", err)
	}
	for _, pattern := range []string{
		"/api/auth/providers", "/api/auth/signin/", "/api/auth/callback/",
		"/api/auth/github", "/api/auth/github/",
		"/api/auth/easy1", "/api/auth/easy1/",
		"/api/auth/easy1auth", "/api/auth/easy1auth/",
	} {
		mux.Handle(pattern, oauthHandler)
	}

	// --- API 网关 / MCP 网关 / cron(OpenAPI 生成路由) ---
	gatewayHandler, err := gateway.New(gateway.Options{
		APIs:       deps.APIService,
		Tokens:     deps.APITokens,
		Transforms: deps.Transforms, // nil 时禁用脚本转换
		Usage:      deps.StatsStore,
		Audit:      deps.Audit,
		Accountant: deps.Accountant,
		Logger:     deps.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("router: gateway: %w", err)
	}
	mcpGatewayHandler, err := mcpgateway.New(mcpgateway.Services{
		Authenticator: deps.APITokens,
		Catalog:       deps.MCPService,
		Credits:       deps.McpLedger,
		Audits:        deps.Audit,
		Stats:         deps.StatsStore,
	}, mcpgateway.HandlerOptions{Logger: deps.Logger})
	if err != nil {
		return nil, fmt.Errorf("router: mcp gateway: %w", err)
	}
	cronHandler, err := cron.NewSyncStatsHandler(cfg.Cron, deps.StatsSync.Sync)
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
