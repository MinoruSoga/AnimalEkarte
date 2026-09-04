package csvimport

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type cutoverBandState int

const (
	cutoverBandEmpty cutoverBandState = iota
	cutoverBandComplete
)

type cutoverSession interface {
	cutoverQuerier
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Begin(context.Context, pgx.TxIsoLevel) (cutoverTransaction, error)
}

type pgxPoolCutoverSession struct {
	conn *pgxpool.Conn
}

func (s pgxPoolCutoverSession) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return s.conn.QueryRow(ctx, sql, args...)
}

func (s pgxPoolCutoverSession) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return s.conn.Exec(ctx, sql, args...)
}

func (s pgxPoolCutoverSession) Begin(ctx context.Context, iso pgx.TxIsoLevel) (cutoverTransaction, error) {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso})
	if err != nil {
		return nil, err
	}
	return pgxCutoverTransaction{Tx: tx}, nil
}

// PreflightCutoverTargetAllowingResume is the STG UAT preflight. A complete
// clinic-band prefix that already matches the manifest is allowed so a later
// table can resume after a previous table had committed. Formal cutover still
// uses PreflightCutoverTarget and requires every band empty.
func PreflightCutoverTargetAllowingResume(ctx context.Context, target cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs) error {
	if err := validateCutoverTarget(ctx, target, manifest, seeds, false); err != nil {
		return err
	}
	return validateCutoverBandResumable(ctx, target, manifest, seeds)
}

func validateCutoverBandResumable(ctx context.Context, q cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs) error {
	seenEmpty := false
	for i, spec := range CutoverTableSpecs() {
		state, _, err := inspectCutoverBand(ctx, q, spec, manifest.Tables[i], manifest.IDBand, seeds)
		if err != nil {
			return err
		}
		switch state {
		case cutoverBandEmpty:
			seenEmpty = true
		case cutoverBandComplete:
			if seenEmpty {
				return fmt.Errorf("%s: target clinic band is already occupied in table %s", CutoverRefBandOccupied, spec.Name)
			}
		}
	}
	return nil
}

func inspectCutoverBand(
	ctx context.Context,
	q cutoverQuerier,
	spec CutoverTableSpec,
	table CutoverManifestTable,
	band CutoverIDBand,
	seeds CutoverSeedIDs,
) (cutoverBandState, int64, error) {
	count, err := cutoverBandCount(ctx, q, spec, band)
	if err != nil {
		return 0, 0, err
	}
	if count == 0 {
		return cutoverBandEmpty, 0, nil
	}
	if err := verifyCutoverTable(ctx, q, spec, table, band, seeds); err != nil {
		return 0, count, err
	}
	return cutoverBandComplete, count, nil
}

// ApplyCutoverCommittingEachTable imports each CSV in its own transaction.
// PlanetScale terminated a multi-million-row single transaction (SQLSTATE-less
// connection close after ~30 minutes). Formal ApplyCutover stays one
// transaction. Successful tables stay committed; retry skips a complete prefix.
func ApplyCutoverCommittingEachTable(
	ctx context.Context,
	pool *pgxpool.Pool,
	bundle CutoverBundle,
	seeds CutoverSeedIDs,
	iso pgx.TxIsoLevel,
) (CutoverResult, error) {
	if iso == "" {
		iso = pgx.RepeatableRead
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return CutoverResult{}, ErrCutoverTransactionNotStarted
	}
	defer conn.Release()
	return applyCutoverCommittingEachTable(ctx, pgxPoolCutoverSession{conn: conn}, bundle, seeds, iso)
}

