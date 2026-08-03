package marketplace

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
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
		return pageOptions{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "page", Reason: "must be a positive integer"})
	}
	limit, err := positiveInt(query.Get("limit"), defaultLimit)
	if err != nil || limit > maxLimit {
		return pageOptions{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "limit", Reason: fmt.Sprintf("must be between 1 and %d", maxLimit)})
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

func writeData[T any](w http.ResponseWriter, status int, data T) {
	_ = appRuntime.WriteData(w, status, data)
}

func writePaginated[T any](w http.ResponseWriter, data T, page, limit, total int) {
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	envelope := appRuntime.Envelope[T]{Success: true, Data: &data, Pagination: &appRuntime.Pagination{Page: page, PageSize: limit, Total: total, TotalPages: pages}}
	_ = appRuntime.WriteEnvelope(w, http.StatusOK, envelope)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	_ = appRuntime.WriteError(w, r, err)
}
