package catalog

import "github.com/shuTwT/nex-api/ent"

type MCPInput struct {
	Name       string
	Identifier string
	Type       string
	Command    *string
	Endpoint   *string
	EnvVars    *string
	Pricing    int
	IsActive   bool
}

type MCPUpdateInput struct {
	Name       *string
	Identifier *string
	Type       *string
	Command    *string
	Endpoint   *string
	EnvVars    *string
	Pricing    *int
	IsActive   *bool
}

type MCPListOptions struct {
	Type   string
	Search string
	Status string
	Page   int
	Limit  int
}

type MCPListResult struct {
	Items []*ent.McpService
	Total int
}

type MCPStats struct {
	TotalServices    int `json:"totalServices"`
	ActiveServices   int `json:"activeServices"`
	InactiveServices int `json:"inactiveServices"`
	TotalCalls       int `json:"totalCalls"`
}
