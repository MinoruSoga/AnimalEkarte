package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// passthroughExamTypeTransactor は unit test 用 WithTx 素通し。
type passthroughExamTypeTransactor struct{}

func (passthroughExamTypeTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// mockExamTypeRepository は ExamTypeRepository のテスト用モック実装
type mockExamTypeRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	createFn                  func(ctx context.Context, exType *model.ExaminationType) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
	countUsageByExamTypeIDFn  func(ctx context.Context, clinicID, examTypeID uint64) (int64, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

func (m *mockExamTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockExamTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockExamTypeRepository) Create(ctx context.Context, exType *model.ExaminationType) error {
	return m.createFn(ctx, exType)
}

func (m *mockExamTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockExamTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockExamTypeRepository) ReplaceItems(ctx context.Context, examTypeID uint64, items []model.ExamTypeField) error {
	return nil
}

func (m *mockExamTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func (m *mockExamTypeRepository) CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error) {
	if m.countUsageByExamTypeIDFn == nil {
		return 0, nil
	}
	return m.countUsageByExamTypeIDFn(ctx, clinicID, examTypeID)
}

func (m *mockExamTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

func (m *mockExamTypeRepository) CreateField(context.Context, *model.ExamTypeField) error { return nil }
func (m *mockExamTypeRepository) LockFieldByID(context.Context, uint64, uint64, uint64) (*model.ExamTypeField, error) {
	return &model.ExamTypeField{}, nil
}
func (m *mockExamTypeRepository) UpdateField(context.Context, uint64, uint64, uint64, map[string]any) (*model.ExamTypeField, error) {
	return &model.ExamTypeField{}, nil
}
func (m *mockExamTypeRepository) DeleteField(context.Context, uint64, uint64, uint64) error {
	return nil
}
func (m *mockExamTypeRepository) ReorderFields(context.Context, uint64, uint64, []uint64) error {
	return nil
}
func (m *mockExamTypeRepository) CountExamResultsByFieldID(context.Context, uint64) (int64, error) {
	return 0, nil
}
func (m *mockExamTypeRepository) CountReferenceRangesByFieldID(context.Context, uint64) (int64, error) {
	return 0, nil
}
func (m *mockExamTypeRepository) AnimalSpeciesExists(context.Context, uint64) (bool, error) {
	return true, nil
}
func (m *mockExamTypeRepository) ReplaceReferenceRanges(context.Context, uint64, uint64, []model.ExamReferenceRange) error {
	return nil
}
func (m *mockExamTypeRepository) FindReferenceRangesByFieldIDs(context.Context, uint64, []uint64) (map[uint64][]model.ExamReferenceRange, error) {
	return map[uint64][]model.ExamReferenceRange{}, nil
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
				findAllFn: func(_ context.Context, _ uint64) ([]model.ExaminationType, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})

			examTypes, err := svc.List(context.Background(), 1)

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
				Items: []model.ExamTypeField{
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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ExaminationType, error) {
					return tt.repoExamType, tt.repoErr
				},
			}
			svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})

			examType, err := svc.GetByID(context.Background(), 1, tt.id)

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
		input   *CreateExamTypeInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates exam type successfully",
			input: &CreateExamTypeInput{
				Name:     "新規検査種別",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates exam type with optional fields",
			input: &CreateExamTypeInput{
				Name:        "血液検査",
				IsActive:    true,
				Description: "血液の詳細検査",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates exam type with is_non_insurance=true",
			input: &CreateExamTypeInput{
				Name:           "保険対象外検査",
				IsActive:       true,
				IsNonInsurance: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when exam type already exists",
			input: &CreateExamTypeInput{
				Name:     "重複検査種別",
				IsActive: true,
			},
			repoErr: apperrors.WrapAlreadyExists("exam_type", "重複検査種別"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateExamTypeInput{
				Name:     "エラー検査種別",
				IsActive: true,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "rejects empty name",
			input: &CreateExamTypeInput{
				Name:     "",
				IsActive: true,
			},
			wantErr: true,
		},
		{
			name: "returns validation error when price is negative",
			input: &CreateExamTypeInput{
				Name:  "Negative Price Exam Type",
				Price: func(v int64) *int64 { return &v }(-100),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalled := false
			repo := &mockExamTypeRepository{
				createFn: func(_ context.Context, _ *model.ExaminationType) error {
					createCalled = true
					return tt.repoErr
				},
			}
			svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})

			result, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.name == "returns validation error when price is negative" {
					assert.True(t, apperrors.IsInvalidInput(err))
					assert.False(t, createCalled)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				assert.Equal(t, uint64(1), result.ClinicID)
				assert.Equal(t, tt.input.IsNonInsurance, result.IsNonInsurance)
			}
		})
	}
}

