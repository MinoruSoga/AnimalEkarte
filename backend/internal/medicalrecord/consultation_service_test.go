package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Consultation モック ----

type mockConsultationRepository struct {
	findAllFn                    func(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	findByIDFn                   func(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	createFn                     func(ctx context.Context, consultation *model.Consultation) error
	updateFieldsFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error)
	deleteFn                     func(ctx context.Context, clinicID, id uint64) error
	countUsageByConsultationIDFn func(ctx context.Context, clinicID, consultationID uint64) (int64, error)
	countChildrenByParentIDFn    func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	reorderErr                   error
}

func (m *mockConsultationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockConsultationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockConsultationRepository) Create(ctx context.Context, consultation *model.Consultation) error {
	return m.createFn(ctx, consultation)
}

func (m *mockConsultationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockConsultationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockConsultationRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockConsultationRepository) CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error) {
	if m.countUsageByConsultationIDFn == nil {
		return 0, nil
	}
	return m.countUsageByConsultationIDFn(ctx, clinicID, consultationID)
}

func (m *mockConsultationRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

// ---- Tests ----

func TestBuildConsultationUpdate(t *testing.T) {
	name := "更新名"
	price := int64(3000)
	isActive := true
	description := "説明"
	timeCondition := "平日"
	duration := 30
	parentID := uint64(1)
	sortOrder := 2
	taxType := "included"
	taxRate := 0.08

	tests := []struct {
		name  string
		input *UpdateConsultationInput
		want  map[string]any
	}{
		{
			name:  "no fields set returns empty map",
			input: &UpdateConsultationInput{},
			want:  map[string]any{},
		},
		{
			name:  "only name set",
			input: &UpdateConsultationInput{Name: &name},
			want:  map[string]any{colConsultationName: name},
		},
		{
			name:  "only price set",
			input: &UpdateConsultationInput{Price: &price},
			want:  map[string]any{colConsultationPrice: price},
		},
		{
			name:  "only is_active set",
			input: &UpdateConsultationInput{IsActive: &isActive},
			want:  map[string]any{colConsultationIsActive: isActive},
		},
		{
			name:  "only description set",
			input: &UpdateConsultationInput{Description: &description},
			want:  map[string]any{colConsultationDescription: description},
		},
		{
			name:  "only time_condition set",
			input: &UpdateConsultationInput{TimeCondition: &timeCondition},
			want:  map[string]any{colConsultationTimeCondition: timeCondition},
		},
		{
			name:  "only duration set",
			input: &UpdateConsultationInput{Duration: &duration},
			want:  map[string]any{colConsultationDuration: duration},
		},
		{
			name:  "parent_id set writes value",
			input: &UpdateConsultationInput{ParentID: &parentID},
			want:  map[string]any{colConsultationParentID: parentID},
		},
		{
			name:  "clear_parent_id writes explicit NULL even with parent_id set",
			input: &UpdateConsultationInput{ParentID: &parentID, ClearParentID: true},
			want:  map[string]any{colConsultationParentID: nil},
		},
		{
			name:  "only sort_order set",
			input: &UpdateConsultationInput{SortOrder: &sortOrder},
			want:  map[string]any{colConsultationSortOrder: sortOrder},
		},
		{
			name:  "only tax_type set",
			input: &UpdateConsultationInput{TaxType: &taxType},
			want:  map[string]any{colConsultationTaxType: model.TaxType(taxType)},
		},
		{
			name:  "only tax_rate set",
			input: &UpdateConsultationInput{TaxRate: &taxRate},
			want:  map[string]any{colConsultationTaxRate: taxRate},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildConsultationUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConsultationService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.Consultation
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns consultations list",
			repoData: []model.Consultation{
				{ID: 1, ClinicID: 1, Name: "相談1", SortOrder: 1, IsActive: true},
				{ID: 2, ClinicID: 1, Name: "相談2", SortOrder: 2, IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no consultations exist",
			repoData: []model.Consultation{},
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
			repo := &mockConsultationRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Consultation, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

			consultations, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, consultations, tt.wantLen)
			}
		})
	}
}

func TestConsultationService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoData     *model.Consultation
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "returns consultation when found",
			id:   1,
			repoData: &model.Consultation{
				ID:        1,
				ClinicID:  1,
				Name:      "相談1",
				SortOrder: 1,
				IsActive:  true,
			},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when consultation does not exist",
			id:           999,
			repoData:     nil,
			repoErr:      apperrors.WrapNotFound("consultation", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoData:     nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Consultation, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

			consultation, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, consultation)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoData, consultation)
			}
		})
	}
}

