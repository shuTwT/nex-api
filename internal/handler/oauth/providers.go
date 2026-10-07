package oauth

import (
	"net/http"

	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
)

type publicProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// providers handles GET /api/auth/providers.
//
// @Summary GET /api/auth/providers
// @ID auth_providers_route_get
// @Tags oauth
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/auth/providers [get]
func (h *Handler) providers(w http.ResponseWriter, r *http.Request) {
	providers, err := h.service.Providers(r.Context())
	if err != nil {
		h.writeOAuthError(w, http.StatusServiceUnavailable, "oauth_unavailable")
		return
	}
	available := make([]publicProvider, 0, len(providers))
	for _, provider := range providers {
		available = append(available, publicProvider{ID: provider.ID, Name: provider.DisplayName()})
	}
	handlerutils.WriteData(w, http.StatusOK, available)
}
