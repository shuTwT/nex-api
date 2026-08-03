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

	"github.com/shuTwT/nex-api/ent"
	infraschedule "github.com/shuTwT/nex-api/internal/infra/schedule"
)

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

func TestServiceCRUDSynchronizesRunningScheduler(t *testing.T) {
	client := newScheduleTestClient(t)
	manager, err := infraschedule.NewScheduleManager(slog.Default())
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
	service, err := NewService(client, manager)
	if err != nil {
		t.Fatal(err)
	}
	manager.Start()
	ctx := context.Background()
	created, err := service.Create(ctx, UpsertInput{Name: "Test", TaskKey: "test_task", ScheduleType: "duration", Expression: "1h", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if !created.Runtime.Scheduled {
		t.Fatal("created job was not scheduled")
	}
	if err := service.RunNow(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-runs:
	case <-time.After(2 * time.Second):
		t.Fatal("created job did not run")
	}
	updated, err := service.Update(ctx, created.ID, UpsertInput{Name: "Test", TaskKey: "test_task", ScheduleType: "cron", Expression: "*/5 * * * *", Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Runtime.Scheduled {
		t.Fatal("disabled job remains scheduled")
	}
	if err := service.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(ctx, created.ID); err == nil {
		t.Fatal("deleted job remains in database")
	}
}
