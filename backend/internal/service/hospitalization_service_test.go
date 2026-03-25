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

// mockHospitalizationRepository は HospitalizationRepository のテスト用モック実装
type mockHospitalizationRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	createFn   func(ctx context.Context, hospitalization *model.Hospitalization) error
	updateFn   func(ctx context.Context, hospitalization *model.Hospitalization) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockHospitalizationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockHospitalizationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockHospitalizationRepository) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	return m.createFn(ctx, hospitalization)
}

func (m *mockHospitalizationRepository) Update(ctx context.Context, hospitalization *model.Hospitalization) error {
	return m.updateFn(ctx, hospitalization)
}

func (m *mockHospitalizationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
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
			svc := NewHospitalizationService(repo)

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
			svc := NewHospitalizationService(repo)

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
		name            string
		hospitalization *model.Hospitalization
		repoErr         error
		wantErr         bool
	}{
		{
			name: "creates hospitalization successfully",
			hospitalization: &model.Hospitalization{
				ClinicID:            1,
				PetID:               5,
				OwnerID:             2,
				HospitalizationType: model.HospitalizationTypeInpatient,
				StartDate:           now,
				EndDate:             now.Add(24 * time.Hour),
				Status:              model.HospitalizationStatusReserved,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when already exists",
			hospitalization: &model.Hospitalization{
				ClinicID:  1,
				PetID:     5,
				OwnerID:   2,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			repoErr: apperrors.WrapAlreadyExists("hospitalization", now.String()),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			hospitalization: &model.Hospitalization{
				ClinicID: 1,
				PetID:    5,
				OwnerID:  2,
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
			svc := NewHospitalizationService(repo)

			err := svc.Create(context.Background(), tt.hospitalization)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHospitalizationService_Update(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		hospitalization *model.Hospitalization
		repoErr         error
		wantErr         bool
	}{
		{
			name: "updates hospitalization successfully",
			hospitalization: &model.Hospitalization{
				ID:        1,
				ClinicID:  1,
				PetID:     5,
				OwnerID:   2,
				Status:    model.HospitalizationStatusAdmitted,
				StartDate: now,
				EndDate:   now.Add(48 * time.Hour),
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error on repository failure",
			hospitalization: &model.Hospitalization{
				ID:       999,
				ClinicID: 1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				updateFn: func(_ context.Context, _ *model.Hospitalization) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationService(repo)

			err := svc.Update(context.Background(), tt.hospitalization)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHospitalizationService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes hospitalization successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns not found error when hospitalization does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("hospitalization", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationService(repo)

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
