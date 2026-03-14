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
	findAllFn  func(ctx context.Context) ([]model.ExamType, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.ExamType, error)
	createFn   func(ctx context.Context, exType *model.ExamType) error
	updateFn   func(ctx context.Context, exType *model.ExamType) error
	deleteFn   func(ctx context.Context, id uint64) error
}

func (m *mockExamTypeRepository) FindAll(ctx context.Context) ([]model.ExamType, error) {
	return m.findAllFn(ctx)
}

func (m *mockExamTypeRepository) FindByID(ctx context.Context, id uint64) (*model.ExamType, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockExamTypeRepository) Create(ctx context.Context, exType *model.ExamType) error {
	return m.createFn(ctx, exType)
}

func (m *mockExamTypeRepository) Update(ctx context.Context, exType *model.ExamType) error {
	return m.updateFn(ctx, exType)
}

func (m *mockExamTypeRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockExamTypeRepository) ReplaceItems(ctx context.Context, examTypeID uint64, items []model.ExamTypeItem) error {
	return nil
}

func TestExamTypeService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.ExamType
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns exam type list",
			repoData: []model.ExamType{
				{ID: 1, Name: "血液検査"},
				{ID: 2, Name: "尿検査"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no exam types exist",
			repoData: []model.ExamType{},
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
				findAllFn: func(_ context.Context) ([]model.ExamType, error) {
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
		repoExamType *model.ExamType
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "returns exam type when found",
			id:   1,
			repoExamType: &model.ExamType{
				ID:   1,
				Name: "血液検査",
				Items: []model.ExamTypeItem{
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
				findByIDFn: func(_ context.Context, _ uint64) (*model.ExamType, error) {
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
		exType  *model.ExamType
		repoErr error
		wantErr bool
	}{
		{
			name: "creates exam type successfully",
			exType: &model.ExamType{
				Name:     "新規検査種別",
				ClinicID: 1,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates exam type with items successfully",
			exType: &model.ExamType{
				Name:     "血液検査",
				ClinicID: 1,
				Items: []model.ExamTypeItem{
					{Name: "白血球数", NormalValue: "5000-10000"},
				},
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when exam type already exists",
			exType: &model.ExamType{
				Name:     "重複検査種別",
				ClinicID: 1,
			},
			repoErr: apperrors.WrapAlreadyExists("exam_type", "重複検査種別"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			exType: &model.ExamType{
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
				createFn: func(_ context.Context, _ *model.ExamType) error {
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
	tests := []struct {
		name    string
		exType  *model.ExamType
		repoErr error
		wantErr bool
	}{
		{
			name: "updates exam type successfully",
			exType: &model.ExamType{
				ID:   1,
				Name: "更新後検査種別",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns not found error when exam type does not exist",
			exType: &model.ExamType{
				ID:   999,
				Name: "存在しない検査種別",
			},
			repoErr: apperrors.WrapNotFound("exam_type", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			exType: &model.ExamType{
				ID:   1,
				Name: "エラーケース",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExamTypeRepository{
				updateFn: func(_ context.Context, _ *model.ExamType) error {
					return tt.repoErr
				},
			}
			svc := NewExamTypeService(repo)

			err := svc.Update(context.Background(), tt.exType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExamTypeService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes exam type successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns not found error when exam type does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("exam_type", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExamTypeRepository{
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
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
