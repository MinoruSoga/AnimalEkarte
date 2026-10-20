package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests compile the actual SQL reader and exercise only boundary refusal;
// they do not execute its SQL or establish transaction/lock behavior.
func TestStayCageEffectiveReader_RefusesInvalidScopeOrMissingDatabase(t *testing.T) {
	reader := NewStayCageEffectiveStateReader(nil)
	for _, tc := range []struct {
		name   string
		clinic uint64
		ids    []uint64
	}{
		{"no clinic", 0, []uint64{1}},
		{"zero stay", 3, []uint64{0}},
		{"database absent", 3, []uint64{1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := reader.ReadEffectiveStatesForCageAssignment(t.Context(), tc.clinic, tc.ids)
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
	got, err := reader.ReadEffectiveStatesForCageAssignment(t.Context(), 3, nil)
	require.NoError(t, err)
	require.Empty(t, got)
}
