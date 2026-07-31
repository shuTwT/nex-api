package mcpgateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

func (h *Handler) validateService(service *ent.McpService) (map[string]string, error) {
	switch service.Type {
	case "stdio":
		if strings.TrimSpace(service.Command) == "" {
			return nil, ErrInvalidService
		}
		envVars, err := parseServiceEnvVars(service.EnvVars)
		if err != nil {
			return nil, err
		}
		if _, err := workerEnvironment(h.options.Stdio.Path, envVars); err != nil {
			return nil, err
		}
		return envVars, nil
	case "sse", "streamableHttp":
		if err := validateHTTPURL(service.Endpoint); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, ErrUnsupportedType
	}
}

func readRequestBody(request *http.Request, maxBytes int64) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}
	if request.ContentLength > maxBytes {
		return nil, ErrInputLimit
	}
	body, err := io.ReadAll(io.LimitReader(request.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("mcp gateway: read request body: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return nil, ErrInputLimit
	}
	return body, nil
}

func clientIP(request *http.Request) string {
	if forwarded := request.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	return request.RemoteAddr
}

func setCORSHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Mcp-Session-Id, Mcp-Protocol-Version")
}

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(errorResponse{Success: false, Error: message})
}
