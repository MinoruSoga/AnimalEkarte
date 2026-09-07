package medicalrecord

// medical_record_cross_tenant_master_fk_write_test.go — BE9-2D ⑦: internal/service
// cross_tenant_master_fk_write_test.go の MedicalRecordService CreateSubRecords 節
// （ChiefComplaint/Diagnosis 4スロット FK guard 3テスト）を同名のまま縦移動。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestMedicalRecordService_CreateSubRecords_RejectsCrossClinicChiefComplaintType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(10)
	const foreignTypeID = uint64(999)

	newSvc := func(saved *bool) *medicalRecordService {
		inquiryRepo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
				*saved = true
				return inquiry, nil
			},
		}
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, mrID uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: mrID}, nil
			},
		}
		return &medicalRecordService{
			inquiryRepo:            inquiryRepo,
			clinicalPlanRepo:       clinicalPlanRepo,
			chiefComplaintTypeRepo: rejectChiefComplaintTypeRepo(ownedTypeID),
			diagTypeRepo:           okDiagnosisTypeRepo(),
			diagNameRepo:           okDiagnosisNameRepo(),
		}
	}

	t.Run("rejects cross-clinic chief_complaint_type_id and does not upsert inquiry", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		foreign := foreignTypeID
		_ = svc.CreateSubRecords(context.Background(), clinicID, 1, CreateSubRecordsInput{ChiefComplaintTypeID: &foreign})
		assert.False(t, saved, "inquiry must NOT be persisted referencing another clinic's chief complaint type")
	})

	t.Run("accepts same-clinic chief_complaint_type_id (no false-reject)", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		owned := ownedTypeID
		_ = svc.CreateSubRecords(context.Background(), clinicID, 1, CreateSubRecordsInput{ChiefComplaintTypeID: &owned})
		assert.True(t, saved)
	})
}

func TestMedicalRecordService_CreateSubRecords_RejectsCrossClinicDiagnosisFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(10)
	const foreignTypeID = uint64(999)

	newSvc := func(updated *bool) *medicalRecordService {
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, mrID uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: mrID}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
				*updated = true
				return nil
			},
		}
		return &medicalRecordService{
			inquiryRepo:            &mockInquiryRepository{},
			clinicalPlanRepo:       clinicalPlanRepo,
			chiefComplaintTypeRepo: okChiefComplaintTypeRepo(),
			diagTypeRepo:           rejectDiagnosisTypeRepo(ownedTypeID),
			diagNameRepo:           okDiagnosisNameRepo(),
		}
	}

	t.Run("rejects cross-clinic diagnosis_1_category_id and does not update clinical plan", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignTypeID
		_ = svc.CreateSubRecords(context.Background(), clinicID, 1, CreateSubRecordsInput{Diagnosis1CategoryID: &foreign})
		assert.False(t, updated, "clinical plan must NOT be updated to reference another clinic's diagnosis type")
	})

	t.Run("accepts same-clinic diagnosis_1_category_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedTypeID
		_ = svc.CreateSubRecords(context.Background(), clinicID, 1, CreateSubRecordsInput{Diagnosis1CategoryID: &owned})
		assert.True(t, updated)
	})
}

func TestMedicalRecordService_CreateSubRecords_RejectsCrossClinicDiagnosisNameFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedNameID = uint64(20)
	const foreignNameID = uint64(888)

	newSvc := func(updated *bool) *medicalRecordService {
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, mrID uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: mrID}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
				*updated = true
				return nil
			},
		}
		return &medicalRecordService{
			inquiryRepo:            &mockInquiryRepository{},
			clinicalPlanRepo:       clinicalPlanRepo,
			chiefComplaintTypeRepo: okChiefComplaintTypeRepo(),
			diagTypeRepo:           okDiagnosisTypeRepo(),
			diagNameRepo:           rejectDiagnosisNameRepo(ownedNameID),
		}
	}

	t.Run("rejects cross-clinic diagnosis_1_name_id and does not update clinical plan", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignNameID
		_ = svc.CreateSubRecords(context.Background(), clinicID, 1, CreateSubRecordsInput{Diagnosis1NameID: &foreign})
		assert.False(t, updated, "clinical plan must NOT be updated to reference another clinic's diagnosis name")
	})

	t.Run("accepts same-clinic diagnosis_1_name_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedNameID
		_ = svc.CreateSubRecords(context.Background(), clinicID, 1, CreateSubRecordsInput{Diagnosis1NameID: &owned})
		assert.True(t, updated)
	})
}

// ── owner/pet InsuranceID (X-14 batch U5) ──
//
// ownerService.CreateWithPets persisted a nested, request-derived InsuranceID
// (input.Pets[i].InsuranceID) without verifying it belongs to the caller's clinic —
// buildOwnerPetModels maps it straight onto model.Pet. Guard: insuranceRepo.FindByID
// (ctx, clinicID, InsuranceID) for every pet before repo.CreateWithPets.
//
// petService.Create/Update already carry the same FindByID guard (pet_service.go);
// they lacked a DEDICATED isolation test distinguishing same-clinic vs cross-clinic
// IDs (TestPetService_Update_InsuranceValidation above always returns NotFound
// regardless of ID). These tests supply that missing runtime evidence so the
// allowlist can move from known-unguarded to guarded.
