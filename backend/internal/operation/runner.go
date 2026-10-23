package operation

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// store is the persistence surface the runner needs. It is declared by the
// consumer so tests can run the runner without a database.
type store interface {
	// ClaimBinding inserts the binding when absent and returns the stored row.
	// It runs inside its own short transaction; when this attempt wins the
	// insert, writeMetadata is invoked exactly once inside that claim
	// transaction so claim metadata commits or rolls back with the binding.
	// It must run outside an ambient transaction.
	ClaimBinding(ctx context.Context, scope Scope, key uuid.UUID, sum []byte, keyVersion string, writeMetadata func(ctx context.Context, b Binding) error) (storedBinding, error)
	// FindBinding returns the binding of this exact scope and key.
	FindBinding(ctx context.Context, scope Scope, key uuid.UUID) (storedBinding, bool, error)
	// LockBindingNoWait locks the binding row; a held row is errBindingBusy.
	// It must run inside an ambient transaction.
	LockBindingNoWait(ctx context.Context, b Binding) error
	// LoadResult returns the success receipt, or (nil, nil) when none exists.
	LoadResult(ctx context.Context, b Binding) (*Result, error)
	// InsertResult writes the success receipt. It must run inside the owner
	// transaction that performed the write.
	InsertResult(ctx context.Context, b Binding, res Result) error
	// HasUnknownOutcome reports whether any unknown outcome exists (sticky).
	HasUnknownOutcome(ctx context.Context, b Binding) (bool, error)
	// MarkApplySavepoint creates the savepoint that RecordUnknownUnderLock
	// rolls back to. It must run before Participant.Apply: a savepoint created
	// after the owner writes cannot undo them.
	MarkApplySavepoint(ctx context.Context) error
	// RecordUnknownUnderLock records the unknown attempt inside the caller's
	// business transaction by rolling back to the apply savepoint (undoing the
	// owner writes) and inserting the unknown row. The unknown commits with
	// that transaction while the binding row lock is still held, so the stop
	// state is durable before the lock is released.
	RecordUnknownUnderLock(ctx context.Context, b Binding, attemptKey uuid.UUID, status Status) error
	// RecordOutcome appends one attempt outcome. It must run outside an
	// ambient transaction.
	RecordOutcome(ctx context.Context, b Binding, attemptKey uuid.UUID, outcome AttemptOutcome, status Status) error
}

// storedBinding is a binding row together with its stored fingerprint.
type storedBinding struct {
	Binding     Binding
	Fingerprint []byte
	KeyVersion  string
}

// errBindingBusy reports that another attempt holds the binding row.
var errBindingBusy = errors.New("operation binding is held by another attempt")

// Runner executes and looks up operation receipts for write owners.
type Runner struct {
	store         store
	tx            persistence.Transactor
	fp            *Fingerprinter
	newAttemptKey func() uuid.UUID
}

// NewRunner returns a runner backed by db. The fingerprinter carries the
// injected key ring; nothing about the keys is configured here.
func NewRunner(db *gorm.DB, fp *Fingerprinter) *Runner {
	return newRunner(newGormStore(db), persistence.NewTransactor(db), fp, uuid.New)
}

// newRunner wires a runner to explicit collaborators so tests can substitute
// the store, the transaction boundary and the attempt-key generator.
func newRunner(s store, tx persistence.Transactor, fp *Fingerprinter, newKey func() uuid.UUID) *Runner {
	if newKey == nil {
		newKey = uuid.New
	}
	return &Runner{store: s, tx: tx, fp: fp, newAttemptKey: newKey}
}

