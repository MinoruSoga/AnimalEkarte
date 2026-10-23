package operation

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// verifiedRequest builds a request whose VerifiedCommand projects to the fixture's
// default content.
func (f *r2Fixture) verifiedRequest(t *testing.T, project CommandProjector) Request {
	t.Helper()
	vc, err := NewVerifiedCommand(testCommandSchema, mustCommandBytes(t, f.content.Fields()), f.content, project)
	require.NoError(t, err)
	return Request{
		Scope:           f.scope,
		Key:             f.key,
		Content:         f.content,
		Preconditions:   Preconditions{"billing_version": "1", "line_version": "1"},
		VerifiedCommand: vc,
	}
}

// flipAfter returns a projector that is faithful for the first n calls and
// returns a changed value from call n+1 on (a non-deterministic projector).
func flipAfter(n int32, calls *atomic.Int32) CommandProjector {
	return func(schema string, command []byte) (map[string]string, error) {
		fields, err := jsonProjector(schema, command)
		if err != nil {
			return nil, err
		}
		if calls.Add(1) > n {
			fields["after_amount"] = "tampered"
		}
		return fields, nil
	}
}

func TestExecute_NilVerifiedCommandBehavesAsContentOnly(t *testing.T) {
	f := newR2Fixture(t)

	out, err := f.runner.Execute(t.Context(), Request{
		Scope:         f.scope,
		Key:           f.key,
		Content:       f.content,
		Preconditions: Preconditions{"billing_version": "1", "line_version": "1"},
	}, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, out.Status)
	require.Equal(t, 1, f.owner.applyCount())
	require.NotNil(t, f.owner.lastReq)
	require.True(t, f.owner.lastReq.VerifiedCommand.IsZero())
}

func TestExecute_VerifiedCommandReachesApplyUnchanged(t *testing.T) {
	f := newR2Fixture(t)
	req := f.verifiedRequest(t, jsonProjector)

	out, err := f.runner.Execute(t.Context(), req, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, out.Status)
	require.Equal(t, 1, f.owner.applyCount())
	seen := f.owner.lastReq.VerifiedCommand
	require.False(t, seen.IsZero())
	require.Equal(t, req.VerifiedCommand.Schema(), seen.Schema())
	require.Equal(t, req.VerifiedCommand.Command(), seen.Command())
	// Apply can rebuild its typed input from the command alone, and it equals
	// the fingerprinted content.
	decoded, err := jsonProjector(seen.Schema(), seen.Command())
	require.NoError(t, err)
	require.Equal(t, f.owner.lastReq.Content.Fields(), decoded)
}

func TestExecute_VerifiedCommandReplayDoesNotReapply(t *testing.T) {
	f := newR2Fixture(t)
	req := f.verifiedRequest(t, jsonProjector)
	first, err := f.runner.Execute(t.Context(), req, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, first.Status)

	second, err := f.runner.Execute(t.Context(), req, f.owner)

	require.NoError(t, err)
	require.Equal(t, StatusReplayed, second.Status)
	require.Equal(t, 1, f.owner.applyCount())
}

func TestExecute_VerifiedCommandContentMismatchStopsBeforeClaim(t *testing.T) {
	f := newR2Fixture(t)
	req := f.verifiedRequest(t, jsonProjector)
	// A different default Content than the one the command was verified against.
	other := r483Content("8", "13000")
	otherContent := mustContent(t, other)
	req.Content = otherContent

	out, err := f.runner.Execute(t.Context(), req, f.owner)

	require.Error(t, err)
	require.Equal(t, Outcome{}, out)
	require.Equal(t, 0, f.store.bindingCount(), "no binding row may exist")
	require.Equal(t, 0, f.owner.applyCount())
	require.Equal(t, 0, f.owner.gateCount())
	require.False(t, f.store.hasResultOfScope(f.scope, f.key))
}

func TestExecute_ProjectorChangeBeforeClaimLeavesNoBindingOrAttempt(t *testing.T) {
	f := newR2Fixture(t)
	var calls atomic.Int32
	// Call 1 is the constructor; call 2 is the runner's pre-claim check.
	req := f.verifiedRequest(t, flipAfter(1, &calls))

	out, err := f.runner.Execute(t.Context(), req, f.owner)

	require.Error(t, err)
	require.Equal(t, Outcome{}, out)
	require.Equal(t, 0, f.store.bindingCount())
	require.Equal(t, 0, f.owner.applyCount())
}

func TestExecute_CommandTamperedBetweenClaimAndApplyStopsBeforeApply(t *testing.T) {
	f := newR2Fixture(t)
	var calls atomic.Int32
	// Calls 1 and 2 (constructor, pre-claim) pass; call 3 (just before Apply)
	// observes a changed projection.
	req := f.verifiedRequest(t, flipAfter(2, &calls))

	out, err := f.runner.Execute(t.Context(), req, f.owner)

	require.Error(t, err)
	require.Equal(t, Outcome{}, out)
	require.Equal(t, int32(3), calls.Load())
	require.Equal(t, 0, f.owner.applyCount(), "Apply must never run")
	require.Equal(t, 1, f.store.bindingCount(), "the claim already happened")
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	require.False(t, f.store.hasResult(stored.Binding.ID), "no success receipt")
	require.Equal(t, "1", f.owner.version("billing_version"), "no owner write")
	attempt, ok := f.store.latestAttempt(stored.Binding.ID)
	require.True(t, ok)
	require.Equal(t, OutcomeRetryableFailure, attempt.Outcome)
}
