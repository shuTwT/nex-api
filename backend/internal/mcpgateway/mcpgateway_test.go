package mcpgateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

func Test_StdioRunner_forwards_bounded_jsonrpc_without_inherited_environment(t *testing.T) {
	// Given
	t.Setenv("MCP_GATEWAY_SECRET", "must-not-leak")
	runner, err := NewStdioRunner(StdioOptions{
		Path:           "/usr/local/bin:/usr/bin:/bin",
		MaxRuntime:     2 * time.Second,
		MaxMemoryBytes: 64 << 20,
		MaxOutputBytes: 4 << 10,
		MaxFrameBytes:  2 << 10,
	})
	requireNoError(t, err)
	command := `if [ "$MCP_ALLOWED" = "yes" ] && [ -z "$MCP_GATEWAY_SECRET" ]; then printf '%s\n' '{"jsonrpc":"2.0","id":1,"result":{"ok":true}}'; else printf '%s\n' '{"jsonrpc":"2.0","id":1,"error":{"message":"bad env"}}'; fi`

	// When
	stream, err := runner.Start(context.Background(), command, map[string]string{"MCP_ALLOWED": "yes"}, []byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	requireNoError(t, err)
	got, readErr := io.ReadAll(stream)
	closeErr := stream.Close()

	// Then
	requireNoError(t, readErr)
	requireNoError(t, closeErr)
	if string(got) != "{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"ok\":true}}\n" {
		t.Fatalf("unexpected JSON-RPC frame: %q", got)
	}
}

func Test_StdioRunner_rejects_output_over_limit(t *testing.T) {
	// Given
	runner, err := NewStdioRunner(StdioOptions{
		MaxRuntime:     2 * time.Second,
		MaxMemoryBytes: 64 << 20,
		MaxOutputBytes: 16,
		MaxFrameBytes:  64,
	})
	requireNoError(t, err)

	// When
	stream, err := runner.Start(context.Background(), `printf '%s\n' '{"jsonrpc":"2.0","result":"this-is-too-large"}'`, nil, nil)
	requireNoError(t, err)
	_, readErr := io.ReadAll(stream)
	_ = stream.Close()

	// Then
	if !errors.Is(readErr, ErrOutputLimit) {
		t.Fatalf("expected output limit error, got %v", readErr)
	}
}

func Test_StdioRunner_stops_worker_when_context_is_canceled(t *testing.T) {
	// Given
	runner, err := NewStdioRunner(StdioOptions{MaxRuntime: 5 * time.Second, MaxMemoryBytes: 256 << 20, MaxOutputBytes: 1 << 10, MaxFrameBytes: 1 << 10})
	requireNoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := runner.Start(ctx, `while :; do :; done`, nil, nil)
	requireNoError(t, err)

	// When
	cancel()
	readDone := make(chan error, 1)
	go func() {
		_, readErr := io.ReadAll(stream)
		readDone <- readErr
	}()

	// Then
	select {
	case <-readDone:
	case <-time.After(2 * time.Second):
		t.Fatal("canceled stdio worker did not stop")
	}
	requireNoError(t, stream.Close())
}

func Test_SSEJSONStream_converts_data_events_and_honors_done(t *testing.T) {
	// Given
	body := io.NopCloser(strings.NewReader("event: message\ndata: {\"jsonrpc\":\"2.0\"}\n\ndata: [DONE]\n\n"))

	// When
	stream := NewSSEJSONStream(context.Background(), body, 1<<10, 1<<10)
	got, err := io.ReadAll(stream)
	requireNoError(t, err)
	requireNoError(t, stream.Close())

	// Then
	if string(got) != "{\"jsonrpc\":\"2.0\"}\n" {
		t.Fatalf("unexpected converted SSE: %q", got)
	}
}

func Test_Handler_options_returns_cors_without_authentication(t *testing.T) {
	// Given
	handler := newTestHandler(t, &ent.McpService{ID: "mcp-1", Name: "Demo", Identifier: "demo", Type: "streamableHttp", IsActive: true})
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/mcp/demo", nil)
	recorder := httptest.NewRecorder()

	// When
	handler.ServeHTTP(recorder, req)

	// Then
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS" {
		t.Fatalf("unexpected allowed methods: %q", got)
	}
}

