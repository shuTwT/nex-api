package mcpgateway

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/shuTwT/nex-api/internal/infra/proxy"
	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
)

type StdioOptions struct {
	Path           string
	Shell          string
	MaxRuntime     time.Duration
	MaxMemoryBytes int64
	MaxOutputBytes int64
	MaxFrameBytes  int
	MaxInputBytes  int64
}

type Invocation struct {
	Headers map[string][]string
	Body    []byte
}

type StreamResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       io.ReadCloser
	SSE        bool
}

type Executor interface {
	Validate(servicecatalog.GatewayMCPService) error
	Invoke(context.Context, servicecatalog.GatewayMCPService, Invocation) (StreamResponse, error)
}

type ProxyExecutor struct {
	transport *proxy.MCPTransport
	options   StdioOptions
}

func NewProxyExecutor(options StdioOptions) (*ProxyExecutor, error) {
	transport, err := proxy.NewMCPTransport(proxy.StdioOptions{Path: options.Path, Shell: options.Shell, MaxRuntime: options.MaxRuntime, MaxMemoryBytes: options.MaxMemoryBytes, MaxOutputBytes: options.MaxOutputBytes, MaxFrameBytes: options.MaxFrameBytes, MaxInputBytes: options.MaxInputBytes})
	if err != nil {
		return nil, err
	}
	return &ProxyExecutor{transport: transport, options: options}, nil
}

func (e *ProxyExecutor) Validate(service servicecatalog.GatewayMCPService) error {
	_, err := e.validate(service)
	return err
}

func (e *ProxyExecutor) Invoke(ctx context.Context, service servicecatalog.GatewayMCPService, invocation Invocation) (StreamResponse, error) {
	env, err := e.validate(service)
	if err != nil {
		return StreamResponse{}, err
	}
	switch service.Type {
	case "stdio":
		stream, err := e.transport.StartStdio(ctx, service.Command, env, invocation.Body)
		if err != nil {
			return StreamResponse{}, err
		}
		return StreamResponse{StatusCode: 200, Headers: map[string][]string{"Content-Type": {"application/json; charset=utf-8"}}, Body: stream}, nil
	case "sse", "streamableHttp":
		response, err := e.transport.ForwardStreamable(ctx, service.Endpoint, invocation.Headers, invocation.Body)
		if err != nil {
			return StreamResponse{}, err
		}
		if service.Type == "sse" {
			if response.StatusCode < 200 || response.StatusCode >= 300 || response.Body == nil {
				if response.Body != nil {
					_ = response.Body.Close()
				}
				return StreamResponse{}, fmt.Errorf("mcp gateway: invalid SSE upstream response")
			}
			return StreamResponse{StatusCode: 200, Headers: map[string][]string{"Content-Type": {"application/json; charset=utf-8"}}, Body: proxy.NewSSEJSONStream(ctx, response.Body, e.options.MaxOutputBytes, e.options.MaxOutputBytes), SSE: true}, nil
		}
		return StreamResponse{StatusCode: response.StatusCode, Headers: response.Headers, Body: response.Body}, nil
	default:
		return StreamResponse{}, ErrUnsupportedType
	}
}

func (e *ProxyExecutor) validate(service servicecatalog.GatewayMCPService) (map[string]string, error) {
	switch service.Type {
	case "stdio":
		if service.Command == "" {
			return nil, ErrInvalidService
		}
		env, err := proxy.ParseServiceEnvVars(service.EnvVars)
		if err != nil {
			return nil, err
		}
		if _, err := proxy.WorkerEnvironment(e.options.Path, env); err != nil {
			return nil, err
		}
		return env, nil
	case "sse", "streamableHttp":
		if err := proxy.ValidateHTTPURL(service.Endpoint); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, ErrUnsupportedType
	}
}