func TestConsultationService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateConsultationInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates consultation successfully",
			input: &CreateConsultationInput{
				Name:      "新規相談",
				SortOrder: 3,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "applies default tax type and rate",
			input: &CreateConsultationInput{
				Name: "デフォルト税率相談",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when consultation already exists",
			input: &CreateConsultationInput{
				Name: "既存相談",
			},
			repoErr: apperrors.WrapAlreadyExists("consultation", "既存相談"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateConsultationInput{
				Name: "エラー相談",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "returns validation error when name is blank",
			input:   &CreateConsultationInput{Name: ""},
			wantErr: true,
		},
		{
			name: "returns validation error when price is negative",
			input: &CreateConsultationInput{
				Name:  "Negative Price Consultation",
				Price: func(v int64) *int64 { return &v }(-100),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalled := false
			repo := &mockConsultationRepository{
				createFn: func(_ context.Context, _ *model.Consultation) error {
					createCalled = true
					return tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

			consultation, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, consultation)
				if tt.name == "returns validation error when price is negative" {
					assert.True(t, apperrors.IsInvalidInput(err))
					assert.False(t, createCalled)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, consultation)
			}
		})
	}
}

// TestConsultationService_Create_TaxDefaults は tax_type/tax_rate のデフォルト・
// カスタム指定・空文字指定の各分岐を検証する（消費税表示の正確性に直結するため個別テスト）。
func TestConsultationService_Create_TaxDefaults(t *testing.T) {
	tests := []struct {
		name        string
		input       *CreateConsultationInput
		wantTaxType model.TaxType
		wantTaxRate float64
	}{
		{
			name:        "nil TaxType/TaxRate use defaults",
			input:       &CreateConsultationInput{Name: "デフォルト"},
			wantTaxType: model.TaxTypeExcluded,
			wantTaxRate: 0.10,
		},
		{
			name:        "empty string TaxType falls back to default",
			input:       &CreateConsultationInput{Name: "空文字税区分", TaxType: strPtr("")},
			wantTaxType: model.TaxTypeExcluded,
			wantTaxRate: 0.10,
		},
		{
			name:        "custom TaxType is applied",
			input:       &CreateConsultationInput{Name: "内税相談", TaxType: strPtr("included")},
			wantTaxType: model.TaxType("included"),
			wantTaxRate: 0.10,
		},
		{
			name:        "custom TaxRate is applied",
			input:       &CreateConsultationInput{Name: "軽減税率相談", TaxRate: float64Ptr(0.08)},
			wantTaxType: model.TaxTypeExcluded,
			wantTaxRate: 0.08,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
				createFn: func(_ context.Context, _ *model.Consultation) error { return nil },
			}
			svc := NewConsultationService(repo)

			consultation, err := svc.Create(context.Background(), 1, tt.input)

			assert.NoError(t, err)
			if assert.NotNil(t, consultation) {
				assert.Equal(t, tt.wantTaxType, consultation.TaxType)
				assert.InDelta(t, tt.wantTaxRate, consultation.TaxRate, 0.0001)
			}
		})
	}
}

func TestConsultationService_Update(t *testing.T) {
	name := "更新後相談"
	isActive := true
	tests := []struct {
		name         string
		input        UpdateConsultationInput
		findByIDErr  error
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "updates consultation successfully",
			input: UpdateConsultationInput{
				Name:     &name,
				IsActive: &isActive,
			},
			findByIDErr: nil,
			repoErr:     nil,
			wantErr:     false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateConsultationInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns not found error when consultation does not exist",
			input: UpdateConsultationInput{
				Name: &name,
			},
			findByIDErr:  apperrors.WrapNotFound("consultation", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateConsultationInput{
				Name: &name,
			},
			findByIDErr: nil,
			repoErr:     errors.New("db error"),
			wantErr:     true,
		},
		{
			name:    "returns validation error when name is blank",
			input:   UpdateConsultationInput{Name: strPtr("")},
			wantErr: true,
		},
		{
			name: "returns validation error when price is negative",
			input: UpdateConsultationInput{
				Price: func(v int64) *int64 { return &v }(-500),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			repo := &mockConsultationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Consultation, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Consultation{ID: 1, ClinicID: 1}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Consultation, error) {
					updateCalled = true
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Consultation{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewConsultationService(repo)

			consultation, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, consultation)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.name == "returns validation error when price is negative" {
					assert.True(t, apperrors.IsInvalidInput(err))
					assert.False(t, updateCalled)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, consultation)
			}
		})
	}
}

func TestConsultationService_Update_NilInput(t *testing.T) {
	repo := &mockConsultationRepository{}
	svc := NewConsultationService(repo)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestConsultationService_Delete(t *testing.T) {
	tests := []struct {
		name             string
		id               uint64
		findByIDErr      error
		childCount       int64
		countChildrenErr error
		usageCount       int64
		countUsageErr    error
		repoErr          error
		wantErr          bool
		wantNF           bool
		wantConflict     bool
	}{
		{
			name:       "deletes consultation successfully when no children and no medical records",
			id:         1,
			childCount: 0,
			usageCount: 0,
			repoErr:    nil,
			wantErr:    false,
		},
		{
			name:        "returns not found error when FindByID fails",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("consultation", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:         "returns conflict error when consultation has children",
			id:           1,
			childCount:   2,
			usageCount:   0,
			repoErr:      nil,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:             "returns error when children count check fails",
			id:               1,
			countChildrenErr: errors.New("db error"),
			wantErr:          true,
		},
		{
			name:          "returns conflict error when consultation is used in medical records",
			id:            1,
			childCount:    0,
			usageCount:    2,
			countUsageErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns error when usage count check fails",
			id:            1,
			childCount:    0,
			usageCount:    0,
			countUsageErr: errors.New("db error"),
			repoErr:       nil,
			wantErr:       true,
		},
		{
			name:       "returns error on repository delete failure",
			id:         1,
			childCount: 0,
			usageCount: 0,
			repoErr:    errors.New("db error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Consultation, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Consultation{ID: id}, nil
				},
				countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.childCount, tt.countChildrenErr
				},
				countUsageByConsultationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, tt.countUsageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

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

func TestConsultationService_Reorder(t *testing.T) {
	tests := []struct {
		name    string
		ids     []uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{2, 1},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when ids is empty",
			ids:     []uint64{},
			repoErr: nil,
			wantErr: true,
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
			repo := &mockConsultationRepository{reorderErr: tt.repoErr}
			svc := NewConsultationService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
