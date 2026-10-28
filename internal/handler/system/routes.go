package system

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/shuTwT/nex-api/internal/pkg/utils/httpkit"
	servicesystem "github.com/shuTwT/nex-api/internal/service/system"
	"github.com/shuTwT/nex-api/pkg/domain/model"
	"net/http"
)

type Handler struct {
	service *servicesystem.Service
	mux     chi.Router
}

type initializeData = model.SystemInitializeResp

func NewHandler(service *servicesystem.Service) http.Handler {
	handler := &Handler{service: service, mux: chi.NewRouter()}
	handler.registerRoutes(handler.mux)
	return handler
}

func RegisterRoutes(r chi.Router, service *servicesystem.Service) error {
	if r == nil {
		return errors.New("system: route mux is nil")
	}
	if service == nil {
		return errors.New("system: service is nil")
	}
	(&Handler{service: service}).registerRoutes(r)
	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpkit.WriteFail(w, r, http.StatusInternalServerError, "system_unavailable", "系统服务不可用", nil)
		return
	}
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes(r chi.Router) {
	r.Get("/api/system/initialized", h.initialized)
	r.Post("/api/system/initialize", h.initialize)
}

// initialized handles GET /api/system/initialized.
//
// @Summary GET /api/system/initialized
// @ID system_initialized_route_get
// @Tags system
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope{data=bool}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/system/initialized [get]
func (h *Handler) initialized(w http.ResponseWriter, r *http.Request) {
	initialized, err := h.service.Initialized(r.Context())
	if err != nil {
		httpkit.WriteFail(w, r, http.StatusInternalServerError, "initialization_status_failed", "无法获取系统初始化状态", err)
		return
	}
	httpkit.WriteData(w, http.StatusOK, map[string]bool{"initialized": initialized})
}

// initialize handles POST /api/system/initialize.
//
// @Summary POST /api/system/initialize
// @ID system_initialize_route_post
// @Tags system
// @Accept json
// @Produce json
// @Param body body model.SystemInitializeReq true "request payload"
// @Success 200 {object} main.SwaggerEnvelope{data=model.SystemInitializeResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/system/initialize [post]
func (h *Handler) initialize(w http.ResponseWriter, r *http.Request) {
	var request model.SystemInitializeReq
	if err := httpkit.DecodeJSON(r, &request); err != nil {
		httpkit.WriteFail(w, r, http.StatusBadRequest, "invalid_request", "请求参数无效", err)
		return
	}
	admin, err := h.service.Initialize(r.Context(), request)
	if errors.Is(err, servicesystem.ErrAlreadyInitialized) {
		httpkit.WriteFail(w, r, http.StatusBadRequest, "already_initialized", "系统已初始化", nil)
		return
	}
	if errors.Is(err, servicesystem.ErrInvalidInitialize) {
		httpkit.WriteFail(w, r, http.StatusBadRequest, "invalid_request", "请求参数无效", err)
		return
	}
	if err != nil {
		httpkit.WriteFail(w, r, http.StatusInternalServerError, "initialization_failed", "系统初始化失败", err)
		return
	}
	httpkit.WriteData(w, http.StatusCreated, initializeData{
		ID: admin.ID, Email: admin.Email, Username: admin.Username, Role: admin.Role,
		Credits: admin.Credits,
	})
}
