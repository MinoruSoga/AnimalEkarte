package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const testClinicID uint64 = 1

// ---- DiagnosisType モック ----

type mockDiagnosisTypeRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error)
	createFn                  func(ctx context.Context, category *model.DiagnosisType) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
	countChildrenByParentIDFn func(ctx context.Context, clinicID, categoryID uint64) (int64, error)
}

func (m *mockDiagnosisTypeRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockDiagnosisTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.DiagnosisType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockDiagnosisTypeRepository) Create(ctx context.Context, category *model.DiagnosisType) error {
	return m.createFn(ctx, category)
}

func (m *mockDiagnosisTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.DiagnosisType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockDiagnosisTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockDiagnosisTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockDiagnosisTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, categoryID)
}

// ---- DiagnosisName モック ----

type mockDiagnosisNameRepository struct {
	findAllFn                             func(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	findByCategoryIDFn                    func(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	findAllByFilterFn                     func(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)
	findByIDFn                            func(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	createFn                              func(ctx context.Context, name *model.DiagnosisName) error
	updateFieldsFn                        func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisName, error)
	deleteFn                              func(ctx context.Context, clinicID, id uint64) error
	reorderFn                             func(ctx context.Context, clinicID uint64, ids []uint64) error
	countClinicalPlansByDiagnosisNameIDFn func(ctx context.Context, clinicID, diagnosisNameID uint64) (int64, error)
}

func (m *mockDiagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockDiagnosisNameRepository) FindAllByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return m.findByCategoryIDFn(ctx, clinicID, categoryID, page, limit)
}

func (m *mockDiagnosisNameRepository) FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
	if m.findAllByFilterFn != nil {
		return m.findAllByFilterFn(ctx, clinicID, typeID)
	}
	return []model.DiagnosisName{}, nil
}

func (m *mockDiagnosisNameRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.DiagnosisName{ID: id, ClinicID: clinicID}, nil
}

func (m *mockDiagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	return m.createFn(ctx, name)
}

func (m *mockDiagnosisNameRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisName, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.DiagnosisName{ID: id, ClinicID: clinicID}, nil
}

func (m *mockDiagnosisNameRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockDiagnosisNameRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockDiagnosisNameRepository) CountUsageByDiagnosisNameID(ctx context.Context, clinicID, diagnosisNameID uint64) (int64, error) {
	if m.countClinicalPlansByDiagnosisNameIDFn != nil {
		return m.countClinicalPlansByDiagnosisNameIDFn(ctx, clinicID, diagnosisNameID)
	}
	return 0, nil
}

func (m *mockDiagnosisNameRepository) FindAllActive(_ context.Context, _ uint64, _ *uint64) ([]model.DiagnosisName, error) {
	return []model.DiagnosisName{}, nil
}

// newCategoryService はテスト用ヘルパー
func newCategoryService(repo *mockDiagnosisTypeRepository) DiagnosisTypeService {
	return NewDiagnosisTypeService(repo)
}

// newNameService はテスト用ヘルパー（categoryRepo FK validation付き）
func newNameService(repo *mockDiagnosisNameRepository, categoryRepo *mockDiagnosisTypeRepository) DiagnosisNameService {
	return NewDiagnosisNameService(repo, categoryRepo)
}

// ---- DiagnosisTypeService テスト ----

func TestDiagnosisTypeService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.DiagnosisType
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns category list",
			repoData: []model.DiagnosisType{
				{ID: 1, Name: "皮膚疾患"},
				{ID: 2, Name: "消化器疾患"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no categories exist",
			repoData: []model.DiagnosisType{},
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
			capturedPage := 0
			capturedLimit := 0
			repo := &mockDiagnosisTypeRepository{
				findAllFn: func(_ context.Context, _ uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
					capturedPage = page
					capturedLimit = limit
					return tt.repoData, int64(len(tt.repoData)), tt.repoErr
				},
			}
			svc := newCategoryService(repo)

			categories, _, err := svc.List(context.Background(), testClinicID, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, categories, tt.wantLen)
				assert.Equal(t, 1, capturedPage)
				assert.Equal(t, 20, capturedLimit)
			}
		})
	}
}

