package operation

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// Fixed fixture for the EMR-483 price-acknowledgement participation (先行例).
//
// The fixture separates the two field groups that R2 requires:
//   - Content: the acknowledged business content (target lines, acknowledged
//     before/after calculation versions, before/after amounts). It is the only
//     fingerprint input besides Scope.OwnerKind and Scope.Intent.
//   - Preconditions: the CAS prerequisites (billing_version, line_version).
//     They are re-verified on every attempt and never fingerprinted.
//
// The fake owner keeps its current billing_version; Apply applies the CAS and
// performs a write that participates in the surrounding transaction.
const fakeBillingID = "42"

// fakeRecordedAt is a fixed receipt timestamp so that a replayed result is
// comparable field by field.
var fakeRecordedAt = time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)

// fakeHMACKeyV1 is a fixed 32-byte test key. Secrets never live in _test.go for
// production use; this value only keys the in-memory fixture.
var fakeHMACKeyV1 = []byte("0123456789abcdef0123456789abcdef")

// TestC06Fixtures_R2 is the fixed fixture set for the C06 selected contract and
// the R2 correction: each case names one boundary and asserts its expected
// result instead of a narrative. Cases share one billing-like participant so
// the six boundaries are distinguishable by name.
func TestC06Fixtures_R2(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{name: "same_content_cas_only_updated", run: r2SameContentCASOnlyUpdated},
		{name: "ack_before_after_version_changed", run: r2AckBeforeAfterVersionChanged},
		{name: "same_key_different_content", run: r2SameKeyDifferentContent},
		{name: "succeeded_result_resent", run: r2SucceededResultResent},
		{name: "unknown_result_lookup", run: r2UnknownResultLookup},
		{name: "current_permission_revoked", run: r2CurrentPermissionRevoked},
	}
	for _, tc := range cases {
		t.Run(tc.name, tc.run)
	}
}

// r2SameContentCASOnlyUpdated: a definite rejection caused by a stale CAS does
// not record a business success, and the human may retry the same key and the
// same content with current preconditions. The fingerprint must not change.
func r2SameContentCASOnlyUpdated(t *testing.T) {
	f := newR2Fixture(t)
	// Another user's write already advanced the owner version to 2.
	f.owner.setVersions("2", "1")
	stale := Preconditions{"billing_version": "1", "line_version": "1"}
	current := Preconditions{"billing_version": "2", "line_version": "1"}

	// Act: the first attempt carries the stale CAS.
	first := f.execute(f.content, stale)

	// Assert: version-conflict, no success receipt, one Apply call.
	require.Equal(t, StatusVersionConflict, first.Status)
	require.Nil(t, first.Result)
	require.Equal(t, 1, f.owner.applyCount())

	// Assert: the binding and its fingerprint survive the rejected attempt.
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	sum, keyVersion := f.fp.Compute(f.scope, f.content)
	require.Equal(t, sum, stored.Fingerprint)
	require.Equal(t, keyVersion, stored.KeyVersion)
	require.False(t, f.store.hasResult(stored.Binding.ID))

	// Assert: the rejection is an attempt outcome, not a success.
	attempt, ok := f.store.latestAttempt(stored.Binding.ID)
	require.True(t, ok)
	require.Equal(t, OutcomeRejected, attempt.Outcome)
	require.Equal(t, first.Status, attempt.Status)

	// Act: after re-checking current values the human retries the same key with
	// the same content and the current preconditions.
	second := f.execute(f.content, current)

	// Assert: the same fingerprint is reused, so the binding is the same row.
	require.Equal(t, StatusSucceeded, second.Status)
	require.NotNil(t, second.Result)
	require.Equal(t, 2, f.owner.applyCount())
	reused, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	require.Equal(t, stored.Binding.ID, reused.Binding.ID)
	require.Equal(t, sum, reused.Fingerprint)
	require.Equal(t, 1, f.store.bindingCount())
	require.Equal(t, fakeBillingID, second.Result.ResultRefs["billing_id"])
	require.Equal(t, "3", second.Result.CommittedVersions["billing_version"])
}

