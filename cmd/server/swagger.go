// Package main carries the global swagger metadata and the shared envelope
// types. Per-endpoint annotations live directly on the handler functions in
// internal/handler/ so the generated spec cannot drift from the real router;
// regenerate with `make generate`.
package main

// @title Nex API
// @version 1.0.0
// @description Contract for the Nex HTTP API.
// @license.name MIT
// @license.identifier MIT
// @license.url https://opensource.org/license/mit
// @BasePath /
// @securityDefinitions.apikey SessionCookie
// @in cookie
// @name session
// @securityDefinitions.apikey ApiTokenAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey CronSecretAuth
// @in header
// @name Authorization
type SwaggerDocument struct{}

// SwaggerEnvelope is the standard JSON response envelope returned by the API.
type SwaggerEnvelope struct {
	Success    bool               `json:"success"`
	Data       any                `json:"data,omitempty"`
	Error      string             `json:"error,omitempty"`
	Pagination *SwaggerPagination `json:"pagination,omitempty"`
}

// SwaggerPagination is the pagination metadata carried by list responses.
type SwaggerPagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// SwaggerRequest represents a JSON request payload. Endpoint handlers validate its fields.
type SwaggerRequest map[string]any

// Endpoint annotations live directly on the handler functions in internal/handler/,
// colocated with the route registrations they document, so the generated spec cannot
// drift from the real router. Regenerate with `make generate`.
