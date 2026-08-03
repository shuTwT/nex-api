package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/ent/enttest"
	serviceoauth "github.com/shuTwT/nex-api/backend/internal/service/oauth"
)

func TestHandlerReadsOAuthProvidersFromSystemSettings(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/oauth.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	providers := []serviceoauth.ProviderConfig{{ID: "github", Name: "数据库 GitHub", ClientID: "database-client-id", ClientSecret: "database-client-secret", AuthorizationURL: "https://oauth.example.test/authorize", TokenURL: "https://oauth.example.test/token", UserInfoURL: "https://oauth.example.test/userinfo", Scopes: "read:user,user:email", UserIDField: "id", EmailField: "email", UsernameField: "login"}}
	payload, err := json.Marshal(providers)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.SystemSetting.Create().SetKey("oauthProviders").SetValue(string(payload)).SetCategory("oauth").Save(t.Context()); err != nil {
		t.Fatal(err)
	}
	service, err := serviceoauth.NewService(client)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := New(service, nil, Config{AppURL: "https://nex.example.test", SessionSecret: []byte("test-session-secret")})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/auth/signin/github?callbackUrl=/console", nil))
	if response.Code != http.StatusFound {
		t.Fatalf("authorize status = %d, body=%s", response.Code, response.Body.String())
	}
	redirect, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if got := redirect.Query().Get("client_id"); got != "database-client-id" {
		t.Fatalf("authorization client_id = %q", got)
	}
	if got := redirect.Query().Get("scope"); got != "read:user user:email" {
		t.Fatalf("authorization scope = %q", got)
	}
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/auth/providers", nil))
	if list.Code != http.StatusOK {
		t.Fatalf("providers status = %d", list.Code)
	}
	if body := list.Body.String(); body == "" || !containsAll(body, "数据库 GitHub", "github") || containsAll(body, "database-client-secret") {
		t.Fatalf("providers response leaked or omitted expected fields: %s", body)
	}
}
func containsAll(value string, expected ...string) bool {
	for _, item := range expected {
		if !strings.Contains(value, item) {
			return false
		}
	}
	return true
}
