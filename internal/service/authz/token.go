package authz

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/apitoken"
	"github.com/shuTwT/nex-api/ent/user"
)

const (
	TokenPrefix        = "sk_"
	TokenEntropySize   = 32
	DefaultPermissions = "read"
)

var (
	ErrInvalidBearer      = errors.New("authz: invalid bearer authorization")
	ErrInvalidToken       = errors.New("authz: invalid API token")
	ErrInactiveToken      = errors.New("authz: inactive API token")
	ErrExpiredToken       = errors.New("authz: expired API token")
	ErrInvalidPermissions = errors.New("authz: invalid API token permissions")
	ErrTokenNotFound      = errors.New("authz: API token not found")
)

func GenerateToken() (string, error) {
	return generateToken(rand.Reader)
}

func GenerateAPIToken() (string, error) {
	return GenerateToken()
}

func generateToken(reader io.Reader) (string, error) {
	bytes := make([]byte, TokenEntropySize)
	if _, err := io.ReadFull(reader, bytes); err != nil {
		return "", fmt.Errorf("generate API token: %w", err)
	}
	return TokenPrefix + hex.EncodeToString(bytes), nil
}

func IsGeneratedToken(token string) bool {
	if len(token) != len(TokenPrefix)+TokenEntropySize*2 || !strings.HasPrefix(token, TokenPrefix) {
		return false
	}
	_, err := hex.DecodeString(token[len(TokenPrefix):])
	return err == nil
}

func ParseBearerToken(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidBearer
	}
	return parts[1], nil
}

func ConstantTimeTokenEqual(expected, presented string) bool {
	if expected == "" || presented == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(presented)) == 1
}

type StoredToken struct {
	ID          string
	UserID      string
	Role        string
	Token       string `json:"-"`
	Permissions string
	IsActive    bool
	ExpiresAt   time.Time
}

type TokenStore interface {
	LookupToken(ctx context.Context, token string) (StoredToken, error)
	TouchLastUsedAt(ctx context.Context, tokenID string, at time.Time) error
}

type Clock interface {
	Now() time.Time
}

type clockFunc func() time.Time

func (f clockFunc) Now() time.Time { return f() }

type TokenServiceOption func(*tokenServiceOptions)

type tokenServiceOptions struct {
	clock Clock
}

func WithClock(clock Clock) TokenServiceOption {
	return func(options *tokenServiceOptions) {
		if clock != nil {
			options.clock = clock
		}
	}
}

type TokenService struct {
	store TokenStore
	clock Clock
}

func NewTokenService(store TokenStore, options ...TokenServiceOption) (*TokenService, error) {
	if store == nil {
		return nil, errors.New("authz: token store is nil")
	}
	serviceOptions := tokenServiceOptions{clock: clockFunc(time.Now)}
	for _, option := range options {
		if option != nil {
			option(&serviceOptions)
		}
	}
	return &TokenService{store: store, clock: serviceOptions.clock}, nil
}

func (s *TokenService) AuthenticateBearer(ctx context.Context, header string) (Principal, error) {
	token, err := ParseBearerToken(header)
	if err != nil {
		return Principal{}, err
	}
	return s.Authenticate(ctx, token)
}

func (s *TokenService) Authenticate(ctx context.Context, presented string) (Principal, error) {
	if ctx == nil {
		return Principal{}, errors.New("authz: token authentication context is nil")
	}
	if presented == "" {
		return Principal{}, ErrInvalidToken
	}
	stored, err := s.store.LookupToken(ctx, presented)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return Principal{}, ErrInvalidToken
		}
		return Principal{}, fmt.Errorf("lookup API token: %w", err)
	}
	if !ConstantTimeTokenEqual(stored.Token, presented) || stored.UserID == "" {
		return Principal{}, ErrInvalidToken
	}
	if !stored.IsActive {
		return Principal{}, fmt.Errorf("%w: %w", ErrInvalidToken, ErrInactiveToken)
	}
	now := s.clock.Now().UTC()
	if !stored.ExpiresAt.IsZero() && !now.Before(stored.ExpiresAt) {
		return Principal{}, fmt.Errorf("%w: %w", ErrInvalidToken, ErrExpiredToken)
	}
	permissions, err := ParsePermissions(stored.Permissions)
	if err != nil {
		return Principal{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	if err := s.store.TouchLastUsedAt(ctx, stored.ID, now); err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return Principal{}, ErrInvalidToken
		}
		return Principal{}, fmt.Errorf("touch API token last used time: %w", err)
	}
	return Principal{
		UserID:      stored.UserID,
		Role:        stored.Role,
		Source:      APITokenCredential,
		TokenID:     stored.ID,
		Permissions: permissions,
	}, nil
}

type EntTokenStore struct {
	client *ent.Client
}

func NewEntTokenStore(client *ent.Client) (*EntTokenStore, error) {
	if client == nil {
		return nil, errors.New("authz: ent client is nil")
	}
	return &EntTokenStore{client: client}, nil
}

func (s *EntTokenStore) LookupToken(ctx context.Context, token string) (StoredToken, error) {
	if ctx == nil {
		return StoredToken{}, errors.New("authz: token lookup context is nil")
	}
	entity, err := s.client.ApiToken.Query().Where(apitoken.Token(token)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return StoredToken{}, ErrTokenNotFound
		}
		return StoredToken{}, fmt.Errorf("query API token: %w", err)
	}
	account, err := s.client.User.Query().Where(user.ID(entity.UserId)).Select(
		user.FieldID,
		user.FieldRole,
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return StoredToken{}, ErrTokenNotFound
		}
		return StoredToken{}, fmt.Errorf("query API token owner: %w", err)
	}
	return StoredToken{
		ID:          entity.ID,
		UserID:      account.ID,
		Role:        account.Role,
		Token:       entity.Token,
		Permissions: entity.Permissions,
		IsActive:    entity.IsActive,
		ExpiresAt:   entity.ExpiresAt,
	}, nil
}

func (s *EntTokenStore) TouchLastUsedAt(ctx context.Context, tokenID string, at time.Time) error {
	if ctx == nil {
		return errors.New("authz: token update context is nil")
	}
	updated, err := s.client.ApiToken.Update().Where(
		apitoken.ID(tokenID),
		apitoken.IsActive(true),
	).SetLastUsedAt(at.UTC()).Save(ctx)
	if err != nil {
		return fmt.Errorf("update API token last used time: %w", err)
	}
	if updated != 1 {
		return ErrTokenNotFound
	}
	return nil
}
