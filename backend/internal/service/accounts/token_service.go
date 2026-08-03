package accounts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/apitoken"
	"github.com/shuTwT/nex-api/backend/pkg/domain/model"
)

type TokenCreateRequest = model.TokenCreateReq
type TokenUpdateRequest = model.TokenUpdateReq

type TokenFilter struct {
	Search string
	Status string
}

type TokenView = model.TokenResp
type TokenCreationView = model.TokenCreateResp
type TokenStats = model.TokenStatsResp

type TokenService struct {
	client *ent.Client
	audit  *AuditService
	now    func() time.Time
}

func NewTokenService(client *ent.Client, audit ...*AuditService) (*TokenService, error) {
	if client == nil {
		return nil, errors.New("accounts: ent client is nil")
	}
	var recorder *AuditService
	if len(audit) > 0 {
		recorder = audit[0]
	}
	return &TokenService{client: client, audit: recorder, now: time.Now}, nil
}

func (s *TokenService) Create(ctx context.Context, ownerID string, request TokenCreateRequest, metadata AuditMetadata) (TokenCreationView, error) {
	if ownerID == "" || strings.TrimSpace(request.Name) == "" {
		return TokenCreationView{}, fmt.Errorf("create token: %w", ErrInvalidRequest)
	}
	permissions := request.Permissions
	if permissions == "" {
		permissions = "read"
	}
	if permissions != "read" && permissions != "read,write" && permissions != "read,write,delete" {
		return TokenCreationView{}, fmt.Errorf("token permissions: %w", ErrInvalidRequest)
	}
	exists, err := s.client.ApiToken.Query().Where(apitoken.UserIdEQ(ownerID), apitoken.NameEQ(strings.TrimSpace(request.Name))).Exist(ctx)
	if err != nil {
		return TokenCreationView{}, fmt.Errorf("check token uniqueness: %w", err)
	}
	if exists {
		return TokenCreationView{}, fmt.Errorf("token name already exists: %w", ErrConflict)
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return TokenCreationView{}, fmt.Errorf("generate API token: %w", err)
	}
	plain := "sk_" + hex.EncodeToString(bytes)
	created, err := s.client.ApiToken.Create().SetUserId(ownerID).SetName(strings.TrimSpace(request.Name)).SetToken(plain).SetPermissions(permissions).SetNillableExpiresAt(request.ExpiresAt).Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return TokenCreationView{}, fmt.Errorf("create token: %w", ErrConflict)
		}
		return TokenCreationView{}, fmt.Errorf("create token: %w", err)
	}
	if err := s.record(ctx, ownerID, "token.create", "token", created.ID, metadata); err != nil {
		return TokenCreationView{}, err
	}
	return TokenCreationView{TokenResp: tokenView(created), Token: plain}, nil
}

func (s *TokenService) Get(ctx context.Context, ownerID, id string) (TokenView, error) {
	entity, err := s.owned(ctx, ownerID, id)
	if err != nil {
		return TokenView{}, err
	}
	return tokenView(entity), nil
}

func (s *TokenService) List(ctx context.Context, ownerID string, filter TokenFilter, page PageRequest) ([]TokenView, PageInfo, error) {
	page = page.normalized()
	query := s.client.ApiToken.Query().Where(apitoken.UserIdEQ(ownerID))
	if search := strings.TrimSpace(filter.Search); search != "" {
		query.Where(apitoken.NameContains(search))
	}
	if filter.Status == "active" {
		query.Where(apitoken.IsActiveEQ(true))
	}
	if filter.Status == "inactive" {
		query.Where(apitoken.IsActiveEQ(false))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, PageInfo{}, fmt.Errorf("count tokens: %w", err)
	}
	entities, err := query.Order(ent.Desc(apitoken.FieldCreatedAt)).Offset((page.Page - 1) * page.Size).Limit(page.Size).All(ctx)
	if err != nil {
		return nil, PageInfo{}, fmt.Errorf("list tokens: %w", err)
	}
	views := make([]TokenView, len(entities))
	for i, entity := range entities {
		views[i] = tokenView(entity)
	}
	return views, pageInfo(page, total), nil
}