// Execute runs one attempt of req for the write owner p.
//
// The binding is claimed first, so a key is never reusable for different
// content. The owner write and its receipt then commit in one transaction. A
// failed transaction is never reported as a failure before the receipt is
// re-read: a committed receipt makes the call succeeded, and a definite owner
// rejection is recorded as rejected instead of retryable_failure.
func (r *Runner) Execute(ctx context.Context, req Request, p Participant) (Outcome, error) {
	if persistence.TxFromContext(ctx) != nil {
		return Outcome{}, apperrors.WrapInternalServerError("operation execution must start outside an ambient transaction")
	}
	if err := validateRequest(req); err != nil {
		return Outcome{}, err
	}
	if p == nil {
		return Outcome{}, apperrors.WrapInternalServerError("operation participant is required")
	}
	if p.Effects() != EffectsDatabaseOnly {
		return Outcome{}, apperrors.WrapInvalidInput("operation supports database-transaction-only effects")
	}
	if r.fp == nil {
		return Outcome{}, apperrors.WrapInternalServerError("operation fingerprinter is required")
	}

	sum, keyVersion := r.fp.Compute(req.Scope, req.Content)

	// Claim transaction: the binding must exist before any business write and
	// stays bound to this content even when the attempt is rejected. A
	// participant that implements ClaimMetadataWriter gets the store's claim
	// callback; the store invokes it once, only on the claim winner.
	var writeMetadata func(context.Context, Binding) error
	if writer, ok := p.(ClaimMetadataWriter); ok {
		writeMetadata = func(ctx context.Context, b Binding) error {
			return writer.CreateClaimMetadata(ctx, b, req)
		}
	}
	stored, err := r.store.ClaimBinding(ctx, req.Scope, req.Key, sum, keyVersion, writeMetadata)
	if err != nil {
		return Outcome{}, err
	}
	binding := stored.Binding

	matches, err := r.fp.Matches(req.Scope, req.Content, stored.Fingerprint, stored.KeyVersion)
	if err != nil {
		return Outcome{}, err
	}
	if !matches {
		// Same key, different content. Never execute and never record an
		// attempt: the binding belongs to the earlier content. The conflict
		// answer itself can leak the stored content (an equality oracle for
		// clinical or financial fields), so the current grant is re-checked
		// before it is exposed: a revoked actor sees not-authorized.
		receipt, loadErr := r.store.LoadResult(ctx, binding)
		if loadErr != nil {
			return disclose(ctx, p, binding, StatusResultUnavailable, nil), nil
		}
		// A success receipt may identify the original target for permission
		// lookup, but conflict never returns that receipt to the caller.
		return discloseUsingReceipt(ctx, p, binding, StatusKeyConflict, receipt, nil), nil
	}

	var outcome Outcome
	var applyFailure error
	callbackSucceeded := false
	txErr := r.tx.WithTx(ctx, func(txCtx context.Context) error {
		callbackErr := r.participate(txCtx, binding, req, p, &outcome, &applyFailure)
		callbackSucceeded = callbackErr == nil
		return callbackErr
	})
	if txErr == nil {
		return outcome, nil
	}
	if callbackSucceeded {
		// COMMIT may still be undecided after its reply is lost. A single
		// lock-free read with no receipt cannot establish rollback.
		status, receipt, settleErr := r.settledState(ctx, binding)
		if settleErr != nil {
			return disclose(ctx, p, binding, StatusResultUnavailable, nil), nil
		}
		if status != StatusNotSucceeded {
			if status == StatusSucceeded && outcome.Status == StatusReplayed {
				status = StatusReplayed
			}
			return disclose(ctx, p, binding, status, receipt), nil
		}
		// Only a receipt/unknown re-read under the acquired binding lock
		// establishes no success. Definite callback failures already rolled
		// back and use the existing classification below.
	}
	return r.resolveFailedAttempt(ctx, binding, p, errors.Join(txErr, applyFailure))
}

// Lookup returns the stored state of one binding without performing the owner
// write. The current grant is re-checked before any stored result is disclosed.
// A recorded unknown returns result-unavailable. A result-store re-read
// failure (receipt/unknown read error) is distinct from a recorded unknown:
// the current authorization is evaluated first — a denied grant stays
// non-disclosing (not-authorized), an authorized caller gets an internal
// error. Execute's post-failure recheck keeps mapping read failure to
// result-unavailable (sticky unknown recorded) and is unchanged.
func (r *Runner) Lookup(ctx context.Context, scope Scope, key uuid.UUID, auth LookupAuthorizer) (Outcome, error) {
	if persistence.TxFromContext(ctx) != nil {
		return Outcome{}, apperrors.WrapInternalServerError("operation lookup must start outside an ambient transaction")
	}
	if err := validateScope(scope); err != nil {
		return Outcome{}, err
	}
	if key == uuid.Nil {
		return Outcome{}, apperrors.WrapInvalidInput("operation key is required")
	}
	if auth == nil {
		return Outcome{}, apperrors.WrapInternalServerError("operation lookup authorizer is required")
	}

	stored, found, err := r.store.FindBinding(ctx, scope, key)
	if err != nil {
		return Outcome{}, err
	}
	if !found {
		return Outcome{Status: StatusScopeNotFound}, nil
	}
	binding := stored.Binding

	status, res, err := r.settledState(ctx, binding)
	if err != nil {
		// The result re-read failed. Evaluate the current authorization on the
		// original anchor first: a denied grant wins non-disclosure. An
		// authorized caller gets an internal error instead — the store error
		// text, binding, and any result are never returned.
		outcome := disclose(ctx, auth, binding, StatusResultUnavailable, nil)
		if outcome.Status == StatusNotAuthorized {
			return outcome, nil
		}
		return Outcome{}, apperrors.WrapInternalServerError("operation result lookup failed")
	}
	return disclose(ctx, auth, binding, status, res), nil
}

