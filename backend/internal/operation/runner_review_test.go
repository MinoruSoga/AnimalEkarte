package operation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

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
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("database result uncertain")})
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
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("database result uncertain")})
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
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("database result uncertain")})
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

// Rejected scopes never reach binding creation or owner code.
type unsupportedEffects struct{ *fakeBillingParticipant }

func (p unsupportedEffects) Effects() EffectScope { return "external-send" }

func TestExecute_RejectsNontransactionalEffectsBeforeClaim(t *testing.T) {
	f := newR2Fixture(t)
	_, err := f.runner.Execute(t.Context(), Request{Scope: f.scope, Key: f.key, Content: f.content}, unsupportedEffects{f.owner})
	require.Error(t, err)
	require.Zero(t, f.store.bindingCount())
	require.Zero(t, f.owner.applyCount())
}

type failedReadStore struct {
	store
	resultErr, unknownErr error
}

func (s failedReadStore) LoadResult(ctx context.Context, b Binding) (*Result, error) {
	if s.resultErr != nil {
		return nil, s.resultErr
	}
	return s.store.LoadResult(ctx, b)
}
func (s failedReadStore) HasUnknownOutcome(ctx context.Context, b Binding) (bool, error) {
	if s.unknownErr != nil {
		return false, s.unknownErr
	}
	return s.store.HasUnknownOutcome(ctx, b)
}
func TestFailedAttempt_AllStoredStatesRecheckGrant(t *testing.T) {
	for _, name := range []string{"busy", "sticky-unknown", "result-read-error", "unknown-read-error"} {
		t.Run(name, func(t *testing.T) {
			f := newR2Fixture(t)
			f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("uncertain database result")})
			f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"})
			b, _ := f.store.bindingFor(f.scope, f.key)
			f.owner.revoke()
			txErr := errors.New("gate failed")
			switch name {
			case "busy":
				txErr = &Busy{}
			case "result-read-error":
				f.runner.store = failedReadStore{store: f.store, resultErr: errors.New("unavailable")}
			case "unknown-read-error":
				f.runner.store = failedReadStore{store: f.store, unknownErr: errors.New("unavailable")}
			}
			got, err := f.runner.resolveFailedAttempt(t.Context(), b.Binding, f.owner, txErr)
			require.NoError(t, err)
			require.Equal(t, StatusNotAuthorized, got.Status)
			require.Nil(t, got.Result)
		})
	}
}

func TestExecute_UnknownCommitFailureRetainsCause(t *testing.T) {
	f := newR2Fixture(t)
	f.owner.failApplyWith(&IndeterminateEffect{Cause: errors.New("database result uncertain")})
	f.runner.tx = &fakeTx{store: f.store, failCommit: errors.New("commit failed")}
	got, err := f.runner.Execute(t.Context(), Request{Scope: f.scope, Key: f.key, Content: f.content, Preconditions: Preconditions{"billing_version": "1", "line_version": "1"}}, f.owner)
	require.NoError(t, err)
	require.Equal(t, StatusResultUnavailable, got.Status)
	b, _ := f.store.bindingFor(f.scope, f.key)
	a, ok := f.store.latestAttempt(b.Binding.ID)
	require.True(t, ok)
	require.Equal(t, OutcomeUnknown, a.Outcome)
	require.Equal(t, "1", f.owner.version("billing_version"))
}

// resolvedLookup models a domain's persisted binding -> resource index. The
// grant permits B; the stored operation targets A. Current input cannot replace A.
type resolvedLookup struct {
	persisted  map[uint64]LookupTarget
	allowed    string
	resolveErr error
	authorized []LookupTarget
}

func (a *resolvedLookup) ResolveLookupTarget(_ context.Context, b Binding, res *Result) (LookupTarget, error) {
	if a.resolveErr != nil {
		return LookupTarget{}, a.resolveErr
	}
	if target, ok := a.persisted[b.ID]; ok {
		return target, nil
	}
	// Success result_refs belong to the helper-loaded receipt for this binding;
	// the fixture's owner resolver recognizes billing_id as its original root.
	if res != nil && res.ResultRefs["billing_id"] != "" {
		return LookupTarget{BindingID: b.ID, ClinicID: b.ClinicID, Kind: "billing", Ref: res.ResultRefs["billing_id"]}, nil
	}
	return LookupTarget{}, errors.New("no original target")
}
func (a *resolvedLookup) AuthorizeLookup(_ context.Context, target LookupTarget) error {
	a.authorized = append(a.authorized, target)
	if target.Ref != a.allowed {
		return errors.New("original resource denied")
	}
	return nil
}

type targetParticipant struct {
	*fakeBillingParticipant
	*resolvedLookup
}

