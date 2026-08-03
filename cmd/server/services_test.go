package main

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/shuTwT/nex-api/internal/infra/config"
)

func TestCleanupStackRunsAcquiredResourcesInReverseOrder(t *testing.T) {
	var cleanup cleanupStack
	var calls []string
	cleanup.add(func() { calls = append(calls, "database") })
	cleanup.add(func() { calls = append(calls, "redis") })
	cleanup.add(func() { calls = append(calls, "scheduler") })

	cleanup.run()

	want := []string{"scheduler", "redis", "database"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("cleanup order = %v, want %v", calls, want)
	}
}

func TestBuildServicesClosesCreatedClientsWhenLaterConstructionFails(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{
		Database: config.Database{
			URL:             "file:" + filepath.Join(t.TempDir(), "nex.db"),
			ConnMaxLifetime: time.Minute,
		},
		Redis: config.Redis{URL: "redis://localhost:6379/0"},
		Upload: config.Upload{
			Directory:    filepath.Join(t.TempDir(), "uploads"),
			AllowedTypes: []string{"*/*"},
		},
		Cron:   config.Cron{Interval: time.Minute},
		Server: config.Server{ShutdownTimeout: time.Second},
	}

	deps, _, err := buildServices(ctx, cfg, slog.Default())
	if err == nil {
		t.Fatal("build services unexpectedly succeeded")
	}
	if deps.Client == nil || deps.Redis == nil {
		t.Fatalf("build services did not expose acquired clients: %#v", deps)
	}
	if err := deps.Redis.Ping(ctx).Err(); !errors.Is(err, redis.ErrClosed) {
		t.Fatalf("redis ping error after rollback = %v, want redis.ErrClosed", err)
	}
	if _, err := deps.Client.User.Query().Count(ctx); err == nil {
		t.Fatal("database query unexpectedly succeeded after rollback")
	}
}
