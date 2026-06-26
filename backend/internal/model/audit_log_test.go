package model

import "testing"

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
