package service

// Cross-tenant master FK write isolation tests (#201 follow-up).
//
// These tests close the "clinic-scoped master FK write hole" class: a service
// Create/Update persisting a request-derived master FK (medicine/vaccine/
// exam_type/procedure/consultation/checkup_type/diagnosis) WITHOUT verifying the
// master belongs to the caller's clinic — the source of #124/#125-type mislinks.
//
// Each fix gets a dual regression test:
//   (i)  clinic A referencing clinic B's master  -> rejected, row NOT persisted
//   (ii) same-clinic master                      -> still succeeds (no false-reject)
//
// Shared permissive ("ok*") and rejecting ("reject*") master-repo mock builders
// live here so every service test reuses one tenancy model. The reject builders
// return NotFound for IDs not owned by the caller clinic, exactly mirroring the
// real *_repository.go FindByID(ctx, clinicID, id) clinic-scoped behavior.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ── permissive builders: any FindByID succeeds (used to keep existing tests green) ──

func okMedicineRepo() repository.MedicineRepository {
	return &mockMedicineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
		return &model.Medicine{ID: id}, nil
	}}
}

func okProcedureRepo() repository.ProcedureRepository {
	return &mockProcedureRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
		return &model.Procedure{ID: id}, nil
	}}
}

func okConsultationRepo() repository.ConsultationRepository {
	return &mockConsultationRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Consultation, error) {
		return &model.Consultation{ID: id}, nil
	}}
}

func okVaccineRepo() repository.VaccineRepository {
	return &mockVaccineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
		return &model.Vaccine{ID: id}, nil
	}}
}

func okExamTypeRepo() repository.ExamTypeRepository {
	return &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
		return &model.ExaminationType{ID: id}, nil
	}}
}

func okCheckupTypeRepo() repository.CheckupTypeRepository {
	return &mockCheckupTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
		return &model.CheckupType{ID: id}, nil
	}}
}

func okDiagnosisTypeRepo() repository.DiagnosisTypeRepository {
	return &mockDiagnosisTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
		return &model.DiagnosisType{ID: id}, nil
	}}
}

func okDiagnosisNameRepo() repository.DiagnosisNameRepository {
	return &mockDiagnosisNameRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
		return &model.DiagnosisName{ID: id}, nil
	}}
}

// rejectByForeignID returns a clinic-scoped FindByID stub: it returns the row for
// ownedID and NotFound for anything else, mimicking real FindByID(ctx, clinicID, id).
func rejectMedicineRepo(ownedID uint64) repository.MedicineRepository {
	return &mockMedicineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("medicine", "foreign")
		}
		return &model.Medicine{ID: id}, nil
	}}
}

func rejectProcedureRepo(ownedID uint64) repository.ProcedureRepository {
	return &mockProcedureRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("procedure", "foreign")
		}
		return &model.Procedure{ID: id}, nil
	}}
}

// ── treatment (CRITICAL): medicine/procedure/consultation ──

func TestTreatmentService_Create_RejectsCrossClinicMasterFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedProcedureID = uint64(10)
	const foreignProcedureID = uint64(999)
	unitPrice := int64(1000)

	newSvc := func(created *bool) TreatmentService {
		repos := &repository.Repositories{
			Treatment: &mockTreatmentRepository{createFn: func(_ context.Context, _ *model.Treatment) error {
				*created = true
				return nil
			}},
			MedicalRecord: &mockMedicalRecordRepoForTreatment{},
			Inventory:     &mockInventoryRepository{},
			Medicine:      okMedicineRepo(),
			Consultation:  okConsultationRepo(),
			Procedure:     rejectProcedureRepo(ownedProcedureID),
		}
		repos.TransactionFn = func(_ context.Context, fn func(*repository.Repositories) error) error { return fn(repos) }
		return NewTreatmentService(repos)
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
		repos := &repository.Repositories{
			Treatment: &mockTreatmentRepository{
				findByIDFn: func(_ context.Context, _, treatmentID uint64) (*model.Treatment, error) {
					return &model.Treatment{ID: treatmentID, MedicalRecordID: 1}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					*updated = true
					return nil
				},
			},
			MedicalRecord: &mockMedicalRecordRepoForTreatment{},
			Inventory:     &mockInventoryRepository{},
			Medicine:      okMedicineRepo(),
			Consultation:  okConsultationRepo(),
			Procedure:     rejectProcedureRepo(ownedProcedureID),
		}
		repos.TransactionFn = func(_ context.Context, fn func(*repository.Repositories) error) error { return fn(repos) }
		return NewTreatmentService(repos)
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

func rejectVaccineRepo(ownedID uint64) repository.VaccineRepository {
	return &mockVaccineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("vaccine", "foreign")
		}
		return &model.Vaccine{ID: id}, nil
	}}
}

