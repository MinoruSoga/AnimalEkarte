package model

// Persisted identifier constants for the EMR-446/466 line staff start-policy
// family and the EMR-492 reservation-master operation anchor (contract §4.2,
// §4.3, docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md).
// Values must match the contract strings exactly; other sessions must not
// re-declare them. The operation registry, lookup allowlists, and ContextGate
// wiring belong to the 489 owner and are intentionally not defined here.

// Line staff start-policy intents (contract §4.2).
const (
	IntentLineStartPolicyCreate      = "line-start-policy.create"
	IntentLineStartPolicyReplace     = "line-start-policy.replace"
	IntentLineStartPolicyReconfigure = "line-start-policy.reconfigure"
	IntentLineStartPolicyDelete      = "line-start-policy.delete"
)

// reservation_master_operation_anchors.target_kind values
// (contract §4.3 and the 492 selected wire contract §3.1/§6).
const (
	AnchorTargetKindType            = "type"
	AnchorTargetKindGroup           = "group"
	AnchorTargetKindLineStartPolicy = "line_start_policy"
)
