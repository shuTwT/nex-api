package ads

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/advertisement"
	"github.com/shuTwT/nex-api/backend/internal/service/apierror"
)

type Service struct{ db *ent.Client }

func NewService(db *ent.Client) (*Service, error) {
	if db == nil {
		return nil, errors.New("ads: database is required")
	}
	return &Service{db: db}, nil
}

type CreateInput struct {
	Image       string
	ImageWidth  int
	ImageHeight int
	Link        string
	Title       string
	Position    string
	IsActive    bool
}

type UpdateInput struct {
	Image       *string
	ImageWidth  *int
	ImageHeight *int
	Link        *string
	Title       *string
	Position    *string
	IsActive    *bool
}

type ListOptions struct {
	Search   string
	Position string
	IsActive *bool
	Page     int
	Limit    int
}

type ListResult struct {
	Items []*ent.Advertisement
	Total int
}

type PositionStat struct {
	Position string         `json:"position"`
	Count    map[string]int `json:"_count"`
}

type Stats struct {
	TotalAds      int            `json:"totalAds"`
	ActiveAds     int            `json:"activeAds"`
	InactiveAds   int            `json:"inactiveAds"`
	PositionStats []PositionStat `json:"positionStats"`
}

func (s *Service) List(ctx context.Context, options ListOptions) (ListResult, error) {
	options = normalizeListOptions(options)
	query := s.db.Advertisement.Query()
	if options.Search != "" {
		query = query.Where(advertisement.TitleContainsFold(options.Search))
	}
	if options.Position != "" {
		query = query.Where(advertisement.Position(options.Position))
	}
	if options.IsActive != nil {
		query = query.Where(advertisement.IsActive(*options.IsActive))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return ListResult{}, fmt.Errorf("count advertisements: %w", err)
	}
	items, err := query.Order(advertisement.ByCreatedAt(sql.OrderDesc())).Offset((options.Page - 1) * options.Limit).Limit(options.Limit).All(ctx)
	if err != nil {
		return ListResult{}, fmt.Errorf("list advertisements: %w", err)
	}
	return ListResult{Items: items, Total: total}, nil
}

func (s *Service) Get(ctx context.Context, id string) (*ent.Advertisement, error) {
	item, err := s.db.Advertisement.Query().Where(advertisement.ID(strings.TrimSpace(id))).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apierror.NewError(apierror.KindNotFound, "not_found", "Advertisement not found", apierror.ErrNotFound)
		}
		return nil, fmt.Errorf("get advertisement: %w", err)
	}
	return item, nil
}

func (s *Service) ByPosition(ctx context.Context, position string) (*ent.Advertisement, error) {
	item, err := s.db.Advertisement.Query().Where(advertisement.Position(strings.TrimSpace(position)), advertisement.IsActive(true)).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get advertisement by position: %w", err)
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*ent.Advertisement, error) {
	if err := validateCreate(input); err != nil {
		return nil, err
	}
	exists, err := s.db.Advertisement.Query().Where(advertisement.Position(input.Position)).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check advertisement position: %w", err)
	}
	if exists {
		return nil, apierror.NewError(apierror.KindConflict, "conflict", "该广告位已被占用", apierror.ErrConflict)
	}
	now := time.Now()
	item, err := s.db.Advertisement.Create().SetImage(input.Image).SetImageWidth(input.ImageWidth).SetImageHeight(input.ImageHeight).SetLink(input.Link).SetTitle(input.Title).SetPosition(input.Position).SetIsActive(input.IsActive).SetCreatedAt(now).SetUpdatedAt(now).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create advertisement: %w", err)
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (*ent.Advertisement, error) {
	if input.Position != nil {
		exists, err := s.db.Advertisement.Query().Where(advertisement.Position(*input.Position), advertisement.IDNEQ(id)).Exist(ctx)
		if err != nil {
			return nil, fmt.Errorf("check advertisement position: %w", err)
		}
		if exists {
			return nil, apierror.NewError(apierror.KindConflict, "conflict", "该广告位已被占用", apierror.ErrConflict)
		}
	}
	builder := s.db.Advertisement.UpdateOneID(id)
	if input.Image != nil {
		builder.SetImage(*input.Image)
	}
	if input.ImageWidth != nil {
		builder.SetImageWidth(*input.ImageWidth)
	}
	if input.ImageHeight != nil {
		builder.SetImageHeight(*input.ImageHeight)
	}
	if input.Link != nil {
		builder.SetLink(*input.Link)
	}
	if input.Title != nil {
		builder.SetTitle(*input.Title)
	}
	if input.Position != nil {
		builder.SetPosition(*input.Position)
	}
	if input.IsActive != nil {
		builder.SetIsActive(*input.IsActive)
	}
	item, err := builder.SetUpdatedAt(time.Now()).Save(ctx)
	if ent.IsNotFound(err) {
		return nil, apierror.NewError(apierror.KindNotFound, "not_found", "Advertisement not found", apierror.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("update advertisement: %w", err)
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.db.Advertisement.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return apierror.NewError(apierror.KindNotFound, "not_found", "Advertisement not found", apierror.ErrNotFound)
		}
		return fmt.Errorf("delete advertisement: %w", err)
	}
	return nil
}

func (s *Service) Toggle(ctx context.Context, id string) (*ent.Advertisement, error) {
	item, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.Update(ctx, id, UpdateInput{IsActive: boolPointer(!item.IsActive)})
}

func (s *Service) Stats(ctx context.Context) (Stats, error) {
	items, err := s.db.Advertisement.Query().All(ctx)
	if err != nil {
		return Stats{}, fmt.Errorf("list advertisement stats: %w", err)
	}
	result := Stats{TotalAds: len(items), PositionStats: make([]PositionStat, 0)}
	counts := make(map[string]int)
	for _, item := range items {
		counts[item.Position]++
		if item.IsActive {
			result.ActiveAds++
		}
	}
	result.InactiveAds = result.TotalAds - result.ActiveAds
	for position, count := range counts {
		result.PositionStats = append(result.PositionStats, PositionStat{Position: position, Count: map[string]int{"position": count}})
	}
	return result, nil
}

func normalizeListOptions(options ListOptions) ListOptions {
	if options.Page < 1 {
		options.Page = 1
	}
	if options.Limit < 1 {
		options.Limit = 10
	}
	if options.Limit > 100 {
		options.Limit = 100
	}
	return options
}

func validateCreate(input CreateInput) error {
	fields := make([]apierror.FieldError, 0, 4)
	if strings.TrimSpace(input.Image) == "" {
		fields = append(fields, apierror.FieldError{Field: "image", Reason: "required"})
	}
	if strings.TrimSpace(input.Link) == "" {
		fields = append(fields, apierror.FieldError{Field: "link", Reason: "required"})
	}
	if strings.TrimSpace(input.Title) == "" {
		fields = append(fields, apierror.FieldError{Field: "title", Reason: "required"})
	}
	if strings.TrimSpace(input.Position) == "" {
		fields = append(fields, apierror.FieldError{Field: "position", Reason: "required"})
	}
	if len(fields) > 0 {
		return apierror.NewValidationError(fields...)
	}
	return nil
}

func boolPointer(value bool) *bool { return &value }
