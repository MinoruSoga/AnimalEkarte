package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestGuardedCheckInRepositoryCAS(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db).(guardedCheckInWriter)
	tx := testNewTransactor(db)
	for _, tc := range []struct {
		name         string
		clinic       uint64
		status       model.ReservationStatus
		stale        bool
		wantConflict bool
	}{
		{"success", 1, model.ReservationStatusConfirmed, false, false},
		{"stale version", 1, model.ReservationStatusConfirmed, true, true},
		{"diagnosis progressed", 1, model.ReservationStatusInConsultation, false, true},
		{"accounting progressed", 1, model.ReservationStatusAccounting, false, true},
		{"completed", 1, model.ReservationStatusCompleted, false, true},
		{"wrong clinic", 2, model.ReservationStatusConfirmed, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current := makeReservation(t, db, 1)
			require.NoError(t, db.Model(current).Update("status", tc.status).Error)
			require.NoError(t, db.First(current, current.ID).Error)
			expected := current.UpdatedAt
			if tc.stale {
				expected = expected.Add(-time.Second)
			}
			checked := time.Now().Truncate(time.Microsecond)
			err := tx.WithTx(context.Background(), func(ctx context.Context) error {
				got, err := repo.checkInConfirmed(ctx, tc.clinic, current.ID, expected, checked)
				if !tc.wantConflict {
					require.NoError(t, err)
					require.Equal(t, model.ReservationStatusCheckedIn, got.Status)
					require.True(t, got.CheckedInAt.Equal(checked))
				}
				return err
			})
			if tc.wantConflict {
				require.True(t, apperrors.IsConflict(err))
				reloaded := &model.Reservation{}
				require.NoError(t, db.First(reloaded, current.ID).Error)
				require.Equal(t, tc.status, reloaded.Status)
				require.Nil(t, reloaded.CheckedInAt)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
