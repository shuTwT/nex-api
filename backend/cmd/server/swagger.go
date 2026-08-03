// Package main contains the source annotations for the public HTTP API.
// Swaggo parses this file to produce Swagger 2.0; the standard swagger2openapi
// converter then writes backend/openapi/openapi.yaml.
package main

// @title Nex API
// @version 1.0.0
// @description Contract for the Nex HTTP API.
// @license.name MIT
// @license.identifier MIT
// @license.url https://opensource.org/license/mit
// @BasePath /
// @securityDefinitions.apikey SessionCookie
// @in cookie
// @name session
// @securityDefinitions.apikey ApiTokenAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey CronSecretAuth
// @in header
// @name Authorization
type SwaggerDocument struct{}

// SwaggerEnvelope is the standard JSON response envelope returned by the API.
type SwaggerEnvelope struct {
	Success    bool               `json:"success"`
	Data       any                `json:"data,omitempty"`
	Error      string             `json:"error,omitempty"`
	Pagination *SwaggerPagination `json:"pagination,omitempty"`
}

// SwaggerPagination is the pagination metadata carried by list responses.
type SwaggerPagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// SwaggerRequest represents a JSON request payload. Endpoint handlers validate its fields.
type SwaggerRequest map[string]any

// SwaggerAdvertisementsIdGET1 documents GET /api/advertisements/{id}.
//
// @Summary GET /api/advertisements/{id}
// @ID advertisements_id_route_get
// @Tags advertisements
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/advertisements/{id} [get]
func SwaggerAdvertisementsIdGET1() {}

// SwaggerAdvertisementsIdPUT2 documents PUT /api/advertisements/{id}.
//
// @Summary PUT /api/advertisements/{id}
// @ID advertisements_id_route_put
// @Tags advertisements
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/advertisements/{id} [put]
func SwaggerAdvertisementsIdPUT2() {}

// SwaggerAdvertisementsIdDELETE3 documents DELETE /api/advertisements/{id}.
//
// @Summary DELETE /api/advertisements/{id}
// @ID advertisements_id_route_delete
// @Tags advertisements
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/advertisements/{id} [delete]
func SwaggerAdvertisementsIdDELETE3() {}

// SwaggerAdvertisementsIdTogglePUT4 documents PUT /api/advertisements/{id}/toggle.
//
// @Summary PUT /api/advertisements/{id}/toggle
// @ID advertisements_id_toggle_route_put
// @Tags advertisements
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/advertisements/{id}/toggle [put]
func SwaggerAdvertisementsIdTogglePUT4() {}

// SwaggerAdvertisementsByPositionPositionGET5 documents GET /api/advertisements/by-position/{position}.
//
// @Summary GET /api/advertisements/by-position/{position}
// @ID advertisements_by_position_position_route_get
// @Tags advertisements
// @Produce json
// @Param position path string true "position"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/advertisements/by-position/{position} [get]
func SwaggerAdvertisementsByPositionPositionGET5() {}

// SwaggerAdvertisementsGET6 documents GET /api/advertisements.
//
// @Summary GET /api/advertisements
// @ID advertisements_route_get
// @Tags advertisements
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/advertisements [get]
func SwaggerAdvertisementsGET6() {}

// SwaggerAdvertisementsPOST7 documents POST /api/advertisements.
//
// @Summary POST /api/advertisements
// @ID advertisements_route_post
// @Tags advertisements
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/advertisements [post]
func SwaggerAdvertisementsPOST7() {}

// SwaggerAdvertisementsStatsGET8 documents GET /api/advertisements/stats.
//
// @Summary GET /api/advertisements/stats
// @ID advertisements_stats_route_get
// @Tags advertisements
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/advertisements/stats [get]
func SwaggerAdvertisementsStatsGET8() {}

// SwaggerApisIdGET9 documents GET /api/apis/{id}.
//
// @Summary GET /api/apis/{id}
// @ID apis_id_route_get
// @Tags apis
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/apis/{id} [get]
func SwaggerApisIdGET9() {}

