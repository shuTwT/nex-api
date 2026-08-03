package proxy

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_StdioRunner_forwards_bounded_jsonrpc_without_inherited_environment(t *testing.T) {
	t.Setenv("MCP_GATEWAY_SECRET", "must-not-leak")
	runner, err := NewStdioRunner(StdioOptions{Path: "/usr/local/bin:/usr/bin:/bin", MaxRuntime: 2 * time.Second, MaxMemoryBytes: 64 << 20, MaxOutputBytes: 4 << 10, MaxFrameBytes: 2 << 10})
	if err != nil {
		t.Fatal(err)
	}
	stream, err := runner.Start(context.Background(), `if [ "$MCP_ALLOWED" = "yes" ] && [ -z "$MCP_GATEWAY_SECRET" ]; then printf '%s\n' '{"jsonrpc":"2.0","id":1,"result":{"ok":true}}'; else printf '%s\n' '{"jsonrpc":"2.0","id":1,"error":{"message":"bad env"}}'; fi`, map[string]string{"MCP_ALLOWED": "yes"}, []byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	if err != nil {
		t.Fatal(err)
	}
	got, readErr := io.ReadAll(stream)
	closeErr := stream.Close()
	if readErr != nil || closeErr != nil {
		t.Fatalf("read=%v close=%v", readErr, closeErr)
	}
	if string(got) != "{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"ok\":true}}\n" {
		t.Fatalf("frame = %q", got)
	}
}
func Test_StdioRunner_rejects_output_over_limit(t *testing.T) {
	runner, err := NewStdioRunner(StdioOptions{MaxRuntime: 2 * time.Second, MaxMemoryBytes: 64 << 20, MaxOutputBytes: 16, MaxFrameBytes: 64})
	if err != nil {
		t.Fatal(err)
	}
	stream, err := runner.Start(context.Background(), `printf '%s\n' '{"jsonrpc":"2.0","result":"this-is-too-large"}'`, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, readErr := io.ReadAll(stream)
	_ = stream.Close()
	if !errors.Is(readErr, ErrOutputLimit) {
		t.Fatalf("error = %v", readErr)
	}
}
func Test_StdioRunner_stops_worker_when_context_is_canceled(t *testing.T) {
	runner, err := NewStdioRunner(StdioOptions{MaxRuntime: 5 * time.Second, MaxMemoryBytes: 256 << 20, MaxOutputBytes: 1 << 10, MaxFrameBytes: 1 << 10})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := runner.Start(ctx, `while :; do :; done`, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	done := make(chan error, 1)
	go func() { _, err := io.ReadAll(stream); done <- err }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop")
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
}
func Test_SSEJSONStream_converts_data_events_and_honors_done(t *testing.T) {
	stream := NewSSEJSONStream(context.Background(), io.NopCloser(strings.NewReader("event: message\ndata: {\"jsonrpc\":\"2.0\"}\n\ndata: [DONE]\n\n")), 1<<10, 1<<10)
	got, err := io.ReadAll(stream)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
	if string(got) != "{\"jsonrpc\":\"2.0\"}\n" {
		t.Fatalf("stream = %q", got)
	}
}
func Test_StreamableForwarding_propagates_context_cancellation(t *testing.T) {
	entered := make(chan struct{})
	client := &http.Client{Transport: cancellationRoundTripper(func(request *http.Request) (*http.Response, error) {
		close(entered)
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		response, err := ForwardStreamableHTTP(ctx, client, "http://upstream.test/mcp", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`)), []byte(`{}`))
		if response != nil {
			_ = response.Body.Close()
		}
		result <- err
	}()
	select {
	case <-entered:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("upstream not called")
	}
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("forward did not return")
	}
}

type cancellationRoundTripper func(*http.Request) (*http.Response, error)

func (f cancellationRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
