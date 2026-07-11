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
	"time"

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

func okHospitalizationPlanRepo() repository.HospitalizationPlanRepository {
	return &mockHospitalizationPlanRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.HospitalizationPlan, error) {
		return &model.HospitalizationPlan{ID: id}, nil
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

func okTrimmingCourseRepo() repository.TrimmingCourseRepository {
	return &mockTrimmingCourseRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		return &model.TrimmingCourse{ID: id}, nil
	}}
}

func okTrimmingOptionRepo() repository.TrimmingOptionRepository {
	return &mockTrimmingOptionRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		return &model.TrimmingOption{ID: id}, nil
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

func rejectHospitalizationPlanRepo(ownedID uint64) repository.HospitalizationPlanRepository {
	return &mockHospitalizationPlanRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.HospitalizationPlan, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("hospitalization_plan", "foreign")
		}
		return &model.HospitalizationPlan{ID: id}, nil
	}}
}

func rejectCageRepo(ownedID uint64) repository.CageRepository {
	return &mockCageRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("cage", "foreign")
		}
		return &model.Cage{ID: id}, nil
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
		return NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, rejectExamTypeRepo(ownedExamTypeID), nil, &mockCheckupTransactor{})
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
		return NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, rejectExamTypeRepo(ownedExamTypeID), nil, &mockCheckupTransactor{})
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
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), rejectMedicineRepo(ownedMedicineID), okProcedureRepo(), okHospitalizationPlanRepo())
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
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), rejectMedicineRepo(ownedMedicineID), okProcedureRepo(), okHospitalizationPlanRepo())
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

// ── care_plan_item (X-14): hospitalization_plan_id ──

func TestCarePlanItemService_Create_RejectsCrossClinicHospitalizationPlanFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedPlanID = uint64(10)
	const foreignPlanID = uint64(999)

	newSvc := func(created *bool) CarePlanItemService {
		repo := &mockCarePlanItemRepository{
			createFn: func(_ context.Context, _ *model.CarePlanItem) error { *created = true; return nil },
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID}, nil
			},
		}
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), rejectHospitalizationPlanRepo(ownedPlanID))
	}

	t.Run("rejects cross-clinic hospitalization_plan_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignPlanID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeFood), Name: "x", HospitalizationPlanID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "care plan item must NOT be persisted referencing another clinic's hospitalization plan")
	})

	t.Run("accepts same-clinic hospitalization_plan_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedPlanID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeFood), Name: "x", HospitalizationPlanID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCarePlanItemService_Update_RejectsCrossClinicHospitalizationPlanFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedPlanID = uint64(10)
	const foreignPlanID = uint64(999)

	newSvc := func(updated *bool) CarePlanItemService {
		repo := &mockCarePlanItemRepository{
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID, HospitalizationID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error { *updated = true; return nil },
		}
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), rejectHospitalizationPlanRepo(ownedPlanID))
	}

	t.Run("rejects cross-clinic hospitalization_plan_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignPlanID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateCarePlanItemInput{HospitalizationPlanID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "care plan item must NOT be updated to reference another clinic's hospitalization plan")
	})

	t.Run("accepts same-clinic hospitalization_plan_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedPlanID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateCarePlanItemInput{HospitalizationPlanID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── hospitalization (X-14): cage_id ──

func TestHospitalizationService_Create_RejectsCrossClinicCageFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedCageID = uint64(10)
	const foreignCageID = uint64(999)

	newSvc := func(created *bool) HospitalizationService {
		repo := &mockHospitalizationRepository{
			createFn: func(_ context.Context, _ *model.Hospitalization) error { *created = true; return nil },
		}
		return NewHospitalizationService(&repository.Repositories{
			Hospitalization: repo,
			Cage:            rejectCageRepo(ownedCageID),
		})
	}

	t.Run("rejects cross-clinic cage_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignCageID
		out, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			OwnerID: 2, PetID: 5, CageID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "hospitalization must NOT be persisted referencing another clinic's cage")
	})

	t.Run("accepts same-clinic cage_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedCageID
		out, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			OwnerID: 2, PetID: 5, CageID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestHospitalizationService_Update_RejectsCrossClinicCageFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedCageID = uint64(10)
	const foreignCageID = uint64(999)

	newSvc := func(updated *bool) HospitalizationService {
		repo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{ID: id, ClinicID: clinicID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
				*updated = true
				return &model.Hospitalization{ID: 1, ClinicID: clinicID}, nil
			},
		}
		return NewHospitalizationService(&repository.Repositories{
			Hospitalization: repo,
			Cage:            rejectCageRepo(ownedCageID),
		})
	}

	t.Run("rejects cross-clinic cage_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignCageID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{CageID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "hospitalization must NOT be updated to reference another clinic's cage")
	})

	t.Run("accepts same-clinic cage_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedCageID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{CageID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
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

func rejectTrimmingCourseRepo(ownedID uint64) repository.TrimmingCourseRepository {
	return &mockTrimmingCourseRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_course", "foreign")
		}
		return &model.TrimmingCourse{ID: id}, nil
	}}
}

