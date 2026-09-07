package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Procedure モック ----

type mockProcedureRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.Procedure, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
	createFn                  func(ctx context.Context, procedure *model.Procedure) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, cmd UpdateProcedureInput) (*model.Procedure, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	countUsageByProcedureIDFn func(ctx context.Context, clinicID, procedureID uint64) (int64, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockProcedureRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockProcedureRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockProcedureRepository) Create(ctx context.Context, procedure *model.Procedure) error {
	return m.createFn(ctx, procedure)
}

func (m *mockProcedureRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateProcedureInput) (*model.Procedure, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, cmd)
	}
	return &model.Procedure{ID: id}, nil
}

func (m *mockProcedureRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockProcedureRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockProcedureRepository) CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error) {
	if m.countUsageByProcedureIDFn == nil {
		return 0, nil
	}
	return m.countUsageByProcedureIDFn(ctx, clinicID, procedureID)
}

func (m *mockProcedureRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

// ---- Tests ----

func TestBuildProcedureUpdate(t *testing.T) {
	name := "処置A"
	price := int64(3000)
	isActive := true
	description := "説明"
	duration := 30
	anesthesia := string(model.AnesthesiaTypeLocal)
	parentID := uint64(5)
	sortOrder := 2
	taxType := string(model.TaxTypeIncluded)
	taxRate := 0.08

	tests := []struct {
		name       string
		input      *UpdateProcedureInput
		wantFields map[string]any
	}{
		{
			name:       "empty input produces empty fields",
			input:      &UpdateProcedureInput{},
			wantFields: map[string]any{},
		},
		{
			name: "all fields set produces full field map",
			input: &UpdateProcedureInput{
				Name:        &name,
				Price:       &price,
				IsActive:    &isActive,
				Description: &description,
				Duration:    &duration,
				Anesthesia:  &anesthesia,
				ParentID:    &parentID,
				SortOrder:   &sortOrder,
				TaxType:     &taxType,
				TaxRate:     &taxRate,
			},
			wantFields: map[string]any{
				colProcedureName:        name,
				colProcedurePrice:       price,
				colProcedureIsActive:    isActive,
				colProcedureDescription: description,
				colProcedureDuration:    duration,
				colProcedureAnesthesia:  model.AnesthesiaType(anesthesia),
				colProcedureParentID:    parentID,
				colProcedureSortOrder:   sortOrder,
				colProcedureTaxType:     model.TaxType(taxType),
				colProcedureTaxRate:     taxRate,
			},
		},
		{
			name: "ClearParentID clears parent_id to nil",
			input: &UpdateProcedureInput{
				ClearParentID: true,
			},
			wantFields: map[string]any{
				colProcedureParentID: nil,
			},
		},
		{
			name: "only name set produces single field",
			input: &UpdateProcedureInput{
				Name: &name,
			},
			wantFields: map[string]any{colProcedureName: name},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildProcedureUpdate(tt.input)
			assert.Equal(t, tt.wantFields, got)
		})
	}
}

func TestProcedureService_List(t *testing.T) {
	tests := []struct {
		name           string
		repoProcedures []model.Procedure
		repoErr        error
		wantLen        int
		wantErr        bool
	}{
		{
			name: "returns all procedures",
			repoProcedures: []model.Procedure{
				{ID: 1, Name: "Surgery A", IsActive: true},
				{ID: 2, Name: "Surgery B", IsActive: true},
				{ID: 3, Name: "Examination", IsActive: true},
			},
			repoErr: nil,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:           "returns empty list when no procedures exist",
			repoProcedures: []model.Procedure{},
			repoErr:        nil,
			wantLen:        0,
			wantErr:        false,
		},
		{
			name:           "propagates repository error",
			repoProcedures: nil,
			repoErr:        errors.New("db error"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Procedure, error) {
					return tt.repoProcedures, tt.repoErr
				},
			}
			svc := NewProcedureService(repo, &mockTransactor{})

			procedures, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, procedures, tt.wantLen)
			}
		})
	}
}

func TestProcedureService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		repoProcedure *model.Procedure
		repoErr       error
		wantErr       bool
	}{
		{
			name: "returns procedure when found",
			id:   1,
			repoProcedure: &model.Procedure{
				ID:       1,
				Name:     "Surgical Procedure",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:          "returns error when procedure not found",
			id:            999,
			repoProcedure: nil,
			repoErr:       errors.New("not found"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Procedure, error) {
					return tt.repoProcedure, tt.repoErr
				},
			}
			svc := NewProcedureService(repo, &mockTransactor{})

			procedure, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, procedure)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, procedure)
				assert.Equal(t, tt.repoProcedure, procedure)
			}
		})
	}
}

func TestProcedureService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateProcedureInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates procedure successfully",
			input: &CreateProcedureInput{
				Name:     "New Procedure",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates procedure with tax defaults applied",
			input: &CreateProcedureInput{
				Name: "Default Tax Procedure",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when repository fails",
			input: &CreateProcedureInput{
				Name: "Failed Procedure",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "returns validation error when name is empty",
			input: &CreateProcedureInput{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "returns validation error when price is negative",
			input: &CreateProcedureInput{
				Name:  "Negative Price Procedure",
				Price: func(v int64) *int64 { return &v }(-100),
			},
			wantErr: true,
		},
		{
			name: "returns validation error when anesthesia type is invalid",
			input: &CreateProcedureInput{
				Name:       "Invalid Anesthesia Procedure",
				Anesthesia: "invalid_type",
			},
			wantErr: true,
		},
		{
			name: "returns validation error when tax type is invalid",
			input: &CreateProcedureInput{
				Name:    "Invalid Tax Procedure",
				TaxType: "invalid_tax",
			},
			wantErr: true,
		},
		{
			name: "creates procedure with anesthesia, parent and custom tax settings",
			input: &CreateProcedureInput{
				Name:       "Full Option Procedure",
				Anesthesia: string(model.AnesthesiaTypeGeneral),
				ParentID:   func(v uint64) *uint64 { return &v }(7),
				TaxType:    string(model.TaxTypeIncluded),
				TaxRate:    func(v float64) *float64 { return &v }(0.08),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				createFn: func(_ context.Context, _ *model.Procedure) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
					return &model.Procedure{ID: id}, nil
				},
			}
			svc := NewProcedureService(repo, &mockTransactor{})

			procedure, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, procedure)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, procedure)
			}
		})
	}
}

