package mcpgateway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

const responseBufferSize = 32 << 10

func forwardStreamableHTTP(ctx context.Context, client *http.Client, endpoint string, incoming *http.Request, body []byte) (*http.Response, error) {
	if ctx == nil {
		return nil, errors.New("mcp gateway: streamable HTTP context is nil")
	}
	if client == nil {
		return nil, errors.New("mcp gateway: HTTP client is nil")
	}
	u, err := parseHTTPURL(endpoint)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("mcp gateway: create streamable HTTP request: %w", err)
	}
	copyRequestHeaders(request.Header, incoming.Header)
	request.Header.Del("Authorization")
	if request.Header.Get("Content-Type") == "" && len(body) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("mcp gateway: call streamable HTTP endpoint: %w", err)
	}
	return response, nil
}

func validateHTTPURL(endpoint string) error {
	_, err := parseHTTPURL(endpoint)
	return err
}

func parseHTTPURL(endpoint string) (*url.URL, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("mcp gateway: invalid streamable HTTP endpoint %q", endpoint)
	}
	return u, nil
}

func copyRequestHeaders(destination, source http.Header) {
	for key, values := range source {
		if hopByHopHeader(key) || strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}

func writeResponseBody(writer http.ResponseWriter, body io.Reader, maxBytes int64) error {
	if maxBytes <= 0 {
		return errors.New("mcp gateway: response limit must be positive")
	}
	limited := &boundedWriter{writer: writer, limit: maxBytes}
	buffer := make([]byte, responseBufferSize)
	flusher, canFlush := writer.(http.Flusher)
	for {
		read, readErr := body.Read(buffer)
		if read > 0 {
			if _, writeErr := limited.Write(buffer[:read]); writeErr != nil {
				return fmt.Errorf("mcp gateway: copy response body: %w", writeErr)
			}
			if canFlush {
				flusher.Flush()
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("mcp gateway: read response body: %w", readErr)
		}
	}
}

func writeUpstreamResponse(writer http.ResponseWriter, response *http.Response, maxBytes int64) error {
	defer response.Body.Close()
	copyResponseHeaders(writer.Header(), response.Header)
	if response.ContentLength >= 0 && response.ContentLength <= maxBytes {
		writer.Header().Set("Content-Length", strconv.FormatInt(response.ContentLength, 10))
	} else {
		writer.Header().Del("Content-Length")
	}
	writer.WriteHeader(response.StatusCode)
	return writeResponseBody(writer, response.Body, maxBytes)
}

func (h *Handler) serveStdio(writer http.ResponseWriter, ctx context.Context, service *ent.McpService, envVars map[string]string, body []byte) {
	stream, err := h.stdio.Start(ctx, service.Command, envVars, body)
	if err != nil {
		writeError(writer, http.StatusBadGateway, "stdio_worker_failed")
		return
	}
	defer stream.Close()
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if err := writeResponseBody(writer, stream, h.options.MaxResponseBytes); err != nil {
		h.logger.WarnContext(ctx, "MCP stdio stream failed", slog.String("identifier", service.Identifier), slog.Any("err", err))
	}
}

func (h *Handler) serveStreamable(writer http.ResponseWriter, ctx context.Context, request *http.Request, service *ent.McpService, body []byte) {
	response, err := forwardStreamableHTTP(ctx, h.options.HTTPClient, service.Endpoint, request, body)
	if err != nil {
		writeError(writer, http.StatusBadGateway, "streamable_http_upstream_failed")
		return
	}
	if err := writeUpstreamResponse(writer, response, h.options.MaxResponseBytes); err != nil {
		h.logger.WarnContext(ctx, "MCP streamable HTTP response failed", slog.String("identifier", service.Identifier), slog.Any("err", err))
	}
}

func copyResponseHeaders(destination, source http.Header) {
	for key, values := range source {
		if hopByHopHeader(key) || strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}

func hopByHopHeader(key string) bool {
	switch strings.ToLower(key) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

type boundedWriter struct {
	writer io.Writer
	limit  int64
	total  int64
}

func (w *boundedWriter) Write(payload []byte) (int, error) {
	remaining := w.limit - w.total
	if remaining <= 0 {
		return 0, ErrOutputLimit
	}
	if int64(len(payload)) > remaining {
		written, err := w.writer.Write(payload[:int(remaining)])
		w.total += int64(written)
		if err != nil {
			return written, err
		}
		return written, ErrOutputLimit
	}
	written, err := w.writer.Write(payload)
	w.total += int64(written)
	return written, err
}

var _ http.Handler = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
