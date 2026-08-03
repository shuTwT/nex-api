package utils

import (
	"net/http"

	"github.com/shuTwT/nex-api/internal/handler/httpkit"
)

// WriteData writes a standard successful API envelope.
func WriteData[T any](w http.ResponseWriter, status int, data T) {
	_ = httpkit.WriteData(w, status, data)
}

// WritePaginated writes a successful HTTP 200 API envelope with pagination metadata.
func WritePaginated[T any](w http.ResponseWriter, data T, page, limit, total int) {
	pages := 0
	if limit > 0 {
		pages = total / limit
		if total%limit != 0 {
			pages++
		}
	}
	envelope := httpkit.Envelope[T]{
		Success: true,
		Data:    &data,
		Pagination: &httpkit.Pagination{
			Page: page, PageSize: limit, Total: total, TotalPages: pages,
		},
	}
	_ = httpkit.WriteEnvelope(w, http.StatusOK, envelope)
}

// WriteError writes an error using the standard API error mapping and envelope.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	_ = httpkit.WriteError(w, r, err)
}
