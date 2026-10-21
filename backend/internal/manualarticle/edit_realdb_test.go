//go:build operationrealdb

package manualarticle

// edit_realdb_test.go — real-DB §7 proofs for EMR-502 existing-article edit:
// audit-failure atomicity, cross-clinic stale-CAS + non-disclosure, concurrent
// same-key claims, and replay stability after a later edit. NOT_RUN — same
// opt-in tag/harness as internal/operation/store_realdb_test.go:
//
//   go test -tags operationrealdb -count=1 ./internal/manualarticle/...
//
// The tag exists because internal/testdb does not compile on main after PR #729;
// this file has never been type-checked or run. Reuses the package's unit-test
// helpers (newCmdForTest, reqForTest, resolvedSourceVersion, mockEditTxAudit).

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupEditRealDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.OperationBinding{},
		&model.OperationResult{},
		&model.OperationAttemptOutcome{},
		&model.ManualArticle{},
		&model.ManualArticleVersion{},
		&model.ManualArticleOperation{},
	))
	// Mirror 001_init.sql scope key + 005 anchor table (no GORM model exists for it).
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS operation_bindings_scope_key_test
		ON operation_bindings (clinic_id, owner_kind, actor_kind, actor_ref, intent, operation_key)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS manual_article_operation_targets (
		clinic_id bigint NOT NULL, binding_id bigint NOT NULL, article_id bigint NOT NULL,
		PRIMARY KEY (clinic_id, binding_id))`).Error)
	require.NoError(t, db.Exec(`TRUNCATE TABLE manual_article_operation_targets,
		manual_article_operations, manual_article_versions, manual_articles,
		operation_attempt_outcomes, operation_results, operation_bindings RESTART IDENTITY CASCADE`).Error)
	testdb.SeedClinicsForFK(t, db, 1, 2)
	return db
}

// seedEditArticle creates article 42 (screens/intro, version 3, current=7,
// published=9) plus its resolved-source current content version — matching the
// package helper fixtures (newCmdForTest expects article "42", expected 3/7).
func seedEditArticle(t *testing.T, db *gorm.DB) {
	t.Helper()
	currentID := uint64(7)
	publishedID := uint64(9)
	article := &model.ManualArticle{
		ID:                 42,
		Category:           model.ManualCategoryScreens,
		Slug:               "intro",
		Version:            3,
		CurrentVersionID:   &currentID,
		PublishedVersionID: &publishedID,
	}
	require.NoError(t, db.Create(article).Error)
	require.NoError(t, db.Create(resolvedSourceVersion(7)).Error)
}

func editTestFingerprinter(t *testing.T) *operation.Fingerprinter {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = 0x42
	}
	fp, err := operation.NewFingerprinter("k1", map[string][]byte{"k1": key})
	require.NoError(t, err)
	return fp
}

func editAllowGrant(context.Context, uint64, uint64) (bool, error) { return true, nil }

func editScopes(clinicID uint64, actorRef string) operation.Scope {
	return operation.Scope{
		ClinicID:  clinicID,
		OwnerKind: OwnerKind,
		ActorKind: operation.ActorStaff,
		ActorRef:  actorRef,
		Intent:    IntentEdit,
	}
}

func editVersionCount(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&model.ManualArticleVersion{}).Where("article_id = 42").Count(&n).Error)
	return n
}

func editAnchorCount(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Table("manual_article_operation_targets").Count(&n).Error)
	return n
}

// §7-a: a failing EditTxAudit aborts the owner tx — no new content version,
// version/current pointer unchanged, no history row, no receipt; the claim-time
// anchor row stays (it belongs to the binding, not the outcome).
func TestEditRealDB_AuditFailureRollsBackOwnerWrites(t *testing.T) {
	db := setupEditRealDB(t)
	seedEditArticle(t, db)
	runner := operation.NewRunner(db, editTestFingerprinter(t))
	repo := New(db)
	cmd := newCmdForTest(t)
	req := reqForTest(t, cmd)

	p := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{err: errors.New("audit down")},
		1, 7, false, model.ManualCategoryScreens, "intro", cmd)
	outcome, err := runner.Execute(context.Background(), req, p)
	require.True(t, err != nil || outcome.Status != operation.StatusSucceeded,
		"audit failure must not report success (outcome=%+v err=%v)", outcome, err)

	require.Equal(t, int64(1), editVersionCount(t, db), "no new content version row")
	var reloaded model.ManualArticle
	require.NoError(t, db.First(&reloaded, 42).Error)
	require.Equal(t, int64(3), reloaded.Version)
	require.NotNil(t, reloaded.CurrentVersionID)
	require.Equal(t, uint64(7), *reloaded.CurrentVersionID)
	var ops, results int64
	require.NoError(t, db.Model(&model.ManualArticleOperation{}).Count(&ops).Error)
	require.Equal(t, int64(0), ops, "no history row")
	require.NoError(t, db.Model(&model.OperationResult{}).Count(&results).Error)
	require.Equal(t, int64(0), results, "no receipt row")
	require.Equal(t, int64(1), editAnchorCount(t, db), "anchor row belongs to the claim and stays")
}

// §7-b: two clinics edit the same global article with the same stale
// expected_version — one succeeds, the loser gets version-conflict; the loser
// clinic's Lookup of the winner's key is scope-not-found (non-disclosure).
func TestEditRealDB_CrossClinicStaleCASAndLookupScope(t *testing.T) {
	db := setupEditRealDB(t)
	seedEditArticle(t, db)
	runner := operation.NewRunner(db, editTestFingerprinter(t))
	repo := New(db)

	cmdA := newCmdForTest(t)
	pA := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
		1, 7, false, model.ManualCategoryScreens, "intro", cmdA)
	outA, err := runner.Execute(context.Background(), reqForTest(t, cmdA), pA)
	require.NoError(t, err)
	require.Equal(t, operation.StatusSucceeded, outA.Status)

	cmdB := newCmdForTest(t) // identical stale expected_version=3 / content_version_id=7
	pB := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
		2, 8, false, model.ManualCategoryScreens, "intro", cmdB)
	outB, err := runner.Execute(context.Background(), reqForTest(t, cmdB), pB)
	require.NoError(t, err, "a rejected attempt reports its status, not an error")
	require.Equal(t, operation.StatusVersionConflict, outB.Status)

	// Clinic B must not see clinic A's binding — scope-not-found, not 404 disclosure.
	pLookup := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
		2, 8, false, model.ManualCategoryScreens, "intro", nil)
	outLookup, err := runner.Lookup(context.Background(), editScopes(2, "8"), cmdA.OperationKey, pLookup)
	require.NoError(t, err)
	require.Equal(t, operation.StatusScopeNotFound, outLookup.Status)
	require.Nil(t, outLookup.Result)
}

// §7-c: concurrent same-key PUT (same clinic/actor/content) — exactly one
// anchor row, one success receipt, the second is replayed or operation-in-progress.
func TestEditRealDB_ConcurrentSameKeySingleAnchor(t *testing.T) {
	db := setupEditRealDB(t)
	seedEditArticle(t, db)
	runner := operation.NewRunner(db, editTestFingerprinter(t))
	repo := New(db)
	cmd := newCmdForTest(t)
	req := reqForTest(t, cmd)

	var wg sync.WaitGroup
	outcomes := make(chan operation.Outcome, 2)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
				1, 7, false, model.ManualCategoryScreens, "intro", cmd)
			outcome, err := runner.Execute(context.Background(), req, p)
			errs <- err
			outcomes <- outcome
		}()
	}
	wg.Wait()
	close(outcomes)
	close(errs)

	// 競合は status で報告される — error は実障害のみ。両 goroutine は nil error。
	for err := range errs {
		require.NoError(t, err)
	}

	require.Equal(t, int64(1), editAnchorCount(t, db), "one binding → one anchor row")
	var results int64
	require.NoError(t, db.Model(&model.OperationResult{}).Count(&results).Error)
	require.Equal(t, int64(1), results, "one success receipt")

	statuses := map[operation.Status]int{}
	count := 0
	for out := range outcomes {
		statuses[out.Status]++
		count++
	}
	require.Equal(t, 2, count)
	require.Equal(t, 1, statuses[operation.StatusSucceeded], "exactly one success")
	require.Equal(t, 1,
		statuses[operation.StatusReplayed]+statuses[operation.StatusInProgress],
		"the loser replays the receipt or reports operation-in-progress")
	// Apply は 1 回のみ — seeded version 7 plus exactly one inserted version row.
	require.Equal(t, int64(2), editVersionCount(t, db))
}

// §7-d: success then same-key replay after a later edit by a different key —
// replay must return the original receipt refs/versions, not the latest state.
func TestEditRealDB_ReplayReturnsOriginalReceipt(t *testing.T) {
	db := setupEditRealDB(t)
	seedEditArticle(t, db)
	runner := operation.NewRunner(db, editTestFingerprinter(t))
	repo := New(db)

	cmd1 := newCmdForTest(t)
	p1 := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
		1, 7, false, model.ManualCategoryScreens, "intro", cmd1)
	out1, err := runner.Execute(context.Background(), reqForTest(t, cmd1), p1)
	require.NoError(t, err)
	require.Equal(t, operation.StatusSucceeded, out1.Status)
	require.NotNil(t, out1.Result)
	newVersionID, err := strconv.ParseUint(out1.Result.ResultRefs["content_version_id"], 10, 64)
	require.NoError(t, err)

	// Later edit by a different operation key: expected 4 / new content id.
	cmd2 := newCmdForTest(t)
	cmd2.OperationKey = uuid.New()
	cmd2.ExpectedVersion = 4
	cmd2.ExpectedContentVersionID = newVersionID
	p2 := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
		1, 7, false, model.ManualCategoryScreens, "intro", cmd2)
	out2, err := runner.Execute(context.Background(), reqForTest(t, cmd2), p2)
	require.NoError(t, err)
	require.Equal(t, operation.StatusSucceeded, out2.Status)

	// Re-run the first key: replay must surface the original receipt.
	pReplay := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
		1, 7, false, model.ManualCategoryScreens, "intro", cmd1)
	outReplay, err := runner.Execute(context.Background(), reqForTest(t, cmd1), pReplay)
	require.NoError(t, err)
	require.Equal(t, operation.StatusReplayed, outReplay.Status)
	require.NotNil(t, outReplay.Result)
	require.Equal(t, out1.Result.ResultRefs["content_version_id"], outReplay.Result.ResultRefs["content_version_id"])
	require.Equal(t, "4", outReplay.Result.CommittedVersions["resulting_version"])
}

// §7-e: 51 edits with distinct keys — the first key's Lookup still returns the
// original receipt; versions referenced by operation history / receipts /
// current / published pointers are never pruned.
func TestEditRealDB_FiftyOneEditsKeepOriginalReceipt(t *testing.T) {
	db := setupEditRealDB(t)
	seedEditArticle(t, db)
	runner := operation.NewRunner(db, editTestFingerprinter(t))
	repo := New(db)

	expectedVersion := int64(3)
	expectedContentID := uint64(7)
	var firstKey uuid.UUID
	var firstRefs map[string]string
	for i := 0; i < 51; i++ {
		cmd := newCmdForTest(t)
		cmd.OperationKey = uuid.New()
		cmd.ExpectedVersion = expectedVersion
		cmd.ExpectedContentVersionID = expectedContentID
		p := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
			1, 7, false, model.ManualCategoryScreens, "intro", cmd)
		out, err := runner.Execute(context.Background(), reqForTest(t, cmd), p)
		require.NoError(t, err, "edit %d", i)
		require.Equal(t, operation.StatusSucceeded, out.Status)
		if i == 0 {
			firstKey = cmd.OperationKey
			firstRefs = out.Result.ResultRefs
		}
		vid, err := strconv.ParseUint(out.Result.ResultRefs["content_version_id"], 10, 64)
		require.NoError(t, err)
		expectedVersion++
		expectedContentID = vid
	}

	// 最初の key の Lookup は原 receipt を返す（現在 pointer から再構成しない）。
	pLookup := NewEditParticipant(repo, editAllowGrant, &mockEditTxAudit{},
		1, 7, false, model.ManualCategoryScreens, "intro", nil)
	out, err := runner.Lookup(context.Background(), editScopes(1, "7"), firstKey, pLookup)
	require.NoError(t, err)
	require.Equal(t, operation.StatusSucceeded, out.Status)
	require.NotNil(t, out.Result)
	require.Equal(t, firstRefs["content_version_id"], out.Result.ResultRefs["content_version_id"])
	require.Equal(t, "4", out.Result.CommittedVersions["resulting_version"])

	// operation 履歴 / receipt / current・published pointer が参照する版は全て残存。
	var ids []uint64
	require.NoError(t, db.Model(&model.ManualArticleOperation{}).
		Where("content_version_id IS NOT NULL").Pluck("content_version_id", &ids).Error)
	var missing int64
	for _, id := range ids {
		var n int64
		require.NoError(t, db.Model(&model.ManualArticleVersion{}).Where("id = ?", id).Count(&n).Error)
		if n == 0 {
			missing++
		}
	}
	require.Equal(t, int64(0), missing, "history-referenced versions must exist")
	var article model.ManualArticle
	require.NoError(t, db.First(&article, 42).Error)
	var n int64
	require.NoError(t, db.Model(&model.ManualArticleVersion{}).Where("id = ?", *article.CurrentVersionID).Count(&n).Error)
	require.Equal(t, int64(1), n, "current pointer must resolve")
	require.NoError(t, db.Model(&model.ManualArticleVersion{}).Where("id = ?", *article.PublishedVersionID).Count(&n).Error)
	require.Equal(t, int64(1), n, "published pointer must resolve")
}
