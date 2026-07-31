package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/accounts"
	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/httpapi/generated"
	"github.com/shuTwT/nex-api/backend/internal/stats"
	"github.com/shuTwT/nex-api/backend/internal/worker"
)

type fakeAPIResolver struct{ api *ent.Api }

func (f fakeAPIResolver) GetByIdentifier(context.Context, string) (*ent.Api, error) {
	return f.api, nil
}

type fakeTokenAuthenticator struct {
	principal authz.Principal
	err       error
}

func (f fakeTokenAuthenticator) AuthenticateBearer(context.Context, string) (authz.Principal, error) {
	return f.principal, f.err
}

type fakeTransformer struct {
	mu   sync.Mutex
	jobs []worker.Job
}

func (f *fakeTransformer) Execute(_ context.Context, job worker.Job) (worker.TransformResult, error) {
	f.mu.Lock()
	f.jobs = append(f.jobs, job)
	f.mu.Unlock()
	if job.Kind == worker.ScriptKindPre {
		return worker.TransformResult{Headers: map[string]string{"X-Transformed": "yes"}, Body: json.RawMessage(`"changed"`), BodySet: true}, nil
	}
	return worker.TransformResult{ResponseBody: json.RawMessage(`"post"`), ResponseBodySet: true, ResponseHeaders: map[string]string{"X-Post": "yes"}}, nil
}

type fakeAccountant struct {
	mu        sync.Mutex
	reserved  []CreditReservation
	finalized []CreditReservation
	refunded  []CreditReservation
}

func (f *fakeAccountant) Reserve(_ context.Context, request CreditRequest) (CreditReservation, error) {
	reservation := CreditReservation{ID: "reservation-1", UserID: request.UserID, APIID: request.APIID, Credits: request.Credits}
	f.mu.Lock()
	f.reserved = append(f.reserved, reservation)
	f.mu.Unlock()
	return reservation, nil
}

func (f *fakeAccountant) Finalize(_ context.Context, reservation CreditReservation) error {
	f.mu.Lock()
	f.finalized = append(f.finalized, reservation)
	f.mu.Unlock()
	return nil
}

func (f *fakeAccountant) Refund(_ context.Context, reservation CreditReservation) error {
	f.mu.Lock()
	f.refunded = append(f.refunded, reservation)
	f.mu.Unlock()
	return nil
}

type fakeUsage struct {
	mu     sync.Mutex
	events []stats.RequestEvent
}

func (f *fakeUsage) Increment(_ context.Context, event stats.RequestEvent) error {
	f.mu.Lock()
	f.events = append(f.events, event)
	f.mu.Unlock()
	return nil
}

type fakeAudit struct {
	mu      sync.Mutex
	entries []accounts.AuditEntry
}

func (f *fakeAudit) Record(_ context.Context, entry accounts.AuditEntry) error {
	f.mu.Lock()
	f.entries = append(f.entries, entry)
	f.mu.Unlock()
	return nil
}

func newTestHandler(t *testing.T, endpoint string, method string, transformer Transformer) (*Handler, *fakeAccountant, *fakeUsage, *fakeAudit) {
	t.Helper()
	accountant := &fakeAccountant{}
	usage := &fakeUsage{}
	audit := &fakeAudit{}
	handler, err := New(Options{
		APIs:       fakeAPIResolver{api: &ent.Api{ID: "api-1", Name: "Weather", Alias: "weather", Endpoint: endpoint, Method: method, Pricing: 7, IsActive: true}},
		Tokens:     fakeTokenAuthenticator{principal: authz.Principal{UserID: "user-1"}},
		Transforms: transformer,
		Usage:      usage,
		Audit:      audit,
		Accountant: accountant,
	})
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	return handler, accountant, usage, audit
}