// r2AckBeforeAfterVersionChanged: changing the acknowledged after-calculation
// version is a different business content. The same key is refused and nothing
// is applied; a new key is a new binding and may run because the earlier result
// is already a definite success.
func r2AckBeforeAfterVersionChanged(t *testing.T) {
	f := newR2Fixture(t)

	// Arrange: the first acknowledgement of billing_version 1 succeeded.
	first := f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"})
	require.Equal(t, StatusSucceeded, first.Status)
	require.Equal(t, 1, f.owner.applyCount())
	oldKey := f.key

	// Act: the same key carries only a different ack_after_calc_version.
	changed, err := NewContent(r483Content("8", "12000"))
	require.NoError(t, err)
	conflict := f.execute(changed, Preconditions{"billing_version": "2", "line_version": "1"})

	// Assert: operation-key-conflict, no Apply, no result, no attempt outcome.
	require.Equal(t, StatusKeyConflict, conflict.Status)
	require.Nil(t, conflict.Result)
	require.Equal(t, 1, f.owner.applyCount())
	stored, ok := f.store.bindingFor(f.scope, oldKey)
	require.True(t, ok)
	require.Equal(t, 0, f.store.attemptCount(stored.Binding.ID))
	require.Equal(t, "2", f.owner.version("billing_version"))
	res, hasResult := f.store.resultFor(stored.Binding.ID)
	require.True(t, hasResult)
	require.Equal(t, first.Result, &res)

	// Act: the changed content uses a new key.
	f.key = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	accepted := f.execute(changed, Preconditions{"billing_version": "2", "line_version": "1"})

	// Assert: a separate binding with its own fingerprint.
	require.Equal(t, StatusSucceeded, accepted.Status)
	require.NotNil(t, accepted.Result)
	require.Equal(t, 2, f.owner.applyCount())
	require.Equal(t, 2, f.store.bindingCount())
	newBinding, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	require.NotEqual(t, stored.Fingerprint, newBinding.Fingerprint)
	require.Equal(t, "3", accepted.Result.CommittedVersions["billing_version"])
}

// r2SameKeyDifferentContent: the same key with a different amount is refused
// before any owner write and returns no result.
func r2SameKeyDifferentContent(t *testing.T) {
	f := newR2Fixture(t)

	// Arrange: a successful acknowledgement for the original amount.
	first := f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"})
	require.Equal(t, StatusSucceeded, first.Status)
	require.Equal(t, 1, f.owner.applyCount())

	// Act: same key, different approved amount.
	changed, err := NewContent(r483Content("7", "13000"))
	require.NoError(t, err)
	conflict := f.execute(changed, Preconditions{"billing_version": "2", "line_version": "1"})

	// Assert: refused, Apply count unchanged, no new result.
	require.Equal(t, StatusKeyConflict, conflict.Status)
	require.Nil(t, conflict.Result)
	require.Equal(t, 1, f.owner.applyCount())
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	res, hasResult := f.store.resultFor(stored.Binding.ID)
	require.True(t, hasResult)
	require.Equal(t, first.Result, &res)
	require.Equal(t, 0, f.store.attemptCount(stored.Binding.ID))
	require.Equal(t, "2", f.owner.version("billing_version"))
}

// r2SucceededResultResent: re-sending a succeeded key replays the stored result
// without running the owner write again.
func r2SucceededResultResent(t *testing.T) {
	f := newR2Fixture(t)
	pre := Preconditions{"billing_version": "1", "line_version": "1"}

	first := f.execute(f.content, pre)
	require.Equal(t, StatusSucceeded, first.Status)
	require.NotNil(t, first.Result)
	require.Equal(t, 1, f.owner.applyCount())

	// Act: same key, same content (preconditions are irrelevant on replay).
	resent := f.execute(f.content, pre)

	// Assert: replayed with the stored result and no additional Apply call.
	require.Equal(t, StatusReplayed, resent.Status)
	require.NotNil(t, resent.Result)
	require.Equal(t, first.Result, resent.Result)
	require.Equal(t, 1, f.owner.applyCount())
	require.Equal(t, "2", f.owner.version("billing_version"))
	require.Equal(t, 1, f.store.bindingCount())
}

