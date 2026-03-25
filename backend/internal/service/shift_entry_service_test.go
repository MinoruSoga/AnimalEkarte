package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ShiftEntry モック ----

type mockShiftEntryRepository struct {
	listFn     func(ctx context.Context, clinicID uint64, filter repository.ShiftEntryFilter) ([]model.ShiftEntry, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error)
	createFn   func(ctx context.Context, entry *model.ShiftEntry) error
	updateFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockShiftEntryRepository) List(ctx context.Context, clinicID uint64, filter repository.ShiftEntryFilter) ([]model.ShiftEntry, error) {
	return m.listFn(ctx, clinicID, filter)
}

func (m *mockShiftEntryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockShiftEntryRepository) Create(ctx context.Context, entry *model.ShiftEntry) error {
	return m.createFn(ctx, entry)
}

func (m *mockShiftEntryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockShiftEntryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

// ---- Tests ----

func TestShiftEntryService_List(t *testing.T) {
	tests := []struct {
		name        string
		clinicID    uint64
		yearMonth   string
		staffID     *uint64
		repoEntries []model.ShiftEntry
		repoErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name:      "returns shifts for clinic without filter",
			clinicID:  1,
			yearMonth: "",
			staffID:   nil,
			repoEntries: []model.ShiftEntry{
				{ID: 1, ClinicID: 1, StaffID: 1, ShiftType: model.ShiftTypeMorning},
				{ID: 2, ClinicID: 1, StaffID: 2, ShiftType: model.ShiftTypeAfternoon},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:        "returns empty list when no shifts exist",
			clinicID:    1,
			yearMonth:   "2024-03",
			staffID:     nil,
			repoEntries: []model.ShiftEntry{},
			repoErr:     nil,
			wantLen:     0,
			wantErr:     false,
		},
		{
			name:        "returns error on invalid yearMonth format",
			clinicID:    1,
			yearMonth:   "2024/03",
			staffID:     nil,
			repoEntries: nil,
			repoErr:     nil,
			wantErr:     true,
		},
		{
			name:        "returns error on invalid yearMonth value",
			clinicID:    1,
			yearMonth:   "2024-13",
			staffID:     nil,
			repoEntries: nil,
			repoErr:     nil,
			wantErr:     true,
		},
		{
			name:        "propagates repository error",
			clinicID:    1,
			yearMonth:   "2024-03",
			staffID:     nil,
			repoEntries: nil,
			repoErr:     errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				listFn: func(_ context.Context, _ uint64, _ repository.ShiftEntryFilter) ([]model.ShiftEntry, error) {
					return tt.repoEntries, tt.repoErr
				},
			}
			svc := NewShiftEntryService(repo)

			entries, err := svc.List(context.Background(), tt.clinicID, tt.yearMonth, tt.staffID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, entries, tt.wantLen)
			}
		})
	}
}

func TestShiftEntryService_Create(t *testing.T) {
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	startTime := "09:00:00"
	endTime := "18:00:00"

	tests := []struct {
		name     string
		clinicID uint64
		input    *CreateShiftEntryInput
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "creates shift entry successfully",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: model.ShiftTypeMorning,
				StartTime: &startTime,
				EndTime:   &endTime,
				Note:      "Regular shift",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "creates shift without times",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: model.ShiftTypeOff,
				Note:      "Day off",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when repository fails",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: model.ShiftTypeMorning,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				createFn: func(_ context.Context, _ *model.ShiftEntry) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
					return &model.ShiftEntry{ID: 1, ClinicID: tt.clinicID}, nil
				},
			}
			svc := NewShiftEntryService(repo)

			entry, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, entry)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
			}
		})
	}
}

func TestShiftEntryService_Update(t *testing.T) {
	newShiftType := model.ShiftTypeAfternoon
	newStartTime := "15:00"
	newNote := "Updated note"

	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		input           *UpdateShiftEntryInput
		repoUpdateErr   error
		repoReturnEntry *model.ShiftEntry
		wantErr         bool
	}{
		{
			name:     "updates shift entry successfully",
			clinicID: 1,
			id:       1,
			input: &UpdateShiftEntryInput{
				ShiftType: &newShiftType,
				StartTime: &newStartTime,
				Note:      &newNote,
			},
			repoUpdateErr: nil,
			repoReturnEntry: &model.ShiftEntry{
				ID:        1,
				ShiftType: newShiftType,
			},
			wantErr: false,
		},
		{
			name:            "returns error when no fields provided",
			clinicID:        1,
			id:              1,
			input:           &UpdateShiftEntryInput{},
			repoUpdateErr:   nil,
			repoReturnEntry: nil,
			wantErr:         true,
		},
		{
			name:     "returns error when update fails",
			clinicID: 1,
			id:       1,
			input: &UpdateShiftEntryInput{
				Note: &newNote,
			},
			repoUpdateErr:   errors.New("db error"),
			repoReturnEntry: nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
					return tt.repoReturnEntry, nil
				},
			}
			svc := NewShiftEntryService(repo)

			entry, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
			}
		})
	}
}

func TestShiftEntryService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "deletes shift entry successfully",
			clinicID: 1,
			id:       1,
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns error when shift not found",
			clinicID: 1,
			id:       999,
			repoErr:  errors.New("not found"),
			wantErr:  true,
		},
		{
			name:     "returns error when delete fails",
			clinicID: 1,
			id:       1,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewShiftEntryService(repo)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
