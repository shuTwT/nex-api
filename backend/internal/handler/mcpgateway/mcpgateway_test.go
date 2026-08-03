package mcpgateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	serviceaccounts "github.com/shuTwT/nex-api/backend/internal/service/accounts"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
	servicemcp "github.com/shuTwT/nex-api/backend/internal/service/mcpgateway"
	"github.com/shuTwT/nex-api/backend/internal/service/stats"
)

type fakeAuth struct{ err error }

func (f fakeAuth) AuthenticateBearer(context.Context, string) (serviceauthz.Principal, error) {
	if f.err != nil {
		return serviceauthz.Principal{}, f.err
	}
	return serviceauthz.Principal{UserID: "user-1", Source: serviceauthz.APITokenCredential}, nil
}

type fakeCatalog struct {
	service servicecatalog.GatewayMCPService
}

func (f fakeCatalog) GatewayMCPService(context.Context, string) (servicecatalog.GatewayMCPService, error) {
	return f.service, nil
}

type fakeCredits struct {
	mu                 sync.Mutex
	calls, lastCredits int
	err                error
}

func (f *fakeCredits) Reserve(_ context.Context, _, _ string, credits int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.lastCredits = credits
	return f.err
}

type fakeStats struct{ events []stats.RequestEvent }

func (f *fakeStats) Increment(_ context.Context, event stats.RequestEvent) error {
	f.events = append(f.events, event)
	return nil
}

type fakeAudit struct{ entries []serviceaccounts.AuditEntry }

