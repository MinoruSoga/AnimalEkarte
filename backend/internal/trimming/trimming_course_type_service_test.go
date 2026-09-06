package trimming

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockTrimmingCourseTypeRepository は TrimmingCourseTypeRepository のテスト用モック (#73)
type mockTrimmingCourseTypeRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error)
	createFn       func(ctx context.Context, m *model.TrimmingCourseType) (*model.TrimmingCourseType, error)
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error)
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

func (m *mockTrimmingCourseTypeRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, cmd)
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
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		got, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseTypeInput{Name: "シャンプー", SortOrder: 1})
		assert.NoError(t, err)
		assert.Equal(t, created, got)
	})

	t.Run("エラー: 名前が空", func(t *testing.T) {
		svc := NewTrimmingCourseTypeService(&mockTrimmingCourseTypeRepository{}, &mockTransactor{})
		got, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseTypeInput{Name: ""})
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("エラー: 名前が空白のみ", func(t *testing.T) {
		svc := NewTrimmingCourseTypeService(&mockTrimmingCourseTypeRepository{}, &mockTransactor{})
		got, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseTypeInput{Name: "   "})
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("エラー: repo がエラーを返す", func(t *testing.T) {
		repo := &mockTrimmingCourseTypeRepository{
			createFn: func(_ context.Context, _ *model.TrimmingCourseType) (*model.TrimmingCourseType, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		got, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseTypeInput{Name: "エラーケース"})
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestTrimmingCourseTypeService_List(t *testing.T) {
	tests := []struct {
		name    string
		items   []model.TrimmingCourseType
		repoErr error
		wantLen int
		wantErr bool
	}{
		{
			name: "正常: 一覧を返す",
			items: []model.TrimmingCourseType{
				{ID: 1, ClinicID: 1, Name: "シャンプー"},
				{ID: 2, ClinicID: 1, Name: "カット"},
			},
			wantLen: 2,
		},
		{
			name:    "正常: 空リスト",
			items:   []model.TrimmingCourseType{},
			wantLen: 0,
		},
		{
			name:    "エラー: repo エラーを伝播",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.TrimmingCourseType, error) {
					return tt.items, tt.repoErr
				},
			}
			svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
			got, err := svc.List(context.Background(), 1)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestTrimmingCourseTypeService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		item    *model.TrimmingCourseType
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "正常: 取得できる",
			item: &model.TrimmingCourseType{ID: 1, Name: "シャンプー"},
		},
		{
			name:    "エラー: 見つからない",
			repoErr: apperrors.WrapNotFound("trimming_course_type", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:    "エラー: repo エラーを伝播",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingCourseType, error) {
					return tt.item, tt.repoErr
				},
			}
			svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
			got, err := svc.GetByID(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.item, got)
		})
	}
}

func TestTrimmingCourseTypeService_Update(t *testing.T) {
	name := "更新後"
	sortOrder := 5
	isActive := true

	t.Run("正常: 更新できる", func(t *testing.T) {
		repo := &mockTrimmingCourseTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourseType, error) {
				return &model.TrimmingCourseType{ID: id}, nil
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
				return &model.TrimmingCourseType{ID: id, Name: name}, nil
			},
		}
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		got, err := svc.Update(context.Background(), 1, 1, &UpdateTrimmingCourseTypeInput{Name: &name, SortOrder: &sortOrder, IsActive: &isActive})
		assert.NoError(t, err)
		assert.Equal(t, name, got.Name)
	})

	t.Run("エラー: input が nil", func(t *testing.T) {
		svc := NewTrimmingCourseTypeService(&mockTrimmingCourseTypeRepository{}, &mockTransactor{})
		got, err := svc.Update(context.Background(), 1, 1, nil)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("エラー: FindByID が見つからない", func(t *testing.T) {
		repo := &mockTrimmingCourseTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingCourseType, error) {
				return nil, apperrors.WrapNotFound("trimming_course_type", "999")
			},
		}
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		got, err := svc.Update(context.Background(), 1, 999, &UpdateTrimmingCourseTypeInput{Name: &name})
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("エラー: 名前が空文字", func(t *testing.T) {
		empty := ""
		repo := &mockTrimmingCourseTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourseType, error) {
				return &model.TrimmingCourseType{ID: id}, nil
			},
		}
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		got, err := svc.Update(context.Background(), 1, 1, &UpdateTrimmingCourseTypeInput{Name: &empty})
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("エラー: 更新フィールドなし", func(t *testing.T) {
		repo := &mockTrimmingCourseTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourseType, error) {
				return &model.TrimmingCourseType{ID: id}, nil
			},
		}
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		got, err := svc.Update(context.Background(), 1, 1, &UpdateTrimmingCourseTypeInput{})
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("エラー: repo Update がエラーを返す", func(t *testing.T) {
		repo := &mockTrimmingCourseTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourseType, error) {
				return &model.TrimmingCourseType{ID: id}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		got, err := svc.Update(context.Background(), 1, 1, &UpdateTrimmingCourseTypeInput{Name: &name})
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestBuildTrimmingCourseTypeUpdate(t *testing.T) {
	name := "新名称"
	sortOrder := 2
	isActive := false

	tests := []struct {
		name  string
		input *UpdateTrimmingCourseTypeInput
		want  map[string]any
	}{
		{
			name:  "全フィールド未指定なら空map",
			input: &UpdateTrimmingCourseTypeInput{},
			want:  map[string]any{},
		},
		{
			name:  "全フィールド指定で全キーが含まれる",
			input: &UpdateTrimmingCourseTypeInput{Name: &name, SortOrder: &sortOrder, IsActive: &isActive},
			want: map[string]any{
				colTrimmingCourseTypeName:      name,
				colTrimmingCourseTypeSortOrder: sortOrder,
				colTrimmingCourseTypeIsActive:  isActive,
			},
		},
		{
			name:  "IsActive のみ指定",
			input: &UpdateTrimmingCourseTypeInput{IsActive: &isActive},
			want:  map[string]any{colTrimmingCourseTypeIsActive: isActive},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTrimmingCourseTypeUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
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
		{
			name: "エラー: 削除対象が見つからない",
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return apperrors.WrapNotFound("trimming_course_type", "999")
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrNotFound,
		},
		{
			name:         "エラー: repo Delete がエラーを返す",
			countUsageFn: func(_ context.Context, _, _ uint64) (int64, error) { return 0, nil },
			deleteFn:     func(_ context.Context, _, _ uint64) error { return errors.New("db error") },
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseTypeRepository{
				countUsageFn: tt.countUsageFn,
				deleteFn:     tt.deleteFn,
			}
			svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
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
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		assert.NoError(t, svc.Reorder(context.Background(), 1, []uint64{2, 1}))
	})

	t.Run("エラー: ID リストが空", func(t *testing.T) {
		svc := NewTrimmingCourseTypeService(&mockTrimmingCourseTypeRepository{}, &mockTransactor{})
		assert.Error(t, svc.Reorder(context.Background(), 1, []uint64{}))
	})

	t.Run("エラー: repo エラーを伝播", func(t *testing.T) {
		repo := &mockTrimmingCourseTypeRepository{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error { return errors.New("db error") },
		}
		svc := NewTrimmingCourseTypeService(repo, &mockTransactor{})
		assert.Error(t, svc.Reorder(context.Background(), 1, []uint64{1, 2}))
	})
}
