package catalog

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	"github.com/shuTwT/nex-api/backend/internal/middleware"
	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
)

type Handler struct {
	apis       *servicecatalog.APIService
	categories *servicecatalog.CategoryService
	mcp        *servicecatalog.MCPService
}

func NewHandler(apis *servicecatalog.APIService, categories *servicecatalog.CategoryService, mcp *servicecatalog.MCPService) (*Handler, error) {
	if apis == nil || categories == nil || mcp == nil {
		return nil, errors.New("catalog: all services are required")
	}
	return &Handler{apis: apis, categories: categories, mcp: mcp}, nil
}

func RegisterRoutes(r chi.Router, handler *Handler) error {
	if r == nil {
		return errors.New("catalog: route mux is nil")
	}
	if handler == nil {
		return errors.New("catalog: handler is nil")
	}
	admin := func(next http.Handler) http.Handler { return middleware.RequireAdmin(next) }
	r.Method(http.MethodGet, "/api/apis", admin(http.HandlerFunc(handler.listAPIs)))
	r.Method(http.MethodPost, "/api/apis", admin(http.HandlerFunc(handler.createAPI)))
	r.Method(http.MethodGet, "/api/apis/stats", admin(http.HandlerFunc(handler.apiStats)))
	r.Method(http.MethodGet, "/api/apis/{id}", admin(http.HandlerFunc(handler.getAPI)))
	r.Method(http.MethodPut, "/api/apis/{id}", admin(http.HandlerFunc(handler.updateAPI)))
	r.Method(http.MethodDelete, "/api/apis/{id}", admin(http.HandlerFunc(handler.deleteAPI)))
	r.Method(http.MethodPut, "/api/apis/{id}/toggle", admin(http.HandlerFunc(handler.toggleAPI)))
	r.Method(http.MethodGet, "/api/categories", admin(http.HandlerFunc(handler.listCategories)))
	r.Method(http.MethodPost, "/api/categories", admin(http.HandlerFunc(handler.createCategory)))
	r.Method(http.MethodGet, "/api/categories/{id}", admin(http.HandlerFunc(handler.getCategory)))
	r.Method(http.MethodPut, "/api/categories/{id}", admin(http.HandlerFunc(handler.updateCategory)))
	r.Method(http.MethodDelete, "/api/categories/{id}", admin(http.HandlerFunc(handler.deleteCategory)))
	r.Method(http.MethodGet, "/api/mcp-services", admin(http.HandlerFunc(handler.listMCP)))
	r.Method(http.MethodPost, "/api/mcp-services", admin(http.HandlerFunc(handler.createMCP)))
	r.Method(http.MethodGet, "/api/mcp-services/stats", admin(http.HandlerFunc(handler.mcpStats)))
	r.Method(http.MethodGet, "/api/mcp-services/{id}", admin(http.HandlerFunc(handler.getMCP)))
	r.Method(http.MethodPut, "/api/mcp-services/{id}", admin(http.HandlerFunc(handler.updateMCP)))
	r.Method(http.MethodDelete, "/api/mcp-services/{id}", admin(http.HandlerFunc(handler.deleteMCP)))
	r.Method(http.MethodPut, "/api/mcp-services/{id}/toggle", admin(http.HandlerFunc(handler.toggleMCP)))
	return nil
}

func decodeJSON[T any](r *http.Request, destination *T) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return servicecatalog.ValidationError("body", "invalid JSON")
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return servicecatalog.ValidationError("body", "must contain exactly one JSON value")
	}
	return nil
}

func parseAPIListOptions(r *http.Request) (servicecatalog.APIListOptions, error) {
	page, limit, err := parsePage(r)
	if err != nil {
		return servicecatalog.APIListOptions{}, err
	}
	return servicecatalog.APIListOptions{CategoryID: r.URL.Query().Get("category"), Search: r.URL.Query().Get("search"), Status: r.URL.Query().Get("status"), Page: page, Limit: limit}, nil
}

func parseMCPListOptions(r *http.Request) (servicecatalog.MCPListOptions, error) {
	page, limit, err := parsePage(r)
	if err != nil {
		return servicecatalog.MCPListOptions{}, err
	}
	return servicecatalog.MCPListOptions{Type: r.URL.Query().Get("type"), Search: r.URL.Query().Get("search"), Status: r.URL.Query().Get("status"), Page: page, Limit: limit}, nil
}

func parsePage(r *http.Request) (int, int, error) {
	page, err := positiveQueryInt(r, "page", 1)
	if err != nil {
		return 0, 0, err
	}
	limit, err := positiveQueryInt(r, "limit", 10)
	if err != nil {
		return 0, 0, err
	}
	if limit > 100 {
		return 0, 0, servicecatalog.ValidationError("limit", "must be at most 100")
	}
	return page, limit, nil
}

func positiveQueryInt(r *http.Request, name string, fallback int) (int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, servicecatalog.ValidationError(name, "must be a positive integer")
	}
	return value, nil
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	if err := appRuntime.WriteData(w, status, data); err != nil {
		return
	}
}

func writePaginated[T any](w http.ResponseWriter, data T, page, limit, total int) {
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	envelope := appRuntime.Envelope[T]{Success: true, Data: &data, Pagination: &appRuntime.Pagination{Page: page, PageSize: limit, Total: total, TotalPages: pages}}
	if err := appRuntime.WriteEnvelope(w, http.StatusOK, envelope); err != nil {
		return
	}
}

func writeCatalogError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		err = appRuntime.NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
	if writeErr := appRuntime.WriteError(w, r, err); writeErr != nil {
		return
	}
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
