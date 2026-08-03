package catalog

import (
	"context"
	"strings"

	"github.com/shuTwT/nex-api/ent/api"
	"github.com/shuTwT/nex-api/ent/mcpservice"
)

// GatewayAPI is the minimal API definition required by the request gateway.
// It prevents HTTP handlers from receiving Ent entities.
type GatewayAPI struct {
	ID         string
	Name       string
	Endpoint   string
	Method     string
	Pricing    int
	PreScript  string
	PostScript string
	IsActive   bool
}

func (s *APIService) GatewayAPI(ctx context.Context, identifier string) (GatewayAPI, error) {
	if strings.TrimSpace(identifier) == "" {
		return GatewayAPI{}, ValidationError("id", "required")
	}
	entity, err := s.db.Api.Query().Where(api.Or(api.ID(identifier), api.Alias(identifier))).Only(ctx)
	if err != nil {
		return GatewayAPI{}, ClassifyNotFound(err, "API")
	}
	return GatewayAPI{ID: entity.ID, Name: entity.Name, Endpoint: entity.Endpoint, Method: entity.Method, Pricing: entity.Pricing, PreScript: entity.PreScript, PostScript: entity.PostScript, IsActive: entity.IsActive}, nil
}

// GatewayMCPService is the minimum configuration required to invoke an MCP
// target without exposing its persistence entity outside the service layer.
type GatewayMCPService struct {
	ID         string
	Identifier string
	Type       string
	Endpoint   string
	Command    string
	EnvVars    string
	Pricing    int
	IsActive   bool
}

func (s *MCPService) GatewayMCPService(ctx context.Context, identifier string) (GatewayMCPService, error) {
	if strings.TrimSpace(identifier) == "" {
		return GatewayMCPService{}, ValidationError("id", "required")
	}
	entity, err := s.db.McpService.Query().Where(mcpservice.Or(mcpservice.ID(identifier), mcpservice.Identifier(identifier))).Only(ctx)
	if err != nil {
		return GatewayMCPService{}, ClassifyNotFound(err, "MCP service")
	}
	return GatewayMCPService{ID: entity.ID, Identifier: entity.Identifier, Type: entity.Type, Endpoint: entity.Endpoint, Command: entity.Command, EnvVars: entity.EnvVars, Pricing: entity.Pricing, IsActive: entity.IsActive}, nil
}
