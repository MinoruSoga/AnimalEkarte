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
	findAllFn            func(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	findByIDFn           func(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	countUsageByCageIDFn func(ctx context.Context, clinicID, id uint64) (int64, error)
	createFn             func(ctx context.Context, cage *model.Cage) error
	updateFieldsFn       func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error)
	deleteFn             func(ctx context.Context, clinicID, id uint64) error
	reorderFn            func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockCageRepository) FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	return m.findAllFn(ctx, clinicID, cageType)
}

func (m *mockCageRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
 if m.findByIDFn != nil {
	return m.findByIDFn(ctx, clinicID, id)
 }
 return nil, nil
}

func (m *mockCageRepository) Create(ctx context.Context, cage *model.Cage) error {
	return m.createFn(ctx, cage)
}

func (m *mockCageRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockCageRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockCageRepository) CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByCageIDFn != nil {
		return m.countUsageByCageIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockCageRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func newTestCageService(repo *mockCageRepository) CageService {
	return NewCageService(repo)
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
				findAllFn: func(_ context.Context, _ uint64, _ *string) ([]model.Cage, error) {
					return tt.repoCages, tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			cages, err := svc.List(context.Background(), 1, tt.cageType)

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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Cage, error) {
					return tt.repoCage, tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			cage, err := svc.GetByID(context.Background(), 1, tt.id)

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
		input   *CreateCageInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates cage successfully",
			input: &CreateCageInput{
				Name:     "新規ケージ",
				CageType: string(model.CageTypeDog),
				CageSize: string(model.CageSizeMedium),
				Price:    &price,
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when cage already exists",
			input: &CreateCageInput{
				Name:     "既存ケージ",
				CageType: string(model.CageTypeCat),
				CageSize: string(model.CageSizeSmall),
			},
			repoErr: apperrors.WrapAlreadyExists("cage", "既存ケージ"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateCageInput{
				Name:     "エラーケージ",
				CageType: string(model.CageTypeGeneral),
				CageSize: string(model.CageSizeLarge),
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
			svc := newTestCageService(repo)

			cage, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cage)
			}
		})
	}
}

func TestCageService_Update(t *testing.T) {
	price := int64(4500)
	name := "更新後ケージ"
	cageType := string(model.CageTypeDog)
	cageSize := string(model.CageSizeLarge)
	isActive := true
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		input    UpdateCageInput
		repoCage *model.Cage
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "updates cage successfully",
			clinicID: 1,
			id:       1,
			input: UpdateCageInput{
				Name:     &name,
				CageType: &cageType,
				CageSize: &cageSize,
				Price:    &price,
				IsActive: &isActive,
			},
			repoCage: &model.Cage{
				ID:       1,
				ClinicID: 1,
				Name:     name,
				CageType: model.CageType(cageType),
				CageSize: model.CageSize(cageSize),
				Price:    &price,
				IsActive: isActive,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when no fields provided",
			clinicID: 1,
			id:       1,
			input:    UpdateCageInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       999,
			input:    UpdateCageInput{Name: &name},
			repoCage: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Cage, error) {
					return &model.Cage{ID: tt.id, ClinicID: tt.clinicID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Cage, error) {
					return tt.repoCage, tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			cage, err := svc.Update(context.Background(), tt.clinicID, tt.id, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cage)
			}
		})
	}
}

func TestCageService_Update_NilInput(t *testing.T) {
	repo := &mockCageRepository{}
	svc := newTestCageService(repo)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCageService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		usageCount   int64
		findByIDErr  error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:    "deletes cage successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns not found error when cage does not exist",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("cage", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:         "使用中のケージは削除できない",
			id:           2,
			usageCount:   1,
			wantErr:      true,
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Cage{ID: id}, nil
				},
				countUsageByCageIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			err := svc.Delete(context.Background(), 1, tt.id)

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
