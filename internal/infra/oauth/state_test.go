package oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestResolveReturnURL_rejectsExternalTargets(t *testing.T) {
	got, err := ResolveReturnURL("https://attacker.example/steal", "https://nex.example.com")
	if err == nil {
		t.Fatal("ResolveReturnURL accepted an external target")
	}
	if got != "" {
		t.Fatalf("ResolveReturnURL returned %q for an invalid target", got)
	}
}
func TestStateCodec_roundTripsPKCEAndNonce(t *testing.T) {
	codec, err := NewStateCodec([]byte("test-state-secret"))
	if err != nil {
		t.Fatal(err)
	}
	state, err := codec.New("easy1auth", "https://nex.example.com/console", true)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	codec.Write(response, state)
	request := httptest.NewRequest("GET", "/api/auth/callback/easy1auth?state="+state.Value, nil)
	for _, cookie := range response.Result().Cookies() {
		request.AddCookie(cookie)
	}
	got, err := codec.Read(request, "easy1auth", state.Value)
	if err != nil {
		t.Fatal(err)
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
func TestValidateOIDCNonce_rejectsMismatchedNonce(t *testing.T) {
	idToken := makeTestIDToken(t, "expected-nonce")
	if err := ValidateOIDCNonce(idToken, "different-nonce"); err == nil {
		t.Fatal("ValidateOIDCNonce accepted a mismatched nonce")
	}
}
func makeTestIDToken(t *testing.T, nonce string) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "none"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]string{"nonce": nonce})
	if err != nil {
		t.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}
