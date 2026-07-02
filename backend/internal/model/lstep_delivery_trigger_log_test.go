package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAllTriggerTypes pins the exhaustive trigger-type list used to validate
// stored TriggerType values. A silent omission here would let an unrecognized
// trigger type slip past request-boundary validation elsewhere.
func TestAllTriggerTypes(t *testing.T) {
	got := AllTriggerTypes()

	want := []string{
		TriggerTypeFirstVisitFollowUp3D,
		TriggerTypeFirstVisitFollowUp7D,
		TriggerTypeNextVisitReminder,
		TriggerTypeVaccineDeadline60,
		TriggerTypeVaccineDeadline30,
		TriggerTypeBirthdayMessage,
		TriggerTypeDormantPrevention180,
		TriggerTypeDormantPrevention210,
		TriggerTypeDormantPrevention240,
		TriggerTypeDormantPrevention365,
		TriggerTypeFilariaAlert,
		TriggerTypeFleaTickAlert,
		TriggerTypeFoodRefillReminder,
		TriggerTypeFirstVisitWelcome,
		TriggerTypeCheckupFollowUp,
	}

	assert.Equal(t, want, got)
	assert.Len(t, got, 15)

	// no duplicates
	seen := map[string]bool{}
	for _, tt := range got {
		assert.False(t, seen[tt], "duplicate trigger type: %s", tt)
		seen[tt] = true
	}
}
