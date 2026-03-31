package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

const testClinicID uint64 = 1

// ---- DiagnosisCategory モック ----

type mockDiagnosisCategoryRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.DiagnosisCategory, error)
	createFn   func(ctx context.Context, category *model.DiagnosisCategory) error
	updateFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
	reorderFn  func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockDiagnosisCategoryRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockDiagnosisCategoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisCategory, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockDiagnosisCategoryRepository) Create(ctx context.Context, category *model.DiagnosisCategory) error {
	return m.createFn(ctx, category)
}

func (m *mockDiagnosisCategoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockDiagnosisCategoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockDiagnosisCategoryRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- DiagnosisName モック ----

type mockDiagnosisNameRepository struct {
	findAllFn          func(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	findByCategoryIDFn func(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	findByIDFn         func(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	createFn           func(ctx context.Context, name *model.DiagnosisName) error
	updateFn           func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn           func(ctx context.Context, clinicID, id uint64) error
	reorderFn          func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockDiagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockDiagnosisNameRepository) FindByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return m.findByCategoryIDFn(ctx, clinicID, categoryID, page, limit)
}

func (m *mockDiagnosisNameRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockDiagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	return m.createFn(ctx, name)
}

func (m *mockDiagnosisNameRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockDiagnosisNameRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockDiagnosisNameRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// newCategoryService はテスト用ヘルパー
func newCategoryService(repo *mockDiagnosisCategoryRepository) DiagnosisCategoryService {
	return NewDiagnosisCategoryService(repo)
}

// newNameService はテスト用ヘルパー（categoryRepo FK validation付き）
func newNameService(repo *mockDiagnosisNameRepository, categoryRepo *mockDiagnosisCategoryRepository) DiagnosisNameService {
	return NewDiagnosisNameService(repo, categoryRepo)
}

// ---- DiagnosisCategoryService テスト ----

func TestDiagnosisCategoryService_List(t *testing.T) {
	tests := []struct {
		name      string
		repoData  []model.DiagnosisCategory
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "returns category list",
			repoData: []model.DiagnosisCategory{
				{ID: 1, Name: "皮膚疾患"},
				{ID: 2, Name: "消化器疾患"},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no categories exist",
			repoData:  []model.DiagnosisCategory{},
			repoTotal: 0,
			repoErr:   nil,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			repoData:  nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPage := 0
			capturedLimit := 0
			repo := &mockDiagnosisCategoryRepository{
				findAllFn: func(_ context.Context, _ uint64, page, limit int) ([]model.DiagnosisCategory, int64, error) {
					capturedPage = page
					capturedLimit = limit
					return tt.repoData, tt.repoTotal, tt.repoErr
				},
			}
			svc := newCategoryService(repo)

			categories, total, err := svc.List(context.Background(), testClinicID, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.Len(t, categories, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, 1, capturedPage)
				assert.Equal(t, 20, capturedLimit)
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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisCategory, error) {
					return tt.repoCategory, tt.repoErr
				},
			}
			svc := newCategoryService(repo)

			category, err := svc.GetByID(context.Background(), testClinicID, tt.id)

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
		name    string
		input   *CreateDiagnosisCategoryInput
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates category successfully",
			input:   &CreateDiagnosisCategoryInput{Name: "新規カテゴリ", IsActive: true},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when category already exists",
			input:   &CreateDiagnosisCategoryInput{Name: "重複カテゴリ"},
			repoErr: apperrors.WrapAlreadyExists("diagnosis_category", "重複カテゴリ"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   &CreateDiagnosisCategoryInput{Name: "エラーカテゴリ"},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisCategoryRepository{
				createFn: func(_ context.Context, _ *model.DiagnosisCategory) error {
					return tt.repoErr
				},
			}
			svc := newCategoryService(repo)

			result, err := svc.Create(context.Background(), testClinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				assert.Equal(t, testClinicID, result.ClinicID)
			}
		})
	}
}

func TestDiagnosisCategoryService_Update(t *testing.T) {
	updatedName := "更新後カテゴリ名"
	isActive := false

	tests := []struct {
		name     string
		id       uint64
		input    *UpdateDiagnosisCategoryInput
		repoErr  error
		fetchRes *model.DiagnosisCategory
		wantErr  bool
	}{
		{
			name:     "updates category successfully",
			id:       1,
			input:    &UpdateDiagnosisCategoryInput{Name: &updatedName, IsActive: &isActive},
			repoErr:  nil,
			fetchRes: &model.DiagnosisCategory{ID: 1, Name: updatedName, IsActive: isActive, ClinicID: testClinicID},
			wantErr:  false,
		},
		{
			name:    "returns invalid input when no fields provided",
			id:      1,
			input:   &UpdateDiagnosisCategoryInput{},
			wantErr: true,
		},
		{
			name:    "returns not found error when category does not exist",
			id:      999,
			input:   &UpdateDiagnosisCategoryInput{Name: &updatedName},
			repoErr: apperrors.WrapNotFound("diagnosis_category", "999"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			input:   &UpdateDiagnosisCategoryInput{Name: &updatedName},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisCategoryRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisCategory, error) {
					return tt.fetchRes, nil
				},
			}
			svc := newCategoryService(repo)

			result, err := svc.Update(context.Background(), testClinicID, tt.id, tt.input)

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
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := newCategoryService(repo)

			err := svc.Delete(context.Background(), testClinicID, tt.id)

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

func TestDiagnosisCategoryService_Reorder(t *testing.T) {
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
			repoErr: apperrors.WrapInvalidInput("diagnosis_category id 999 not found in this clinic"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisCategoryRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := newCategoryService(repo)

			err := svc.Reorder(context.Background(), testClinicID, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- DiagnosisNameService テスト ----

// defaultCategoryRepo は FindByID が常に成功する categoryRepo モック
func defaultCategoryRepo() *mockDiagnosisCategoryRepository {
	return &mockDiagnosisCategoryRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisCategory, error) {
			return &model.DiagnosisCategory{ID: 1}, nil
		},
		// FindAll は DiagnosisNameService では使用されないが、テスト追加時の panic防止
		findAllFn: func(_ context.Context, _ uint64, _, _ int) ([]model.DiagnosisCategory, int64, error) {
			return []model.DiagnosisCategory{}, 0, nil
		},
	}
}

func TestDiagnosisNameService_List(t *testing.T) {
	tests := []struct {
		name      string
		repoData  []model.DiagnosisName
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "returns diagnosis name list",
			repoData: []model.DiagnosisName{
				{ID: 1, Name: "アトピー性皮膚炎", DiagnosisCategoryID: 1},
				{ID: 2, Name: "食物アレルギー", DiagnosisCategoryID: 1},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no names exist",
			repoData:  []model.DiagnosisName{},
			repoTotal: 0,
			repoErr:   nil,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			repoData:  nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPage := 0
			capturedLimit := 0
			repo := &mockDiagnosisNameRepository{
				findAllFn: func(_ context.Context, _ uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
					capturedPage = page
					capturedLimit = limit
					return tt.repoData, tt.repoTotal, tt.repoErr
				},
			}
			svc := newNameService(repo, defaultCategoryRepo())

			names, total, err := svc.List(context.Background(), testClinicID, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.Len(t, names, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, 1, capturedPage)
				assert.Equal(t, 20, capturedLimit)
			}
		})
	}
}

func TestDiagnosisNameService_ListByCategoryID(t *testing.T) {
	tests := []struct {
		name       string
		categoryID uint64
		repoData   []model.DiagnosisName
		repoTotal  int64
		repoErr    error
		wantLen    int
		wantTotal  int64
		wantErr    bool
	}{
		{
			name:       "returns names filtered by category ID",
			categoryID: 1,
			repoData: []model.DiagnosisName{
				{ID: 1, Name: "アトピー性皮膚炎", DiagnosisCategoryID: 1},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:       "returns empty list when no names exist in category",
			categoryID: 99,
			repoData:   []model.DiagnosisName{},
			repoTotal:  0,
			repoErr:    nil,
			wantLen:    0,
			wantTotal:  0,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			categoryID: 1,
			repoData:   nil,
			repoTotal:  0,
			repoErr:    errors.New("db error"),
			wantLen:    0,
			wantTotal:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPage := 0
			capturedLimit := 0
			repo := &mockDiagnosisNameRepository{
				findByCategoryIDFn: func(_ context.Context, _, _ uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
					capturedPage = page
					capturedLimit = limit
					return tt.repoData, tt.repoTotal, tt.repoErr
				},
			}
			svc := newNameService(repo, defaultCategoryRepo())

			names, total, err := svc.ListByCategoryID(context.Background(), testClinicID, tt.categoryID, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.Len(t, names, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, 1, capturedPage)
				assert.Equal(t, 20, capturedLimit)
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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisName, error) {
					return tt.repoName, tt.repoErr
				},
			}
			svc := newNameService(repo, defaultCategoryRepo())

			name, err := svc.GetByID(context.Background(), testClinicID, tt.id)

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
		name        string
		input       *CreateDiagnosisNameInput
		repoErr     error
		categoryErr error
		wantErr     bool
	}{
		{
			name:    "creates diagnosis name successfully",
			input:   &CreateDiagnosisNameInput{Name: "新規病名", DiagnosisCategoryID: 1, IsActive: true},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns invalid input when category not found",
			input:       &CreateDiagnosisNameInput{Name: "病名", DiagnosisCategoryID: 999},
			categoryErr: apperrors.WrapNotFound("diagnosis_category", "999"),
			wantErr:     true,
		},
		{
			name:    "returns error when name already exists",
			input:   &CreateDiagnosisNameInput{Name: "重複病名", DiagnosisCategoryID: 1},
			repoErr: apperrors.WrapAlreadyExists("diagnosis_name", "重複病名"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   &CreateDiagnosisNameInput{Name: "エラー病名", DiagnosisCategoryID: 1},
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
			categoryErr := tt.categoryErr
			categoryRepo := &mockDiagnosisCategoryRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisCategory, error) {
					if categoryErr != nil {
						return nil, categoryErr
					}
					return &model.DiagnosisCategory{ID: 1}, nil
				},
			}
			svc := newNameService(repo, categoryRepo)

			result, err := svc.Create(context.Background(), testClinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				assert.Equal(t, testClinicID, result.ClinicID)
			}
		})
	}
}

func TestDiagnosisNameService_Update(t *testing.T) {
	updatedName := "更新後病名"
	isActive := false
	newCatID := uint64(2)

	tests := []struct {
		name        string
		id          uint64
		input       *UpdateDiagnosisNameInput
		repoErr     error
		fetchRes    *model.DiagnosisName
		categoryErr error
		wantErr     bool
	}{
		{
			name:     "updates diagnosis name successfully",
			id:       1,
			input:    &UpdateDiagnosisNameInput{Name: &updatedName, IsActive: &isActive},
			repoErr:  nil,
			fetchRes: &model.DiagnosisName{ID: 1, Name: updatedName, IsActive: isActive, ClinicID: testClinicID},
			wantErr:  false,
		},
		{
			name:    "returns invalid input when no fields provided",
			id:      1,
			input:   &UpdateDiagnosisNameInput{},
			wantErr: true,
		},
		{
			name:        "returns invalid input when category_id not found",
			id:          1,
			input:       &UpdateDiagnosisNameInput{DiagnosisCategoryID: &newCatID},
			categoryErr: apperrors.WrapNotFound("diagnosis_category", "2"),
			wantErr:     true,
		},
		{
			name:    "returns not found error when name does not exist",
			id:      999,
			input:   &UpdateDiagnosisNameInput{Name: &updatedName},
			repoErr: apperrors.WrapNotFound("diagnosis_name", "999"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			input:   &UpdateDiagnosisNameInput{Name: &updatedName},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisNameRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisName, error) {
					return tt.fetchRes, nil
				},
			}
			categoryErr := tt.categoryErr
			categoryRepo := &mockDiagnosisCategoryRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisCategory, error) {
					if categoryErr != nil {
						return nil, categoryErr
					}
					return &model.DiagnosisCategory{ID: 1}, nil
				},
			}
			svc := newNameService(repo, categoryRepo)

			result, err := svc.Update(context.Background(), testClinicID, tt.id, tt.input)

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
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := newNameService(repo, defaultCategoryRepo())

			err := svc.Delete(context.Background(), testClinicID, tt.id)

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

func TestDiagnosisNameService_Reorder(t *testing.T) {
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
			repoErr: apperrors.WrapInvalidInput("diagnosis_name id 999 not found in this clinic"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisNameRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := newNameService(repo, defaultCategoryRepo())

			err := svc.Reorder(context.Background(), testClinicID, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildDiagnosisCategoryUpdateFields(t *testing.T) {
	name := "テスト"
	isActive := false
	desc := "説明"
	sortOrder := 0

	t.Run("includes only non-nil fields", func(t *testing.T) {
		input := &UpdateDiagnosisCategoryInput{
			Name:      &name,
			IsActive:  &isActive,
			SortOrder: &sortOrder,
		}
		fields := buildDiagnosisCategoryUpdateFields(input)
		assert.Equal(t, name, fields[colDiagnosisCategoryName])
		assert.Equal(t, isActive, fields[colDiagnosisCategoryIsActive])
		assert.Equal(t, sortOrder, fields[colDiagnosisCategorySortOrder])
		assert.NotContains(t, fields, colDiagnosisCategoryDescription)
	})

	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		input := &UpdateDiagnosisCategoryInput{}
		fields := buildDiagnosisCategoryUpdateFields(input)
		assert.Empty(t, fields)
	})

	t.Run("includes description when provided", func(t *testing.T) {
		input := &UpdateDiagnosisCategoryInput{Description: &desc}
		fields := buildDiagnosisCategoryUpdateFields(input)
		assert.Equal(t, desc, fields[colDiagnosisCategoryDescription])
	})
}

func TestBuildDiagnosisNameUpdateFields(t *testing.T) {
	name := "病名テスト"
	isActive := false
	catID := uint64(5)
	sortOrder := 0

	t.Run("includes only non-nil fields", func(t *testing.T) {
		input := &UpdateDiagnosisNameInput{
			Name:                &name,
			IsActive:            &isActive,
			DiagnosisCategoryID: &catID,
			SortOrder:           &sortOrder,
		}
		fields := buildDiagnosisNameUpdateFields(input)
		assert.Equal(t, name, fields[colDiagnosisNameName])
		assert.Equal(t, isActive, fields[colDiagnosisNameIsActive])
		assert.Equal(t, catID, fields[colDiagnosisNameDiagnosisCategoryID])
		assert.Equal(t, sortOrder, fields[colDiagnosisNameSortOrder])
		assert.NotContains(t, fields, colDiagnosisNameDescription)
	})

	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		input := &UpdateDiagnosisNameInput{}
		fields := buildDiagnosisNameUpdateFields(input)
		assert.Empty(t, fields)
	})
}
