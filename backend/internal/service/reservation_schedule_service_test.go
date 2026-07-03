package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationScheduleRepository は ReservationScheduleRepository のテスト用モック実装
type mockReservationScheduleRepository struct {
	findAllByMonthFn          func(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error)
	findAllBreaksByEntryIDsFn func(ctx context.Context, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error)
	findAllByDateFn           func(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
	findAllBreaksByEntryIDFn  func(ctx context.Context, entryID uint64) ([]model.ShiftEntryBreak, error)
	saveFn                    func(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error
	deleteFn                  func(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

func (m *mockReservationScheduleRepository) FindAllByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error) {
	if m.findAllByMonthFn != nil {
		return m.findAllByMonthFn(ctx, clinicID, staffID, month)
	}
	return nil, nil
}

func (m *mockReservationScheduleRepository) FindAllBreaksByEntryIDs(ctx context.Context, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
	if m.findAllBreaksByEntryIDsFn != nil {
		return m.findAllBreaksByEntryIDsFn(ctx, entryIDs)
	}
	return map[uint64][]model.ShiftEntryBreak{}, nil
}

func (m *mockReservationScheduleRepository) FindAllByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error) {
	if m.findAllByDateFn != nil {
		return m.findAllByDateFn(ctx, clinicID, staffID, date)
	}
	return nil, nil
}

func (m *mockReservationScheduleRepository) FindAllBreaksByEntryID(ctx context.Context, entryID uint64) ([]model.ShiftEntryBreak, error) {
	if m.findAllBreaksByEntryIDFn != nil {
		return m.findAllBreaksByEntryIDFn(ctx, entryID)
	}
	return nil, nil
}

func (m *mockReservationScheduleRepository) Save(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, clinicID, entry, breaks)
	}
	return nil
}

func (m *mockReservationScheduleRepository) Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, staffID, date)
	}
	return nil
}

func TestNewReservationScheduleService(t *testing.T) {
	svc := NewReservationScheduleService(&mockReservationScheduleRepository{})
	assert.NotNil(t, svc)
}

