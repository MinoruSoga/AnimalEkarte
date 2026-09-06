package medicalrecord

// Cross-tenant master FK write isolation tests (#201 follow-up) for the services migrated to
// internal/medicalrecord (BE9-2D), ported from internal/service/cross_tenant_master_fk_write_test.go.
// Each fix keeps its dual regression: (i) clinic A referencing clinic B's master is rejected and
// NOT persisted; (ii) same-clinic master still succeeds (no false-reject).
//
// okVaccineRepo/rejectVaccineRepo and okCheckupTypeRepo/rejectCheckupTypeRepo were *moved* here
// (their only consumers — the vaccination/checkup service tests and these sections — all migrated).
// okChiefComplaintTypeRepo/rejectChiefComplaintTypeRepo are *replicated*: the residual
// MedicalRecordService cross-tenant section still uses the internal/service copies, so both exist.
// The lab-import section (BE9-2D sub-batch③) adds okExamTypeRepo/rejectExamTypeRepo,
// okPetRepo/rejectPetRepo and okMedicalRecordRepo/rejectMedicalRecordRepo just below.
// The mock repositories (mockVaccineRepository, mockCheckupTypeRepository, mockChiefComplaintTypeRepository,
// mockMedicalRecordRepository, mockExamTypeRepository, and the lab-local mockPetRepository declared here)
// are the medicalrecord-package copies declared in this file or the sibling test files.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ── shared permissive / rejecting master-repo builders (clinic-scoped FindByID semantics) ──

func okVaccineRepo() VaccineRepository {
	return &mockVaccineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
		return &model.Vaccine{ID: id}, nil
	}}
}

func rejectVaccineRepo(ownedID uint64) VaccineRepository {
	return &mockVaccineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("vaccine", "foreign")
		}
		return &model.Vaccine{ID: id}, nil
	}}
}

func okCheckupTypeRepo() CheckupTypeRepository {
	return &mockCheckupTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
		return &model.CheckupType{ID: id}, nil
	}}
}

func rejectCheckupTypeRepo(ownedID uint64) CheckupTypeRepository {
	return &mockCheckupTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("checkup_type", "foreign")
		}
		return &model.CheckupType{ID: id}, nil
	}}
}

func okChiefComplaintTypeRepo() ChiefComplaintTypeRepository {
	return &mockChiefComplaintTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ChiefComplaintType, error) {
		return &model.ChiefComplaintType{ID: id}, nil
	}}
}

func rejectChiefComplaintTypeRepo(ownedID uint64) ChiefComplaintTypeRepository {
	return &mockChiefComplaintTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ChiefComplaintType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("chief_complaint_type", "foreign")
		}
		return &model.ChiefComplaintType{ID: id}, nil
	}}
}

// ── vaccination (CRITICAL #125): vaccine_id ──

// passthroughTxForCrossTenant is WithTx passthrough for constructor-updated services under test.
type passthroughTxForCrossTenant struct{}

func (passthroughTxForCrossTenant) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type draftMedicalRecordLocker struct{}

func (draftMedicalRecordLocker) LockByIDForUpdate(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
}

func TestVaccinationService_Create_RejectsCrossClinicVaccine(t *testing.T) {
	const clinicID = uint64(1)
	const ownedVaccineID = uint64(10)
	const foreignVaccineID = uint64(999)

	newSvc := func(created *bool) VaccinationService {
		repo := &mockVaccinationRepository{
			createFn: func(_ context.Context, vaccination *model.Vaccination) error {
				*created = true
				vaccination.ID = 1
				return nil
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
				return &model.Vaccination{ID: id, ClinicID: clinicID, VaccineID: ownedVaccineID}, nil
			},
		}
		return newTestVaccinationService(repo, rejectVaccineRepo(ownedVaccineID), nil)
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
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
				*updated = true
				return &model.Vaccination{ID: 1}, nil
			},
		}
		return newTestVaccinationService(repo, rejectVaccineRepo(ownedVaccineID), nil)
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

