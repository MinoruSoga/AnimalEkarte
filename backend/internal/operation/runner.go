package operation

import (
	"context"
	"errors"
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
	// It must run outside an ambient transaction.
	ClaimBinding(ctx context.Context, scope Scope, key uuid.UUID, sum []byte, keyVersion string) (storedBinding, error)
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
	if r.fp == nil {
		return Outcome{}, apperrors.WrapInternalServerError("operation fingerprinter is required")
	}

	sum, keyVersion := r.fp.Compute(req.Scope, req.Content)

	// Claim transaction: the binding must exist before any business write and
	// stays bound to this content even when the attempt is rejected.
	stored, err := r.store.ClaimBinding(ctx, req.Scope, req.Key, sum, keyVersion)
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
		// attempt: the binding belongs to the earlier content.
		return Outcome{Status: StatusKeyConflict, Binding: &binding}, nil
	}

	var outcome Outcome
	txErr := r.tx.WithTx(ctx, func(txCtx context.Context) error {
		return r.participate(txCtx, binding, req.Preconditions, p, &outcome)
	})
	if txErr == nil {
		return outcome, nil
	}
	return r.resolveFailedAttempt(ctx, binding, p, txErr)
}

// Lookup returns the stored state of one binding without performing the owner
// write. The current grant is re-checked before any stored result is disclosed.
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
		return Outcome{}, err
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
			status = StatusInProgress
			return nil
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
	if err != nil {
		return "", nil, err
	}
	return status, res, nil
}

// disclose applies the current grant to an outcome. A denied grant hides the
// status detail and the result body behind not-authorized.
func disclose(ctx context.Context, auth LookupAuthorizer, binding Binding, status Status, res *Result) Outcome {
	if err := auth.AuthorizeLookup(ctx, binding, res); err != nil {
		return Outcome{Status: StatusNotAuthorized, Binding: &binding}
	}
	return Outcome{Status: status, Binding: &binding, Result: res}
}

// participate is the outermost business transaction body: owner gate, binding
// row lock, receipt check, owner write and receipt insert.
func (r *Runner) participate(ctx context.Context, binding Binding, pre Preconditions, p Participant, outcome *Outcome) error {
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
		// An unresolved effect must not be executed again.
		*outcome = Outcome{Status: StatusResultUnavailable, Binding: &binding}
		return nil
	}

	applied, err := p.Apply(ctx, binding, pre)
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
		return Outcome{Status: StatusResultUnavailable, Binding: &binding}, nil //nolint:nilerr // 意図的: 成否を再読取できないときは結果不明として停止し、再実行させない
	}
	if committed != nil {
		return disclose(ctx, p, binding, StatusSucceeded, committed), nil
	}
	if errors.Is(txErr, errBindingBusy) {
		// Another attempt is inside the business transaction. Nothing is
		// recorded: the winner owns the outcome row.
		return Outcome{Status: StatusInProgress, Binding: &binding}, nil
	}
	unknown, err := r.store.HasUnknownOutcome(ctx, binding)
	if err != nil {
		return Outcome{Status: StatusResultUnavailable, Binding: &binding}, nil //nolint:nilerr // 意図的: unknown 有無を確認できないときは記録せず結果不明として停止する
	}
	if unknown {
		// Unknown is sticky: no failure row is recorded over it.
		return Outcome{Status: StatusResultUnavailable, Binding: &binding}, nil
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
		StatusEditRightLost, StatusNotAuthorized, StatusScopeNotFound:
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
	case ActorStaff, ActorCustomer, ActorSystem:
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
