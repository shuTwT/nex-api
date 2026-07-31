package system

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

const systemUserSchema = `CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT NOT NULL DEFAULT '', "email" TEXT NOT NULL UNIQUE, "emailVerified" DATETIME, "image" TEXT NOT NULL DEFAULT '', "username" TEXT NOT NULL UNIQUE, "password" TEXT NOT NULL, "role" TEXT NOT NULL, "credits" INTEGER NOT NULL, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);`

func TestService_Initialize_createsExactlyOneAdmin_whenRequestsRace(t *testing.T) {
	// Given
	client := newSystemTestClient(t)
	service, err := NewService(client)
	if err != nil {
		t.Fatalf("NewService returned an error: %v", err)
	}
	requests := []InitializeRequest{
		{Email: "first@example.com", Username: "first", Password: "password-one", ConfirmPassword: "password-one"},
		{Email: "second@example.com", Username: "second", Password: "password-two", ConfirmPassword: "password-two"},
	}
	results := make(chan error, len(requests))
	var waitGroup sync.WaitGroup
	for _, request := range requests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, initializeErr := service.Initialize(context.Background(), request)
			results <- initializeErr
		}()
		break
	}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		_, initializeErr := service.Initialize(context.Background(), requests[1])
		results <- initializeErr
	}()
	waitGroup.Wait()
	close(results)
	var successCount int
	for initializeErr := range results {
		if initializeErr == nil {
			successCount++
			continue
		}
		t.Logf("initialize attempt failed: %v", initializeErr)
	}

	// When
	users, err := client.User.Query().All(context.Background())

	// Then
	if err != nil {
		t.Fatalf("query initialized users: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("user count = %d, want exactly one", len(users))
	}
	if users[0].Role != "admin" {
		t.Fatalf("created role = %q, want admin", users[0].Role)
	}
	if successCount != 1 {
		t.Fatalf("successful initializations = %d, want one", successCount)
	}
}

func newSystemTestClient(t *testing.T) *ent.Client {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "system.db")
	database, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.ExecContext(context.Background(), systemUserSchema); err != nil {
		t.Fatalf("create system tables: %v", err)
	}
	client, err := ent.Open(dialect.SQLite, databasePath)
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}