// ── checkup_type (self-ref parent_id, X-14 batch3) ──

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
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateCheckupTypeInput) (*model.CheckupType, error) {
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

// ── vaccine (self-ref parent_id, X-14 batch3) ──

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
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateVaccineInput) (*model.Vaccine, error) {
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

// ── inquiry (X-14 batch U4): chief_complaint_type_id ──

func TestInquiryService_Save_RejectsCrossClinicChiefComplaintType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(10)
	const foreignTypeID = uint64(999)

	newSvc := func(saved *bool) InquiryService {
		repo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
				*saved = true
				return inquiry, nil
			},
		}
		return NewInquiryService(repo, rejectChiefComplaintTypeRepo(ownedTypeID))
	}

	t.Run("rejects cross-clinic chief_complaint_type_id and does not persist", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		foreign := foreignTypeID
		out, err := svc.Save(context.Background(), UpsertInquiryInput{
			ClinicID: clinicID, MedicalRecordID: 1, ChiefComplaintTypeID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, saved, "inquiry must NOT be persisted referencing another clinic's chief complaint type")
	})

	t.Run("accepts same-clinic chief_complaint_type_id (no false-reject)", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		owned := ownedTypeID
		out, err := svc.Save(context.Background(), UpsertInquiryInput{
			ClinicID: clinicID, MedicalRecordID: 1, ChiefComplaintTypeID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, saved)
	})
}

// ── lab import master-repo builders (BE9-2D sub-batch③, ported from internal/service) ──
//
// okExamTypeRepo/rejectExamTypeRepo reuse the package's existing mockExamTypeRepository
// (exam_type_service_test.go). okMedicalRecordRepo/rejectMedicalRecordRepo reuse the existing
// mockMedicalRecordRepository (service_deps_mock_test.go). okPetRepo/rejectPetRepo use the local
// mockPetRepository declared just below — medicalrecord had no pet mock before the lab move.
// The lab exam service consumes these through the narrow ExamTypeRepository / petFinder /
// medicalRecordFinder consumer-side views, which each mock satisfies structurally.

func okExamTypeRepo() ExamTypeRepository {
	return &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
		return &model.ExaminationType{ID: id}, nil
	}}
}

func rejectExamTypeRepo(ownedID uint64) ExamTypeRepository {
	return &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("exam_type", "foreign")
		}
		return &model.ExaminationType{ID: id}, nil
	}}
}

func examTypeRepoWithOwnedFields(ownedID uint64, fieldIDs ...uint64) ExamTypeRepository {
	items := make([]model.ExamTypeField, 0, len(fieldIDs))
	for _, fieldID := range fieldIDs {
		items = append(items, model.ExamTypeField{ID: fieldID, ExamTypeID: ownedID})
	}
	return &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("exam_type", "foreign")
		}
		return &model.ExaminationType{ID: id, Items: items}, nil
	}}
}

func okPetRepo() petFinder {
	return &mockPetRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
		return &model.Pet{ID: id}, nil
	}}
}

func rejectPetRepo(ownedID uint64) petFinder {
	return &mockPetRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("pet", "foreign")
		}
		return &model.Pet{ID: id}, nil
	}}
}

func okMedicalRecordRepo() medicalRecordFinder {
	return &mockMedicalRecordRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
		return &model.MedicalRecord{ID: id}, nil
	}}
}

func rejectMedicalRecordRepo(ownedID uint64) medicalRecordFinder {
	return &mockMedicalRecordRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("medical_record", "foreign")
		}
		return &model.MedicalRecord{ID: id}, nil
	}}
}

// mockPetRepository satisfies petFinder (FindByID) — the lab exam service's minimal pet view.
// Local medicalrecord copy of internal/service's shared pet mock (BE9-2D sub-batch③).
type mockPetRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
}

func (m *mockPetRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

// ── lab import (X-14 batch U3): ExamTypeID ──
//
// labImportExaminationService.PersistExam/PersistBatch and labResultImportService.Commit
// persisted a request-derived ExamTypeID without verifying it belongs to the caller's
// clinic (same class as examinationService.Create, #124). Guard: examTypeRepo.FindByID
// (ctx, clinicID, ExamTypeID) before dup-check/create, mirroring examination_service.go.
// stubExamRepo/stubDupChecker/newStubLabJobService/syntheticFixtureBatch are shared
// package-level test helpers from lab_import_examination_service_test.go and
// lab_result_import_service_test.go.

func TestLabImportExaminationService_PersistExam_RejectsCrossClinicExamType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const foreignExamTypeID = uint64(999)

	newSvc := func(examRepo *stubExamRepo) *labImportExaminationService {
		return NewLabImportExaminationService(examRepo, &stubDupChecker{}, rejectExamTypeRepo(ownedExamTypeID), okPetRepo(), okMedicalRecordRepo(), passthroughTxForCrossTenant{}).(*labImportExaminationService)
	}

	t.Run("rejects cross-clinic exam_type_id and does not persist", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := newSvc(examRepo)
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:   clinicID,
			ExamTypeID: foreignExamTypeID,
			Date:       time.Now(),
			JobID:      uuid.New(),
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.Empty(t, examRepo.exams, "lab import exam must NOT be persisted referencing another clinic's exam_type")
	})

	t.Run("accepts same-clinic exam_type_id (no false-reject)", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := newSvc(examRepo)
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:   clinicID,
			ExamTypeID: ownedExamTypeID,
			Date:       time.Now(),
			JobID:      uuid.New(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Len(t, examRepo.exams, 1)
	})
}

