package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/api"
	"github.com/shuTwT/nex-api/ent/apicategory"
	"github.com/shuTwT/nex-api/internal/service/apierror"
)

type APIInput struct {
	Name          string
	Alias         string
	Description   string
	Endpoint      string
	Method        string
	CategoryID    string
	Pricing       int
	Documentation *string
	PreScript     *string
	PostScript    *string
	IsActive      bool
}

type APIUpdateInput struct {
	Name          *string
	Alias         *string
	Description   *string
	Endpoint      *string
	Method        *string
	CategoryID    *string
	Pricing       *int
	Documentation *string
	PreScript     *string
	PostScript    *string
	IsActive      *bool
}

type APIListOptions struct {
	CategoryID string
	Search     string
	Status     string
	Page       int
	Limit      int
}

type APIListResult struct {
	Items []*ent.Api
	Total int
}

type APIStats struct {
	TotalAPIs       int `json:"totalApis"`
	ActiveAPIs      int `json:"activeApis"`
	InactiveAPIs    int `json:"inactiveApis"`
	TotalCalls      int `json:"totalCalls"`
	CategoriesCount int `json:"categoriesCount"`
}

type APIService struct {
	db *ent.Client
}

func NewAPIService(db *ent.Client) (*APIService, error) {
	if db == nil {
		return nil, errors.New("catalog: database client is nil")
	}
	return &APIService{db: db}, nil
}

func (s *APIService) Create(ctx context.Context, input APIInput) (*ent.Api, error) {
	if err := validateAPIInput(input); err != nil {
		return nil, err
	}
	exists, err := s.db.ApiCategory.Query().Where(apicategory.ID(input.CategoryID)).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check API category: %w", err)
	}
	if !exists {
		return nil, apierror.NewError(apierror.KindNotFound, "not_found", "API category not found", apierror.ErrNotFound)
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin API create: %w", err)
	}
	created, err := tx.Api.Create().
		SetName(input.Name).
		SetAlias(input.Alias).
		SetDescription(input.Description).
		SetEndpoint(input.Endpoint).
		SetMethod(strings.ToUpper(strings.TrimSpace(input.Method))).
		SetCategoryId(input.CategoryID).
		SetPricing(input.Pricing).
		SetIsActive(input.IsActive).
		SetUpdatedAt(time.Now()).
		SetNillableDocumentation(input.Documentation).
		SetNillablePreScript(input.PreScript).
		SetNillablePostScript(input.PostScript).
		Save(ctx)
	if err != nil {
		return nil, abortTx(tx, classifyMutationError(err, "API alias or endpoint already exists"))
	}
	if err := writeAudit(ctx, tx, "创建 API", "API 管理", fmt.Sprintf("创建了 API: %s (%s)", input.Name, input.Alias), "info", "success"); err != nil {
		return nil, abortTx(tx, fmt.Errorf("audit API create: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit API create: %w", err)
	}
	return created.Unwrap(), nil
}

func (s *APIService) Get(ctx context.Context, id string) (*ent.Api, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ValidationError("id", "required")
	}
	item, err := s.db.Api.Query().Where(api.ID(id)).WithCategory().WithParameters().WithResponses().Only(ctx)
	if err != nil {
		return nil, ClassifyNotFound(err, "API")
	}
	return item, nil
}

func (s *APIService) GetByIdentifier(ctx context.Context, identifier string) (*ent.Api, error) {
	if strings.TrimSpace(identifier) == "" {
		return nil, ValidationError("id", "required")
	}
	item, err := s.db.Api.Query().Where(api.Or(api.ID(identifier), api.Alias(identifier))).WithCategory().WithParameters().WithResponses().Only(ctx)
	if err != nil {
		return nil, ClassifyNotFound(err, "API")
	}
	return item, nil
}

func (s *APIService) Update(ctx context.Context, id string, input APIUpdateInput) (*ent.Api, error) {
	if err := validateAPIUpdate(input); err != nil {
		return nil, err
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin API update: %w", err)
	}
	builder := tx.Api.UpdateOneID(id)
	applyAPIUpdate(builder, input)
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, abortTx(tx, classifyMutationError(err, "API alias or endpoint already exists"))
	}
	if err := writeAudit(ctx, tx, "更新 API", "API 管理", fmt.Sprintf("更新了 API: %s", updated.Name), "info", "success"); err != nil {
		return nil, abortTx(tx, fmt.Errorf("audit API update: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit API update: %w", err)
	}
	return updated.Unwrap(), nil
}

func (s *APIService) Delete(ctx context.Context, id string) error {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin API delete: %w", err)
	}
	item, err := tx.Api.Query().Where(api.ID(id)).Only(ctx)
	if err != nil {
		return abortTx(tx, ClassifyNotFound(err, "API"))
	}
	dependent, err := dependentAPICounts(ctx, tx, id)
	if err != nil {
		return abortTx(tx, fmt.Errorf("check API dependencies: %w", err))
	}
	if dependent.total() > 0 {
		return abortTx(tx, dependencyConflict("API", item.Name, dependent.describe()))
	}
	if err := tx.Api.DeleteOneID(id).Exec(ctx); err != nil {
		return abortTx(tx, classifyMutationError(err, "API has dependent records"))
	}
	if err := writeAudit(ctx, tx, "删除 API", "API 管理", fmt.Sprintf("删除了 API: %s", item.Name), "warning", "success"); err != nil {
		return abortTx(tx, fmt.Errorf("audit API delete: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit API delete: %w", err)
	}
	return nil
}

func (s *APIService) List(ctx context.Context, options APIListOptions) (APIListResult, error) {
	options = normalizeListOptions(options)
	query := s.db.Api.Query()
	if options.CategoryID != "" && options.CategoryID != "all" {
		query = query.Where(api.CategoryId(options.CategoryID))
	}
	if options.Search != "" {
		query = query.Where(api.Or(api.NameContainsFold(options.Search), api.DescriptionContainsFold(options.Search), api.EndpointContainsFold(options.Search)))
	}
	switch options.Status {
	case "active":
		query = query.Where(api.IsActive(true))
	case "inactive":
		query = query.Where(api.IsActive(false))
	case "":
	default:
		return APIListResult{}, ValidationError("status", "must be active or inactive")
	}
	total, err := query.Count(ctx)
	if err != nil {
		return APIListResult{}, fmt.Errorf("count APIs: %w", err)
	}
	items, err := query.WithCategory().Order(api.ByCreatedAt(sql.OrderDesc())).Offset((options.Page - 1) * options.Limit).Limit(options.Limit).All(ctx)
	if err != nil {
		return APIListResult{}, fmt.Errorf("list APIs: %w", err)
	}
	return APIListResult{Items: items, Total: total}, nil
}

func (s *APIService) Toggle(ctx context.Context, id string) (*ent.Api, error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin API toggle: %w", err)
	}
	item, err := tx.Api.Query().Where(api.ID(id)).Only(ctx)
	if err != nil {
		return nil, abortTx(tx, ClassifyNotFound(err, "API"))
	}
	updated, err := tx.Api.UpdateOneID(id).SetIsActive(!item.IsActive).Save(ctx)
	if err != nil {
		return nil, abortTx(tx, fmt.Errorf("toggle API: %w", err))
	}
	if err := writeAudit(ctx, tx, "切换 API 状态", "API 管理", fmt.Sprintf("切换了 API: %s", item.Name), "info", "success"); err != nil {
		return nil, abortTx(tx, fmt.Errorf("audit API toggle: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit API toggle: %w", err)
	}
	return updated.Unwrap(), nil
}
