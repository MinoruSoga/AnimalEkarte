package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- BE9-2D ④b test harness ----
// （mock/builder 群は treatment_mocks_test.go）

// newTreatmentSvc は個別依存注入コンストラクタへのテスト共通配線。
// 旧 harness の repos.TransactionFn インライン実行は mockTransactor の WithTx 素通し（fn(ctx)）が等価。
// dose 再検証パスに入らないテスト用に medicine/procedure/consultation は ok*Repo、vital/doseParam は
// 空応答 mock を配線する。
func newTreatmentSvc(repo TreatmentRepository, mrRepo treatmentMedicalRecordRepo, invRepo treatmentInventoryRepo, auditTx AuditTxLogger) TreatmentService {
	return NewTreatmentServiceWithAudit(
		repo, mrRepo, okMedicineRepo(), okProcedureRepo(), okConsultationRepo(), invRepo,
		benignVitalRepo(), &mockMedicineDoseParamRepository{}, &mockTransactor{}, auditTx)
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
			svc := newTreatmentSvc(repo, &mockMedicalRecordRepository{}, &mockInventoryRepository{}, nil)

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
		var gotFilter model.PetTreatmentHistoryFilter
		repo := &mockTreatmentRepository{
			findHistoryByPetIDFn: func(_ context.Context, gotClinic, gotPet uint64, f model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
				assert.Equal(t, clinicID, gotClinic)
				assert.Equal(t, petID, gotPet)
				gotFilter = f
				return []model.Treatment{{ID: 1}, {ID: 2}}, 2, nil
			},
		}
		svc := newTreatmentSvc(repo, &mockMedicalRecordRepository{}, &mockInventoryRepository{}, nil)

		treatments, total, err := svc.ListPetHistory(context.Background(), clinicID, petID, model.PetTreatmentHistoryFilter{ItemType: &medicine}, 1, 100)

		assert.NoError(t, err)
		assert.Len(t, treatments, 2)
		assert.Equal(t, int64(2), total)
		if assert.NotNil(t, gotFilter.ItemType) {
			assert.Equal(t, medicine, *gotFilter.ItemType)
		}
	})

	t.Run("rejects invalid item_type before hitting repository", func(t *testing.T) {
		called := false
		repo := &mockTreatmentRepository{
			findHistoryByPetIDFn: func(_ context.Context, _, _ uint64, _ model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
				called = true
				return nil, 0, nil
			},
		}
		svc := newTreatmentSvc(repo, &mockMedicalRecordRepository{}, &mockInventoryRepository{}, nil)

		_, _, err := svc.ListPetHistory(context.Background(), clinicID, petID, model.PetTreatmentHistoryFilter{ItemType: &invalid}, 1, 100)

		assert.Error(t, err)
		assert.False(t, called, "repository should not be called for invalid item_type")
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockTreatmentRepository{
			findHistoryByPetIDFn: func(_ context.Context, _, _ uint64, _ model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}
		svc := newTreatmentSvc(repo, &mockMedicalRecordRepository{}, &mockInventoryRepository{}, nil)

		_, _, err := svc.ListPetHistory(context.Background(), clinicID, petID, model.PetTreatmentHistoryFilter{}, 1, 100)

		assert.Error(t, err)
	})

	t.Run("passes AnesthesiaOnly=true to repository", func(t *testing.T) {
		var gotFilter model.PetTreatmentHistoryFilter
		repo := &mockTreatmentRepository{
			findHistoryByPetIDFn: func(_ context.Context, _, _ uint64, f model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
				gotFilter = f
				return []model.Treatment{{ID: 10}}, 1, nil
			},
		}
		svc := newTreatmentSvc(repo, &mockMedicalRecordRepository{}, &mockInventoryRepository{}, nil)

		_, _, err := svc.ListPetHistory(context.Background(), clinicID, petID, model.PetTreatmentHistoryFilter{AnesthesiaOnly: true}, 1, 100)

		assert.NoError(t, err)
		assert.True(t, gotFilter.AnesthesiaOnly)
		assert.False(t, gotFilter.IsSurgery)
		assert.Nil(t, gotFilter.ItemType)
	})

	t.Run("passes IsSurgery=true to repository", func(t *testing.T) {
		var gotFilter model.PetTreatmentHistoryFilter
		repo := &mockTreatmentRepository{
			findHistoryByPetIDFn: func(_ context.Context, _, _ uint64, f model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
				gotFilter = f
				return []model.Treatment{{ID: 20}}, 1, nil
			},
		}
		svc := newTreatmentSvc(repo, &mockMedicalRecordRepository{}, &mockInventoryRepository{}, nil)

		_, _, err := svc.ListPetHistory(context.Background(), clinicID, petID, model.PetTreatmentHistoryFilter{IsSurgery: true}, 1, 100)

		assert.NoError(t, err)
		assert.False(t, gotFilter.AnesthesiaOnly)
		assert.True(t, gotFilter.IsSurgery)
		assert.Nil(t, gotFilter.ItemType)
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
			// mockTransactor: DB 不要でトランザクションをインライン実行
			svc := newTreatmentSvc(repo, draftMedicalRecordRepo(), &mockInventoryRepository{}, nil)

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

// TestTreatmentService_Create_DecreaseStock は FOLLOWUP-X14A / INV-SEC P1 の回帰テスト。
// medicines.id と inventory_items.id は独立した採番空間の別テーブルであり、
// MedicineID を在庫 ID として代用せず、明示された InventoryID と認証済み clinicID のみを
// DecreaseStock へ渡す。
func TestTreatmentService_Create_DecreaseStock(t *testing.T) {
	const (
		clinicID        = uint64(1)
		medicalRecordID = uint64(1)
	)
	unitPrice := int64(1000)
	quantity := 2.0

	t.Run("MedicineID only (InventoryID nil): stock is NOT decreased", func(t *testing.T) {
		// Arrange
		medicineID := uint64(999) // inventory_items.id としては無関係の値（採番衝突を模擬）
		called := false
		treatmentRepo := &mockTreatmentRepository{
			createFn: func(_ context.Context, _ *model.Treatment) error { return nil },
		}
		invRepo := &mockInventoryRepository{
			decreaseStockFn: func(_ context.Context, _, _ uint64, _ float64) error {
				called = true
				return nil
			},
		}
		svc := newTreatmentSvc(treatmentRepo, draftMedicalRecordRepo(), invRepo, nil)
		input := &CreateTreatmentInput{
			ItemType:   model.TreatmentItemTypeMedicine,
			MedicineID: &medicineID,
			UnitPrice:  unitPrice,
			Quantity:   quantity,
		}

		// Act
		treatment, err := svc.Create(context.Background(), clinicID, medicalRecordID, input)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, treatment)
		assert.False(t, called, "DecreaseStock must not be called when only MedicineID is set (medicines.id is not an inventory_items.id)")
	})

	t.Run("InventoryID explicitly set: stock IS decreased for that inventory id", func(t *testing.T) {
		// Arrange
		medicineID := uint64(999) // 同時に指定されても DecreaseStock の対象に影響しないことを確認する
		inventoryID := uint64(42)
		callCount := 0
		var gotClinicID uint64
		var gotID uint64
		var gotQty float64
		treatmentRepo := &mockTreatmentRepository{
			createFn: func(_ context.Context, _ *model.Treatment) error { return nil },
		}
		invRepo := &mockInventoryRepository{
			decreaseStockFn: func(_ context.Context, passedClinicID, id uint64, qty float64) error {
				callCount++
				gotClinicID = passedClinicID
				gotID = id
				gotQty = qty
				return nil
			},
		}
		svc := newTreatmentSvc(treatmentRepo, draftMedicalRecordRepo(), invRepo, nil)
		input := &CreateTreatmentInput{
			ItemType:    model.TreatmentItemTypeMedicine,
			MedicineID:  &medicineID,
			InventoryID: &inventoryID,
			UnitPrice:   unitPrice,
			Quantity:    quantity,
		}

		// Act
		treatment, err := svc.Create(context.Background(), clinicID, medicalRecordID, input)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, treatment)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, clinicID, gotClinicID)
		assert.Equal(t, inventoryID, gotID)
		assert.Equal(t, quantity, gotQty)
	})
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
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateTreatmentInput) error {
					return tt.updateErr
				},
			}
			svc := newTreatmentSvc(repo, draftMedicalRecordRepo(), &mockInventoryRepository{}, nil)

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
			// mockTransactor: DB 不要でトランザクションをインライン実行（BE-refactor.md H-8b）
			svc := newTreatmentSvc(repo, draftMedicalRecordRepo(), &mockInventoryRepository{}, nil)

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
		parentStatus    model.MedicalRecordStatus
		repoErr         error
		wantErr         bool
		wantConflict    bool
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
			parentStatus: model.MedicalRecordStatusDraft,
			repoErr:      nil,
			wantErr:      false,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &BulkUpdateTreatmentsInput{
				Treatments: []BulkTreatmentItem{
					{ID: 1, SortOrder: 1},
				},
			},
			parentStatus: model.MedicalRecordStatusDraft,
			repoErr:      errors.New("db error"),
			wantErr:      true,
		},
		{
			name:            "handles empty treatments list",
			medicalRecordID: 1,
			input: &BulkUpdateTreatmentsInput{
				Treatments: []BulkTreatmentItem{},
			},
			parentStatus: model.MedicalRecordStatusDraft,
			repoErr:      nil,
			wantErr:      false,
		},
		{
			name:            "returns conflict when parent medical record is finalized",
			medicalRecordID: 1,
			input: &BulkUpdateTreatmentsInput{
				Treatments: []BulkTreatmentItem{
					{ID: 1, SortOrder: 1},
				},
			},
			parentStatus: model.MedicalRecordStatusFinalized,
			repoErr:      nil,
			wantErr:      true,
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentRepository{
				bulkUpdateSortOrderFn: func(_ context.Context, _ []TreatmentSortUpdate) error {
					return tt.repoErr
				},
			}
			mrRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{Status: tt.parentStatus}, nil
				},
			}
			// mockTransactor: DB 不要でトランザクションをインライン実行（BE-refactor.md H-8c）
			svc := newTreatmentSvc(repo, mrRepo, &mockInventoryRepository{}, nil)

			err := svc.BulkUpdateSortOrder(context.Background(), clinicID, tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err), "expected conflict but got: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