// r2UnknownResultLookup: an unresolved attempt stops further execution and is
// not rescued with a new key; a new key with the old CAS is still refused by the
// owner CAS, so nothing is acknowledged twice.
func r2UnknownResultLookup(t *testing.T) {
	f := newR2Fixture(t)
	ctx := t.Context()

	// Arrange: an earlier acknowledgement already advanced billing_version to 2.
	acknowledged := f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"})
	require.Equal(t, StatusSucceeded, acknowledged.Status)

	// Arrange: a second key runs with current preconditions, but the owner
	// cannot tell whether its effect happened (e.g. the external send lost its
	// response). The participant reports an indeterminate effect.
	unknownKey := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("database result uncertain")})
	attempted, err := f.runner.Execute(ctx, Request{
		Scope:         f.scope,
		Key:           unknownKey,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "2", "line_version": "1"},
	}, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, attempted.Status)
	require.Nil(t, attempted.Result)
	unknownBinding, ok := f.store.bindingFor(f.scope, unknownKey)
	require.True(t, ok)
	unknownAttempt, ok := f.store.latestAttempt(unknownBinding.Binding.ID)
	require.True(t, ok)
	require.Equal(t, OutcomeUnknown, unknownAttempt.Outcome)
	require.False(t, f.store.hasResult(unknownBinding.Binding.ID))
	f.owner.failApplyWith(nil)
	appliesBefore := f.owner.applyCount()

	// Act: the same key is executed again with current preconditions.
	blocked, err := f.runner.Execute(ctx, Request{
		Scope:         f.scope,
		Key:           unknownKey,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "2", "line_version": "1"},
	}, f.owner)

	// Assert: result-unavailable, no Apply call, no result.
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, blocked.Status)
	require.Nil(t, blocked.Result)
	require.Equal(t, appliesBefore, f.owner.applyCount())

	// Assert: the result lookup reports the same unknown boundary.
	looked, err := f.runner.Lookup(ctx, f.scope, unknownKey, allowLookup{})
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, looked.Status)
	require.Nil(t, looked.Result)

	// Act: a new key with the old CAS.
	retryKey := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	retried, err := f.runner.Execute(ctx, Request{
		Scope:         f.scope,
		Key:           retryKey,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)

	// Assert: the owner CAS refuses it, so the effective version stays advanced.
	require.NoError(t, err)
	require.Equal(t, StatusVersionConflict, retried.Status)
	require.Equal(t, appliesBefore+1, f.owner.applyCount())
	require.Equal(t, "2", f.owner.version("billing_version"))
}

// r2CurrentPermissionRevoked: a stored success is not disclosed when the current
// grant no longer allows the operation, and the outcome carries no result body.
func r2CurrentPermissionRevoked(t *testing.T) {
	f := newR2Fixture(t)
	ctx := t.Context()

	succeeded := f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"})
	require.Equal(t, StatusSucceeded, succeeded.Status)

	// Control: the current grant may read the stored result.
	allowed, err := f.runner.Lookup(ctx, f.scope, f.key, allowLookup{})
	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, allowed.Status)
	require.NotNil(t, allowed.Result)

	// Act: the grant was revoked after the success.
	revoked, err := f.runner.Lookup(ctx, f.scope, f.key, denyLookup{})

	// Assert: not-authorized without a result body.
	require.NoError(t, err)
	require.Equal(t, StatusNotAuthorized, revoked.Status)
	require.Nil(t, revoked.Result)

	// Act: the same key is re-sent through Execute after the revocation.
	f.owner.revoke()
	resent := f.execute(f.content, Preconditions{"billing_version": "2", "line_version": "1"})

	// Assert: the replay path re-checks the current grant too and runs nothing.
	require.Equal(t, StatusNotAuthorized, resent.Status)
	require.Nil(t, resent.Result)
	require.Equal(t, 1, f.owner.applyCount())
}

// TestExecute_RejectsAmbientTransaction pins the outermost-transaction rule: the
// helper must not join a caller transaction, so an ambient tx is a fail-closed
// error and nothing is written.
func TestExecute_RejectsAmbientTransaction(t *testing.T) {
	f := newR2Fixture(t)
	ctx := persistence.WithTxValue(t.Context(), &gorm.DB{})

	out, err := f.runner.Execute(ctx, Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)

	require.Error(t, err)
	require.Equal(t, 0, f.owner.applyCount())
	require.Equal(t, 0, f.owner.gateCount())
	require.Equal(t, 0, f.store.bindingCount())
	require.Nil(t, out.Result)
}

// TestExecute_RetryableFailureRecordsOutcomeAndReturnsError pins that a
// non-business failure records retryable_failure in its own transaction, returns
// the error, and leaves no success receipt or owner write behind.
func TestExecute_RetryableFailureRecordsOutcomeAndReturnsError(t *testing.T) {
	f := newR2Fixture(t)
	f.owner.failApplyWith(errors.New("connection reset by peer"))

	out, err := f.runner.Execute(t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)

	require.Error(t, err)
	require.Nil(t, out.Result)
	require.False(t, f.store.hasResultOfScope(f.scope, f.key))

	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	attempt, ok := f.store.latestAttempt(stored.Binding.ID)
	require.True(t, ok)
	require.Equal(t, OutcomeRetryableFailure, attempt.Outcome)

	// The owner write of the failed transaction was rolled back.
	require.Equal(t, "1", f.owner.version("billing_version"))
}

// TestExecute_BindingBusyReturnsInProgress pins that a held binding is reported
// as operation-in-progress and is not recorded as an attempt outcome.
func TestExecute_BindingBusyReturnsInProgress(t *testing.T) {
	f := newR2Fixture(t)
	f.store.holdBinding(f.key)

	out, err := f.runner.Execute(t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusInProgress, out.Status)
	require.Nil(t, out.Result)
	require.Equal(t, 0, f.owner.applyCount())

	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	require.Equal(t, 0, f.store.attemptCount(stored.Binding.ID))
}