// SwaggerApisIdPUT10 documents PUT /api/apis/{id}.
//
// @Summary PUT /api/apis/{id}
// @ID apis_id_route_put
// @Tags apis
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/apis/{id} [put]
func SwaggerApisIdPUT10() {}

// SwaggerApisIdDELETE11 documents DELETE /api/apis/{id}.
//
// @Summary DELETE /api/apis/{id}
// @ID apis_id_route_delete
// @Tags apis
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/apis/{id} [delete]
func SwaggerApisIdDELETE11() {}

// SwaggerApisIdTogglePUT12 documents PUT /api/apis/{id}/toggle.
//
// @Summary PUT /api/apis/{id}/toggle
// @ID apis_id_toggle_route_put
// @Tags apis
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/apis/{id}/toggle [put]
func SwaggerApisIdTogglePUT12() {}

// SwaggerApisGET13 documents GET /api/apis.
//
// @Summary GET /api/apis
// @ID apis_route_get
// @Tags apis
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/apis [get]
func SwaggerApisGET13() {}

// SwaggerApisPOST14 documents POST /api/apis.
//
// @Summary POST /api/apis
// @ID apis_route_post
// @Tags apis
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/apis [post]
func SwaggerApisPOST14() {}

// SwaggerApisStatsGET15 documents GET /api/apis/stats.
//
// @Summary GET /api/apis/stats
// @ID apis_stats_route_get
// @Tags apis
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/apis/stats [get]
func SwaggerApisStatsGET15() {}

// SwaggerAuditLogsIdPUT16 documents PUT /api/audit-logs/{id}.
//
// @Summary PUT /api/audit-logs/{id}
// @ID audit_logs_id_route_put
// @Tags audit-logs
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/audit-logs/{id} [put]
func SwaggerAuditLogsIdPUT16() {}

// SwaggerAuditLogsIdDELETE17 documents DELETE /api/audit-logs/{id}.
//
// @Summary DELETE /api/audit-logs/{id}
// @ID audit_logs_id_route_delete
// @Tags audit-logs
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/audit-logs/{id} [delete]
func SwaggerAuditLogsIdDELETE17() {}

// SwaggerAuditLogsExportGET18 documents GET /api/audit-logs/export.
//
// @Summary GET /api/audit-logs/export
// @ID audit_logs_export_route_get
// @Tags audit-logs
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/audit-logs/export [get]
func SwaggerAuditLogsExportGET18() {}

// SwaggerAuditLogsGET19 documents GET /api/audit-logs.
//
// @Summary GET /api/audit-logs
// @ID audit_logs_route_get
// @Tags audit-logs
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/audit-logs [get]
func SwaggerAuditLogsGET19() {}

// SwaggerAuditLogsPOST20 documents POST /api/audit-logs.
//
// @Summary POST /api/audit-logs
// @ID audit_logs_route_post
// @Tags audit-logs
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/audit-logs [post]
func SwaggerAuditLogsPOST20() {}

// SwaggerAuditLogsStatsGET21 documents GET /api/audit-logs/stats.
//
// @Summary GET /api/audit-logs/stats
// @ID audit_logs_stats_route_get
// @Tags audit-logs
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/audit-logs/stats [get]
func SwaggerAuditLogsStatsGET21() {}

// SwaggerAuthLogoutPOST22 documents POST /api/auth/logout.
//
// @Summary POST /api/auth/logout
// @ID auth_logout_route_post
// @Tags auth
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/auth/logout [post]
func SwaggerAuthLogoutPOST22() {}

// SwaggerAuthMeGET23 documents GET /api/auth/me.
//
// @Summary GET /api/auth/me
// @ID auth_me_route_get
// @Tags auth
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/auth/me [get]
func SwaggerAuthMeGET23() {}

// SwaggerCategoriesIdPUT24 documents PUT /api/categories/{id}.
//
// @Summary PUT /api/categories/{id}
// @ID categories_id_route_put
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/categories/{id} [put]
func SwaggerCategoriesIdPUT24() {}