// ── lab import (P1-2, PR #186 review): PetID / MedicalRecordID ──
//
// persistExam copied a request-derived pet_id/medical_record_id into a new exam
// after only exam_type ownership was checked, letting a lab-import commit link
// a clinic-A exam to clinic-B pet/medical_record data. Guard: petRepo.FindByID /
// medicalRecordRepo.FindByID (ctx, clinicID, id) before Create, mirroring the
// exam_type check above.

func TestLabImportExaminationService_PersistExam_RejectsCrossClinicPet(t *testing.T) {
	const clinicID = uint64(1)
	const ownedPetID = uint64(50)
	const foreignPetID = uint64(999)

	newSvc := func(examRepo *stubExamRepo) *labImportExaminationService {
		return NewLabImportExaminationService(examRepo, &stubDupChecker{}, okExamTypeRepo(), rejectPetRepo(ownedPetID), okMedicalRecordRepo(), passthroughTxForCrossTenant{}).(*labImportExaminationService)
	}

	t.Run("rejects cross-clinic pet_id and does not persist", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := newSvc(examRepo)
		foreign := foreignPetID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:   clinicID,
			ExamTypeID: 10,
			PetID:      &foreign,
			Date:       time.Now(),
			JobID:      uuid.New(),
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.Empty(t, examRepo.exams, "lab import exam must NOT be persisted referencing another clinic's pet")
	})

	t.Run("accepts same-clinic pet_id (no false-reject)", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := newSvc(examRepo)
		owned := ownedPetID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:   clinicID,
			ExamTypeID: 10,
			PetID:      &owned,
			Date:       time.Now(),
			JobID:      uuid.New(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Len(t, examRepo.exams, 1)
	})
}

func TestLabImportExaminationService_PersistExam_RejectsCrossClinicMedicalRecord(t *testing.T) {
	const clinicID = uint64(1)
	const ownedRecordID = uint64(70)
	const foreignRecordID = uint64(999)

	newSvc := func(examRepo *stubExamRepo) *labImportExaminationService {
		return NewLabImportExaminationService(examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), rejectMedicalRecordRepo(ownedRecordID), passthroughTxForCrossTenant{}).(*labImportExaminationService)
	}

	t.Run("rejects cross-clinic medical_record_id and does not persist", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := newSvc(examRepo)
		foreign := foreignRecordID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:        clinicID,
			ExamTypeID:      10,
			MedicalRecordID: &foreign,
			Date:            time.Now(),
			JobID:           uuid.New(),
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.Empty(t, examRepo.exams, "lab import exam must NOT be persisted referencing another clinic's medical_record")
	})

	t.Run("accepts same-clinic medical_record_id (no false-reject)", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := newSvc(examRepo)
		owned := ownedRecordID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:        clinicID,
			ExamTypeID:      10,
			MedicalRecordID: &owned,
			Date:            time.Now(),
			JobID:           uuid.New(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Len(t, examRepo.exams, 1)
	})
}

func TestLabImportExaminationService_PersistBatch_RejectsCrossClinicExamType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const foreignExamTypeID = uint64(999)

	examRepo := newStubExamRepo()
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{}, rejectExamTypeRepo(ownedExamTypeID), okPetRepo(), okMedicalRecordRepo(), passthroughTxForCrossTenant{})

	jobID := uuid.New()
	inputs := []LabExamPersistInput{
		{ClinicID: clinicID, ExamTypeID: foreignExamTypeID, Date: time.Now(), JobID: jobID},
		{ClinicID: clinicID, ExamTypeID: ownedExamTypeID, Date: time.Now(), JobID: jobID},
	}

	results, err := svc.PersistBatch(context.Background(), inputs)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Error(t, results[0].RowError, "foreign exam_type_id row must be recorded as RowError")
	assert.NoError(t, results[1].RowError, "same-clinic exam_type_id row must NOT be rejected")
	assert.Len(t, examRepo.exams, 1, "only the same-clinic row must be persisted")
}

