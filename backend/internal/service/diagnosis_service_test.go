package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- DiagnosisCategory モック ----

type mockDiagnosisCategoryRepository struct {
	findAllFn  func(ctx context.Context) ([]model.DiagnosisCategory, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.DiagnosisCategory, error)
	createFn   func(ctx context.Context, category *model.DiagnosisCategory) error
	updateFn   func(ctx context.Context, category *model.DiagnosisCategory) error
	deleteFn   func(ctx context.Context, id uint64) error
}

func (m *mockDiagnosisCategoryRepository) FindAll(ctx context.Context) ([]model.DiagnosisCategory, error) {
	return m.findAllFn(ctx)
}

func (m *mockDiagnosisCategoryRepository) FindByID(ctx context.Context, id uint64) (*model.DiagnosisCategory, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockDiagnosisCategoryRepository) Create(ctx context.Context, category *model.DiagnosisCategory) error {
	return m.createFn(ctx, category)
}

func (m *mockDiagnosisCategoryRepository) Update(ctx context.Context, category *model.DiagnosisCategory) error {
	return m.updateFn(ctx, category)
}

func (m *mockDiagnosisCategoryRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

// ---- DiagnosisName モック ----

type mockDiagnosisNameRepository struct {
	findAllFn          func(ctx context.Context) ([]model.DiagnosisName, error)
	findByCategoryIDFn func(ctx context.Context, categoryID uint64) ([]model.DiagnosisName, error)
	findByIDFn         func(ctx context.Context, id uint64) (*model.DiagnosisName, error)
	createFn           func(ctx context.Context, name *model.DiagnosisName) error
	updateFn           func(ctx context.Context, name *model.DiagnosisName) error
	deleteFn           func(ctx context.Context, id uint64) error
}

func (m *mockDiagnosisNameRepository) FindAll(ctx context.Context) ([]model.DiagnosisName, error) {
	return m.findAllFn(ctx)
}

func (m *mockDiagnosisNameRepository) FindByCategoryID(ctx context.Context, categoryID uint64) ([]model.DiagnosisName, error) {
	return m.findByCategoryIDFn(ctx, categoryID)
}

func (m *mockDiagnosisNameRepository) FindByID(ctx context.Context, id uint64) (*model.DiagnosisName, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockDiagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	return m.createFn(ctx, name)
}

func (m *mockDiagnosisNameRepository) Update(ctx context.Context, name *model.DiagnosisName) error {
	return m.updateFn(ctx, name)
}

func (m *mockDiagnosisNameRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

// ---- DiagnosisCategoryService テスト ----

func TestDiagnosisCategoryService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.DiagnosisCategory
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns category list",
			repoData: []model.DiagnosisCategory{
				{ID: 1, Name: "皮膚疾患"},
				{ID: 2, Name: "消化器疾患"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no categories exist",
			repoData: []model.DiagnosisCategory{},
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
			repo := &mockDiagnosisCategoryRepository{
				findAllFn: func(_ context.Context) ([]model.DiagnosisCategory, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewDiagnosisCategoryService(repo)

			categories, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, categories, tt.wantLen)
			}
		})
	}
}

func TestDiagnosisCategoryService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoCategory *model.DiagnosisCategory
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "returns category when found",
			id:           1,
			repoCategory: &model.DiagnosisCategory{ID: 1, Name: "皮膚疾患"},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when category does not exist",
			id:           999,
			repoCategory: nil,
			repoErr:      apperrors.WrapNotFound("diagnosis_category", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoCategory: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisCategoryRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.DiagnosisCategory, error) {
					return tt.repoCategory, tt.repoErr
				},
			}
			svc := NewDiagnosisCategoryService(repo)

			category, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, category)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoCategory, category)
			}
		})
	}
}

