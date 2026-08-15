package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type reservationScheduleWriterStub struct {
	saveFn func(
		ctx context.Context,
		clinicID uint64,
		entry *model.ShiftEntry,
		breaks []model.ShiftEntryBreak,
	) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error)
}

func (s *reservationScheduleWriterStub) SaveByStaffDate(
	ctx context.Context,
	clinicID uint64,
	entry *model.ShiftEntry,
	breaks []model.ShiftEntryBreak,
) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
	return s.saveFn(ctx, clinicID, entry, breaks)
}

func (s *reservationScheduleWriterStub) DeleteByStaffDate(
	_ context.Context,
	_, _ uint64,
	_ time.Time,
) error {
	return nil
}

func TestReservationScheduleRepository_SaveReturnsWriterTransactionResult(t *testing.T) {
	saveErr := errors.New("save failed")
	tests := []struct {
		name        string
		writerEntry *model.ShiftEntry
		writerBreak []model.ShiftEntryBreak
		created     bool
		writerErr   error
	}{
		{
			name:        "created aggregate",
			writerEntry: &model.ShiftEntry{ID: 10, ClinicID: 1, StaffID: 2},
			writerBreak: []model.ShiftEntryBreak{{ID: 20, ShiftEntryID: 10}},
			created:     true,
		},
		{
			name:      "writer error",
			writerErr: saveErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputEntry := &model.ShiftEntry{ClinicID: 1, StaffID: 2}
			inputBreaks := []model.ShiftEntryBreak{{BreakStart: "12:00", BreakEnd: "13:00"}}
			writer := &reservationScheduleWriterStub{
				saveFn: func(
					_ context.Context,
					clinicID uint64,
					entry *model.ShiftEntry,
					breaks []model.ShiftEntryBreak,
				) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Same(t, inputEntry, entry)
					assert.Equal(t, inputBreaks, breaks)
					return tt.writerEntry, tt.writerBreak, tt.created, tt.writerErr
				},
			}
			repo := NewReservationScheduleRepository(nil, writer)

			entry, breaks, created, err := repo.Save(
				context.Background(),
				1,
				inputEntry,
				inputBreaks,
			)

			if tt.writerErr != nil {
				assert.ErrorIs(t, err, tt.writerErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Same(t, tt.writerEntry, entry)
			assert.Equal(t, tt.writerBreak, breaks)
			assert.Equal(t, tt.created, created)
		})
	}
}