func (f *fakeAudit) Record(_ context.Context, entry serviceaccounts.AuditEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type fakeExecutor struct {
	response servicemcp.StreamResponse
	err      error
}

type recordingCatalog struct {
	service    servicecatalog.GatewayMCPService
	identifier string
}

func (f *recordingCatalog) GatewayMCPService(_ context.Context, identifier string) (servicecatalog.GatewayMCPService, error) {
	f.identifier = identifier
	return f.service, nil
}

type cancellationExecutor struct{ canceled bool }

func (f *cancellationExecutor) Validate(servicecatalog.GatewayMCPService) error { return nil }
func (f *cancellationExecutor) Invoke(ctx context.Context, _ servicecatalog.GatewayMCPService, _ servicemcp.Invocation) (servicemcp.StreamResponse, error) {
	f.canceled = errors.Is(ctx.Err(), context.Canceled)
	return servicemcp.StreamResponse{}, ctx.Err()
}

func (f fakeExecutor) Validate(servicecatalog.GatewayMCPService) error { return nil }
func (f fakeExecutor) Invoke(context.Context, servicecatalog.GatewayMCPService, servicemcp.Invocation) (servicemcp.StreamResponse, error) {
	return f.response, f.err
}
func newTestHandler(t *testing.T, service servicecatalog.GatewayMCPService, credits *fakeCredits, statsSink *fakeStats, audits *fakeAudit, response servicemcp.StreamResponse) *Handler {
	t.Helper()
	handler, err := New(Services{Authenticator: fakeAuth{}, Catalog: fakeCatalog{service}, Credits: credits, Stats: statsSink, Audits: audits}, HandlerOptions{MaxRequestBytes: 1 << 20, MaxResponseBytes: 1 << 20, MaxRuntime: 2 * time.Second, Executor: fakeExecutor{response: response}})
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
func Test_Handler_options_returns_cors_without_authentication(t *testing.T) {
	handler := newTestHandler(t, servicecatalog.GatewayMCPService{ID: "mcp-1", Identifier: "demo", Type: "streamableHttp", IsActive: true}, &fakeCredits{}, &fakeStats{}, &fakeAudit{}, servicemcp.StreamResponse{})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/api/v1/mcp/demo", nil))
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Methods") != "POST, OPTIONS" {
		t.Fatalf("options = %d %v", rec.Code, rec.Header())
	}
}
func Test_Handler_streamable_http_forwards_content_type_and_accounts_before_call(t *testing.T) {
	credits := &fakeCredits{}
	statsSink := &fakeStats{}
	audits := &fakeAudit{}
	handler := newTestHandler(t, servicecatalog.GatewayMCPService{ID: "mcp-1", Identifier: "demo", Type: "streamableHttp", Pricing: 3, IsActive: true}, credits, statsSink, audits, servicemcp.StreamResponse{StatusCode: http.StatusOK, Headers: map[string][]string{"Content-Type": {"application/json"}, "Mcp-Session-Id": {"session-1"}}, Body: io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0","id":1,"result":{}}`))})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" || rec.Header().Get("Mcp-Session-Id") != "session-1" || rec.Body.String() != `{"jsonrpc":"2.0","id":1,"result":{}}` {
		t.Fatalf("response = %d %s", rec.Code, rec.Body.String())
	}
	if credits.calls != 1 || credits.lastCredits != 3 || len(statsSink.events) != 1 || statsSink.events[0].Alias != "mcp:demo" || len(audits.entries) != 1 {
		t.Fatalf("accounting state")
	}
}
func Test_Handler_sse_converts_upstream_events_to_json_lines(t *testing.T) {
	handler := newTestHandler(t, servicecatalog.GatewayMCPService{ID: "mcp-1", Identifier: "demo", Type: "sse", IsActive: true}, &fakeCredits{}, &fakeStats{}, &fakeAudit{}, servicemcp.StreamResponse{StatusCode: http.StatusOK, Headers: map[string][]string{"Content-Type": {"application/json; charset=utf-8"}}, Body: io.NopCloser(strings.NewReader("{\"jsonrpc\":\"2.0\"}\n")), SSE: true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json; charset=utf-8" || rec.Body.String() != "{\"jsonrpc\":\"2.0\"}\n" {
		t.Fatalf("response=%d %q", rec.Code, rec.Body.String())
	}
}
func Test_Handler_stdio_returns_jsonrpc_frames(t *testing.T) {
	handler := newTestHandler(t, servicecatalog.GatewayMCPService{ID: "mcp-1", Identifier: "demo", Type: "stdio", Command: "cmd", IsActive: true}, &fakeCredits{}, &fakeStats{}, &fakeAudit{}, servicemcp.StreamResponse{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":1}\n"))})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "{\"jsonrpc\":\"2.0\",\"id\":1}\n" {
		t.Fatalf("response=%d %q", rec.Code, rec.Body.String())
	}
}
func Test_Handler_rejects_insufficient_credits_without_calling_upstream(t *testing.T) {
	credits := &fakeCredits{err: servicemcp.ErrInsufficientCredits}
	handler := newTestHandler(t, servicecatalog.GatewayMCPService{ID: "mcp-1", Identifier: "demo", Type: "streamableHttp", Pricing: 5, IsActive: true}, credits, &fakeStats{}, &fakeAudit{}, servicemcp.StreamResponse{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusPaymentRequired || credits.calls != 1 {
		t.Fatalf("response=%d calls=%d", rec.Code, credits.calls)
	}
}

func Test_Handler_ChiRouter_preserves_path_value_and_route_statuses(t *testing.T) {
	service := servicecatalog.GatewayMCPService{ID: "mcp-1", Identifier: "demo", Type: "streamableHttp", IsActive: true}
	catalog := &recordingCatalog{service: service}
	handler, err := New(Services{Authenticator: fakeAuth{}, Catalog: catalog, Credits: &fakeCredits{}}, HandlerOptions{Executor: fakeExecutor{response: servicemcp.StreamResponse{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader(""))}}})
	if err != nil {
		t.Fatal(err)
	}

	post := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", nil)
	post.Header.Set("Authorization", "Bearer token")
	postResponse := httptest.NewRecorder()
	handler.ServeHTTP(postResponse, post)
	if postResponse.Code != http.StatusNoContent || catalog.identifier != "demo" {
		t.Fatalf("post = %d identifier = %q", postResponse.Code, catalog.identifier)
	}

	pathValueRequest := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", nil)
	pathValueRequest.SetPathValue("identifier", "from-path-value")
	pathValueRequest.Header.Set("Authorization", "Bearer token")
	handler.postRoute(httptest.NewRecorder(), pathValueRequest)
	if catalog.identifier != "from-path-value" {
		t.Fatalf("path value identifier = %q", catalog.identifier)
	}

	methodResponse := httptest.NewRecorder()
	handler.ServeHTTP(methodResponse, httptest.NewRequest(http.MethodGet, "/api/v1/mcp/demo", nil))
	if methodResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status = %d", methodResponse.Code)
	}

	notFoundResponse := httptest.NewRecorder()
	handler.ServeHTTP(notFoundResponse, httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil))
	if notFoundResponse.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d", notFoundResponse.Code)
	}
}

func TestRegisterRoutes_uses_chi_router_and_propagates_cancellation(t *testing.T) {
	executor := &cancellationExecutor{}
	router := chi.NewRouter()
	service := servicecatalog.GatewayMCPService{ID: "mcp-1", Identifier: "demo", Type: "streamableHttp", IsActive: true}
	if err := RegisterRoutes(router, Services{Authenticator: fakeAuth{}, Catalog: fakeCatalog{service}, Credits: &fakeCredits{}}, HandlerOptions{Executor: executor}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/demo", nil).WithContext(ctx)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadGateway || !executor.canceled {
		t.Fatalf("response = %d canceled = %v", response.Code, executor.canceled)
	}
}
