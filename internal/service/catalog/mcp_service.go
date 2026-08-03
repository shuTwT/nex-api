package catalog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/mcpservice"
	"github.com/shuTwT/nex-api/internal/service/apierror"
)

type MCPService struct {
	db *ent.Client
}

type McpService = MCPService

func NewMCPService(db *ent.Client) (*MCPService, error) {
	if db == nil {
		return nil, fmt.Errorf("catalog: database client is nil")
	}
	return &MCPService{db: db}, nil
}

func NewMcpService(db *ent.Client) (*MCPService, error) { return NewMCPService(db) }

func (s *MCPService) Create(ctx context.Context, input MCPInput) (*ent.McpService, error) {
	if err := validateMCPInput(input); err != nil {
		return nil, err
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin MCP create: %w", err)
	}
	builder := tx.McpService.Create().SetName(input.Name).SetIdentifier(input.Identifier).SetType(input.Type).SetPricing(input.Pricing).SetIsActive(input.IsActive).SetUpdatedAt(time.Now()).SetNillableEnvVars(input.EnvVars)
	if input.Type == "stdio" {
		builder.SetNillableCommand(input.Command)
	} else {
		builder.SetNillableEndpoint(input.Endpoint)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, abortTx(tx, classifyMutationError(err, "MCP identifier already exists"))
	}
	if err := writeAudit(ctx, tx, "创建 MCP 服务", "MCP 服务管理", fmt.Sprintf("创建了 MCP 服务: %s (%s)", input.Name, input.Identifier), "info", "success"); err != nil {
		return nil, abortTx(tx, fmt.Errorf("audit MCP create: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit MCP create: %w", err)
	}
	return created.Unwrap(), nil
}

func (s *MCPService) Get(ctx context.Context, id string) (*ent.McpService, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ValidationError("id", "required")
	}
	item, err := s.db.McpService.Query().Where(mcpservice.ID(id)).Only(ctx)
	if err != nil {
		return nil, ClassifyNotFound(err, "MCP service")
	}
	return item, nil
}

func (s *MCPService) GetByIdentifier(ctx context.Context, identifier string) (*ent.McpService, error) {
	if strings.TrimSpace(identifier) == "" {
		return nil, ValidationError("id", "required")
	}
	item, err := s.db.McpService.Query().Where(mcpservice.Or(mcpservice.ID(identifier), mcpservice.Identifier(identifier))).Only(ctx)
	if err != nil {
		return nil, ClassifyNotFound(err, "MCP service")
	}
	return item, nil
}