// ── vaccination (CRITICAL #125): vaccine_id ──

func TestVaccinationService_Create_RejectsCrossClinicVaccine(t *testing.T) {
	const clinicID = uint64(1)
	const ownedVaccineID = uint64(10)
	const foreignVaccineID = uint64(999)

	newSvc := func(created *bool) VaccinationService {
		repo := &mockVaccinationRepository{
			createFn: func(_ context.Context, _ *model.Vaccination) error { *created = true; return nil },
		}
		return NewVaccinationService(repo, rejectVaccineRepo(ownedVaccineID), nil)
	}

	t.Run("rejects cross-clinic vaccine_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateVaccinationInput{VaccineID: foreignVaccineID})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "vaccination must NOT be persisted referencing another clinic's vaccine")
	})

	t.Run("accepts same-clinic vaccine_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateVaccinationInput{VaccineID: ownedVaccineID})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestVaccinationService_Update_RejectsCrossClinicVaccine(t *testing.T) {
	const clinicID = uint64(1)
	const ownedVaccineID = uint64(10)
	const foreignVaccineID = uint64(999)

	newSvc := func(updated *bool) VaccinationService {
		repo := &mockVaccinationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
				return &model.Vaccination{ID: id, ClinicID: clinicID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Vaccination, error) {
				*updated = true
				return &model.Vaccination{ID: 1}, nil
			},
		}
		return NewVaccinationService(repo, rejectVaccineRepo(ownedVaccineID), nil)
	}

	t.Run("rejects cross-clinic vaccine_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignVaccineID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateVaccinationInput{VaccineID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "vaccination must NOT be updated to reference another clinic's vaccine")
	})
}

func rejectExamTypeRepo(ownedID uint64) repository.ExamTypeRepository {
	return &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("exam_type", "foreign")
		}
		return &model.ExaminationType{ID: id}, nil
	}}
}

// ── examination (CRITICAL #124): exam_type_id ──

func TestExaminationService_Create_RejectsCrossClinicExamType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const foreignExamTypeID = uint64(999)

	newSvc := func(created *bool) ExaminationService {
		repo := &mockExaminationRepository{
			createFn: func(_ context.Context, _ *model.Examination) error { *created = true; return nil },
		}
		return NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, rejectExamTypeRepo(ownedExamTypeID), nil, nil)
	}

	t.Run("rejects cross-clinic exam_type_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{ExamTypeID: foreignExamTypeID})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "examination must NOT be persisted referencing another clinic's exam_type")
	})

	t.Run("accepts same-clinic exam_type_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{ExamTypeID: ownedExamTypeID})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestExaminationService_Update_RejectsCrossClinicExamType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const foreignExamTypeID = uint64(999)

	newSvc := func(updated *bool) ExaminationService {
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
				return &model.Examination{ID: id, ClinicID: clinicID, Status: model.ExaminationStatusPending}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
				*updated = true
				return &model.Examination{ID: 1}, nil
			},
		}
		return NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, rejectExamTypeRepo(ownedExamTypeID), nil, nil)
	}

	t.Run("rejects cross-clinic exam_type_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignExamTypeID
		out, err := svc.Update(context.Background(), clinicID, 1, UpdateExaminationInput{ExamTypeID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "examination must NOT be updated to reference another clinic's exam_type")
	})
}

// ── care_plan_item (HIGH): medicine_id / procedure_id ──

func TestCarePlanItemService_Create_RejectsCrossClinicMasterFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedMedicineID = uint64(10)
	const foreignMedicineID = uint64(999)

	newSvc := func(created *bool) CarePlanItemService {
		repo := &mockCarePlanItemRepository{
			createFn: func(_ context.Context, _ *model.CarePlanItem) error { *created = true; return nil },
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID}, nil
			},
		}
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), rejectMedicineRepo(ownedMedicineID), okProcedureRepo())
	}

	t.Run("rejects cross-clinic medicine_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignMedicineID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeMedicine), Name: "x", MedicineID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "care plan item must NOT be persisted referencing another clinic's medicine")
	})

	t.Run("accepts same-clinic medicine_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedMedicineID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeMedicine), Name: "x", MedicineID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCarePlanItemService_Update_RejectsCrossClinicMasterFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedMedicineID = uint64(10)
	const foreignMedicineID = uint64(999)

	newSvc := func(updated *bool) CarePlanItemService {
		repo := &mockCarePlanItemRepository{
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID, HospitalizationID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { *updated = true; return nil },
		}
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), rejectMedicineRepo(ownedMedicineID), okProcedureRepo())
	}

	t.Run("rejects cross-clinic medicine_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignMedicineID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateCarePlanItemInput{MedicineID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "care plan item must NOT be updated to reference another clinic's medicine")
	})
}

