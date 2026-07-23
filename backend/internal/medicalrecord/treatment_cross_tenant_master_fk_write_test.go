package medicalrecord

// treatment_cross_tenant_master_fk_write_test.go — BE9-2D sub-batch④b: internal/service
// cross_tenant_master_fk_write_test.go の treatment 節（Create/Update × MasterFK/InventoryFK の
// 4 test）を同名のまま縦移動。mock/builder は treatment_mocks_test.go の narrow-view 版を使う。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ── treatment (CRITICAL): medicine/procedure/consultation ──

func TestTreatmentService_Create_RejectsCrossClinicMasterFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedProcedureID = uint64(10)
	const foreignProcedureID = uint64(999)
	unitPrice := int64(1000)

	newSvc := func(created *bool) TreatmentService {
		treatRepo := &mockTreatmentRepository{createFn: func(_ context.Context, _ *model.Treatment) error {
			*created = true
			return nil
		}}
		return NewTreatmentServiceWithAudit(
			treatRepo, draftMedicalRecordRepo(), okMedicineRepo(), rejectProcedureRepo(ownedProcedureID),
			okConsultationRepo(), &mockInventoryRepository{}, benignVitalRepo(),
			&mockMedicineDoseParamRepository{}, &mockTransactor{}, nil)
	}

	t.Run("rejects cross-clinic procedure_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignProcedureID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateTreatmentInput{
			ItemType:    model.TreatmentItemTypeProcedure,
			ProcedureID: &foreign,
			UnitPrice:   unitPrice,
			Quantity:    1,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "treatment row must NOT be persisted when referencing another clinic's procedure")
	})

	t.Run("accepts same-clinic procedure_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedProcedureID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateTreatmentInput{
			ItemType:    model.TreatmentItemTypeProcedure,
			ProcedureID: &owned,
			UnitPrice:   unitPrice,
			Quantity:    1,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestTreatmentService_Update_RejectsCrossClinicMasterFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedProcedureID = uint64(10)
	const foreignProcedureID = uint64(999)

	newSvc := func(updated *bool) TreatmentService {
		treatRepo := &mockTreatmentRepository{
			findByIDFn: func(_ context.Context, _, treatmentID uint64) (*model.Treatment, error) {
				return &model.Treatment{ID: treatmentID, MedicalRecordID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				*updated = true
				return nil
			},
		}
		return NewTreatmentServiceWithAudit(
			treatRepo, draftMedicalRecordRepo(), okMedicineRepo(), rejectProcedureRepo(ownedProcedureID),
			okConsultationRepo(), &mockInventoryRepository{}, benignVitalRepo(),
			&mockMedicineDoseParamRepository{}, &mockTransactor{}, nil)
	}

	t.Run("rejects cross-clinic procedure_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignProcedureID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateTreatmentInput{ProcedureID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "treatment must NOT be updated to reference another clinic's procedure")
	})
}

// ── treatment InventoryID (X-14a / INV-SEC P1): the pre-persist ownership guard rejects
// cross-clinic treatment links; DecreaseStock independently scopes the stock update by clinicID. ──

func TestTreatmentService_Create_RejectsCrossClinicInventoryFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedInventoryID = uint64(20)
	const foreignInventoryID = uint64(888)
	unitPrice := int64(1000)

	newSvc := func(created *bool) TreatmentService {
		treatRepo := &mockTreatmentRepository{createFn: func(_ context.Context, _ *model.Treatment) error {
			*created = true
			return nil
		}}
		return NewTreatmentServiceWithAudit(
			treatRepo, draftMedicalRecordRepo(), okMedicineRepo(), okProcedureRepo(),
			okConsultationRepo(), rejectInventoryRepo(ownedInventoryID), benignVitalRepo(),
			&mockMedicineDoseParamRepository{}, &mockTransactor{}, nil)
	}

	t.Run("rejects cross-clinic inventory_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignInventoryID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateTreatmentInput{
			ItemType:    model.TreatmentItemTypeProcedure,
			InventoryID: &foreign,
			UnitPrice:   unitPrice,
			Quantity:    1,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "treatment row must NOT be persisted when referencing another clinic's inventory item")
	})

	t.Run("accepts same-clinic inventory_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedInventoryID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateTreatmentInput{
			ItemType:    model.TreatmentItemTypeProcedure,
			InventoryID: &owned,
			UnitPrice:   unitPrice,
			Quantity:    1,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestTreatmentService_Update_RejectsCrossClinicInventoryFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedInventoryID = uint64(20)
	const foreignInventoryID = uint64(888)

	newSvc := func(updated *bool) TreatmentService {
		treatRepo := &mockTreatmentRepository{
			findByIDFn: func(_ context.Context, _, treatmentID uint64) (*model.Treatment, error) {
				return &model.Treatment{ID: treatmentID, MedicalRecordID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				*updated = true
				return nil
			},
		}
		return NewTreatmentServiceWithAudit(
			treatRepo, draftMedicalRecordRepo(), okMedicineRepo(), okProcedureRepo(),
			okConsultationRepo(), rejectInventoryRepo(ownedInventoryID), benignVitalRepo(),
			&mockMedicineDoseParamRepository{}, &mockTransactor{}, nil)
	}

	t.Run("rejects cross-clinic inventory_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignInventoryID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateTreatmentInput{InventoryID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "treatment must NOT be updated to reference another clinic's inventory item")
	})

	t.Run("accepts same-clinic inventory_id on update (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedInventoryID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateTreatmentInput{InventoryID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}
