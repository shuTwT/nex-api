package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/shuTwT/nex-api/ent/systemsetting"
	pay "github.com/shuTwT/nex-api/internal/infra/pay"
)

func (s *Service) loadConfiguration(ctx context.Context) (pay.PaymentConfiguration, error) {
	settings, err := s.client.SystemSetting.Query().Where(systemsetting.Category("payment")).All(ctx)
	if err != nil {
		return pay.PaymentConfiguration{}, fmt.Errorf("get payment settings: %w", err)
	}
	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}

	return pay.PaymentConfiguration{
		WeChat: pay.WeChatConfiguration{
			AppID:            values["wechatPayAppId"],
			MchID:            values["wechatPayMchId"],
			APIKey:           values["wechatPayApiKey"],
			PrivateKey:       values["wechatPayPrivateKey"],
			PublicKey:        values["wechatPayPublicKey"],
			PaymentPublicKey: values["wechatPayPaymentPublicKey"],
			PublicKeyID:      values["wechatPayPublicKeyId"],
			NotifyURL:        pay.SettingURL(values["wechatPayNotifyUrl"], s.appURL, "/api/payment/callback/wechat"),
			Debug:            pay.SettingBool(values, "wechatPayDebug", false),
		},
		Alipay: pay.AlipayConfiguration{
			AppID:           values["alipayAppId"],
			PrivateKey:      values["alipayPrivateKey"],
			AlipayPublicKey: values["alipayPublicKey"],
			NotifyURL:       pay.SettingURL(values["alipayNotifyUrl"], s.appURL, "/api/payment/callback/alipay"),
			ReturnURL:       pay.SettingURL(values["alipayReturnUrl"], s.appURL, "/payment/result"),
			Sandbox:         pay.SettingBool(values, "alipaySandbox", false),
		},
		Mock: pay.MockPaymentConfiguration{
			Enabled:     pay.SettingBool(values, "mockPaymentEnabled", false),
			AutoSuccess: pay.SettingBool(values, "mockPaymentAutoSuccess", true),
			Delay:       time.Duration(pay.SettingInt(values, "mockPaymentDelay", 2000)) * time.Millisecond,
		},
		CreditPrice:   parseSettingFloat(values, "creditPrice", 1),
		MinRecharge:   parseSettingFloat(values, "minRecharge", 10),
		AlipayEnabled: pay.SettingBool(values, "alipayEnabled", false),
		WeChatEnabled: pay.SettingBool(values, "wechatEnabled", false),
	}, nil
}

func (s *Service) resolveProvider(ctx context.Context, method pay.PaymentMethod) (pay.Provider, string, error) {
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
	case pay.PaymentMethodWeChat:
		if !pay.WechatConfigured(configuration.WeChat) {
			return nil, "", fmt.Errorf("wechat: %w", pay.ErrProviderUnavailable)
		}
		return s.buildProvider(method, configuration), configuration.WeChat.NotifyURL, nil
	case pay.PaymentMethodAlipay:
		if !pay.AlipayConfigured(configuration.Alipay) {
			return nil, "", fmt.Errorf("alipay: %w", pay.ErrProviderUnavailable)
		}
		return s.buildProvider(method, configuration), configuration.Alipay.NotifyURL, nil
	case pay.PaymentMethodMock:
		if !configuration.Mock.Enabled {
			return nil, "", fmt.Errorf("mock: %w", pay.ErrProviderUnavailable)
		}
		return s.buildProvider(method, configuration), "", nil
	default:
		return nil, "", fmt.Errorf("unsupported payment method %q", method)
	}
}

func (s *Service) buildProvider(method pay.PaymentMethod, configuration pay.PaymentConfiguration) pay.Provider {
	if s.providerFactory != nil {
		return s.providerFactory(method, configuration)
	}
	switch method {
	case pay.PaymentMethodWeChat:
		return pay.NewWeChatProvider(configuration.WeChat, nil)
	case pay.PaymentMethodAlipay:
		return pay.NewAlipayProvider(configuration.Alipay, nil)
	case pay.PaymentMethodMock:
		return pay.NewMockProvider(configuration.Mock)
	default:
		return nil
	}
}