// SwaggerCategoriesIdDELETE25 documents DELETE /api/categories/{id}.
//
// @Summary DELETE /api/categories/{id}
// @ID categories_id_route_delete
// @Tags categories
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/categories/{id} [delete]
func SwaggerCategoriesIdDELETE25() {}

// SwaggerCategoriesGET26 documents GET /api/categories.
//
// @Summary GET /api/categories
// @ID categories_route_get
// @Tags categories
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/categories [get]
func SwaggerCategoriesGET26() {}

// SwaggerCategoriesPOST27 documents POST /api/categories.
//
// @Summary POST /api/categories
// @ID categories_route_post
// @Tags categories
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/categories [post]
func SwaggerCategoriesPOST27() {}

// SwaggerCronSyncStatsPOST28 documents POST /api/cron/sync-stats.
//
// @Summary POST /api/cron/sync-stats
// @ID cron_sync_stats_route_post
// @Tags cron
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security CronSecretAuth
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/cron/sync-stats [post]
func SwaggerCronSyncStatsPOST28() {}

// SwaggerDashboardActivityGET29 documents GET /api/dashboard/activity.
//
// @Summary GET /api/dashboard/activity
// @ID dashboard_activity_route_get
// @Tags dashboard
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/dashboard/activity [get]
func SwaggerDashboardActivityGET29() {}

// SwaggerDashboardStatsGET30 documents GET /api/dashboard/stats.
//
// @Summary GET /api/dashboard/stats
// @ID dashboard_stats_route_get
// @Tags dashboard
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/dashboard/stats [get]
func SwaggerDashboardStatsGET30() {}

// SwaggerDashboardTopApisGET31 documents GET /api/dashboard/top-apis.
//
// @Summary GET /api/dashboard/top-apis
// @ID dashboard_top_apis_route_get
// @Tags dashboard
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/dashboard/top-apis [get]
func SwaggerDashboardTopApisGET31() {}

// SwaggerDashboardUsageTrendGET32 documents GET /api/dashboard/usage-trend.
//
// @Summary GET /api/dashboard/usage-trend
// @ID dashboard_usage_trend_route_get
// @Tags dashboard
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/dashboard/usage-trend [get]
func SwaggerDashboardUsageTrendGET32() {}

// SwaggerMarketplaceApisIdGET33 documents GET /api/marketplace/apis/{id}.
//
// @Summary GET /api/marketplace/apis/{id}
// @ID marketplace_apis_id_route_get
// @Tags marketplace
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/marketplace/apis/{id} [get]
func SwaggerMarketplaceApisIdGET33() {}

// SwaggerMarketplaceApisGET34 documents GET /api/marketplace/apis.
//
// @Summary GET /api/marketplace/apis
// @ID marketplace_apis_route_get
// @Tags marketplace
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/marketplace/apis [get]
func SwaggerMarketplaceApisGET34() {}

// SwaggerMarketplaceMcpServicesGET35 documents GET /api/marketplace/mcp-services.
//
// @Summary GET /api/marketplace/mcp-services
// @ID marketplace_mcp_services_route_get
// @Tags marketplace
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/marketplace/mcp-services [get]
func SwaggerMarketplaceMcpServicesGET35() {}

// SwaggerMarketplaceMcpStatsGET36 documents GET /api/marketplace/mcp-stats.
//
// @Summary GET /api/marketplace/mcp-stats
// @ID marketplace_mcp_stats_route_get
// @Tags marketplace
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/marketplace/mcp-stats [get]
func SwaggerMarketplaceMcpStatsGET36() {}

// SwaggerMarketplaceStatsGET37 documents GET /api/marketplace/stats.
//
// @Summary GET /api/marketplace/stats
// @ID marketplace_stats_route_get
// @Tags marketplace
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/marketplace/stats [get]
func SwaggerMarketplaceStatsGET37() {}

