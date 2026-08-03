package system

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	servicesystem "github.com/shuTwT/nex-api/backend/internal/service/system"
	"io"
	"net/http"
)

type Handler struct {
	service *servicesystem.Service
	mux     chi.Router
}

type responseEnvelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type initializeData struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Credits  int    `json:"credits"`
	Password string `json:"password,omitempty"`
}

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
		writeJSON(w, http.StatusInternalServerError, responseEnvelope{Success: false, Error: "system_unavailable"})
		return
	}
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes(r chi.Router) {
	r.Get("/api/system/initialized", h.initialized)
	r.Post("/api/system/initialize", h.initialize)
}

func (h *Handler) initialized(w http.ResponseWriter, r *http.Request) {
	initialized, err := h.service.Initialized(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, responseEnvelope{Success: false, Error: "initialization_status_failed"})
		return
	}
	writeJSON(w, http.StatusOK, responseEnvelope{
		Success: true,
		Data:    map[string]bool{"initialized": initialized},
	})
}

func (h *Handler) initialize(w http.ResponseWriter, r *http.Request) {
	var request servicesystem.InitializeRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{Success: false, Error: "invalid_request"})
		return
	}
	admin, err := h.service.Initialize(r.Context(), request)
	if errors.Is(err, servicesystem.ErrAlreadyInitialized) {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{Success: false, Error: "already_initialized"})
		return
	}
	if errors.Is(err, servicesystem.ErrInvalidInitialize) {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{Success: false, Error: "invalid_request"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, responseEnvelope{Success: false, Error: "initialization_failed"})
		return
	}
	writeJSON(w, http.StatusCreated, responseEnvelope{
		Success: true,
		Data: initializeData{
			ID: admin.ID, Email: admin.Email, Username: admin.Username, Role: admin.Role,
			Credits: admin.Credits,
		},
	})
}

func decodeJSON(r *http.Request, destination interface{}) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("system: multiple JSON values")
		}
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}
