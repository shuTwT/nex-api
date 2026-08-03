package system

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/ent"
)

const systemUserSchema = `CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT NOT NULL DEFAULT '', "email" TEXT NOT NULL UNIQUE, "emailVerified" DATETIME, "image" TEXT NOT NULL DEFAULT '', "username" TEXT NOT NULL UNIQUE, "password" TEXT NOT NULL, "role" TEXT NOT NULL, "credits" INTEGER NOT NULL, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);`

func TestService_Initialize_createsExactlyOneAdmin_whenRequestsRace(t *testing.T) {
	client := newSystemTestClient(t)
	service, err := NewService(client)
	if err != nil {
		t.Fatal(err)
	}
	requests := []InitializeRequest{{Email: "first@example.com", Username: "first", Password: "password-one", ConfirmPassword: "password-one"}, {Email: "second@example.com", Username: "second", Password: "password-two", ConfirmPassword: "password-two"}}
	results := make(chan error, len(requests))
	var wg sync.WaitGroup
	for _, request := range requests {
		request := request
		wg.Add(1)
		go func() { defer wg.Done(); _, err := service.Initialize(context.Background(), request); results <- err }()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	users, err := client.User.Query().All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Fatalf("user count = %d, want exactly one", len(users))
	}
	if users[0].Role != "admin" {
		t.Fatalf("created role = %q, want admin", users[0].Role)
	}
	if success != 1 {
		t.Fatalf("successful initializations = %d, want one", success)
	}
}
func newSystemTestClient(t *testing.T) *ent.Client {
	t.Helper()
	path := filepath.Join(t.TempDir(), "system.db")
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err = database.ExecContext(context.Background(), systemUserSchema); err != nil {
		t.Fatal(err)
	}
	client, err := ent.Open(dialect.SQLite, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}
