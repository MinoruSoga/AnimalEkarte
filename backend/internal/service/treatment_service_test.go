package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- Treatment モック ----

type mockTreatmentRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, medicalRecordID uint64) ([]model.Treatment, error)
	findByIDFn              func(ctx context.Context, treatmentID uint64) (*model.Treatment, error)
	createFn                func(ctx context.Context, treatment *model.Treatment) error
	updateFn                func(ctx context.Context, treatmentID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, treatmentID uint64) error
	bulkUpdateSortOrderFn   func(ctx context.Context, updates []repository.TreatmentSortUpdate) error
}

func (m *mockTreatmentRepository) ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.Treatment, error) {
	return m.listByMedicalRecordIDFn(ctx, medicalRecordID)
}

func (m *mockTreatmentRepository) FindByID(ctx context.Context, treatmentID uint64) (*model.Treatment, error) {
	return m.findByIDFn(ctx, treatmentID)
}

func (m *mockTreatmentRepository) Create(ctx context.Context, treatment *model.Treatment) error {
	return m.createFn(ctx, treatment)
}

func (m *mockTreatmentRepository) Update(ctx context.Context, treatmentID uint64, fields map[string]any) error {
	return m.updateFn(ctx, treatmentID, fields)
}

func (m *mockTreatmentRepository) Delete(ctx context.Context, treatmentID uint64) error {
	return m.deleteFn(ctx, treatmentID)
}

func (m *mockTreatmentRepository) BulkUpdateSortOrder(ctx context.Context, updates []repository.TreatmentSortUpdate) error {
	return m.bulkUpdateSortOrderFn(ctx, updates)
}

// ---- Tests ----

func TestTreatmentService_List(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		repoTreatments  []model.Treatment
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:            "returns treatments for medical record",
			medicalRecordID: 1,
			repoTreatments: []model.Treatment{
				{ID: 1, MedicalRecordID: 1, ItemType: model.TreatmentItemTypeConsultation, Status: model.TreatmentStatusPending},
				{ID: 2, MedicalRecordID: 1, ItemType: model.TreatmentItemTypeProcedure, Status: model.TreatmentStatusCompleted},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty list when no treatments exist",
			medicalRecordID: 999,
			repoTreatments:  []model.Treatment{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoTreatments:  nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _ uint64) ([]model.Treatment, error) {
					return tt.repoTreatments, tt.repoErr
				},
			}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment: repo,
				Inventory: &mockInventoryRepository{},
			})

			treatments, err := svc.List(context.Background(), tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, treatments, tt.wantLen)
			}
		})
	}
}

func TestTreatmentService_Create(t *testing.T) {
	procedureID := uint64(1)
	unitPrice := int64(10000)
	discountRate := 0.1
	quantity := 1.0

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *CreateTreatmentInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "creates treatment with valid item type and default status",
			medicalRecordID: 1,
			input: &CreateTreatmentInput{
				ItemType:     model.TreatmentItemTypeProcedure,
				ProcedureID:  &procedureID,
				UnitPrice:    unitPrice,
				Quantity:     quantity,
				Selected:     true,
				DiscountRate: discountRate,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "creates treatment with custom status",
			medicalRecordID: 1,
			input: &CreateTreatmentInput{
				ItemType:  model.TreatmentItemTypeConsultation,
				Status:    string(model.TreatmentStatusCompleted),
				UnitPrice: unitPrice,
				Quantity:  quantity,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "returns error on invalid item type",
			medicalRecordID: 1,
			input: &CreateTreatmentInput{
				ItemType:  model.TreatmentItemType("invalid_type"),
				UnitPrice: unitPrice,
				Quantity:  quantity,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:            "returns error on invalid status",
			medicalRecordID: 1,
			input: &CreateTreatmentInput{
				ItemType:  model.TreatmentItemTypeConsultation,
				Status:    "invalid_status",
				UnitPrice: unitPrice,
				Quantity:  quantity,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &CreateTreatmentInput{
				ItemType:  model.TreatmentItemTypeMedicine,
				UnitPrice: unitPrice,
				Quantity:  quantity,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentRepository{
				createFn: func(_ context.Context, _ *model.Treatment) error {
					return tt.repoErr
				},
			}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment: repo,
				Inventory: &mockInventoryRepository{},
			})

			treatment, err := svc.Create(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, treatment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, treatment)
				assert.Equal(t, tt.medicalRecordID, treatment.MedicalRecordID)
				assert.Equal(t, tt.input.ItemType, treatment.ItemType)
			}
		})
	}
}

