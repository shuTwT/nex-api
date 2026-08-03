package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

type StreamableResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       io.ReadCloser
}

type MCPTransport struct {
	client *http.Client
	stdio  *StdioRunner
}

func NewMCPTransport(options StdioOptions) (*MCPTransport, error) {
	stdio, err := NewStdioRunner(options)
	if err != nil {
		return nil, err
	}
	return &MCPTransport{client: NewHTTPClient(), stdio: stdio}, nil
}

func (t *MCPTransport) StartStdio(ctx context.Context, command string, env map[string]string, body []byte) (io.ReadCloser, error) {
	return t.stdio.Start(ctx, command, env, body)
}

func (t *MCPTransport) ForwardStreamable(ctx context.Context, endpoint string, headers map[string][]string, body []byte) (StreamableResponse, error) {
	u, err := ParseHTTPURL(endpoint)
	if err != nil {
		return StreamableResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return StreamableResponse{}, fmt.Errorf("mcp gateway: create streamable HTTP request: %w", err)
	}
	for key, values := range headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	request.Header.Del("Authorization")
	if request.Header.Get("Content-Type") == "" && len(body) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := t.client.Do(request)
	if err != nil {
		return StreamableResponse{}, fmt.Errorf("mcp gateway: call streamable HTTP endpoint: %w", err)
	}
	responseHeaders := make(map[string][]string, len(response.Header))
	for key, values := range response.Header {
		responseHeaders[key] = append([]string(nil), values...)
	}
	return StreamableResponse{StatusCode: response.StatusCode, Headers: responseHeaders, Body: response.Body}, nil
}