// TestLookup_ScopeNotFound pins that a lookup without a binding for the scope
// and key is scope-not-found, not not-succeeded.
func TestLookup_ScopeNotFound(t *testing.T) {
	f := newR2Fixture(t)

	out, err := f.runner.Lookup(t.Context(), f.scope, uuid.MustParse("99999999-9999-4999-8999-999999999999"), allowLookup{})

	require.NoError(t, err)
	require.Equal(t, StatusScopeNotFound, out.Status)
	require.Nil(t, out.Result)
	require.Nil(t, out.Binding)
}

// TestExecute_UnknownStaysStickyAfterGateFailure pins that a failure recorded
// after an unknown outcome cannot release the stop: unknown -> gate rejected ->
// gate recovered -> same key re-sent must still not run the owner write.
func TestExecute_UnknownStaysStickyAfterGateFailure(t *testing.T) {
	f := newR2Fixture(t)
	ctx := t.Context()
	pre := Preconditions{"billing_version": "1", "line_version": "1"}

	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("database result uncertain")})
	first, err := f.runner.Execute(ctx, Request{Scope: f.scope, Key: f.key, Content: f.content, Preconditions: pre}, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, first.Status)
	f.owner.failApplyWith(nil)
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	appliesAfterUnknown := f.owner.applyCount()

	// Act: the gate rejects once, then recovers.
	f.owner.failGateWith(errors.New("gate unavailable"))
	_, err = f.runner.Execute(ctx, Request{Scope: f.scope, Key: f.key, Content: f.content, Preconditions: pre}, f.owner)
	require.NoError(t, err)
	f.owner.failGateWith(nil)
	again, err := f.runner.Execute(ctx, Request{Scope: f.scope, Key: f.key, Content: f.content, Preconditions: pre}, f.owner)

	// Assert: still result-unavailable, never applied, and no failure row was
	// recorded over the unknown outcome.
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, again.Status)
	require.Equal(t, appliesAfterUnknown, f.owner.applyCount())
	attempt, ok := f.store.latestAttempt(stored.Binding.ID)
	require.True(t, ok)
	require.Equal(t, OutcomeUnknown, attempt.Outcome)
}

// TestLookup_SeesSuccessCommittedBeforeLock pins that a success committed
// between the first receipt read and the binding lock is reported as succeeded,
// never as not-succeeded (which would invite a retry).
func TestLookup_SeesSuccessCommittedBeforeLock(t *testing.T) {
	f := newR2Fixture(t)
	ctx := t.Context()
	claimed, err := f.store.ClaimBinding(ctx, f.scope, f.key, []byte("unused-in-fake"), "k1")
	require.NoError(t, err)
	f.store.onLock(func() {
		f.store.commitResultForTest(claimed.Binding.ID, Result{
			ResultRefs: map[string]string{"billing_id": fakeBillingID},
			RecordedAt: fakeRecordedAt,
		})
	})

	out, err := f.runner.Lookup(ctx, f.scope, f.key, allowLookup{})

	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, out.Status)
	require.NotNil(t, out.Result)
	require.Equal(t, fakeBillingID, out.Result.ResultRefs["billing_id"])
}

// TestExecute_GateFailureWithCommittedReceiptRechecksGrant pins that the
// failure path which finds a committed receipt does not disclose it when the
// current grant was revoked.
func TestExecute_GateFailureWithCommittedReceiptRechecksGrant(t *testing.T) {
	f := newR2Fixture(t)
	first := f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"})
	require.Equal(t, StatusSucceeded, first.Status)

	f.owner.revoke()
	f.owner.failGateWith(errors.New("gate unavailable"))
	out, err := f.runner.Execute(t.Context(), Request{
		Scope: f.scope, Key: f.key, Content: f.content,
		Preconditions: Preconditions{"billing_version": "2", "line_version": "1"},
	}, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusNotAuthorized, out.Status)
	require.Nil(t, out.Result)
	require.Equal(t, 1, f.owner.applyCount())
}

// r2Fixture wires the in-memory store, the fake transaction boundary and the
// fake billing owner against the package Runner.
type r2Fixture struct {
	t       *testing.T
	store   *fakeStore
	fp      *Fingerprinter
	owner   *fakeBillingParticipant
	runner  *Runner
	scope   Scope
	key     uuid.UUID
	content Content
}

