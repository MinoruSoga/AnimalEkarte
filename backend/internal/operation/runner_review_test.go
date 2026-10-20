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

// TestExecute_UnknownRecordingFailureFailsClosed pins the coderabbitai finding
// "return result-unavailable when recording the unknown outcome fails": a
// failed attempt-record never turns into a retriable failure, because a retry
// would re-run Apply on top of an effect whose outcome nobody knows. The
// recording error is logged and the response stops the binding.
func TestExecute_UnknownRecordingFailureFailsClosed(t *testing.T) {
	f := newR2Fixture(t)
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("send response lost")})
	f.store.failRecordOutcome(errors.New("attempt store unavailable"))

	out, err := f.runner.Execute(t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, out.Status)
	require.Nil(t, out.Result)
	require.Equal(t, 1, f.owner.applyCount())
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	require.Zero(t, f.store.attemptCount(stored.Binding.ID))
}

// TestExecute_AppliesTheFingerprintedContent pins the codex P1 finding "bind
// the applied payload to the fingerprinted content": Apply receives the exact
// Request whose Content was fingerprinted, so the owner write and the stored
// receipt cannot diverge from the bound content.
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
	require.Equal(t, f.content.fields, f.owner.lastReq.Content.fields)
}
