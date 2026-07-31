package catalog

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/apicategory"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

type CategoryInput struct {
	Name        string
	Description string
	Icon        *string
}

type CategoryUpdateInput struct {
	Name        *string
	Description *string
	Icon        *string
}

type CategoryListItem struct {
	Category *ent.ApiCategory
	APICount int
}

type CategoryService struct {
	db *ent.Client
}

func NewCategoryService(db *ent.Client) (*CategoryService, error) {
	if db == nil {
		return nil, fmt.Errorf("catalog: database client is nil")
	}
	return &CategoryService{db: db}, nil
}

func (s *CategoryService) Create(ctx context.Context, input CategoryInput) (*ent.ApiCategory, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, validationError("name", "required")
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin category create: %w", err)
	}
	created, err := tx.ApiCategory.Create().SetName(input.Name).SetDescription(input.Description).SetNillableIcon(input.Icon).Save(ctx)
	if err != nil {
		return nil, abortTx(tx, classifyMutationError(err, "category name already exists"))
	}
	if err := writeAudit(ctx, tx, "创建 API 分类", "API 分类管理", fmt.Sprintf("创建了分类: %s", input.Name), "info", "success"); err != nil {
		return nil, abortTx(tx, fmt.Errorf("audit category create: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit category create: %w", err)
	}
	return created.Unwrap(), nil
}

func (s *CategoryService) Get(ctx context.Context, id string) (*ent.ApiCategory, error) {
	if strings.TrimSpace(id) == "" {
		return nil, validationError("id", "required")
	}
	category, err := s.db.ApiCategory.Query().Where(apicategory.ID(id)).WithApis().Only(ctx)
	if err != nil {
		return nil, classifyNotFound(err, "API category")
	}
	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, id string, input CategoryUpdateInput) (*ent.ApiCategory, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return nil, validationError("name", "must not be empty")
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin category update: %w", err)
	}
	builder := tx.ApiCategory.UpdateOneID(id)
	if input.Name != nil {
		builder.SetName(*input.Name)
	}
	if input.Description != nil {
		builder.SetDescription(*input.Description)
	}
	if input.Icon != nil {
		builder.SetIcon(*input.Icon)
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, abortTx(tx, classifyMutationError(err, "category name already exists"))
	}
	if err := writeAudit(ctx, tx, "更新 API 分类", "API 分类管理", fmt.Sprintf("更新了分类: %s", updated.Name), "info", "success"); err != nil {
		return nil, abortTx(tx, fmt.Errorf("audit category update: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit category update: %w", err)
	}
	return updated.Unwrap(), nil
}

func (s *CategoryService) Delete(ctx context.Context, id string) error {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin category delete: %w", err)
	}
	category, err := tx.ApiCategory.Query().Where(apicategory.ID(id)).Only(ctx)
	if err != nil {
		return abortTx(tx, classifyNotFound(err, "API category"))
	}
	apiCount, err := tx.ApiCategory.Query().Where(apicategory.ID(id)).QueryApis().Count(ctx)
	if err != nil {
		return abortTx(tx, fmt.Errorf("check category dependencies: %w", err))
	}
	if apiCount > 0 {
		return abortTx(tx, runtime.NewAPIError(http.StatusConflict, "dependent_records", fmt.Sprintf("cannot delete category %q: %d APIs still reference it", category.Name, apiCount), runtime.ErrConflict))
	}
	if err := tx.ApiCategory.DeleteOneID(id).Exec(ctx); err != nil {
		return abortTx(tx, classifyMutationError(err, "category has dependent APIs"))
	}
	if err := writeAudit(ctx, tx, "删除 API 分类", "API 分类管理", fmt.Sprintf("删除了分类: %s", category.Name), "warning", "success"); err != nil {
		return abortTx(tx, fmt.Errorf("audit category delete: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit category delete: %w", err)
	}
	return nil
}

func (s *CategoryService) List(ctx context.Context) ([]CategoryListItem, error) {
	categories, err := s.db.ApiCategory.Query().WithApis().Order(apicategory.ByName()).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	items := make([]CategoryListItem, len(categories))
	for index, category := range categories {
		items[index] = CategoryListItem{Category: category, APICount: len(category.Edges.Apis)}
	}
	return items, nil
}