// settledState determines the binding state. A receipt is immutable, so a
// visible one settles the answer; otherwise the receipt and the unknown state
// are re-read under the binding lock, so a success committed after the first
// read is never reported as not-succeeded.
func (r *Runner) settledState(ctx context.Context, binding Binding) (Status, *Result, error) {
	res, err := r.store.LoadResult(ctx, binding)
	if err != nil {
		return "", nil, err
	}
	if res != nil {
		return StatusSucceeded, res, nil
	}

	status := StatusNotSucceeded
	err = r.tx.WithTx(ctx, func(txCtx context.Context) error {
		lockErr := r.store.LockBindingNoWait(txCtx, binding)
		if errors.Is(lockErr, errBindingBusy) {
			// 55P03 has aborted the transaction block: the only way out is a
			// rollback, so hand the busy error to the transactor instead of
			// letting it commit an aborted transaction.
			return errBindingBusy
		}
		if lockErr != nil {
			return lockErr
		}
		locked, loadErr := r.store.LoadResult(txCtx, binding)
		if loadErr != nil {
			return loadErr
		}
		if locked != nil {
			status, res = StatusSucceeded, locked
			return nil
		}
		unknown, unknownErr := r.store.HasUnknownOutcome(txCtx, binding)
		if unknownErr != nil {
			return unknownErr
		}
		if unknown {
			status = StatusResultUnavailable
		}
		return nil
	})
	if errors.Is(err, errBindingBusy) {
		// The lock was held: roll back happened, nothing is disclosed.
		return StatusInProgress, nil, nil
	}
	if err != nil {
		return "", nil, err
	}
	return status, res, nil
}

// disclose applies the current grant to an outcome. A denied grant hides the
// status detail and the result body behind not-authorized.
func disclose(ctx context.Context, auth LookupAuthorizer, binding Binding, status Status, res *Result) Outcome {
	return discloseUsingReceipt(ctx, auth, binding, status, res, res)
}

// Authorization evidence and the response body are separate: key-conflict may
// resolve the original resource from a receipt without disclosing its body.
func discloseUsingReceipt(ctx context.Context, auth LookupAuthorizer, binding Binding, status Status, receipt, result *Result) Outcome {
	target, err := auth.ResolveLookupTarget(ctx, binding, copyLookupReceipt(receipt))
	if err != nil || target.BindingID != binding.ID || target.ClinicID != binding.ClinicID || strings.TrimSpace(target.Kind) == "" || strings.TrimSpace(target.Ref) == "" {
		return Outcome{Status: StatusNotAuthorized, Binding: &binding}
	}
	if err := auth.AuthorizeLookup(ctx, target); err != nil {
		return Outcome{Status: StatusNotAuthorized, Binding: &binding}
	}
	return Outcome{Status: status, Binding: &binding, Result: result}
}

// copyLookupReceipt gives the owner resolver a detached view. It must not be
// able to change either the stored success receipt or the disclosed result.
func copyLookupReceipt(res *Result) *Result {
	if res == nil {
		return nil
	}
	copyMap := func(src map[string]string) map[string]string {
		if src == nil {
			return nil
		}
		dst := make(map[string]string, len(src))
		for key, value := range src {
			dst[key] = value
		}
		return dst
	}
	return &Result{
		ResultRefs:        copyMap(res.ResultRefs),
		CommittedVersions: copyMap(res.CommittedVersions),
		EffectSummary:     copyMap(res.EffectSummary),
		RecordedAt:        res.RecordedAt,
	}
}

