package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- HospitalizationPlan モック ----

type mockHospitalizationPlanRepository struct {
	findAllFn                    func(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error)
	findByIDFn                   func(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error)
	createFn                     func(ctx context.Context, plan *model.HospitalizationPlan) error
	updateFn                     func(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error)
	deleteFn                     func(ctx context.Context, clinicID, id uint64) error
	reorderFn                    func(ctx context.Context, clinicID uint64, ids []uint64) error
	countCarePlanItemsByPlanIDFn func(ctx context.Context, clinicID, planID uint64) (int64, error)
}

func (m *mockHospitalizationPlanRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockHospitalizationPlanRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.HospitalizationPlan{ID: id, ClinicID: clinicID}, nil
}

func (m *mockHospitalizationPlanRepository) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	return m.createFn(ctx, plan)
}

func (m *mockHospitalizationPlanRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
	return m.updateFn(ctx, clinicID, id, cmd)
}

func (m *mockHospitalizationPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockHospitalizationPlanRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockHospitalizationPlanRepository) CountUsageByHospitalizationPlanID(ctx context.Context, clinicID, planID uint64) (int64, error) {
	if m.countCarePlanItemsByPlanIDFn != nil {
		return m.countCarePlanItemsByPlanIDFn(ctx, clinicID, planID)
	}
	return 0, nil
}

// ---- Tests ----

func TestBuildHospitalizationPlanUpdate(t *testing.T) {
	name := "更新後プラン"
	price := int64(5000)
	isActive := true
	description := "説明文"
	bodySize := "large"
	billingUnit := "day"
	sortOrder := 3
	taxType := "excluded"
	taxRate := 0.08

	tests := []struct {
		name       string
		input      UpdateHospitalizationPlanInput
		wantFields map[string]any
	}{
		{
			name:       "no fields set returns empty map",
			input:      UpdateHospitalizationPlanInput{},
			wantFields: map[string]any{},
		},
		{
			name: "all fields set are reflected with real column names",
			input: UpdateHospitalizationPlanInput{
				Name:        &name,
				Price:       &price,
				IsActive:    &isActive,
				Description: &description,
				BodySize:    &bodySize,
				BillingUnit: &billingUnit,
				SortOrder:   &sortOrder,
				TaxType:     &taxType,
				TaxRate:     &taxRate,
			},
			wantFields: map[string]any{
				"name":         name,
				"price":        price,
				"is_active":    isActive,
				"description":  description,
				"body_size":    model.BodySize(bodySize),
				"billing_unit": model.BillingUnit(billingUnit),
				"sort_order":   sortOrder,
				"tax_type":     model.TaxType(taxType),
				"tax_rate":     taxRate,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := buildHospitalizationPlanUpdate(tt.input)
			assert.Equal(t, tt.wantFields, fields)
		})
	}
}

func TestHospitalizationPlanService_List(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		repoPlan []model.HospitalizationPlan
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns hospitalization plans for clinic",
			clinicID: 1,
			repoPlan: []model.HospitalizationPlan{
				{ID: 1, ClinicID: 1, Name: "Plan 1", IsActive: true},
				{ID: 2, ClinicID: 1, Name: "Plan 2", IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no plans exist",
			clinicID: 999,
			repoPlan: []model.HospitalizationPlan{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			clinicID: 1,
			repoPlan: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.HospitalizationPlan, error) {
					return tt.repoPlan, tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			plans, err := svc.List(context.Background(), tt.clinicID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, plans, tt.wantLen)
			}
		})
	}
}

func TestHospitalizationPlanService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       uint64
		repoPlan *model.HospitalizationPlan
		repoErr  error
		wantErr  bool
	}{
		{
			name: "returns plan when found",
			id:   1,
			repoPlan: &model.HospitalizationPlan{
				ID:       1,
				ClinicID: 1,
				Name:     "Hospitalization Plan",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when plan not found",
			id:       999,
			repoPlan: nil,
			repoErr:  errors.New("not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.HospitalizationPlan, error) {
					return tt.repoPlan, tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			plan, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
				assert.Equal(t, tt.repoPlan, plan)
			}
		})
	}
}

func TestHospitalizationPlanService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateHospitalizationPlanInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates hospitalization plan successfully",
			input: &CreateHospitalizationPlanInput{
				Name:     "New Plan",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when repository fails",
			input: &CreateHospitalizationPlanInput{
				Name: "Failed Plan",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				createFn: func(_ context.Context, _ *model.HospitalizationPlan) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			plan, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}
		})
	}
}

