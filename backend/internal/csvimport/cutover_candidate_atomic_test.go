package csvimport

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Major 4266618373: a same-clinic medicines name INSERT between the name
// probe and the medicines COPY would fail only the medicines table on the
// per-table path, leaving staffs/procedures committed — a prefix that
// fresh-only decision A then rejects. Candidate22 therefore applies in ONE
// atomic transaction (existing applyCutoverWithBegin closure over the SAME
// session Begin): any failure rolls the whole bundle back with zero
// committed transactions.

// candidateConflictSession fails exactly the medicines COPY.
type candidateConflictSession struct {
	candidateFreshApplySession
}

func (s *candidateConflictSession) Begin(ctx context.Context, iso pgx.TxIsoLevel) (cutoverTransaction, error) {
	tx, err := s.candidateFreshApplySession.Begin(ctx, iso)
	if err == nil {
		tx.(*candidateAtomicTransaction).copyErrorContains = pgx.Identifier{cutoverCopyStagingRelation("medicines")}.Sanitize()
	}
	return tx, err
}

func TestCandidateAtomicMedicinesConflictRollsBackEverything(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &candidateConflictSession{}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil {
		t.Fatal("applyCutoverCommittingEachTable accepted a medicines copy conflict")
	}
	if !strings.Contains(err.Error(), "medicines") {
		t.Fatalf("error=%v, want medicines COPY conflict", err)
	}
	copied, rolledBack := 0, false
	for _, tx := range session.txs {
		copied += tx.copyCalls
		rolledBack = rolledBack || (tx.rolledBack && !tx.committed)
	}
	if copied != 2 || !rolledBack {
		t.Fatalf("prior COPYs=%d rollback=%v, want staffs/procedures COPYs and rollback", copied, rolledBack)
	}
	committed := 0
	for _, tx := range session.txs {
		if tx.committed {
			committed++
		}
	}
	if committed != 0 {
		t.Fatalf("committed transactions = %d, want 0 (whole bundle must roll back)", committed)
	}
}

func TestCandidateAtomicFreshAppliesInSingleTransaction(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &candidateFreshApplySession{}
	result, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err != nil {
		t.Fatal(err)
	}
	want := len(CutoverTableSpecs(CutoverProfileMedicinesCandidate))
	if len(session.txs) != 1 {
		t.Fatalf("transactions = %d, want 1 atomic transaction", len(session.txs))
	}
	tx := session.txs[0]
	if !tx.committed {
		t.Fatal("atomic transaction was not committed")
	}
	copyCalls := 0
	for _, sqlText := range tx.copySQLs {
		for _, spec := range CutoverTableSpecs(CutoverProfileMedicinesCandidate) {
			if strings.Contains(sqlText, spec.Name) {
				copyCalls++
				break
			}
		}
	}
	if copyCalls != want {
		t.Fatalf("distinct candidate table COPYs = %d, want %d", copyCalls, want)
	}
	if result.Counts["medicines"] != 1 {
		t.Fatalf("medicines count = %d, want 1", result.Counts["medicines"])
	}
	// The advisory xact lock and the in-transaction empty-band recheck run
	// inside the SAME transaction, before any COPY.
	locked, tablesLocked := false, false
	for _, q := range tx.execQueries {
		if strings.Contains(q, "pg_advisory_xact_lock") {
			locked = true
		}
		if strings.HasPrefix(q, "LOCK TABLE ") && strings.Contains(q, `"medicines"`) && strings.Contains(q, "SHARE ROW EXCLUSIVE") {
			tablesLocked = true
		}
		if strings.HasPrefix(q, "COPY ") && (!locked || !tablesLocked) {
			t.Fatal("COPY preceded same-transaction locks")
		}
	}
	if !locked {
		t.Fatal("atomic transaction missing pg_advisory_xact_lock")
	}
	if !tablesLocked {
		t.Fatal("atomic transaction missing all-table lock")
	}
}
