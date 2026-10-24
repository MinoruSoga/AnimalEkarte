package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCreateSubRecords(t *testing.T) {
	t.Run("no inquiry input: SaveByMedicalRecordID is not called", func(t *testing.T) {
		saveCalled := false
		inquiryRepo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, _ InquiryUpsertFields) (*model.Inquiry, error) {
				saveCalled = true
				return &model.Inquiry{}, nil
			},
		}
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1}, nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: inquiryRepo, clinicalPlanRepo: clinicalPlanRepo}

		_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{})

		assert.False(t, saveCalled, "inquiry upsert should be skipped when no inquiry fields are provided")
	})

	t.Run("inquiry input present: SaveByMedicalRecordID is called with mapped fields", func(t *testing.T) {
		var savedInquiry *InquiryUpsertFields
		typeID := uint64(3)
		complaint := "食欲不振"
		notes := "2日前から"
		inquiryRepo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, fields InquiryUpsertFields) (*model.Inquiry, error) {
				savedInquiry = &fields
				return &model.Inquiry{}, nil
			},
		}
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1}, nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: inquiryRepo, clinicalPlanRepo: clinicalPlanRepo, chiefComplaintTypeRepo: okChiefComplaintTypeRepo()}

		_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{
			ChiefComplaintTypeID: &typeID,
			ChiefComplaint:       &complaint,
			Notes:                &notes,
		})

		if assert.NotNil(t, savedInquiry) {
			assert.Equal(t, uint64(10), savedInquiry.MedicalRecordID)
			if assert.NotNil(t, savedInquiry.ChiefComplaintTypeID) {
				assert.Equal(t, &typeID, *savedInquiry.ChiefComplaintTypeID)
			}
			if assert.NotNil(t, savedInquiry.ChiefComplaint) {
				assert.Equal(t, complaint, *savedInquiry.ChiefComplaint)
			}
			if assert.NotNil(t, savedInquiry.Notes) {
				assert.Equal(t, notes, *savedInquiry.Notes)
			}
		}
	})

	t.Run("inquiry upsert failure is logged and swallowed (best-effort)", func(t *testing.T) {
		complaint := "嘔吐"
		inquiryRepo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, _ InquiryUpsertFields) (*model.Inquiry, error) {
				return nil, errors.New("db error")
			},
		}
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1}, nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: inquiryRepo, clinicalPlanRepo: clinicalPlanRepo}

		assert.NotPanics(t, func() {
			_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{ChiefComplaint: &complaint})
		})
	})

	t.Run("clinical plan lookup returns a non-not-found error: aborts without creating or updating", func(t *testing.T) {
		findCalls := 0
		createCalled := false
		updateCalled := false
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				findCalls++
				return nil, errors.New("db down")
			},
			createFn: func(_ context.Context, _ *model.ClinicalPlan) error {
				createCalled = true
				return nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
				updateCalled = true
				return nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		plan := "policy"
		_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Plan: &plan})

		assert.Equal(t, 1, findCalls)
		assert.False(t, createCalled)
		assert.False(t, updateCalled)
	})

	t.Run("clinical plan not found: creates a new empty plan", func(t *testing.T) {
		var createdPlan *model.ClinicalPlan
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return nil, apperrors.WrapNotFound("clinical_plan", "10")
			},
			createFn: func(_ context.Context, plan *model.ClinicalPlan) error {
				createdPlan = plan
				return nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{})

		if assert.NotNil(t, createdPlan) {
			assert.Equal(t, uint64(10), createdPlan.MedicalRecordID)
		}
	})

	t.Run("clinical plan not found and create fails: aborts without updating", func(t *testing.T) {
		updateCalled := false
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return nil, apperrors.WrapNotFound("clinical_plan", "10")
			},
			createFn: func(_ context.Context, _ *model.ClinicalPlan) error {
				return errors.New("create failed")
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
				updateCalled = true
				return nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		plan := "policy"
		_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Plan: &plan})

		assert.False(t, updateCalled)
	})

	t.Run("no plan-related fields: Update is not called", func(t *testing.T) {
		updateCalled := false
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
				updateCalled = true
				return nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{})

		assert.False(t, updateCalled)
	})

	t.Run("plan-related fields present: Update gets the scalars and diagnosis rows are replaced", func(t *testing.T) {
		var updated UpdateClinicalPlanInput
		var replaced []ClinicalPlanDiagnosisInput
		var createdWithClinic uint64
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, cmd UpdateClinicalPlanInput) error {
				updated = cmd
				return nil
			},
			replaceDiagnosesFn: func(_ context.Context, clinicID, planID uint64, rows []ClinicalPlanDiagnosisInput) error {
				createdWithClinic = clinicID
				assert.Equal(t, uint64(7), planID)
				replaced = rows
				return nil
			},
		}
		svc := &medicalRecordService{
			inquiryRepo:      &mockInquiryRepository{},
			clinicalPlanRepo: clinicalPlanRepo,
			diagTypeRepo:     okDiagnosisTypeRepo(),
			// AUD-007: name.DiagnosisTypeID が対応 type と一致するよう返す
			diagNameRepo: &mockDiagnosisNameRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
					typeID := id // default
					switch id {
					case 2:
						typeID = 1
					case 4:
						typeID = 3
					}
					return &model.DiagnosisName{ID: id, DiagnosisTypeID: typeID}, nil
				},
			},
		}

		_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{
			Plan:       strPtr("policy"),
			Assessment: strPtr("assessment"),
			Diagnoses: []ClinicalPlanDiagnosisInput{
				{Slot: 1, DiagnosisTypeID: uint64Ptr(1), DiagnosisNameID: uint64Ptr(2), DiagnosisNameSnapshot: strPtr("snap-1")},
				{Slot: 2, DiagnosisTypeID: uint64Ptr(3), DiagnosisNameID: uint64Ptr(4)},
			},
		})

		require.NotNil(t, updated.TreatmentPolicy)
		require.NotNil(t, updated.DiagnosisDetails)
		assert.Equal(t, "policy", *updated.TreatmentPolicy)
		assert.Equal(t, "assessment", *updated.DiagnosisDetails)
		assert.Nil(t, updated.Diagnoses, "diagnosis rows are written by ReplaceDiagnoses, not the plan column update")
		require.Len(t, replaced, 2)
		assert.Equal(t, uint64(1), createdWithClinic)
		require.NotNil(t, replaced[0].DiagnosisNameSnapshot)
		assert.Equal(t, "snap-1", *replaced[0].DiagnosisNameSnapshot, "snapshots are stored as sent")
	})

	// EMR-495 F-01: 診断行だけの入力でも plan の version を CAS つきで進め、診断行の置換より先に行う。
	t.Run("diagnosis-only input advances the plan version under CAS before replacing rows", func(t *testing.T) {
		var order []string
		var gotCmd UpdateClinicalPlanInput
		var gotExpected *int
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7, Version: 3}, nil
			},
			updateWithVersionFn: func(_ context.Context, _, planID uint64, cmd UpdateClinicalPlanInput, expectedVersion *int) error {
				order = append(order, "update")
				assert.Equal(t, uint64(7), planID)
				gotCmd = cmd
				gotExpected = expectedVersion
				return nil
			},
			replaceDiagnosesFn: func(_ context.Context, _, _ uint64, _ []ClinicalPlanDiagnosisInput) error {
				order = append(order, "replace")
				return nil
			},
		}
		svc := &medicalRecordService{
			inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo,
			diagTypeRepo: okDiagnosisTypeRepo(), diagNameRepo: okDiagnosisNameRepo(),
		}

		err := svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{
			Diagnoses: []ClinicalPlanDiagnosisInput{{Slot: 1}},
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"update", "replace"}, order)
		require.NotNil(t, gotExpected)
		assert.Equal(t, 3, *gotExpected, "CAS は直前に読んだ version に対して行う")
		require.NotNil(t, gotCmd.persistVersion)
		assert.Equal(t, 4, *gotCmd.persistVersion)
		assert.Nil(t, gotCmd.TreatmentPolicy, "本文列は触らず version だけを進める")
		assert.Nil(t, gotCmd.DiagnosisDetails)
	})

	t.Run("diagnosis-only input leaves rows untouched when the plan version CAS is lost", func(t *testing.T) {
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7, Version: 3}, nil
			},
			updateWithVersionFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput, _ *int) error {
				return apperrors.WrapConflict("version mismatch")
			},
			replaceDiagnosesFn: func(_ context.Context, _, _ uint64, _ []ClinicalPlanDiagnosisInput) error {
				t.Fatal("must not replace rows after losing the plan version CAS")
				return nil
			},
		}
		svc := &medicalRecordService{
			inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo,
			diagTypeRepo: okDiagnosisTypeRepo(), diagNameRepo: okDiagnosisNameRepo(),
		}

		err := svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{
			Diagnoses: []ClinicalPlanDiagnosisInput{{Slot: 1}},
		})

		require.Error(t, err)
	})

	t.Run("plan created from sub-records carries the trusted clinic_id", func(t *testing.T) {
		var createdPlan *model.ClinicalPlan
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return nil, apperrors.WrapNotFound("clinical_plan", "10")
			},
			createFn: func(_ context.Context, plan *model.ClinicalPlan) error {
				createdPlan = plan
				return nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		_ = svc.CreateSubRecords(context.Background(), 42, 10, CreateSubRecordsInput{})

		require.NotNil(t, createdPlan)
		assert.Equal(t, uint64(42), createdPlan.ClinicID)
	})

	t.Run("more than the diagnosis row cap skips every plan write", func(t *testing.T) {
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
				t.Fatal("must not update")
				return nil
			},
			replaceDiagnosesFn: func(_ context.Context, _, _ uint64, _ []ClinicalPlanDiagnosisInput) error {
				t.Fatal("must not replace")
				return nil
			},
		}
		svc := &medicalRecordService{
			inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo,
			diagTypeRepo: okDiagnosisTypeRepo(), diagNameRepo: okDiagnosisNameRepo(),
		}
		tooMany := make([]ClinicalPlanDiagnosisInput, maxClinicalPlanDiagnosisRows+1)
		for i := range tooMany {
			tooMany[i] = ClinicalPlanDiagnosisInput{Slot: 1, SortOrder: i}
		}

		err := svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Plan: strPtr("p"), Diagnoses: tooMany})

		require.Error(t, err)
	})

	t.Run("clinical plan update failure is logged and swallowed (best-effort)", func(t *testing.T) {
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
				return errors.New("update failed")
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		assert.NotPanics(t, func() {
			_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Plan: strPtr("policy")})
		})
	})

	// MRC-14 / EMR-495 A: CreateSubRecords も clinical plan と同じ type↔name 整合を全行に適用する。
	t.Run("CreateSubRecords diagnosis row type/name mismatch skips every plan write", func(t *testing.T) {
		writeCalled := false
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
				writeCalled = true
				return nil
			},
			replaceDiagnosesFn: func(_ context.Context, _, _ uint64, _ []ClinicalPlanDiagnosisInput) error {
				writeCalled = true
				return nil
			},
		}
		diagTypeRepo := &mockDiagnosisTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
				return &model.DiagnosisType{ID: id}, nil
			},
		}
		diagNameRepo := &mockDiagnosisNameRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
				// name belongs to type 99, not the requested type 3
				return &model.DiagnosisName{ID: id, DiagnosisTypeID: 99}, nil
			},
		}
		svc := &medicalRecordService{
			inquiryRepo:      &mockInquiryRepository{},
			clinicalPlanRepo: clinicalPlanRepo,
			diagTypeRepo:     diagTypeRepo,
			diagNameRepo:     diagNameRepo,
		}
		_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{
			Diagnoses: []ClinicalPlanDiagnosisInput{{Slot: 2, DiagnosisTypeID: uint64Ptr(3), DiagnosisNameID: uint64Ptr(4)}},
		})
		assert.False(t, writeCalled, "diagnosis type/name mismatch must skip the clinical plan writes")
	})
}

func TestHasInquirySubRecordInput(t *testing.T) {
	assert.False(t, hasInquirySubRecordInput(CreateSubRecordsInput{}))
	assert.True(t, hasInquirySubRecordInput(CreateSubRecordsInput{ChiefComplaintTypeID: uint64Ptr(1)}))
	assert.True(t, hasInquirySubRecordInput(CreateSubRecordsInput{ChiefComplaint: strPtr("x")}))
	assert.True(t, hasInquirySubRecordInput(CreateSubRecordsInput{Notes: strPtr("x")}))
}
