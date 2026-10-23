package model

// Persisted identifier constants for the EMR-446/466 LINE acceptance family
// (contract §3.2, docs/work/emr/
// emr-clinic-line-time-selected-contract-20261022.md). Values must match the
// contract strings exactly; other sessions must not re-declare them. The
// operation registry, lookup allowlists, and ContextGate wiring belong to the
// 489 owner and are intentionally not defined here.

// LineTimeOwnerKindReservation is the operation_bindings.owner_kind for the
// LINE acceptance and line start-policy families (contract §3.2, §4.2).
const LineTimeOwnerKindReservation = "reservation"

// LINE acceptance intents (contract §3.2).
const (
	IntentLineAcceptanceInitialize = "line-acceptance.initialize"
	IntentLineAcceptanceConfirm    = "line-acceptance.confirm"
	IntentLineAcceptanceUpdate     = "line-acceptance.update"
)

// LineAcceptanceSchemaV1 is the wire schema marker for the LINE acceptance
// settings family (contract §3.2).
const LineAcceptanceSchemaV1 = "line-acceptance/v1"
