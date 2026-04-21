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

// ---- ClinicHoliday モック ----

type mockClinicHolidayRepository struct {
	findByYearMonthFn func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	upsertFn          func(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error)
	deleteFn          func(ctx context.Context, clinicID uint64, date time.Time) error
}

func (m *mockClinicHolidayRepository) FindByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	return m.findByYearMonthFn(ctx, clinicID, yearMonth)
}

func (m *mockClinicHolidayRepository) Upsert(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
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
		name     string
		deleteFn func(ctx context.Context, clinicID uint64, date time.Time) error
		wantErr  bool
	}{
		{
			name: "removes holiday successfully",
			deleteFn: func(_ context.Context, _ uint64, _ time.Time) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "returns nil when holiday does not exist (idempotent)",
			deleteFn: func(_ context.Context, _ uint64, _ time.Time) error {
				return apperrors.WrapNotFound("clinic_holiday", "2026-04-01")
			},
			wantErr: false,
		},
		{
			name: "propagates non-NotFound repository error",
			deleteFn: func(_ context.Context, _ uint64, _ time.Time) error {
				return errors.New("db connection error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicHolidayRepository{
				deleteFn: tt.deleteFn,
			}
			svc := NewClinicHolidayService(repo)
			err := svc.Remove(context.Background(), 1, date)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
