package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockTrimmingCourseTypeRepository は TrimmingCourseTypeRepository のテスト用モック (#73)
type mockTrimmingCourseTypeRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error)
	createFn       func(ctx context.Context, m *model.TrimmingCourseType) (*model.TrimmingCourseType, error)
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourseType, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
	countUsageFn   func(ctx context.Context, clinicID, id uint64) (int64, error)
	reorderFn      func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockTrimmingCourseTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockTrimmingCourseTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockTrimmingCourseTypeRepository) Create(ctx context.Context, t *model.TrimmingCourseType) (*model.TrimmingCourseType, error) {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	return t, nil
}

func (m *mockTrimmingCourseTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourseType, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockTrimmingCourseTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockTrimmingCourseTypeRepository) CountUsageByCourseTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageFn != nil {
		return m.countUsageFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockTrimmingCourseTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- テスト ----

func TestTrimmingCourseTypeService_Create(t *testing.T) {
	created := &model.TrimmingCourseType{ID: 1, ClinicID: 1, Name: "シャンプー", SortOrder: 1}

	t.Run("正常: 作成済みレコードを返す", func(t *testing.T) {
		repo := &mockTrimmingCourseTypeRepository{
			createFn: func(_ context.Context, _ *model.TrimmingCourseType) (*model.TrimmingCourseType, error) {
				return created, nil
			},
		}
		svc := NewTrimmingCourseTypeService(repo)
		got, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseTypeInput{Name: "シャンプー", SortOrder: 1})
		assert.NoError(t, err)
		assert.Equal(t, created, got)
	})

	t.Run("エラー: 名前が空", func(t *testing.T) {
		svc := NewTrimmingCourseTypeService(&mockTrimmingCourseTypeRepository{})
		got, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseTypeInput{Name: ""})
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestTrimmingCourseTypeService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		countUsageFn func(ctx context.Context, clinicID, id uint64) (int64, error)
		deleteFn     func(ctx context.Context, clinicID, id uint64) error
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:         "正常: 未使用の種別を削除",
			countUsageFn: func(_ context.Context, _, _ uint64) (int64, error) { return 0, nil },
			deleteFn:     func(_ context.Context, _, _ uint64) error { return nil },
		},
		{
			name:         "エラー: 使用中の種別 → ErrConflict",
			countUsageFn: func(_ context.Context, _, _ uint64) (int64, error) { return 3, nil },
			wantErr:      true,
			wantErrIs:    apperrors.ErrConflict,
		},
		{
			name:         "エラー: CountUsage がエラー",
			countUsageFn: func(_ context.Context, _, _ uint64) (int64, error) { return 0, errors.New("db error") },
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseTypeRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourseType, error) {
					return &model.TrimmingCourseType{ID: id}, nil
				},
				countUsageFn: tt.countUsageFn,
				deleteFn:     tt.deleteFn,
			}
			svc := NewTrimmingCourseTypeService(repo)
			err := svc.Delete(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "want errors.Is(%v), got %v", tt.wantErrIs, err)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTrimmingCourseTypeService_Reorder(t *testing.T) {
	t.Run("正常", func(t *testing.T) {
		repo := &mockTrimmingCourseTypeRepository{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error { return nil },
		}
		svc := NewTrimmingCourseTypeService(repo)
		assert.NoError(t, svc.Reorder(context.Background(), 1, []uint64{2, 1}))
	})

	t.Run("エラー: ID リストが空", func(t *testing.T) {
		svc := NewTrimmingCourseTypeService(&mockTrimmingCourseTypeRepository{})
		assert.Error(t, svc.Reorder(context.Background(), 1, []uint64{}))
	})
}
