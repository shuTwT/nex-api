package catalog

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

type Handler struct {
	apis       *APIService
	categories *CategoryService
	mcp        *MCPService
}

func NewHandler(apis *APIService, categories *CategoryService, mcp *MCPService) (*Handler, error) {
	if apis == nil || categories == nil || mcp == nil {
		return nil, errors.New("catalog: all services are required")
	}
	return &Handler{apis: apis, categories: categories, mcp: mcp}, nil
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil {
		return errors.New("catalog: route mux is nil")
	}
	if handler == nil {
		return errors.New("catalog: handler is nil")
	}
	admin := func(next http.Handler) http.Handler { return authz.RequireAdmin(next) }
	mux.Handle("GET /api/apis", admin(http.HandlerFunc(handler.listAPIs)))
	mux.Handle("POST /api/apis", admin(http.HandlerFunc(handler.createAPI)))
	mux.Handle("GET /api/apis/stats", admin(http.HandlerFunc(handler.apiStats)))
	mux.Handle("GET /api/apis/{id}", admin(http.HandlerFunc(handler.getAPI)))
	mux.Handle("PUT /api/apis/{id}", admin(http.HandlerFunc(handler.updateAPI)))
	mux.Handle("DELETE /api/apis/{id}", admin(http.HandlerFunc(handler.deleteAPI)))
	mux.Handle("PUT /api/apis/{id}/toggle", admin(http.HandlerFunc(handler.toggleAPI)))
	mux.Handle("GET /api/categories", admin(http.HandlerFunc(handler.listCategories)))
	mux.Handle("POST /api/categories", admin(http.HandlerFunc(handler.createCategory)))
	mux.Handle("GET /api/categories/{id}", admin(http.HandlerFunc(handler.getCategory)))
	mux.Handle("PUT /api/categories/{id}", admin(http.HandlerFunc(handler.updateCategory)))
	mux.Handle("DELETE /api/categories/{id}", admin(http.HandlerFunc(handler.deleteCategory)))
	mux.Handle("GET /api/mcp-services", admin(http.HandlerFunc(handler.listMCP)))
	mux.Handle("POST /api/mcp-services", admin(http.HandlerFunc(handler.createMCP)))
	mux.Handle("GET /api/mcp-services/stats", admin(http.HandlerFunc(handler.mcpStats)))
	mux.Handle("GET /api/mcp-services/{id}", admin(http.HandlerFunc(handler.getMCP)))
	mux.Handle("PUT /api/mcp-services/{id}", admin(http.HandlerFunc(handler.updateMCP)))
	mux.Handle("DELETE /api/mcp-services/{id}", admin(http.HandlerFunc(handler.deleteMCP)))
	mux.Handle("PUT /api/mcp-services/{id}/toggle", admin(http.HandlerFunc(handler.toggleMCP)))
	return nil
}

func decodeJSON[T any](r *http.Request, destination *T) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return validationError("body", "invalid JSON")
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return validationError("body", "must contain exactly one JSON value")
	}
	return nil
}

func parseAPIListOptions(r *http.Request) (APIListOptions, error) {
	page, limit, err := parsePage(r)
	if err != nil {
		return APIListOptions{}, err
	}
	return APIListOptions{CategoryID: r.URL.Query().Get("category"), Search: r.URL.Query().Get("search"), Status: r.URL.Query().Get("status"), Page: page, Limit: limit}, nil
}

func parseMCPListOptions(r *http.Request) (MCPListOptions, error) {
	page, limit, err := parsePage(r)
	if err != nil {
		return MCPListOptions{}, err
	}
	return MCPListOptions{Type: r.URL.Query().Get("type"), Search: r.URL.Query().Get("search"), Status: r.URL.Query().Get("status"), Page: page, Limit: limit}, nil
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
		return 0, 0, validationError("limit", "must be at most 100")
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
		return 0, validationError(name, "must be a positive integer")
	}
	return value, nil
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	if err := runtime.WriteData(w, status, data); err != nil {
		return
	}
}

func writePaginated[T any](w http.ResponseWriter, data T, page, limit, total int) {
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	envelope := runtime.Envelope[T]{Success: true, Data: &data, Pagination: &runtime.Pagination{Page: page, PageSize: limit, Total: total, TotalPages: pages}}
	if err := runtime.WriteEnvelope(w, http.StatusOK, envelope); err != nil {
		return
	}
}

func writeCatalogError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		err = runtime.NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
	if writeErr := runtime.WriteError(w, r, err); writeErr != nil {
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
