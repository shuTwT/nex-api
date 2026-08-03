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

	"github.com/shuTwT/nex-api/internal/infra/worker"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
	serviceauthz "github.com/shuTwT/nex-api/internal/service/authz"
	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
	servicegateway "github.com/shuTwT/nex-api/internal/service/gateway"
	"github.com/shuTwT/nex-api/internal/service/stats"
)

type fakeAPIResolver struct{ api servicecatalog.GatewayAPI }

func (f fakeAPIResolver) GatewayAPI(context.Context, string) (servicecatalog.GatewayAPI, error) {
	return f.api, nil
}

type fakeTokenAuthenticator struct {
	principal serviceauthz.Principal
	err       error
}

func (f fakeTokenAuthenticator) AuthenticateBearer(context.Context, string) (serviceauthz.Principal, error) {
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
	mu                            sync.Mutex
	reserved, finalized, refunded []servicegateway.CreditReservation
}

func (f *fakeAccountant) Reserve(_ context.Context, r servicegateway.CreditRequest) (servicegateway.CreditReservation, error) {
	v := servicegateway.CreditReservation{ID: "reservation-1", UserID: r.UserID, APIID: r.APIID, Credits: r.Credits}
	f.mu.Lock()
	f.reserved = append(f.reserved, v)
	f.mu.Unlock()
	return v, nil
}
func (f *fakeAccountant) Finalize(_ context.Context, v servicegateway.CreditReservation) error {
	f.mu.Lock()
	f.finalized = append(f.finalized, v)
	f.mu.Unlock()
	return nil
}
func (f *fakeAccountant) Refund(_ context.Context, v servicegateway.CreditReservation) error {
	f.mu.Lock()
	f.refunded = append(f.refunded, v)
	f.mu.Unlock()
	return nil
}

type fakeUsage struct {
	mu     sync.Mutex
	events []stats.RequestEvent
}

func (f *fakeUsage) Increment(_ context.Context, e stats.RequestEvent) error {
	f.mu.Lock()
	f.events = append(f.events, e)
	f.mu.Unlock()
	return nil
}

type fakeAudit struct {
	mu      sync.Mutex
	entries []serviceaccounts.AuditEntry
}

func (f *fakeAudit) Record(_ context.Context, e serviceaccounts.AuditEntry) error {
	f.mu.Lock()
	f.entries = append(f.entries, e)
	f.mu.Unlock()
	return nil
}
func newTestHandler(t *testing.T, endpoint, method string, transformer servicegateway.Transformer) (*Handler, *fakeAccountant, *fakeUsage, *fakeAudit) {
	t.Helper()
	accountant := &fakeAccountant{}
	usage := &fakeUsage{}
	audit := &fakeAudit{}
	handler, err := New(Options{APIs: fakeAPIResolver{servicecatalog.GatewayAPI{ID: "api-1", Name: "Weather", Endpoint: endpoint, Method: method, Pricing: 7, IsActive: true}}, Tokens: fakeTokenAuthenticator{principal: serviceauthz.Principal{UserID: "user-1"}}, Transforms: transformer, Usage: usage, Audit: audit, Accountant: accountant})
	if err != nil {
		t.Fatal(err)
	}
	return handler, accountant, usage, audit
}
func TestGateway_forwardsAllMethodsAndPreservesCompletedResponse(t *testing.T) {
	var methods []string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		if r.Header.Get("Authorization") != "" || r.Header.Get("Connection") != "" {
			t.Error("sensitive header forwarded")
		}
		if r.URL.Query().Get("city") != "Shanghai" || len(r.URL.Query()["tag"]) != 2 {
			t.Errorf("query not forwarded: %v", r.URL.Query())
		}
		if r.Header.Get("X-Forwarded-Test") != "yes" {
			t.Error("header missing")
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Upstream", "kept")
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.WriteString(w, "upstream-body")
	}))
	defer upstream.Close()
	handler, accountant, usage, _ := newTestHandler(t, upstream.URL, "ALL", nil)
	for _, test := range []struct {
		name, method string
	}{{"get", http.MethodGet}, {"post", http.MethodPost}, {"put", http.MethodPut}, {"patch", http.MethodPatch}, {"delete", http.MethodDelete}} {
		t.Run(test.name, func(t *testing.T) {
			body := ""
			if test.method != http.MethodGet && test.method != http.MethodDelete {
				body = "payload"
			}
			req := httptest.NewRequest(test.method, upstream.URL+"?city=Shanghai&tag=a&tag=b", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer secret")
			req.Header.Set("Connection", "close")
			req.Header.Set("X-Forwarded-Test", "yes")
			rec := httptest.NewRecorder()
			handler.serve(rec, req, "weather")
			if rec.Code != http.StatusTeapot || rec.Body.String() != "upstream-body" || rec.Header().Get("X-Upstream") != "kept" {
				t.Fatalf("unexpected response: %d %q", rec.Code, rec.Body.String())
			}
		})
	}
	if strings.Join(methods, ",") != "GET,POST,PUT,PATCH,DELETE" {
		t.Fatalf("methods: %v", methods)
	}
	if len(accountant.finalized) != 5 || len(accountant.refunded) != 0 || len(usage.events) != 5 {
		t.Fatalf("accounting state")
	}
}
func TestGateway_transformsRequestAndResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "changed" || r.Header.Get("X-Transformed") != "yes" {
			t.Errorf("pre-transform not forwarded")
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, "upstream")
	}))
	defer upstream.Close()
	transformer := &fakeTransformer{}
	handler, accountant, _, _ := newTestHandler(t, upstream.URL, "POST", transformer)
	handler.apis = fakeAPIResolver{servicecatalog.GatewayAPI{ID: "api-1", Name: "Weather", Endpoint: upstream.URL, Method: "POST", Pricing: 7, IsActive: true, PreScript: "x", PostScript: "x"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/weather", strings.NewReader("original"))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	handler.serve(rec, req, "weather")
	if rec.Code != http.StatusAccepted || rec.Body.String() != "post" || rec.Header().Get("X-Post") != "yes" {
		t.Fatalf("unexpected transformed response: %d %q", rec.Code, rec.Body.String())
	}
	if len(transformer.jobs) != 2 || len(accountant.finalized) != 1 || len(accountant.refunded) != 0 {
		t.Fatalf("transform/accounting mismatch")
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
	handler.serve(rec, req, "weather")
	if rec.Code != http.StatusBadGateway || len(accountant.finalized) != 0 || len(accountant.refunded) != 1 {
		t.Fatalf("network failure")
	}
	unauthorized := httptest.NewRecorder()
	handler.tokens = fakeTokenAuthenticator{err: serviceauthz.ErrInvalidToken}
	handler.serve(unauthorized, req, "weather")
	if unauthorized.Code != http.StatusUnauthorized || len(accountant.reserved) != 1 {
		t.Fatalf("auth failure")
	}
	method := httptest.NewRecorder()
	handler.tokens = fakeTokenAuthenticator{principal: serviceauthz.Principal{UserID: "user-1"}}
	handler.apis = fakeAPIResolver{servicecatalog.GatewayAPI{ID: "api-1", Name: "Weather", Endpoint: endpoint, Method: "GET", Pricing: 7, IsActive: true}}
	handler.serve(method, req, "weather")
	if method.Code != http.StatusMethodNotAllowed || len(accountant.reserved) != 1 {
		t.Fatalf("method failure")
	}
}
