package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/infra/config"
	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/session"
	"github.com/shuTwT/nex-api/backend/ent/user"
)

const (
	defaultSessionCookieName = "nex_session"
	defaultCSRFTokenName     = "nex_csrf"
	defaultSessionTTL        = 30 * 24 * time.Hour
	sessionTokenBytes        = 32
	csrfTokenBytes           = 32
)

var (
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrUnauthenticated    = errors.New("auth: unauthenticated")
	ErrForbidden          = errors.New("auth: forbidden")
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Credits  int    `json:"credits"`
}

type AuthContext struct {
	User      User
	SessionID string
	ExpiresAt time.Time
	token     string
}

// Token returns the raw session token of this authentication context. It is
// only populated by Login/CreateSession and is used by the handler layer to
// set the session cookie.
func (c AuthContext) Token() string { return c.token }

type Clock interface {
	Now() time.Time
}

type clockFunc func() time.Time

func (f clockFunc) Now() time.Time { return f() }

type Service struct {
	client            *ent.Client
	sessionSecret     []byte
	sessionCookieName string
	csrfCookieName    string
	sessionTTL        time.Duration
	clock             Clock
	tokenGenerator    func(int) ([]byte, error)
	secureCookies     bool
	rateLimiter       *loginRateLimiter
}

type Option func(*serviceOptions)

type serviceOptions struct {
	sessionTTL     time.Duration
	clock          Clock
	tokenGenerator func(int) ([]byte, error)
	loginLimit     int
	loginWindow    time.Duration
	secureCookies  bool
}

func WithClock(clock Clock) Option {
	return func(options *serviceOptions) {
		if clock != nil {
			options.clock = clock
		}
	}
}

func WithSessionTTL(ttl time.Duration) Option {
	return func(options *serviceOptions) {
		if ttl > 0 {
			options.sessionTTL = ttl
		}
	}
}

func WithLoginRateLimit(maxAttempts int, window time.Duration) Option {
	return func(options *serviceOptions) {
		if maxAttempts > 0 {
			options.loginLimit = maxAttempts
		}
		if window > 0 {
			options.loginWindow = window
		}
	}
}

// WithSecureCookies 控制会话/CSRF cookie 是否带 Secure 标志。
// 默认 true(生产安全默认);开发环境(HTTP)应传 false,
// 否则浏览器会拒绝保存 Secure cookie 导致登录静默失败。
func WithSecureCookies(secure bool) Option {
	return func(options *serviceOptions) {
		options.secureCookies = secure
	}
}

func NewService(client *ent.Client, authConfig config.Auth, options ...Option) (*Service, error) {
	if client == nil {
		return nil, errors.New("auth: ent client is nil")
	}
	serviceOptions := serviceOptions{
		sessionTTL:     defaultSessionTTL,
		clock:          clockFunc(time.Now),
		tokenGenerator: randomBytes,
		loginLimit:     5,
		loginWindow:    15 * time.Minute,
		secureCookies:  true,
	}
	for _, option := range options {
		if option != nil {
			option(&serviceOptions)
		}
	}
	cookieName := strings.TrimSpace(authConfig.SessionCookieName)
	if cookieName == "" {
		cookieName = defaultSessionCookieName
	}
	return &Service{
		client:            client,
		sessionSecret:     []byte(authConfig.SessionSecret),
		sessionCookieName: cookieName,
		csrfCookieName:    defaultCSRFTokenName,
		sessionTTL:        serviceOptions.sessionTTL,
		clock:             serviceOptions.clock,
		tokenGenerator:    serviceOptions.tokenGenerator,
		secureCookies:     serviceOptions.secureCookies,
		rateLimiter:       newLoginRateLimiter(serviceOptions.loginLimit, serviceOptions.loginWindow, serviceOptions.clock),
	}, nil
}

func New(client *ent.Client, authConfig config.Auth, options ...Option) (*Service, error) {
	return NewService(client, authConfig, options...)
}

func (s *Service) Login(ctx context.Context, email, password, currentToken string) (AuthContext, error) {
	if ctx == nil {
		return AuthContext{}, errors.New("auth: login context is nil")
	}
	account, err := s.client.User.Query().Where(user.Email(strings.TrimSpace(email))).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return AuthContext{}, ErrInvalidCredentials
		}
		return AuthContext{}, fmt.Errorf("query login user: %w", err)
	}
	matched, err := VerifyPassword(password, account.Password)
	if err != nil {
		return AuthContext{}, fmt.Errorf("verify login password: %w", err)
	}
	if !matched {
		return AuthContext{}, ErrInvalidCredentials
	}
	if err := s.deleteSession(ctx, currentToken); err != nil {
		return AuthContext{}, fmt.Errorf("rotate login session: %w", err)
	}
	return s.createSession(ctx, toUser(account))
}

