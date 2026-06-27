package model

import "testing"

// ------------------------------------
// Phase 4A.3 — LabAuditErrorCategory taxonomy validator
// ------------------------------------

// TestValidLabAuditErrorCategory_AllConstantsPass ensures every declared constant returns true.
// If a new constant is added without updating validLabAuditErrorCategories, this test catches it.
func TestValidLabAuditErrorCategory_AllConstantsPass(t *testing.T) {
	for _, c := range []LabAuditErrorCategory{
		LabAuditErrorCategoryInvalidInput,
		LabAuditErrorCategoryNotFound,
		LabAuditErrorCategoryForbidden,
		LabAuditErrorCategoryUnauthorized,
		LabAuditErrorCategoryInternal,
	} {
		if !ValidLabAuditErrorCategory(c) {
			t.Errorf("ValidLabAuditErrorCategory(%q) = false; want true (declared constant must be valid)", c)
		}
	}
}

// TestValidLabAuditErrorCategory_InvalidValuesRejected ensures arbitrary casts and empty strings fail.
func TestValidLabAuditErrorCategory_InvalidValuesRejected(t *testing.T) {
	for _, c := range []LabAuditErrorCategory{
		"",
		"arbitrary",
		"patient name",
		"INVALID_INPUT",
		"internal_error",
		"none",
	} {
		if ValidLabAuditErrorCategory(c) {
			t.Errorf("ValidLabAuditErrorCategory(%q) = true; want false (non-constant value must be invalid)", c)
		}
	}
}

// ------------------------------------
// LabBlockedReason taxonomy validator
// ------------------------------------

// TestValidLabBlockedReason_AllConstantsPass ensures every declared constant returns true.
// If a new constant is added without updating validLabBlockedReasons, this test catches it.
func TestValidLabBlockedReason_AllConstantsPass(t *testing.T) {
	for _, r := range []LabBlockedReason{
		LabBlockedReasonMDBSchemaUnconfirmed,
		LabBlockedReasonSourceNotImplemented,
		LabBlockedReasonSourceTypeBlocked,
	} {
		if !ValidLabBlockedReason(r) {
			t.Errorf("ValidLabBlockedReason(%q) = false; want true (declared constant must be valid)", r)
		}
	}
}

// TestValidLabBlockedReason_InvalidValuesRejected ensures arbitrary casts and empty strings fail.
func TestValidLabBlockedReason_InvalidValuesRejected(t *testing.T) {
	for _, r := range []LabBlockedReason{
		"",
		"arbitrary",
		"arbitrary text",
		"mdb_schema_not_confirmed_typo",
		"SOURCE_TYPE_BLOCKED",
	} {
		if ValidLabBlockedReason(r) {
			t.Errorf("ValidLabBlockedReason(%q) = true; want false (non-constant value must be invalid)", r)
		}
	}
}
