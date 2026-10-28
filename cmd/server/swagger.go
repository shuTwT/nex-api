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
type SwaggerDocument struct{}

// SwaggerEnvelope is the standard JSON response envelope returned by the API.
// Code 0 means success; any other value is an entry of the error code table in
// pkg/constant (HTTP status * 100 + sequence). List endpoints put their items
// and paging metadata inside data (see SwaggerPage).
type SwaggerEnvelope struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// SwaggerPage is the data payload of a paginated response.
type SwaggerPage struct {
	List       any `json:"list"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}

// SwaggerRequest represents a JSON request payload. Endpoint handlers validate its fields.
type SwaggerRequest map[string]any

// Endpoint annotations live directly on the handler functions in internal/handler/,
// colocated with the route registrations they document, so the generated spec cannot
// drift from the real router. Regenerate with `make generate`.
