package auth

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/infra/config"
	"github.com/shuTwT/nex-api/backend/ent"
)

const legacyPasswordHash = "00112233445566778899aabbccddeeff:5699cfee2c5c280e66678242092f368ce88ff05305af2c75a9e629d473deb2165b3797e0e31ec3cda30414573befb697f928384c38b187e8c176107e5be20f01"
const authSchema = `CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT NOT NULL DEFAULT '', "email" TEXT NOT NULL, "emailVerified" DATETIME, "image" TEXT NOT NULL DEFAULT '', "username" TEXT NOT NULL, "password" TEXT NOT NULL, "role" TEXT NOT NULL, "credits" INTEGER NOT NULL, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL); CREATE TABLE "Session" ("id" TEXT PRIMARY KEY NOT NULL, "sessionToken" TEXT NOT NULL UNIQUE, "userId" TEXT NOT NULL, "expires" DATETIME NOT NULL, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);`

func TestVerifyPassword_acceptsLegacyScryptEncoding(t *testing.T) {
	matched, err := VerifyPassword("correct horse battery staple", legacyPasswordHash)
	if err != nil {
		t.Fatalf("VerifyPassword returned an error: %v", err)
	}
	if !matched {
		t.Fatal("VerifyPassword rejected the legacy hash")
	}
}

func TestVerifyPassword_rejectsWrongAndMalformedValues(t *testing.T) {
	for _, test := range []struct{ name, password, hashed string }{
		{"wrong password", "wrong", legacyPasswordHash}, {"missing separator", "correct horse battery staple", "not-a-hash"}, {"wrong key length", "correct horse battery staple", "00112233445566778899aabbccddeeff:00"},
	} {
		t.Run(test.name, func(t *testing.T) {
			matched, err := VerifyPassword(test.password, test.hashed)
			if err != nil {
				t.Fatalf("VerifyPassword returned an error: %v", err)
			}
			if matched {
				t.Fatal("VerifyPassword accepted an invalid value")
			}
		})
	}
}

func TestService_rotationInvalidatesPreviousSession(t *testing.T) {
	service, _ := newTestService(t, "rotation", "user")
	first, err := service.Login(context.Background(), "user@example.com", "correct horse battery staple", "")
	if err != nil {
		t.Fatalf("initial login: %v", err)
	}
	rotated, err := service.RotateSession(context.Background(), first)
	if err != nil {
		t.Fatalf("rotate session: %v", err)
	}
	if first.token == rotated.token {
		t.Fatal("session rotation reused the old token")
	}
	if _, err := service.Authenticate(context.Background(), first.token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("old session error = %v", err)
	}
	if _, err := service.Authenticate(context.Background(), rotated.token); err != nil {
		t.Fatalf("rotated session was not accepted: %v", err)
	}
}

func newTestService(t *testing.T, name, role string, options ...Option) (*Service, *ent.Client) {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), name+".db")
	database, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err = database.ExecContext(context.Background(), authSchema); err != nil {
		t.Fatalf("create auth tables: %v", err)
	}
	client, err := ent.Open(dialect.SQLite, databasePath)
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if _, err = client.User.Create().SetID("user-1").SetEmail("user@example.com").SetUsername("user").SetPassword(legacyPasswordHash).SetRole(role).SetCredits(100).SetCreatedAt(time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)).SetUpdatedAt(time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)).Save(context.Background()); err != nil {
		t.Fatalf("create test user: %v", err)
	}
	options = append([]Option{WithClock(clockFunc(func() time.Time { return time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC) }))}, options...)
	service, err := NewService(client, config.Auth{SessionSecret: "test-session-secret"}, options...)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	return service, client
}