func (s *Service) Authenticate(ctx context.Context, token string) (AuthContext, error) {
	if ctx == nil {
		return AuthContext{}, errors.New("auth: authentication context is nil")
	}
	if token == "" {
		return AuthContext{}, ErrUnauthenticated
	}
	storedToken := s.persistedToken(token)
	storedSession, err := s.client.Session.Query().Where(session.SessionToken(storedToken)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return AuthContext{}, ErrUnauthenticated
		}
		return AuthContext{}, fmt.Errorf("query session: %w", err)
	}
	if !s.clock.Now().Before(storedSession.Expires) {
		return AuthContext{}, ErrUnauthenticated
	}
	account, err := s.client.User.Get(ctx, storedSession.UserId)
	if err != nil {
		if ent.IsNotFound(err) {
			return AuthContext{}, ErrUnauthenticated
		}
		return AuthContext{}, fmt.Errorf("query session user: %w", err)
	}
	return AuthContext{
		User:      toUser(account),
		SessionID: storedSession.ID,
		ExpiresAt: storedSession.Expires,
	}, nil
}

func (s *Service) RotateSession(ctx context.Context, current AuthContext) (AuthContext, error) {
	if current.SessionID == "" || current.User.ID == "" {
		return AuthContext{}, ErrUnauthenticated
	}
	if err := s.client.Session.DeleteOneID(current.SessionID).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return AuthContext{}, ErrUnauthenticated
		}
		return AuthContext{}, fmt.Errorf("delete session during rotation: %w", err)
	}
	return s.createSession(ctx, current.User)
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if ctx == nil {
		return errors.New("auth: logout context is nil")
	}
	if err := s.deleteSession(ctx, token); err != nil {
		return fmt.Errorf("delete logout session: %w", err)
	}
	return nil
}

func (s *Service) CreateSession(ctx context.Context, account User) (AuthContext, error) {
	if ctx == nil {
		return AuthContext{}, errors.New("auth: session context is nil")
	}
	if account.ID == "" || account.Email == "" {
		return AuthContext{}, errors.New("auth: session user is invalid")
	}
	return s.createSession(ctx, account)
}

func (s *Service) createSession(ctx context.Context, account User) (AuthContext, error) {
	rawToken, err := s.tokenGenerator(sessionTokenBytes)
	if err != nil {
		return AuthContext{}, fmt.Errorf("generate session token: %w", err)
	}
	if len(rawToken) != sessionTokenBytes {
		return AuthContext{}, errors.New("auth: session token generator returned invalid length")
	}
	now := s.clock.Now().UTC()
	expiresAt := now.Add(s.sessionTTL)
	token := base64.RawURLEncoding.EncodeToString(rawToken)
	created, err := s.client.Session.Create().
		SetSessionToken(s.persistedToken(token)).
		SetUserID(account.ID).
		SetExpires(expiresAt).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return AuthContext{}, fmt.Errorf("create session: %w", err)
	}
	return AuthContext{User: account, SessionID: created.ID, ExpiresAt: expiresAt, token: token}, nil
}

func (s *Service) deleteSession(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if _, err := s.client.Session.Delete().Where(session.SessionToken(s.persistedToken(token))).Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) persistedToken(token string) string {
	digest := hmac.New(sha256.New, s.sessionSecret)
	if _, err := digest.Write([]byte(token)); err != nil {
		return ""
	}
	return hex.EncodeToString(digest.Sum(nil))
}

func (s *Service) SessionCookieName() string { return s.sessionCookieName }

func (s *Service) CSRFTokenCookieName() string { return s.csrfCookieName }

// SecureCookies reports whether session/CSRF cookies must carry the Secure
// flag; the handler layer uses it when writing cookies.
func (s *Service) SecureCookies() bool { return s.secureCookies }

// SessionTTL reports the session lifetime used for cookie expiration.
func (s *Service) SessionTTL() time.Duration { return s.sessionTTL }

func toUser(account *ent.User) User {
	return User{ID: account.ID, Email: account.Email, Username: account.Username, Role: account.Role, Credits: account.Credits}
}

func randomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}
