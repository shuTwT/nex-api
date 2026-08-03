package accounts

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/apiusage"
	"github.com/shuTwT/nex-api/ent/subscription"
	"github.com/shuTwT/nex-api/ent/user"
	"github.com/shuTwT/nex-api/internal/service/auth"
)

type UserService struct {
	client *ent.Client
	audit  *AuditService
	now    func() time.Time
}

func NewUserService(client *ent.Client, audit ...*AuditService) (*UserService, error) {
	if client == nil {
		return nil, errors.New("accounts: ent client is nil")
	}
	var recorder *AuditService
	if len(audit) > 0 {
		recorder = audit[0]
	}
	return &UserService{client: client, audit: recorder, now: time.Now}, nil
}

func (s *UserService) Create(ctx context.Context, request UserCreateRequest, metadata AuditMetadata) (UserView, error) {
	if err := validateUserCreate(request); err != nil {
		return UserView{}, err
	}
	role := request.Role
	if role == "" {
		role = "user"
	}
	credits := request.Credits
	if credits == 0 {
		credits = 1000
	}
	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		return UserView{}, fmt.Errorf("hash user password: %w", err)
	}
	existing, err := s.client.User.Query().Where(user.Or(user.EmailEQ(request.Email), user.UsernameEQ(request.Username))).Exist(ctx)
	if err != nil {
		return UserView{}, fmt.Errorf("check user uniqueness: %w", err)
	}
	if existing {
		return UserView{}, fmt.Errorf("user email or username: %w", ErrConflict)
	}
	created, err := s.client.User.Create().SetEmail(request.Email).SetUsername(request.Username).
		SetPassword(hash).SetRole(role).SetCredits(credits).Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return UserView{}, fmt.Errorf("create user: %w", ErrConflict)
		}
		return UserView{}, fmt.Errorf("create user: %w", err)
	}
	if err := s.record(ctx, created.ID, "user.create", "user", created.ID, metadata); err != nil {
		return UserView{}, err
	}
	return userView(created), nil
}

func (s *UserService) Get(ctx context.Context, id string) (UserView, error) {
	if strings.TrimSpace(id) == "" {
		return UserView{}, fmt.Errorf("get user: %w", ErrNotFound)
	}
	entity, err := s.client.User.Query().Where(user.IDEQ(id)).
		WithSubscriptions(func(q *ent.SubscriptionQuery) { q.Order(ent.Desc(subscription.FieldCreatedAt)).Limit(1) }).
		WithApiUsage(func(q *ent.ApiUsageQuery) { q.Order(ent.Desc(apiusage.FieldCreatedAt)).Limit(10).WithAPI() }).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return UserView{}, fmt.Errorf("get user: %w", ErrNotFound)
		}
		return UserView{}, fmt.Errorf("get user: %w", err)
	}
	return userView(entity), nil
}

func (s *UserService) Update(ctx context.Context, id string, request UserUpdateRequest, metadata AuditMetadata) (UserView, error) {
	if id == "" {
		return UserView{}, fmt.Errorf("update user: %w", ErrNotFound)
	}
	if request.Email != nil && strings.TrimSpace(*request.Email) == "" || request.Username != nil && strings.TrimSpace(*request.Username) == "" {
		return UserView{}, fmt.Errorf("update user: %w", ErrInvalidRequest)
	}
	query := s.client.User.UpdateOneID(id)
	if request.Email != nil {
		query.SetEmail(strings.TrimSpace(strings.ToLower(*request.Email)))
	}
	if request.Username != nil {
		query.SetUsername(strings.TrimSpace(*request.Username))
	}
	if request.Role != nil {
		if *request.Role != "user" && *request.Role != "admin" {
			return UserView{}, fmt.Errorf("update user role: %w", ErrInvalidRequest)
		}
		query.SetRole(*request.Role)
	}
	if request.Credits != nil {
		query.SetCredits(*request.Credits)
	}
	updated, err := query.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return UserView{}, fmt.Errorf("update user: %w", ErrNotFound)
		}
		if ent.IsConstraintError(err) {
			return UserView{}, fmt.Errorf("update user: %w", ErrConflict)
		}
		return UserView{}, fmt.Errorf("update user: %w", err)
	}
	if err := s.record(ctx, id, "user.update", "user", id, metadata); err != nil {
		return UserView{}, err
	}
	return s.Get(ctx, updated.ID)
}

