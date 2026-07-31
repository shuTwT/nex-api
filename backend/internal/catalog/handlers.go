package catalog

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

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
	mux.HandleFunc("GET /api/apis", handler.listAPIs)
	mux.HandleFunc("POST /api/apis", handler.createAPI)
	mux.HandleFunc("GET /api/apis/stats", handler.apiStats)
	mux.HandleFunc("GET /api/apis/{id}", handler.getAPI)
	mux.HandleFunc("PUT /api/apis/{id}", handler.updateAPI)
	mux.HandleFunc("DELETE /api/apis/{id}", handler.deleteAPI)
	mux.HandleFunc("PUT /api/apis/{id}/toggle", handler.toggleAPI)
	mux.HandleFunc("GET /api/categories", handler.listCategories)
	mux.HandleFunc("POST /api/categories", handler.createCategory)
	mux.HandleFunc("GET /api/categories/{id}", handler.getCategory)
	mux.HandleFunc("PUT /api/categories/{id}", handler.updateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", handler.deleteCategory)
	mux.HandleFunc("GET /api/mcp-services", handler.listMCP)
	mux.HandleFunc("POST /api/mcp-services", handler.createMCP)
	mux.HandleFunc("GET /api/mcp-services/stats", handler.mcpStats)
	mux.HandleFunc("GET /api/mcp-services/{id}", handler.getMCP)
	mux.HandleFunc("PUT /api/mcp-services/{id}", handler.updateMCP)
	mux.HandleFunc("DELETE /api/mcp-services/{id}", handler.deleteMCP)
	mux.HandleFunc("PUT /api/mcp-services/{id}/toggle", handler.toggleMCP)
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
