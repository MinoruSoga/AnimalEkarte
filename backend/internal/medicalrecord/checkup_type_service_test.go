package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockCheckupTypeRepository は CheckupTypeRepository のテスト用モック実装
type mockCheckupTypeRepository struct {
	findAllFn                   func(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	findByIDFn                  func(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	createFn                    func(ctx context.Context, checkupType *model.CheckupType) error
	updateFieldsFn              func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error)
	deleteFn                    func(ctx context.Context, clinicID, id uint64) error
	reorderFn                   func(ctx context.Context, clinicID uint64, ids []uint64) error
	countUsageByCheckupTypeIDFn func(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error)
	countChildrenByParentIDFn   func(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

func (m *mockCheckupTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockCheckupTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockCheckupTypeRepository) Create(ctx context.Context, checkupType *model.CheckupType) error {
	return m.createFn(ctx, checkupType)
}

func (m *mockCheckupTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockCheckupTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockCheckupTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func (m *mockCheckupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
	if m.countUsageByCheckupTypeIDFn == nil {
		return 0, nil
	}
	return m.countUsageByCheckupTypeIDFn(ctx, clinicID, checkupTypeID)
}

func (m *mockCheckupTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

func TestCheckupTypeService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.CheckupType
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns checkup type list",
			repoData: []model.CheckupType{
				{ID: 1, Name: "定期健診"},
				{ID: 2, Name: "ワクチン前健診"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no checkup types exist",
			repoData: []model.CheckupType{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			repoData: nil,
			repoErr:  errors.New("db connection error"),
			wantLen:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.CheckupType, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			checkupTypes, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, checkupTypes, tt.wantLen)
			}
		})
	}
}

func TestCheckupTypeService_GetByID(t *testing.T) {
	tests := []struct {
		name            string
		id              uint64
		repoCheckupType *model.CheckupType
		repoErr         error
		wantErr         bool
		wantNotFound    bool
	}{
		{
			name:            "returns checkup type when found",
			id:              1,
			repoCheckupType: &model.CheckupType{ID: 1, Name: "定期健診"},
			repoErr:         nil,
			wantErr:         false,
			wantNotFound:    false,
		},
		{
			name:            "returns not found error when checkup type does not exist",
			id:              999,
			repoCheckupType: nil,
			repoErr:         apperrors.WrapNotFound("checkup_type", "999"),
			wantErr:         true,
			wantNotFound:    true,
		},
		{
			name:            "returns error on repository failure",
			id:              1,
			repoCheckupType: nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
			wantNotFound:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
					return tt.repoCheckupType, tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			checkupType, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, checkupType)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoCheckupType, checkupType)
			}
		})
	}
}

