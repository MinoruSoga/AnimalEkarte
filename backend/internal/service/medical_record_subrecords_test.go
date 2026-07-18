package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCreateSubRecords(t *testing.T) {
	t.Run("no inquiry input: SaveByMedicalRecordID is not called", func(t *testing.T) {
		saveCalled := false
		inquiryRepo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, _ *model.Inquiry) (*model.Inquiry, error) {
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

		svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{})

		assert.False(t, saveCalled, "inquiry upsert should be skipped when no inquiry fields are provided")
	})

	t.Run("inquiry input present: SaveByMedicalRecordID is called with mapped fields", func(t *testing.T) {
		var savedInquiry *model.Inquiry
		typeID := uint64(3)
		complaint := "食欲不振"
		notes := "2日前から"
		inquiryRepo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
				savedInquiry = inquiry
				return inquiry, nil
			},
		}
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1}, nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: inquiryRepo, clinicalPlanRepo: clinicalPlanRepo, chiefComplaintTypeRepo: okChiefComplaintTypeRepo()}

		svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{
			ChiefComplaintTypeID: &typeID,
			ChiefComplaint:       &complaint,
			Notes:                &notes,
		})

		if assert.NotNil(t, savedInquiry) {
			assert.Equal(t, uint64(10), savedInquiry.MedicalRecordID)
			assert.Equal(t, &typeID, savedInquiry.ChiefComplaintTypeID)
			assert.Equal(t, complaint, savedInquiry.ChiefComplaint)
			assert.Equal(t, notes, savedInquiry.Notes)
		}
	})

	t.Run("inquiry upsert failure is logged and swallowed (best-effort)", func(t *testing.T) {
		complaint := "嘔吐"
		inquiryRepo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, _ *model.Inquiry) (*model.Inquiry, error) {
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
			svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{ChiefComplaint: &complaint})
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
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				updateCalled = true
				return nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		plan := "policy"
		svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Plan: &plan})

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

		svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{})

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
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				updateCalled = true
				return nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		plan := "policy"
		svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Plan: &plan})

		assert.False(t, updateCalled)
	})

	t.Run("no plan-related fields: Update is not called", func(t *testing.T) {
		updateCalled := false
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				updateCalled = true
				return nil
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{})

		assert.False(t, updateCalled)
	})

	t.Run("plan-related fields present: Update is called with mapped fields", func(t *testing.T) {
		var updatedFields map[string]any
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
				updatedFields = fields
				return nil
			},
		}
		svc := &medicalRecordService{
			inquiryRepo:      &mockInquiryRepository{},
			clinicalPlanRepo: clinicalPlanRepo,
			diagTypeRepo:     okDiagnosisTypeRepo(),
			diagNameRepo:     okDiagnosisNameRepo(),
		}

		svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{
			Plan:                 strPtr("policy"),
			Assessment:           strPtr("assessment"),
			Diagnosis1CategoryID: uint64Ptr(1),
			Diagnosis1NameID:     uint64Ptr(2),
			Diagnosis2TypeID:     uint64Ptr(3),
			Diagnosis2NameID:     uint64Ptr(4),
		})

		if assert.NotNil(t, updatedFields) {
			assert.Equal(t, "policy", updatedFields["treatment_policy"])
			assert.Equal(t, "assessment", updatedFields["diagnosis_details"])
			assert.Equal(t, uint64(1), updatedFields["diagnosis_type_id"])
			assert.Equal(t, uint64(2), updatedFields["diagnosis_name_id"])
			assert.Equal(t, uint64(3), updatedFields["diagnosis_2_type_id"])
			assert.Equal(t, uint64(4), updatedFields["diagnosis_2_name_id"])
		}
	})

	t.Run("clinical plan update failure is logged and swallowed (best-effort)", func(t *testing.T) {
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return errors.New("update failed")
			},
		}
		svc := &medicalRecordService{inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: clinicalPlanRepo}

		assert.NotPanics(t, func() {
			svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Plan: strPtr("policy")})
		})
	})
}

func TestHasInquirySubRecordInput(t *testing.T) {
	assert.False(t, hasInquirySubRecordInput(CreateSubRecordsInput{}))
	assert.True(t, hasInquirySubRecordInput(CreateSubRecordsInput{ChiefComplaintTypeID: uint64Ptr(1)}))
	assert.True(t, hasInquirySubRecordInput(CreateSubRecordsInput{ChiefComplaint: strPtr("x")}))
	assert.True(t, hasInquirySubRecordInput(CreateSubRecordsInput{Notes: strPtr("x")}))
}
