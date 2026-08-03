package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/ent/enttest"
	serviceoauth "github.com/shuTwT/nex-api/internal/service/oauth"
)

func TestHandlerReadsGitHubOAuthAppFromSystemSettings(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/oauth.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	for key, value := range map[string]string{
		"githubOAuthClientId":     "database-client-id",
		"githubOAuthClientSecret": "database-client-secret",
	} {
		if _, err := client.SystemSetting.Create().SetKey(key).SetValue(value).SetCategory("oauth").Save(t.Context()); err != nil {
			t.Fatal(err)
		}
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
	if body := list.Body.String(); body == "" || !containsAll(body, "GitHub", "github") || containsAll(body, "database-client-secret") {
		t.Fatalf("providers response leaked or omitted expected fields: %s", body)
	}
}

func TestHandlerListsCustomOIDCProviderFromSystemSettings(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/oidc.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	payload, err := json.Marshal(serviceoauth.OIDCProviderConfig{
		Name:         "公司统一登录",
		Issuer:       "https://login.example.test",
		ClientID:     "oidc-client-id",
		ClientSecret: "oidc-client-secret",
		Scopes:       "openid profile email",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.SystemSetting.Create().SetKey("oidcProvider").SetValue(string(payload)).SetCategory("oauth").Save(t.Context()); err != nil {
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
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/auth/providers", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("providers status = %d", response.Code)
	}
	if body := response.Body.String(); body == "" || !containsAll(body, "公司统一登录", "custom-oidc") || containsAll(body, "oidc-client-secret") {
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
