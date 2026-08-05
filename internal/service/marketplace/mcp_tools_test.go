package marketplace

import (
	"context"
	"io"
	"strings"
	"testing"

	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
	servicemcpgateway "github.com/shuTwT/nex-api/internal/service/mcpgateway"
)

type discoveryExecutor struct {
	responses   []servicemcpgateway.StreamResponse
	invocations []servicemcpgateway.Invocation
}

func (e *discoveryExecutor) Validate(servicecatalog.GatewayMCPService) error { return nil }

func (e *discoveryExecutor) Invoke(_ context.Context, _ servicecatalog.GatewayMCPService, invocation servicemcpgateway.Invocation) (servicemcpgateway.StreamResponse, error) {
	headers := make(map[string][]string, len(invocation.Headers))
	for key, values := range invocation.Headers {
		headers[key] = append([]string(nil), values...)
	}
	e.invocations = append(e.invocations, servicemcpgateway.Invocation{Headers: headers, Body: append([]byte(nil), invocation.Body...)})
	response := e.responses[0]
	e.responses = e.responses[1:]
	return response, nil
}

func TestListStdioToolsUsesOneInitializedSession(t *testing.T) {
	executor := &discoveryExecutor{responses: []servicemcpgateway.StreamResponse{{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(
			`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-06-18"}}` + "\n" +
				`{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"search","description":"Search documents","inputSchema":{"type":"object"}}]}}` + "\n",
		)),
	}}}
	service := &Service{executor: executor}

	tools, err := service.listStdioTools(context.Background(), servicecatalog.GatewayMCPService{Type: "stdio"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name != "search" {
		t.Fatalf("tools = %#v", tools)
	}
	if len(executor.invocations) != 1 || !strings.Contains(string(executor.invocations[0].Body), `"method":"notifications/initialized"`) {
		t.Fatalf("invocations = %#v", executor.invocations)
	}
}

func TestListRemoteToolsCarriesSessionHeaderAndParsesSSE(t *testing.T) {
	executor := &discoveryExecutor{responses: []servicemcpgateway.StreamResponse{
		{StatusCode: 200, Headers: map[string][]string{"Mcp-Session-Id": {"session-1"}}, Body: io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-06-18"}}`))},
		{StatusCode: 202, Body: io.NopCloser(strings.NewReader(""))},
		{StatusCode: 200, Body: io.NopCloser(strings.NewReader("event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{\"tools\":[{\"name\":\"create_page\"}]}}\n\n"))},
	}}
	service := &Service{executor: executor}

	tools, err := service.listRemoteTools(context.Background(), servicecatalog.GatewayMCPService{Type: "streamableHttp"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name != "create_page" {
		t.Fatalf("tools = %#v", tools)
	}
	if len(executor.invocations) != 3 || headerValue(executor.invocations[2].Headers, "Mcp-Session-Id") != "session-1" {
		t.Fatalf("session header was not propagated: %#v", executor.invocations)
	}
	if headerValue(executor.invocations[2].Headers, "Mcp-Protocol-Version") != "2025-06-18" {
		t.Fatalf("protocol version header was not propagated: %#v", executor.invocations)
	}
}