func Test_Handler_streamable_http_forwards_content_type_and_accounts_before_call(t *testing.T) {
	// Given
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Mcp-Session-Id", "session-1")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`))
	}))
	t.Cleanup(upstream.Close)
	service := &ent.McpService{ID: "mcp-1", Name: "Demo", Identifier: "demo", Type: "streamableHttp", Endpoint: upstream.URL, Pricing: 3, IsActive: true}
	credits := &fakeCredits{}
	statsSink := &fakeStats{}
	auditSink := &fakeAudit{}
	handler := newTestHandlerWithServices(t, service, credits, statsSink, auditSink)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	// When
	handler.ServeHTTP(recorder, req)

	// Then
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if got := recorder.Header().Get("Mcp-Session-Id"); got != "session-1" {
		t.Fatalf("upstream header was not preserved: %q", got)
	}
	if got := recorder.Body.String(); got != `{"jsonrpc":"2.0","id":1,"result":{}}` {
		t.Fatalf("unexpected response body: %q", got)
	}
	if credits.calls != 1 || credits.lastCredits != 3 {
		t.Fatalf("credits were not reserved before transport: %+v", credits)
	}
	if len(statsSink.events) != 1 || statsSink.events[0].Alias != "mcp:demo" {
		t.Fatalf("stats were not recorded: %+v", statsSink.events)
	}
	if len(auditSink.entries) != 1 || auditSink.entries[0].Status != "success" {
		t.Fatalf("audit was not recorded: %+v", auditSink.entries)
	}
}

func Test_Handler_sse_converts_upstream_events_to_json_lines(t *testing.T) {
	// Given
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"jsonrpc\":\"2.0\"}\n\ndata: [DONE]\n\n")
	}))
	t.Cleanup(upstream.Close)
	service := &ent.McpService{ID: "mcp-1", Name: "Demo", Identifier: "demo", Type: "sse", Endpoint: upstream.URL, IsActive: true}
	handler := newTestHandler(t, service)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", strings.NewReader(`{"jsonrpc":"2.0"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()

	// When
	handler.ServeHTTP(recorder, req)

	// Then
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if got := recorder.Body.String(); got != "{\"jsonrpc\":\"2.0\"}\n" {
		t.Fatalf("unexpected SSE conversion: %q", got)
	}
}

func Test_Handler_stdio_returns_jsonrpc_frames(t *testing.T) {
	// Given
	service := &ent.McpService{ID: "mcp-1", Name: "Demo", Identifier: "demo", Type: "stdio", Command: `printf '%s\n' '{"jsonrpc":"2.0","id":1}'`, IsActive: true}
	handler := newTestHandler(t, service)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", strings.NewReader(`{"jsonrpc":"2.0","id":1}`))
	req.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()

	// When
	handler.ServeHTTP(recorder, req)

	// Then
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Body.String(); got != "{\"jsonrpc\":\"2.0\",\"id\":1}\n" {
		t.Fatalf("unexpected stdio response: %q", got)
	}
}

func Test_StreamableForwarding_propagates_context_cancellation(t *testing.T) {
	// Given
	entered := make(chan struct{})
	client := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		close(entered)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		response, err := forwardStreamableHTTP(ctx, client, "http://upstream.test/mcp", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`)), []byte(`{}`))
		if response != nil {
			response.Body.Close()
		}
		result <- err
	}()

	// When
	select {
	case <-entered:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("upstream did not receive request")
	}

	// Then
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("expected cancellation error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("forwarding did not return after cancellation")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func Test_Handler_rejects_insufficient_credits_without_calling_upstream(t *testing.T) {
	// Given
	called := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(upstream.Close)
	service := &ent.McpService{ID: "mcp-1", Name: "Demo", Identifier: "demo", Type: "streamableHttp", Endpoint: upstream.URL, Pricing: 5, IsActive: true}
	credits := &fakeCredits{err: ErrInsufficientCredits}
	handler := newTestHandlerWithServices(t, service, credits, &fakeStats{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", strings.NewReader(`{"jsonrpc":"2.0"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()

	// When
	handler.ServeHTTP(recorder, req)

	// Then
	if recorder.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", recorder.Code)
	}
	if called {
		t.Fatal("upstream was called despite insufficient credits")
	}
}
