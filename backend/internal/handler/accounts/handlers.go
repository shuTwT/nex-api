package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	serviceaccounts "github.com/shuTwT/nex-api/backend/internal/service/accounts"
	"github.com/shuTwT/nex-api/backend/internal/service/authz"
)

type Services struct {
	Users    *serviceaccounts.UserService
	Tokens   *serviceaccounts.TokenService
	Profiles *serviceaccounts.ProfileService
	Audits   *serviceaccounts.AuditService
}

type Handler struct {
	services Services
	mux      chi.Router
}

func NewHandler(services Services) http.Handler {
	handler := &Handler{services: services, mux: chi.NewRouter()}
	handler.registerRoutes(handler.mux)
	return handler
}

func RegisterRoutes(r chi.Router, services Services) error {
	if r == nil {
		return errors.New("accounts: route mux is nil")
	}
	if services.Users == nil || services.Tokens == nil || services.Profiles == nil || services.Audits == nil {
		return errors.New("accounts: services are incomplete")
	}
	(&Handler{services: services}).registerRoutes(r)
	return nil
}

func (h *Handler) registerRoutes(r chi.Router) {
	r.Get("/api/users", h.listUsers)
	r.Post("/api/users", h.createUser)
	r.Get("/api/users/stats", h.userStats)
	r.Get("/api/users/{id}", h.getUser)
	r.Put("/api/users/{id}", h.updateUser)
	r.Delete("/api/users/{id}", h.deleteUser)
	r.Get("/api/tokens", h.listTokens)
	r.Post("/api/tokens", h.createToken)
	r.Get("/api/tokens/stats", h.tokenStats)
	r.Get("/api/tokens/{id}", h.getToken)
	r.Put("/api/tokens/{id}", h.updateToken)
	r.Delete("/api/tokens/{id}", h.deleteToken)
	r.Put("/api/tokens/{id}/toggle", h.toggleToken)
	r.Get("/api/personal/profile", h.getProfile)
	r.Put("/api/personal/profile", h.updateProfile)
	r.Put("/api/personal/profile/password", h.updatePassword)
	r.Get("/api/audit-logs", h.listAudits)
	r.Post("/api/audit-logs", h.createAudit)
	r.Get("/api/audit-logs/stats", h.auditStats)
	r.Get("/api/audit-logs/export", h.exportAudits)
	r.Get("/api/audit-logs/{id}", h.getAudit)
	r.Put("/api/audit-logs/{id}", h.updateAudit)
	r.Delete("/api/audit-logs/{id}", h.deleteAudit)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

func principal(ctx context.Context) (authz.Principal, error) {
	value, err := authz.RequestPrincipal(ctx)
	if err != nil {
		if authz.IsUnauthenticated(err) {
			return authz.Principal{}, appRuntime.ErrUnauthorized
		}
		return authz.Principal{}, appRuntime.ErrForbidden
	}
	return value, nil
}

func admin(ctx context.Context) (authz.Principal, error) {
	value, err := principal(ctx)
	if err != nil {
		return authz.Principal{}, err
	}
	if value.Role != "admin" {
		return authz.Principal{}, appRuntime.ErrForbidden
	}
	return value, nil
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values")
	}
	return nil
}

func parsePage(r *http.Request) (serviceaccounts.PageRequest, error) {
	page, size := 1, 10
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil {
			return serviceaccounts.PageRequest{}, fmt.Errorf("page: %w", serviceaccounts.ErrInvalidRequest)
		}
	}
	if value := r.URL.Query().Get("limit"); value != "" {
		size, err = strconv.Atoi(value)
		if err != nil {
			return serviceaccounts.PageRequest{}, fmt.Errorf("limit: %w", serviceaccounts.ErrInvalidRequest)
		}
	}
	if page < 1 || size < 1 || size > 100 {
		return serviceaccounts.PageRequest{}, fmt.Errorf("pagination: %w", serviceaccounts.ErrInvalidRequest)
	}
	return serviceaccounts.PageRequest{Page: page, Size: size}, nil
}

func auditFilter(r *http.Request) (serviceaccounts.AuditFilter, error) {
	filter := serviceaccounts.AuditFilter{Search: r.URL.Query().Get("search"), Level: r.URL.Query().Get("level"), Status: r.URL.Query().Get("status")}
	for key, destination := range map[string]**time.Time{"startDate": &filter.StartDate, "endDate": &filter.EndDate} {
		value := r.URL.Query().Get(key)
		if value == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return serviceaccounts.AuditFilter{}, fmt.Errorf("%s: %w", key, serviceaccounts.ErrInvalidRequest)
		}
		*destination = &parsed
	}
	return filter, nil
}

func requestMetadata(r *http.Request) serviceaccounts.AuditMetadata {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	metadata, err := json.Marshal(map[string]string{"method": r.Method, "path": r.URL.Path})
	if err != nil {
		metadata = []byte(`{"method":"unknown","path":"unknown"}`)
	}
	return serviceaccounts.AuditMetadata{IP: ip, UserAgent: r.UserAgent(), Metadata: string(metadata)}
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	if err := appRuntime.WriteData(w, status, data); err != nil {
		return
	}
}

func writePage[T any](w http.ResponseWriter, status int, data []T, info serviceaccounts.PageInfo) {
	envelope := appRuntime.NewSuccessEnvelope(data)
	envelope.Pagination = &appRuntime.Pagination{Page: info.Page, PageSize: info.PageSize, Total: info.Total, TotalPages: info.TotalPages}
	if err := appRuntime.WriteEnvelope(w, status, envelope); err != nil {
		return
	}
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	if writeErr := appRuntime.WriteError(w, r, err); writeErr != nil {
		return
	}
}