func newR2Fixture(t *testing.T) *r2Fixture {
	t.Helper()
	store := newFakeStore()
	tx := &fakeTx{store: store}
	fp := newTestFingerprinter(t)
	owner := newFakeBillingParticipant(store)
	content, err := NewContent(r483Content("7", "12000"))
	require.NoError(t, err)

	return &r2Fixture{
		t:       t,
		store:   store,
		fp:      fp,
		owner:   owner,
		runner:  newRunner(store, tx, fp, deterministicKeyFunc()),
		scope:   testScope(),
		key:     uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		content: content,
	}
}

func (f *r2Fixture) execute(content Content, pre Preconditions) Outcome {
	f.t.Helper()
	out, err := f.runner.Execute(f.t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       content,
		Preconditions: pre,
	}, f.owner)
	require.NoError(f.t, err)
	return out
}

// r483Content is the acknowledged business content of one price acknowledgement:
// target lines, the acknowledged before/after calculation versions and the
// before/after amounts. CAS versions are not part of it.
func r483Content(ackAfterCalcVersion, afterAmount string) map[string]string {
	return map[string]string{
		"billing_id":              fakeBillingID,
		"item_ids":                "1001,1002",
		"ack_before_calc_version": "7",
		"ack_after_calc_version":  ackAfterCalcVersion,
		"before_amount":           "12000",
		"after_amount":            afterAmount,
	}
}

func testScope() Scope {
	return Scope{
		ClinicID:  7,
		OwnerKind: "billing",
		ActorKind: ActorStaff,
		ActorRef:  "12",
		Intent:    "billing.price_acknowledgement",
	}
}

func newTestFingerprinter(t *testing.T) *Fingerprinter {
	t.Helper()
	fp, err := NewFingerprinter("k1", map[string][]byte{"k1": fakeHMACKeyV1})
	require.NoError(t, err)
	require.NotNil(t, fp)
	return fp
}

// deterministicKeyFunc returns server-style attempt keys in a stable order.
func deterministicKeyFunc() func() uuid.UUID {
	var seq uint64
	return func() uuid.UUID {
		seq++
		return uuid.MustParse(fmt.Sprintf("00000000-0000-4000-8000-%012d", seq))
	}
}

// fakeAttempt is one recorded operation_attempt_outcomes row.
type fakeAttempt struct {
	AttemptKey uuid.UUID
	Outcome    AttemptOutcome
	Status     Status
}

// fakeJournal holds the undos of one simulated transaction. A failed fn discards
// what it inserted; a successful fn keeps it.
type fakeJournal struct {
	undos []func()
	// savepointIndex is the undo-stack position captured by
	// MarkApplySavepoint (-1 when the savepoint was not created). It models
	// PostgreSQL ROLLBACK TO SAVEPOINT, which undoes only the writes made
	// after the savepoint was established.
	savepointIndex int
}

// errAbortedCommit models PostgreSQL rejecting the COMMIT of a transaction
// block that a NOWAIT lock error already aborted.
var errAbortedCommit = errors.New("fake commit: transaction is aborted")

// fakeTx is the persistence.Transactor used by the fixtures. It runs fn and, when
// fn returns an error, rolls back what fn wrote through fakeStore. A callback
// that returns nil on an aborted transaction block fails the commit, like the
// real PostgreSQL transactor would.
type fakeTx struct {
	store      *fakeStore
	failCommit error
}

func (t *fakeTx) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	journal := t.store.begin()
	t.store.mu.Lock()
	t.store.txAborted = false
	t.store.mu.Unlock()
	if err := fn(ctx); err != nil {
		t.store.rollback(journal)
		t.store.mu.Lock()
		t.store.txRollbacks++
		t.store.mu.Unlock()
		return err
	}
	t.store.mu.Lock()
	aborted := t.store.txAborted
	t.store.mu.Unlock()
	if aborted {
		t.store.rollback(journal)
		return errAbortedCommit
	}
	if t.failCommit != nil {
		t.store.rollback(journal)
		return t.failCommit
	}
	t.store.commit(journal)
	return nil
}

// fakeStore is the in-memory implementation of the unexported store interface.
// It models row identity, uniqueness and transaction rollback of the rows the
// fixture writes; it models nothing else about PostgreSQL.
type fakeStore struct {
	mu       sync.Mutex
	bindings map[string]storedBinding
	results  map[uint64]Result
	outcomes map[uint64][]fakeAttempt
	nextID   uint64

	journals   []*fakeJournal
	held       map[uuid.UUID]bool
	beforeLock func()
	// txAborted models a PostgreSQL transaction block aborted by a NOWAIT
	// lock error: a callback that returns nil afterwards cannot commit.
	txAborted bool
	// recordErr makes every subsequent outcome write (RecordOutcome and
	// RecordUnknownUnderLock) fail, modeling a broken attempt store.
	recordErr error
	// unknownInTx is set when RecordUnknownUnderLock ran with an open
	// transaction, pinning that the stop state is recorded under the lock.
	unknownInTx bool
	// txRollbacks counts transaction rollbacks. The unknown stop state must
	// commit with the owner transaction (rollback count stays zero for that
	// attempt); a rollback would release the binding lock before the stop
	// state is durable — the race the review flagged.
	txRollbacks int
	// opSeq is a monotonic operation counter shared by MarkApplySavepoint and
	// the participant fake, so tests can pin that the savepoint is created
	// before the owner write runs.
	opSeq       int
	lastMarkSeq int
	markSeen    bool
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		bindings: map[string]storedBinding{},
		results:  map[uint64]Result{},
		outcomes: map[uint64][]fakeAttempt{},
		held:     map[uuid.UUID]bool{},
	}
}

