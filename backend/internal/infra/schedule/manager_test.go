package schedule

import (
	"context"
	"database/sql"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/ent"
)

func TestScheduleManagerUpsertRunAndDisable(t *testing.T) {
	manager, err := NewScheduleManager(slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Shutdown(context.Background()) })
	runs := make(chan struct{}, 1)
	if err := manager.RegisterTask("test_task", "test", func(context.Context) error {
		runs <- struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	definition := Definition{ID: "job-1", Name: "Test", TaskKey: "test_task", ScheduleType: "duration", Expression: "1h", Enabled: true}
	if err := manager.Upsert(definition); err != nil {
		t.Fatal(err)
	}
	manager.Start()
	if runtime := manager.Runtime("job-1"); !runtime.Scheduled || runtime.NextRun == nil {
		t.Fatalf("unexpected runtime: %+v", runtime)
	}
	if err := manager.RunNow("job-1"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-runs:
	case <-time.After(2 * time.Second):
		t.Fatal("job did not run")
	}
	definition.Enabled = false
	if err := manager.Upsert(definition); err != nil {
		t.Fatal(err)
	}
	if runtime := manager.Runtime("job-1"); runtime.Scheduled {
		t.Fatalf("disabled job remains scheduled: %+v", runtime)
	}
}

func TestScheduleManagerRejectsUnknownTaskAndInvalidSchedule(t *testing.T) {
	manager, err := NewScheduleManager(slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Shutdown(context.Background()) })
	if err := manager.RegisterTask("known", "", func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}
	tests := []Definition{
		{ID: "1", Name: "Unknown", TaskKey: "unknown", ScheduleType: "duration", Expression: "1m", Enabled: true},
		{ID: "2", Name: "Bad cron", TaskKey: "known", ScheduleType: "cron", Expression: "bad", Enabled: true},
		{ID: "3", Name: "Bad duration", TaskKey: "known", ScheduleType: "duration", Expression: "0s", Enabled: true},
	}
	for _, definition := range tests {
		if err := manager.Upsert(definition); err == nil {
			t.Fatalf("expected validation error for %+v", definition)
		}
	}
}

func newScheduleTestClient(t *testing.T) *ent.Client {
	t.Helper()
	path := filepath.Join(t.TempDir(), "schedule.db")
	database, err := sql.Open("sqlite3", path+"?_fk=1")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	client, err := ent.Open(dialect.SQLite, path+"?_fk=1")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	return client
}
