package marketplace

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/ent/mcpservice"
	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
	servicemcpgateway "github.com/shuTwT/nex-api/internal/service/mcpgateway"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

const (
	mcpDiscoveryTimeout = 15 * time.Second
	mcpDiscoveryLimit   = 2 << 20
)

type mcpRPCResponse struct {
	ID     json.RawMessage `json:"id"`
	Result struct {
		ProtocolVersion string                     `json:"protocolVersion"`
		Tools           []model.MarketplaceMCPTool `json:"tools"`
		NextCursor      string                     `json:"nextCursor"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ListMCPTools performs the MCP initialization lifecycle and returns the
// service's advertised tools without consuming user credits.
func (s *Service) ListMCPTools(ctx context.Context, id string) ([]model.MarketplaceMCPTool, error) {
	item, err := s.db.McpService.Query().Where(mcpservice.ID(id), mcpservice.IsActive(true)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("find marketplace MCP service: %w", err)
	}
	service := servicecatalog.GatewayMCPService{
		ID: item.ID, Identifier: item.Identifier, Type: item.Type,
		Endpoint: item.Endpoint, Command: item.Command, EnvVars: item.EnvVars,
		Pricing: item.Pricing, IsActive: item.IsActive,
	}
	if err := s.executor.Validate(service); err != nil {
		return nil, err
	}
	discoveryContext, cancel := context.WithTimeout(ctx, mcpDiscoveryTimeout)
	defer cancel()
	if service.Type == "stdio" {
		return s.listStdioTools(discoveryContext, service)
	}
	return s.listRemoteTools(discoveryContext, service)
}

func (s *Service) listStdioTools(ctx context.Context, service servicecatalog.GatewayMCPService) ([]model.MarketplaceMCPTool, error) {
	payload := bytes.Join([][]byte{initializeRequest(), initializedNotification(), toolsListRequest()}, []byte("\n"))
	response, err := s.executor.Invoke(ctx, service, servicemcpgateway.Invocation{Body: payload})
	if err != nil {
		return nil, err
	}
	frames, err := readMCPFrames(response.Body)
	if err != nil {
		return nil, err
	}
	return toolsFromFrames(frames)
}

func (s *Service) listRemoteTools(ctx context.Context, service servicecatalog.GatewayMCPService) ([]model.MarketplaceMCPTool, error) {
	headers := map[string][]string{
		"Accept":       {"application/json, text/event-stream"},
		"Content-Type": {"application/json"},
	}
	initialize, err := s.invokeRemote(ctx, service, headers, initializeRequest())
	if err != nil {
		return nil, err
	}
	if sessionID := headerValue(initialize.headers, "Mcp-Session-Id"); sessionID != "" {
		headers["Mcp-Session-Id"] = []string{sessionID}
	}
	initialized, err := rpcResponseFromFrames(initialize.frames, "1")
	if err != nil {
		return nil, fmt.Errorf("initialize MCP service: %w", err)
	}
	if initialized.Result.ProtocolVersion != "" {
		headers["Mcp-Protocol-Version"] = []string{initialized.Result.ProtocolVersion}
	}
	if _, err := s.invokeRemote(ctx, service, headers, initializedNotification()); err != nil {
		return nil, fmt.Errorf("finish MCP initialization: %w", err)
	}
	listed, err := s.invokeRemote(ctx, service, headers, toolsListRequest())
	if err != nil {
		return nil, err
	}
	return toolsFromFrames(listed.frames)
}

type remoteMCPResponse struct {
	headers map[string][]string
	frames  [][]byte
}

func (s *Service) invokeRemote(ctx context.Context, service servicecatalog.GatewayMCPService, headers map[string][]string, payload []byte) (remoteMCPResponse, error) {
	response, err := s.executor.Invoke(ctx, service, servicemcpgateway.Invocation{Headers: headers, Body: payload})
	if err != nil {
		return remoteMCPResponse{}, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if response.Body != nil {
			_ = response.Body.Close()
		}
		return remoteMCPResponse{}, fmt.Errorf("MCP upstream returned HTTP %d", response.StatusCode)
	}
	if response.Body == nil {
		return remoteMCPResponse{headers: response.Headers}, nil
	}
	frames, err := readMCPFrames(response.Body)
	if err != nil {
		return remoteMCPResponse{}, err
	}
	return remoteMCPResponse{headers: response.Headers, frames: frames}, nil
}

func readMCPFrames(body io.ReadCloser) ([][]byte, error) {
	defer body.Close()
	payload, err := io.ReadAll(io.LimitReader(body, mcpDiscoveryLimit+1))
	if err != nil {
		return nil, fmt.Errorf("read MCP response: %w", err)
	}
	if len(payload) > mcpDiscoveryLimit {
		return nil, errors.New("MCP discovery response is too large")
	}
	payload = bytes.TrimSpace(payload)
	if len(payload) == 0 {
		return nil, nil
	}
	if json.Valid(payload) {
		return [][]byte{payload}, nil
	}
	var frames [][]byte
	scanner := bufio.NewScanner(bytes.NewReader(payload))
	scanner.Buffer(make([]byte, 64<<10), mcpDiscoveryLimit)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if bytes.HasPrefix(line, []byte("data:")) {
			line = bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
		}
		if len(line) > 0 && json.Valid(line) {
			frames = append(frames, append([]byte(nil), line...))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan MCP response: %w", err)
	}
	if len(frames) == 0 {
		return nil, errors.New("MCP service returned no valid JSON-RPC response")
	}
	return frames, nil
}

func toolsFromFrames(frames [][]byte) ([]model.MarketplaceMCPTool, error) {
	response, err := rpcResponseFromFrames(frames, "2")
	if err != nil {
		return nil, fmt.Errorf("list MCP tools: %w", err)
	}
	if response.Result.Tools == nil {
		return []model.MarketplaceMCPTool{}, nil
	}
	return response.Result.Tools, nil
}

func rpcResponseFromFrames(frames [][]byte, id string) (mcpRPCResponse, error) {
	for _, frame := range frames {
		var response mcpRPCResponse
		if err := json.Unmarshal(frame, &response); err != nil {
			continue
		}
		if strings.Trim(string(response.ID), `"`) != id {
			continue
		}
		if response.Error != nil {
			return mcpRPCResponse{}, fmt.Errorf("JSON-RPC %d: %s", response.Error.Code, response.Error.Message)
		}
		return response, nil
	}
	return mcpRPCResponse{}, fmt.Errorf("JSON-RPC response %s is missing", id)
}

func headerValue(headers map[string][]string, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func initializeRequest() []byte {
	return []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"nex-api-marketplace","version":"1.0.0"}}}`)
}

func initializedNotification() []byte {
	return []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
}

func toolsListRequest() []byte {
	return []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
}