// SwaggerMcpServicesIdPUT38 documents PUT /api/mcp-services/{id}.
//
// @Summary PUT /api/mcp-services/{id}
// @ID mcp_services_id_route_put
// @Tags mcp-services
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/mcp-services/{id} [put]
func SwaggerMcpServicesIdPUT38() {}

// SwaggerMcpServicesIdDELETE39 documents DELETE /api/mcp-services/{id}.
//
// @Summary DELETE /api/mcp-services/{id}
// @ID mcp_services_id_route_delete
// @Tags mcp-services
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/mcp-services/{id} [delete]
func SwaggerMcpServicesIdDELETE39() {}

// SwaggerMcpServicesIdTogglePUT40 documents PUT /api/mcp-services/{id}/toggle.
//
// @Summary PUT /api/mcp-services/{id}/toggle
// @ID mcp_services_id_toggle_route_put
// @Tags mcp-services
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/mcp-services/{id}/toggle [put]
func SwaggerMcpServicesIdTogglePUT40() {}

// SwaggerMcpServicesGET41 documents GET /api/mcp-services.
//
// @Summary GET /api/mcp-services
// @ID mcp_services_route_get
// @Tags mcp-services
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/mcp-services [get]
func SwaggerMcpServicesGET41() {}

// SwaggerMcpServicesPOST42 documents POST /api/mcp-services.
//
// @Summary POST /api/mcp-services
// @ID mcp_services_route_post
// @Tags mcp-services
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/mcp-services [post]
func SwaggerMcpServicesPOST42() {}

// SwaggerMcpServicesStatsGET43 documents GET /api/mcp-services/stats.
//
// @Summary GET /api/mcp-services/stats
// @ID mcp_services_stats_route_get
// @Tags mcp-services
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/mcp-services/stats [get]
func SwaggerMcpServicesStatsGET43() {}

// SwaggerMembershipCurrentGET44 documents GET /api/membership/current.
//
// @Summary GET /api/membership/current
// @ID membership_current_route_get
// @Tags membership
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/membership/current [get]
func SwaggerMembershipCurrentGET44() {}

// SwaggerMembershipPlansGET45 documents GET /api/membership/plans.
//
// @Summary GET /api/membership/plans
// @ID membership_plans_route_get
// @Tags membership
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/membership/plans [get]
func SwaggerMembershipPlansGET45() {}

// SwaggerMembershipSubscribePOST46 documents POST /api/membership/subscribe.
//
// @Summary POST /api/membership/subscribe
// @ID membership_subscribe_route_post
// @Tags membership
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/membership/subscribe [post]
func SwaggerMembershipSubscribePOST46() {}

// SwaggerPaymentOutTradeNoCancelPOST47 documents POST /api/payment/{outTradeNo}/cancel.
//
// @Summary POST /api/payment/{outTradeNo}/cancel
// @ID payment_outtradeno_cancel_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param outTradeNo path string true "outTradeNo"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/{outTradeNo}/cancel [post]
func SwaggerPaymentOutTradeNoCancelPOST47() {}

// SwaggerPaymentOutTradeNoGET48 documents GET /api/payment/{outTradeNo}.
//
// @Summary GET /api/payment/{outTradeNo}
// @ID payment_outtradeno_route_get
// @Tags payment
// @Produce json
// @Param outTradeNo path string true "outTradeNo"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/{outTradeNo} [get]
func SwaggerPaymentOutTradeNoGET48() {}

// SwaggerPaymentOutTradeNoStatusGET49 documents GET /api/payment/{outTradeNo}/status.
//
// @Summary GET /api/payment/{outTradeNo}/status
// @ID payment_outtradeno_status_route_get
// @Tags payment
// @Produce json
// @Param outTradeNo path string true "outTradeNo"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/{outTradeNo}/status [get]
func SwaggerPaymentOutTradeNoStatusGET49() {}

// SwaggerPaymentBusinessRechargePOST50 documents POST /api/payment/business/recharge.
//
// @Summary POST /api/payment/business/recharge
// @ID payment_business_recharge_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/business/recharge [post]
func SwaggerPaymentBusinessRechargePOST50() {}