func TestGateway_forwardsAllMethodsAndPreservesCompletedResponse(t *testing.T) {
	var gotMethods []string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethods = append(gotMethods, r.Method)
		if r.Header.Get("Authorization") != "" || r.Header.Get("Connection") != "" {
			t.Errorf("sensitive or hop-by-hop header forwarded")
		}
		if r.URL.Query().Get("city") != "Shanghai" || len(r.URL.Query()["tag"]) != 2 {
			t.Errorf("query not forwarded: %v", r.URL.Query())
		}
		if r.Header.Get("X-Forwarded-Test") != "yes" {
			t.Errorf("end-to-end header missing")
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Upstream", "kept")
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.WriteString(w, "upstream-body")
	}))
	defer upstream.Close()
	handler, accountant, usage, _ := newTestHandler(t, upstream.URL, "ALL", nil)

	tests := []struct {
		name   string
		method string
		call   func(http.ResponseWriter, *http.Request)
	}{
		{"get", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
			handler.V1AliasRouteGet(w, r, "weather", generated.V1AliasRouteGetParams{})
		}},
		{"post", http.MethodPost, func(w http.ResponseWriter, r *http.Request) { handler.V1AliasRoutePost(w, r, "weather") }},
		{"put", http.MethodPut, func(w http.ResponseWriter, r *http.Request) { handler.V1AliasRoutePut(w, r, "weather") }},
		{"patch", http.MethodPatch, func(w http.ResponseWriter, r *http.Request) { handler.V1AliasRoutePatch(w, r, "weather") }},
		{"delete", http.MethodDelete, func(w http.ResponseWriter, r *http.Request) { handler.V1AliasRouteDelete(w, r, "weather") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := ""
			if test.method != http.MethodGet && test.method != http.MethodDelete {
				body = "payload"
			}
			req := httptest.NewRequest(test.method, upstream.URL+"?city=Shanghai&tag=a&tag=b", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer secret")
			req.Header.Set("Connection", "close")
			req.Header.Set("X-Forwarded-Test", "yes")
			req.Header.Set("Content-Type", "text/plain")
			rec := httptest.NewRecorder()

			test.call(rec, req)

			if rec.Code != http.StatusTeapot || rec.Body.String() != "upstream-body" || rec.Header().Get("X-Upstream") != "kept" {
				t.Fatalf("unexpected response: %d %q headers=%v", rec.Code, rec.Body.String(), rec.Header())
			}
		})
	}
	if strings.Join(gotMethods, ",") != "GET,POST,PUT,PATCH,DELETE" {
		t.Fatalf("methods: %v", gotMethods)
	}
	if len(accountant.finalized) != 5 || len(accountant.refunded) != 0 || len(usage.events) != 5 {
		t.Fatalf("accounting: reserved=%d finalized=%d refunded=%d usage=%d", len(accountant.reserved), len(accountant.finalized), len(accountant.refunded), len(usage.events))
	}
}

func TestGateway_transformsRequestAndResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "changed" || r.Header.Get("X-Transformed") != "yes" {
			t.Errorf("pre-transform not forwarded: body=%q headers=%v", body, r.Header)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, "upstream")
	}))
	defer upstream.Close()
	transformer := &fakeTransformer{}
	handler, accountant, _, _ := newTestHandler(t, upstream.URL, "POST", transformer)
	handler.apis = fakeAPIResolver{api: &ent.Api{ID: "api-1", Name: "Weather", Alias: "weather", Endpoint: upstream.URL, Method: "POST", Pricing: 7, IsActive: true, PreScript: "headers['X-Transformed'] = 'yes'", PostScript: "responseBody = 'post'"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/weather", strings.NewReader("original"))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	handler.V1AliasRoutePost(rec, req, "weather")

	if rec.Code != http.StatusAccepted || rec.Body.String() != "post" || rec.Header().Get("X-Post") != "yes" {
		t.Fatalf("unexpected transformed response: %d %q %v", rec.Code, rec.Body.String(), rec.Header())
	}
	if len(transformer.jobs) != 2 || len(accountant.finalized) != 1 || len(accountant.refunded) != 0 {
		t.Fatalf("transform/accounting: jobs=%d finalized=%d refunded=%d", len(transformer.jobs), len(accountant.finalized), len(accountant.refunded))
	}
}

func TestGateway_refundsNetworkFailureAndRejectsAuthAndMethod(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	endpoint := upstream.URL
	upstream.Close()
	handler, accountant, _, _ := newTestHandler(t, endpoint, "POST", nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/weather", strings.NewReader("body"))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	handler.V1AliasRoutePost(rec, req, "weather")
	if rec.Code != http.StatusBadGateway || len(accountant.finalized) != 0 || len(accountant.refunded) != 1 {
		t.Fatalf("network failure: status=%d finalized=%d refunded=%d", rec.Code, len(accountant.finalized), len(accountant.refunded))
	}

	unauthorized := httptest.NewRecorder()
	handler.tokens = fakeTokenAuthenticator{err: authz.ErrInvalidToken}
	handler.V1AliasRoutePost(unauthorized, req, "weather")
	if unauthorized.Code != http.StatusUnauthorized || len(accountant.reserved) != 1 {
		t.Fatalf("auth failure: status=%d reservations=%d", unauthorized.Code, len(accountant.reserved))
	}

	method := httptest.NewRecorder()
	handler.tokens = fakeTokenAuthenticator{principal: authz.Principal{UserID: "user-1"}}
	handler.apis = fakeAPIResolver{api: &ent.Api{ID: "api-1", Name: "Weather", Endpoint: endpoint, Method: "GET", Pricing: 7, IsActive: true}}
	handler.V1AliasRoutePost(method, req, "weather")
	if method.Code != http.StatusMethodNotAllowed || len(accountant.reserved) != 1 {
		t.Fatalf("method failure: status=%d reservations=%d", method.Code, len(accountant.reserved))
	}
}