func TestLabResultImportService_Commit_RejectsCrossClinicExamType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const foreignExamTypeID = uint64(999)

	jobSvc := newStubLabJobService()
	examRepo := newStubExamRepo()
	examSvc := NewLabImportExaminationService(examRepo, &stubDupChecker{}, rejectExamTypeRepo(ownedExamTypeID), okPetRepo(), okMedicalRecordRepo(), passthroughTxForCrossTenant{})
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{
		{ClinicID: clinicID, ExamTypeID: foreignExamTypeID, Date: time.Now()},
	}

	resp, err := svc.Commit(context.Background(), clinicID, batch, inputs)
	assert.NoError(t, err, "Commit must not return a function-level error for a per-row rejection")
	assert.Equal(t, 1, resp.FailedCount, "foreign exam_type_id row must be counted as failed, not persisted")
	assert.Equal(t, 0, resp.PersistedCount)
	assert.Empty(t, examRepo.exams, "no exam must be persisted for the foreign exam_type_id row")
}

func TestLabImportExaminationService_PersistExam_RejectsCrossClinicExamTypeField(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const ownedFieldID = uint64(100)
	const foreignFieldID = uint64(999)

	newSvc := func(examRepo *stubExamRepo) *labImportExaminationService {
		return NewLabImportExaminationService(
			examRepo,
			&stubDupChecker{},
			examTypeRepoWithOwnedFields(ownedExamTypeID, ownedFieldID),
			okPetRepo(),
			okMedicalRecordRepo(),
			passthroughTxForCrossTenant{},
		).(*labImportExaminationService)
	}

	t.Run("rejects cross-clinic exam_type_field_id and does not persist", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := newSvc(examRepo)
		foreign := foreignFieldID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:   clinicID,
			ExamTypeID: ownedExamTypeID,
			Date:       time.Now(),
			JobID:      uuid.New(),
			Items:      []LabExamItemInput{{Name: "BUN", ExamTypeFieldID: &foreign}},
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, out)
		assert.Empty(t, examRepo.exams, "lab import exam must NOT be persisted referencing another clinic's exam_type_field")
	})

	t.Run("accepts same-clinic exam_type_field_id (no false-reject)", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := newSvc(examRepo)
		owned := ownedFieldID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:   clinicID,
			ExamTypeID: ownedExamTypeID,
			Date:       time.Now(),
			JobID:      uuid.New(),
			Items:      []LabExamItemInput{{Name: "BUN", ExamTypeFieldID: &owned}},
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Len(t, examRepo.exams, 1)
	})
}

func TestLabImportExaminationService_PersistBatch_RejectsCrossClinicExamTypeField(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const ownedFieldID = uint64(100)
	const foreignFieldID = uint64(999)

	examRepo := newStubExamRepo()
	svc := NewLabImportExaminationService(
		examRepo,
		&stubDupChecker{},
		examTypeRepoWithOwnedFields(ownedExamTypeID, ownedFieldID),
		okPetRepo(),
		okMedicalRecordRepo(),
		passthroughTxForCrossTenant{},
	)

	owned := ownedFieldID
	foreign := foreignFieldID
	jobID := uuid.New()
	inputs := []LabExamPersistInput{
		{ClinicID: clinicID, ExamTypeID: ownedExamTypeID, Date: time.Now(), JobID: jobID, Items: []LabExamItemInput{{Name: "ALT", ExamTypeFieldID: &foreign}}},
		{ClinicID: clinicID, ExamTypeID: ownedExamTypeID, Date: time.Now(), JobID: jobID, Items: []LabExamItemInput{{Name: "BUN", ExamTypeFieldID: &owned}}},
	}

	results, err := svc.PersistBatch(context.Background(), inputs)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Error(t, results[0].RowError, "foreign exam_type_field_id row must be recorded as RowError")
	assert.True(t, apperrors.IsInvalidInput(results[0].RowError))
	assert.NoError(t, results[1].RowError, "same-clinic exam_type_field_id row must NOT be rejected")
	assert.Len(t, examRepo.exams, 1, "only the same-clinic field row must be persisted")
}

