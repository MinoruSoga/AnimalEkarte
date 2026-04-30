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

// ---- Checkup モック ----

type mockCheckupRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	listByClinicFn          func(ctx context.Context, clinicID uint64, filters repository.CheckupFilters) ([]model.Checkup, error)
	findByOwnerIDFn         func(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error)
	findByIDFn              func(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error)
	createFn                func(ctx context.Context, checkup *model.Checkup) error
	updateFn                func(ctx context.Context, clinicID, checkupID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, checkupID uint64) error
}

func (m *mockCheckupRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockCheckupRepository) FindByClinicID(ctx context.Context, clinicID uint64, filters repository.CheckupFilters) ([]model.Checkup, error) {
	if m.listByClinicFn != nil {
		return m.listByClinicFn(ctx, clinicID, filters)
	}
	return nil, nil
}

func (m *mockCheckupRepository) FindByID(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, checkupID)
	}
	return nil, nil
}

func (m *mockCheckupRepository) FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error) {
	if m.findByOwnerIDFn != nil {
		return m.findByOwnerIDFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockCheckupRepository) Create(ctx context.Context, checkup *model.Checkup) error {
	return m.createFn(ctx, checkup)
}

func (m *mockCheckupRepository) Update(ctx context.Context, clinicID, checkupID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, checkupID, fields)
}

func (m *mockCheckupRepository) Delete(ctx context.Context, clinicID, checkupID uint64) error {
	return m.deleteFn(ctx, clinicID, checkupID)
}

// ---- Tests ----

func TestCheckupService_List(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		repoCheckups    []model.Checkup
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:            "returns checkups for medical record",
			medicalRecordID: 1,
			repoCheckups: []model.Checkup{
				{ID: 1, MedicalRecordID: 1, CheckupTypeID: 1, Result: "Normal"},
				{ID: 2, MedicalRecordID: 1, CheckupTypeID: 2, Result: "Abnormal"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty list when no checkups exist",
			medicalRecordID: 999,
			repoCheckups:    []model.Checkup{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoCheckups:    nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return tt.repoCheckups, tt.repoErr
				},
			}
			svc := NewCheckupService(repo)

			checkups, err := svc.List(context.Background(), 1, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, checkups, tt.wantLen)
			}
		})
	}
}

func TestCheckupService_Create(t *testing.T) {
	now := time.Now()
	petID := uint64(5)
	doctorID := uint64(10)

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *CreateCheckupInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "creates checkup successfully",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
				PetID:         &petID,
				DoctorID:      &doctorID,
				Result:        "Normal findings",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "creates checkup with minimal fields",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				createFn: func(_ context.Context, _ *model.Checkup) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{ID: 1, MedicalRecordID: tt.medicalRecordID}, nil
				},
			}
			svc := NewCheckupService(repo)

			checkup, err := svc.Create(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, checkup)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checkup)
			}
		})
	}
}

func TestCheckupService_Update(t *testing.T) {
	newResult := "Updated result"
	newCheckupTypeID := uint64(2)

	tests := []struct {
		name                       string
		medicalRecordID            uint64
		checkupID                  uint64
		input                      *UpdateCheckupInput
		repoCheckupMedicalRecordID uint64
		repoUpdateErr              error
		repoReturnCheckup          *model.Checkup
		wantErr                    bool
	}{
		{
			name:            "updates checkup successfully",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result:        &newResult,
				CheckupTypeID: &newCheckupTypeID,
			},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              nil,
			repoReturnCheckup: &model.Checkup{
				ID:              1,
				MedicalRecordID: 1,
				Result:          newResult,
			},
			wantErr: false,
		},
		{
			name:                       "returns error when no fields provided",
			medicalRecordID:            1,
			checkupID:                  1,
			input:                      &UpdateCheckupInput{},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              nil,
			repoReturnCheckup:          nil,
			wantErr:                    true,
		},
		{
			name:            "returns error when checkup doesn't belong to medical record",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result: &newResult,
			},
			repoCheckupMedicalRecordID: 2, // Different medical record
			repoUpdateErr:              nil,
			repoReturnCheckup: &model.Checkup{
				ID:              1,
				MedicalRecordID: 2,
			},
			wantErr: true,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result: &newResult,
			},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              errors.New("db error"),
			repoReturnCheckup:          nil,
			wantErr:                    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{
						ID:              tt.checkupID,
						MedicalRecordID: tt.repoCheckupMedicalRecordID,
					}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewCheckupService(repo)

			checkup, err := svc.Update(context.Background(), 1, tt.medicalRecordID, tt.checkupID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checkup)
			}
		})
	}
}

func TestCheckupService_Delete(t *testing.T) {
	tests := []struct {
		name                       string
		medicalRecordID            uint64
		checkupID                  uint64
		repoCheckupMedicalRecordID uint64
		repoDeleteErr              error
		wantErr                    bool
	}{
		{
			name:                       "deletes checkup successfully",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 1,
			repoDeleteErr:              nil,
			wantErr:                    false,
		},
		{
			name:                       "returns error when checkup doesn't belong to medical record",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 2, // Different medical record
			repoDeleteErr:              nil,
			wantErr:                    true,
		},
		{
			name:                       "returns error when delete fails",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 1,
			repoDeleteErr:              errors.New("db error"),
			wantErr:                    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{
						ID:              tt.checkupID,
						MedicalRecordID: tt.repoCheckupMedicalRecordID,
					}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoDeleteErr
				},
			}
			svc := NewCheckupService(repo)

			err := svc.Delete(context.Background(), 1, tt.medicalRecordID, tt.checkupID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
