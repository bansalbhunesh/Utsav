//go:build integration
// +build integration

package httpserver

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/bhune/utsav/services/api/internal/db"
	"github.com/bhune/utsav/services/api/internal/migrate"
)

func TestMigrationsAgainstPostgresContainer(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("utsav"),
		postgres.WithUsername("utsav"),
		postgres.WithPassword("utsav"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	if err != nil {
		t.Skipf("docker/testcontainers unavailable: %v", err)
		return
	}
	defer func() { _ = pg.Terminate(context.Background()) }()

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("postgres connection string: %v", err)
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	migPath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "db", "migrations"))
	if err := migrate.Up(dsn, migPath); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	pool, err := db.Connect(ctx, dsn, true)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	var tableExists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema='public' AND table_name='events'
		)`).Scan(&tableExists); err != nil {
		t.Fatalf("events table existence query failed: %v", err)
	}
	if !tableExists {
		t.Fatal("expected events table after migrations")
	}

	var columnExists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema='public' AND table_name='gallery_assets' AND column_name='status'
		)`).Scan(&columnExists); err != nil {
		t.Fatalf("gallery_assets.status column query failed: %v", err)
	}
	if !columnExists {
		t.Fatal("expected gallery_assets.status column after migrations")
	}
}