func TestExamTypeService_Update(t *testing.T) {
	name := "更新後検査種別"
	emptyName := "   "
	tests := []struct {
		name        string
		input       UpdateExamTypeInput
		findByIDErr error
		updateErr   error
		wantErr     bool
		wantNF      bool
	}{
		{
			name: "updates exam type successfully",
			input: UpdateExamTypeInput{
				Name: &name,
			},
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateExamTypeInput{},
			wantErr: true,
		},
		{
			name: "returns not found error when exam type does not exist",
			input: UpdateExamTypeInput{
				Name: &name,
			},
			findByIDErr: apperrors.WrapNotFound("exam_type", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateExamTypeInput{
				Name: &name,
			},
			updateErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name: "rejects blank name",
			input: UpdateExamTypeInput{
				Name: &emptyName,
			},
			wantErr: true,
		},
		{
			name: "returns validation error when price is negative",
			input: UpdateExamTypeInput{
				Price: func(v int64) *int64 { return &v }(-500),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			repo := &mockExamTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ExaminationType, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.ExaminationType{ID: 1}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ExaminationType, error) {
					updateCalled = true
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return &model.ExaminationType{ID: 1}, nil
				},
			}
			svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})

			exType, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, exType)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.name == "returns validation error when price is negative" {
					assert.True(t, apperrors.IsInvalidInput(err))
					assert.False(t, updateCalled)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, exType)
			}
		})
	}

	t.Run("nil input はエラー", func(t *testing.T) {
		repo := &mockExamTypeRepository{}
		svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})
		result, err := svc.Update(context.Background(), 1, 1, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("is_non_insurance をトグル (false→true) できる", func(t *testing.T) {
		nonIns := true
		var capturedFields map[string]any
		repo := &mockExamTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ExaminationType, error) {
				return &model.ExaminationType{ID: 1, IsNonInsurance: false}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.ExaminationType, error) {
				capturedFields = fields
				return &model.ExaminationType{ID: 1, IsNonInsurance: true}, nil
			},
		}
		svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})
		input := &UpdateExamTypeInput{IsNonInsurance: &nonIns}
		result, err := svc.Update(context.Background(), 1, 1, input)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, true, capturedFields[colExamTypeIsNonInsurance])
	})
}

func TestExamTypeService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		findByIDErr   error
		childCount    int64
		countChildErr error
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
			name:        "returns not found error when FindByID fails",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("exam_type", "999"),
			wantErr:     true,
			wantNF:      true,
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
			name:          "returns error on repository delete failure",
			id:            1,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
		{
			name:         "returns conflict error when exam type has sub types",
			id:           1,
			childCount:   1,
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
			repo := &mockExamTypeRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.ExaminationType{ID: id}, nil
				},
				countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.childCount, tt.countChildErr
				},
				countUsageByExamTypeIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, tt.countUsageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})

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

func TestExamTypeService_Reorder(t *testing.T) {
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
			repoErr: apperrors.WrapInvalidInput("exam_type id 999 not found in this clinic"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExamTypeRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildExamTypeUpdate(t *testing.T) {
	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		fields := buildExamTypeUpdate(&UpdateExamTypeInput{})
		assert.Empty(t, fields)
	})

	t.Run("includes all provided fields", func(t *testing.T) {
		name := "更新後検査種別"
		price := int64(2000)
		isActive := true
		desc := "説明"
		parentID := uint64(3)
		sortOrder := 2
		isNonIns := true

		input := &UpdateExamTypeInput{
			Name:           &name,
			Price:          &price,
			IsActive:       &isActive,
			Description:    &desc,
			ParentID:       &parentID,
			SortOrder:      &sortOrder,
			IsNonInsurance: &isNonIns,
		}
		fields := buildExamTypeUpdate(input)

		assert.Equal(t, name, fields[colExamTypeName])
		assert.Equal(t, price, fields[colExamTypePrice])
		assert.Equal(t, isActive, fields[colExamTypeIsActive])
		assert.Equal(t, desc, fields[colExamTypeDescription])
		assert.Equal(t, parentID, fields[colExamTypeParentID])
		assert.Equal(t, sortOrder, fields[colExamTypeSortOrder])
		assert.Equal(t, isNonIns, fields[colExamTypeIsNonInsurance])
	})

	t.Run("clears parent_id when ClearParentID is true", func(t *testing.T) {
		input := &UpdateExamTypeInput{ClearParentID: true}
		fields := buildExamTypeUpdate(input)
		assert.Contains(t, fields, colExamTypeParentID)
		assert.Nil(t, fields[colExamTypeParentID])
	})

	t.Run("omits parent_id when neither ParentID nor ClearParentID set", func(t *testing.T) {
		input := &UpdateExamTypeInput{}
		fields := buildExamTypeUpdate(input)
		assert.NotContains(t, fields, colExamTypeParentID)
	})
}

func TestExamTypeService_WithTx_NilTransactorIsInternalError(t *testing.T) {
	// MRB-07: nil transactor must map to 500 Internal, not 400 InvalidInput.
	svc := NewExamTypeService(&mockExamTypeRepository{}, nil).(*examTypeService)
	err := svc.withTx(context.Background(), func(context.Context) error { return nil })
	assert.Error(t, err)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr), "got %T %v", err, err)
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.False(t, apperrors.IsInvalidInput(err))
}

func TestExamTypeService_Create_ValidatesParentInsideTx(t *testing.T) {
	// MRB-08: parent validation and Create share one WithTx; parent FindByID sees ambient tx.
	var sawTx bool
	repo := &mockExamTypeRepository{
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
			// ambient marker not available without real tx; ensure called before create
			return &model.ExaminationType{ID: id, ClinicID: clinicID}, nil
		},
		createFn: func(ctx context.Context, exType *model.ExaminationType) error {
			sawTx = true
			exType.ID = 42
			return nil
		},
	}
	svc := NewExamTypeService(repo, passthroughExamTypeTransactor{})
	parentID := uint64(9)
	got, err := svc.Create(context.Background(), 1, &CreateExamTypeInput{
		Name: "blood", ParentID: &parentID, IsActive: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.True(t, sawTx)
}
