package csvimport

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCutoverSanitizedWriteErrorOmitsConnectionResetText(t *testing.T) {
	err := cutoverSanitizedWriteError("payment_splits", "INSERT", errors.New("read tcp 10.0.0.1:5432: connection reset by peer"))
	if err == nil || !strings.Contains(err.Error(), "connection closed during CSV INSERT") {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(err.Error(), "10.0.0.1") || strings.Contains(err.Error(), "connection reset by peer") {
		t.Fatalf("leaked driver text: %v", err)
	}
}

func TestPreflightAllowingResumeAcceptsCompletePrefix(t *testing.T) {
	manifest := cutoverManifestForTargetTests()
	for i := range manifest.Tables {
		manifest.Tables[i].RowCount = 1
	}
	querier := resumeTargetQuerier{complete: map[string]bool{"staffs": true}}
	if err := PreflightCutoverTargetAllowingResume(context.Background(), querier, manifest, validCutoverSeeds()); err != nil {
		t.Fatal(err)
	}
	err := PreflightCutoverTarget(context.Background(), occupiedBandTargetQuerier{}, manifest, validCutoverSeeds())
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("formal preflight error = %v, want occupied", err)
	}
}

func TestPreflightAllowingResumeRejectsCompleteTableAfterEmpty(t *testing.T) {
	manifest := cutoverManifestForTargetTests()
	for i := range manifest.Tables {
		manifest.Tables[i].RowCount = 1
	}
	err := PreflightCutoverTargetAllowingResume(
		context.Background(),
		resumeTargetQuerier{complete: map[string]bool{"procedures": true}},
		manifest,
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) || !strings.Contains(err.Error(), "procedures") {
		t.Fatalf("error = %v, want occupied procedures after empty staffs", err)
	}
}

func TestApplyCutoverCommittingEachTableCommitsImportedTablesAndFinalVerify(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	session := &fakeCutoverSession{}
	result, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err != nil {
		t.Fatal(err)
	}
	wantTx := len(CutoverTableSpecs()) + 1
	if len(session.txs) != wantTx {
		t.Fatalf("transactions = %d, want %d", len(session.txs), wantTx)
	}
	copyCalls := 0
	for i, tx := range session.txs {
		if !tx.committed {
			t.Fatalf("transaction %d was not committed", i)
		}
		copyCalls += tx.copyCalls
	}
	if copyCalls != len(CutoverTableSpecs()) {
		t.Fatalf("copy calls = %d, want %d", copyCalls, len(CutoverTableSpecs()))
	}
	if result.Counts["staffs"] != 1 {
		t.Fatalf("staffs count = %d", result.Counts["staffs"])
	}
}

func TestApplyCutoverCommittingEachTableSkipsCompletePrefix(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	session := &fakeCutoverSession{inspectCounts: map[string]int64{"staffs": 1}}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err != nil {
		t.Fatal(err)
	}
	copyCalls := 0
	for _, tx := range session.txs {
		copyCalls += tx.copyCalls
	}
	if copyCalls != len(CutoverTableSpecs())-1 {
		t.Fatalf("copy calls = %d, want %d", copyCalls, len(CutoverTableSpecs())-1)
	}
	if len(session.txs) != len(CutoverTableSpecs()) {
		t.Fatalf("transactions = %d, want %d (skipped staffs + final verify)", len(session.txs), len(CutoverTableSpecs()))
	}
}

func TestApplyCutoverCommittingEachTableRejectsCompleteTableAfterEmpty(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	session := &fakeCutoverSession{inspectCounts: map[string]int64{"procedures": 1}}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) || !strings.Contains(err.Error(), "procedures") {
		t.Fatalf("error = %v, want occupied procedures", err)
	}
}

type fakeCutoverSession struct {
	inspectCounts map[string]int64
	txs           []*fakeCutoverTransaction
}

func (s *fakeCutoverSession) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	switch {
	case strings.Contains(query, "clinic_id <> $3"):
		return staticRow{values: []any{int64(0)}}
	case strings.Contains(query, "count(*)") && strings.Contains(query, "id >= $1 AND id < $2"):
		name := tableNameFromBandQuery(query)
		if s.inspectCounts != nil {
			return staticRow{values: []any{s.inspectCounts[name]}}
		}
		return staticRow{values: []any{int64(0)}}
	default:
		return validTargetQuerier{}.QueryRow(context.Background(), query)
	}
}

func (s *fakeCutoverSession) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("SELECT 1"), nil
}

func (s *fakeCutoverSession) Begin(context.Context, pgx.TxIsoLevel) (cutoverTransaction, error) {
	tx := &fakeCutoverTransaction{}
	s.txs = append(s.txs, tx)
	return tx, nil
}

type resumeTargetQuerier struct {
	complete map[string]bool
}

func (q resumeTargetQuerier) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	switch {
	case strings.Contains(query, "clinic_id <> $3"):
		return staticRow{values: []any{int64(0)}}
	case strings.Contains(query, "count(*)") && strings.Contains(query, "id >= $1 AND id < $2"):
		name := tableNameFromBandQuery(query)
		if q.complete[name] {
			return staticRow{values: []any{int64(1)}}
		}
		return staticRow{values: []any{int64(0)}}
	default:
		return validTargetQuerier{}.QueryRow(context.Background(), query)
	}
}

func tableNameFromBandQuery(query string) string {
	best := ""
	for _, spec := range CutoverTableSpecs() {
		needle := "FROM " + pgx.Identifier{spec.Name}.Sanitize()
		if strings.Contains(query, needle) && len(spec.Name) >= len(best) {
			best = spec.Name
		}
	}
	return best
}