func TestDiagnosisTypeService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoCategory *model.DiagnosisType
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "returns category when found",
			id:           1,
			repoCategory: &model.DiagnosisType{ID: 1, Name: "皮膚疾患"},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when category does not exist",
			id:           999,
			repoCategory: nil,
			repoErr:      apperrors.WrapNotFound("diagnosis_type", "999"),
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
			repo := &mockDiagnosisTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisType, error) {
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

func TestDiagnosisTypeService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateDiagnosisTypeInput
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates category successfully",
			input:   &CreateDiagnosisTypeInput{Name: "新規カテゴリ", IsActive: true},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when category already exists",
			input:   &CreateDiagnosisTypeInput{Name: "重複カテゴリ"},
			repoErr: apperrors.WrapAlreadyExists("diagnosis_type", "重複カテゴリ"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   &CreateDiagnosisTypeInput{Name: "エラーカテゴリ"},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "rejects empty name",
			input:   &CreateDiagnosisTypeInput{Name: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisTypeRepository{
				createFn: func(_ context.Context, _ *model.DiagnosisType) error {
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

func TestDiagnosisTypeService_Update(t *testing.T) {
	updatedName := "更新後カテゴリ名"
	isActive := false
	emptyName := "   "

	tests := []struct {
		name            string
		id              uint64
		input           *UpdateDiagnosisTypeInput
		findByIDErr     error
		updateFieldsErr error
		fetchRes        *model.DiagnosisType
		wantErr         bool
		wantNF          bool
	}{
		{
			name:            "updates category successfully",
			id:              1,
			input:           &UpdateDiagnosisTypeInput{Name: &updatedName, IsActive: &isActive},
			updateFieldsErr: nil,
			fetchRes:        &model.DiagnosisType{ID: 1, Name: updatedName, IsActive: isActive, ClinicID: testClinicID},
			wantErr:         false,
		},
		{
			name:    "returns invalid input when no fields provided",
			id:      1,
			input:   &UpdateDiagnosisTypeInput{},
			wantErr: true,
		},
		{
			name:            "returns not found error when category does not exist",
			id:              999,
			input:           &UpdateDiagnosisTypeInput{Name: &updatedName},
			updateFieldsErr: apperrors.WrapNotFound("diagnosis_type", "999"),
			wantErr:         true,
		},
		{
			name:            "returns error on repository failure",
			id:              1,
			input:           &UpdateDiagnosisTypeInput{Name: &updatedName},
			updateFieldsErr: errors.New("db error"),
			wantErr:         true,
		},
		{
			name:        "returns not found error when FindByID fails",
			id:          999,
			input:       &UpdateDiagnosisTypeInput{Name: &updatedName},
			findByIDErr: apperrors.WrapNotFound("diagnosis_type", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:    "returns invalid input when name is blank",
			id:      1,
			input:   &UpdateDiagnosisTypeInput{Name: &emptyName},
			wantErr: true,
		},
		{
			name:    "returns invalid input when input is nil",
			id:      1,
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisTypeRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.DiagnosisType{ID: id, ClinicID: testClinicID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.DiagnosisType, error) {
					return tt.fetchRes, tt.updateFieldsErr
				},
			}
			svc := newCategoryService(repo)

			result, err := svc.Update(context.Background(), testClinicID, tt.id, tt.input)

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

func TestDiagnosisTypeService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		nameCount     int64
		countNamesErr error
		findByIDErr   error
		repoErr       error
		wantErr       bool
		wantNF        bool
		wantConflict  bool
	}{
		{
			name:          "deletes category successfully when no diagnosis names exist",
			id:            1,
			nameCount:     0,
			countNamesErr: nil,
			repoErr:       nil,
			wantErr:       false,
		},
		{
			name:          "returns conflict error when category has diagnosis names",
			id:            1,
			nameCount:     3,
			countNamesErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns error when diagnosis name count check fails",
			id:            1,
			nameCount:     0,
			countNamesErr: errors.New("db error"),
			repoErr:       nil,
			wantErr:       true,
		},
		{
			name:        "returns not found error when category does not exist",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("diagnosis_type", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:          "returns error on repository failure",
			id:            1,
			nameCount:     0,
			countNamesErr: nil,
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisTypeRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.DiagnosisType{ID: id}, nil
				},
				countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.nameCount, tt.countNamesErr
				},
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
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiagnosisTypeService_Reorder(t *testing.T) {
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
			repoErr: apperrors.WrapInvalidInput("diagnosis_type id 999 not found in this clinic"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisTypeRepository{
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
func defaultCategoryRepo() *mockDiagnosisTypeRepository {
	return &mockDiagnosisTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisType, error) {
			return &model.DiagnosisType{ID: 1}, nil
		},
		// FindAll は DiagnosisNameService では使用されないが、テスト追加時の panic防止
		findAllFn: func(_ context.Context, _ uint64, _, _ int) ([]model.DiagnosisType, int64, error) {
			return []model.DiagnosisType{}, 0, nil
		},
	}
}

func TestDiagnosisNameService_List(t *testing.T) {
	tests := []struct {
		name        string
		typeID      *uint64
		repoData    []model.DiagnosisName
		repoErr     error
		wantLen     int
		wantErr     bool
		useCategory bool
	}{
		{
			name: "returns all diagnosis names when typeID is nil",
			repoData: []model.DiagnosisName{
				{ID: 1, Name: "アトピー性皮膚炎", DiagnosisTypeID: 1},
				{ID: 2, Name: "食物アレルギー", DiagnosisTypeID: 1},
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
			name:     "propagates repository error on FindAll",
			repoData: nil,
			repoErr:  errors.New("db connection error"),
			wantErr:  true,
		},
		{
			name:        "returns names filtered by typeID when provided",
			typeID:      func() *uint64 { v := uint64(1); return &v }(),
			repoData:    []model.DiagnosisName{{ID: 1, Name: "アトピー性皮膚炎", DiagnosisTypeID: 1}},
			repoErr:     nil,
			wantLen:     1,
			wantErr:     false,
			useCategory: true,
		},
		{
			name:        "propagates repository error on FindAllByCategoryID",
			typeID:      func() *uint64 { v := uint64(1); return &v }(),
			repoData:    nil,
			repoErr:     errors.New("db error"),
			wantErr:     true,
			useCategory: true,
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
					return tt.repoData, int64(len(tt.repoData)), tt.repoErr
				},
				findByCategoryIDFn: func(_ context.Context, _, _ uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
					capturedPage = page
					capturedLimit = limit
					return tt.repoData, int64(len(tt.repoData)), tt.repoErr
				},
			}
			svc := newNameService(repo, defaultCategoryRepo())

			names, _, err := svc.List(context.Background(), testClinicID, tt.typeID, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, names, tt.wantLen)
				assert.Equal(t, 1, capturedPage)
				assert.Equal(t, 20, capturedLimit)
			}
		})
	}
}

// TestDiagnosisNameService_ListNames は #418 ページネーションなし一覧取得の委譲を検証する。
func TestDiagnosisNameService_ListNames(t *testing.T) {
	tests := []struct {
		name     string
		typeID   *uint64
		repoData []model.DiagnosisName
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns all diagnosis names when typeID is nil",
			repoData: []model.DiagnosisName{
				{ID: 1, Name: "アトピー性皮膚炎", DiagnosisTypeID: 1},
				{ID: 2, Name: "食物アレルギー", DiagnosisTypeID: 1},
			},
			wantLen: 2,
		},
		{
			name:     "returns names filtered by typeID when provided",
			typeID:   func() *uint64 { v := uint64(1); return &v }(),
			repoData: []model.DiagnosisName{{ID: 1, Name: "アトピー性皮膚炎", DiagnosisTypeID: 1}},
			wantLen:  1,
		},
		{
			name:     "returns empty list when no names exist",
			repoData: []model.DiagnosisName{},
			wantLen:  0,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedTypeID *uint64
			repo := &mockDiagnosisNameRepository{
				findAllByFilterFn: func(_ context.Context, _ uint64, typeID *uint64) ([]model.DiagnosisName, error) {
					capturedTypeID = typeID
					return tt.repoData, tt.repoErr
				},
			}
			svc := newNameService(repo, defaultCategoryRepo())

			names, err := svc.ListNames(context.Background(), testClinicID, tt.typeID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, names, tt.wantLen)
				assert.Equal(t, tt.typeID, capturedTypeID)
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
			input:   &CreateDiagnosisNameInput{Name: "新規病名", DiagnosisTypeID: 1, IsActive: true},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns invalid input when category not found",
			input:       &CreateDiagnosisNameInput{Name: "病名", DiagnosisTypeID: 999},
			categoryErr: apperrors.WrapNotFound("diagnosis_type", "999"),
			wantErr:     true,
		},
		{
			name:    "returns error when name already exists",
			input:   &CreateDiagnosisNameInput{Name: "重複病名", DiagnosisTypeID: 1},
			repoErr: apperrors.WrapAlreadyExists("diagnosis_name", "重複病名"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   &CreateDiagnosisNameInput{Name: "エラー病名", DiagnosisTypeID: 1},
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
			categoryRepo := &mockDiagnosisTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisType, error) {
					if categoryErr != nil {
						return nil, categoryErr
					}
					return &model.DiagnosisType{ID: 1}, nil
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
		name            string
		id              uint64
		input           *UpdateDiagnosisNameInput
		findByIDErr     error
		updateFieldsErr error
		fetchRes        *model.DiagnosisName
		categoryErr     error
		wantErr         bool
		wantNF          bool
	}{
		{
			name:            "updates diagnosis name successfully",
			id:              1,
			input:           &UpdateDiagnosisNameInput{Name: &updatedName, IsActive: &isActive},
			updateFieldsErr: nil,
			fetchRes:        &model.DiagnosisName{ID: 1, Name: updatedName, IsActive: isActive, ClinicID: testClinicID},
			wantErr:         false,
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
			input:       &UpdateDiagnosisNameInput{DiagnosisTypeID: &newCatID},
			categoryErr: apperrors.WrapNotFound("diagnosis_type", "2"),
			wantErr:     true,
		},
		{
			name:            "returns not found error when name does not exist",
			id:              999,
			input:           &UpdateDiagnosisNameInput{Name: &updatedName},
			updateFieldsErr: apperrors.WrapNotFound("diagnosis_name", "999"),
			wantErr:         true,
		},
		{
			name:            "returns error on repository failure",
			id:              1,
			input:           &UpdateDiagnosisNameInput{Name: &updatedName},
			updateFieldsErr: errors.New("db error"),
			wantErr:         true,
		},
		{
			name:        "returns not found error when FindByID fails",
			id:          999,
			input:       &UpdateDiagnosisNameInput{Name: &updatedName},
			findByIDErr: apperrors.WrapNotFound("diagnosis_name", "999"),
			wantErr:     true,
			wantNF:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisNameRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.DiagnosisName{ID: id, ClinicID: testClinicID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.DiagnosisName, error) {
					return tt.fetchRes, tt.updateFieldsErr
				},
			}
			categoryErr := tt.categoryErr
			categoryRepo := &mockDiagnosisTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.DiagnosisType, error) {
					if categoryErr != nil {
						return nil, categoryErr
					}
					return &model.DiagnosisType{ID: 1}, nil
				},
			}
			svc := newNameService(repo, categoryRepo)

			result, err := svc.Update(context.Background(), testClinicID, tt.id, tt.input)

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

func TestDiagnosisNameService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		planCount    int64
		countErr     error
		findByIDErr  error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:    "deletes diagnosis name successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:        "returns not found error when name does not exist",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("diagnosis_name", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
		{
			name:         "使用中の診断名は削除できない",
			id:           2,
			planCount:    3,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:     "returns error when usage count check fails",
			id:       1,
			countErr: errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDiagnosisNameRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.DiagnosisName{ID: id}, nil
				},
				countClinicalPlansByDiagnosisNameIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.planCount, tt.countErr
				},
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
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
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

func TestBuildDiagnosisTypeUpdate(t *testing.T) {
	name := "テスト"
	isActive := false
	desc := "説明"
	sortOrder := 0

	t.Run("includes only non-nil fields", func(t *testing.T) {
		input := &UpdateDiagnosisTypeInput{
			Name:      &name,
			IsActive:  &isActive,
			SortOrder: &sortOrder,
		}
		fields := buildDiagnosisTypeUpdate(input)
		assert.Equal(t, name, fields[colDiagnosisName])
		assert.Equal(t, isActive, fields[colDiagnosisIsActive])
		assert.Equal(t, sortOrder, fields[colDiagnosisSortOrder])
		assert.NotContains(t, fields, colDiagnosisDescription)
	})

	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		input := &UpdateDiagnosisTypeInput{}
		fields := buildDiagnosisTypeUpdate(input)
		assert.Empty(t, fields)
	})

	t.Run("includes description when provided", func(t *testing.T) {
		input := &UpdateDiagnosisTypeInput{Description: &desc}
		fields := buildDiagnosisTypeUpdate(input)
		assert.Equal(t, desc, fields[colDiagnosisDescription])
	})
}

func TestBuildDiagnosisNameUpdate(t *testing.T) {
	name := "病名テスト"
	isActive := false
	catID := uint64(5)
	sortOrder := 0

	t.Run("includes only non-nil fields", func(t *testing.T) {
		input := &UpdateDiagnosisNameInput{
			Name:            &name,
			IsActive:        &isActive,
			DiagnosisTypeID: &catID,
			SortOrder:       &sortOrder,
		}
		fields := buildDiagnosisNameUpdate(input)
		assert.Equal(t, name, fields[colDiagnosisName])
		assert.Equal(t, isActive, fields[colDiagnosisIsActive])
		assert.Equal(t, catID, fields[colDiagnosisNameDiagnosisTypeID])
		assert.Equal(t, sortOrder, fields[colDiagnosisSortOrder])
		assert.NotContains(t, fields, colDiagnosisDescription)
	})

	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		input := &UpdateDiagnosisNameInput{}
		fields := buildDiagnosisNameUpdate(input)
		assert.Empty(t, fields)
	})
}
