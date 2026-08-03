// Package settings provides the system settings CRUD service.
package settings

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/service/apierror"
	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/systemsetting"
)

// UpdateItem is a single key/value pair in a bulk settings update.
type UpdateItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DefaultSetting describes a built-in setting default.
type DefaultSetting struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// DefaultGroups groups defaults by operation/payment/oauth sub-sections.
type DefaultGroups struct {
	Basic        []DefaultSetting `json:"basic,omitempty"`
	Announcement []DefaultSetting `json:"announcement,omitempty"`
	Alipay       []DefaultSetting `json:"alipay,omitempty"`
	Wechat       []DefaultSetting `json:"wechat,omitempty"`
}

// DefaultSettings is the full defaults document exposed by the API.
type DefaultSettings struct {
	General   []DefaultSetting `json:"general"`
	Operation DefaultGroups    `json:"operation"`
	Payment   DefaultGroups    `json:"payment"`
	OAuth     DefaultGroups    `json:"oauth"`
}

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
	return DefaultSettings{
		General:   []DefaultSetting{{"siteName", "API 网关", "general", "网站名称"}, {"siteDescription", "一站式 API 服务平台", "general", "网站描述"}, {"siteLogo", "", "general", "网站 Logo"}, {"contactEmail", "support@example.com", "general", "联系邮箱"}},
		Operation: DefaultGroups{Basic: []DefaultSetting{{"registrationEnabled", "true", "operation", "是否允许用户注册"}, {"defaultCredits", "1000", "operation", "新用户默认积分"}, {"inviteRewards", "100", "operation", "邀请奖励积分"}, {"maintenanceMode", "false", "operation", "是否开启维护模式"}}, Announcement: []DefaultSetting{{"announcementEnabled", "false", "operation", "是否启用公告"}, {"announcementContent", "", "operation", "公告内容"}}},
		Payment:   DefaultGroups{Basic: []DefaultSetting{{"alipayEnabled", "false", "payment", "是否开启支付宝支付"}, {"wechatEnabled", "false", "payment", "是否开启微信支付"}, {"creditPrice", "1", "payment", "每积分价格（元）"}, {"minRecharge", "10", "payment", "最低充值金额（元）"}, {"mockPaymentEnabled", "false", "payment", "是否启用模拟支付"}, {"mockPaymentAutoSuccess", "true", "payment", "模拟支付自动成功"}, {"mockPaymentDelay", "2000", "payment", "模拟支付延迟时间（毫秒）"}}, Alipay: []DefaultSetting{{"alipayAppId", "", "payment", "支付宝 AppID"}, {"alipayPrivateKey", "", "payment", "支付宝私钥"}, {"alipayPublicKey", "", "payment", "支付宝公钥"}, {"alipayNotifyUrl", "", "payment", "支付宝回调地址"}, {"alipayReturnUrl", "", "payment", "支付宝返回地址"}, {"alipaySandbox", "false", "payment", "支付宝沙箱模式"}}, Wechat: []DefaultSetting{{"wechatPayAppId", "", "payment", "微信支付 AppID"}, {"wechatPayMchId", "", "payment", "微信支付商户号"}, {"wechatPayApiKey", "", "payment", "微信支付 API 密钥"}, {"wechatPayPrivateKey", "", "payment", "微信支付私钥"}, {"wechatPayPublicKey", "", "payment", "微信支付公钥"}, {"wechatPayPaymentPublicKey", "", "payment", "微信支付平台公钥"}, {"wechatPayPublicKeyId", "", "payment", "微信支付公钥 ID"}, {"wechatPayNotifyUrl", "", "payment", "微信支付回调地址"}, {"wechatPayDebug", "false", "payment", "微信支付调试模式"}}},
		OAuth:     DefaultGroups{Basic: []DefaultSetting{{"oauthProviders", "[]", "oauth", "OAuth 提供商配置"}}},
	}
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
