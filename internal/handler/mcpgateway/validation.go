package mcpgateway

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shuTwT/nex-api/internal/pkg/utils/httpkit"
)

func readRequestBody(request *http.Request, maxBytes int64) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}
	if request.ContentLength > maxBytes {
		return nil, errRequestTooLarge
	}
	body, err := io.ReadAll(io.LimitReader(request.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("mcp gateway: read request body: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return nil, errRequestTooLarge
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

// errorMessages holds the human-readable msg for each error slug this handler
// emits; the numeric response code comes from constant.CodeFor.
var errorMessages = map[string]string{
	"mcp_gateway_unavailable":      "MCP gateway is unavailable",
	"invalid_api_token":            "invalid API token",
	"mcp_service_not_found":        "MCP service not found",
	"mcp_service_lookup_failed":    "MCP service lookup failed",
	"mcp_service_disabled":         "MCP service is disabled",
	"invalid_mcp_service":          "MCP service is misconfigured",
	"unsupported_mcp_service_type": "MCP service type is not supported",
	"request_body_too_large":       "request body is too large",
	"insufficient_credits":         "insufficient credits",
	"credit_reservation_failed":    "credit reservation failed",
	"mcp_upstream_failed":          "MCP upstream request failed",
}

func writeError(writer http.ResponseWriter, request *http.Request, status int, slug string) {
	message, ok := errorMessages[slug]
	if !ok {
		message = http.StatusText(status)
	}
	httpkit.WriteFail(writer, request, status, slug, message, nil)
}