func TestHospitalizationPlanService_Create_ValidationError(t *testing.T) {
	repo := &mockHospitalizationPlanRepository{
		createFn: func(_ context.Context, _ *model.HospitalizationPlan) error {
			t.Fatal("plan must not be created when name is empty")
			return nil
		},
	}
	svc := NewHospitalizationPlanService(repo)

	plan, err := svc.Create(context.Background(), 1, &CreateHospitalizationPlanInput{Name: "  "})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, plan)
}

func TestHospitalizationPlanService_Create_WithBodySizeAndBillingUnit(t *testing.T) {
	taxRate := 0.08
	var captured *model.HospitalizationPlan
	repo := &mockHospitalizationPlanRepository{
		createFn: func(_ context.Context, plan *model.HospitalizationPlan) error {
			captured = plan
			return nil
		},
	}
	svc := NewHospitalizationPlanService(repo)

	plan, err := svc.Create(context.Background(), 1, &CreateHospitalizationPlanInput{
		Name:        "個室プラン",
		BodySize:    "large",
		BillingUnit: "day",
		TaxType:     "included",
		TaxRate:     &taxRate,
	})

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.NotNil(t, captured.BodySize)
	assert.Equal(t, model.BodySize("large"), *captured.BodySize)
	assert.NotNil(t, captured.BillingUnit)
	assert.Equal(t, model.BillingUnit("day"), *captured.BillingUnit)
	assert.Equal(t, model.TaxType("included"), captured.TaxType)
	assert.Equal(t, taxRate, captured.TaxRate)
}

func TestHospitalizationPlanService_Update(t *testing.T) {
	name := "Updated Plan"
	isActive := false
	tests := []struct {
		name    string
		input   UpdateHospitalizationPlanInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates hospitalization plan successfully",
			input: UpdateHospitalizationPlanInput{
				Name:     &name,
				IsActive: &isActive,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateHospitalizationPlanInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when repository fails",
			input: UpdateHospitalizationPlanInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.HospitalizationPlan{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewHospitalizationPlanService(repo)

			plan, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}
		})
	}
}

func TestHospitalizationPlanService_Update_InputNil(t *testing.T) {
	repo := &mockHospitalizationPlanRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.HospitalizationPlan, error) {
			t.Fatal("plan must not be looked up when input is nil")
			return nil, nil
		},
	}
	svc := NewHospitalizationPlanService(repo)

	plan, err := svc.Update(context.Background(), 1, 1, nil)

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, plan)
}

func TestHospitalizationPlanService_Update_ParentNotFound(t *testing.T) {
	name := "更新後プラン"
	repo := &mockHospitalizationPlanRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.HospitalizationPlan, error) {
			return nil, apperrors.WrapNotFound("hospitalization_plan", "999")
		},
	}
	svc := NewHospitalizationPlanService(repo)

	plan, err := svc.Update(context.Background(), 1, 999, &UpdateHospitalizationPlanInput{Name: &name})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, plan)
}

func TestHospitalizationPlanService_Update_InvalidName(t *testing.T) {
	blank := "   "
	repo := &mockHospitalizationPlanRepository{
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
			t.Fatal("plan must not be updated when name fails validation")
			return nil, nil
		},
	}
	svc := NewHospitalizationPlanService(repo)

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationPlanInput{Name: &blank})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, plan)
}

func TestHospitalizationPlanService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		findErr       error
		carePlanCount int64
		countErr      error
		repoErr       error
		wantErr       bool
		wantConflict  bool
	}{
		{
			name:    "deletes hospitalization plan successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when plan not found",
			id:      999,
			repoErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:          "使用中の入院プランは削除できない",
			id:            2,
			carePlanCount: 3,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:    "returns error when the parent lookup fails",
			id:      3,
			findErr: apperrors.WrapNotFound("hospitalization_plan", "3"),
			wantErr: true,
		},
		{
			name:     "returns error when usage count check fails",
			id:       4,
			countErr: errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.HospitalizationPlan{ID: id, ClinicID: clinicID}, nil
				},
				countCarePlanItemsByPlanIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.carePlanCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			err := svc.Delete(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHospitalizationPlanService_Reorder(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		ids      []uint64
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "reorders plans successfully",
			clinicID: 1,
			ids:      []uint64{2, 1, 3},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns error when repository fails",
			clinicID: 1,
			ids:      []uint64{1, 2},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
		{
			name:     "returns error when ids is empty",
			clinicID: 1,
			ids:      []uint64{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			err := svc.Reorder(context.Background(), tt.clinicID, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