// participate is the outermost business transaction body: owner gate, binding
// row lock, receipt check, owner write and receipt insert.
func (r *Runner) participate(ctx context.Context, binding Binding, req Request, p Participant, outcome *Outcome, applyFailure *error) error {
	if err := p.Gate(ctx); err != nil {
		return err
	}
	if err := r.store.LockBindingNoWait(ctx, binding); err != nil {
		return err
	}

	existing, err := r.store.LoadResult(ctx, binding)
	if err != nil {
		return err
	}
	if existing != nil {
		*outcome = disclose(ctx, p, binding, StatusReplayed, existing)
		return nil
	}

	unknown, err := r.store.HasUnknownOutcome(ctx, binding)
	if err != nil {
		return err
	}
	if unknown {
		// An unresolved effect must not be executed again, and the stored
		// stop-state is disclosed only to the current grant, like every other
		// stored state.
		*outcome = disclose(ctx, p, binding, StatusResultUnavailable, nil)
		return nil
	}

	// The savepoint must exist before the owner writes: a savepoint created
	// after an indeterminate Apply could not undo them.
	if err := r.store.MarkApplySavepoint(ctx); err != nil {
		return err
	}
	if !req.VerifiedCommand.IsZero() {
		// Re-verify immediately before Apply. The claim already happened, so a
		// mismatch here is a non-deterministic projector or corrupted request
		// memory: a programming error, not a business rejection. Returning it as
		// a plain error rolls back the transaction before Apply and is recorded
		// as retryable_failure by resolveFailedAttempt (never as a rejection that
		// could be disclosed as a business result, never as a success receipt).
		if err := req.VerifiedCommand.verifyAgainst(req.Content); err != nil {
			return apperrors.Wrap(err, "operation verified command changed before apply")
		}
	}
	applied, err := p.Apply(ctx, binding, req)
	var indeterminate *IndeterminateEffect
	if errors.As(err, &indeterminate) {
		*applyFailure = indeterminate // retain this cause if the later commit fails
		// The effect may or may not have happened. Undo the owner writes and
		// commit the unknown row inside this transaction, while the binding
		// row lock is still held: the stop state is durable before the lock is
		// released, so no concurrent execution can run Apply twice.
		attemptKey := r.newAttemptKey()
		if attemptKey == uuid.Nil {
			return errors.Join(indeterminate, apperrors.WrapInternalServerError("operation attempt key is required"))
		}
		if unknownErr := r.store.RecordUnknownUnderLock(ctx, binding, attemptKey, StatusResultUnavailable); unknownErr != nil {
			// Recording failed under the lock and the transaction will roll
			// back. resolveFailedAttempt falls through to the
			// outside-transaction best-effort record; the insert error must
			// not mask the indeterminate cause.
			return errors.Join(indeterminate, unknownErr)
		}
		*outcome = disclose(ctx, p, binding, StatusResultUnavailable, nil)
		return nil
	}
	if err != nil {
		return err
	}
	if err := r.store.InsertResult(ctx, binding, applied); err != nil {
		return err
	}
	// Return the stored receipt (with its database recorded_at) so a later
	// replay of the same key returns exactly what this call returned.
	stored, err := r.store.LoadResult(ctx, binding)
	if err != nil {
		return err
	}
	if stored == nil {
		return apperrors.WrapInternalServerError("operation result was not stored")
	}
	*outcome = Outcome{Status: StatusSucceeded, Binding: &binding, Result: stored}
	return nil
}

