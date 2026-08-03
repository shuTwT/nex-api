package accounts

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/apiusage"
	"github.com/shuTwT/nex-api/ent/user"
	"github.com/shuTwT/nex-api/internal/service/auth"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type ProfileUpdateRequest = model.ProfileUpdateReq
type PasswordUpdateRequest = model.PasswordUpdateReq
type ProfileView = model.ProfileResp

type ProfileService struct {
	client *ent.Client
	audit  *AuditService
}

func NewProfileService(client *ent.Client, audit ...*AuditService) (*ProfileService, error) {
	if client == nil {
		return nil, errors.New("accounts: ent client is nil")
	}
	var recorder *AuditService
	if len(audit) > 0 {
		recorder = audit[0]
	}
	return &ProfileService{client: client, audit: recorder}, nil
}

func (s *ProfileService) Get(ctx context.Context, ownerID string) (ProfileView, error) {
	entity, err := s.client.User.Query().Where(user.IDEQ(ownerID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ProfileView{}, fmt.Errorf("get profile: %w", ErrNotFound)
		}
		return ProfileView{}, fmt.Errorf("get profile: %w", err)
	}
	spent := 0
	usage, err := s.client.ApiUsage.Query().Where(apiusage.UserIdEQ(ownerID)).All(ctx)
	if err != nil {
		return ProfileView{}, fmt.Errorf("sum profile usage: %w", err)
	}
	for _, item := range usage {
		spent += item.Credits
	}
	return ProfileView{ID: entity.ID, Name: entity.Name, Email: entity.Email, Image: entity.Image, Username: entity.Username, Role: entity.Role, Credits: entity.Credits, CreatedAt: entity.CreatedAt, TotalCreditsSpent: spent, TotalRequests: len(usage)}, nil
}

func (s *ProfileService) Update(ctx context.Context, ownerID string, request ProfileUpdateRequest, metadata AuditMetadata) (ProfileView, error) {
	query := s.client.User.UpdateOneID(ownerID)
	if request.Name != nil {
		query.SetName(strings.TrimSpace(*request.Name))
	}
	if request.Email != nil {
		value := strings.TrimSpace(strings.ToLower(*request.Email))
		if !strings.Contains(value, "@") {
			return ProfileView{}, fmt.Errorf("update profile email: %w", ErrInvalidRequest)
		}
		query.SetEmail(value)
	}
	if request.Username != nil {
		value := strings.TrimSpace(*request.Username)
		if len(value) < 3 || len(value) > 20 {
			return ProfileView{}, fmt.Errorf("update profile username: %w", ErrInvalidRequest)
		}
		query.SetUsername(value)
	}
	updated, err := query.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ProfileView{}, fmt.Errorf("update profile: %w", ErrNotFound)
		}
		if ent.IsConstraintError(err) {
			return ProfileView{}, fmt.Errorf("update profile: %w", ErrConflict)
		}
		return ProfileView{}, fmt.Errorf("update profile: %w", err)
	}
	if err := s.record(ctx, ownerID, "profile.update", "profile", ownerID, metadata); err != nil {
		return ProfileView{}, err
	}
	return s.Get(ctx, updated.ID)
}

func (s *ProfileService) UpdatePassword(ctx context.Context, ownerID string, request PasswordUpdateRequest, metadata AuditMetadata) error {
	if len(request.NewPassword) < 8 || len(request.NewPassword) > 100 || request.CurrentPassword == "" {
		return fmt.Errorf("update password: %w", ErrInvalidRequest)
	}
	entity, err := s.client.User.Query().Where(user.IDEQ(ownerID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("load password: %w", ErrNotFound)
		}
		return fmt.Errorf("load password: %w", err)
	}
	valid, err := auth.VerifyPassword(request.CurrentPassword, entity.Password)
	if err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	if !valid {
		return fmt.Errorf("verify password: %w", ErrInvalidRequest)
	}
	hash, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if _, err := s.client.User.UpdateOne(entity).SetPassword(hash).Save(ctx); err != nil {
		return fmt.Errorf("save password: %w", err)
	}
	return s.record(ctx, ownerID, "profile.password.update", "profile", ownerID, metadata)
}

func (s *ProfileService) record(ctx context.Context, userID, action, resource, details string, metadata AuditMetadata) error {
	if s.audit == nil {
		return nil
	}
	if err := s.audit.Record(ctx, AuditEntry{UserID: userID, Action: action, Resource: resource, Details: details, IPAddress: metadata.IP, UserAgent: metadata.UserAgent, Metadata: metadata.Metadata}); err != nil {
		return fmt.Errorf("record %s audit: %w", action, err)
	}
	return nil
}
