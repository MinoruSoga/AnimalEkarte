package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// This fake distinguishes base-DB transactions from participation in an ambient
// transaction: nested boundaries independently commit, as the real transactor did.
func TestReservationCancel_DeleteFailureRollsBackPatchedFields(t *testing.T) {
	for _, reschedule := range []bool{false, true} {
		t.Run(map[bool]string{false: "owner patch", true: "owner and schedule patch"}[reschedule], func(t *testing.T) {
			start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)
			stored := model.Reservation{ID: 10, ClinicID: 3, OwnerID: uint64Ptr(1), Status: model.ReservationStatusPending, StartTime: start, EndTime: start.Add(time.Hour)}
			states := make(map[*gorm.DB]*model.Reservation)
			boundaries := 0
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				boundaries++
				handle := &gorm.DB{}
				snapshot := stored
				states[handle] = &snapshot
				err := fn(persistence.WithTxValue(ctx, handle))
				if err == nil {
					stored = snapshot
				}
				delete(states, handle)
				return err
			}}
			read := func(ctx context.Context) *model.Reservation {
				if current := states[persistence.TxFromContext(ctx)]; current != nil {
					return current
				}
				return &stored
			}
			deleteErr := errors.New("forced soft-delete failure")
			repo := &mockReservationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) { copy := stored; return &copy, nil },
				lockAndFindByIDFn: func(ctx context.Context, _ uint64, _ uint64) (*model.Reservation, error) {
					require.NotNil(t, persistence.TxFromContext(ctx))
					copy := *read(ctx)
					return &copy, nil
				},
				countOnDutyDoctorsFn: func(context.Context, uint64, time.Time) (int64, error) { return 1, nil },
				countConflictsFn:     func(context.Context, uint64, time.Time, time.Time, *uint64) (int64, error) { return 0, nil },
				updateFieldsFn: func(ctx context.Context, _ uint64, _ uint64, fields map[string]any) (*model.Reservation, error) {
					current := read(ctx)
					current.OwnerID = uint64Ptr(2)
					current.Status = model.ReservationStatusCancelled
					if value, ok := fields["start_time"].(time.Time); ok {
						current.StartTime = value
					}
					copy := *current
					return &copy, nil
				},
				deleteFn: func(context.Context, uint64, uint64) error { return deleteErr },
			}
			syncer := &recordingNextVisitTagSyncer{}
			svc := newReservationService(repo, nil, tx, nil, nil, nil, nil, openDayHolidayFinder(), syncer)
			cancelled := model.ReservationStatusCancelled
			input := &UpdateReservationInput{OwnerID: uint64Ptr(2), Status: &cancelled}
			if reschedule {
				newStart, newEnd := start.Add(2*time.Hour), start.Add(3*time.Hour)
				input.StartTime = &newStart
				input.EndTime = &newEnd
			}
			_, err := svc.Update(context.Background(), 3, 10, input)
			require.ErrorIs(t, err, deleteErr)
			assert.Equal(t, 1, boundaries, "cancel must own exactly one base-DB transaction")
			require.NotNil(t, stored.OwnerID)
			assert.Equal(t, uint64(1), *stored.OwnerID)
			assert.Equal(t, model.ReservationStatusPending, stored.Status)
			assert.Equal(t, start, stored.StartTime)
			assert.Empty(t, syncer.calls)
		})
	}
}
