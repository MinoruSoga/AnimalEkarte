package operation

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// metadataParticipant is a ClaimMetadataWriter test double wrapping the
// existing fake billing participant. It records each callback invocation.
type metadataParticipant struct {
	*fakeBillingParticipant
	calls   int
	lastB   Binding
	lastReq Request
	err     error
}

func (p *metadataParticipant) CreateClaimMetadata(_ context.Context, b Binding, req Request) error {
	p.calls++
	p.lastB = b
	p.lastReq = req
	return p.err
}

var _ ClaimMetadataWriter = (*metadataParticipant)(nil)

func (f *r2Fixture) request(content Content) Request {
	return Request{Scope: f.scope, Key: f.key, Content: content}
}

// TestExecute_ClaimMetadata_InvokedOnceOnWinner: a participant implementing
// ClaimMetadataWriter gets its callback exactly once inside the claim
// transaction, before any business write.
func TestExecute_ClaimMetadata_InvokedOnceOnWinner(t *testing.T) {
	f := newR2Fixture(t)
	p := &metadataParticipant{fakeBillingParticipant: f.owner}

	outcome, err := f.runner.Execute(t.Context(), f.request(f.content), p)
	require.NoError(t, err)
	assert.Equal(t, StatusSucceeded, outcome.Status)
	assert.Equal(t, 1, p.calls)
	stored, ok := f.store.bindingFor(f.scope, f.key)
	require.True(t, ok)
	assert.Equal(t, stored.Binding.ID, p.lastB.ID)
	assert.Equal(t, stored.Binding.ClinicID, p.lastB.ClinicID)
	assert.Equal(t, f.key, p.lastB.Key)
	// The callback sees the same fingerprinted request the Apply receives.
	assert.Equal(t, f.content.Fields(), p.lastReq.Content.Fields())
}

// TestExecute_ClaimMetadataErrorRollsBackBinding: a callback failure rolls the
// just-claimed binding back with the metadata — no binding row remains and no
// attempt outcome is recorded.
func TestExecute_ClaimMetadataErrorRollsBackBinding(t *testing.T) {
	f := newR2Fixture(t)
	p := &metadataParticipant{
		fakeBillingParticipant: f.owner,
		err:                    errors.New("metadata insert failed"),
	}

	_, err := f.runner.Execute(t.Context(), f.request(f.content), p)
	require.Error(t, err)
	_, found := f.store.bindingFor(f.scope, f.key)
	assert.False(t, found, "binding must roll back with the failed metadata callback")
	assert.Equal(t, 1, p.calls)
	assert.Equal(t, 0, f.owner.applyCount(), "Apply must not run after a failed claim")
}

// TestExecute_ClaimMetadata_NotCalledForLoser: a same-scope/key claim that
// loses to the original binding never invokes the callback.
func TestExecute_ClaimMetadata_NotCalledForLoser(t *testing.T) {
	f := newR2Fixture(t)

	first := &metadataParticipant{fakeBillingParticipant: f.owner}
	outcome, err := f.runner.Execute(t.Context(), f.request(f.content), first)
	require.NoError(t, err)
	assert.Equal(t, StatusSucceeded, outcome.Status)
	require.Equal(t, 1, first.calls)

	// Same key and same content: replay. The binding already exists so the
	// claim loses and the callback must not run.
	second := &metadataParticipant{fakeBillingParticipant: f.owner}
	replayed, err := f.runner.Execute(t.Context(), f.request(f.content), second)
	require.NoError(t, err)
	assert.Equal(t, StatusReplayed, replayed.Status)
	assert.Equal(t, 0, second.calls, "loser never invokes the metadata callback")
}

// TestExecute_ParticipantWithoutMetadataWriter_Unchanged: a participant that
// does not implement ClaimMetadataWriter behaves exactly as before (the
// runner passes no callback and the fake store records no invocation).
func TestExecute_ParticipantWithoutMetadataWriter_Unchanged(t *testing.T) {
	f := newR2Fixture(t)

	outcome, err := f.runner.Execute(t.Context(), f.request(f.content), f.owner)
	require.NoError(t, err)
	assert.Equal(t, StatusSucceeded, outcome.Status)
	assert.Equal(t, 0, f.store.claimMetadataCalls)
	assert.Equal(t, 1, f.owner.applyCount())
}

// TestLookup_ClaimMetadataNotInvoked: Lookup never claims and never invokes
// the metadata writer.
func TestLookup_ClaimMetadataNotInvoked(t *testing.T) {
	f := newR2Fixture(t)
	p := &metadataParticipant{fakeBillingParticipant: f.owner}

	outcome, err := f.runner.Execute(t.Context(), f.request(f.content), p)
	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, outcome.Status)
	before := p.calls

	looked, err := f.runner.Lookup(t.Context(), f.scope, f.key, p)
	require.NoError(t, err)
	assert.Equal(t, StatusSucceeded, looked.Status)
	assert.Equal(t, before, p.calls, "Lookup must not invoke the claim metadata callback")
}

// TestExecute_ClaimMetadata_UnknownKeyIsNewBinding: a new operation key claims
// a new binding and invokes the callback for that binding.
func TestExecute_ClaimMetadata_UnknownKeyIsNewBinding(t *testing.T) {
	f := newR2Fixture(t)
	p := &metadataParticipant{fakeBillingParticipant: f.owner}

	_, err := f.runner.Execute(t.Context(), f.request(f.content), p)
	require.NoError(t, err)

	other := f.request(f.content)
	other.Key = uuid.New()
	_, err = f.runner.Execute(t.Context(), other, p)
	require.NoError(t, err)
	assert.Equal(t, 2, p.calls)
	assert.Equal(t, other.Key, p.lastB.Key)
}
