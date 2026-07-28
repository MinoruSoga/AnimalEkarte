package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- InquiryTemplate モック ----

type mockInquiryTemplateRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error)
	createFn       func(ctx context.Context, template *model.InquiryTemplate) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InquiryTemplate, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
	reorderErr     error
	countUsageFn   func(ctx context.Context, clinicID, id uint64) (int64, error)
}

func (m *mockInquiryTemplateRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockInquiryTemplateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.InquiryTemplate{ID: id, ClinicID: clinicID}, nil
}

func (m *mockInquiryTemplateRepository) Create(ctx context.Context, template *model.InquiryTemplate) error {
	return m.createFn(ctx, template)
}

func (m *mockInquiryTemplateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InquiryTemplate, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockInquiryTemplateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockInquiryTemplateRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockInquiryTemplateRepository) CountUsageByInquiryTemplateID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageFn != nil {
		return m.countUsageFn(ctx, clinicID, id)
	}
	return 0, nil
}

// ---- Tests ----

func TestInquiryTemplateService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.InquiryTemplate
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns inquiry templates list",
			repoData: []model.InquiryTemplate{
				{ID: 1, ClinicID: 1, Category: "問い合わせ", Title: "テンプレート1", IsActive: true},
				{ID: 2, ClinicID: 1, Category: "予約", Title: "テンプレート2", IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no templates exist",
			repoData: []model.InquiryTemplate{},
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
			repo := &mockInquiryTemplateRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.InquiryTemplate, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewInquiryTemplateService(repo)

			templates, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, templates, tt.wantLen)
			}
		})
	}
}

func TestInquiryTemplateService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoTemplate *model.InquiryTemplate
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "returns inquiry template when found",
			id:   1,
			repoTemplate: &model.InquiryTemplate{
				ID:       1,
				ClinicID: 1,
				Category: "問い合わせ",
				Title:    "テンプレート1",
				Content:  "テンプレートの内容",
				IsActive: true,
			},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when template does not exist",
			id:           999,
			repoTemplate: nil,
			repoErr:      apperrors.WrapNotFound("inquiry_template", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoTemplate: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInquiryTemplateRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.InquiryTemplate, error) {
					return tt.repoTemplate, tt.repoErr
				},
			}
			svc := NewInquiryTemplateService(repo)

			template, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, template)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoTemplate, template)
			}
		})
	}
}

func TestInquiryTemplateService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateInquiryTemplateInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates inquiry template successfully",
			input: &CreateInquiryTemplateInput{
				Category: "新規カテゴリ",
				Title:    "新規テンプレート",
				Content:  "テンプレートの内容",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when template already exists",
			input: &CreateInquiryTemplateInput{
				Title: "既存テンプレート",
			},
			repoErr: apperrors.WrapAlreadyExists("inquiry_template", "既存テンプレート"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateInquiryTemplateInput{
				Title: "エラーテンプレート",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "returns validation error when title is empty",
			input: &CreateInquiryTemplateInput{
				Title: "",
			},
			wantErr: true,
		},
		{
			name: "returns validation error when title is whitespace only",
			input: &CreateInquiryTemplateInput{
				Title: "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInquiryTemplateRepository{
				createFn: func(_ context.Context, _ *model.InquiryTemplate) error {
					return tt.repoErr
				},
			}
			svc := NewInquiryTemplateService(repo)

			tmpl, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tmpl)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)
			}
		})
	}
}

func TestInquiryTemplateService_Update(t *testing.T) {
	updatedTitle := "更新テンプレート"
	updatedCategory := "更新カテゴリ"
	updatedSortOrder := 5

	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		input        *UpdateInquiryTemplateInput
		repoTemplate *model.InquiryTemplate
		repoErr      error
		findByIDErr  error
		wantErr      bool
	}{
		{
			name:     "updates inquiry template successfully",
			clinicID: 1,
			id:       1,
			input: &UpdateInquiryTemplateInput{
				Title:     &updatedTitle,
				Category:  &updatedCategory,
				SortOrder: &updatedSortOrder,
			},
			repoTemplate: &model.InquiryTemplate{
				ID:        1,
				ClinicID:  1,
				Title:     "更新テンプレート",
				Category:  "更新カテゴリ",
				SortOrder: 5,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when no fields provided",
			clinicID: 1,
			id:       1,
			input:    &UpdateInquiryTemplateInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns not found error when template does not exist",
			clinicID: 1,
			id:       999,
			input: &UpdateInquiryTemplateInput{
				Title: &updatedTitle,
			},
			repoErr: apperrors.WrapNotFound("inquiry_template", "999"),
			wantErr: true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input: &UpdateInquiryTemplateInput{
				Title: &updatedTitle,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:     "updates only active flag",
			clinicID: 1,
			id:       1,
			input: &UpdateInquiryTemplateInput{
				IsActive: func(b bool) *bool { return &b }(false),
			},
			repoTemplate: &model.InquiryTemplate{
				ID:       1,
				ClinicID: 1,
				IsActive: false,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when FindByID fails before update",
			clinicID: 1,
			id:       999,
			input: &UpdateInquiryTemplateInput{
				Title: &updatedTitle,
			},
			findByIDErr: apperrors.WrapNotFound("inquiry_template", "999"),
			wantErr:     true,
		},
		{
			name:     "returns validation error when title is whitespace only",
			clinicID: 1,
			id:       1,
			input: &UpdateInquiryTemplateInput{
				Title: func(s string) *string { return &s }("   "),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInquiryTemplateRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.InquiryTemplate{ID: id, ClinicID: clinicID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.InquiryTemplate, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.repoTemplate, nil
				},
			}
			svc := NewInquiryTemplateService(repo)

			template, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoTemplate, template)
			}
		})
	}
}

func TestInquiryTemplateService_Update_NilInput(t *testing.T) {
	repo := &mockInquiryTemplateRepository{}
	svc := NewInquiryTemplateService(repo)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

func TestInquiryTemplateService_Reorder(t *testing.T) {
	tests := []struct {
		name       string
		ids        []uint64
		reorderErr error
		wantErr    bool
	}{
		{
			name:    "reorders templates successfully",
			ids:     []uint64{3, 1, 2},
			wantErr: false,
		},
		{
			name:    "returns error for empty ids",
			ids:     []uint64{},
			wantErr: true,
		},
		{
			name:       "propagates repository error",
			ids:        []uint64{1, 2},
			reorderErr: errors.New("reorder failed"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInquiryTemplateRepository{
				reorderErr: tt.reorderErr,
			}
			svc := NewInquiryTemplateService(repo)
			err := svc.Reorder(context.Background(), 1, tt.ids)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInquiryTemplateService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		findByIDErr  error
		usageCount   int64
		usageErr     error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:    "deletes inquiry template successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when template does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("inquiry_template", "999"),
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
		{
			name:        "returns not found error when FindByID fails",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("inquiry_template", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:       "returns error when usage count check fails",
			id:         1,
			usageCount: 0,
			usageErr:   errors.New("db error"),
			wantErr:    true,
		},
		{
			name:         "returns conflict error when template is in use",
			id:           1,
			usageCount:   3,
			wantErr:      true,
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInquiryTemplateRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.InquiryTemplate{ID: id, ClinicID: clinicID}, nil
				},
				countUsageFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, tt.usageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewInquiryTemplateService(repo)

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