func (p targetParticipant) ResolveLookupTarget(ctx context.Context, b Binding, res *Result) (LookupTarget, error) {
	return p.resolvedLookup.ResolveLookupTarget(ctx, b, res)
}
func (p targetParticipant) AuthorizeLookup(ctx context.Context, target LookupTarget) error {
	return p.resolvedLookup.AuthorizeLookup(ctx, target)
}

func TestExecute_KeyConflictAuthorizesPersistedOriginalNotCurrentTarget(t *testing.T) {
	f := newR2Fixture(t)
	// A rejected operation still binds the key but has no success receipt.
	f.owner.setVersions("2", "1")
	f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"})
	b, _ := f.store.bindingFor(f.scope, f.key)
	auth := &resolvedLookup{persisted: map[uint64]LookupTarget{b.Binding.ID: {BindingID: b.Binding.ID, ClinicID: b.Binding.ClinicID, Kind: "billing", Ref: "A"}}, allowed: "B"}
	fields := f.content.Fields()
	fields["billing_id"] = "B"
	content, err := NewContent(fields)
	require.NoError(t, err)
	got, err := f.runner.Execute(t.Context(), Request{Scope: f.scope, Key: f.key, Content: content}, targetParticipant{f.owner, auth})
	require.NoError(t, err)
	require.Equal(t, StatusNotAuthorized, got.Status)
	require.Len(t, auth.authorized, 1)
	require.Equal(t, "A", auth.authorized[0].Ref)
	require.Equal(t, 1, f.owner.applyCount())
}
func TestDisclosure_UnresolvedOrWrongBindingTargetFailsClosed(t *testing.T) {
	b := Binding{ID: 10, ClinicID: 1}
	for _, name := range []string{"missing", "resolver-error", "other-clinic", "other-binding", "empty-resource"} {
		t.Run(name, func(t *testing.T) {
			a := &resolvedLookup{allowed: "B", persisted: map[uint64]LookupTarget{}}
			switch name {
			case "resolver-error":
				a.resolveErr = errors.New("unavailable")
			case "other-clinic":
				a.persisted[10] = LookupTarget{BindingID: 10, ClinicID: 2, Kind: "billing", Ref: "B"}
			case "other-binding":
				a.persisted[10] = LookupTarget{BindingID: 11, ClinicID: 1, Kind: "billing", Ref: "B"}
			case "empty-resource":
				a.persisted[10] = LookupTarget{BindingID: 10, ClinicID: 1, Kind: "billing"}
			}
			got := disclose(t.Context(), a, b, StatusKeyConflict, nil)
			require.Equal(t, StatusNotAuthorized, got.Status)
			require.Empty(t, a.authorized)
		})
	}
}
func TestDisclosure_MayResolveOriginalFromBoundSuccessReceipt(t *testing.T) {
	b := Binding{ID: 10, ClinicID: 1}
	a := &resolvedLookup{allowed: "A"}
	got := disclose(t.Context(), a, b, StatusReplayed, &Result{ResultRefs: map[string]string{"billing_id": "A"}})
	require.Equal(t, StatusReplayed, got.Status)
	require.Equal(t, "A", a.authorized[0].Ref)
}

// actorScopeProbe catches any invalid spelling reaching persistence.
type actorScopeProbe struct {
	store
	claims, lookups int
}

