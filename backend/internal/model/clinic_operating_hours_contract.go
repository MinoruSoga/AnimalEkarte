package model

// Persisted identifier constants for the EMR-446/466 common clinic
// operating-hours family (contract §2.2, docs/work/emr/
// emr-clinic-line-time-selected-contract-20261022.md). Values must match the
// contract strings exactly; other sessions must not re-declare them. The
// operation registry, lookup allowlists, and ContextGate wiring belong to the
// 489 owner and are intentionally not defined here.

// LineTimeOwnerKindClinic is the operation_bindings.owner_kind for the common
// clinic operating-hours family (contract §2.2).
const LineTimeOwnerKindClinic = "clinic"

// Common clinic operating-hours intents (contract §2.2).
const (
	IntentClinicOperatingHoursInitialize              = "clinic.operating-hours.initialize"
	IntentClinicOperatingHoursUpdate                  = "clinic.operating-hours.update"
	IntentClinicOperatingHoursHolidaySetFromShifts    = "clinic.operating-hours.holiday-set-from-shifts"
	IntentClinicOperatingHoursHolidayUnsetFromShifts  = "clinic.operating-hours.holiday-unset-from-shifts"
	IntentClinicOperatingHoursHolidaySetFromClosing   = "clinic.operating-hours.holiday-set-from-closing"
	IntentClinicOperatingHoursHolidayUnsetFromClosing = "clinic.operating-hours.holiday-unset-from-closing"
)
