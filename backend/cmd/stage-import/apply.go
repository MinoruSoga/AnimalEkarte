package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// tablePlan is the pre-import accounting for one table: how many stage rows are
// importable, how many are skipped (by status), how many importable-status rows
// are still skipped because their hard-FK parent is not importable, and how many
// old_db rows the apply would delete from the target.
type tablePlan struct {
	Table            string
	Importable       int64
	SkippedByStatus  map[string]int64 // mapping_status -> count (needs_review/archive_only/blocked)
	SkippedParentGap int64            // confirmed/inferred rows whose hard-FK parent is excluded
	DeleteFromTarget int64
}

// importPlan is the whole dry-run picture.
type importPlan struct {
	Tables []tablePlan
}

// buildPlan computes per-table importable / skipped / delete counts. Read-only:
// it issues only SELECTs against stage and target. The importable count uses the
// EXACT same predicate (importableWhere) the apply uses, so dry-run and apply
// agree to the row.
func buildPlan(ctx context.Context, stagePool, targetPool *pgxpool.Pool, runID string) (importPlan, error) {
	var p importPlan
	for _, t := range importOrder {
		tp := tablePlan{Table: t, SkippedByStatus: map[string]int64{}}

		// Importable: identical predicate to the apply SELECT (status + run-id +
		// pets owner-present + hard-FK-parent-importable cascade).
		if err := stagePool.QueryRow(ctx, fmt.Sprintf(
			`SELECT count(*) FROM animalekarte_stage.%s WHERE %s`,
			t, importableWhere(t, runID),
		)).Scan(&tp.Importable); err != nil {
			return p, fmt.Errorf("count importable %s: %w", t, err)
		}

		// Skipped by status (status not importable), grouped so each excluded
		// bucket is countable and reported (nothing silently dropped).
		skipRows, err := stagePool.Query(ctx, fmt.Sprintf(
			`SELECT mapping_status, count(*) FROM animalekarte_stage.%s
			 WHERE NOT (%s)%s GROUP BY mapping_status`,
			t, statusInClause(), runIDClause(runID)))
		if err != nil {
			return p, fmt.Errorf("count skipped %s: %w", t, err)
		}
		for skipRows.Next() {
			var status string
			var n int64
			if err := skipRows.Scan(&status, &n); err != nil {
				skipRows.Close()
				return p, fmt.Errorf("scan skipped %s: %w", t, err)
			}
			tp.SkippedByStatus[status] = n
		}
		skipRows.Close()
		if err := skipRows.Err(); err != nil {
			return p, err
		}

		// Parent-gap: rows whose OWN status is importable but are still excluded
		// because their hard-FK parent is not importable (= status-importable minus
		// fully-importable). Only the hard-FK children have this gap.
		var statusImportable int64
		if err := stagePool.QueryRow(ctx, fmt.Sprintf(
			`SELECT count(*) FROM animalekarte_stage.%s WHERE %s%s`,
			t, statusInClause(), runIDClause(runID),
		)).Scan(&statusImportable); err != nil {
			return p, fmt.Errorf("count status-importable %s: %w", t, err)
		}
		tp.SkippedParentGap = statusImportable - tp.Importable

		// old_db rows the destructive apply would delete from the target.
		if err := targetPool.QueryRow(ctx, fmt.Sprintf(
			`SELECT count(*) FROM public.%s WHERE %s`, t, deleteScope(t),
		)).Scan(&tp.DeleteFromTarget); err != nil {
			return p, fmt.Errorf("count target delete %s: %w", t, err)
		}

		p.Tables = append(p.Tables, tp)
	}
	return p, nil
}

// printPlan logs the dry-run plan: per table importable, skipped-by-status, and
// the old_db delete count the apply would perform.
func printPlan(logger *slog.Logger, p importPlan, opt options) {
	mode := "DRY-RUN (no writes)"
	if opt.apply {
		mode = "APPLY (destructive: delete old_db rows + reinsert)"
	}
	logger.Info("═══════════ stage-import plan ═══════════", slog.String("mode", mode))
	var totalImport, totalSkip, totalParentGap, totalDelete int64
	for _, tp := range p.Tables {
		var skipTotal int64
		for _, n := range tp.SkippedByStatus {
			skipTotal += n
		}
		totalImport += tp.Importable
		totalSkip += skipTotal
		totalParentGap += tp.SkippedParentGap
		totalDelete += tp.DeleteFromTarget
		logger.Info("plan",
			slog.String("table", tp.Table),
			slog.Int64("importable", tp.Importable),
			slog.Int64("skipped_by_status", skipTotal),
			slog.Any("skipped_status_buckets", tp.SkippedByStatus),
			slog.Int64("skipped_parent_excluded", tp.SkippedParentGap),
			slog.Int64("target_old_db_to_delete", tp.DeleteFromTarget))
	}
	logger.Info("plan totals",
		slog.Int64("importable", totalImport),
		slog.Int64("skipped_by_status", totalSkip),
		slog.Int64("skipped_parent_excluded", totalParentGap),
		slog.Int64("target_old_db_to_delete", totalDelete))
}