func TestReservationScheduleService_ListByMonth(t *testing.T) {
	clinicID, staffID := uint64(1), uint64(2)

	t.Run("returns entries with breaks", func(t *testing.T) {
		repo := &mockReservationScheduleRepository{
			findAllByMonthFn: func(_ context.Context, _, _ uint64, _ string) ([]model.ShiftEntry, error) {
				return []model.ShiftEntry{{ID: 1}, {ID: 2}}, nil
			},
			findAllBreaksByEntryIDsFn: func(_ context.Context, _ []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
				return map[uint64][]model.ShiftEntryBreak{
					1: {{ID: 10, BreakStart: "12:00:00", BreakEnd: "13:00:00"}},
				}, nil
			},
		}
		svc := NewReservationScheduleService(repo)
		result, err := svc.ListByMonth(context.Background(), clinicID, staffID, "2026-07")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Len(t, result[0].Breaks, 1)
		assert.Empty(t, result[1].Breaks)
	})

	t.Run("returns empty slice when no entries", func(t *testing.T) {
		repo := &mockReservationScheduleRepository{
			findAllByMonthFn: func(_ context.Context, _, _ uint64, _ string) ([]model.ShiftEntry, error) {
				return []model.ShiftEntry{}, nil
			},
		}
		svc := NewReservationScheduleService(repo)
		result, err := svc.ListByMonth(context.Background(), clinicID, staffID, "2026-07")
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("propagates find all by month error", func(t *testing.T) {
		repo := &mockReservationScheduleRepository{
			findAllByMonthFn: func(_ context.Context, _, _ uint64, _ string) ([]model.ShiftEntry, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewReservationScheduleService(repo)
		result, err := svc.ListByMonth(context.Background(), clinicID, staffID, "2026-07")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("propagates find all breaks error", func(t *testing.T) {
		repo := &mockReservationScheduleRepository{
			findAllByMonthFn: func(_ context.Context, _, _ uint64, _ string) ([]model.ShiftEntry, error) {
				return []model.ShiftEntry{{ID: 1}}, nil
			},
			findAllBreaksByEntryIDsFn: func(_ context.Context, _ []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
				return nil, errors.New("breaks db error")
			},
		}
		svc := NewReservationScheduleService(repo)
		result, err := svc.ListByMonth(context.Background(), clinicID, staffID, "2026-07")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestReservationScheduleService_Save(t *testing.T) {
	clinicID, staffID := uint64(1), uint64(2)
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	start := "09:00:00"
	end := "18:00:00"

	tests := []struct {
		name         string
		input        *CreateReservationScheduleInput
		findByDateFn func(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
		saveFn       func(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error
		findBreaksFn func(ctx context.Context, entryID uint64) ([]model.ShiftEntryBreak, error)
		wantErr      bool
		wantIsNew    bool
	}{
		{
			name:  "creates new schedule when none exists (not found)",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end},
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return nil, apperrors.WrapNotFound("shift_entry", "1")
			},
			saveFn: func(_ context.Context, _ uint64, entry *model.ShiftEntry, _ []model.ShiftEntryBreak) error {
				entry.ID = 100
				return nil
			},
			wantErr:   false,
			wantIsNew: true,
		},
		{
			name:  "updates existing schedule",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end, Breaks: []ReservationScheduleBreakInput{{Start: "12:00:00", End: "13:00:00"}}},
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{ID: 5}, nil
			},
			saveFn: func(_ context.Context, _ uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error {
				entry.ID = 5
				assert.Len(t, breaks, 1)
				return nil
			},
			wantErr:   false,
			wantIsNew: false,
		},
		{
			name:  "invalid shift times returns error before find",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &end, WorkEnd: &start}, // end before start
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return nil, errors.New("should not be called")
			},
			wantErr: true,
		},
		{
			name:  "find by date repository error propagates",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end},
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "repo save error propagates",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end},
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{ID: 5}, nil
			},
			saveFn: func(_ context.Context, _ uint64, _ *model.ShiftEntry, _ []model.ShiftEntryBreak) error {
				return errors.New("save error")
			},
			wantErr: true,
		},
		{
			name:  "find breaks after save error propagates",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end},
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{ID: 5}, nil
			},
			saveFn: func(_ context.Context, _ uint64, entry *model.ShiftEntry, _ []model.ShiftEntryBreak) error {
				entry.ID = 5
				return nil
			},
			findBreaksFn: func(_ context.Context, _ uint64) ([]model.ShiftEntryBreak, error) {
				return nil, errors.New("breaks error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationScheduleRepository{
				findAllByDateFn:          tt.findByDateFn,
				saveFn:                   tt.saveFn,
				findAllBreaksByEntryIDFn: tt.findBreaksFn,
			}
			svc := NewReservationScheduleService(repo)
			result, isNew, err := svc.Save(context.Background(), clinicID, staffID, date, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantIsNew, isNew)
			}
		})
	}
}

func TestReservationScheduleService_Delete(t *testing.T) {
	clinicID, staffID := uint64(1), uint64(2)
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		findByDateFn func(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
		deleteFn     func(ctx context.Context, clinicID, staffID uint64, date time.Time) error
		wantErr      bool
	}{
		{
			name: "deletes successfully",
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{ID: 1}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64, _ time.Time) error { return nil },
			wantErr:  false,
		},
		{
			name: "returns error when schedule not found",
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return nil, apperrors.WrapNotFound("shift_entry", "1")
			},
			wantErr: true,
		},
		{
			name: "returns error on repository delete failure",
			findByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{ID: 1}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64, _ time.Time) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationScheduleRepository{
				findAllByDateFn: tt.findByDateFn,
				deleteFn:        tt.deleteFn,
			}
			svc := NewReservationScheduleService(repo)
			err := svc.Delete(context.Background(), clinicID, staffID, date)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