func TestDiagnosisCategoryService_Create(t *testing.T) {
	tests := []struct {
		name     string
		category *model.DiagnosisCategory
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "creates category successfully",
			category: &model.DiagnosisCategory{Name: "新規カテゴリ", ClinicID: 1},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns error when category already exists",
			category: &model.DiagnosisCategory{Name: "重複カテゴリ", ClinicID: 1},
			repoErr:  apperrors.WrapAlreadyExists("diagnosis_category", "重複カテゴリ"),
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			category: &model.DiagnosisCategory{Name: "エラーカテゴリ", ClinicID: 1},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisCategoryRepository{
				createFn: func(_ context.Context, _ *model.DiagnosisCategory) error {
					return tt.repoErr
				},
			}
			svc := NewDiagnosisCategoryService(repo)

			err := svc.Create(context.Background(), tt.category)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiagnosisCategoryService_Update(t *testing.T) {
	tests := []struct {
		name     string
		category *model.DiagnosisCategory
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "updates category successfully",
			category: &model.DiagnosisCategory{ID: 1, Name: "更新後カテゴリ名"},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns not found error when category does not exist",
			category: &model.DiagnosisCategory{ID: 999, Name: "存在しないカテゴリ"},
			repoErr:  apperrors.WrapNotFound("diagnosis_category", "999"),
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			category: &model.DiagnosisCategory{ID: 1, Name: "エラーケース"},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisCategoryRepository{
				updateFn: func(_ context.Context, _ *model.DiagnosisCategory) error {
					return tt.repoErr
				},
			}
			svc := NewDiagnosisCategoryService(repo)

			err := svc.Update(context.Background(), tt.category)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiagnosisCategoryService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes category successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns not found error when category does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("diagnosis_category", "999"),
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
			repo := &mockDiagnosisCategoryRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewDiagnosisCategoryService(repo)

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

// ---- DiagnosisNameService テスト ----

func TestDiagnosisNameService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.DiagnosisName
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns diagnosis name list",
			repoData: []model.DiagnosisName{
				{ID: 1, Name: "アトピー性皮膚炎", DiagnosisCategoryID: 1},
				{ID: 2, Name: "食物アレルギー", DiagnosisCategoryID: 1},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no names exist",
			repoData: []model.DiagnosisName{},
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
			repo := &mockDiagnosisNameRepository{
				findAllFn: func(_ context.Context) ([]model.DiagnosisName, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewDiagnosisNameService(repo)

			names, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, names, tt.wantLen)
			}
		})
	}
}

func TestDiagnosisNameService_ListByCategoryID(t *testing.T) {
	tests := []struct {
		name       string
		categoryID uint64
		repoData   []model.DiagnosisName
		repoErr    error
		wantLen    int
		wantErr    bool
	}{
		{
			name:       "returns names filtered by category ID",
			categoryID: 1,
			repoData: []model.DiagnosisName{
				{ID: 1, Name: "アトピー性皮膚炎", DiagnosisCategoryID: 1},
			},
			repoErr: nil,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:       "returns empty list when no names exist in category",
			categoryID: 99,
			repoData:   []model.DiagnosisName{},
			repoErr:    nil,
			wantLen:    0,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			categoryID: 1,
			repoData:   nil,
			repoErr:    errors.New("db error"),
			wantLen:    0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisNameRepository{
				findByCategoryIDFn: func(_ context.Context, _ uint64) ([]model.DiagnosisName, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewDiagnosisNameService(repo)

			names, err := svc.ListByCategoryID(context.Background(), tt.categoryID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, names, tt.wantLen)
			}
		})
	}
}

func TestDiagnosisNameService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoName     *model.DiagnosisName
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "returns diagnosis name when found",
			id:           1,
			repoName:     &model.DiagnosisName{ID: 1, Name: "アトピー性皮膚炎"},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when name does not exist",
			id:           999,
			repoName:     nil,
			repoErr:      apperrors.WrapNotFound("diagnosis_name", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoName:     nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisNameRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.DiagnosisName, error) {
					return tt.repoName, tt.repoErr
				},
			}
			svc := NewDiagnosisNameService(repo)

			name, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, name)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoName, name)
			}
		})
	}
}

func TestDiagnosisNameService_Create(t *testing.T) {
	tests := []struct {
		name    string
		dname   *model.DiagnosisName
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates diagnosis name successfully",
			dname:   &model.DiagnosisName{Name: "新規病名", ClinicID: 1, DiagnosisCategoryID: 1},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when name already exists",
			dname:   &model.DiagnosisName{Name: "重複病名", ClinicID: 1, DiagnosisCategoryID: 1},
			repoErr: apperrors.WrapAlreadyExists("diagnosis_name", "重複病名"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			dname:   &model.DiagnosisName{Name: "エラー病名", ClinicID: 1, DiagnosisCategoryID: 1},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisNameRepository{
				createFn: func(_ context.Context, _ *model.DiagnosisName) error {
					return tt.repoErr
				},
			}
			svc := NewDiagnosisNameService(repo)

			err := svc.Create(context.Background(), tt.dname)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiagnosisNameService_Update(t *testing.T) {
	tests := []struct {
		name    string
		dname   *model.DiagnosisName
		repoErr error
		wantErr bool
	}{
		{
			name:    "updates diagnosis name successfully",
			dname:   &model.DiagnosisName{ID: 1, Name: "更新後病名"},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when name does not exist",
			dname:   &model.DiagnosisName{ID: 999, Name: "存在しない病名"},
			repoErr: apperrors.WrapNotFound("diagnosis_name", "999"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			dname:   &model.DiagnosisName{ID: 1, Name: "エラーケース"},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisNameRepository{
				updateFn: func(_ context.Context, _ *model.DiagnosisName) error {
					return tt.repoErr
				},
			}
			svc := NewDiagnosisNameService(repo)

			err := svc.Update(context.Background(), tt.dname)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiagnosisNameService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes diagnosis name successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns not found error when name does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("diagnosis_name", "999"),
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
			repo := &mockDiagnosisNameRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewDiagnosisNameService(repo)

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
