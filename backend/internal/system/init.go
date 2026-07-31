package system

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/auth"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

var (
	ErrAlreadyInitialized = errors.New("system: already initialized")
	ErrInvalidInitialize  = errors.New("system: invalid initialization request")
)

type InitializeRequest struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type Service struct {
	client *ent.Client
}

var initializeMu sync.Mutex

func NewService(client *ent.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("system: ent client is nil")
	}
	return &Service{client: client}, nil
}

func (s *Service) Initialized(ctx context.Context) (bool, error) {
	if ctx == nil {
		return false, errors.New("system: initialization context is nil")
	}
	initialized, err := s.client.User.Query().Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("system: check initialized state: %w", err)
	}
	return initialized, nil
}

func (s *Service) Initialize(ctx context.Context, request InitializeRequest) (*ent.User, error) {
	if ctx == nil {
		return nil, errors.New("system: initialization context is nil")
	}
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Username = strings.TrimSpace(request.Username)
	if err := validateInitializeRequest(request); err != nil {
		return nil, err
	}
	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		return nil, fmt.Errorf("system: hash admin password: %w", err)
	}

	initializeMu.Lock()
	defer initializeMu.Unlock()
	initialized, err := s.client.User.Query().Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("system: check before initialize: %w", err)
	}
	if initialized {
		return nil, ErrAlreadyInitialized
	}
	admin, err := s.client.User.Create().
		SetName(request.Username).
		SetEmail(request.Email).
		SetUsername(request.Username).
		SetPassword(hashedPassword).
		SetRole("admin").
		SetCredits(10000).
		SetCreatedAt(time.Now().UTC()).
		SetUpdatedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("system: create admin: %w", err)
	}
	return admin, nil
}

func validateInitializeRequest(request InitializeRequest) error {
	parsed, err := mail.ParseAddress(request.Email)
	if err != nil || parsed.Address != request.Email || request.Username == "" || request.Password == "" || request.ConfirmPassword == "" || len(request.Password) < 8 || request.Password != request.ConfirmPassword {
		return ErrInvalidInitialize
	}
	return nil
}
