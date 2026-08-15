package billing

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockInsuranceRepository は InsuranceRepository のテスト用モック実装
type mockInsuranceRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
	createFn                  func(ctx context.Context, insurance *model.Insurance) error
	updateFn                  func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderErr                error
	countUsageByInsuranceIDFn func(ctx context.Context, clinicID, insuranceID uint64) (int64, error)
}

func (m *mockInsuranceRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockInsuranceRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockInsuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	return m.createFn(ctx, insurance)
}

func (m *mockInsuranceRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error) {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockInsuranceRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockInsuranceRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockInsuranceRepository) CountUsageByInsuranceID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByInsuranceIDFn != nil {
		return m.countUsageByInsuranceIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

func TestInsuranceService_List(t *testing.T) {
	tests := []struct {
		name           string
		repoInsurances []model.Insurance
		repoErr        error
		wantLen        int
		wantErr        bool
	}{
		{
			name: "returns insurance list",
			repoInsurances: []model.Insurance{
				{ID: 1, Name: "アニコム損保", IsActive: true},
				{ID: 2, Name: "アイペット損保", IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:           "returns empty list when no insurances exist",
			repoInsurances: []model.Insurance{},
			repoErr:        nil,
			wantLen:        0,
			wantErr:        false,
		},
		{
			name:           "propagates repository error",
			repoInsurances: nil,
			repoErr:        errors.New("db connection error"),
			wantLen:        0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Insurance, error) {
					return tt.repoInsurances, tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

			insurances, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, insurances, tt.wantLen)
			}
		})
	}
}

func TestInsuranceService_GetByID(t *testing.T) {
	coverageRate := 70
	tests := []struct {
		name          string
		id            uint64
		repoInsurance *model.Insurance
		repoErr       error
		wantErr       bool
		wantNF        bool
	}{
		{
			name: "returns insurance when found",
			id:   1,
			repoInsurance: &model.Insurance{
				ID:           1,
				Name:         "アニコム損保",
				IsActive:     true,
				CoverageRate: coverageRate,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:          "returns not found error when insurance does not exist",
			id:            999,
			repoInsurance: nil,
			repoErr:       apperrors.WrapNotFound("insurance", "999"),
			wantErr:       true,
			wantNF:        true,
		},
		{
			name:          "returns error on repository failure",
			id:            1,
			repoInsurance: nil,
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Insurance, error) {
					return tt.repoInsurance, tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

			insurance, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoInsurance, insurance)
			}
		})
	}
}

func TestInsuranceService_Create(t *testing.T) {
	rate70 := 70
	rateNeg := -1
	rate101 := 101
	tests := []struct {
		name    string
		input   *CreateInsuranceInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates insurance successfully",
			input: &CreateInsuranceInput{
				Name:         "新規保険",
				IsActive:     true,
				CoverageRate: &rate70,
				ContactPhone: "03-1234-5678",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates insurance successfully when coverage_rate is nil (default 0)",
			input: &CreateInsuranceInput{
				Name:         "デフォルト保険",
				IsActive:     true,
				CoverageRate: nil,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when coverage_rate is negative (BUG-398)",
			input: &CreateInsuranceInput{
				Name:         "負数保険",
				IsActive:     true,
				CoverageRate: &rateNeg,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when coverage_rate exceeds 100 (BUG-398)",
			input: &CreateInsuranceInput{
				Name:         "超過保険",
				IsActive:     true,
				CoverageRate: &rate101,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when insurance already exists",
			input: &CreateInsuranceInput{
				Name:     "既存保険",
				IsActive: true,
			},
			repoErr: apperrors.WrapAlreadyExists("insurance", "既存保険"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateInsuranceInput{
				Name:     "エラー保険",
				IsActive: true,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				createFn: func(_ context.Context, _ *model.Insurance) error {
					return tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

			insurance, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, insurance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, insurance)
			}
		})
	}
}

func TestInsuranceService_Update(t *testing.T) {
	name := "更新後保険"
	isActive := true
	coverageRate := 80
	rateNeg := -5
	rate200 := 200
	tests := []struct {
		name        string
		input       UpdateInsuranceInput
		repoErr     error
		findByIDErr error
		wantErr     bool
	}{
		{
			name: "updates insurance successfully",
			input: UpdateInsuranceInput{
				Name:         &name,
				IsActive:     &isActive,
				CoverageRate: &coverageRate,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateInsuranceInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when coverage_rate is negative (BUG-398)",
			input: UpdateInsuranceInput{
				CoverageRate: &rateNeg,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when coverage_rate exceeds 100 (BUG-398)",
			input: UpdateInsuranceInput{
				CoverageRate: &rate200,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateInsuranceInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "returns error when insurance not found",
			input: UpdateInsuranceInput{
				Name: &name,
			},
			findByIDErr: apperrors.WrapNotFound("insurance", "999"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Insurance, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Insurance{ID: 1, ClinicID: 1}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Insurance, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Insurance{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewInsuranceService(repo)

			insurance, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, insurance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, insurance)
			}
		})
	}
}

func TestInsuranceService_Update_NilInput(t *testing.T) {
	repo := &mockInsuranceRepository{}
	svc := NewInsuranceService(repo)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestInsuranceService_Reorder(t *testing.T) {
	tests := []struct {
		name    string
		ids     []uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{3, 1, 2},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "returns error when ids is empty",
			ids:     []uint64{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{reorderErr: tt.repoErr}
			svc := NewInsuranceService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInsuranceService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		petCount     int64
		countErr     error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:    "deletes insurance successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when insurance does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("insurance", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:         "使用中の保険は削除できない",
			id:           2,
			petCount:     2,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:     "returns error when usage count check fails",
			id:       3,
			countErr: errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Insurance, error) {
					if tt.wantNF {
						return nil, tt.repoErr
					}
					return &model.Insurance{ID: id}, nil
				},
				countUsageByInsuranceIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.petCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					if tt.wantNF {
						return nil
					}
					return tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

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

// TestBuildInsuranceUpdate は buildInsuranceUpdate の全フィールド網羅とゼロ値挙動を検証する。
func TestBuildInsuranceUpdate(t *testing.T) {
	name := "新保険名"
	isActive := true
	description := "説明文"
	coverageRate := 60
	contactPhone := "03-0000-0000"
	sortOrder := 5

	t.Run("maps all provided fields", func(t *testing.T) {
		input := &UpdateInsuranceInput{
			Name:         &name,
			IsActive:     &isActive,
			Description:  &description,
			CoverageRate: &coverageRate,
			ContactPhone: &contactPhone,
			SortOrder:    &sortOrder,
		}
		fields := buildInsuranceUpdate(input)
		assert.Equal(t, name, fields[colInsuranceName])
		assert.Equal(t, isActive, fields[colInsuranceIsActive])
		assert.Equal(t, description, fields[colInsuranceDescription])
		assert.Equal(t, coverageRate, fields[colInsuranceCoverageRate])
		assert.Equal(t, contactPhone, fields[colInsuranceContactPhone])
		assert.Equal(t, sortOrder, fields[colInsuranceSortOrder])
		assert.Len(t, fields, 6)
	})

	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		fields := buildInsuranceUpdate(&UpdateInsuranceInput{})
		assert.Empty(t, fields)
	})
}
