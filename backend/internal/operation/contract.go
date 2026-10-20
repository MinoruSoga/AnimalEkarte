// Package operation is the C06 common operation receipt capability: a durable,
// non-reusable binding of an operation key to its scope and business content,
// a success receipt written in the same transaction as the owner write, and
// failed/unknown attempts recorded separately.
//
// The package owns no domain rule. Each write owner supplies a Participant that
// re-verifies its own preconditions, authorization and locks, and performs the
// owner write. Storage is operation_bindings / operation_results /
// operation_attempt_outcomes (backend/migrations/001_init.sql; success is the
// existence of an operation_results row, there is no state column).
//
// Business content and execution preconditions are distinct types on purpose:
// Content is fingerprinted and immutable for one key, Preconditions (CAS
// versions, current fences) are never fingerprinted and are re-verified on every
// attempt, so a human may retry the same content with the same key after a
// stale-precondition rejection.
package operation

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ActorKind mirrors the operation_bindings.actor_kind CHECK set.
type ActorKind string

const (
	ActorStaff    ActorKind = "staff"
	ActorCustomer ActorKind = "customer"
	ActorSystem   ActorKind = "system"
)

// Scope is the permanent binding scope. It must be built from the authenticated
// server context only; a client-supplied clinic or actor never selects a row.
type Scope struct {
	ClinicID  uint64
	OwnerKind string // write owner, e.g. "billing"
	ActorKind ActorKind
	ActorRef  string // staff/customer: canonical positive uint64 decimal; system: nonblank named principal.
	Intent    string // owner-defined operation family, e.g. "billing.price_acknowledgement"
}

// Content is the immutable business content of one intent (target, acknowledged
// versions, amounts...). It is the only fingerprint input besides OwnerKind and
// Intent. Values are canonical strings (decimal integers, not floats). Build it
// with NewContent.
type Content struct {
	fields map[string]string
}

// Fields returns a detached copy of the canonical content values. Owner
// participants derive their write from this map — the exact fingerprinted
// values — and never from a separately captured command. Mutating the returned
// map does not change the content or its fingerprint.
func (c Content) Fields() map[string]string {
	out := make(map[string]string, len(c.fields))
	for k, v := range c.fields {
		out[k] = v
	}
	return out
}

// Preconditions are execution prerequisites (CAS versions, edit-right fences).
// They are never fingerprinted. The Participant re-verifies them under lock.
type Preconditions map[string]string

// Request is one execution attempt of an intent.
type Request struct {
	Scope         Scope
	Key           uuid.UUID
	Content       Content
	Preconditions Preconditions
}

// Binding identifies a stored operation_bindings row.
type Binding struct {
	ID       uint64
	ClinicID uint64
	Key      uuid.UUID
}

// Result is the stored success receipt (operation_results). Maps are stored as
// JSON objects of strings; versions are canonical decimal strings.
type Result struct {
	ResultRefs        map[string]string
	CommittedVersions map[string]string
	EffectSummary     map[string]string
	RecordedAt        time.Time
}

// Status is the result category returned to the caller. Values are candidate
// wire codes from the C06 result category table.
type Status string

const (
	// StatusSucceeded: the owner write and its receipt committed by this call.
	StatusSucceeded Status = "succeeded"
	// StatusReplayed: an earlier success for the same key and content; the stored
	// result is returned and the owner write is not executed again.
	StatusReplayed Status = "replayed"
	// StatusNotSucceeded: binding exists without success; the same key and same
	// content may be retried after the human re-checks current preconditions.
	StatusNotSucceeded Status = "not-succeeded"
	// StatusKeyConflict: the key is bound to different content. Never re-executes.
	StatusKeyConflict Status = "operation-key-conflict"
	// StatusInProgress: another attempt holds the binding; query again with the
	// same key, never create with a new key.
	StatusInProgress Status = "operation-in-progress"
	// StatusResultUnavailable: success/failure cannot be determined. Stop; no
	// automatic compensation, re-create or external resend.
	StatusResultUnavailable Status = "result-unavailable"
	// StatusScopeNotFound: no binding for this scope and key.
	StatusScopeNotFound Status = "scope-not-found"
	// StatusNotAuthorized: the current grant does not allow the operation or its
	// result. No result body is returned.
	StatusNotAuthorized Status = "not-authorized"

	// Owner rejection categories (returned by a Participant via *Rejection).
	StatusValidation        Status = "validation"
	StatusVersionConflict   Status = "version-conflict"
	StatusDependencyChanged Status = "dependency-changed"
	StatusEditRightLost     Status = "edit-right-lost"
)