func rejectTrimmingOptionRepo(ownedID uint64) repository.TrimmingOptionRepository {
	return &mockTrimmingOptionRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_option", "foreign")
		}
		return &model.TrimmingOption{ID: id}, nil
	}}
}

// ── billing_item (P1, X-4): trimming_course_id / trimming_option_id ──

func TestBillingItemService_CreateItem_RejectsCrossClinicTrimmingFK(t *testing.T) {
	const clinicID = uint64(1)
	const billingID = uint64(10)
	const ownedCourseID = uint64(300)
	const foreignCourseID = uint64(999)
	const ownedOptionID = uint64(400)
	const foreignOptionID = uint64(998)

	newSvc := func(created *bool, courseRepo repository.TrimmingCourseRepository, optionRepo repository.TrimmingOptionRepository) BillingItemService {
		repo := defaultMockBillingItemRepo()
		repo.createFn = func(_ context.Context, item *model.BillingItem) error { *created = true; item.ID = 1; return nil }
		billingRepo := &mockAccountingRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id}, nil
		}}
		return NewBillingItemService(repo, billingRepo, defaultMockTreatmentRepo(), &mockTransactor{}, courseRepo, optionRepo)
	}

	baseInput := func() *CreateBillingItemInput {
		return &CreateBillingItemInput{
			ClinicID:  clinicID,
			BillingID: billingID,
			Category:  string(model.ItemCategoryProcedure),
			Name:      "トリミング",
			UnitPrice: 5000,
			Quantity:  1,
		}
	}

	t.Run("rejects cross-clinic trimming_course_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID), okTrimmingOptionRepo())
		foreign := foreignCourseID
		input := baseInput()
		input.TrimmingCourseID = &foreign
		out, err := svc.CreateItem(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "billing item must NOT be persisted referencing another clinic's trimming course")
	})

	t.Run("rejects cross-clinic trimming_option_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, okTrimmingCourseRepo(), rejectTrimmingOptionRepo(ownedOptionID))
		foreign := foreignOptionID
		input := baseInput()
		input.TrimmingOptionID = &foreign
		out, err := svc.CreateItem(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "billing item must NOT be persisted referencing another clinic's trimming option")
	})

	t.Run("accepts same-clinic trimming_course_id/trimming_option_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID), rejectTrimmingOptionRepo(ownedOptionID))
		owned1, owned2 := ownedCourseID, ownedOptionID
		input := baseInput()
		input.TrimmingCourseID = &owned1
		input.TrimmingOptionID = &owned2
		out, err := svc.CreateItem(context.Background(), input)
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

// ── accounting (G14-1): payment_method_id ──
//
// payment_method_id is not guarded by a plain FindByID(clinicID, id) call like the
// other master FKs above — it is guarded by resolvePaymentMethodMasterID
// (accounting_service_builders.go), which resolves the request's method(ENUM) to
// the caller clinic's payment_methods master id and rejects any explicitly
// supplied id that does not match ("... or equivalent" per the lint's own
// guarded definition). This test proves that guard holds against a *real*
// foreign clinic's resolved master id, not just an arbitrary mismatched number.

func TestAccountingService_Update_RejectsForeignPaymentMethodID(t *testing.T) {
	const clinicA = uint64(1)
	const clinicB = uint64(2)
	const clinicACashID = uint64(101)
	const clinicBCashID = uint64(201)
	billingAmount := int64(5000)
	cashKey := "cash"

	payMethodRepo := &mockPaymentMethodMasterRepository{
		findAllFn: func(_ context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
			switch clinicID {
			case clinicA:
				return []model.PaymentMethodMaster{{ID: clinicACashID, ClinicID: clinicA, SystemKey: &cashKey}}, nil
			case clinicB:
				return []model.PaymentMethodMaster{{ID: clinicBCashID, ClinicID: clinicB, SystemKey: &cashKey}}, nil
			default:
				return nil, nil
			}
		},
	}

	newSvc := func(saved *bool) AccountingService {
		repo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
				return &model.Billing{ID: id, ClinicID: clinicID}, nil
			},
			updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
				return &model.Billing{ID: id, ClinicID: clinicID}, nil
			},
			savePaymentSplitsFn: func(_ context.Context, _ []model.PaymentSplit) error {
				*saved = true
				return nil
			},
		}
		return NewAccountingService(repo, nil, &mockTransactor{}, &mockAccountingAuditService{}, payMethodRepo)
	}

	t.Run("rejects clinic B's payment_method_id on clinic A's billing update and does not persist", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		foreign := clinicBCashID // clinic B の cash master id（実在するが clinic A のものではない）
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      clinicA,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, PaymentMethodID: &foreign, Amount: 5000, ReceivedAmount: 5000},
			},
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, out)
		assert.False(t, saved, "billing must NOT be persisted with another clinic's payment_method_id")
	})

	t.Run("accepts clinic A's own payment_method_id (no false-reject)", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		owned := clinicACashID
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      clinicA,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, PaymentMethodID: &owned, Amount: 5000, ReceivedAmount: 5000},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, saved)
	})
}