// applyHooks are test-only injection points. nil in production.
//   - failAfterTable, when set and returning a non-nil error for a given table,
//     aborts the transaction immediately after that table's DELETE — the moment a
//     non-atomic implementation would leave a half-scrubbed DB. The test points it
//     at the FIRST (small, indexed) delete so the rollback guarantee is proven
//     without paying for the full multi-million-row scrub.
type applyHooks struct {
	failAfterTable map[string]error
}

// applyImport performs the destructive import inside a SINGLE transaction: scrub
// old_db-derived rows (FK-safe child->parent), then COPY in the importable stage
// rows (FK-safe parent->child). Any error rolls the whole thing back, leaving the
// target unchanged (no half-imported state). After commit, BIGSERIAL sequences
// for explicit-id tables are advanced so app inserts won't collide.
//
// Returns inserted[table] and deleted[table] counts.
func applyImport(
	ctx context.Context,
	targetPool, stagePool *pgxpool.Pool,
	refs targetRefs,
	runID string,
	logger *slog.Logger,
	hooks *applyHooks,
) (map[string]int64, map[string]int64, error) {
	inserted := map[string]int64{}
	deleted := map[string]int64{}

	tx, err := targetPool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	// Rollback is a no-op after a successful Commit; it is the safety net on any
	// early return below.
	defer tx.Rollback(ctx) //nolint:errcheck

	// 0. Ensure the FK index the scoped delete relies on exists. exam_results has
	//    ~1.3M rows and its exam_id FK is unindexed in the base schema, so deleting
	//    old_db exams would seq-scan it repeatedly (minutes). IF NOT EXISTS makes
	//    this idempotent and harmless when the index is already present.
	if _, err := tx.Exec(ctx,
		`CREATE INDEX IF NOT EXISTS idx_exam_results_exam_id ON public.exam_results(exam_id)`,
	); err != nil {
		return nil, nil, fmt.Errorf("ensure exam_results.exam_id index (rolled back): %w", err)
	}

	// 1a. Pre-scrub high-volume CASCADE children of medical_records that this
	//     importer does not re-insert. Without this, the medical_records DELETE
	//     cascades into ~2.8M dependent rows in one statement and dominates runtime.
	for _, t := range cascadeChildScrubOrder {
		tag, err := tx.Exec(ctx, fmt.Sprintf(
			`DELETE FROM public.%s WHERE %s`, t, cascadeChildScope()))
		if err != nil {
			return nil, nil, fmt.Errorf("scrub cascade child %s (rolled back): %w", t, err)
		}
		deleted[t] = tag.RowsAffected()
		logger.Info("scrubbed cascade child", slog.String("table", t), slog.Int64("deleted", deleted[t]))
	}

	// 1. Scrub old_db-derived rows, children first.
	for _, t := range deleteOrder {
		tag, err := tx.Exec(ctx, fmt.Sprintf(
			`DELETE FROM public.%s WHERE %s`, t, deleteScope(t)))
		if err != nil {
			return nil, nil, fmt.Errorf("delete old_db %s (rolled back): %w", t, err)
		}
		deleted[t] = tag.RowsAffected()
		logger.Info("scrubbed old_db rows", slog.String("table", t), slog.Int64("deleted", deleted[t]))

		// Injected failure point (tests only): abort right after this table's delete
		// to prove the delete is rolled back rather than left committed.
		if hooks != nil {
			if injErr, ok := hooks.failAfterTable[t]; ok && injErr != nil {
				return nil, nil, fmt.Errorf("injected failure after delete %s (rolled back): %w", t, injErr)
			}
		}
	}

	// 2. Re-insert importable stage rows, parents first.
	for _, t := range importOrder {
		n, err := copyRowsFromStage(ctx, tx, stagePool, t, runID, refs)
		if err != nil {
			return nil, nil, fmt.Errorf("insert %s (rolled back): %w", t, err)
		}
		inserted[t] = n
		logger.Info("inserted stage rows", slog.String("table", t), slog.Int64("inserted", n))
	}

	// 3. Advance BIGSERIAL sequences for tables that received explicit ids, so the
	//    next app-level INSERT does not collide with an imported id.
	if err := advanceSequences(ctx, tx, importOrder); err != nil {
		return nil, nil, fmt.Errorf("advance sequences (rolled back): %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit (rolled back): %w", err)
	}
	return inserted, deleted, nil
}

// advanceSequences sets each table's id sequence to max(id)+1.
func advanceSequences(ctx context.Context, tx pgx.Tx, tables []string) error {
	for _, t := range tables {
		// setval(..., max(id)+1, false): next nextval() returns max(id)+1.
		_, err := tx.Exec(ctx, fmt.Sprintf(
			`SELECT setval(pg_get_serial_sequence('public.%s','id'), COALESCE((SELECT max(id) FROM public.%s),0)+1, false)`,
			t, t))
		if err != nil {
			return fmt.Errorf("setval %s: %w", t, err)
		}
	}
	return nil
}
