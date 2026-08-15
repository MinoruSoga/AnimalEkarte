package staff_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func functionSource(t *testing.T, source, signature string) string {
	t.Helper()
	start := strings.Index(source, signature)
	require.NotEqual(t, -1, start, "missing function %s", signature)
	next := strings.Index(source[start+len(signature):], "\nfunc ")
	if next == -1 {
		return source[start:]
	}
	return source[start : start+len(signature)+next]
}

func TestShiftEntryRepository_ScheduleMutationSourceContract(t *testing.T) {
	source, err := os.ReadFile("shift_entry_repository.go")
	require.NoError(t, err)
	text := string(source)

	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "save and delete share the canonical staff assignment shift lock order",
			check: func(t *testing.T) {
				ownershipScope := functionSource(
					t,
					text,
					"func lockShiftScheduleMutationScope(",
				)
				staffLock := strings.Index(
					ownershipScope,
					`Where("staffs.id = ? AND staffs.deleted_at IS NULL"`,
				)
				assignmentLock := strings.Index(
					ownershipScope,
					`Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL"`,
				)
				require.NotEqual(t, -1, staffLock)
				require.NotEqual(t, -1, assignmentLock)
				assert.Less(t, staffLock, assignmentLock)

				methods := []struct {
					name      string
					signature string
				}{
					{
						name:      "save",
						signature: "func (r *shiftEntryRepository) SaveByStaffDate(",
					},
					{
						name:      "delete",
						signature: "func (r *shiftEntryRepository) DeleteByStaffDate(",
					},
				}
				for _, method := range methods {
					t.Run(method.name, func(t *testing.T) {
						body := functionSource(t, text, method.signature)
						ownershipLock := strings.Index(
							body,
							"lockShiftScheduleMutationScope",
						)
						shiftLock := strings.Index(
							body,
							"lockShiftEntryByStaffDateForUpdate",
						)
						require.NotEqual(t, -1, ownershipLock)
						require.NotEqual(t, -1, shiftLock)
						assert.Less(t, ownershipLock, shiftLock)
					})
				}
			},
		},
		{
			name: "schedule ownership locks are exclusive",
			check: func(t *testing.T) {
				method := functionSource(
					t,
					text,
					"func lockShiftScheduleMutationScope(",
				)
				assert.GreaterOrEqual(t, strings.Count(method, `Strength: "UPDATE"`), 2)
			},
		},
		{
			name: "upsert returns persisted aggregate and transaction result",
			check: func(t *testing.T) {
				assert.Contains(
					t,
					text,
					"(*model.ShiftEntry, []model.ShiftEntryBreak, bool, error)",
				)
				assert.Contains(t, text, "RowsAffected")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}
