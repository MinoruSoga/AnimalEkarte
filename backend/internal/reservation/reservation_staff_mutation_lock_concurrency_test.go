package reservation

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestReservationStaffRepository_LockForMutationRequiresTransaction(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		staffID  uint64
	}{
		{name: "missing ambient transaction", clinicID: 1, staffID: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewReservationStaffRepository(nil, nil)

			staff, err := repo.LockForMutation(
				context.Background(),
				tt.clinicID,
				tt.staffID,
			)

			assert.Error(t, err)
			assert.Nil(t, staff)
		})
	}
}

func TestReservationStaffRepository_LockForMutationSerializesSameStaff(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
	}{
		{name: "same clinic mutation", clinicID: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupReservationRepoTestDB(t)
			repo := NewReservationStaffRepository(db, nil)
			staff := makeDoctor(t, db, tt.clinicID, "exclusive mutation lock")
			require.NoError(t, db.Create(&model.StaffClinicAssignment{
				StaffID:  staff.ID,
				ClinicID: tt.clinicID,
				IsMain:   true,
			}).Error)

			holder := db.Begin()
			require.NoError(t, holder.Error)
			defer holder.Rollback()
			locked, err := repo.LockForMutation(
				persistence.WithTxValue(context.Background(), holder),
				tt.clinicID,
				staff.ID,
			)
			require.NoError(t, err)
			require.Equal(t, staff.ID, locked.ID)

			contender := db.Begin()
			require.NoError(t, contender.Error)
			defer contender.Rollback()
			require.NoError(t, contender.Exec("SET LOCAL lock_timeout = '100ms'").Error)
			_, contenderErr := repo.LockForMutation(
				persistence.WithTxValue(context.Background(), contender),
				tt.clinicID,
				staff.ID,
			)

			require.Error(t, contenderErr)
			assert.True(
				t,
				strings.Contains(contenderErr.Error(), "55P03") ||
					strings.Contains(contenderErr.Error(), "lock timeout"),
				"same-staff mutation must wait for exclusive ownership, got: %v",
				contenderErr,
			)
		})
	}
}
