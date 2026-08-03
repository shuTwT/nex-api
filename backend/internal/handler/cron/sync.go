package cron

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
)

type SyncFunc func(context.Context) error

type Config struct {
	Enabled bool
	Secret  string
}

type SyncStatsHandler struct {
	enabled bool
	secret  string
	sync    SyncFunc
}

func NewSyncStatsHandler(value any, syncFn SyncFunc) (*SyncStatsHandler, error) {
	cfg, err := normalizeConfig(value)
	if err != nil {
		return nil, err
	}
	if syncFn == nil {
		return nil, errors.New("cron: sync function is nil")
	}
	secret := strings.TrimSpace(cfg.Secret)
	if cfg.Enabled && secret == "" {
		return nil, errors.New("cron: secret is required when cron is enabled")
	}
	return &SyncStatsHandler{enabled: cfg.Enabled, secret: secret, sync: syncFn}, nil
}

func normalizeConfig(value any) (Config, error) {
	if cfg, ok := value.(Config); ok {
		return cfg, nil
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return Config{}, errors.New("cron: invalid configuration")
	}
	enabled, secret := v.FieldByName("Enabled"), v.FieldByName("Secret")
	if !enabled.IsValid() || enabled.Kind() != reflect.Bool || !secret.IsValid() || secret.Kind() != reflect.String {
		return Config{}, errors.New("cron: invalid configuration")
	}
	return Config{Enabled: enabled.Bool(), Secret: secret.String()}, nil
}

func (h *SyncStatsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost || !h.enabled {
		h.writeJSON(writer, http.StatusNotFound, false, "未找到")
		return
	}
	if !authorized(request.Header.Get("Authorization"), h.secret) {
		h.writeJSON(writer, http.StatusUnauthorized, false, "未授权")
		return
	}
	if err := h.sync(request.Context()); err != nil {
		h.writeJSON(writer, http.StatusInternalServerError, false, "数据同步失败")
		return
	}
	h.writeJSON(writer, http.StatusOK, true, "数据同步成功")
}

// RegisterRoutes exposes the cron endpoint on the application's Chi router.
func (h *SyncStatsHandler) RegisterRoutes(router chi.Router) error {
	if router == nil {
		return errors.New("cron: route router is nil")
	}
	router.Post("/api/cron/sync-stats", h.ServeHTTP)
	return nil
}

func authorized(header, secret string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	provided := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) == 1
}

func (h *SyncStatsHandler) writeJSON(writer http.ResponseWriter, status int, success bool, message string) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{"success": success, "message": message})
}