func rejectMerchandiseItemRepo(ownedID uint64) repository.MerchandiseItemRepository {
	return &mockMerchandiseItemRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("merchandise_item", "foreign")
		}
		return &model.MerchandiseItem{ID: id}, nil
	}}
}

// ── campaign (target-item FK): TargetItemIDs → campaign_target_items.merchandise_item_id ──

func TestCampaignService_Create_RejectsCrossClinicTargetItemFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedItemID = uint64(10)
	const foreignItemID = uint64(999)

	newSvc := func(created *bool) CampaignService {
		repo := &mockCampaignRepository{
			createFn: func(_ context.Context, m *model.Campaign) (*model.Campaign, error) {
				*created = true
				m.ID = 1
				return m, nil
			},
		}
		return NewCampaignService(repo, rejectMerchandiseItemRepo(ownedItemID))
	}

	baseInput := func(itemID uint64) *CreateCampaignInput {
		return &CreateCampaignInput{
			Name:          "Autumn Sale",
			StartDate:     time.Now(),
			EndDate:       time.Now().Add(24 * time.Hour),
			DiscountType:  model.CampaignDiscountTypeRate,
			DiscountValue: 10.0,
			TargetItemIDs: []uint64{itemID},
		}
	}

	t.Run("rejects cross-clinic target_item_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, baseInput(foreignItemID))
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "campaign must NOT be persisted referencing another clinic's merchandise item")
	})

	t.Run("accepts same-clinic target_item_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, baseInput(ownedItemID))
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCampaignService_Update_RejectsCrossClinicTargetItemFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedItemID = uint64(10)
	const foreignItemID = uint64(999)

	newSvc := func(replaced *bool) CampaignService {
		current := &model.Campaign{ID: 100, ClinicID: clinicID}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
			replaceTargetsFn: func(_ context.Context, _ uint64, _ []model.ItemCategory, _ []uint64) error {
				*replaced = true
				return nil
			},
		}
		return NewCampaignService(repo, rejectMerchandiseItemRepo(ownedItemID))
	}

	t.Run("rejects cross-clinic target_item_id on update and does not persist", func(t *testing.T) {
		replaced := false
		svc := newSvc(&replaced)
		ids := []uint64{foreignItemID}
		out, err := svc.Update(context.Background(), clinicID, 100, &UpdateCampaignInput{TargetItemIDs: &ids})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, replaced, "campaign targets must NOT be replaced referencing another clinic's merchandise item")
	})

	t.Run("accepts same-clinic target_item_id (no false-reject)", func(t *testing.T) {
		replaced := false
		svc := newSvc(&replaced)
		ids := []uint64{ownedItemID}
		out, err := svc.Update(context.Background(), clinicID, 100, &UpdateCampaignInput{TargetItemIDs: &ids})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, replaced)
	})
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

