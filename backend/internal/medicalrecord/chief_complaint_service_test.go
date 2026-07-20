package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ChiefComplaintType モック ----

type mockChiefComplaintTypeRepository struct {
	findAllFn                        func(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error)
	findByIDFn                       func(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error)
	countUsageByChiefComplaintTypeFn func(ctx context.Context, clinicID, id uint64) (int64, error)
	createFn                         func(ctx context.Context, category *model.ChiefComplaintType) error
	updateFieldsFn                   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error)
	deleteFn                         func(ctx context.Context, clinicID, id uint64) error
	reorderErr                       error
}

func (m *mockChiefComplaintTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockChiefComplaintTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ChiefComplaintType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockChiefComplaintTypeRepository) CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByChiefComplaintTypeFn != nil {
		return m.countUsageByChiefComplaintTypeFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockChiefComplaintTypeRepository) Create(ctx context.Context, category *model.ChiefComplaintType) error {
	return m.createFn(ctx, category)
}

func (m *mockChiefComplaintTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.ChiefComplaintType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockChiefComplaintTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockChiefComplaintTypeRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

// ---- Tests ----

func TestChiefComplaintTypeService_List(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		repoData []model.ChiefComplaintType
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns all categories for clinic",
			clinicID: 1,
			repoData: []model.ChiefComplaintType{
				{ID: 1, ClinicID: 1, Name: "症状A", IsActive: true},
				{ID: 2, ClinicID: 1, Name: "症状B", IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no categories exist",
			clinicID: 999,
			repoData: []model.ChiefComplaintType{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			clinicID: 1,
			repoData: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintTypeRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ChiefComplaintType, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewChiefComplaintTypeService(repo)

			categories, err := svc.List(context.Background(), tt.clinicID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, categories, tt.wantLen)
			}
		})
	}
}

func TestChiefComplaintTypeService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       uint64
		repoData *model.ChiefComplaintType
		repoErr  error
		wantErr  bool
	}{
		{
			name: "returns category when found",
			id:   1,
			repoData: &model.ChiefComplaintType{
				ID:       1,
				ClinicID: 1,
				Name:     "症状Category",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when category not found",
			id:       999,
			repoData: nil,
			repoErr:  errors.New("not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ChiefComplaintType, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewChiefComplaintTypeService(repo)

			category, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, category)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, category)
				assert.Equal(t, tt.repoData, category)
			}
		})
	}
}

