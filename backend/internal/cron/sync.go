package cron

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/config"
)

type SyncFunc func(context.Context) error

type SyncStatsHandler struct {
	enabled bool
	secret  string
	sync    SyncFunc
}

func NewSyncStatsHandler(cfg config.Cron, syncFn SyncFunc) (*SyncStatsHandler, error) {
	if syncFn == nil {
		return nil, errors.New("cron: sync function is nil")
	}
	secret := strings.TrimSpace(cfg.Secret)
	if cfg.Enabled && secret == "" {
		return nil, errors.New("cron: secret is required when cron is enabled")
	}
	return &SyncStatsHandler{enabled: cfg.Enabled, secret: secret, sync: syncFn}, nil
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

func (h *SyncStatsHandler) CronSyncStatsRoutePost(writer http.ResponseWriter, request *http.Request) {
	h.ServeHTTP(writer, request)
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
