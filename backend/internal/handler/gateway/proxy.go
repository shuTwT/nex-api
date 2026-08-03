package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	servicegateway "github.com/shuTwT/nex-api/backend/internal/service/gateway"
)

func writeResponse(w http.ResponseWriter, response servicegateway.ResponseData) {
	for key, values := range response.Headers {
		w.Header().Del(key)
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	if _, err := w.Write(response.Body); err != nil {
		return
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	payload, err := json.Marshal(errorEnvelope{Success: false, Error: message, Code: code})
	if err != nil {
		h.logger.Error("gateway error response encoding failed", "err", err)
		return
	}
	if _, err := w.Write(append(payload, '\n')); err != nil {
		h.logger.Error("gateway error response write failed", "err", err)
	}
}

type errorEnvelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

func methodAllowed(configured, method string) bool {
	configured = strings.TrimSpace(strings.ToUpper(configured))
	method = strings.TrimSpace(strings.ToUpper(method))
	return configured == "ALL" || configured == method
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	return r.RemoteAddr
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