func applyCutoverCommittingEachTable(
	ctx context.Context,
	session cutoverSession,
	bundle CutoverBundle,
	seeds CutoverSeedIDs,
	iso pgx.TxIsoLevel,
) (CutoverResult, error) {
	if _, err := session.Exec(ctx, `SELECT pg_advisory_lock($1)`, cutoverAdvisoryLockKey); err != nil {
		return CutoverResult{}, fmt.Errorf("acquire cutover lock: %w", err)
	}
	defer func() {
		_, _ = session.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, cutoverAdvisoryLockKey)
	}()
	if _, err := session.Exec(ctx, `SELECT set_config('app.bypass_rls', 'on', false)`); err != nil {
		return CutoverResult{}, fmt.Errorf("configure cutover RLS bypass: %w", err)
	}
	if err := validateCutoverTarget(ctx, session, bundle.Manifest, seeds, false); err != nil {
		return CutoverResult{}, err
	}

	counts := make(map[string]int64, len(bundle.Manifest.Tables))
	seenEmpty := false
	for i, spec := range CutoverTableSpecs() {
		manifestTable := bundle.Manifest.Tables[i]
		state, count, err := inspectCutoverBand(ctx, session, spec, manifestTable, bundle.Manifest.IDBand, seeds)
		if err != nil {
			return CutoverResult{}, err
		}
		if state == cutoverBandComplete {
			if seenEmpty {
				return CutoverResult{}, fmt.Errorf("%s: target clinic band is already occupied in table %s", CutoverRefBandOccupied, spec.Name)
			}
			counts[spec.Name] = count
			continue
		}
		seenEmpty = true
		imported, err := importOneCutoverTable(ctx, session, iso, bundle, spec, manifestTable, seeds)
		if err != nil {
			return CutoverResult{}, err
		}
		counts[spec.Name] = imported
	}

	if err := finalizeCutoverAfterTables(ctx, session, iso, bundle, seeds); err != nil {
		return CutoverResult{}, err
	}
	return CutoverResult{
		CompletedAt: time.Now().UTC(),
		ClinicCode:  bundle.Manifest.ClinicCode,
		RunID:       bundle.Manifest.SourceRunID,
		IDBand:      bundle.Manifest.IDBand,
		Counts:      counts,
	}, nil
}

func importOneCutoverTable(
	ctx context.Context,
	session cutoverSession,
	iso pgx.TxIsoLevel,
	bundle CutoverBundle,
	spec CutoverTableSpec,
	manifestTable CutoverManifestTable,
	seeds CutoverSeedIDs,
) (int64, error) {
	tx, err := session.Begin(ctx, iso)
	if err != nil {
		return 0, fmt.Errorf("begin cutover table %s", spec.Name)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout = '10s'`); err != nil {
		return 0, fmt.Errorf("configure cutover lock timeout: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('app.bypass_rls', 'on', true)`); err != nil {
		return 0, fmt.Errorf("configure cutover RLS bypass: %w", err)
	}
	if err := lockCutoverTable(ctx, tx, spec.Name); err != nil {
		return 0, err
	}
	occupied, err := cutoverBandOccupied(ctx, tx, spec, bundle.Manifest.IDBand)
	if err != nil {
		return 0, err
	}
	if occupied {
		return 0, fmt.Errorf("%s: target clinic band is already occupied in table %s", CutoverRefBandOccupied, spec.Name)
	}
	path := filepath.Join(bundle.SourceDir, manifestTable.File)
	count, err := copyCutoverTable(ctx, tx, path, spec, manifestTable, seeds)
	if err != nil {
		return 0, err
	}
	if err := verifyCutoverTable(ctx, tx, spec, manifestTable, bundle.Manifest.IDBand, seeds); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		if errors.Is(err, pgx.ErrTxCommitRollback) {
			return 0, fmt.Errorf("commit rejected and transaction rolled back: %w", err)
		}
		return 0, fmt.Errorf("%w: run read-only verify before any retry or restore", ErrCutoverCommitOutcomeUnknown)
	}
	return count, nil
}

func finalizeCutoverAfterTables(
	ctx context.Context,
	session cutoverSession,
	iso pgx.TxIsoLevel,
	bundle CutoverBundle,
	seeds CutoverSeedIDs,
) error {
	tx, err := session.Begin(ctx, iso)
	if err != nil {
		return fmt.Errorf("begin cutover final verification")
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout = '10s'`); err != nil {
		return fmt.Errorf("configure cutover lock timeout: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('app.bypass_rls', 'on', true)`); err != nil {
		return fmt.Errorf("configure cutover RLS bypass: %w", err)
	}
	if err := verifyCutoverRows(ctx, tx, bundle.Manifest, seeds, bundle.Provenance); err != nil {
		return err
	}
	if err := advanceCutoverSequences(ctx, tx); err != nil {
		return err
	}
	if err := verifyCutoverSequences(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		if errors.Is(err, pgx.ErrTxCommitRollback) {
			return fmt.Errorf("commit rejected and transaction rolled back: %w", err)
		}
		return fmt.Errorf("%w: run read-only verify before any retry or restore", ErrCutoverCommitOutcomeUnknown)
	}
	return nil
}