func TestProcedureService_Update(t *testing.T) {
	name := "Updated Procedure"
	isActive := false
	tests := []struct {
		name        string
		input       UpdateProcedureInput
		repoErr     error
		findByIDErr error
		wantErr     bool
		wantNF      bool
	}{
		{
			name: "updates procedure successfully",
			input: UpdateProcedureInput{
				Name:     &name,
				IsActive: &isActive,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateProcedureInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns not found error when procedure does not exist",
			input: UpdateProcedureInput{
				Name: &name,
			},
			repoErr: apperrors.WrapNotFound("procedure", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error when repository fails",
			input: UpdateProcedureInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "returns error when FindByID fails before update",
			input: UpdateProcedureInput{
				Name: &name,
			},
			findByIDErr: apperrors.WrapNotFound("procedure", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name: "returns validation error when name is whitespace only",
			input: UpdateProcedureInput{
				Name: func(s string) *string { return &s }("   "),
			},
			wantErr: true,
		},
		{
			name: "returns validation error when price is negative",
			input: UpdateProcedureInput{
				Price: func(v int64) *int64 { return &v }(-500),
			},
			wantErr: true,
		},
		{
			name: "returns validation error when anesthesia type is invalid",
			input: UpdateProcedureInput{
				Anesthesia: func(s string) *string { return &s }("bogus"),
			},
			wantErr: true,
		},
		{
			name: "returns validation error when tax type is invalid",
			input: UpdateProcedureInput{
				TaxType: func(s string) *string { return &s }("bogus_tax"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Procedure{ID: id}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateProcedureInput) (*model.Procedure, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Procedure{ID: 1}, nil
				},
			}
			svc := NewProcedureService(repo, &mockTransactor{})

			procedure, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, procedure)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, procedure)
			}
		})
	}
}

func TestProcedureService_Update_NilInput(t *testing.T) {
	repo := &mockProcedureRepository{}
	svc := NewProcedureService(repo, &mockTransactor{})
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestProcedureService_Delete(t *testing.T) {
	tests := []struct {
		name             string
		id               uint64
		childCount       int64
		countChildrenErr error
		usageCount       int64
		countUsageErr    error
		findByIDErr      error
		repoErr          error
		wantErr          bool
		wantNotFound     bool
		wantConflict     bool
	}{
		{
			name:             "deletes procedure successfully when no children and no medical records use it",
			id:               1,
			childCount:       0,
			countChildrenErr: nil,
			usageCount:       0,
			countUsageErr:    nil,
			repoErr:          nil,
			wantErr:          false,
		},
		{
			name:             "returns conflict error when procedure has children (BUG-390)",
			id:               1,
			childCount:       3,
			countChildrenErr: nil,
			usageCount:       0,
			countUsageErr:    nil,
			repoErr:          nil,
			wantErr:          true,
			wantConflict:     true,
		},
		{
			name:             "returns error when children count check fails (BUG-390)",
			id:               1,
			childCount:       0,
			countChildrenErr: errors.New("db error"),
			usageCount:       0,
			countUsageErr:    nil,
			repoErr:          nil,
			wantErr:          true,
		},
		{
			name:             "returns conflict error when procedure is used in medical records",
			id:               1,
			childCount:       0,
			countChildrenErr: nil,
			usageCount:       2,
			countUsageErr:    nil,
			repoErr:          nil,
			wantErr:          true,
			wantConflict:     true,
		},
		{
			name:             "returns error when usage count check fails",
			id:               1,
			childCount:       0,
			countChildrenErr: nil,
			usageCount:       0,
			countUsageErr:    errors.New("db error"),
			repoErr:          nil,
			wantErr:          true,
		},
		{
			name:         "returns not found error when procedure does not exist",
			id:           999,
			findByIDErr:  apperrors.WrapNotFound("procedure", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:             "returns error on repository failure",
			id:               1,
			childCount:       0,
			countChildrenErr: nil,
			usageCount:       0,
			countUsageErr:    nil,
			repoErr:          errors.New("db error"),
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Procedure{ID: id}, nil
				},
				countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.childCount, tt.countChildrenErr
				},
				countUsageByProcedureIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, tt.countUsageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewProcedureService(repo, &mockTransactor{})

			err := svc.Delete(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
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

func TestProcedureService_Reorder(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		ids      []uint64
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "reorders procedures successfully",
			clinicID: 1,
			ids:      []uint64{3, 1, 2},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns error when ids is empty",
			clinicID: 1,
			ids:      []uint64{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error when repository fails",
			clinicID: 1,
			ids:      []uint64{1, 2},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := NewProcedureService(repo, &mockTransactor{})

			err := svc.Reorder(context.Background(), tt.clinicID, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				// Validate empty ids error is about invalid input
				if len(tt.ids) == 0 {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