// SwaggerPaymentBusinessSubscriptionPOST51 documents POST /api/payment/business/subscription.
//
// @Summary POST /api/payment/business/subscription
// @ID payment_business_subscription_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/business/subscription [post]
func SwaggerPaymentBusinessSubscriptionPOST51() {}

// SwaggerPaymentCallbackAlipayPOST52 documents POST /api/payment/callback/alipay.
//
// @Summary POST /api/payment/callback/alipay
// @ID payment_callback_alipay_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/callback/alipay [post]
func SwaggerPaymentCallbackAlipayPOST52() {}

// SwaggerPaymentCallbackMockPOST53 documents POST /api/payment/callback/mock.
//
// @Summary POST /api/payment/callback/mock
// @ID payment_callback_mock_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/callback/mock [post]
func SwaggerPaymentCallbackMockPOST53() {}

// SwaggerPaymentCallbackWechatPOST54 documents POST /api/payment/callback/wechat.
//
// @Summary POST /api/payment/callback/wechat
// @ID payment_callback_wechat_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/callback/wechat [post]
func SwaggerPaymentCallbackWechatPOST54() {}

// SwaggerPaymentMethodsGET55 documents GET /api/payment/methods.
//
// @Summary GET /api/payment/methods
// @ID payment_methods_route_get
// @Tags payment
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/methods [get]
func SwaggerPaymentMethodsGET55() {}

// SwaggerPaymentMethodsPOST56 documents POST /api/payment/methods.
//
// @Summary POST /api/payment/methods
// @ID payment_methods_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/methods [post]
func SwaggerPaymentMethodsPOST56() {}

// SwaggerPaymentSettingsGET57 documents GET /api/payment/settings.
//
// @Summary GET /api/payment/settings
// @ID payment_settings_route_get
// @Tags payment
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/settings [get]
func SwaggerPaymentSettingsGET57() {}

// SwaggerPaymentUserGET58 documents GET /api/payment/user.
//
// @Summary GET /api/payment/user
// @ID payment_user_route_get
// @Tags payment
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/payment/user [get]
func SwaggerPaymentUserGET58() {}

// SwaggerPersonalProfileGET59 documents GET /api/personal/profile.
//
// @Summary GET /api/personal/profile
// @ID personal_profile_route_get
// @Tags personal
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/personal/profile [get]
func SwaggerPersonalProfileGET59() {}

// SwaggerPersonalRedeemLookupPOST60 documents POST /api/personal/redeem/lookup.
//
// @Summary POST /api/personal/redeem/lookup
// @ID personal_redeem_lookup_route_post
// @Tags personal
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/personal/redeem/lookup [post]
func SwaggerPersonalRedeemLookupPOST60() {}

// SwaggerPersonalRedeemPOST61 documents POST /api/personal/redeem.
//
// @Summary POST /api/personal/redeem
// @ID personal_redeem_route_post
// @Tags personal
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/personal/redeem [post]
func SwaggerPersonalRedeemPOST61() {}

// SwaggerRechargePOST62 documents POST /api/recharge.
//
// @Summary POST /api/recharge
// @ID recharge_route_post
// @Tags recharge
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/recharge [post]
func SwaggerRechargePOST62() {}

// SwaggerRedemptionCodesIdDELETE63 documents DELETE /api/redemption-codes/{id}.
//
// @Summary DELETE /api/redemption-codes/{id}
// @ID redemption_codes_id_route_delete
// @Tags redemption-codes
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/redemption-codes/{id} [delete]
func SwaggerRedemptionCodesIdDELETE63() {}

// SwaggerRedemptionCodesBatchDELETE64 documents DELETE /api/redemption-codes/batch.
//
// @Summary DELETE /api/redemption-codes/batch
// @ID redemption_codes_batch_route_delete
// @Tags redemption-codes
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/redemption-codes/batch [delete]
func SwaggerRedemptionCodesBatchDELETE64() {}

