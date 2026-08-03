package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const responseBufferSize = 32 << 10

func ForwardStreamableHTTP(ctx context.Context, client *http.Client, endpoint string, incoming *http.Request, body []byte) (*http.Response, error) {
	if ctx == nil {
		return nil, errors.New("mcp gateway: streamable HTTP context is nil")
	}
	if client == nil {
		return nil, errors.New("mcp gateway: HTTP client is nil")
	}
	u, err := ParseHTTPURL(endpoint)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("mcp gateway: create streamable HTTP request: %w", err)
	}
	CopyRequestHeaders(request.Header, incoming.Header)
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

func ValidateHTTPURL(endpoint string) error {
	_, err := ParseHTTPURL(endpoint)
	return err
}

func ParseHTTPURL(endpoint string) (*url.URL, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("mcp gateway: invalid streamable HTTP endpoint %q", endpoint)
	}
	return u, nil
}

func CopyRequestHeaders(destination, source http.Header) {
	for key, values := range source {
		if IsHopByHopHeader(key) || strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}

func WriteResponseBody(writer http.ResponseWriter, body io.Reader, maxBytes int64) error {
	if maxBytes <= 0 {
		return errors.New("mcp gateway: response limit must be positive")
	}
	limited := &BoundedWriter{writer: writer, limit: maxBytes}
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

func WriteUpstreamResponse(writer http.ResponseWriter, response *http.Response, maxBytes int64) error {
	defer response.Body.Close()
	CopyResponseHeaders(writer.Header(), response.Header)
	if response.ContentLength >= 0 && response.ContentLength <= maxBytes {
		writer.Header().Set("Content-Length", strconv.FormatInt(response.ContentLength, 10))
	} else {
		writer.Header().Del("Content-Length")
	}
	writer.WriteHeader(response.StatusCode)
	return WriteResponseBody(writer, response.Body, maxBytes)
}

func CopyResponseHeaders(destination, source http.Header) {
	blocked := connectionHeaders(source)
	for key, values := range source {
		lower := strings.ToLower(key)
		if IsHopByHopHeader(lower) || blocked[lower] {
			continue
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}

type BoundedWriter struct {
	writer io.Writer
	limit  int64
	total  int64
}

func (w *BoundedWriter) Write(payload []byte) (int, error) {
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
