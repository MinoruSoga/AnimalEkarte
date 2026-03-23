package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockCageRepository は CageRepository のテスト用モック実装
type mockCageRepository struct {
	findAllFn  func(ctx context.Context, cageType *string) ([]model.Cage, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.Cage, error)
	createFn   func(ctx context.Context, cage *model.Cage) error
	updateFn   func(ctx context.Context, cage *model.Cage) error
	deleteFn   func(ctx context.Context, id uint64) error
	reorderFn  func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockCageRepository) FindAll(ctx context.Context, cageType *string) ([]model.Cage, error) {
	return m.findAllFn(ctx, cageType)
}

func (m *mockCageRepository) FindByID(ctx context.Context, id uint64) (*model.Cage, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockCageRepository) Create(ctx context.Context, cage *model.Cage) error {
	return m.createFn(ctx, cage)
}

func (m *mockCageRepository) Update(ctx context.Context, cage *model.Cage) error {
	return m.updateFn(ctx, cage)
}

func (m *mockCageRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockCageRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func TestCageService_List(t *testing.T) {
	cageType := string(model.CageTypeDog)

	tests := []struct {
		name      string
		cageType  *string
		repoCages []model.Cage
		repoErr   error
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns all cage list",
			repoCages: []model.Cage{
				{ID: 1, ClinicID: 1, Name: "犬用ケージA", CageType: model.CageTypeDog, CageSize: model.CageSizeSmall, IsActive: true},
				{ID: 2, ClinicID: 1, Name: "猫用ケージA", CageType: model.CageTypeCat, CageSize: model.CageSizeMedium, IsActive: true},
				{ID: 3, ClinicID: 1, Name: "ICU-1", CageType: model.CageTypeICU, CageSize: model.CageSizeLarge, IsActive: true},
			},
			repoErr: nil,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:      "returns empty list when no cages exist",
			repoCages: []model.Cage{},
			repoErr:   nil,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:     "filters by cage_type",
			cageType: &cageType,
			repoCages: []model.Cage{
				{ID: 1, ClinicID: 1, Name: "犬用ケージA", CageType: model.CageTypeDog, CageSize: model.CageSizeSmall},
			},
			repoErr: nil,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:      "propagates repository error",
			repoCages: nil,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				findAllFn: func(_ context.Context, _ *string) ([]model.Cage, error) {
					return tt.repoCages, tt.repoErr
				},
			}
			svc := NewCageService(repo)

			cages, err := svc.List(context.Background(), tt.cageType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, cages, tt.wantLen)
			}
		})
	}
}

func TestCageService_GetByID(t *testing.T) {
	price := int64(5000)
	tests := []struct {
		name     string
		id       uint64
		repoCage *model.Cage
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name: "returns cage when found",
			id:   1,
			repoCage: &model.Cage{
				ID:       1,
				ClinicID: 1,
				Name:     "犬用ケージA",
				CageType: model.CageTypeDog,
				CageSize: model.CageSizeSmall,
				Price:    &price,
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns not found error when cage does not exist",
			id:       999,
			repoCage: nil,
			repoErr:  apperrors.WrapNotFound("cage", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			id:       1,
			repoCage: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Cage, error) {
					return tt.repoCage, tt.repoErr
				},
			}
			svc := NewCageService(repo)

			cage, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoCage, cage)
			}
		})
	}
}

func TestCageService_Create(t *testing.T) {
	price := int64(3000)
	tests := []struct {
		name    string
		cage    *model.Cage
		repoErr error
		wantErr bool
	}{
		{
			name: "creates cage successfully",
			cage: &model.Cage{
				ClinicID: 1,
				Name:     "新規ケージ",
				CageType: model.CageTypeDog,
				CageSize: model.CageSizeMedium,
				Price:    &price,
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when cage already exists",
			cage: &model.Cage{
				ClinicID: 1,
				Name:     "既存ケージ",
				CageType: model.CageTypeCat,
				CageSize: model.CageSizeSmall,
			},
			repoErr: apperrors.WrapAlreadyExists("cage", "既存ケージ"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			cage: &model.Cage{
				ClinicID: 1,
				Name:     "エラーケージ",
				CageType: model.CageTypeGeneral,
				CageSize: model.CageSizeLarge,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				createFn: func(_ context.Context, _ *model.Cage) error {
					return tt.repoErr
				},
			}
			svc := NewCageService(repo)

			err := svc.Create(context.Background(), tt.cage)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCageService_Update(t *testing.T) {
	price := int64(4500)
	tests := []struct {
		name    string
		cage    *model.Cage
		repoErr error
		wantErr bool
	}{
		{
			name: "updates cage successfully",
			cage: &model.Cage{
				ID:       1,
				ClinicID: 1,
				Name:     "更新後ケージ",
				CageType: model.CageTypeDog,
				CageSize: model.CageSizeLarge,
				Price:    &price,
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error on repository failure",
			cage: &model.Cage{
				ID:       999,
				ClinicID: 1,
				Name:     "エラーケージ",
				CageType: model.CageTypeCat,
				CageSize: model.CageSizeSmall,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				updateFn: func(_ context.Context, _ *model.Cage) error {
					return tt.repoErr
				},
			}
			svc := NewCageService(repo)

			err := svc.Update(context.Background(), tt.cage)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCageService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes cage successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when cage does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("cage", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewCageService(repo)

			err := svc.Delete(context.Background(), tt.id)

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
