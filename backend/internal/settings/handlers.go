package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/systemsetting"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

type Handler struct{ db *ent.Client }

func NewHandler(db *ent.Client) (*Handler, error) {
	if db == nil {
		return nil, errors.New("settings: database is required")
	}
	return &Handler{db: db}, nil
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("settings: mux and handler are required")
	}
	mux.HandleFunc("GET /api/system-settings", handler.list)
	mux.HandleFunc("PUT /api/system-settings", handler.update)
	mux.HandleFunc("GET /api/system-settings/defaults", handler.defaults)
	mux.HandleFunc("GET /api/system-settings/announcement", handler.announcement)
	return nil
}

type settingUpdate struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type updateRequest struct {
	Settings []settingUpdate `json:"settings"`
}

type defaultSetting struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type defaultGroups struct {
	Basic        []defaultSetting `json:"basic,omitempty"`
	Announcement []defaultSetting `json:"announcement,omitempty"`
	Alipay       []defaultSetting `json:"alipay,omitempty"`
	Wechat       []defaultSetting `json:"wechat,omitempty"`
}

type defaultSettings struct {
	General   []defaultSetting `json:"general"`
	Operation defaultGroups    `json:"operation"`
	Payment   defaultGroups    `json:"payment"`
	OAuth     defaultGroups    `json:"oauth"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	query := h.db.SystemSetting.Query()
	if category := strings.TrimSpace(r.URL.Query().Get("category")); category != "" {
		query = query.Where(systemsetting.Category(category))
	}
	items, err := query.Order(systemsetting.ByKey()).All(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("list system settings: %w", err))
		return
	}
	writeData(w, http.StatusOK, items)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var request updateRequest
	if err := decodeJSON(r, &request); err != nil || len(request.Settings) == 0 {
		if err == nil {
			err = runtime.NewValidationError(runtime.FieldError{Field: "settings", Reason: "must not be empty"})
		}
		writeError(w, r, err)
		return
	}
	tx, err := h.db.Tx(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("begin settings update: %w", err))
		return
	}
	now := time.Now()
	for _, item := range request.Settings {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			_ = tx.Rollback()
			writeError(w, r, runtime.NewValidationError(runtime.FieldError{Field: "key", Reason: "required"}))
			return
		}
		current, findErr := tx.SystemSetting.Query().Where(systemsetting.Key(key)).Only(r.Context())
		switch {
		case findErr == nil:
			if _, err = current.Update().SetValue(item.Value).SetUpdatedAt(now).Save(r.Context()); err != nil {
				_ = tx.Rollback()
				writeError(w, r, fmt.Errorf("update system setting %q: %w", key, err))
				return
			}
		case ent.IsNotFound(findErr):
			if _, err = tx.SystemSetting.Create().SetKey(key).SetValue(item.Value).SetCategory(categoryForKey(key)).SetUpdatedAt(now).Save(r.Context()); err != nil {
				_ = tx.Rollback()
				writeError(w, r, fmt.Errorf("create system setting %q: %w", key, err))
				return
			}
		default:
			_ = tx.Rollback()
			writeError(w, r, fmt.Errorf("find system setting %q: %w", key, findErr))
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, r, fmt.Errorf("commit settings update: %w", err))
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "设置已更新"})
}

func (h *Handler) defaults(w http.ResponseWriter, _ *http.Request) {
	writeData(w, http.StatusOK, defaultValues())
}

func (h *Handler) announcement(w http.ResponseWriter, r *http.Request) {
	items, err := h.db.SystemSetting.Query().Where(systemsetting.KeyIn("announcementEnabled", "announcementContent")).All(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("load public announcement: %w", err))
		return
	}
	values := make(map[string]string, len(items))
	for _, item := range items {
		values[item.Key] = item.Value
	}
	writeData(w, http.StatusOK, map[string]any{"enabled": values["announcementEnabled"] == "true", "content": values["announcementContent"]})
}

func defaultValues() defaultSettings {
	return defaultSettings{
		General:   []defaultSetting{{"siteName", "API 网关", "general", "网站名称"}, {"siteDescription", "一站式 API 服务平台", "general", "网站描述"}, {"siteLogo", "", "general", "网站 Logo"}, {"contactEmail", "support@example.com", "general", "联系邮箱"}},
		Operation: defaultGroups{Basic: []defaultSetting{{"registrationEnabled", "true", "operation", "是否允许用户注册"}, {"defaultCredits", "1000", "operation", "新用户默认积分"}, {"inviteRewards", "100", "operation", "邀请奖励积分"}, {"maintenanceMode", "false", "operation", "是否开启维护模式"}}, Announcement: []defaultSetting{{"announcementEnabled", "false", "operation", "是否启用公告"}, {"announcementContent", "", "operation", "公告内容"}}},
		Payment:   defaultGroups{Basic: []defaultSetting{{"alipayEnabled", "false", "payment", "是否开启支付宝支付"}, {"wechatEnabled", "false", "payment", "是否开启微信支付"}, {"creditPrice", "1", "payment", "每积分价格（元）"}, {"minRecharge", "10", "payment", "最低充值金额（元）"}, {"mockPaymentEnabled", "false", "payment", "是否启用模拟支付"}, {"mockPaymentAutoSuccess", "true", "payment", "模拟支付自动成功"}, {"mockPaymentDelay", "2000", "payment", "模拟支付延迟时间（毫秒）"}}, Alipay: []defaultSetting{{"alipayAppId", "", "payment", "支付宝 AppID"}, {"alipayPrivateKey", "", "payment", "支付宝私钥"}, {"alipayPublicKey", "", "payment", "支付宝公钥"}, {"alipayNotifyUrl", "", "payment", "支付宝回调地址"}, {"alipayReturnUrl", "", "payment", "支付宝返回地址"}, {"alipaySandbox", "false", "payment", "支付宝沙箱模式"}}, Wechat: []defaultSetting{{"wechatPayAppId", "", "payment", "微信支付 AppID"}, {"wechatPayMchId", "", "payment", "微信支付商户号"}, {"wechatPayApiKey", "", "payment", "微信支付 API 密钥"}, {"wechatPayPrivateKey", "", "payment", "微信支付私钥"}, {"wechatPayPublicKey", "", "payment", "微信支付公钥"}, {"wechatPayPaymentPublicKey", "", "payment", "微信支付平台公钥"}, {"wechatPayPublicKeyId", "", "payment", "微信支付公钥 ID"}, {"wechatPayNotifyUrl", "", "payment", "微信支付回调地址"}, {"wechatPayDebug", "false", "payment", "微信支付调试模式"}}},
		OAuth:     defaultGroups{Basic: []defaultSetting{{"oauthProviders", "[]", "oauth", "OAuth 提供商配置"}}},
	}
}

func categoryForKey(key string) string {
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

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	principal, err := authz.AdminPolicy(r.Context())
	if err == nil && principal.UserID != "" {
		return true
	}
	status := http.StatusUnauthorized
	if errors.Is(err, authz.ErrForbidden) {
		status = http.StatusForbidden
	}
	writeError(w, r, runtime.NewAPIError(status, map[bool]string{true: "forbidden", false: "unauthorized"}[status == http.StatusForbidden], map[bool]string{true: "access denied", false: "authentication required"}[status == http.StatusForbidden], err))
	return false
}

func decodeJSON[T any](r *http.Request, destination *T) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return runtime.NewValidationError(runtime.FieldError{Field: "body", Reason: "invalid JSON"})
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return runtime.NewValidationError(runtime.FieldError{Field: "body", Reason: "must contain exactly one JSON value"})
	}
	return nil
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	_ = runtime.WriteData(w, status, data)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) { _ = runtime.WriteError(w, r, err) }