// failRecordOutcome makes every subsequent outcome write fail, modeling an
// unavailable attempt store. The residual risk this models (no durable stop
// state when persistence keeps failing) is pinned by
// TestExecute_UnknownPersistenceFailureKeepsResponseStopped.
func (s *fakeStore) failRecordOutcome(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordErr = err
}

// unknownWasRecordedInTx reports whether RecordUnknownUnderLock ran inside an
// open transaction, which is what keeps the stop state durable before the
// binding row lock is released.
func (s *fakeStore) unknownWasRecordedInTx() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.unknownInTx
}

// rollbackCount reports how many transactions rolled back. The unknown stop
// state must commit with the owner transaction, so the attempt that records it
// must not have rolled back at all.
func (s *fakeStore) rollbackCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.txRollbacks
}

// MarkApplySavepoint captures the current undo-stack position of the open
// transaction. RecordUnknownUnderLock later rolls back to it, so only the
// owner writes made after this point are undone.
func (s *fakeStore) MarkApplySavepoint(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n := len(s.journals); n == 0 {
		return errors.New("fake savepoint requires an open transaction")
	}
	journal := s.journals[len(s.journals)-1]
	journal.savepointIndex = len(journal.undos)
	s.lastMarkSeq = s.opSeq
	s.opSeq++
	s.markSeen = true
	return nil
}

// savepointBeforeApply reports whether MarkApplySavepoint ran after the last
// unknown record and before the participant's Apply (order pinned via opSeq).
func (s *fakeStore) applySavepointOrdering() (markSeq int, marked bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastMarkSeq, s.markSeen
}

// RecordUnknownUnderLock models the savepoint-based in-transaction unknown
// record: it rolls back the owner writes made after MarkApplySavepoint and
// appends the unknown attempt row, which then commits with the caller's
// transaction while the binding lock is still held.
func (s *fakeStore) RecordUnknownUnderLock(_ context.Context, b Binding, attemptKey uuid.UUID, status Status) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.recordErr != nil {
		return s.recordErr
	}
	s.unknownInTx = len(s.journals) > 0
	if n := len(s.journals); n == 0 {
		return errors.New("fake unknown record requires an open transaction")
	}
	journal := s.journals[len(s.journals)-1]
	if journal.savepointIndex < 0 {
		return errors.New("fake unknown record requires a prior apply savepoint")
	}
	for i := len(journal.undos) - 1; i >= journal.savepointIndex; i-- {
		journal.undos[i]()
	}
	journal.undos = journal.undos[:journal.savepointIndex]
	s.outcomes[b.ID] = append(s.outcomes[b.ID], fakeAttempt{
		AttemptKey: attemptKey,
		Outcome:    OutcomeUnknown,
		Status:     status,
	})
	s.addRollbackLocked(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		recorded := s.outcomes[b.ID]
		if len(recorded) > 0 {
			s.outcomes[b.ID] = recorded[:len(recorded)-1]
		}
	})
	return nil
}

func (s *fakeStore) ClaimBinding(_ context.Context, scope Scope, key uuid.UUID, sum []byte, keyVersion string) (storedBinding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.bindings[fakeBindingKey(scope, key)]; ok {
		return existing, nil
	}

	s.nextID++
	created := storedBinding{
		Binding:     Binding{ID: s.nextID, ClinicID: scope.ClinicID, Key: key},
		Fingerprint: append([]byte(nil), sum...),
		KeyVersion:  keyVersion,
	}
	s.bindings[fakeBindingKey(scope, key)] = created
	s.addRollbackLocked(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.bindings, fakeBindingKey(scope, key))
	})
	return created, nil
}

func (s *fakeStore) FindBinding(_ context.Context, scope Scope, key uuid.UUID) (storedBinding, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.bindings[fakeBindingKey(scope, key)]
	return existing, ok, nil
}