// resolveFailedAttempt classifies a rolled-back attempt. The receipt is re-read
// before anything is recorded: a receipt that committed while the client lost
// the response is a success, not a failure.
func (r *Runner) resolveFailedAttempt(ctx context.Context, binding Binding, p Participant, txErr error) (Outcome, error) {
	committed, err := r.store.LoadResult(ctx, binding)
	if err != nil {
		return disclose(ctx, p, binding, StatusResultUnavailable, nil), nil //nolint:nilerr // 意図的: 成否を再読取できないときは結果不明として停止し、再実行させない
	}
	if committed != nil {
		return disclose(ctx, p, binding, StatusSucceeded, committed), nil
	}
	unknown, err := r.store.HasUnknownOutcome(ctx, binding)
	if err != nil {
		return disclose(ctx, p, binding, StatusResultUnavailable, nil), nil //nolint:nilerr // 意図的: unknown 有無を確認できないときは記録せず結果不明として停止する
	}
	if unknown {
		// Unknown is sticky: no failure row is recorded over it.
		return disclose(ctx, p, binding, StatusResultUnavailable, nil), nil
	}

	var busy *Busy
	if errors.Is(txErr, errBindingBusy) || errors.As(txErr, &busy) {
		// Another attempt holds the binding row or a resource behind the owner
		// gate. Nothing is recorded: the winner owns the outcome row.
		return disclose(ctx, p, binding, StatusInProgress, nil), nil
	}

	var indeterminate *IndeterminateEffect
	if errors.As(txErr, &indeterminate) {
		// The in-transaction unknown record failed and the transaction rolled
		// back. Best effort: record unknown outside the transaction. If both
		// recordings fail, no durable stop remains; the current response stops.
		// A later DB-only attempt re-locks the binding and re-reads receipts
		// before Apply. Rollback undid this attempt's database writes; this is
		// not permission to retry an unsupported external effect.
		if err := r.recordOutcome(ctx, binding, OutcomeUnknown, StatusResultUnavailable); err != nil {
			// Preserve the conservative unknown response. The recording error
			// goes to the log; no persistent unknown row is claimed to exist.
			slog.Error("operation unknown outcome recording failed",
				"clinic_id", binding.ClinicID, "operation_key", binding.Key, "err", err)
			return disclose(ctx, p, binding, StatusResultUnavailable, nil), nil //nolint:nilerr // 意図的: unknown記録に失敗しても結果不明として停止し、再実行を誘発しない
		}
		return disclose(ctx, p, binding, StatusResultUnavailable, nil), nil
	}

	attemptOutcome := OutcomeRetryableFailure
	status := Status("")
	var rejection *Rejection
	if errors.As(txErr, &rejection) && isRejectionStatus(rejection.Status) {
		attemptOutcome = OutcomeRejected
		status = rejection.Status
	}
	if err := r.recordOutcome(ctx, binding, attemptOutcome, status); err != nil {
		return Outcome{}, err
	}
	if attemptOutcome == OutcomeRejected {
		return Outcome{Status: status, Binding: &binding}, nil
	}
	return Outcome{}, apperrors.Wrap(txErr, "operation execution failed")
}

// recordOutcome appends one attempt outcome outside any ambient transaction.
func (r *Runner) recordOutcome(ctx context.Context, binding Binding, outcome AttemptOutcome, status Status) error {
	attemptKey := r.newAttemptKey()
	if attemptKey == uuid.Nil {
		return apperrors.WrapInternalServerError("operation attempt key is required")
	}
	return r.store.RecordOutcome(ctx, binding, attemptKey, outcome, status)
}

// isRejectionStatus reports whether s is a definite owner rejection category.
// Any other status on a *Rejection is a programming error and is treated as a
// retryable failure, so it can never be disclosed as a business result.
func isRejectionStatus(s Status) bool {
	switch s {
	case StatusValidation, StatusVersionConflict, StatusDependencyChanged,
		StatusEditRightLost, StatusNotAuthorized:
		return true
	default:
		return false
	}
}

func validateRequest(req Request) error {
	if err := validateScope(req.Scope); err != nil {
		return err
	}
	if req.Key == uuid.Nil {
		return apperrors.WrapInvalidInput("operation key is required")
	}
	if req.Content.fields == nil {
		return apperrors.WrapInvalidInput("operation content must be built with NewContent")
	}
	if !req.VerifiedCommand.IsZero() {
		// Pre-claim re-verification: a command that does not project to the
		// request content stops here, before any binding or attempt row exists.
		if err := req.VerifiedCommand.verifyAgainst(req.Content); err != nil {
			return err
		}
	}
	return nil
}

func validateScope(scope Scope) error {
	if scope.ClinicID == 0 {
		return apperrors.WrapInvalidInput("operation clinic is required")
	}
	if strings.TrimSpace(scope.OwnerKind) == "" {
		return apperrors.WrapInvalidInput("operation owner is required")
	}
	switch scope.ActorKind {
	case ActorStaff, ActorCustomer:
		id, err := strconv.ParseUint(scope.ActorRef, 10, 64)
		if err != nil || id == 0 || strconv.FormatUint(id, 10) != scope.ActorRef {
			return apperrors.WrapInvalidInput("operation actor must be a canonical positive decimal ID")
		}
	case ActorSystem:
		// Named system principals are not numeric staff/customer identities.
	default:
		return apperrors.WrapInvalidInput("operation actor kind is not supported")
	}
	if strings.TrimSpace(scope.ActorRef) == "" {
		return apperrors.WrapInvalidInput("operation actor is required")
	}
	if strings.TrimSpace(scope.Intent) == "" {
		return apperrors.WrapInvalidInput("operation intent is required")
	}
	return nil
}
