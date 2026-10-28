package gateway

import (
	"io"
	"net/http"
	"strings"

	"github.com/shuTwT/nex-api/internal/pkg/utils/httpkit"
	servicegateway "github.com/shuTwT/nex-api/internal/service/gateway"
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

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	httpkit.WriteFail(w, r, status, code, message, nil)
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
