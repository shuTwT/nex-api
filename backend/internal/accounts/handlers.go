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

	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

type Services struct {
	Users    *UserService
	Tokens   *TokenService
	Profiles *ProfileService
	Audits   *AuditService
}

type Handler struct {
	services Services
	mux      *http.ServeMux
}

func NewHandler(services Services) http.Handler {
	handler := &Handler{services: services, mux: http.NewServeMux()}
	handler.registerRoutes()
	return handler
}

func RegisterRoutes(mux *http.ServeMux, services Services) error {
	if mux == nil {
		return errors.New("accounts: route mux is nil")
	}
	if services.Users == nil || services.Tokens == nil || services.Profiles == nil || services.Audits == nil {
		return errors.New("accounts: services are incomplete")
	}
	for _, path := range []string{"/api/users", "/api/users/", "/api/tokens", "/api/tokens/", "/api/personal/profile", "/api/personal/profile/", "/api/audit-logs", "/api/audit-logs/"} {
		mux.Handle(path, NewHandler(services))
	}
	return nil
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("GET /api/users", h.listUsers)
	h.mux.HandleFunc("POST /api/users", h.createUser)
	h.mux.HandleFunc("GET /api/users/stats", h.userStats)
	h.mux.HandleFunc("GET /api/users/{id}", h.getUser)
	h.mux.HandleFunc("PUT /api/users/{id}", h.updateUser)
	h.mux.HandleFunc("DELETE /api/users/{id}", h.deleteUser)
	h.mux.HandleFunc("GET /api/tokens", h.listTokens)
	h.mux.HandleFunc("POST /api/tokens", h.createToken)
	h.mux.HandleFunc("GET /api/tokens/stats", h.tokenStats)
	h.mux.HandleFunc("GET /api/tokens/{id}", h.getToken)
	h.mux.HandleFunc("PUT /api/tokens/{id}", h.updateToken)
	h.mux.HandleFunc("DELETE /api/tokens/{id}", h.deleteToken)
	h.mux.HandleFunc("PUT /api/tokens/{id}/toggle", h.toggleToken)
	h.mux.HandleFunc("GET /api/personal/profile", h.getProfile)
	h.mux.HandleFunc("PUT /api/personal/profile", h.updateProfile)
	h.mux.HandleFunc("PUT /api/personal/profile/password", h.updatePassword)
	h.mux.HandleFunc("GET /api/audit-logs", h.listAudits)
	h.mux.HandleFunc("POST /api/audit-logs", h.createAudit)
	h.mux.HandleFunc("GET /api/audit-logs/stats", h.auditStats)
	h.mux.HandleFunc("GET /api/audit-logs/export", h.exportAudits)
	h.mux.HandleFunc("GET /api/audit-logs/{id}", h.getAudit)
	h.mux.HandleFunc("PUT /api/audit-logs/{id}", h.updateAudit)
	h.mux.HandleFunc("DELETE /api/audit-logs/{id}", h.deleteAudit)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

func principal(ctx context.Context) (authz.Principal, error) {
	value, err := authz.RequestPrincipal(ctx)
	if err != nil {
		if authz.IsUnauthenticated(err) {
			return authz.Principal{}, runtime.ErrUnauthorized
		}
		return authz.Principal{}, runtime.ErrForbidden
	}
	return value, nil
}

func admin(ctx context.Context) (authz.Principal, error) {
	value, err := principal(ctx)
	if err != nil {
		return authz.Principal{}, err
	}
	if value.Role != "admin" {
		return authz.Principal{}, runtime.ErrForbidden
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

func parsePage(r *http.Request) (PageRequest, error) {
	page, size := 1, 10
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil {
			return PageRequest{}, fmt.Errorf("page: %w", ErrInvalidRequest)
		}
	}
	if value := r.URL.Query().Get("limit"); value != "" {
		size, err = strconv.Atoi(value)
		if err != nil {
			return PageRequest{}, fmt.Errorf("limit: %w", ErrInvalidRequest)
		}
	}
	if page < 1 || size < 1 || size > 100 {
		return PageRequest{}, fmt.Errorf("pagination: %w", ErrInvalidRequest)
	}
	return PageRequest{Page: page, Size: size}, nil
}

func auditFilter(r *http.Request) (AuditFilter, error) {
	filter := AuditFilter{Search: r.URL.Query().Get("search"), Level: r.URL.Query().Get("level"), Status: r.URL.Query().Get("status")}
	for key, destination := range map[string]**time.Time{"startDate": &filter.StartDate, "endDate": &filter.EndDate} {
		value := r.URL.Query().Get(key)
		if value == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return AuditFilter{}, fmt.Errorf("%s: %w", key, ErrInvalidRequest)
		}
		*destination = &parsed
	}
	return filter, nil
}

func requestMetadata(r *http.Request) AuditMetadata {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	metadata, err := json.Marshal(map[string]string{"method": r.Method, "path": r.URL.Path})
	if err != nil {
		metadata = []byte(`{"method":"unknown","path":"unknown"}`)
	}
	return AuditMetadata{IP: ip, UserAgent: r.UserAgent(), Metadata: string(metadata)}
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	if err := runtime.WriteData(w, status, data); err != nil {
		return
	}
}

func writePage[T any](w http.ResponseWriter, status int, data []T, info PageInfo) {
	envelope := runtime.NewSuccessEnvelope(data)
	envelope.Pagination = &runtime.Pagination{Page: info.Page, PageSize: info.PageSize, Total: info.Total, TotalPages: info.TotalPages}
	if err := runtime.WriteEnvelope(w, status, envelope); err != nil {
		return
	}
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	if writeErr := runtime.WriteError(w, r, err); writeErr != nil {
		return
	}
}