func (s *fakeStore) LockBindingNoWait(_ context.Context, b Binding) error {
	s.mu.Lock()
	hook := s.beforeLock
	s.beforeLock = nil
	s.mu.Unlock()
	if hook != nil {
		hook()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.held[b.Key] {
		// A NOWAIT lock error aborts the surrounding transaction block.
		s.txAborted = true
		return errBindingBusy
	}
	return nil
}

func (s *fakeStore) LoadResult(_ context.Context, b Binding) (*Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, ok := s.results[b.ID]
	if !ok {
		return nil, nil
	}
	loaded := cloneResult(stored)
	return &loaded, nil
}

func (s *fakeStore) InsertResult(_ context.Context, b Binding, res Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.results[b.ID]; exists {
		return errors.New("operation result already exists for binding")
	}
	s.results[b.ID] = cloneResult(res)
	s.addRollbackLocked(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.results, b.ID)
	})
	return nil
}

func (s *fakeStore) HasUnknownOutcome(_ context.Context, b Binding) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, recorded := range s.outcomes[b.ID] {
		if recorded.Outcome == OutcomeUnknown {
			return true, nil
		}
	}
	return false, nil
}

func (s *fakeStore) RecordOutcome(_ context.Context, b Binding, attemptKey uuid.UUID, outcome AttemptOutcome, status Status) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.recordErr != nil {
		err := s.recordErr
		s.recordErr = nil
		return err
	}
	s.outcomes[b.ID] = append(s.outcomes[b.ID], fakeAttempt{
		AttemptKey: attemptKey,
		Outcome:    outcome,
		Status:     status,
	})
	s.addRollbackLocked(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		recorded := s.outcomes[b.ID]
		if len(recorded) > 0 {
			s.outcomes[b.ID] = recorded[:len(recorded)-1]
		}
	})
	return nil
}

func (s *fakeStore) begin() *fakeJournal {
	s.mu.Lock()
	defer s.mu.Unlock()

	journal := &fakeJournal{savepointIndex: -1}
	s.journals = append(s.journals, journal)
	return journal
}

func (s *fakeStore) commit(journal *fakeJournal) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n := len(s.journals); n > 0 && s.journals[n-1] == journal {
		s.journals = s.journals[:n-1]
	}
}

func (s *fakeStore) rollback(journal *fakeJournal) {
	s.mu.Lock()
	if n := len(s.journals); n > 0 && s.journals[n-1] == journal {
		s.journals = s.journals[:n-1]
	}
	s.mu.Unlock()

	for i := len(journal.undos) - 1; i >= 0; i-- {
		journal.undos[i]()
	}
}

// addRollback registers an undo in the transaction that is currently open. It is
// used by fakes whose owner write participates in the caller transaction.
func (s *fakeStore) addRollback(undo func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addRollbackLocked(undo)
}

func (s *fakeStore) addRollbackLocked(undo func()) {
	if n := len(s.journals); n > 0 {
		s.journals[n-1].undos = append(s.journals[n-1].undos, undo)
	}
}

// onLock runs fn once, just before the next binding lock is taken, to model a
// concurrent attempt committing in between.
func (s *fakeStore) onLock(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.beforeLock = fn
}

// commitResultForTest stores a receipt outside any simulated transaction, as a
// concurrent attempt that already committed.
func (s *fakeStore) commitResultForTest(bindingID uint64, res Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[bindingID] = cloneResult(res)
}

// holdBinding makes the next Nowait lock of the key fail, as a concurrent
// attempt would.
func (s *fakeStore) holdBinding(key uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.held[key] = true
}

func (s *fakeStore) bindingFor(scope Scope, key uuid.UUID) (storedBinding, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.bindings[fakeBindingKey(scope, key)]
	return existing, ok
}

func (s *fakeStore) bindingCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.bindings)
}

func (s *fakeStore) resultFor(bindingID uint64) (Result, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.results[bindingID]
	return stored, ok
}

func (s *fakeStore) hasResult(bindingID uint64) bool {
	_, ok := s.resultFor(bindingID)
	return ok
}

func (s *fakeStore) hasResultOfScope(scope Scope, key uuid.UUID) bool {
	stored, ok := s.bindingFor(scope, key)
	if !ok {
		return false
	}
	return s.hasResult(stored.Binding.ID)
}

func (s *fakeStore) attemptCount(bindingID uint64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.outcomes[bindingID])
}

func (s *fakeStore) latestAttempt(bindingID uint64) (fakeAttempt, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	recorded := s.outcomes[bindingID]
	if len(recorded) == 0 {
		return fakeAttempt{}, false
	}
	return recorded[len(recorded)-1], true
}

func fakeBindingKey(scope Scope, key uuid.UUID) string {
	return fmt.Sprintf("%d|%s|%s|%s|%s|%s",
		scope.ClinicID, scope.OwnerKind, scope.ActorKind, scope.ActorRef, scope.Intent, key)
}

