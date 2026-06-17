package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- MedicalRecord スタブ（BulkUpdateSortOrder テナント検証用）----

type mockMedicalRecordRepoForTreatment struct{}

func (m *mockMedicalRecordRepoForTreatment) FindAll(_ context.Context, _ []uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
	return nil, 0, nil
}
func (m *mockMedicalRecordRepoForTreatment) FindByID(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
}
func (m *mockMedicalRecordRepoForTreatment) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *mockMedicalRecordRepoForTreatment) Create(_ context.Context, _ *model.MedicalRecord) error {
	return nil
}
func (m *mockMedicalRecordRepoForTreatment) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
	return &model.MedicalRecord{}, nil
}
func (m *mockMedicalRecordRepoForTreatment) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockMedicalRecordRepoForTreatment) CountByPetID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockMedicalRecordRepoForTreatment) CountEstimatesByMedicalRecordID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockMedicalRecordRepoForTreatment) FindOwnerVisitSummary(_ context.Context, _, _ uint64) (*repository.OwnerVisitSummary, error) {
	return &repository.OwnerVisitSummary{}, nil
}

func (m *mockMedicalRecordRepoForTreatment) FindLatestByOwner(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepoForTreatment) FindDormantOwnerEntries(_ context.Context, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepoForTreatment) FindOwnersByFirstVisitDate(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepoForTreatment) FindOwnersByLastVisitDays(_ context.Context, _ uint64, _ int, _ time.Time) ([]uint64, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepoForTreatment) FindOwnersByNextVisitRecommended(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepoForTreatment) CountByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockMedicalRecordRepoForTreatment) DeleteDraftByAppointmentID(_ context.Context, _, _ uint64) error {
	return nil
}

// ---- Treatment モック ----

type mockTreatmentRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error)
	findByIDFn              func(ctx context.Context, clinicID, treatmentID uint64) (*model.Treatment, error)
	createFn                func(ctx context.Context, treatment *model.Treatment) error
	updateFn                func(ctx context.Context, clinicID, treatmentID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, treatmentID uint64) error
	bulkUpdateSortOrderFn   func(ctx context.Context, updates []repository.TreatmentSortUpdate) error
	findHistoryByPetIDFn    func(ctx context.Context, clinicID, petID uint64, itemType *model.TreatmentItemType, page, limit int) ([]model.Treatment, int64, error)
}

func (m *mockTreatmentRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockTreatmentRepository) FindByID(ctx context.Context, clinicID, treatmentID uint64) (*model.Treatment, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, treatmentID)
	}
	return nil, nil
}

func (m *mockTreatmentRepository) Create(ctx context.Context, treatment *model.Treatment) error {
	return m.createFn(ctx, treatment)
}

func (m *mockTreatmentRepository) Update(ctx context.Context, clinicID, treatmentID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, treatmentID, fields)
}

func (m *mockTreatmentRepository) Delete(ctx context.Context, clinicID, treatmentID uint64) error {
	return m.deleteFn(ctx, clinicID, treatmentID)
}

func (m *mockTreatmentRepository) BulkUpdateSortOrder(ctx context.Context, updates []repository.TreatmentSortUpdate) error {
	return m.bulkUpdateSortOrderFn(ctx, updates)
}

func (m *mockTreatmentRepository) FindUnbilledByPetID(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
	return nil, nil
}

func (m *mockTreatmentRepository) FindHistoryByPetID(ctx context.Context, clinicID, petID uint64, itemType *model.TreatmentItemType, page, limit int) ([]model.Treatment, int64, error) {
	if m.findHistoryByPetIDFn != nil {
		return m.findHistoryByPetIDFn(ctx, clinicID, petID, itemType, page, limit)
	}
	return nil, 0, nil
}

func (m *mockTreatmentRepository) CountFinalizedUnconfirmedByPetAndDate(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}

// ---- Tests ----

func TestTreatmentService_List(t *testing.T) {
	const clinicID = uint64(1)

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
				listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
					return tt.repoTreatments, tt.repoErr
				},
			}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment:     repo,
				MedicalRecord: &mockMedicalRecordRepoForTreatment{},
				Inventory:     &mockInventoryRepository{},
			})

			treatments, err := svc.List(context.Background(), clinicID, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, treatments, tt.wantLen)
			}
		})
	}
}