func TestChiefComplaintTypeService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateChiefComplaintTypeInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates category successfully",
			input: &CreateChiefComplaintTypeInput{
				Name:     "新症状",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates category with description",
			input: &CreateChiefComplaintTypeInput{
				Name:        "嘔吐",
				Description: "食後に嘔吐する",
				IsActive:    true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when repository fails",
			input: &CreateChiefComplaintTypeInput{
				Name:     "失敗症状",
				IsActive: true,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintTypeRepository{
				createFn: func(_ context.Context, _ *model.ChiefComplaintType) error {
					return tt.repoErr
				},
			}
			svc := NewChiefComplaintTypeService(repo)

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

func TestChiefComplaintTypeService_Create_ValidationError(t *testing.T) {
	repo := &mockChiefComplaintTypeRepository{
		createFn: func(_ context.Context, _ *model.ChiefComplaintType) error {
			t.Fatal("repository must not be called when name validation fails")
			return nil
		},
	}
	svc := NewChiefComplaintTypeService(repo)

	result, err := svc.Create(context.Background(), 1, &CreateChiefComplaintTypeInput{Name: "   "})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestChiefComplaintTypeService_Update(t *testing.T) {
	newName := "更新症状"
	newSortOrder := 3
	isActive := false

	tests := []struct {
		name       string
		clinicID   uint64
		categoryID uint64
		input      *UpdateChiefComplaintTypeInput
		repoData   *model.ChiefComplaintType
		repoErr    error
		wantErr    bool
	}{
		{
			name:       "updates category successfully",
			clinicID:   1,
			categoryID: 1,
			input: &UpdateChiefComplaintTypeInput{
				Name:      &newName,
				SortOrder: &newSortOrder,
			},
			repoData: &model.ChiefComplaintType{
				ID:        1,
				ClinicID:  1,
				Name:      newName,
				SortOrder: newSortOrder,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:       "returns error when no fields provided",
			clinicID:   1,
			categoryID: 1,
			input:      &UpdateChiefComplaintTypeInput{},
			repoData: &model.ChiefComplaintType{
				ID:       1,
				ClinicID: 1,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:       "returns error when update fails",
			clinicID:   1,
			categoryID: 1,
			input: &UpdateChiefComplaintTypeInput{
				IsActive: &isActive,
			},
			repoData: &model.ChiefComplaintType{
				ID:       1,
				ClinicID: 1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintTypeRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ChiefComplaintType, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewChiefComplaintTypeService(repo)

			category, err := svc.Update(context.Background(), tt.clinicID, tt.categoryID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, category)
			}
		})
	}
}

func TestChiefComplaintTypeService_Update_FindByIDError(t *testing.T) {
	repo := &mockChiefComplaintTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ChiefComplaintType, error) {
			return nil, apperrors.WrapNotFound("chief_complaint_type", "999")
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ChiefComplaintType, error) {
			t.Fatal("repository Update must not be called when the category is not found")
			return nil, nil
		},
	}
	svc := NewChiefComplaintTypeService(repo)

	newName := "更新症状"
	result, err := svc.Update(context.Background(), 1, 999, &UpdateChiefComplaintTypeInput{Name: &newName})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestChiefComplaintTypeService_Update_InvalidName(t *testing.T) {
	repo := &mockChiefComplaintTypeRepository{
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ChiefComplaintType, error) {
			t.Fatal("repository Update must not be called when name validation fails")
			return nil, nil
		},
	}
	svc := NewChiefComplaintTypeService(repo)

	blankName := "   "
	result, err := svc.Update(context.Background(), 1, 1, &UpdateChiefComplaintTypeInput{Name: &blankName})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestChiefComplaintTypeService_Update_NilInput(t *testing.T) {
	repo := &mockChiefComplaintTypeRepository{}
	svc := NewChiefComplaintTypeService(repo)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestChiefComplaintTypeService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		inquiryCount int64
		inquiryErr   error
		repoErr      error
		wantErr      bool
		wantConflict bool
	}{
		{
			name:         "deletes category successfully",
			id:           1,
			inquiryCount: 0,
			inquiryErr:   nil,
			repoErr:      nil,
			wantErr:      false,
		},
		{
			name:         "returns conflict when inquiries reference category",
			id:           2,
			inquiryCount: 3,
			inquiryErr:   nil,
			repoErr:      nil,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:         "returns error when inquiry count check fails",
			id:           3,
			inquiryCount: 0,
			inquiryErr:   errors.New("db error"),
			repoErr:      nil,
			wantErr:      true,
		},
		{
			name:         "returns error when category not found",
			id:           999,
			inquiryCount: 0,
			inquiryErr:   nil,
			repoErr:      errors.New("not found"),
			wantErr:      true,
		},
		{
			name:         "returns error when repository delete fails",
			id:           4,
			inquiryCount: 0,
			inquiryErr:   nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintTypeRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.ChiefComplaintType, error) {
					// id=999 は「not found」ケース — FindByID でエラーを返す
					if id == 999 {
						return nil, tt.repoErr
					}
					return &model.ChiefComplaintType{ID: id}, nil
				},
				countUsageByChiefComplaintTypeFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.inquiryCount, tt.inquiryErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewChiefComplaintTypeService(repo)

			err := svc.Delete(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, errors.Is(err, apperrors.ErrConflict),
						"expected conflict error, got: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChiefComplaintTypeService_Reorder(t *testing.T) {
	tests := []struct {
		name             string
		ids              []uint64
		repoErr          error
		wantErr          bool
		wantInvalidInput bool
	}{
		{
			name: "reorders successfully",
			ids:  []uint64{3, 1, 2},
		},
		{
			name:             "returns invalid input when ids is empty",
			ids:              []uint64{},
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintTypeRepository{reorderErr: tt.repoErr}
			svc := NewChiefComplaintTypeService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
