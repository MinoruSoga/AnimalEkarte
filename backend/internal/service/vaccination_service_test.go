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

// mockVaccinationRepository は VaccinationRepository のテスト用モック実装
type mockVaccinationRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	createFn       func(ctx context.Context, vaccination *model.Vaccination) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockVaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockVaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockVaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	return m.createFn(ctx, vaccination)
}

func (m *mockVaccinationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockVaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func TestVaccinationService_List(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name             string
		clinicID         uint64
		petID            *uint64
		ownerID          *uint64
		page             int
		limit            int
		repoVaccinations []model.Vaccination
		repoTotal        int64
		repoErr          error
		wantLen          int
		wantTotal        int64
		wantErr          bool
	}{
		{
			name:     "returns all vaccinations without filter",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			page:     1,
			limit:    20,
			repoVaccinations: []model.Vaccination{
				{ID: 1, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
				{ID: 2, MedicalRecordID: ptrUint64(2), VaccineID: 2, Date: now},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			petID:    ptrUint64(10),
			ownerID:  nil,
			page:     1,
			limit:    20,
			repoVaccinations: []model.Vaccination{
				{ID: 1, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			petID:    nil,
			ownerID:  ptrUint64(5),
			page:     1,
			limit:    20,
			repoVaccinations: []model.Vaccination{
				{ID: 2, MedicalRecordID: ptrUint64(2), VaccineID: 2, Date: now},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:             "returns empty list when no vaccinations exist",
			clinicID:         1,
			petID:            nil,
			ownerID:          nil,
			page:             1,
			limit:            20,
			repoVaccinations: []model.Vaccination{},
			repoTotal:        0,
			repoErr:          nil,
			wantLen:          0,
			wantTotal:        0,
			wantErr:          false,
		},
		{
			name:             "propagates repository error",
			clinicID:         1,
			petID:            nil,
			ownerID:          nil,
			page:             1,
			limit:            20,
			repoVaccinations: nil,
			repoTotal:        0,
			repoErr:          errors.New("db connection error"),
			wantLen:          0,
			wantTotal:        0,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPetID := (*uint64)(nil)
			capturedOwnerID := (*uint64)(nil)
			repo := &mockVaccinationRepository{
				findAllFn: func(_ context.Context, _ uint64, petID *uint64, ownerID *uint64, _, _ *string, _, _ int) ([]model.Vaccination, int64, error) {
					capturedPetID = petID
					capturedOwnerID = ownerID
					return tt.repoVaccinations, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewVaccinationService(repo)

			vaccinations, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, nil, nil, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vaccinations, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, tt.petID, capturedPetID)
				assert.Equal(t, tt.ownerID, capturedOwnerID)
			}
		})
	}
}

func TestVaccinationService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		repoVaccination *model.Vaccination
		repoErr         error
		wantVaccination *model.Vaccination
		wantErr         error
	}{
		{
			name:            "returns vaccination when found",
			clinicID:        1,
			id:              10,
			repoVaccination: &model.Vaccination{ID: 10, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
			repoErr:         nil,
			wantVaccination: &model.Vaccination{ID: 10, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
			wantErr:         nil,
		},
		{
			name:            "returns not found error when vaccination does not exist",
			clinicID:        1,
			id:              999,
			repoVaccination: nil,
			repoErr:         apperrors.WrapNotFound("vaccination", "999"),
			wantVaccination: nil,
			wantErr:         apperrors.ErrNotFound,
		},
		{
			name:            "returns error on repository failure",
			clinicID:        1,
			id:              10,
			repoVaccination: nil,
			repoErr:         errors.New("db error"),
			wantVaccination: nil,
			wantErr:         errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return tt.repoVaccination, tt.repoErr
				},
			}
			svc := NewVaccinationService(repo)

			vaccination, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVaccination, vaccination)
			}
		})
	}
}

func TestVaccinationService_GetByID_NotFound(t *testing.T) {
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, apperrors.WrapNotFound("vaccination", "999")
		},
	}
	svc := NewVaccinationService(repo)

	vaccination, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, vaccination)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestVaccinationService_Create(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		clinicID uint64
		input    *CreateVaccinationInput
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "creates vaccination successfully",
			clinicID: 1,
			input: &CreateVaccinationInput{
				MedicalRecordID: ptrUint64(1),
				VaccineID:       1,
				Date:            now,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when vaccine_id is zero",
			clinicID: 1,
			input:    &CreateVaccinationInput{VaccineID: 0, Date: now},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input: &CreateVaccinationInput{
				MedicalRecordID: ptrUint64(1),
				VaccineID:       2,
				Date:            now,
			},
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				createFn: func(_ context.Context, _ *model.Vaccination) error {
					return tt.repoErr
				},
			}
			svc := NewVaccinationService(repo)

			vaccination, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vaccination)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vaccination)
			}
		})
	}
}

func TestVaccinationService_Update(t *testing.T) {
	now := time.Now()
	supplemental := "追記情報"
	tests := []struct {
		name    string
		input   UpdateVaccinationInput
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates vaccination successfully",
			input: UpdateVaccinationInput{
				Date:         &now,
				Supplemental: &supplemental,
			},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateVaccinationInput{},
			repoErr: nil,
			wantErr: true,
			wantNF:  false,
		},
		{
			name: "returns not found error when vaccination does not exist",
			input: UpdateVaccinationInput{
				Supplemental: &supplemental,
			},
			repoErr: apperrors.WrapNotFound("vaccination", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateVaccinationInput{
				Supplemental: &supplemental,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Vaccination, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Vaccination{ID: 1}, nil
				},
			}
			svc := NewVaccinationService(repo)

			vaccination, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vaccination)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vaccination)
			}
		})
	}
}

func TestVaccinationService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes vaccination successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
			wantNF:   false,
		},
		{
			name:     "returns not found error when vaccination does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("vaccination", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoErr:  errors.New("db error"),
			wantErr:  true,
			wantNF:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewVaccinationService(repo)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
