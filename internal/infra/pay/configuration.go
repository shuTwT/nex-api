package pay

import (
	"strconv"
	"strings"
	"time"
)

// WeChatConfiguration holds the WeChat Pay client settings loaded from the
// system settings table by the payment service.
type WeChatConfiguration struct {
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

// AlipayConfiguration holds the Alipay client settings.
type AlipayConfiguration struct {
	AppID           string
	PrivateKey      string
	AlipayPublicKey string
	NotifyURL       string
	ReturnURL       string
	Sandbox         bool
}

// MockPaymentConfiguration controls the mock payment provider.
type MockPaymentConfiguration struct {
	Enabled     bool
	AutoSuccess bool
	Delay       time.Duration
}

// PaymentConfiguration aggregates all payment provider settings.
type PaymentConfiguration struct {
	WeChat        WeChatConfiguration
	Alipay        AlipayConfiguration
	Mock          MockPaymentConfiguration
	CreditPrice   float64
	MinRecharge   float64
	AlipayEnabled bool
	WeChatEnabled bool
}

// WechatConfigured reports whether the WeChat provider can be built.
func WechatConfigured(value WeChatConfiguration) bool {
	if !AllPresent(value.AppID, value.MchID, value.APIKey, value.PrivateKey) {
		return false
	}
	return strings.TrimSpace(value.PublicKey) != "" || AllPresent(value.PaymentPublicKey, value.PublicKeyID)
}

// AlipayConfigured reports whether the Alipay provider can be built.
func AlipayConfigured(value AlipayConfiguration) bool {
	return AllPresent(value.AppID, value.PrivateKey, value.AlipayPublicKey)
}

// AllPresent reports whether every value is non-blank.
func AllPresent(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

// SettingURL falls back to appURL+path when the configured URL is blank.
func SettingURL(value, appURL, path string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return strings.TrimRight(strings.TrimSpace(appURL), "/") + path
}

// SettingBool parses a boolean setting with a fallback default.
func SettingBool(values map[string]string, key string, fallback bool) bool {
	value, ok := values[key]
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}

// SettingInt parses an integer setting with a fallback default.
func SettingInt(values map[string]string, key string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(values[key]))
	if err != nil {
		return fallback
	}
	return parsed
}