// SwaggerRedemptionCodesBatchPOST65 documents POST /api/redemption-codes/batch.
//
// @Summary POST /api/redemption-codes/batch
// @ID redemption_codes_batch_route_post
// @Tags redemption-codes
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/redemption-codes/batch [post]
func SwaggerRedemptionCodesBatchPOST65() {}

// SwaggerRedemptionCodesExportGET66 documents GET /api/redemption-codes/export.
//
// @Summary GET /api/redemption-codes/export
// @ID redemption_codes_export_route_get
// @Tags redemption-codes
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/redemption-codes/export [get]
func SwaggerRedemptionCodesExportGET66() {}

// SwaggerRedemptionCodesPlansGET67 documents GET /api/redemption-codes/plans.
//
// @Summary GET /api/redemption-codes/plans
// @ID redemption_codes_plans_route_get
// @Tags redemption-codes
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/redemption-codes/plans [get]
func SwaggerRedemptionCodesPlansGET67() {}

// SwaggerRedemptionCodesGET68 documents GET /api/redemption-codes.
//
// @Summary GET /api/redemption-codes
// @ID redemption_codes_route_get
// @Tags redemption-codes
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/redemption-codes [get]
func SwaggerRedemptionCodesGET68() {}

// SwaggerRedemptionCodesPOST69 documents POST /api/redemption-codes.
//
// @Summary POST /api/redemption-codes
// @ID redemption_codes_route_post
// @Tags redemption-codes
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/redemption-codes [post]
func SwaggerRedemptionCodesPOST69() {}

// SwaggerStatsAliasGET70 documents GET /api/stats/{alias}.
//
// @Summary GET /api/stats/{alias}
// @ID stats_alias_route_get
// @Tags stats
// @Produce json
// @Param alias path string true "alias"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/stats/{alias} [get]
func SwaggerStatsAliasGET70() {}

// SwaggerStatsGET71 documents GET /api/stats.
//
// @Summary GET /api/stats
// @ID stats_route_get
// @Tags stats
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/stats [get]
func SwaggerStatsGET71() {}

// SwaggerSubscriptionPlansIdGET72 documents GET /api/subscription-plans/{id}.
//
// @Summary GET /api/subscription-plans/{id}
// @ID subscription_plans_id_route_get
// @Tags subscription-plans
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/subscription-plans/{id} [get]
func SwaggerSubscriptionPlansIdGET72() {}

// SwaggerSubscriptionPlansIdPUT73 documents PUT /api/subscription-plans/{id}.
//
// @Summary PUT /api/subscription-plans/{id}
// @ID subscription_plans_id_route_put
// @Tags subscription-plans
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/subscription-plans/{id} [put]
func SwaggerSubscriptionPlansIdPUT73() {}

// SwaggerSubscriptionPlansIdDELETE74 documents DELETE /api/subscription-plans/{id}.
//
// @Summary DELETE /api/subscription-plans/{id}
// @ID subscription_plans_id_route_delete
// @Tags subscription-plans
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/subscription-plans/{id} [delete]
func SwaggerSubscriptionPlansIdDELETE74() {}

// SwaggerSubscriptionPlansGET75 documents GET /api/subscription-plans.
//
// @Summary GET /api/subscription-plans
// @ID subscription_plans_route_get
// @Tags subscription-plans
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/subscription-plans [get]
func SwaggerSubscriptionPlansGET75() {}

// SwaggerSubscriptionPlansPOST76 documents POST /api/subscription-plans.
//
// @Summary POST /api/subscription-plans
// @ID subscription_plans_route_post
// @Tags subscription-plans
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/subscription-plans [post]
func SwaggerSubscriptionPlansPOST76() {}

// SwaggerSystemInitializePOST77 documents POST /api/system/initialize.
//
// @Summary POST /api/system/initialize
// @ID system_initialize_route_post
// @Tags system
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/system/initialize [post]
func SwaggerSystemInitializePOST77() {}

