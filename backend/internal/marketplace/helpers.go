package marketplace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
	"github.com/shuTwT/nex-api/backend/internal/stats"
)

type pageOptions struct {
	page     int
	limit    int
	search   string
	category string
}

type apiView struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Alias          string `json:"alias"`
	Endpoint       string `json:"endpoint"`
	Method         string `json:"method"`
	Pricing        int    `json:"pricing"`
	Category       string `json:"category"`
	IsFree         bool   `json:"isFree"`
	IsActive       bool   `json:"isActive,omitempty"`
	TodayCallCount int64  `json:"todayCallCount"`
	UserCount      int    `json:"userCount"`
	TotalCallCount int64  `json:"totalCallCount"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

type mcpView struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Identifier     string `json:"identifier"`
	Type           string `json:"type"`
	Pricing        int    `json:"pricing"`
	IsFree         bool   `json:"isFree"`
	IsActive       bool   `json:"isActive,omitempty"`
	TodayCallCount int64  `json:"todayCallCount"`
	UserCount      int    `json:"userCount"`
	TotalCallCount int64  `json:"totalCallCount"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

func parsePage(r *http.Request, defaultLimit, maxLimit int) (pageOptions, error) {
	query := r.URL.Query()
	page, err := positiveInt(query.Get("page"), 1)
	if err != nil {
		return pageOptions{}, runtime.NewValidationError(runtime.FieldError{Field: "page", Reason: "must be a positive integer"})
	}
	limit, err := positiveInt(query.Get("limit"), defaultLimit)
	if err != nil || limit > maxLimit {
		return pageOptions{}, runtime.NewValidationError(runtime.FieldError{Field: "limit", Reason: fmt.Sprintf("must be between 1 and %d", maxLimit)})
	}
	return pageOptions{page: page, limit: limit, search: strings.TrimSpace(query.Get("search")), category: strings.TrimSpace(query.Get("category"))}, nil
}

func positiveInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, errors.New("invalid positive integer")
	}
	return value, nil
}

func (h *Handler) toAPIView(ctx context.Context, item *ent.Api, snapshot stats.Snapshot, detail bool) apiView {
	trend, _ := h.stats.APIRequestTrend(ctx, item.Alias, 24)
	return apiView{ID: item.ID, Name: item.Name, Description: item.Description, Alias: item.Alias, Endpoint: item.Endpoint, Method: item.Method, Pricing: item.Pricing, Category: item.Edges.Category.Name, IsFree: item.Pricing == 0, IsActive: detail && item.IsActive, TodayCallCount: sumInt64(trend), UserCount: countAPIUsers(snapshot.UserAPIs, item.Alias), TotalCallCount: canonicalOrDatabase(snapshot.APIs, item.Alias, int64(item.CallCount)), CreatedAt: item.CreatedAt.UTC().Format(timeFormat), UpdatedAt: item.UpdatedAt.UTC().Format(timeFormat)}
}

func (h *Handler) toMCPView(ctx context.Context, item *ent.McpService, snapshot stats.Snapshot, detail bool) mcpView {
	trend, _ := h.stats.MCPRequestTrend(ctx, item.Identifier, 24)
	return mcpView{ID: item.ID, Name: item.Name, Identifier: item.Identifier, Type: item.Type, Pricing: item.Pricing, IsFree: item.Pricing == 0, IsActive: detail && item.IsActive, TodayCallCount: sumInt64(trend), UserCount: countMCPUsers(snapshot.UserMCPs, item.Identifier), TotalCallCount: canonicalOrDatabase(snapshot.MCPs, item.Identifier, int64(item.CallCount)), CreatedAt: item.CreatedAt.UTC().Format(timeFormat), UpdatedAt: item.UpdatedAt.UTC().Format(timeFormat)}
}

const timeFormat = "2006-01-02T15:04:05.000Z"

func canonicalOrDatabase(values map[string]int64, key string, fallback int64) int64 {
	value, ok := values[key]
	if ok {
		return value
	}
	return fallback
}

func countAPIUsers(values map[stats.UserAPIKey]int64, alias string) int {
	count := 0
	for key, value := range values {
		if key.Alias == alias && value > 0 {
			count++
		}
	}
	return count
}

func countMCPUsers(values map[stats.UserMCPKey]int64, identifier string) int {
	count := 0
	for key, value := range values {
		if key.Identifier == identifier && value > 0 {
			count++
		}
	}
	return count
}

func sumInt64(values []int64) int64 {
	var total int64
	for _, value := range values {
		total += value
	}
	return total
}

func decodeJSON[T any](r *http.Request, destination *T) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return runtime.NewValidationError(runtime.FieldError{Field: "body", Reason: "invalid JSON"})
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return runtime.NewValidationError(runtime.FieldError{Field: "body", Reason: "must contain exactly one JSON value"})
	}
	return nil
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	_ = runtime.WriteData(w, status, data)
}

func writePaginated[T any](w http.ResponseWriter, data T, page, limit, total int) {
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	envelope := runtime.Envelope[T]{Success: true, Data: &data, Pagination: &runtime.Pagination{Page: page, PageSize: limit, Total: total, TotalPages: pages}}
	_ = runtime.WriteEnvelope(w, http.StatusOK, envelope)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	_ = runtime.WriteError(w, r, err)
}
