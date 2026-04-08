package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ChiefComplaintCategory モック ----

type mockChiefComplaintCategoryRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintCategory, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.ChiefComplaintCategory, error)
	createFn   func(ctx context.Context, category *model.ChiefComplaintCategory) error
	updateFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn   func(ctx context.Context, id uint64) error
}

func (m *mockChiefComplaintCategoryRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintCategory, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockChiefComplaintCategoryRepository) FindByID(ctx context.Context, id uint64) (*model.ChiefComplaintCategory, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockChiefComplaintCategoryRepository) Create(ctx context.Context, category *model.ChiefComplaintCategory) error {
	return m.createFn(ctx, category)
}

func (m *mockChiefComplaintCategoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockChiefComplaintCategoryRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

// ---- InquiryRepository モック ----

type mockInquiryRepository struct {
	upsertFn func(ctx context.Context, inquiry *model.Inquiry) (*model.Inquiry, error)
	countFn  func(ctx context.Context, categoryID uint64) (int64, error)
}

func (m *mockInquiryRepository) UpsertByMedicalRecordID(ctx context.Context, inquiry *model.Inquiry) (*model.Inquiry, error) {
	return m.upsertFn(ctx, inquiry)
}

func (m *mockInquiryRepository) CountByChiefComplaintCategoryID(ctx context.Context, categoryID uint64) (int64, error) {
	return m.countFn(ctx, categoryID)
}

// ---- Tests ----

func TestChiefComplaintCategoryService_List(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		repoData []model.ChiefComplaintCategory
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns all categories for clinic",
			clinicID: 1,
			repoData: []model.ChiefComplaintCategory{
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
			repoData: []model.ChiefComplaintCategory{},
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
			repo := &mockChiefComplaintCategoryRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ChiefComplaintCategory, error) {
					return tt.repoData, tt.repoErr
				},
			}
			inquiryRepo := &mockInquiryRepository{}
			svc := NewChiefComplaintCategoryService(repo, inquiryRepo)

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

func TestChiefComplaintCategoryService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       uint64
		repoData *model.ChiefComplaintCategory
		repoErr  error
		wantErr  bool
	}{
		{
			name: "returns category when found",
			id:   1,
			repoData: &model.ChiefComplaintCategory{
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
			repo := &mockChiefComplaintCategoryRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.ChiefComplaintCategory, error) {
					return tt.repoData, tt.repoErr
				},
			}
			inquiryRepo := &mockInquiryRepository{}
			svc := NewChiefComplaintCategoryService(repo, inquiryRepo)

			category, err := svc.GetByID(context.Background(), tt.id)

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

func TestChiefComplaintCategoryService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *model.ChiefComplaintCategory
		repoErr error
		wantErr bool
	}{
		{
			name: "creates category successfully",
			input: &model.ChiefComplaintCategory{
				ClinicID: 1,
				Name:     "新症状",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when repository fails",
			input: &model.ChiefComplaintCategory{
				ClinicID: 1,
				Name:     "失敗症状",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintCategoryRepository{
				createFn: func(_ context.Context, _ *model.ChiefComplaintCategory) error {
					return tt.repoErr
				},
			}
			inquiryRepo := &mockInquiryRepository{}
			svc := NewChiefComplaintCategoryService(repo, inquiryRepo)

			err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChiefComplaintCategoryService_Update(t *testing.T) {
	newName := "更新症状"
	newSortOrder := 3
	isActive := false

	tests := []struct {
		name       string
		clinicID   uint64
		categoryID uint64
		input      *UpdateChiefComplaintCategoryInput
		repoData   *model.ChiefComplaintCategory
		repoErr    error
		wantErr    bool
	}{
		{
			name:       "updates category successfully",
			clinicID:   1,
			categoryID: 1,
			input: &UpdateChiefComplaintCategoryInput{
				Name:      &newName,
				SortOrder: &newSortOrder,
			},
			repoData: &model.ChiefComplaintCategory{
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
			input:      &UpdateChiefComplaintCategoryInput{},
			repoData: &model.ChiefComplaintCategory{
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
			input: &UpdateChiefComplaintCategoryInput{
				IsActive: &isActive,
			},
			repoData: &model.ChiefComplaintCategory{
				ID:       1,
				ClinicID: 1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintCategoryRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _ uint64) (*model.ChiefComplaintCategory, error) {
					return tt.repoData, nil
				},
			}
			inquiryRepo := &mockInquiryRepository{}
			svc := NewChiefComplaintCategoryService(repo, inquiryRepo)

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

func TestChiefComplaintCategoryService_Delete(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockChiefComplaintCategoryRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			inquiryRepo := &mockInquiryRepository{
				countFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.inquiryCount, tt.inquiryErr
				},
			}
			svc := NewChiefComplaintCategoryService(repo, inquiryRepo)

			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, errors.Is(err, apperrors.ErrAlreadyExists),
						"expected conflict error, got: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
