package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockCheckupTypeRepository は CheckupTypeRepository のテスト用モック実装
type mockCheckupTypeRepository struct {
	findAllFn                   func(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	findByIDFn                  func(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	createFn                    func(ctx context.Context, checkupType *model.CheckupType) error
	updateFieldsFn              func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error)
	deleteFn                    func(ctx context.Context, clinicID, id uint64) error
	countUsageByCheckupTypeIDFn func(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error)
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

func (m *mockCheckupTypeRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockCheckupTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockCheckupTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return nil
}

func (m *mockCheckupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
	if m.countUsageByCheckupTypeIDFn == nil {
		return 0, nil
	}
	return m.countUsageByCheckupTypeIDFn(ctx, clinicID, checkupTypeID)
}

func (m *mockCheckupTypeRepository) CountChildrenByParentID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
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
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		input    *UpdateCheckupTypeInput
		repoData *model.CheckupType
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "updates checkup type successfully",
			clinicID: 1,
			id:       1,
			input:    &UpdateCheckupTypeInput{Name: &name, IsActive: &isActive},
			repoData: &model.CheckupType{ID: 1, Name: name, IsActive: isActive},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns error when no fields provided",
			clinicID: 1,
			id:       1,
			input:    &UpdateCheckupTypeInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns not found error when checkup type does not exist",
			clinicID: 1,
			id:       999,
			input:    &UpdateCheckupTypeInput{Name: &name},
			repoData: nil,
			repoErr:  apperrors.WrapNotFound("checkup_type", "999"),
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input:    &UpdateCheckupTypeInput{Name: &name},
			repoData: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.CheckupType, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			result, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
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
		usageCount    int64
		countUsageErr error
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
			name:          "returns not found error when checkup type does not exist",
			id:            999,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       apperrors.WrapNotFound("checkup_type", "999"),
			wantErr:       true,
			wantNF:        true,
		},
		{
			name:          "returns error on repository failure",
			id:            1,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
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
