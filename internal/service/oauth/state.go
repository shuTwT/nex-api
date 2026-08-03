package oauth

import (
	infraoauth "github.com/shuTwT/nex-api/internal/infra/oauth"
	"time"
)

type OAuthState struct {
	Provider, Value, ReturnURL, CodeVerifier, CodeChallenge, Nonce string
	IssuedAt                                                       time.Time
}
type StateCookie struct {
	Name, Value, Path string
	Expires           time.Time
	MaxAge            int
	HttpOnly, Secure  bool
}
type StateManager struct{ codec *infraoauth.StateCodec }

func NewStateManager(secret []byte) (*StateManager, error) {
	codec, err := infraoauth.NewStateCodec(secret)
	if err != nil {
		return nil, err
	}
	return &StateManager{codec: codec}, nil
}
func (m *StateManager) New(provider, returnURL string, pkce bool) (OAuthState, StateCookie, error) {
	state, err := m.codec.New(provider, returnURL, pkce)
	if err != nil {
		return OAuthState{}, StateCookie{}, err
	}
	cookie, err := m.codec.Encode(state)
	if err != nil {
		return OAuthState{}, StateCookie{}, err
	}
	return fromInfraState(state), StateCookie{Name: cookie.Name, Value: cookie.Value, Path: cookie.Path, Expires: cookie.Expires, MaxAge: cookie.MaxAge, HttpOnly: cookie.HttpOnly, Secure: cookie.Secure}, nil
}
func (m *StateManager) Read(cookie, provider, value string) (OAuthState, error) {
	state, err := m.codec.Decode(cookie, provider, value)
	return fromInfraState(state), err
}
func (m *StateManager) CookieName() string { return m.codec.CookieName() }
func ResolveReturnURL(raw, base string) (string, error) {
	return infraoauth.ResolveReturnURL(raw, base)
}
func fromInfraState(value infraoauth.OAuthState) OAuthState {
	return OAuthState{Provider: value.Provider, Value: value.Value, ReturnURL: value.ReturnURL, CodeVerifier: value.CodeVerifier, CodeChallenge: value.CodeChallenge, Nonce: value.Nonce, IssuedAt: value.IssuedAt}
}
func toInfraState(value OAuthState) infraoauth.OAuthState {
	return infraoauth.OAuthState{Provider: value.Provider, Value: value.Value, ReturnURL: value.ReturnURL, CodeVerifier: value.CodeVerifier, CodeChallenge: value.CodeChallenge, Nonce: value.Nonce, IssuedAt: value.IssuedAt}
}
