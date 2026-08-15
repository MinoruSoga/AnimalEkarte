package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationScheduleRepository は ReservationScheduleRepository のテスト用モック実装
type mockReservationScheduleRepository struct {
	findAllByMonthFn                func(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error)
	findAllByStaffIDsAndDateRangeFn func(ctx context.Context, clinicID uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error)
	findAllBreaksByEntryIDsFn       func(ctx context.Context, clinicID uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error)
	findAllByDateFn                 func(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
	findAllBreaksByEntryIDFn        func(ctx context.Context, clinicID, entryID uint64) ([]model.ShiftEntryBreak, error)
	saveFn                          func(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error)
	deleteFn                        func(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

func (m *mockReservationScheduleRepository) FindAllByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error) {
	if m.findAllByMonthFn != nil {
		return m.findAllByMonthFn(ctx, clinicID, staffID, month)
	}
	return nil, nil
}

func (m *mockReservationScheduleRepository) FindAllByStaffIDsAndDateRange(ctx context.Context, clinicID uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error) {
	if m.findAllByStaffIDsAndDateRangeFn != nil {
		return m.findAllByStaffIDsAndDateRangeFn(ctx, clinicID, staffIDs, from, to)
	}
	return nil, nil
}

func (m *mockReservationScheduleRepository) FindAllBreaksByEntryIDs(ctx context.Context, clinicID uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
	if m.findAllBreaksByEntryIDsFn != nil {
		return m.findAllBreaksByEntryIDsFn(ctx, clinicID, entryIDs)
	}
	return map[uint64][]model.ShiftEntryBreak{}, nil
}

func (m *mockReservationScheduleRepository) FindAllByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error) {
	if m.findAllByDateFn != nil {
		return m.findAllByDateFn(ctx, clinicID, staffID, date)
	}
	return nil, nil
}

func (m *mockReservationScheduleRepository) FindAllBreaksByEntryID(ctx context.Context, clinicID, entryID uint64) ([]model.ShiftEntryBreak, error) {
	if m.findAllBreaksByEntryIDFn != nil {
		return m.findAllBreaksByEntryIDFn(ctx, clinicID, entryID)
	}
	return nil, nil
}

func (m *mockReservationScheduleRepository) Save(
	ctx context.Context,
	clinicID uint64,
	entry *model.ShiftEntry,
	breaks []model.ShiftEntryBreak,
) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
	if m.saveFn != nil {
		return m.saveFn(ctx, clinicID, entry, breaks)
	}
	return nil, nil, false, nil
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
			findAllBreaksByEntryIDsFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
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
			findAllBreaksByEntryIDsFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
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
		name       string
		input      *CreateReservationScheduleInput
		saveFn     func(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error)
		want       *ScheduleEntry
		wantErr    bool
		wantIsNew  bool
		wantSaves  int
		assertSave func(t *testing.T, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak)
	}{
		{
			name:      "nil input is rejected before save",
			input:     nil,
			wantErr:   true,
			wantSaves: 0,
		},
		{
			name:  "returns transaction-created aggregate",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end},
			saveFn: func(_ context.Context, _ uint64, _ *model.ShiftEntry, _ []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
				return &model.ShiftEntry{
						ID:        100,
						ClinicID:  clinicID,
						StaffID:   staffID,
						Date:      date,
						ShiftType: model.ShiftTypeFull,
						StartTime: &start,
						EndTime:   &end,
					},
					[]model.ShiftEntryBreak{{ID: 10, ShiftEntryID: 100, BreakStart: "12:00:00", BreakEnd: "13:00:00"}},
					true,
					nil
			},
			want: &ScheduleEntry{
				Entry: model.ShiftEntry{
					ID:        100,
					ClinicID:  clinicID,
					StaffID:   staffID,
					Date:      date,
					ShiftType: model.ShiftTypeFull,
					StartTime: &start,
					EndTime:   &end,
				},
				Breaks: []model.ShiftEntryBreak{{ID: 10, ShiftEntryID: 100, BreakStart: "12:00:00", BreakEnd: "13:00:00"}},
			},
			wantErr:   false,
			wantIsNew: true,
			wantSaves: 1,
		},
		{
			name:  "returns transaction-updated aggregate",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end, Breaks: []ReservationScheduleBreakInput{{Start: "12:00:00", End: "13:00:00"}}},
			saveFn: func(_ context.Context, _ uint64, _ *model.ShiftEntry, _ []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
				return &model.ShiftEntry{ID: 5, ClinicID: clinicID, StaffID: staffID, Date: date, ShiftType: model.ShiftTypeFull},
					[]model.ShiftEntryBreak{{ID: 6, ShiftEntryID: 5, BreakStart: "12:00:00", BreakEnd: "13:00:00"}},
					false,
					nil
			},
			want: &ScheduleEntry{
				Entry:  model.ShiftEntry{ID: 5, ClinicID: clinicID, StaffID: staffID, Date: date, ShiftType: model.ShiftTypeFull},
				Breaks: []model.ShiftEntryBreak{{ID: 6, ShiftEntryID: 5, BreakStart: "12:00:00", BreakEnd: "13:00:00"}},
			},
			wantErr:   false,
			wantIsNew: false,
			wantSaves: 1,
			assertSave: func(t *testing.T, _ *model.ShiftEntry, breaks []model.ShiftEntryBreak) {
				assert.Len(t, breaks, 1)
			},
		},
		{
			name:    "invalid shift times returns error before save",
			input:   &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &end, WorkEnd: &start}, // end before start
			wantErr: true,
		},
		{
			name:  "repo save error propagates",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end},
			saveFn: func(_ context.Context, _ uint64, _ *model.ShiftEntry, _ []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
				return nil, nil, false, errors.New("save error")
			},
			wantErr:   true,
			wantSaves: 1,
		},
		{
			name:  "nil transaction result is rejected",
			input: &CreateReservationScheduleInput{ShiftType: string(model.ShiftTypeFull), WorkStart: &start, WorkEnd: &end},
			saveFn: func(_ context.Context, _ uint64, _ *model.ShiftEntry, _ []model.ShiftEntryBreak) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
				return nil, nil, true, nil
			},
			wantErr:   true,
			wantSaves: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readCalls := 0
			saveCalls := 0
			repo := &mockReservationScheduleRepository{
				findAllByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.ShiftEntry, error) {
					readCalls++
					return nil, errors.New("save must not pre-read")
				},
				findAllBreaksByEntryIDFn: func(_ context.Context, _, _ uint64) ([]model.ShiftEntryBreak, error) {
					readCalls++
					return nil, errors.New("save must not post-read")
				},
				saveFn: func(
					ctx context.Context,
					gotClinicID uint64,
					entry *model.ShiftEntry,
					breaks []model.ShiftEntryBreak,
				) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
					saveCalls++
					if tt.assertSave != nil {
						tt.assertSave(t, entry, breaks)
					}
					return tt.saveFn(ctx, gotClinicID, entry, breaks)
				},
			}
			svc := NewReservationScheduleService(repo)
			result, isNew, err := svc.Save(context.Background(), clinicID, staffID, date, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
				assert.Equal(t, tt.wantIsNew, isNew)
			}
			assert.Zero(t, readCalls, "Save must use only the same-transaction repository result")
			assert.Equal(t, tt.wantSaves, saveCalls)
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