// SwaggerSystemInitializedGET78 documents GET /api/system/initialized.
//
// @Summary GET /api/system/initialized
// @ID system_initialized_route_get
// @Tags system
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/system/initialized [get]
func SwaggerSystemInitializedGET78() {}

// SwaggerSystemSettingsAnnouncementGET79 documents GET /api/system-settings/announcement.
//
// @Summary GET /api/system-settings/announcement
// @ID system_settings_announcement_route_get
// @Tags system-settings
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/system-settings/announcement [get]
func SwaggerSystemSettingsAnnouncementGET79() {}

// SwaggerSystemSettingsDefaultsGET80 documents GET /api/system-settings/defaults.
//
// @Summary GET /api/system-settings/defaults
// @ID system_settings_defaults_route_get
// @Tags system-settings
// @Produce json
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/system-settings/defaults [get]
func SwaggerSystemSettingsDefaultsGET80() {}

// SwaggerSystemSettingsGET81 documents GET /api/system-settings.
//
// @Summary GET /api/system-settings
// @ID system_settings_route_get
// @Tags system-settings
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/system-settings [get]
func SwaggerSystemSettingsGET81() {}

// SwaggerSystemSettingsPUT82 documents PUT /api/system-settings.
//
// @Summary PUT /api/system-settings
// @ID system_settings_route_put
// @Tags system-settings
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/system-settings [put]
func SwaggerSystemSettingsPUT82() {}

// SwaggerTokensIdPUT83 documents PUT /api/tokens/{id}.
//
// @Summary PUT /api/tokens/{id}
// @ID tokens_id_route_put
// @Tags tokens
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/tokens/{id} [put]
func SwaggerTokensIdPUT83() {}

// SwaggerTokensIdDELETE84 documents DELETE /api/tokens/{id}.
//
// @Summary DELETE /api/tokens/{id}
// @ID tokens_id_route_delete
// @Tags tokens
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/tokens/{id} [delete]
func SwaggerTokensIdDELETE84() {}

// SwaggerTokensIdTogglePUT85 documents PUT /api/tokens/{id}/toggle.
//
// @Summary PUT /api/tokens/{id}/toggle
// @ID tokens_id_toggle_route_put
// @Tags tokens
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/tokens/{id}/toggle [put]
func SwaggerTokensIdTogglePUT85() {}

// SwaggerTokensGET86 documents GET /api/tokens.
//
// @Summary GET /api/tokens
// @ID tokens_route_get
// @Tags tokens
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/tokens [get]
func SwaggerTokensGET86() {}

// SwaggerTokensPOST87 documents POST /api/tokens.
//
// @Summary POST /api/tokens
// @ID tokens_route_post
// @Tags tokens
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/tokens [post]
func SwaggerTokensPOST87() {}

// SwaggerTokensStatsGET88 documents GET /api/tokens/stats.
//
// @Summary GET /api/tokens/stats
// @ID tokens_stats_route_get
// @Tags tokens
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/tokens/stats [get]
func SwaggerTokensStatsGET88() {}

// SwaggerUploadFilenameGET89 documents GET /api/upload/{filename}.
//
// @Summary GET /api/upload/{filename}
// @ID upload_filename_route_get
// @Tags upload
// @Produce json
// @Param filename path string true "filename"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/upload/{filename} [get]
func SwaggerUploadFilenameGET89() {}

// SwaggerUploadPOST90 documents POST /api/upload.
//
// @Summary POST /api/upload
// @ID upload_route_post
// @Tags upload
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/upload [post]
func SwaggerUploadPOST90() {}

// SwaggerUsageGET91 documents GET /api/usage.
//
// @Summary GET /api/usage
// @ID usage_route_get
// @Tags usage
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/usage [get]
func SwaggerUsageGET91() {}

// SwaggerUsersIdGET92 documents GET /api/users/{id}.
//
// @Summary GET /api/users/{id}
// @ID users_id_route_get
// @Tags users
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/users/{id} [get]
func SwaggerUsersIdGET92() {}

