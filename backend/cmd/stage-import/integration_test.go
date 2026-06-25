package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Integration tests require BOTH the AnimalEkarte target DB and the old_db stage
// DB to be reachable. They are gated behind STAGE_IMPORT_INTEGRATION=1 so the
// default `go test ./...` (and `make test`/ci-local) stays hermetic. When the env
// is set they run inside the backend container with the same DB_* / STAGE_DB_*
// env the importer uses; see `make stage-import-rollback-test`.

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("STAGE_IMPORT_INTEGRATION") != "1" {
		t.Skip("set STAGE_IMPORT_INTEGRATION=1 (with DB_* + STAGE_DB_* env) to run")
	}
}

func mustPools(t *testing.T) (context.Context, *pgxpool.Pool, *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	td, err := targetDSN()
	if err != nil {
		t.Fatalf("targetDSN: %v", err)
	}
	tp, err := pgxpool.New(ctx, td.connString())
	if err != nil {
		t.Fatalf("target pool: %v", err)
	}
	if err := tp.Ping(ctx); err != nil {
		t.Fatalf("ping target: %v", err)
	}
	sp, err := pgxpool.New(ctx, stageDSN().connStringReadOnly())
	if err != nil {
		t.Fatalf("stage pool: %v", err)
	}
	if err := sp.Ping(ctx); err != nil {
		t.Fatalf("ping stage: %v", err)
	}
	t.Cleanup(func() { tp.Close(); sp.Close() })
	return ctx, tp, sp
}

func snapshotCounts(t *testing.T, ctx context.Context, pool *pgxpool.Pool) map[string]int64 {
	t.Helper()
	counts := map[string]int64{}
	for _, tbl := range importOrder {
		var n int64
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM public."+tbl).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", tbl, err)
		}
		counts[tbl] = n
	}
	return counts
}

// TestApplyImportRollsBackOnInjectedFailure proves the destructive import is
// atomic: when a failure is injected right after the delete phase, the whole
// transaction rolls back and EVERY business-table row count is unchanged — no
// half-scrubbed, half-loaded state. This is the core safety guarantee.
func TestApplyImportRollsBackOnInjectedFailure(t *testing.T) {
	requireIntegration(t)
	ctx, tp, sp := mustPools(t)
	silent := slog.New(slog.NewTextHandler(io.Discard, nil))

	refs, err := resolveTargetRefs(ctx, tp)
	if err != nil {
		t.Fatalf("resolveTargetRefs: %v", err)
	}

	before := snapshotCounts(t, ctx, tp)

	// Fail right after the FIRST delete (deleteOrder[0]) so we prove that delete is
	// rolled back without paying for the full multi-million-row scrub.
	firstTable := deleteOrder[0]
	hooks := &applyHooks{failAfterTable: map[string]error{
		firstTable: errors.New("injected: simulate crash after first delete"),
	}}
	_, _, err = applyImport(ctx, tp, sp, refs, "", silent, hooks)
	if err == nil {
		t.Fatal("applyImport must return an error when the injected failure fires")
	}

	after := snapshotCounts(t, ctx, tp)
	for _, tbl := range importOrder {
		if before[tbl] != after[tbl] {
			t.Errorf("rollback failed: %s count changed %d -> %d (the delete was not rolled back)",
				tbl, before[tbl], after[tbl])
		}
	}
}

// TestStageConnectionIsReadOnly proves the stage pool refuses writes, so the
// importer can never mutate the source DB even via a bug.
func TestStageConnectionIsReadOnly(t *testing.T) {
	requireIntegration(t)
	ctx, _, sp := mustPools(t)
	_, err := sp.Exec(ctx, "CREATE TEMP TABLE _stage_import_write_probe (x int)")
	if err == nil {
		t.Fatal("stage connection accepted a write — it must be read-only")
	}
}