func (s *UserService) Delete(ctx context.Context, id string, metadata AuditMetadata) error {
	if err := s.client.User.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("delete user: %w", ErrNotFound)
		}
		return fmt.Errorf("delete user: %w", err)
	}
	return s.record(ctx, id, "user.delete", "user", id, metadata)
}

func (s *UserService) List(ctx context.Context, filter UserListFilter, page PageRequest) ([]UserView, PageInfo, error) {
	page = page.normalized()
	query := s.client.User.Query()
	if filter.Role != "" && filter.Role != "all" {
		query.Where(user.RoleEQ(filter.Role))
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		query.Where(user.Or(user.UsernameContains(search), user.EmailContains(search)))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, PageInfo{}, fmt.Errorf("count users: %w", err)
	}
	entities, err := query.Order(ent.Desc(user.FieldCreatedAt)).Offset((page.Page - 1) * page.Size).Limit(page.Size).All(ctx)
	if err != nil {
		return nil, PageInfo{}, fmt.Errorf("list users: %w", err)
	}
	views := make([]UserView, len(entities))
	for i, entity := range entities {
		views[i] = userView(entity)
	}
	return views, pageInfo(page, total), nil
}

func (s *UserService) Stats(ctx context.Context) (UserStats, error) {
	now := s.now().UTC()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	activeSince := now.Add(-30 * 24 * time.Hour)
	total, err := s.client.User.Query().Count(ctx)
	if err != nil {
		return UserStats{}, fmt.Errorf("count users: %w", err)
	}
	active, err := s.client.User.Query().Where(user.HasApiUsageWith(apiusage.CreatedAtGTE(activeSince))).Count(ctx)
	if err != nil {
		return UserStats{}, fmt.Errorf("count active users: %w", err)
	}
	admins, err := s.client.User.Query().Where(user.RoleEQ("admin")).Count(ctx)
	if err != nil {
		return UserStats{}, fmt.Errorf("count admin users: %w", err)
	}
	newUsers, err := s.client.User.Query().Where(user.CreatedAtGTE(month)).Count(ctx)
	if err != nil {
		return UserStats{}, fmt.Errorf("count new users: %w", err)
	}
	return UserStats{TotalUsers: total, ActiveUsers: active, AdminUsers: admins, NewUsersThisMonth: newUsers}, nil
}

func (s *UserService) record(ctx context.Context, userID, action, resource, details string, metadata AuditMetadata) error {
	if s.audit == nil {
		return nil
	}
	if err := s.audit.Record(ctx, AuditEntry{UserID: userID, Action: action, Resource: resource, Details: details, IPAddress: metadata.IP, UserAgent: metadata.UserAgent, Metadata: metadata.Metadata}); err != nil {
		return fmt.Errorf("record %s audit: %w", action, err)
	}
	return nil
}

func validateUserCreate(request UserCreateRequest) error {
	if !strings.Contains(request.Email, "@") || strings.TrimSpace(request.Username) == "" || len(request.Username) < 3 || len(request.Username) > 20 || len(request.Password) < 8 || len(request.Password) > 100 {
		return fmt.Errorf("create user: %w", ErrInvalidRequest)
	}
	if request.Role != "" && request.Role != "user" && request.Role != "admin" {
		return fmt.Errorf("create user role: %w", ErrInvalidRequest)
	}
	return nil
}

func userView(entity *ent.User) UserView {
	view := UserView{ID: entity.ID, Name: entity.Name, Email: entity.Email, Username: entity.Username, Role: entity.Role, Credits: entity.Credits, CreatedAt: entity.CreatedAt, UpdatedAt: entity.UpdatedAt}
	if subscriptions := entity.Edges.Subscriptions; len(subscriptions) > 0 {
		item := subscriptions[0]
		view.Subscription = &SubscriptionView{ID: item.ID, PlanName: item.PlanName, Credits: item.Credits, Price: item.Price, StartDate: item.StartDate, EndDate: item.EndDate, IsActive: item.IsActive}
	}
	if entity.Edges.ApiUsage != nil {
		view.APIUsage = make([]UsageView, len(entity.Edges.ApiUsage))
		for i, item := range entity.Edges.ApiUsage {
			view.APIUsage[i] = UsageView{ID: item.ID, Credits: item.Credits, Status: item.Status, CreatedAt: item.CreatedAt}
			if item.Edges.API != nil {
				view.APIUsage[i].API = &APIView{Name: item.Edges.API.Name, Endpoint: item.Edges.API.Endpoint}
			}
		}
	}
	return view
}