// SwaggerUsersIdPUT93 documents PUT /api/users/{id}.
//
// @Summary PUT /api/users/{id}
// @ID users_id_route_put
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/users/{id} [put]
func SwaggerUsersIdPUT93() {}

// SwaggerUsersIdDELETE94 documents DELETE /api/users/{id}.
//
// @Summary DELETE /api/users/{id}
// @ID users_id_route_delete
// @Tags users
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/users/{id} [delete]
func SwaggerUsersIdDELETE94() {}

// SwaggerUsersGET95 documents GET /api/users.
//
// @Summary GET /api/users
// @ID users_route_get
// @Tags users
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/users [get]
func SwaggerUsersGET95() {}

// SwaggerUsersPOST96 documents POST /api/users.
//
// @Summary POST /api/users
// @ID users_route_post
// @Tags users
// @Accept json
// @Produce json
// @Param body body SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/users [post]
func SwaggerUsersPOST96() {}

// SwaggerUsersStatsGET97 documents GET /api/users/stats.
//
// @Summary GET /api/users/stats
// @ID users_stats_route_get
// @Tags users
// @Produce json
// @Security SessionCookie
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/users/stats [get]
func SwaggerUsersStatsGET97() {}

// SwaggerV1AliasGET98 documents GET /api/v1/{alias}.
//
// @Summary GET /api/v1/{alias}
// @ID v1_alias_route_get
// @Tags gateway
// @Produce json
// @Param alias path string true "alias"
// @Security ApiTokenAuth
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/{alias} [get]
func SwaggerV1AliasGET98() {}

// SwaggerV1AliasPOST99 documents POST /api/v1/{alias}.
//
// @Summary POST /api/v1/{alias}
// @ID v1_alias_route_post
// @Tags gateway
// @Accept json
// @Produce json
// @Param alias path string true "alias"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security ApiTokenAuth
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/{alias} [post]
func SwaggerV1AliasPOST99() {}

// SwaggerV1AliasPUT100 documents PUT /api/v1/{alias}.
//
// @Summary PUT /api/v1/{alias}
// @ID v1_alias_route_put
// @Tags gateway
// @Accept json
// @Produce json
// @Param alias path string true "alias"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security ApiTokenAuth
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/{alias} [put]
func SwaggerV1AliasPUT100() {}

// SwaggerV1AliasDELETE101 documents DELETE /api/v1/{alias}.
//
// @Summary DELETE /api/v1/{alias}
// @ID v1_alias_route_delete
// @Tags gateway
// @Produce json
// @Param alias path string true "alias"
// @Security ApiTokenAuth
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/{alias} [delete]
func SwaggerV1AliasDELETE101() {}

// SwaggerV1AliasPATCH102 documents PATCH /api/v1/{alias}.
//
// @Summary PATCH /api/v1/{alias}
// @ID v1_alias_route_patch
// @Tags gateway
// @Accept json
// @Produce json
// @Param alias path string true "alias"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security ApiTokenAuth
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/{alias} [patch]
func SwaggerV1AliasPATCH102() {}

// SwaggerV1McpIdentifierOPTIONS103 documents OPTIONS /api/v1/mcp/{identifier}.
//
// @Summary OPTIONS /api/v1/mcp/{identifier}
// @ID v1_mcp_identifier_route_options
// @Tags gateway
// @Produce json
// @Param identifier path string true "identifier"
// @Security ApiTokenAuth
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/mcp/{identifier} [options]
func SwaggerV1McpIdentifierOPTIONS103() {}

// SwaggerV1McpIdentifierPOST104 documents POST /api/v1/mcp/{identifier}.
//
// @Summary POST /api/v1/mcp/{identifier}
// @ID v1_mcp_identifier_route_post
// @Tags gateway
// @Accept json
// @Produce json
// @Param identifier path string true "identifier"
// @Param body body SwaggerRequest false "JSON request payload"
// @Security ApiTokenAuth
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 401 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/mcp/{identifier} [post]
func SwaggerV1McpIdentifierPOST104() {}