// fakeBillingParticipant is the EMR-483 price-acknowledgement owner fake. It
// re-verifies the CAS under the helper's lock and performs the owner write.
type fakeBillingParticipant struct {
	store *fakeStore

	mu       sync.Mutex
	versions map[string]string
	applyErr error
	gateErr  error
	revoked  bool
	applies  int
	gates    int
	lastReq  *Request
	applySeq int
}

func newFakeBillingParticipant(store *fakeStore) *fakeBillingParticipant {
	return &fakeBillingParticipant{
		store:    store,
		versions: map[string]string{"billing_version": "1", "line_version": "1"},
	}
}

func (p *fakeBillingParticipant) Effects() EffectScope { return EffectsDatabaseOnly }

func (p *fakeBillingParticipant) Gate(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.gates++
	return p.gateErr
}

func (p *fakeBillingParticipant) Apply(_ context.Context, _ Binding, req Request) (Result, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.applies++
	p.applySeq = p.store.opSeq
	captured := req
	p.lastReq = &captured

	pre := req.Preconditions
	if pre["billing_version"] != p.versions["billing_version"] ||
		pre["line_version"] != p.versions["line_version"] {
		return Result{}, &Rejection{Status: StatusVersionConflict}
	}

	// The owner write belongs to the caller transaction and is discarded when
	// that transaction rolls back.
	previous := p.versions["billing_version"]
	p.store.addRollback(func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.versions["billing_version"] = previous
	})
	p.versions["billing_version"] = incrementDecimal(previous)

	if p.applyErr != nil {
		return Result{}, fmt.Errorf("fake owner write failed: %w", p.applyErr)
	}

	return Result{
		ResultRefs: map[string]string{"billing_id": fakeBillingID},
		CommittedVersions: map[string]string{
			"billing_version": p.versions["billing_version"],
			"line_version":    p.versions["line_version"],
		},
		RecordedAt: fakeRecordedAt,
	}, nil
}

// The fixture's persisted owner mapping has one billing root for all bindings.
func (p *fakeBillingParticipant) ResolveLookupTarget(_ context.Context, b Binding, _ *Result) (LookupTarget, error) {
	return LookupTarget{BindingID: b.ID, ClinicID: b.ClinicID, Kind: "billing", Ref: fakeBillingID}, nil
}

func (p *fakeBillingParticipant) AuthorizeLookup(_ context.Context, _ LookupTarget) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.revoked {
		return errors.New("operation grant was revoked")
	}
	return nil
}

func (p *fakeBillingParticipant) revoke() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.revoked = true
}

// applyOrderSeq reports the shared operation counter value captured when Apply
// ran. Compared against the store's savepoint marker it pins that the
// savepoint was created before the owner write.
func (p *fakeBillingParticipant) applyOrderSeq() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.applySeq
}

func (p *fakeBillingParticipant) failGateWith(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.gateErr = err
}

func (p *fakeBillingParticipant) setVersions(billingVersion, lineVersion string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.versions["billing_version"] = billingVersion
	p.versions["line_version"] = lineVersion
}

func (p *fakeBillingParticipant) failApplyWith(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.applyErr = err
}

func (p *fakeBillingParticipant) version(name string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.versions[name]
}

func (p *fakeBillingParticipant) applyCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.applies
}

func (p *fakeBillingParticipant) gateCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.gates
}

// allowLookup is the current grant that still allows reading the result.
type allowLookup struct{}

func (allowLookup) ResolveLookupTarget(_ context.Context, b Binding, _ *Result) (LookupTarget, error) {
	return LookupTarget{BindingID: b.ID, ClinicID: b.ClinicID, Kind: "fixture", Ref: "fixture"}, nil
}
func (allowLookup) AuthorizeLookup(_ context.Context, _ LookupTarget) error {
	return nil
}

// denyLookup is the revoked grant: the stored result must not be disclosed.
type denyLookup struct{ allowLookup }

func (denyLookup) AuthorizeLookup(_ context.Context, _ LookupTarget) error {
	return errors.New("operation grant was revoked")
}

func incrementDecimal(value string) string {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return value
	}
	return strconv.Itoa(parsed + 1)
}

func cloneResult(in Result) Result {
	return Result{
		ResultRefs:        cloneStringMap(in.ResultRefs),
		CommittedVersions: cloneStringMap(in.CommittedVersions),
		EffectSummary:     cloneStringMap(in.EffectSummary),
		RecordedAt:        in.RecordedAt,
	}
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// fakeTx must satisfy the production transaction boundary.
var _ persistence.Transactor = (*fakeTx)(nil)