func TestTreatmentService_Update(t *testing.T) {
	newUnitPrice := int64(15000)
	newQuantity := 2.0
	newStatus := string(model.TreatmentStatusCompleted)
	newMemo := "Updated treatment"

	tests := []struct {
		name             string
		medicalRecordID  uint64
		treatmentID      uint64
		input            *UpdateTreatmentInput
		repoTreatment    *model.Treatment
		findByIDErr      error
		updateErr        error
		wantErr          bool
		wantErrSubstring string
	}{
		{
			name:            "updates treatment successfully",
			medicalRecordID: 1,
			treatmentID:     1,
			input: &UpdateTreatmentInput{
				UnitPrice: &newUnitPrice,
				Quantity:  &newQuantity,
			},
			repoTreatment: &model.Treatment{
				ID:              1,
				MedicalRecordID: 1,
				ItemType:        model.TreatmentItemTypeConsultation,
				UnitPrice:       newUnitPrice,
				Quantity:        newQuantity,
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     false,
		},
		{
			name:            "returns error when no fields provided",
			medicalRecordID: 1,
			treatmentID:     1,
			input:           &UpdateTreatmentInput{},
			repoTreatment: &model.Treatment{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns not found when treatment does not belong to medical record",
			medicalRecordID: 1,
			treatmentID:     999,
			input: &UpdateTreatmentInput{
				Memo: &newMemo,
			},
			repoTreatment: &model.Treatment{
				ID:              999,
				MedicalRecordID: 2, // Different medical record
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns error when treatment not found",
			medicalRecordID: 1,
			treatmentID:     999,
			input: &UpdateTreatmentInput{
				UnitPrice: &newUnitPrice,
			},
			repoTreatment: nil,
			findByIDErr:   apperrors.WrapNotFound("treatment", "999"),
			updateErr:     nil,
			wantErr:       true,
		},
		{
			name:            "returns error on invalid status",
			medicalRecordID: 1,
			treatmentID:     1,
			input: &UpdateTreatmentInput{
				Status: func(s string) *string { return &s }("invalid_status"),
			},
			repoTreatment: &model.Treatment{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			treatmentID:     1,
			input: &UpdateTreatmentInput{
				UnitPrice: &newUnitPrice,
			},
			repoTreatment: &model.Treatment{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByIDErr: nil,
			updateErr:   errors.New("db error"),
			wantErr:     true,
		},
		{
			name:            "updates only status field",
			medicalRecordID: 1,
			treatmentID:     1,
			input: &UpdateTreatmentInput{
				Status: &newStatus,
			},
			repoTreatment: &model.Treatment{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.TreatmentStatusCompleted,
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Treatment, error) {
					return tt.repoTreatment, tt.findByIDErr
				},
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.updateErr
				},
			}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment: repo,
				Inventory: &mockInventoryRepository{},
			})

			treatment, err := svc.Update(context.Background(), tt.medicalRecordID, tt.treatmentID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, treatment)
			}
		})
	}
}

func TestTreatmentService_Delete(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		treatmentID     uint64
		repoTreatment   *model.Treatment
		findByIDErr     error
		deleteErr       error
		wantErr         bool
	}{
		{
			name:            "deletes treatment successfully",
			medicalRecordID: 1,
			treatmentID:     1,
			repoTreatment: &model.Treatment{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByIDErr: nil,
			deleteErr:   nil,
			wantErr:     false,
		},
		{
			name:            "returns not found when treatment does not belong to medical record",
			medicalRecordID: 1,
			treatmentID:     999,
			repoTreatment: &model.Treatment{
				ID:              999,
				MedicalRecordID: 2, // Different medical record
			},
			findByIDErr: nil,
			deleteErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns error when treatment not found",
			medicalRecordID: 1,
			treatmentID:     999,
			repoTreatment:   nil,
			findByIDErr:     apperrors.WrapNotFound("treatment", "999"),
			deleteErr:       nil,
			wantErr:         true,
		},
		{
			name:            "returns error when delete fails",
			medicalRecordID: 1,
			treatmentID:     1,
			repoTreatment: &model.Treatment{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByIDErr: nil,
			deleteErr:   errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Treatment, error) {
					return tt.repoTreatment, tt.findByIDErr
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment: repo,
				Inventory: &mockInventoryRepository{},
			})

			err := svc.Delete(context.Background(), tt.medicalRecordID, tt.treatmentID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTreatmentService_BulkUpdateSortOrder(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *BulkUpdateTreatmentsInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "bulk updates sort order successfully",
			medicalRecordID: 1,
			input: &BulkUpdateTreatmentsInput{
				Treatments: []BulkTreatmentItem{
					{ID: 1, SortOrder: 3},
					{ID: 2, SortOrder: 1},
					{ID: 3, SortOrder: 2},
				},
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &BulkUpdateTreatmentsInput{
				Treatments: []BulkTreatmentItem{
					{ID: 1, SortOrder: 1},
				},
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:            "handles empty treatments list",
			medicalRecordID: 1,
			input: &BulkUpdateTreatmentsInput{
				Treatments: []BulkTreatmentItem{},
			},
			repoErr: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentRepository{
				bulkUpdateSortOrderFn: func(_ context.Context, _ []repository.TreatmentSortUpdate) error {
					return tt.repoErr
				},
			}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment: repo,
				Inventory: &mockInventoryRepository{},
			})

			err := svc.BulkUpdateSortOrder(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
