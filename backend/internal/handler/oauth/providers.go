package oauth

import (
	"net/http"

	handlerutils "github.com/shuTwT/nex-api/backend/internal/pkg/utils"
)

type publicProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

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