func (s *MCPService) Update(ctx context.Context, id string, input MCPUpdateInput) (*ent.McpService, error) {
	if err := validateMCPUpdate(input); err != nil {
		return nil, err
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin MCP update: %w", err)
	}
	builder := tx.McpService.UpdateOneID(id)
	if input.Name != nil {
		builder.SetName(*input.Name)
	}
	if input.Identifier != nil {
		builder.SetIdentifier(*input.Identifier)
	}
	if input.Type != nil {
		builder.SetType(*input.Type)
		if *input.Type == "stdio" {
			builder.ClearEndpoint()
		} else {
			builder.ClearCommand()
		}
	}
	if input.Command != nil {
		builder.SetCommand(*input.Command)
	}
	if input.Endpoint != nil {
		builder.SetEndpoint(*input.Endpoint)
	}
	if input.EnvVars != nil {
		builder.SetEnvVars(*input.EnvVars)
	}
	if input.Pricing != nil {
		builder.SetPricing(*input.Pricing)
	}
	if input.IsActive != nil {
		builder.SetIsActive(*input.IsActive)
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, abortTx(tx, classifyMutationError(err, "MCP identifier already exists"))
	}
	if err := writeAudit(ctx, tx, "更新 MCP 服务", "MCP 服务管理", fmt.Sprintf("更新了 MCP 服务: %s (%s)", updated.Name, updated.Identifier), "info", "success"); err != nil {
		return nil, abortTx(tx, fmt.Errorf("audit MCP update: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit MCP update: %w", err)
	}
	return updated.Unwrap(), nil
}

func (s *MCPService) Delete(ctx context.Context, id string) error {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin MCP delete: %w", err)
	}
	item, err := tx.McpService.Query().Where(mcpservice.ID(id)).Only(ctx)
	if err != nil {
		return abortTx(tx, ClassifyNotFound(err, "MCP service"))
	}
	usageCount, err := tx.McpService.Query().Where(mcpservice.ID(id)).QueryUsageRecords().Count(ctx)
	if err != nil {
		return abortTx(tx, fmt.Errorf("check MCP dependencies: %w", err))
	}
	if usageCount > 0 {
		return abortTx(tx, apierror.NewError(apierror.KindConflict, "dependent_records", fmt.Sprintf("cannot delete MCP service %q: %d usage records still reference it", item.Name, usageCount), apierror.ErrConflict))
	}
	if err := tx.McpService.DeleteOneID(id).Exec(ctx); err != nil {
		return abortTx(tx, classifyMutationError(err, "MCP service has dependent usage records"))
	}
	if err := writeAudit(ctx, tx, "删除 MCP 服务", "MCP 服务管理", fmt.Sprintf("删除了 MCP 服务: %s", item.Name), "warning", "success"); err != nil {
		return abortTx(tx, fmt.Errorf("audit MCP delete: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit MCP delete: %w", err)
	}
	return nil
}

func (s *MCPService) List(ctx context.Context, options MCPListOptions) (MCPListResult, error) {
	options = normalizeMCPListOptions(options)
	query := s.db.McpService.Query()
	if options.Type != "" && options.Type != "all" {
		if _, ok := validMCPTypes[options.Type]; !ok {
			return MCPListResult{}, ValidationError("type", "must be stdio, sse, or streamableHttp")
		}
		query = query.Where(mcpservice.Type(options.Type))
	}
	if options.Search != "" {
		query = query.Where(mcpservice.Or(mcpservice.NameContainsFold(options.Search), mcpservice.IdentifierContainsFold(options.Search)))
	}
	switch options.Status {
	case "active":
		query = query.Where(mcpservice.IsActive(true))
	case "inactive":
		query = query.Where(mcpservice.IsActive(false))
	case "":
	default:
		return MCPListResult{}, ValidationError("status", "must be active or inactive")
	}
	total, err := query.Count(ctx)
	if err != nil {
		return MCPListResult{}, fmt.Errorf("count MCP services: %w", err)
	}
	items, err := query.Order(mcpservice.ByCreatedAt(sql.OrderDesc())).Offset((options.Page - 1) * options.Limit).Limit(options.Limit).All(ctx)
	if err != nil {
		return MCPListResult{}, fmt.Errorf("list MCP services: %w", err)
	}
	return MCPListResult{Items: items, Total: total}, nil
}

func (s *MCPService) Toggle(ctx context.Context, id string) (*ent.McpService, error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin MCP toggle: %w", err)
	}
	item, err := tx.McpService.Query().Where(mcpservice.ID(id)).Only(ctx)
	if err != nil {
		return nil, abortTx(tx, ClassifyNotFound(err, "MCP service"))
	}
	updated, err := tx.McpService.UpdateOneID(id).SetIsActive(!item.IsActive).Save(ctx)
	if err != nil {
		return nil, abortTx(tx, fmt.Errorf("toggle MCP service: %w", err))
	}
	if err := writeAudit(ctx, tx, "切换 MCP 服务状态", "MCP 服务管理", fmt.Sprintf("切换了 MCP 服务: %s", item.Name), "info", "success"); err != nil {
		return nil, abortTx(tx, fmt.Errorf("audit MCP toggle: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit MCP toggle: %w", err)
	}
	return updated.Unwrap(), nil
}
