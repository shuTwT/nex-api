package authz

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/ent"
)

const authzSchema = `CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT, "email" TEXT NOT NULL, "emailVerified" DATETIME, "image" TEXT, "username" TEXT NOT NULL, "password" TEXT NOT NULL, "role" TEXT NOT NULL DEFAULT 'user', "credits" INTEGER NOT NULL DEFAULT 1000, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL); CREATE TABLE "ApiToken" ("id" TEXT PRIMARY KEY NOT NULL, "userId" TEXT NOT NULL, "name" TEXT NOT NULL, "token" TEXT NOT NULL UNIQUE, "permissions" TEXT NOT NULL, "lastUsedAt" DATETIME, "expiresAt" DATETIME, "isActive" BOOLEAN NOT NULL DEFAULT 1, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);`

func TestGenerateToken_hasExactFormat(t *testing.T) {
	token, err := generateToken(bytes.NewReader(bytes.Repeat([]byte{0xab}, TokenEntropySize)))
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if token != "sk_"+strings.Repeat("ab", TokenEntropySize) || !IsGeneratedToken(token) {
		t.Fatalf("token format = %q", token)
	}
}
func TestParseBearerToken_rejectsNonBearerHeaders(t *testing.T) {
	for _, test := range []struct {
		name, header string
		valid        bool
	}{{"case insensitive bearer", "bEaReR sk_value", true}, {"missing header", "", false}, {"wrong scheme", "Basic sk_value", false}, {"extra values", "Bearer sk_value extra", false}} {
		t.Run(test.name, func(t *testing.T) {
			token, err := ParseBearerToken(test.header)
			if test.valid && (err != nil || token != "sk_value") {
				t.Fatalf("parsed bearer = %q, error = %v", token, err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidBearer) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
func TestPermissions_mapMethodsToScopes(t *testing.T) {
	for _, test := range []struct {
		permissions, method string
		allowed             bool
	}{{"read", http.MethodGet, true}, {"read", http.MethodPost, false}, {"read,write", http.MethodPatch, true}, {"read,write", http.MethodDelete, false}, {"read,write,delete", http.MethodDelete, true}} {
		permissions, err := ParsePermissions(test.permissions)
		if err != nil {
			t.Fatal(err)
		}
		if permissions.AllowsMethod(test.method) != test.allowed {
			t.Errorf("%q %s permission mismatch", test.permissions, test.method)
		}
	}
	for _, raw := range []string{"write", "read,delete"} {
		if _, err := ParsePermissions(raw); !errors.Is(err, ErrInvalidPermissions) {
			t.Errorf("permissions %q error = %v", raw, err)
		}
	}
}
func TestTokenService_authenticatesAndAtomicallyTouchesActiveToken(t *testing.T) {
	service, client, now := newTokenService(t)
	seedToken(t, client, "token-1", "user-1", "sk_active", "read", true, time.Time{})
	principal, err := service.AuthenticateBearer(context.Background(), "Bearer sk_active")
	if err != nil {
		t.Fatalf("authenticate token: %v", err)
	}
	if principal.UserID != "user-1" || principal.Source != APITokenCredential {
		t.Fatalf("principal = %+v", principal)
	}
	stored, err := client.ApiToken.Get(context.Background(), "token-1")
	if err != nil {
		t.Fatal(err)
	}
	if !stored.LastUsedAt.Equal(now) {
		t.Fatalf("lastUsedAt = %v, want %v", stored.LastUsedAt, now)
	}
}
func TestTokenService_rejectsInactiveExpiredAndInvalidPermissionTokens(t *testing.T) {
	service, client, now := newTokenService(t)
	seedToken(t, client, "inactive", "user-1", "sk_inactive", "read", false, time.Time{})
	seedToken(t, client, "expired", "user-1", "sk_expired", "read", true, now)
	seedToken(t, client, "malformed", "user-1", "sk_malformed", "root", true, time.Time{})
	for _, token := range []string{"sk_inactive", "sk_expired", "sk_malformed", "sk_missing"} {
		if _, err := service.Authenticate(context.Background(), token); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("token %q error = %v", token, err)
		}
	}
}

func newTokenService(t *testing.T) (*TokenService, *ent.Client, time.Time) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "authz.db")
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = database.ExecContext(context.Background(), authzSchema); err != nil {
		t.Fatal(err)
	}
	_ = database.Close()
	client, err := ent.Open(dialect.SQLite, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	for _, account := range []struct{ id, role string }{{"user-1", "user"}, {"user-2", "admin"}} {
		if _, err = client.User.Create().SetID(account.id).SetEmail(account.id + "@example.com").SetUsername(account.id).SetPassword("redacted").SetRole(account.role).SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	store, err := NewEntTokenStore(client)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewTokenService(store, WithClock(clockFunc(func() time.Time { return now })))
	if err != nil {
		t.Fatal(err)
	}
	return service, client, now
}
func seedToken(t *testing.T, client *ent.Client, id, userID, token, permissions string, active bool, expiresAt time.Time) {
	t.Helper()
	builder := client.ApiToken.Create().SetID(id).SetUserID(userID).SetName(id).SetToken(token).SetPermissions(permissions).SetIsActive(active).SetCreatedAt(time.Now()).SetUpdatedAt(time.Now())
	if !expiresAt.IsZero() {
		builder.SetExpiresAt(expiresAt)
	}
	if _, err := builder.Save(context.Background()); err != nil {
		t.Fatal(err)
	}
}
