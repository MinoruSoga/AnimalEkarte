package medicalrecord

// masters_cross_tenant_master_fk_write_test.go — BE9-2D ⑥: internal/service
// cross_tenant_master_fk_write_test.go の consultation/procedure/medicine 節
// （ParentFK/InventoryFK guard 計8テスト）を同名のまま縦移動。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestConsultationService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) ConsultationService {
		repo := &mockConsultationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Consultation, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("consultation", "foreign")
				}
				return &model.Consultation{ID: id}, nil
			},
			createFn: func(_ context.Context, _ *model.Consultation) error { *created = true; return nil },
		}
		return NewConsultationService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateConsultationInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "consultation must NOT be persisted referencing another clinic's parent consultation")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateConsultationInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestConsultationService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) ConsultationService {
		repo := &mockConsultationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Consultation, error) {
				if id == entityID || id == ownedParentID {
					return &model.Consultation{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("consultation", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateConsultationInput) (*model.Consultation, error) {
				*updated = true
				return &model.Consultation{ID: id}, nil
			},
		}
		return NewConsultationService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateConsultationInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "consultation must NOT be updated to reference another clinic's parent consultation")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateConsultationInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// TestExamTypeService_Create/Update_RejectsCrossClinicParentFK moved to
// internal/medicalrecord/exam_type_cross_tenant_test.go (BE9-2C): ExaminationTypeService /
// NewExamTypeService / CreateExamTypeInput / UpdateExamTypeInput no longer exist in this
// package once internal/service/exam_type_service.go is deleted by that batch (zero
// remaining fan-in — see internal/repository/exam_type_repository.go for the still-live
// repository-level facade other not-yet-migrated services depend on).

func TestProcedureService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) ProcedureService {
		repo := &mockProcedureRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("procedure", "foreign")
				}
				return &model.Procedure{ID: id}, nil
			},
			createFn: func(_ context.Context, _ *model.Procedure) error { *created = true; return nil },
		}
		return NewProcedureService(repo, &mockTransactor{})
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateProcedureInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "procedure must NOT be persisted referencing another clinic's parent procedure")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateProcedureInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestProcedureService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) ProcedureService {
		repo := &mockProcedureRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
				if id == entityID || id == ownedParentID {
					return &model.Procedure{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("procedure", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateProcedureInput) (*model.Procedure, error) {
				*updated = true
				return &model.Procedure{ID: id}, nil
			},
		}
		return NewProcedureService(repo, &mockTransactor{})
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateProcedureInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "procedure must NOT be updated to reference another clinic's parent procedure")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateProcedureInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── X-14 batch U2: medicineService self-ref ParentID + InventoryID (InventoryItem) ──
//
// Unlike the five self-ref-only services above, medicineService.Create unconditionally
// creates a linked InventoryItem in the same tx (BUG-320), so its own repo/inventoryRepo
// mocks need createFn wired even in the "rejects" tests' shared okInventoryRepo/
// okMedicineRepo-shaped setup — see okInventoryRepo/rejectInventoryRepo above.

func TestMedicineService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) MedicineService {
		repo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("medicine", "foreign")
				}
				return &model.Medicine{ID: id}, nil
			},
			createFn: func(_ context.Context, medicine *model.Medicine) error {
				*created = true
				medicine.ID = 1
				return nil
			},
		}
		return NewMedicineServiceWithAudit(repo, okInventoryRepo(), &mockTransactor{}, nil)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "medicine must NOT be persisted referencing another clinic's parent medicine")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestMedicineService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) MedicineService {
		repo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
				if id == entityID || id == ownedParentID {
					return &model.Medicine{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("medicine", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
				*updated = true
				return &model.Medicine{ID: id}, nil
			},
		}
		return NewMedicineServiceWithAudit(repo, &mockInventoryRepository{}, &mockTransactor{}, nil)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateMedicineInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "medicine must NOT be updated to reference another clinic's parent medicine")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateMedicineInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

func TestMedicineService_Create_RejectsCrossClinicInventoryFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedInventoryID = uint64(20)
	const foreignInventoryID = uint64(888)

	newSvc := func(created *bool, inventoryRepo medicineInventoryRepo) MedicineService {
		repo := &mockMedicineRepository{
			createFn: func(_ context.Context, medicine *model.Medicine) error {
				*created = true
				medicine.ID = 1
				return nil
			},
		}
		return NewMedicineServiceWithAudit(repo, inventoryRepo, &mockTransactor{}, nil)
	}

	t.Run("rejects cross-clinic inventory_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectInventoryRepo(ownedInventoryID))
		foreign := foreignInventoryID
		out, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{Name: "x", InventoryID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "medicine must NOT be persisted referencing another clinic's inventory item")
	})

	t.Run("accepts same-clinic inventory_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectInventoryRepo(ownedInventoryID))
		owned := ownedInventoryID
		out, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{Name: "x", InventoryID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestMedicineService_Update_RejectsCrossClinicInventoryFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedInventoryID = uint64(20)
	const foreignInventoryID = uint64(888)

	newSvc := func(updated *bool, inventoryRepo medicineInventoryRepo) MedicineService {
		repo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
				return &model.Medicine{ID: id}, nil
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
				*updated = true
				return &model.Medicine{ID: id}, nil
			},
		}
		return NewMedicineServiceWithAudit(repo, inventoryRepo, &mockTransactor{}, nil)
	}

	t.Run("rejects cross-clinic inventory_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated, rejectInventoryRepo(ownedInventoryID))
		foreign := foreignInventoryID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateMedicineInput{InventoryID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "medicine must NOT be updated to reference another clinic's inventory item")
	})

	t.Run("accepts same-clinic inventory_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated, rejectInventoryRepo(ownedInventoryID))
		owned := ownedInventoryID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateMedicineInput{InventoryID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── X-14 batch U4: inquiryService.Save / medicalRecordService.CreateSubRecords ──
//
// inquiryService.Save persisted a request-derived ChiefComplaintTypeID without verifying
// clinic ownership (inquiry_service.go). medicalRecordService.CreateSubRecords carries the
// same hole for ChiefComplaintTypeID plus four diagnosis FKs, bypassing
// clinicalPlanService's validateDiagnosisFKs entirely (medical_record_subrecords.go).
