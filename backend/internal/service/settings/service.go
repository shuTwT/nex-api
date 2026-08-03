// Package settings provides the system settings CRUD service.
package settings

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/systemsetting"
	"github.com/shuTwT/nex-api/backend/internal/service/apierror"
	"github.com/shuTwT/nex-api/backend/pkg/domain/model"
)

type UpdateItem = model.SystemSettingUpdateDTO
type DefaultSetting = model.SystemSettingDefaultDTO
type DefaultGroups = model.SystemSettingDefaultGroupsDTO
type DefaultSettings = model.SystemSettingsDefaultsResp

// Service owns system settings persistence.
type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) (*Service, error) {
	if db == nil {
		return nil, errors.New("settings: database is required")
	}
	return &Service{db: db}, nil
}

// List returns settings, optionally filtered by category.
func (s *Service) List(ctx context.Context, category string) ([]*ent.SystemSetting, error) {
	query := s.db.SystemSetting.Query()
	if category = strings.TrimSpace(category); category != "" {
		query = query.Where(systemsetting.Category(category))
	}
	items, err := query.Order(systemsetting.ByKey()).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list system settings: %w", err)
	}
	return items, nil
}

// Update applies a bulk settings update in one transaction.
func (s *Service) Update(ctx context.Context, items []UpdateItem) error {
	if len(items) == 0 {
		return apierror.NewValidationError(apierror.FieldError{Field: "settings", Reason: "must not be empty"})
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin settings update: %w", err)
	}
	now := time.Now()
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			_ = tx.Rollback()
			return apierror.NewValidationError(apierror.FieldError{Field: "key", Reason: "required"})
		}
		current, findErr := tx.SystemSetting.Query().Where(systemsetting.Key(key)).Only(ctx)
		switch {
		case findErr == nil:
			if _, err = current.Update().SetValue(item.Value).SetUpdatedAt(now).Save(ctx); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("update system setting %q: %w", key, err)
			}
		case ent.IsNotFound(findErr):
			if _, err = tx.SystemSetting.Create().SetKey(key).SetValue(item.Value).SetCategory(CategoryForKey(key)).SetUpdatedAt(now).Save(ctx); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("create system setting %q: %w", key, err)
			}
		default:
			_ = tx.Rollback()
			return fmt.Errorf("find system setting %q: %w", key, findErr)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settings update: %w", err)
	}
	return nil
}

// Defaults returns the built-in setting defaults document.
func (s *Service) Defaults() DefaultSettings {
	return model.DefaultSystemSettings()
}

// Announcement returns the public announcement settings.
func (s *Service) Announcement(ctx context.Context) (map[string]any, error) {
	items, err := s.db.SystemSetting.Query().Where(systemsetting.KeyIn("announcementEnabled", "announcementContent")).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load public announcement: %w", err)
	}
	values := make(map[string]string, len(items))
	for _, item := range items {
		values[item.Key] = item.Value
	}
	return map[string]any{"enabled": values["announcementEnabled"] == "true", "content": values["announcementContent"]}, nil
}

// CategoryForKey derives the setting category from its key prefix.
func CategoryForKey(key string) string {
	switch {
	case strings.HasPrefix(key, "alipay"), strings.HasPrefix(key, "wechat"), strings.HasPrefix(key, "credit"), strings.HasPrefix(key, "minRecharge"), strings.HasPrefix(key, "mockPayment"):
		return "payment"
	case strings.HasPrefix(key, "oauth"):
		return "oauth"
	case strings.HasPrefix(key, "registration"), strings.HasPrefix(key, "defaultCredits"), strings.HasPrefix(key, "inviteRewards"), strings.HasPrefix(key, "maintenance"), strings.HasPrefix(key, "announcement"):
		return "operation"
	default:
		return "general"
	}
}
