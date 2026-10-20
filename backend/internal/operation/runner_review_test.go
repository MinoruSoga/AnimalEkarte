package operation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests added for the 2026-10-20 PR #736 review round. Each test names the
// review finding it pins so the review trail stays traceable.

// TestExecute_GateBusyRecordsNothingAndReportsInProgress pins the codex P2
// finding "map shared-gate contention to operation-in-progress": an owner gate
// that cannot take its lock reports *Busy, the runner answers with the same
// operation-in-progress contract as the binding row lock, and no failure
// attempt is recorded over the contention.
func TestExecute_GateBusyRecordsNothingAndReportsInProgress(t *testing.T) {
	f := newR2Fixture(t)
	f.owner.failGateWith(&Busy{Cause: errors.New("owner gate row is locked")})

	out, err := f.runner.Execute(t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusInProgress, out.Status)
	require.Nil(t, out.Result)
	require.Zero(t, f.owner.applyCount())
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	require.Zero(t, f.store.attemptCount(stored.Binding.ID))
}

// TestExecute_UnknownDisclosureRequiresCurrentGrant pins the codex P2 finding
// "recheck authorization before exposing an unknown state": a stored unknown is
// disclosed through the current grant like every stored state, so a revoked
// actor sees not-authorized instead of learning that the binding is stuck.
func TestExecute_UnknownDisclosureRequiresCurrentGrant(t *testing.T) {
	f := newR2Fixture(t)
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("send response lost")})
	ctx := t.Context()
	req := Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}

	out, err := f.runner.Execute(ctx, req, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, out.Status)

	// Same key, current CAS, but the grant is gone now.
	f.owner.revoke()
	blocked, err := f.runner.Execute(ctx, req, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusNotAuthorized, blocked.Status)
	require.Nil(t, blocked.Result)
}

// TestLookup_BusyBindingRollsBackAbortedTransaction pins the coderabbitai
// finding on runner.go settledState: NOWAIT 55P03 aborts the transaction block,
// so the busy error must leave the callback (rollback) and be classified as
// operation-in-progress outside the transaction. With the old code the callback
// returned nil, the transactor tried to commit an aborted block, and Lookup
// surfaced that commit error instead of operation-in-progress.
func TestLookup_BusyBindingRollsBackAbortedTransaction(t *testing.T) {
	f := newR2Fixture(t)
	// A binding exists without a success: the first attempt is rejected.
	f.owner.setVersions("2", "1")
	rejected, err := f.runner.Execute(t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusVersionConflict, rejected.Status)

	// Another attempt holds the binding row while the lookup runs.
	f.store.holdBinding(f.key)
	looked, err := f.runner.Lookup(t.Context(), f.scope, f.key, allowLookup{})
	require.NoError(t, err)
	require.Equal(t, StatusInProgress, looked.Status)
	require.Nil(t, looked.Result)
}

// TestExecute_UnknownCommitsUnderLockBeforeRelease pins the codex P1 /
// coderabbitai finding "keep the binding stopped while recording unknown
// effects": the unknown row is recorded inside the business transaction while
// the binding row lock is still held, so no concurrent execution can acquire
// the binding and re-run Apply in the window between the rollback and the
// unknown record. The owner write is undone (savepoint) and the versions stay
// untouched.
func TestExecute_UnknownCommitsUnderLockBeforeRelease(t *testing.T) {
	f := newR2Fixture(t)
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("send response lost")})
	req := Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}

	out, err := f.runner.Execute(t.Context(), req, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, out.Status)
	require.Equal(t, 1, f.owner.applyCount())
	// The savepoint undid the owner write of the rolled-back attempt.
	require.Equal(t, "1", f.owner.version("billing_version"))
	// The stop state was recorded inside the open transaction (under the
	// binding lock), not after the rollback released it.
	require.True(t, f.store.unknownWasRecordedInTx())
	// The transaction itself must not have rolled back: a rollback would have
	// released the binding lock before the stop state became durable — the
	// exact race window the review flagged.
	require.Zero(t, f.store.rollbackCount())
	// The apply savepoint was created before the owner write ran, so rolling
	// back to it actually undid the write (a savepoint created after the
	// writes would undo nothing — review round 4).
	markSeq, marked := f.store.applySavepointOrdering()
	require.True(t, marked)
	require.Greater(t, f.owner.applyOrderSeq(), markSeq)
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	require.Equal(t, 1, f.store.attemptCount(stored.Binding.ID))

	// A later execution of the same key observes the committed unknown and
	// never reaches Apply again.
	retried, err := f.runner.Execute(t.Context(), req, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, retried.Status)
	require.Equal(t, 1, f.owner.applyCount())
}

