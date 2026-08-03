package oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/config"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/enttest"
)

func TestResolveReturnURL_rejectsExternalTargets(t *testing.T) {
	// Given
	baseURL := "https://nex.example.com"

	// When
	got, err := ResolveReturnURL("https://attacker.example/steal", baseURL)

	// Then
	if err == nil {
		t.Fatal("ResolveReturnURL accepted an external target")
	}
	if got != "" {
		t.Fatalf("ResolveReturnURL returned %q for an invalid target", got)
	}
}

func TestStateCodec_roundTripsPKCEAndNonce(t *testing.T) {
	// Given
	codec, err := NewStateCodec([]byte("test-state-secret"))
	if err != nil {
		t.Fatalf("NewStateCodec returned an error: %v", err)
	}
	state, err := codec.New("easy1auth", "https://nex.example.com/console", true)
	if err != nil {
		t.Fatalf("StateCodec.New returned an error: %v", err)
	}
	response := httptest.NewRecorder()
	codec.Write(response, state)

	// When
	request := httptest.NewRequest("GET", "/api/auth/callback/easy1auth?state="+state.Value, nil)
	for _, cookie := range response.Result().Cookies() {
		request.AddCookie(cookie)
	}
	got, err := codec.Read(request, "easy1auth", state.Value)

	// Then
	if err != nil {
		t.Fatalf("StateCodec.Read returned an error: %v", err)
	}
	if got.CodeVerifier == "" || got.Nonce == "" {
		t.Fatalf("state lost PKCE or nonce material: %+v", got)
	}
	if got.ReturnURL != state.ReturnURL || got.Provider != state.Provider {
		t.Fatalf("state round trip mismatch: got %+v want %+v", got, state)
	}
	if state.CodeChallenge == "" {
		t.Fatal("state did not expose a PKCE challenge")
	}
}

func makeTestIDToken(t *testing.T, nonce string) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "none"})
	if err != nil {
		t.Fatalf("marshal ID token header: %v", err)
	}
	payload, err := json.Marshal(map[string]string{"nonce": nonce})
	if err != nil {
		t.Fatalf("marshal ID token payload: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}

func TestValidateOIDCNonce_rejectsMismatchedNonce(t *testing.T) {
	// Given
	idToken := makeTestIDToken(t, "expected-nonce")

	// When
	err := ValidateOIDCNonce(idToken, "different-nonce")

	// Then
	if err == nil {
		t.Fatal("ValidateOIDCNonce accepted a mismatched nonce")
	}
}

func TestHandlerReadsOAuthProvidersFromSystemSettings(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/oauth.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	providers := []providerConfig{{
		ID:               "github",
		Name:             "数据库 GitHub",
		ClientID:         "database-client-id",
		ClientSecret:     "database-client-secret",
		AuthorizationURL: "https://oauth.example.test/authorize",
		TokenURL:         "https://oauth.example.test/token",
		UserInfoURL:      "https://oauth.example.test/userinfo",
		Scopes:           "read:user,user:email",
		UserIDField:      "id",
		EmailField:       "email",
		UsernameField:    "login",
	}}
	payload, err := json.Marshal(providers)
	if err != nil {
		t.Fatalf("marshal providers: %v", err)
	}
	if _, err := client.SystemSetting.Create().SetKey(oauthProvidersSettingKey).SetValue(string(payload)).SetCategory("oauth").Save(t.Context()); err != nil {
		t.Fatalf("create OAuth system setting: %v", err)
	}

	handler, err := New(client, nil, config.Config{
		AppURL: "https://nex.example.test",
		Auth:   config.Auth{SessionSecret: "test-session-secret"},
	})
	if err != nil {
		t.Fatalf("create OAuth handler: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/auth/signin/github?callbackUrl=/console", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Fatalf("authorize status = %d, want %d; body=%s", response.Code, http.StatusFound, response.Body.String())
	}
	redirect, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse authorization redirect: %v", err)
	}
	if got := redirect.Query().Get("client_id"); got != "database-client-id" {
		t.Fatalf("authorization client_id = %q, want database setting", got)
	}
	if got := redirect.Query().Get("scope"); got != "read:user user:email" {
		t.Fatalf("authorization scope = %q, want normalized database setting", got)
	}

	providersRequest := httptest.NewRequest(http.MethodGet, "/api/auth/providers", nil)
	providersResponse := httptest.NewRecorder()
	handler.ServeHTTP(providersResponse, providersRequest)
	if providersResponse.Code != http.StatusOK {
		t.Fatalf("providers status = %d, want %d", providersResponse.Code, http.StatusOK)
	}
	if body := providersResponse.Body.String(); body == "" || !containsAll(body, "数据库 GitHub", "github") || containsAll(body, "database-client-secret") {
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