func rejectDiagnosisTypeRepo(ownedID uint64) repository.DiagnosisTypeRepository {
	return &mockDiagnosisTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("diagnosis_type", "foreign")
		}
		return &model.DiagnosisType{ID: id}, nil
	}}
}

// ── clinical_plan (HIGH): diagnosis_type_id / diagnosis_name_id (x2 slots) ──

func TestClinicalPlanService_Update_RejectsCrossClinicDiagnosisFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(10)
	const foreignTypeID = uint64(999)

	newSvc := func(updated *bool) ClinicalPlanService {
		repo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, mrID uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: mrID}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { *updated = true; return nil },
		}
		return NewClinicalPlanService(repo, okMedRecForPlan(), rejectDiagnosisTypeRepo(ownedTypeID), okDiagnosisNameRepo())
	}

	t.Run("rejects cross-clinic diagnosis_type_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignTypeID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{DiagnosisTypeID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "clinical plan must NOT be updated to reference another clinic's diagnosis type")
	})

	t.Run("rejects cross-clinic diagnosis_2_category_id (second slot)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignTypeID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{Diagnosis2CategoryID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated)
	})

	t.Run("accepts same-clinic diagnosis_type_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedTypeID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{DiagnosisTypeID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

func rejectCheckupTypeRepo(ownedID uint64) repository.CheckupTypeRepository {
	return &mockCheckupTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("checkup_type", "foreign")
		}
		return &model.CheckupType{ID: id}, nil
	}}
}

// ── checkup (MEDIUM, clinical screening record): checkup_type_id ──

func TestCheckupService_Create_RejectsCrossClinicCheckupType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedCheckupTypeID = uint64(10)
	const foreignCheckupTypeID = uint64(999)

	newSvc := func(created *bool) CheckupService {
		repo := &mockCheckupRepository{
			createFn: func(_ context.Context, _ *model.Checkup) error { *created = true; return nil },
			findByIDFn: func(_ context.Context, _, checkupID uint64) (*model.Checkup, error) {
				return &model.Checkup{ID: checkupID}, nil
			},
		}
		medRec := &mockMedicalRecordRepository{findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		}}
		return NewCheckupService(repo, medRec, rejectCheckupTypeRepo(ownedCheckupTypeID), nil, nil)
	}

	t.Run("rejects cross-clinic checkup_type_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), 1, &CreateCheckupInput{ClinicID: clinicID, CheckupTypeID: foreignCheckupTypeID})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "checkup must NOT be persisted referencing another clinic's checkup_type")
	})

	t.Run("accepts same-clinic checkup_type_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), 1, &CreateCheckupInput{ClinicID: clinicID, CheckupTypeID: ownedCheckupTypeID})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func rejectDiagnosisNameRepo(ownedID uint64) repository.DiagnosisNameRepository {
	return &mockDiagnosisNameRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("diagnosis_name", "foreign")
		}
		return &model.DiagnosisName{ID: id}, nil
	}}
}

func TestClinicalPlanService_Update_RejectsCrossClinicDiagnosisName(t *testing.T) {
	const clinicID = uint64(1)
	const ownedNameID = uint64(20)
	const foreignNameID = uint64(888)

	newSvc := func(updated *bool) ClinicalPlanService {
		repo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, mrID uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: mrID}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { *updated = true; return nil },
		}
		return NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), rejectDiagnosisNameRepo(ownedNameID))
	}

	t.Run("rejects cross-clinic diagnosis_name_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignNameID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{DiagnosisNameID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "clinical plan must NOT be updated to reference another clinic's diagnosis name")
	})

	t.Run("rejects cross-clinic diagnosis_2_name_id (second slot)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignNameID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{Diagnosis2NameID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated)
	})

	t.Run("accepts same-clinic diagnosis_name_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedNameID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{DiagnosisNameID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}