// ── X-14 batch: self-ref ParentID ownership guard ──
// (checkupType/consultation/examType/procedure/vaccine)
//
// Each of these five master-data services carries a self-referencing ParentID (a
// sub-category pointing at its own parent row in the same table). Prior to this batch,
// Create/Update persisted a request-supplied ParentID without verifying it belongs to
// the caller's clinic. Each service has a single repo dependency (self-ref), so the
// mock below wires findByIDFn (used both for the parent-ownership guard and, on Update,
// the pre-existing self-entity existence check) alongside createFn/updateFieldsFn.

func TestCheckupTypeService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) CheckupTypeService {
		repo := &mockCheckupTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("checkup_type", "foreign")
				}
				return &model.CheckupType{ID: id}, nil
			},
			createFn: func(_ context.Context, _ *model.CheckupType) error { *created = true; return nil },
		}
		return NewCheckupTypeService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateCheckupTypeInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "checkup type must NOT be persisted referencing another clinic's parent checkup type")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateCheckupTypeInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCheckupTypeService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) CheckupTypeService {
		repo := &mockCheckupTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
				if id == entityID || id == ownedParentID {
					return &model.CheckupType{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("checkup_type", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.CheckupType, error) {
				*updated = true
				return &model.CheckupType{ID: id}, nil
			},
		}
		return NewCheckupTypeService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateCheckupTypeInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "checkup type must NOT be updated to reference another clinic's parent checkup type")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateCheckupTypeInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

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
			updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Consultation, error) {
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

func TestExamTypeService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) ExaminationTypeService {
		repo := &mockExamTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("exam_type", "foreign")
				}
				return &model.ExaminationType{ID: id}, nil
			},
			createFn: func(_ context.Context, _ *model.ExaminationType) error { *created = true; return nil },
		}
		return NewExamTypeService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateExamTypeInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "exam type must NOT be persisted referencing another clinic's parent exam type")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateExamTypeInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestExamTypeService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) ExaminationTypeService {
		repo := &mockExamTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
				if id == entityID || id == ownedParentID {
					return &model.ExaminationType{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("exam_type", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.ExaminationType, error) {
				*updated = true
				return &model.ExaminationType{ID: id}, nil
			},
		}
		return NewExamTypeService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateExamTypeInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "exam type must NOT be updated to reference another clinic's parent exam type")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateExamTypeInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

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
		return NewProcedureService(repo)
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
			updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Procedure, error) {
				*updated = true
				return &model.Procedure{ID: id}, nil
			},
		}
		return NewProcedureService(repo)
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

func TestVaccineService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) VaccineService {
		repo := &mockVaccineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("vaccine", "foreign")
				}
				return &model.Vaccine{ID: id}, nil
			},
			createFn: func(_ context.Context, _ *model.Vaccine) error { *created = true; return nil },
		}
		return NewVaccineService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateVaccineInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "vaccine must NOT be persisted referencing another clinic's parent vaccine")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateVaccineInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestVaccineService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) VaccineService {
		repo := &mockVaccineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
				if id == entityID || id == ownedParentID {
					return &model.Vaccine{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("vaccine", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Vaccine, error) {
				*updated = true
				return &model.Vaccine{ID: id}, nil
			},
		}
		return NewVaccineService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateVaccineInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "vaccine must NOT be updated to reference another clinic's parent vaccine")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateVaccineInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}