func TestLabResultImportService_Commit_RejectsCrossClinicExamTypeField(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const ownedFieldID = uint64(100)
	const foreignFieldID = uint64(999)

	jobSvc := newStubLabJobService()
	examRepo := newStubExamRepo()
	examSvc := NewLabImportExaminationService(
		examRepo,
		&stubDupChecker{},
		examTypeRepoWithOwnedFields(ownedExamTypeID, ownedFieldID),
		okPetRepo(),
		okMedicalRecordRepo(),
		passthroughTxForCrossTenant{},
	)
	svc := NewLabResultImportService(jobSvc, examSvc)

	foreign := foreignFieldID
	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{
		{ClinicID: clinicID, ExamTypeID: ownedExamTypeID, Date: time.Now(), Items: []LabExamItemInput{{Name: "BUN", ExamTypeFieldID: &foreign}}},
	}

	resp, err := svc.Commit(context.Background(), clinicID, batch, inputs)
	assert.NoError(t, err, "Commit must not return a function-level error for a per-row rejection")
	assert.Equal(t, 1, resp.FailedCount, "foreign exam_type_field_id row must be counted as failed, not persisted")
	assert.Equal(t, 0, resp.PersistedCount)
	assert.Empty(t, examRepo.exams, "no exam must be persisted for the foreign exam_type_field_id row")
}

// ── clinical_plan (HIGH): diagnosis_type_id / diagnosis_name_id (x2 slots) ──
// BE9-2D sub-batch④a: clinicalPlanService とともに internal/service/cross_tenant_master_fk_write_test.go
// から移設。診断マスタ FK builder は in-package の DiagnosisTypeRepository/DiagnosisNameRepository を返す
// （service 側の同名 builder は medical_record_subrecords_test 等の residual consumer のため残置）。

func okDiagnosisTypeRepo() DiagnosisTypeRepository {
	return &mockDiagnosisTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
		return &model.DiagnosisType{ID: id}, nil
	}}
}

func okDiagnosisNameRepo() DiagnosisNameRepository {
	return &mockDiagnosisNameRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
		return &model.DiagnosisName{ID: id}, nil
	}}
}

func rejectDiagnosisTypeRepo(ownedID uint64) DiagnosisTypeRepository {
	return &mockDiagnosisTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("diagnosis_type", "foreign")
		}
		return &model.DiagnosisType{ID: id}, nil
	}}
}

func rejectDiagnosisNameRepo(ownedID uint64) DiagnosisNameRepository {
	return &mockDiagnosisNameRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("diagnosis_name", "foreign")
		}
		return &model.DiagnosisName{ID: id}, nil
	}}
}

func TestClinicalPlanService_Update_RejectsCrossClinicDiagnosisFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(10)
	const foreignTypeID = uint64(999)

	newSvc := func(updated *bool) ClinicalPlanService {
		repo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, mrID uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: mrID}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error { *updated = true; return nil },
		}
		return NewClinicalPlanService(repo, okMedRecForPlan(), rejectDiagnosisTypeRepo(ownedTypeID), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})
	}

	t.Run("rejects cross-clinic diagnosis_type_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignTypeID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{DiagnosisTypeID: &foreign,
			ActorID: clinicalPlanTestActorID(),
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "clinical plan must NOT be updated to reference another clinic's diagnosis type")
	})

	t.Run("rejects cross-clinic diagnosis_2_type_id (second slot)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignTypeID
		foreignPtr := &foreign
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{Diagnosis2TypeID: &foreignPtr,
			ActorID: clinicalPlanTestActorID(),
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated)
	})

	t.Run("accepts same-clinic diagnosis_type_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedTypeID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{DiagnosisTypeID: &owned,
			ActorID: clinicalPlanTestActorID(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
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
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error { *updated = true; return nil },
		}
		return NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), rejectDiagnosisNameRepo(ownedNameID), &mockCheckupTransactor{}, &mockAuditTxLogger{})
	}

	t.Run("rejects cross-clinic diagnosis_name_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignNameID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{DiagnosisNameID: &foreign,
			ActorID: clinicalPlanTestActorID(),
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "clinical plan must NOT be updated to reference another clinic's diagnosis name")
	})

	t.Run("rejects cross-clinic diagnosis_2_name_id (second slot)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignNameID
		foreignPtr := &foreign
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{Diagnosis2NameID: &foreignPtr,
			ActorID: clinicalPlanTestActorID(),
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated)
	})

	t.Run("accepts same-clinic diagnosis_name_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedNameID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateClinicalPlanInput{DiagnosisNameID: &owned,
			ActorID: clinicalPlanTestActorID(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}
