package clinic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ClinicHoliday モック ----

type mockClinicHolidayRepository struct {
	findByDateFn      func(context.Context, uint64, time.Time) (*model.ClinicHoliday, error)
	findByYearMonthFn func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	upsertFn          func(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error)
	deleteFn          func(ctx context.Context, clinicID uint64, date time.Time) error
}

func (m *mockClinicHolidayRepository) FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	return m.findByYearMonthFn(ctx, clinicID, yearMonth)
}

func (m *mockClinicHolidayRepository) Save(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
	return m.upsertFn(ctx, clinicID, holiday)
}

func (m *mockClinicHolidayRepository) Delete(ctx context.Context, clinicID uint64, date time.Time) error {
	return m.deleteFn(ctx, clinicID, date)
}

// ---- Tests ----

func TestClinicHolidayService_List(t *testing.T) {
	date1 := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		yearMonth string
		repoData  []model.ClinicHoliday
		repoErr   error
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "returns holidays for yearMonth",
			yearMonth: "2026-04",
			repoData: []model.ClinicHoliday{
				{ID: 1, ClinicID: 1, Date: date1, Reason: "休診"},
				{ID: 2, ClinicID: 1, Date: date2, Reason: "院内整備"},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:      "returns all holidays when yearMonth is empty",
			yearMonth: "",
			repoData:  []model.ClinicHoliday{{ID: 1, ClinicID: 1, Date: date1}},
			wantLen:   1,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no holidays",
			yearMonth: "2026-04",
			repoData:  []model.ClinicHoliday{},
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			yearMonth: "2026-04",
			repoErr:   errors.New("db error"),
			wantLen:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicHolidayRepository{
				findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewClinicHolidayService(repo)
			holidays, err := svc.List(context.Background(), 1, tt.yearMonth)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, holidays, tt.wantLen)
			}
		})
	}
}

func TestClinicHolidayService_Set(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		date     time.Time
		reason   string
		upsertFn func(ctx context.Context, clinicID uint64, h *model.ClinicHoliday) (*model.ClinicHoliday, error)
		wantErr  bool
	}{
		{
			name:   "creates holiday successfully",
			date:   date,
			reason: "休診日",
			upsertFn: func(_ context.Context, _ uint64, h *model.ClinicHoliday) (*model.ClinicHoliday, error) {
				h.ID = 1
				return h, nil
			},
			wantErr: false,
		},
		{
			name:   "propagates repository error",
			date:   date,
			reason: "休診日",
			upsertFn: func(_ context.Context, _ uint64, _ *model.ClinicHoliday) (*model.ClinicHoliday, error) {
				return nil, errors.New("upsert failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicHolidayRepository{
				upsertFn: tt.upsertFn,
			}
			svc := NewClinicHolidayService(repo)
			result, err := svc.Set(context.Background(), 1, tt.date, tt.reason)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.reason, result.Reason)
			}
		})
	}
}

func TestClinicHolidayService_Remove(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		findByDateFn   func(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error)
		deleteFn       func(ctx context.Context, clinicID uint64, date time.Time) error
		wantErr        bool
		wantDeleteCall bool
	}{
		{
			name: "removes holiday successfully",
			findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClinicHoliday, error) {
				return &model.ClinicHoliday{ID: 1, Date: date}, nil
			},
			deleteFn: func(_ context.Context, _ uint64, _ time.Time) error {
				return nil
			},
			wantErr:        false,
			wantDeleteCall: true,
		},
		{
			name: "returns nil when holiday does not exist (idempotent) — FindByDate NotFound short-circuits before Delete",
			findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClinicHoliday, error) {
				return nil, apperrors.WrapNotFound("clinic_holiday", "2026-04-01")
			},
			wantErr:        false,
			wantDeleteCall: false,
		},
		{
			name: "propagates non-NotFound FindByDate repository error",
			findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClinicHoliday, error) {
				return nil, errors.New("db connection error on find")
			},
			wantErr:        true,
			wantDeleteCall: false,
		},
		{
			name: "returns nil when Delete races to NotFound (already deleted)",
			findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClinicHoliday, error) {
				return &model.ClinicHoliday{ID: 1, Date: date}, nil
			},
			deleteFn: func(_ context.Context, _ uint64, _ time.Time) error {
				return apperrors.WrapNotFound("clinic_holiday", "2026-04-01")
			},
			wantErr:        false,
			wantDeleteCall: true,
		},
		{
			name: "propagates non-NotFound Delete repository error",
			findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClinicHoliday, error) {
				return &model.ClinicHoliday{ID: 1, Date: date}, nil
			},
			deleteFn: func(_ context.Context, _ uint64, _ time.Time) error {
				return errors.New("db connection error")
			},
			wantErr:        true,
			wantDeleteCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteCalled := false
			repo := &mockClinicHolidayRepository{
				findByDateFn: tt.findByDateFn,
				deleteFn: func(ctx context.Context, clinicID uint64, date time.Time) error {
					deleteCalled = true
					if tt.deleteFn != nil {
						return tt.deleteFn(ctx, clinicID, date)
					}
					return nil
				},
			}
			svc := NewClinicHolidayService(repo)
			err := svc.Remove(context.Background(), 1, date)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantDeleteCall, deleteCalled)
		})
	}
}

func (m *mockClinicHolidayRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
	if m.findByDateFn != nil {
		return m.findByDateFn(ctx, clinicID, date)
	}
	return nil, nil
}