func TestCheckupTypeService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateCheckupTypeInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates checkup type successfully",
			input: &CreateCheckupTypeInput{
				Name:     "新規健診種別",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates checkup type with optional fields",
			input: &CreateCheckupTypeInput{
				Name:      "定期健診",
				IsActive:  true,
				Interval:  "6ヶ月",
				TargetAge: "成犬",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when checkup type already exists",
			input: &CreateCheckupTypeInput{
				Name:     "重複健診種別",
				IsActive: true,
			},
			repoErr: apperrors.WrapAlreadyExists("checkup_type", "重複健診種別"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateCheckupTypeInput{
				Name:     "エラー健診種別",
				IsActive: true,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "rejects empty name",
			input: &CreateCheckupTypeInput{
				Name:     "",
				IsActive: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				createFn: func(_ context.Context, _ *model.CheckupType) error {
					return tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			result, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				assert.Equal(t, uint64(1), result.ClinicID)
			}
		})
	}
}

func TestCheckupTypeService_Update(t *testing.T) {
	name := "更新後健診種別"
	isActive := true
	emptyName := "   "
	tests := []struct {
		name        string
		clinicID    uint64
		id          uint64
		input       *UpdateCheckupTypeInput
		findByIDErr error
		repoData    *model.CheckupType
		updateErr   error
		wantErr     bool
		wantNF      bool
	}{
		{
			name:     "updates checkup type successfully",
			clinicID: 1,
			id:       1,
			input:    &UpdateCheckupTypeInput{Name: &name, IsActive: &isActive},
			repoData: &model.CheckupType{ID: 1, Name: name, IsActive: isActive},
			wantErr:  false,
		},
		{
			name:     "returns error when no fields provided",
			clinicID: 1,
			id:       1,
			input:    &UpdateCheckupTypeInput{},
			wantErr:  true,
		},
		{
			name:        "returns not found error when checkup type does not exist",
			clinicID:    1,
			id:          999,
			input:       &UpdateCheckupTypeInput{Name: &name},
			findByIDErr: apperrors.WrapNotFound("checkup_type", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:      "returns error on repository failure",
			clinicID:  1,
			id:        1,
			input:     &UpdateCheckupTypeInput{Name: &name},
			updateErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:     "returns invalid input error when nil input given",
			clinicID: 1,
			id:       1,
			input:    nil,
			wantErr:  true,
		},
		{
			name:     "rejects blank name",
			clinicID: 1,
			id:       1,
			input:    &UpdateCheckupTypeInput{Name: &emptyName},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.CheckupType{ID: tt.id}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.CheckupType, error) {
					return tt.repoData, tt.updateErr
				},
			}
			svc := NewCheckupTypeService(repo)

			result, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestCheckupTypeService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		childCount    int64
		countChildErr error
		usageCount    int64
		countUsageErr error
		findByIDErr   error
		repoErr       error
		wantErr       bool
		wantNF        bool
		wantConflict  bool
	}{
		{
			name:          "deletes checkup type successfully when no checkups use it",
			id:            1,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       nil,
			wantErr:       false,
		},
		{
			name:          "returns conflict error when checkup type is used in checkup records",
			id:            1,
			usageCount:    3,
			countUsageErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns error when usage count check fails",
			id:            1,
			usageCount:    0,
			countUsageErr: errors.New("db error"),
			repoErr:       nil,
			wantErr:       true,
		},
		{
			name:        "returns not found error when checkup type does not exist",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("checkup_type", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:          "returns error on repository failure",
			id:            1,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
		{
			name:         "returns conflict error when checkup type has sub types",
			id:           1,
			childCount:   2,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:          "returns error when child count check fails",
			id:            1,
			countChildErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.CheckupType{ID: id}, nil
				},
				countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.childCount, tt.countChildErr
				},
				countUsageByCheckupTypeIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, tt.countUsageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

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

func TestCheckupTypeService_Reorder(t *testing.T) {
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
			name:    "returns invalid input when ids is empty",
			ids:     []uint64{},
			wantErr: true,
		},
		{
			name:    "returns error when id not in clinic",
			ids:     []uint64{999},
			repoErr: apperrors.WrapInvalidInput("checkup_type id 999 not found in this clinic"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildCheckupTypeUpdate(t *testing.T) {
	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		fields := buildCheckupTypeUpdate(&UpdateCheckupTypeInput{})
		assert.Empty(t, fields)
	})

	t.Run("includes all provided fields", func(t *testing.T) {
		name := "更新後健診種別"
		price := int64(3000)
		isActive := true
		desc := "説明"
		interval := "6ヶ月"
		targetAge := "成犬"
		parentID := uint64(2)
		sortOrder := 5

		input := &UpdateCheckupTypeInput{
			Name:        &name,
			Price:       &price,
			IsActive:    &isActive,
			Description: &desc,
			Interval:    &interval,
			TargetAge:   &targetAge,
			ParentID:    &parentID,
			SortOrder:   &sortOrder,
		}
		fields := buildCheckupTypeUpdate(input)

		assert.Equal(t, name, fields[colCheckupTypeName])
		assert.Equal(t, price, fields[colCheckupTypePrice])
		assert.Equal(t, isActive, fields[colCheckupTypeIsActive])
		assert.Equal(t, desc, fields[colCheckupTypeDescription])
		assert.Equal(t, interval, fields[colCheckupTypeInterval])
		assert.Equal(t, targetAge, fields[colCheckupTypeTargetAge])
		assert.Equal(t, parentID, fields[colCheckupTypeParentID])
		assert.Equal(t, sortOrder, fields[colCheckupTypeSortOrder])
	})

	t.Run("clears parent_id when ClearParentID is true", func(t *testing.T) {
		input := &UpdateCheckupTypeInput{ClearParentID: true}
		fields := buildCheckupTypeUpdate(input)
		assert.Contains(t, fields, colCheckupTypeParentID)
		assert.Nil(t, fields[colCheckupTypeParentID])
	})

	t.Run("omits parent_id when neither ParentID nor ClearParentID set", func(t *testing.T) {
		input := &UpdateCheckupTypeInput{}
		fields := buildCheckupTypeUpdate(input)
		assert.NotContains(t, fields, colCheckupTypeParentID)
	})
}