func TestTreatmentService_ListPetHistory(t *testing.T) {
	const (
		clinicID = uint64(1)
		petID    = uint64(7)
	)
	medicine := model.TreatmentItemTypeMedicine
	invalid := model.TreatmentItemType("bogus")

	t.Run("returns treatments and total on success", func(t *testing.T) {
		var gotItemType *model.TreatmentItemType
		repo := &mockTreatmentRepository{
			findHistoryByPetIDFn: func(_ context.Context, gotClinic, gotPet uint64, it *model.TreatmentItemType, _, _ int) ([]model.Treatment, int64, error) {
				assert.Equal(t, clinicID, gotClinic)
				assert.Equal(t, petID, gotPet)
				gotItemType = it
				return []model.Treatment{{ID: 1}, {ID: 2}}, 2, nil
			},
		}
		svc := NewTreatmentService(&repository.Repositories{Treatment: repo})

		treatments, total, err := svc.ListPetHistory(context.Background(), clinicID, petID, &medicine, 1, 100)

		assert.NoError(t, err)
		assert.Len(t, treatments, 2)
		assert.Equal(t, int64(2), total)
		if assert.NotNil(t, gotItemType) {
			assert.Equal(t, medicine, *gotItemType)
		}
	})

	t.Run("rejects invalid item_type before hitting repository", func(t *testing.T) {
		called := false
		repo := &mockTreatmentRepository{
			findHistoryByPetIDFn: func(_ context.Context, _, _ uint64, _ *model.TreatmentItemType, _, _ int) ([]model.Treatment, int64, error) {
				called = true
				return nil, 0, nil
			},
		}
		svc := NewTreatmentService(&repository.Repositories{Treatment: repo})

		_, _, err := svc.ListPetHistory(context.Background(), clinicID, petID, &invalid, 1, 100)

		assert.Error(t, err)
		assert.False(t, called, "repository should not be called for invalid item_type")
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockTreatmentRepository{
			findHistoryByPetIDFn: func(_ context.Context, _, _ uint64, _ *model.TreatmentItemType, _, _ int) ([]model.Treatment, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}
		svc := NewTreatmentService(&repository.Repositories{Treatment: repo})

		_, _, err := svc.ListPetHistory(context.Background(), clinicID, petID, nil, 1, 100)

		assert.Error(t, err)
	})
}

func TestTreatmentService_Create(t *testing.T) {
	const clinicID = uint64(1)
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
				IsSelected:   true,
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
			invRepo := &mockInventoryRepository{}
			// TransactionFn: DB 不要でトランザクションをインライン実行
			repos := &repository.Repositories{
				Treatment:     repo,
				MedicalRecord: &mockMedicalRecordRepoForTreatment{},
				Inventory:     invRepo,
			}
			repos.TransactionFn = func(ctx context.Context, fn func(*repository.Repositories) error) error {
				return fn(repos)
			}
			svc := NewTreatmentService(repos)

			treatment, err := svc.Create(context.Background(), clinicID, tt.medicalRecordID, tt.input)

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
	const clinicID = uint64(1)
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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Treatment, error) {
					return tt.repoTreatment, tt.findByIDErr
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.updateErr
				},
			}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment:     repo,
				MedicalRecord: &mockMedicalRecordRepoForTreatment{},
				Inventory:     &mockInventoryRepository{},
			})

			treatment, err := svc.Update(context.Background(), clinicID, tt.medicalRecordID, tt.treatmentID, tt.input)

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
	const clinicID = uint64(1)

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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Treatment, error) {
					return tt.repoTreatment, tt.findByIDErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment:     repo,
				MedicalRecord: &mockMedicalRecordRepoForTreatment{},
				Inventory:     &mockInventoryRepository{},
			})

			err := svc.Delete(context.Background(), clinicID, tt.medicalRecordID, tt.treatmentID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTreatmentService_BulkUpdateSortOrder(t *testing.T) {
	const clinicID = uint64(1)

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
			mrRepo := &mockMedicalRecordRepoForTreatment{}
			svc := NewTreatmentService(&repository.Repositories{
				Treatment:     repo,
				Inventory:     &mockInventoryRepository{},
				MedicalRecord: mrRepo,
			})

			err := svc.BulkUpdateSortOrder(context.Background(), clinicID, tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
