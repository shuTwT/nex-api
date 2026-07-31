package config

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

const RedactedValue = "[REDACTED]"

type Config struct {
	Environment string   `mapstructure:"environment"`
	AppURL      string   `mapstructure:"app_url"`
	Server      Server   `mapstructure:"server"`
	Database    Database `mapstructure:"database"`
	Redis       Redis    `mapstructure:"redis"`
	Auth        Auth     `mapstructure:"auth"`
	Payment     Payment  `mapstructure:"payment"`
	Upload      Upload   `mapstructure:"upload"`
	Cron        Cron     `mapstructure:"cron"`
	Log         Log      `mapstructure:"log"`
}

type Server struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
	MaxBodyBytes      int64         `mapstructure:"max_body_bytes"`
	MaxHeaderBytes    int           `mapstructure:"max_header_bytes"`
	CORSOrigins       []string      `mapstructure:"cors_origins"`
}

type Database struct {
	URL             string        `mapstructure:"url"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type Redis struct {
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	TLS      bool   `mapstructure:"tls"`
}

type Auth struct {
	SessionSecret       string `mapstructure:"session_secret"`
	JWTSecret           string `mapstructure:"jwt_secret"`
	NextAuthSecret      string `mapstructure:"nextauth_secret"`
	SessionCookieName   string `mapstructure:"session_cookie_name"`
	GitHubClientID      string `mapstructure:"github_client_id"`
	GitHubClientSecret  string `mapstructure:"github_client_secret"`
	SSOClientID         string `mapstructure:"sso_client_id"`
	SSOClientSecret     string `mapstructure:"sso_client_secret"`
	SSOAuthorizationURL string `mapstructure:"sso_authorization_url"`
	SSOTokenURL         string `mapstructure:"sso_token_url"`
	SSOUserInfoURL      string `mapstructure:"sso_user_info_url"`
	SSOScope            string `mapstructure:"sso_scope"`
}

type Payment struct {
	WeChat WeChatPayment `mapstructure:"wechat"`
	Alipay AlipayPayment `mapstructure:"alipay"`
	Mock   MockPayment   `mapstructure:"mock"`
}

type WeChatPayment struct {
	AppID            string `mapstructure:"app_id"`
	MchID            string `mapstructure:"mch_id"`
	APIKey           string `mapstructure:"api_key"`
	PrivateKey       string `mapstructure:"private_key"`
	PublicKey        string `mapstructure:"public_key"`
	PaymentPublicKey string `mapstructure:"payment_public_key"`
	PublicKeyID      string `mapstructure:"public_key_id"`
	NotifyURL        string `mapstructure:"notify_url"`
	Debug            bool   `mapstructure:"debug"`
}

type AlipayPayment struct {
	AppID           string `mapstructure:"app_id"`
	PrivateKey      string `mapstructure:"private_key"`
	AlipayPublicKey string `mapstructure:"public_key"`
	NotifyURL       string `mapstructure:"notify_url"`
	ReturnURL       string `mapstructure:"return_url"`
	Sandbox         bool   `mapstructure:"sandbox"`
}

type MockPayment struct {
	Enabled     bool          `mapstructure:"enabled"`
	AutoSuccess bool          `mapstructure:"auto_success"`
	Delay       time.Duration `mapstructure:"delay"`
}

type Upload struct {
	Directory     string   `mapstructure:"directory"`
	MaxBytes      int64    `mapstructure:"max_bytes"`
	AllowedTypes  []string `mapstructure:"allowed_types"`
	CreateOnStart bool     `mapstructure:"create_on_start"`
}

type Cron struct {
	Enabled  bool          `mapstructure:"enabled"`
	Secret   string        `mapstructure:"secret"`
	Interval time.Duration `mapstructure:"interval"`
}

type Log struct {
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	AddSource bool   `mapstructure:"add_source"`
}

func (c Config) Redacted() map[string]string {
	return map[string]string{
		"environment":                       c.Environment,
		"app_url":                           c.AppURL,
		"server.host":                       c.Server.Host,
		"server.port":                       formatInt(c.Server.Port),
		"database.url":                      RedactURL(c.Database.URL),
		"redis.url":                         RedactURL(c.Redis.URL),
		"redis.password":                    RedactSecret(c.Redis.Password),
		"auth.session_secret":               RedactSecret(c.Auth.SessionSecret),
		"auth.jwt_secret":                   RedactSecret(c.Auth.JWTSecret),
		"auth.nextauth_secret":              RedactSecret(c.Auth.NextAuthSecret),
		"auth.github_client_id":             c.Auth.GitHubClientID,
		"auth.github_client_secret":         RedactSecret(c.Auth.GitHubClientSecret),
		"auth.sso_client_id":                c.Auth.SSOClientID,
		"auth.sso_client_secret":            RedactSecret(c.Auth.SSOClientSecret),
		"payment.wechat.api_key":            RedactSecret(c.Payment.WeChat.APIKey),
		"payment.wechat.private_key":        RedactSecret(c.Payment.WeChat.PrivateKey),
		"payment.wechat.public_key":         RedactSecret(c.Payment.WeChat.PublicKey),
		"payment.wechat.payment_public_key": RedactSecret(c.Payment.WeChat.PaymentPublicKey),
		"payment.alipay.private_key":        RedactSecret(c.Payment.Alipay.PrivateKey),
		"payment.alipay.public_key":         RedactSecret(c.Payment.Alipay.AlipayPublicKey),
		"cron.secret":                       RedactSecret(c.Cron.Secret),
	}
}

func RedactSecret(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return RedactedValue
}

func RedactURL(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return RedactedValue
	}
	if parsed.User != nil {
		parsed.User = url.User(parsed.User.Username())
	}
	for _, key := range []string{"password", "secret", "token", "key"} {
		values := parsed.Query()
		if values.Has(key) {
			values.Set(key, RedactedValue)
		}
		parsed.RawQuery = values.Encode()
	}
	return parsed.String()
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}