// AttemptOutcome mirrors the operation_attempt_outcomes.outcome CHECK set.
type AttemptOutcome string

const (
	// OutcomeRejected: a definite owner rejection; no business effect.
	OutcomeRejected AttemptOutcome = "rejected"
	// OutcomeRetryableFailure: the transaction rolled back for a non-business reason.
	OutcomeRetryableFailure AttemptOutcome = "retryable_failure"
	// OutcomeUnknown: the owner cannot determine its transactional database effect.
	// Executions for the binding stop until resolved.
	OutcomeUnknown AttemptOutcome = "unknown"
)

// Rejection is returned by Participant.Apply for a definite business rejection.
// The transaction rolls back and no success receipt exists.
type Rejection struct {
	Status Status
}

func (r *Rejection) Error() string { return "operation rejected: " + string(r.Status) }

// IndeterminateEffect conservatively stops an uncertain database attempt. It
// does not make nontransactional effects safe: external sends, filesystem writes
// and other effects outside the owner transaction are unsupported. Such work
// needs a separately owned transactional outbox and delivery/reconciliation.
type IndeterminateEffect struct {
	Cause error
}

func (e *IndeterminateEffect) Error() string { return "operation effect is indeterminate" }

func (e *IndeterminateEffect) Unwrap() error { return e.Cause }

// Busy reports that a resource the attempt must hold is locked by another
// attempt right now: the operation binding row, or a row behind the owner
// gate. Gate or Apply may return it (wrapping the driver error as needed) so
// the runner classifies the attempt as StatusInProgress and records nothing:
// the client queries the same key again later. A new key never rescues a busy
// binding.
type Busy struct {
	Cause error
}

func (e *Busy) Error() string { return "operation resource is held by another attempt" }

func (e *Busy) Unwrap() error { return e.Cause }

// EffectScope declares which effects an owner performs. Only database writes
// within the supplied outer transaction (including an outbox INSERT) are supported.
type EffectScope string

const EffectsDatabaseOnly EffectScope = "database-transaction-only"

// Participant is implemented by each write owner (consumer-side adapter).
// AuthorizeLookup is also applied before a stored result is replayed, so a
// resend never discloses a result the current grant may no longer read.
type Participant interface {
	LookupAuthorizer
	// Effects must explicitly select EffectsDatabaseOnly. A zero/unknown scope
	// is rejected before claiming a binding or calling Gate/Apply.
	Effects() EffectScope
	// Gate acquires the owner's shared gate as the first step of the outermost
	// business transaction (e.g. a reservation-graph participant). Owners
	// without a shared gate return nil. The helper never takes a gate itself.
	Gate(ctx context.Context) error
	// Apply runs inside the outermost business transaction after the binding
	// row lock. It re-verifies current authorization, scope and Preconditions
	// from req, performs the owner write derived from req.Content — the exact
	// value that was fingerprinted for this binding — and returns the receipt.
	// Taking the Request here is what keeps the fingerprinted content and the
	// written content from diverging: a receipt can never be stored under a
	// payload other than the one bound to the key.
	Apply(ctx context.Context, b Binding, req Request) (Result, error)
}

// LookupTarget is the original resource resolved from persisted owner metadata
// keyed by Binding.ID + ClinicID. It is never built from the current Request.
// A participant with more than one resource must resolve its authoritative
// aggregate/authorization root; this helper does not invent that mapping.
type LookupTarget struct {
	BindingID uint64
	ClinicID  uint64
	Kind      string
	Ref       string
}

// LookupAuthorizer resolves the stored operation's original target before
// re-checking the current grant. ResolveLookupTarget must use persisted owner
// metadata indexed by binding (including bindings without a success receipt).
// res is a detached copy of the receipt read by the helper for this binding;
// resolver mutations are discarded. Its result_refs may be
// used only if they authoritatively identify that original resource, with the
// same clinic and domain identity verified. Never use captured current-request
// content/target as a substitute, including on key-conflict. If the original
// target is absent, deleted without an authorization anchor, ambiguous or
// otherwise unresolvable, return an error: disclosure is denied. A consumer
// without this resolver remains blocked before connecting result disclosure.
type LookupAuthorizer interface {
	ResolveLookupTarget(ctx context.Context, b Binding, res *Result) (LookupTarget, error)
	AuthorizeLookup(ctx context.Context, target LookupTarget) error
}

// Outcome is what Execute and Lookup return. Result is set only for
// StatusSucceeded and StatusReplayed.
type Outcome struct {
	Status  Status
	Binding *Binding
	Result  *Result
}
