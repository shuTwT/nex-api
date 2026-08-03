package settings

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
	servicesettings "github.com/shuTwT/nex-api/backend/internal/service/settings"
)

type Handler struct{ service *servicesettings.Service }

func NewHandler(service *servicesettings.Service) (*Handler, error) {
	if service == nil {
		return nil, errors.New("settings: service is required")
	}
	return &Handler{service: service}, nil
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("settings: mux and handler are required")
	}
	mux.HandleFunc("GET /api/system-settings", handler.list)
	mux.HandleFunc("PUT /api/system-settings", handler.update)
	mux.HandleFunc("GET /api/system-settings/defaults", handler.defaults)
	mux.HandleFunc("GET /api/system-settings/announcement", handler.announcement)
	return nil
}

type updateRequest struct {
	Settings []servicesettings.UpdateItem `json:"settings"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	items, err := h.service.List(r.Context(), r.URL.Query().Get("category"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var request updateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	if err := h.service.Update(r.Context(), request.Settings); err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "设置已更新"})
}

func (h *Handler) defaults(w http.ResponseWriter, _ *http.Request) {
	writeData(w, http.StatusOK, h.service.Defaults())
}

func (h *Handler) announcement(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.Announcement(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, values)
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	principal, err := serviceauthz.AdminPolicy(r.Context())
	if err == nil && principal.UserID != "" {
		return true
	}
	status := http.StatusUnauthorized
	if errors.Is(err, serviceauthz.ErrForbidden) {
		status = http.StatusForbidden
	}
	code, message := "unauthorized", "authentication required"
	if status == http.StatusForbidden {
		code, message = "forbidden", "access denied"
	}
	writeError(w, r, appRuntime.NewAPIError(status, code, message, err))
	return false
}

func decodeJSON[T any](r *http.Request, destination *T) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "body", Reason: "invalid JSON"})
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "body", Reason: "must contain exactly one JSON value"})
	}
	return nil
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	_ = appRuntime.WriteData(w, status, data)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	_ = appRuntime.WriteError(w, r, err)
}