func (s *TokenService) Update(ctx context.Context, ownerID, id string, request TokenUpdateRequest, metadata AuditMetadata) (TokenView, error) {
	if strings.TrimSpace(request.Name) == "" {
		return TokenView{}, fmt.Errorf("update token: %w", ErrInvalidRequest)
	}
	if request.Permissions == "" {
		request.Permissions = "read"
	}
	if request.Permissions != "read" && request.Permissions != "read,write" && request.Permissions != "read,write,delete" {
		return TokenView{}, fmt.Errorf("token permissions: %w", ErrInvalidRequest)
	}
	entity, err := s.owned(ctx, ownerID, id)
	if err != nil {
		return TokenView{}, err
	}
	exists, err := s.client.ApiToken.Query().Where(apitoken.UserIdEQ(ownerID), apitoken.NameEQ(strings.TrimSpace(request.Name)), apitoken.IDNEQ(id)).Exist(ctx)
	if err != nil {
		return TokenView{}, fmt.Errorf("check token uniqueness: %w", err)
	}
	if exists {
		return TokenView{}, fmt.Errorf("token name already exists: %w", ErrConflict)
	}
	query := s.client.ApiToken.UpdateOne(entity).SetName(strings.TrimSpace(request.Name)).SetPermissions(request.Permissions).SetNillableExpiresAt(request.ExpiresAt)
	if request.IsActive != nil {
		query.SetIsActive(*request.IsActive)
	}
	updated, err := query.Save(ctx)
	if err != nil {
		return TokenView{}, fmt.Errorf("update token: %w", err)
	}
	if err := s.record(ctx, ownerID, "token.update", "token", id, metadata); err != nil {
		return TokenView{}, err
	}
	return tokenView(updated), nil
}

func (s *TokenService) Toggle(ctx context.Context, ownerID, id string, metadata AuditMetadata) (TokenView, error) {
	entity, err := s.owned(ctx, ownerID, id)
	if err != nil {
		return TokenView{}, err
	}
	updated, err := s.client.ApiToken.UpdateOne(entity).SetIsActive(!entity.IsActive).Save(ctx)
	if err != nil {
		return TokenView{}, fmt.Errorf("toggle token: %w", err)
	}
	if err := s.record(ctx, ownerID, "token.toggle", "token", id, metadata); err != nil {
		return TokenView{}, err
	}
	return tokenView(updated), nil
}

func (s *TokenService) Delete(ctx context.Context, ownerID, id string, metadata AuditMetadata) error {
	entity, err := s.owned(ctx, ownerID, id)
	if err != nil {
		return err
	}
	if err := s.client.ApiToken.DeleteOne(entity).Exec(ctx); err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	return s.record(ctx, ownerID, "token.delete", "token", id, metadata)
}

func (s *TokenService) Stats(ctx context.Context, ownerID string) (TokenStats, error) {
	now := s.now().UTC()
	base := s.client.ApiToken.Query().Where(apitoken.UserIdEQ(ownerID))
	total, err := base.Count(ctx)
	if err != nil {
		return TokenStats{}, fmt.Errorf("count tokens: %w", err)
	}
	active, err := s.client.ApiToken.Query().Where(apitoken.UserIdEQ(ownerID), apitoken.IsActiveEQ(true), apitoken.Or(apitoken.ExpiresAtIsNil(), apitoken.ExpiresAtGT(now))).Count(ctx)
	if err != nil {
		return TokenStats{}, fmt.Errorf("count active tokens: %w", err)
	}
	inactive, err := s.client.ApiToken.Query().Where(apitoken.UserIdEQ(ownerID), apitoken.IsActiveEQ(false)).Count(ctx)
	if err != nil {
		return TokenStats{}, fmt.Errorf("count inactive tokens: %w", err)
	}
	expired, err := s.client.ApiToken.Query().Where(apitoken.UserIdEQ(ownerID), apitoken.ExpiresAtLT(now)).Count(ctx)
	if err != nil {
		return TokenStats{}, fmt.Errorf("count expired tokens: %w", err)
	}
	return TokenStats{TotalTokens: total, ActiveTokens: active, InactiveTokens: inactive, ExpiredTokens: expired}, nil
}

func (s *TokenService) owned(ctx context.Context, ownerID, id string) (*ent.ApiToken, error) {
	entity, err := s.client.ApiToken.Query().Where(apitoken.IDEQ(id), apitoken.UserIdEQ(ownerID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("token: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("load token: %w", err)
	}
	return entity, nil
}

func (s *TokenService) record(ctx context.Context, userID, action, resource, details string, metadata AuditMetadata) error {
	if s.audit == nil {
		return nil
	}
	if err := s.audit.Record(ctx, AuditEntry{UserID: userID, Action: action, Resource: resource, Details: details, IPAddress: metadata.IP, UserAgent: metadata.UserAgent, Metadata: metadata.Metadata}); err != nil {
		return fmt.Errorf("record %s audit: %w", action, err)
	}
	return nil
}

func tokenView(entity *ent.ApiToken) TokenView {
	view := TokenView{ID: entity.ID, Name: entity.Name, Permissions: entity.Permissions, IsActive: entity.IsActive, CreatedAt: entity.CreatedAt, UpdatedAt: entity.UpdatedAt}
	if !entity.LastUsedAt.IsZero() {
		value := entity.LastUsedAt
		view.LastUsedAt = &value
	}
	if !entity.ExpiresAt.IsZero() {
		value := entity.ExpiresAt
		view.ExpiresAt = &value
	}
	return view
}
