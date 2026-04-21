package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockHospitalizationRepository は HospitalizationRepository のテスト用モック実装
type mockHospitalizationRepository struct {
	findAllFn                               func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	findByIDFn                              func(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	createFn                                func(ctx context.Context, hospitalization *model.Hospitalization) error
	updateFieldsFn                          func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error)
	deleteFn                                func(ctx context.Context, clinicID, id uint64) error
	countCarePlanItemsByHospitalizationIDFn func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
}

func (m *mockHospitalizationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockHospitalizationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockHospitalizationRepository) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	return m.createFn(ctx, hospitalization)
}

func (m *mockHospitalizationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockHospitalizationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockHospitalizationRepository) ExistsByCageID(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockHospitalizationRepository) CountCarePlanItemsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	if m.countCarePlanItemsByHospitalizationIDFn == nil {
		return 0, nil
	}
	return m.countCarePlanItemsByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockHospitalizationRepository) CountDailyRecordsByHospitalizationID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockHospitalizationRepository) CountTreatmentPlansByHospitalizationID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func TestHospitalizationService_List(t *testing.T) {
	petID := uint64(5)
	ownerID := uint64(2)
	status := string(model.HospitalizationStatusAdmitted)

	tests := []struct {
		name      string
		clinicID  uint64
		petID     *uint64
		ownerID   *uint64
		status    *string
		page      int
		limit     int
		repoItems []model.Hospitalization
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:     "returns hospitalization list with total count",
			clinicID: 1,
			page:     1,
			limit:    20,
			repoItems: []model.Hospitalization{
				{ID: 1, ClinicID: 1, PetID: 5, OwnerID: 2, Status: model.HospitalizationStatusAdmitted},
				{ID: 2, ClinicID: 1, PetID: 6, OwnerID: 3, Status: model.HospitalizationStatusReserved},
			},
			repoTotal: 2,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no hospitalizations exist",
			clinicID:  1,
			page:      1,
			limit:     20,
			repoItems: []model.Hospitalization{},
			repoTotal: 0,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			petID:    &petID,
			page:     1,
			limit:    20,
			repoItems: []model.Hospitalization{
				{ID: 1, ClinicID: 1, PetID: 5, OwnerID: 2},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			ownerID:  &ownerID,
			page:     1,
			limit:    20,
			repoItems: []model.Hospitalization{
				{ID: 1, ClinicID: 1, PetID: 5, OwnerID: 2},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by status",
			clinicID: 1,
			status:   &status,
			page:     1,
			limit:    20,
			repoItems: []model.Hospitalization{
				{ID: 1, ClinicID: 1, Status: model.HospitalizationStatusAdmitted},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			clinicID:  1,
			page:      1,
			limit:     20,
			repoItems: nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _, _ *string, _, _ int) ([]model.Hospitalization, int64, error) {
					return tt.repoItems, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewHospitalizationService(&repository.Repositories{Hospitalization: repo})

			items, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, tt.status, nil, nil, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, items, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestHospitalizationService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoItem *model.Hospitalization
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "returns hospitalization when found",
			clinicID: 1,
			id:       10,
			repoItem: &model.Hospitalization{
				ID:        10,
				ClinicID:  1,
				PetID:     5,
				OwnerID:   2,
				Status:    model.HospitalizationStatusAdmitted,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns not found error when hospitalization does not exist",
			clinicID: 1,
			id:       999,
			repoItem: nil,
			repoErr:  apperrors.WrapNotFound("hospitalization", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoItem: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return tt.repoItem, tt.repoErr
				},
			}
			svc := NewHospitalizationService(&repository.Repositories{Hospitalization: repo})

			item, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoItem, item)
			}
		})
	}
}

func TestHospitalizationService_Create(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		clinicID uint64
		input    *CreateHospitalizationInput
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "creates hospitalization successfully",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				OwnerID:             2,
				PetID:               5,
				HospitalizationType: model.HospitalizationTypeInpatient,
				StartDate:           now,
				EndDate:             now.Add(24 * time.Hour),
				Status:              model.HospitalizationStatusReserved,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "defaults status to reserved when empty",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				OwnerID:             2,
				PetID:               5,
				HospitalizationType: model.HospitalizationTypeHotel,
				StartDate:           now,
				EndDate:             now.Add(24 * time.Hour),
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when already exists",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				OwnerID:   2,
				PetID:     5,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			repoErr: apperrors.WrapAlreadyExists("hospitalization", now.String()),
			wantErr: true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				OwnerID: 2,
				PetID:   5,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				createFn: func(_ context.Context, _ *model.Hospitalization) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationService(&repository.Repositories{Hospitalization: repo})

			hosp, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, hosp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hosp)
			}
		})
	}
}

func TestHospitalizationService_Update(t *testing.T) {
	now := time.Now()
	statusAdmitted := model.HospitalizationStatusAdmitted
	tests := []struct {
		name    string
		input   UpdateHospitalizationInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates hospitalization successfully",
			input: UpdateHospitalizationInput{
				Status:    &statusAdmitted,
				StartDate: &now,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateHospitalizationInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   UpdateHospitalizationInput{Status: &statusAdmitted},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Hospitalization{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewHospitalizationService(&repository.Repositories{Hospitalization: repo})

			hosp, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, hosp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hosp)
			}
		})
	}
}

func TestHospitalizationService_Delete(t *testing.T) {
	tests := []struct {
		name                 string
		clinicID             uint64
		id                   uint64
		carePlanItemCount    int64
		countCarePlanItemErr error
		repoErr              error
		wantErr              bool
		wantNF               bool
		wantConflict         bool
	}{
		{
			name:                 "deletes hospitalization successfully when no care plan items exist",
			clinicID:             1,
			id:                   10,
			carePlanItemCount:    0,
			countCarePlanItemErr: nil,
			repoErr:              nil,
			wantErr:              false,
		},
		{
			name:                 "returns conflict error when hospitalization has care plan items",
			clinicID:             1,
			id:                   10,
			carePlanItemCount:    5,
			countCarePlanItemErr: nil,
			repoErr:              nil,
			wantErr:              true,
			wantConflict:         true,
		},
		{
			name:                 "returns error when care plan item count check fails",
			clinicID:             1,
			id:                   10,
			carePlanItemCount:    0,
			countCarePlanItemErr: errors.New("db error"),
			repoErr:              nil,
			wantErr:              true,
		},
		{
			name:                 "returns not found error when hospitalization does not exist",
			clinicID:             1,
			id:                   999,
			carePlanItemCount:    0,
			countCarePlanItemErr: nil,
			repoErr:              apperrors.WrapNotFound("hospitalization", "999"),
			wantErr:              true,
			wantNF:               true,
		},
		{
			name:                 "returns error on repository failure",
			clinicID:             1,
			id:                   10,
			carePlanItemCount:    0,
			countCarePlanItemErr: nil,
			repoErr:              errors.New("db error"),
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				countCarePlanItemsByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.carePlanItemCount, tt.countCarePlanItemErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationService(&repository.Repositories{Hospitalization: repo})

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
