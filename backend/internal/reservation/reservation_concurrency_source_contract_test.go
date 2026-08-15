package reservation

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sourceMethod(t *testing.T, source, signature string) string {
	t.Helper()
	start := strings.Index(source, signature)
	require.NotEqual(t, -1, start, "missing method %q", signature)
	method := source[start:]
	if next := strings.Index(method[len(signature):], "\nfunc "); next >= 0 {
		method = method[:len(signature)+next]
	}
	return method
}

func TestReservationStaffRepository_MutationLockSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("reservation_staff_repository.go")
	require.NoError(t, err)
	source := string(sourceBytes)

	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "staff then assignment use exclusive ownership locks",
			check: func(t *testing.T) {
				method := sourceMethod(t, source, "func lockReservationStaffMutationScope(")
				assert.GreaterOrEqual(t, strings.Count(method, `Strength: "UPDATE"`), 2)
				staffLock := strings.Index(method, "model.Staff")
				assignmentLock := strings.Index(method, "model.StaffClinicAssignment")
				require.NotEqual(t, -1, staffLock)
				require.NotEqual(t, -1, assignmentLock)
				assert.Less(t, staffLock, assignmentLock)
			},
		},
		{
			name: "full replacement takes ownership before reservation type locks",
			check: func(t *testing.T) {
				method := sourceMethod(t, source, "func lockReservationJunctionWriteScope(")
				ownershipLock := strings.Index(method, "lockReservationStaffMutationScope")
				typeLock := strings.Index(method, "lockReservationTypesForShare")
				require.NotEqual(t, -1, ownershipLock)
				require.NotEqual(t, -1, typeLock)
				assert.Less(t, ownershipLock, typeLock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestReservationStaffService_MutationsUseExclusiveOwnershipLock(t *testing.T) {
	sourceBytes, err := os.ReadFile("reservation_staff_service.go")
	require.NoError(t, err)
	source := string(sourceBytes)

	for _, signature := range []string{
		"func (s *reservationStaffService) Update(",
		"func (s *reservationStaffService) PatchStatus(",
	} {
		t.Run(signature, func(t *testing.T) {
			method := sourceMethod(t, source, signature)
			lock := strings.Index(method, "LockForMutation")
			update := strings.Index(method, "s.repo.Update")
			require.NotEqual(t, -1, lock)
			require.NotEqual(t, -1, update)
			assert.Less(t, lock, update)
		})
	}
}

func TestReservationScheduleService_SaveUsesAtomicRepositoryResult(t *testing.T) {
	sourceBytes, err := os.ReadFile("reservation_schedule_service.go")
	require.NoError(t, err)
	method := sourceMethod(
		t,
		string(sourceBytes),
		"func (s *reservationScheduleService) Save(",
	)

	assert.NotContains(t, method, "FindAllByDate")
	assert.NotContains(t, method, "FindAllBreaksByEntryID")
	assert.Contains(t, method, "savedEntry, savedBreaks, created, err := s.repo.Save")
}
