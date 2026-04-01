package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockExamTypeRepository は ExamTypeRepository のテスト用モック実装
type mockExamTypeRepository struct {
	findAllFn                func(ctx context.Context) ([]model.ExaminationType, error)
	findByIDFn               func(ctx context.Context, id uint64) (*model.ExaminationType, error)
	createFn                 func(ctx context.Context, exType *model.ExaminationType) error
	updateFieldsFn           func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error)
	deleteFn                 func(ctx context.Context, id uint64) error
	countUsageByExamTypeIDFn func(ctx context.Context, examTypeID uint64) (int64, error)
}

func (m *mockExamTypeRepository) FindAll(ctx context.Context) ([]model.ExaminationType, error) {
	return m.findAllFn(ctx)
}

func (m *mockExamTypeRepository) FindByID(ctx context.Context, id uint64) (*model.ExaminationType, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockExamTypeRepository) Create(ctx context.Context, exType *model.ExaminationType) error {
	return m.createFn(ctx, exType)
}

func (m *mockExamTypeRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockExamTypeRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockExamTypeRepository) ReplaceItems(ctx context.Context, examTypeID uint64, items []model.ExaminationTypeItem) error {
	return nil
}

func (m *mockExamTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return nil
}

func (m *mockExamTypeRepository) CountUsageByExamTypeID(ctx context.Context, examTypeID uint64) (int64, error) {
	if m.countUsageByExamTypeIDFn == nil {
		return 0, nil
	}
	return m.countUsageByExamTypeIDFn(ctx, examTypeID)
}

func TestExamTypeService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.ExaminationType
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns exam type list",
			repoData: []model.ExaminationType{
				{ID: 1, Name: "血液検査"},
				{ID: 2, Name: "尿検査"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no exam types exist",
			repoData: []model.ExaminationType{},
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
			repo := &mockExamTypeRepository{
				findAllFn: func(_ context.Context) ([]model.ExaminationType, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewExamTypeService(repo)

			examTypes, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, examTypes, tt.wantLen)
			}
		})
	}
}

func TestExamTypeService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoExamType *model.ExaminationType
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "returns exam type when found",
			id:   1,
			repoExamType: &model.ExaminationType{
				ID:   1,
				Name: "血液検査",
				Items: []model.ExaminationTypeItem{
					{ID: 1, ExamTypeID: 1, Name: "白血球数"},
					{ID: 2, ExamTypeID: 1, Name: "赤血球数"},
				},
			},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when exam type does not exist",
			id:           999,
			repoExamType: nil,
			repoErr:      apperrors.WrapNotFound("exam_type", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoExamType: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExamTypeRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.ExaminationType, error) {
					return tt.repoExamType, tt.repoErr
				},
			}
			svc := NewExamTypeService(repo)

			examType, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, examType)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoExamType, examType)
			}
		})
	}
}

func TestExamTypeService_Create(t *testing.T) {
	tests := []struct {
		name    string
		exType  *model.ExaminationType
		repoErr error
		wantErr bool
	}{
		{
			name: "creates exam type successfully",
			exType: &model.ExaminationType{
				Name:     "新規検査種別",
				ClinicID: 1,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates exam type with items successfully",
			exType: &model.ExaminationType{
				Name:     "血液検査",
				ClinicID: 1,
				Items: []model.ExaminationTypeItem{
					{Name: "白血球数", NormalValue: "5000-10000"},
				},
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when exam type already exists",
			exType: &model.ExaminationType{
				Name:     "重複検査種別",
				ClinicID: 1,
			},
			repoErr: apperrors.WrapAlreadyExists("exam_type", "重複検査種別"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			exType: &model.ExaminationType{
				Name:     "エラー検査種別",
				ClinicID: 1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExamTypeRepository{
				createFn: func(_ context.Context, _ *model.ExaminationType) error {
					return tt.repoErr
				},
			}
			svc := NewExamTypeService(repo)

			err := svc.Create(context.Background(), tt.exType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExamTypeService_Update(t *testing.T) {
	name := "更新後検査種別"
	tests := []struct {
		name    string
		input   UpdateExamTypeInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates exam type successfully",
			input: UpdateExamTypeInput{
				Name: &name,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateExamTypeInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns not found error when exam type does not exist",
			input: UpdateExamTypeInput{
				Name: &name,
			},
			repoErr: apperrors.WrapNotFound("exam_type", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateExamTypeInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExamTypeRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ExaminationType, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.ExaminationType{ID: 1}, nil
				},
			}
			svc := NewExamTypeService(repo)

			exType, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, exType)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, exType)
			}
		})
	}
}

func TestExamTypeService_Delete(t *testing.T) {
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
			name:          "deletes exam type successfully when no exams use it",
			id:            1,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       nil,
			wantErr:       false,
		},
		{
			name:          "returns conflict error when exam type is used in exam records",
			id:            1,
			usageCount:    4,
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
			name:          "returns not found error when exam type does not exist",
			id:            999,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       apperrors.WrapNotFound("exam_type", "999"),
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
			repo := &mockExamTypeRepository{
				countUsageByExamTypeIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.usageCount, tt.countUsageErr
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewExamTypeService(repo)

			err := svc.Delete(context.Background(), tt.id)

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
