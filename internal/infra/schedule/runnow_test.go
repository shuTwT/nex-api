package schedule

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

func TestScheduleManagerRunNowFailsFastWhenNotStarted(t *testing.T) {
	manager, err := NewScheduleManager(slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Shutdown(context.Background()) })
	ran := make(chan struct{}, 1)
	if err := manager.RegisterTask("test_task", "test", func(context.Context) error {
		ran <- struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	definition := Definition{ID: "job-1", Name: "Test", TaskKey: "test_task", ScheduleType: "duration", Expression: "1h", Enabled: true}
	if err := manager.Upsert(definition); err != nil {
		t.Fatal(err)
	}

	begin := time.Now()
	err = manager.RunNow("job-1")
	if !errors.Is(err, ErrNotStarted) {
		t.Fatalf("RunNow error = %v, want ErrNotStarted", err)
	}
	if elapsed := time.Since(begin); elapsed > 500*time.Millisecond {
		t.Fatalf("RunNow took %v, want an immediate failure", elapsed)
	}
	select {
	case <-ran:
		t.Fatal("task ran even though the scheduler was not started")
	default:
	}
}