func (s *actorScopeProbe) ClaimBinding(context.Context, Scope, uuid.UUID, []byte, string) (storedBinding, error) {
	s.claims++
	return storedBinding{}, errors.New("unexpected claim")
}
func (s *actorScopeProbe) FindBinding(context.Context, Scope, uuid.UUID) (storedBinding, bool, error) {
	s.lookups++
	return storedBinding{}, false, errors.New("unexpected lookup")
}
func TestActorScope_RejectsNonCanonicalIDsBeforeStoreOrOwner(t *testing.T) {
	for _, kind := range []ActorKind{ActorStaff, ActorCustomer} {
		for _, ref := range []string{"012", " 12", "12 ", "+12", "0", "-12", "18446744073709551616", "", "12.0"} {
			t.Run(string(kind)+"/"+ref, func(t *testing.T) {
				f := newR2Fixture(t)
				probe := &actorScopeProbe{store: f.store}
				f.runner.store = probe
				scope := f.scope
				scope.ActorKind, scope.ActorRef = kind, ref
				_, err := f.runner.Execute(t.Context(), Request{Scope: scope, Key: f.key, Content: f.content}, f.owner)
				require.Error(t, err)
				_, err = f.runner.Lookup(t.Context(), scope, f.key, allowLookup{})
				require.Error(t, err)
				require.Zero(t, probe.claims)
				require.Zero(t, probe.lookups)
				require.Zero(t, f.owner.gates)
				require.Zero(t, f.owner.applyCount())
			})
		}
	}
}
func TestActorScope_CanonicalIDReplaysOnceAndSystemNamesRemainValid(t *testing.T) {
	for _, kind := range []ActorKind{ActorStaff, ActorCustomer, ActorSystem} {
		t.Run(string(kind), func(t *testing.T) {
			f := newR2Fixture(t)
			f.scope.ActorKind = kind
			if kind == ActorSystem {
				f.scope.ActorRef = "scheduled-billing"
			}
			require.Equal(t, StatusSucceeded, f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"}).Status)
			require.Equal(t, StatusReplayed, f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"}).Status)
			got, err := f.runner.Lookup(t.Context(), f.scope, f.key, allowLookup{})
			require.NoError(t, err)
			require.Equal(t, StatusSucceeded, got.Status)
			require.Equal(t, 1, f.owner.applyCount())
		})
	}
}

// An owner resolver may inspect a receipt, but cannot rewrite what is returned.
type mutatingReceiptLookup struct {
	allowLookup
	received *Result
}

func (a *mutatingReceiptLookup) ResolveLookupTarget(ctx context.Context, b Binding, res *Result) (LookupTarget, error) {
	a.received = res
	if res != nil {
		for _, fields := range []map[string]string{res.ResultRefs, res.CommittedVersions, res.EffectSummary} {
			for key := range fields {
				fields[key] = "changed"
				delete(fields, key)
			}
			if fields != nil {
				fields["injected"] = "changed"
			}
		}
		res.RecordedAt = res.RecordedAt.Add(time.Hour)
	}
	return a.allowLookup.ResolveLookupTarget(ctx, b, res)
}

type mutatingReceiptParticipant struct {
	*fakeBillingParticipant
	auth *mutatingReceiptLookup
}

func (p mutatingReceiptParticipant) ResolveLookupTarget(ctx context.Context, b Binding, res *Result) (LookupTarget, error) {
	return p.auth.ResolveLookupTarget(ctx, b, res)
}
func (p mutatingReceiptParticipant) AuthorizeLookup(ctx context.Context, target LookupTarget) error {
	return p.auth.AuthorizeLookup(ctx, target)
}

// lostCommitResponseTx commits the fake transaction, then loses its reply.
// It models only the helper recovery branch, not a real PostgreSQL failure.
type lostCommitResponseTx struct{ *fakeTx }

func (tx lostCommitResponseTx) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if err := tx.fakeTx.WithTx(ctx, fn); err != nil {
		return err
	}
	return errors.New("commit reply lost after commit")
}
func TestReceiptResolver_CannotMutateLookupReplayOrCommitRecovery(t *testing.T) {
	for _, path := range []string{"lookup", "replay", "commit-recovery"} {
		t.Run(path, func(t *testing.T) {
			f := newR2Fixture(t)
			first := f.execute(f.content, Preconditions{"billing_version": "1", "line_version": "1"})
			expected := cloneResult(*first.Result)
			expected.EffectSummary = map[string]string{"effect": "saved"}
			f.store.commitResultForTest(first.Binding.ID, expected)
			auth := &mutatingReceiptLookup{}
			participant := mutatingReceiptParticipant{f.owner, auth}
			var got Outcome
			var err error
			if path == "lookup" {
				got, err = f.runner.Lookup(t.Context(), f.scope, f.key, auth)
			} else {
				if path == "commit-recovery" {
					f.runner.tx = lostCommitResponseTx{&fakeTx{store: f.store}}
				}
				got, err = f.runner.Execute(t.Context(), Request{Scope: f.scope, Key: f.key, Content: f.content}, participant)
			}
			require.NoError(t, err)
			require.NotNil(t, auth.received)
			require.NotSame(t, auth.received, got.Result)
			require.Equal(t, expected, *got.Result)
			stored, err := f.store.LoadResult(t.Context(), *first.Binding)
			require.NoError(t, err)
			require.Equal(t, expected, *stored)
			require.Equal(t, 1, f.owner.applyCount())
		})
	}
}
func TestCopyLookupReceipt_PreservesNilMapsAndTimestamp(t *testing.T) {
	require.Nil(t, copyLookupReceipt(nil))
	original := &Result{RecordedAt: fakeRecordedAt}
	copied := copyLookupReceipt(original)
	require.NotSame(t, original, copied)
	require.Equal(t, original, copied)
	require.Nil(t, copied.ResultRefs)
	require.Nil(t, copied.CommittedVersions)
	require.Nil(t, copied.EffectSummary)
}
