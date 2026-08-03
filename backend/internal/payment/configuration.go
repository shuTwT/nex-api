package payment

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/database/ent/systemsetting"
)

type weChatConfiguration struct {
	AppID            string
	MchID            string
	APIKey           string
	PrivateKey       string
	PublicKey        string
	PaymentPublicKey string
	PublicKeyID      string
	NotifyURL        string
	Debug            bool
}

type alipayConfiguration struct {
	AppID           string
	PrivateKey      string
	AlipayPublicKey string
	NotifyURL       string
	ReturnURL       string
	Sandbox         bool
}

type mockPaymentConfiguration struct {
	Enabled     bool
	AutoSuccess bool
	Delay       time.Duration
}

type paymentConfiguration struct {
	WeChat        weChatConfiguration
	Alipay        alipayConfiguration
	Mock          mockPaymentConfiguration
	CreditPrice   float64
	MinRecharge   float64
	AlipayEnabled bool
	WeChatEnabled bool
}

func (s *Service) loadConfiguration(ctx context.Context) (paymentConfiguration, error) {
	settings, err := s.client.SystemSetting.Query().Where(systemsetting.Category("payment")).All(ctx)
	if err != nil {
		return paymentConfiguration{}, fmt.Errorf("get payment settings: %w", err)
	}
	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}

	return paymentConfiguration{
		WeChat: weChatConfiguration{
			AppID:            values["wechatPayAppId"],
			MchID:            values["wechatPayMchId"],
			APIKey:           values["wechatPayApiKey"],
			PrivateKey:       values["wechatPayPrivateKey"],
			PublicKey:        values["wechatPayPublicKey"],
			PaymentPublicKey: values["wechatPayPaymentPublicKey"],
			PublicKeyID:      values["wechatPayPublicKeyId"],
			NotifyURL:        settingURL(values["wechatPayNotifyUrl"], s.appURL, "/api/payment/callback/wechat"),
			Debug:            settingBool(values, "wechatPayDebug", false),
		},
		Alipay: alipayConfiguration{
			AppID:           values["alipayAppId"],
			PrivateKey:      values["alipayPrivateKey"],
			AlipayPublicKey: values["alipayPublicKey"],
			NotifyURL:       settingURL(values["alipayNotifyUrl"], s.appURL, "/api/payment/callback/alipay"),
			ReturnURL:       settingURL(values["alipayReturnUrl"], s.appURL, "/payment/result"),
			Sandbox:         settingBool(values, "alipaySandbox", false),
		},
		Mock: mockPaymentConfiguration{
			Enabled:     settingBool(values, "mockPaymentEnabled", false),
			AutoSuccess: settingBool(values, "mockPaymentAutoSuccess", true),
			Delay:       time.Duration(settingInt(values, "mockPaymentDelay", 2000)) * time.Millisecond,
		},
		CreditPrice:   parseSettingFloat(values, "creditPrice", 1),
		MinRecharge:   parseSettingFloat(values, "minRecharge", 10),
		AlipayEnabled: settingBool(values, "alipayEnabled", false),
		WeChatEnabled: settingBool(values, "wechatEnabled", false),
	}, nil
}

func (s *Service) resolveProvider(ctx context.Context, method PaymentMethod) (Provider, string, error) {
	if s.providers != nil {
		provider, ok := s.providers[method]
		if !ok {
			return nil, "", fmt.Errorf("unsupported payment method %q", method)
		}
		return provider, "", nil
	}

	configuration, err := s.loadConfiguration(ctx)
	if err != nil {
		return nil, "", err
	}
	switch method {
	case PaymentMethodWeChat:
		if !wechatConfigured(configuration.WeChat) {
			return nil, "", fmt.Errorf("wechat: %w", ErrProviderUnavailable)
		}
		return s.buildProvider(method, configuration), configuration.WeChat.NotifyURL, nil
	case PaymentMethodAlipay:
		if !alipayConfigured(configuration.Alipay) {
			return nil, "", fmt.Errorf("alipay: %w", ErrProviderUnavailable)
		}
		return s.buildProvider(method, configuration), configuration.Alipay.NotifyURL, nil
	case PaymentMethodMock:
		if !configuration.Mock.Enabled {
			return nil, "", fmt.Errorf("mock: %w", ErrProviderUnavailable)
		}
		return s.buildProvider(method, configuration), "", nil
	default:
		return nil, "", fmt.Errorf("unsupported payment method %q", method)
	}
}

func (s *Service) buildProvider(method PaymentMethod, configuration paymentConfiguration) Provider {
	if s.providerFactory != nil {
		return s.providerFactory(method, configuration)
	}
	switch method {
	case PaymentMethodWeChat:
		return NewWeChatProvider(configuration.WeChat, nil)
	case PaymentMethodAlipay:
		return NewAlipayProvider(configuration.Alipay, nil)
	case PaymentMethodMock:
		return NewMockProvider(configuration.Mock)
	default:
		return nil
	}
}

func wechatConfigured(value weChatConfiguration) bool {
	if !allPresent(value.AppID, value.MchID, value.APIKey, value.PrivateKey) {
		return false
	}
	return strings.TrimSpace(value.PublicKey) != "" || allPresent(value.PaymentPublicKey, value.PublicKeyID)
}

func alipayConfigured(value alipayConfiguration) bool {
	return allPresent(value.AppID, value.PrivateKey, value.AlipayPublicKey)
}

func allPresent(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func settingURL(value, appURL, path string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return strings.TrimRight(strings.TrimSpace(appURL), "/") + path
}

func settingBool(values map[string]string, key string, fallback bool) bool {
	value, ok := values[key]
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}

func settingInt(values map[string]string, key string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(values[key]))
	if err != nil {
		return fallback
	}
	return parsed
}