// TestExecute_UnknownPersistenceFailureKeepsResponseStopped pins the second
// half of the same finding: when the unknown persistence keeps failing, only
// the current response stops. There is no durable stop state, so a later
// execution re-runs Apply — the documented residual that a durable
// binding-owned stop state (shared 001 DDL, owned by the storage-contract
// tickets) would remove. This test exists to make the residual explicit, not
// to endorse it.
func TestExecute_UnknownPersistenceFailureKeepsResponseStopped(t *testing.T) {
	f := newR2Fixture(t)
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("send response lost")})
	f.store.failRecordOutcome(errors.New("attempt store unavailable"))
	req := Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}

	// Both the in-transaction record and the outside-transaction fallback
	// fail: the response still stops, no attempt row exists.
	out, err := f.runner.Execute(t.Context(), req, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, out.Status)
	require.Equal(t, 1, f.owner.applyCount())
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	require.Zero(t, f.store.attemptCount(stored.Binding.ID))

	// Residual (documented): with no durable stop state, the later execution
	// re-runs Apply. Removing this requires the binding-owned stop state DDL.
	retried, err := f.runner.Execute(t.Context(), req, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, retried.Status)
	require.Equal(t, 2, f.owner.applyCount())
}

// TestExecute_AppliesTheFingerprintedContent pins the codex P1 finding "bind
// the applied payload to the fingerprinted content": Apply receives the exact
// Request whose Content was fingerprinted, and the content values are readable
// by an owner adapter through the exported accessor — so the owner write and
// the stored receipt cannot diverge from the bound content.
func TestExecute_AppliesTheFingerprintedContent(t *testing.T) {
	f := newR2Fixture(t)

	out, err := f.runner.Execute(t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, out.Status)
	require.NotNil(t, f.owner.lastReq)
	require.Equal(t, f.key, f.owner.lastReq.Key)
	require.Equal(t, f.scope, f.owner.lastReq.Scope)
	// Read through the exported accessor, exactly like an adapter in another
	// package would.
	require.Equal(t, f.content.Fields(), f.owner.lastReq.Content.Fields())
}

// TestExecute_KeyConflictHidesFromRevokedGrant pins the codex P2 finding
// "authorize before reporting operation-key conflicts": the conflict answer is
// an equality oracle over the fingerprinted content, so a revoked actor must
// see not-authorized instead of being able to distinguish a wrong content
// guess from the stored one.
func TestExecute_KeyConflictHidesFromRevokedGrant(t *testing.T) {
	f := newR2Fixture(t)
	// Bind the key to the fixture content with a first successful execution.
	f.owner.setVersions("2", "1")
	out, err := f.runner.Execute(t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "2", "line_version": "1"},
	}, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, out.Status)

	other, err := NewContent(r483Content("9", "99999"))
	require.NoError(t, err)
	probe := Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       other,
		Preconditions: Preconditions{"billing_version": "2", "line_version": "1"},
	}

	// A current grant sees the conflict.
	conflict, err := f.runner.Execute(t.Context(), probe, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusKeyConflict, conflict.Status)

	// The same probe from a revoked actor learns nothing.
	f.owner.revoke()
	denied, err := f.runner.Execute(t.Context(), probe, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusNotAuthorized, denied.Status)
}
