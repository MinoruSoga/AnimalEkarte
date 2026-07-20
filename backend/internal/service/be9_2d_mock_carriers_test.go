package service

// BE9-2D mock carriers. These six test doubles were defined in the service test files that moved
// to internal/medicalrecord (checkup / checkup_field_result / vaccination / prescription / inquiry
// service tests). Residual internal/service tests still depend on them — lstep tag-sync tests use
// mockCheckupRepository/mockVaccinationRepository/mockPrescriptionRepository; examination /
// medical_record_image / vital / cross_tenant tests use mockCheckupTransactor/mockAuditTxLogger;
// medical_record_subrecords + the residual MedicalRecordService cross-tenant section use
// mockInquiryRepository — so the definitions are retained here verbatim (same fields, methods and
// *AuditLogInput signature) as a shared carrier, following the sub-batch① "残置" precedent.

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- carrier: mockCheckupRepository (from checkup_service_test.go) ----

type mockCheckupRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	listByClinicFn          func(ctx context.Context, clinicID uint64, filters repository.CheckupFilters, page, limit int) ([]model.Checkup, int64, error)
	findByOwnerIDFn         func(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error)
	findByIDFn              func(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error)
	createFn                func(ctx context.Context, checkup *model.Checkup) error
	updateFn                func(ctx context.Context, clinicID, checkupID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, checkupID uint64) error
}

func (m *mockCheckupRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockCheckupRepository) FindByClinicID(ctx context.Context, clinicID uint64, filters repository.CheckupFilters, page, limit int) ([]model.Checkup, int64, error) {
	if m.listByClinicFn != nil {
		return m.listByClinicFn(ctx, clinicID, filters, page, limit)
	}
	return nil, 0, nil
}

func (m *mockCheckupRepository) FindByID(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, checkupID)
	}
	return nil, nil
}

func (m *mockCheckupRepository) FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error) {
	if m.findByOwnerIDFn != nil {
		return m.findByOwnerIDFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockCheckupRepository) Create(ctx context.Context, checkup *model.Checkup) error {
	return m.createFn(ctx, checkup)
}

func (m *mockCheckupRepository) Update(ctx context.Context, clinicID, checkupID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, checkupID, fields)
}

func (m *mockCheckupRepository) Delete(ctx context.Context, clinicID, checkupID uint64) error {
	return m.deleteFn(ctx, clinicID, checkupID)
}

// ---- carrier: mockAuditTxLogger + mockCheckupTransactor (from checkup_field_result_service_test.go) ----

type mockAuditTxLogger struct {
	logEntryTxFn func(ctx context.Context, input *AuditLogInput) error
}

func (m *mockAuditTxLogger) LogEntryTx(ctx context.Context, input *AuditLogInput) error {
	if m.logEntryTxFn != nil {
		return m.logEntryTxFn(ctx, input)
	}
	return nil
}

type mockCheckupTransactor struct {
	repository.Transactor
}

func (m *mockCheckupTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// ---- carrier: mockVaccinationRepository (from vaccination_service_test.go) ----

type mockVaccinationRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	findByOwnerFn  func(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
	createFn       func(ctx context.Context, vaccination *model.Vaccination) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockVaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockVaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockVaccinationRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockVaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	return m.createFn(ctx, vaccination)
}

func (m *mockVaccinationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockVaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockVaccinationRepository) FindOwnersByVaccineDeadline(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}

// ---- carrier: mockPrescriptionRepository (from prescription_service_test.go) ----

type mockPrescriptionRepository struct {
	findByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error)
	findByIDFn              func(ctx context.Context, clinicID, id uint64) (*model.Prescription, error)
	findActiveByOwnerFn     func(ctx context.Context, clinicID, ownerID uint64) ([]model.Prescription, error)
	createFn                func(ctx context.Context, prescription *model.Prescription) error
	updateFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockPrescriptionRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
	if m.findByMedicalRecordIDFn != nil {
		return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
	}
	return []model.Prescription{}, nil
}

func (m *mockPrescriptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Prescription, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Prescription{}, nil
}

func (m *mockPrescriptionRepository) FindActiveByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Prescription, error) {
	if m.findActiveByOwnerFn != nil {
		return m.findActiveByOwnerFn(ctx, clinicID, ownerID)
	}
	return []model.Prescription{}, nil
}

func (m *mockPrescriptionRepository) Create(ctx context.Context, prescription *model.Prescription) error {
	if m.createFn != nil {
		return m.createFn(ctx, prescription)
	}
	return nil
}

func (m *mockPrescriptionRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return nil
}

func (m *mockPrescriptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

// ---- carrier: mockInquiryRepository (from inquiry_service_test.go) ----

type mockInquiryRepository struct {
	upsertFn func(ctx context.Context, clinicID uint64, inquiry *model.Inquiry) (*model.Inquiry, error)
}

func (m *mockInquiryRepository) SaveByMedicalRecordID(ctx context.Context, clinicID uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
	return m.upsertFn(ctx, clinicID, inquiry)
}

func (m *mockInquiryRepository) CountByChiefComplaintTypeID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

// ---- carrier: mockVitalRepository (from vital_service_test.go) ----
// vitalService moved to internal/medicalrecord in BE9-2D sub-batch④a, but treatment_dose_save_test.go
// (treatment stays in internal/service until sub-batch④b) still constructs this double, so its
// definition is retained here verbatim.

type mockVitalRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error)
	findByIDFn              func(ctx context.Context, clinicID, vitalID uint64) (*model.VitalRecord, error)
	createFn                func(ctx context.Context, vital *model.VitalRecord) error
	updateFn                func(ctx context.Context, clinicID, vitalID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, vitalID uint64) error
}

func (m *mockVitalRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockVitalRepository) FindByID(ctx context.Context, clinicID, vitalID uint64) (*model.VitalRecord, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, vitalID)
	}
	return nil, nil
}

func (m *mockVitalRepository) Create(ctx context.Context, vital *model.VitalRecord) error {
	return m.createFn(ctx, vital)
}

func (m *mockVitalRepository) Update(ctx context.Context, clinicID, vitalID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, vitalID, fields)
}

func (m *mockVitalRepository) Delete(ctx context.Context, clinicID, vitalID uint64) error {
	return m.deleteFn(ctx, clinicID, vitalID)
}

// ---- carrier: mockClinicalPlanRepository (from clinical_plan_service_test.go) ----
// clinicalPlanService moved to internal/medicalrecord in BE9-2D sub-batch④a, but the residual
// cross_tenant_master_fk_write_test.go / medical_record_auto_create_test.go /
// medical_record_subrecords_test.go still construct this double, so its definition is retained
// here verbatim. updateFn keeps the 4-arg shape (BUG-416③ compat); expectedVersion-checking tests
// set updateWithVersionFn.

type mockClinicalPlanRepository struct {
	findByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	createFn                func(ctx context.Context, plan *model.ClinicalPlan) error
	updateFn                func(ctx context.Context, clinicID, planID uint64, fields map[string]any) error
	updateWithVersionFn     func(ctx context.Context, clinicID, planID uint64, fields map[string]any, expectedVersion *int) error
	deleteFn                func(ctx context.Context, clinicID, planID uint64) error
}

func (m *mockClinicalPlanRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockClinicalPlanRepository) Create(ctx context.Context, plan *model.ClinicalPlan) error {
	return m.createFn(ctx, plan)
}

func (m *mockClinicalPlanRepository) Update(ctx context.Context, clinicID, planID uint64, fields map[string]any, expectedVersion *int) error {
	if m.updateWithVersionFn != nil {
		return m.updateWithVersionFn(ctx, clinicID, planID, fields, expectedVersion)
	}
	return m.updateFn(ctx, clinicID, planID, fields)
}

func (m *mockClinicalPlanRepository) Delete(ctx context.Context, clinicID, planID uint64) error {
	return m.deleteFn(ctx, clinicID, planID)
}

// ── BE9-2D ⑤ carrier ──
// mockHospitalizationRepository は旧 hospitalization_service_test.go の同名 mock の残置コピー
// （accounting_fk_clinic_isolation_test / accounting_service_test が使用・解消=accounting/billing domain 移行時）。
type mockHospitalizationRepository struct {
	findAllFn                                func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	findByIDFn                               func(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	createFn                                 func(ctx context.Context, hospitalization *model.Hospitalization) error
	updateFieldsFn                           func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error)
	updateIfNotDischargedFn                  func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error)
	deleteFn                                 func(ctx context.Context, clinicID, id uint64) error
	countCarePlanItemsByHospitalizationIDFn  func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	countDailyRecordsByHospitalizationIDFn   func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	countTreatmentPlansByHospitalizationIDFn func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
}

func (m *mockHospitalizationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockHospitalizationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockHospitalizationRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockHospitalizationRepository) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	return m.createFn(ctx, hospitalization)
}

func (m *mockHospitalizationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockHospitalizationRepository) UpdateIfNotDischarged(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error) {
	if m.updateIfNotDischargedFn != nil {
		return m.updateIfNotDischargedFn(ctx, clinicID, id, fields)
	}
	return nil, apperrors.WrapNotFound("hospitalization", "updateIfNotDischargedFn not stubbed")
}

func (m *mockHospitalizationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockHospitalizationRepository) CountByCageID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockHospitalizationRepository) CountCarePlanItemsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	if m.countCarePlanItemsByHospitalizationIDFn == nil {
		return 0, nil
	}
	return m.countCarePlanItemsByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockHospitalizationRepository) CountDailyRecordsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	if m.countDailyRecordsByHospitalizationIDFn == nil {
		return 0, nil
	}
	return m.countDailyRecordsByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockHospitalizationRepository) CountTreatmentPlansByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	if m.countTreatmentPlansByHospitalizationIDFn == nil {
		return 0, nil
	}
	return m.countTreatmentPlansByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

// ── BE9-2D ⑥ carriers ──
// medicine/procedure/consultation の service test 移動後も、残留する cross_tenant builder
// （ok/reject 系: estimate/billing_item/examination 等が使用）と clinic_service_test（strPtr）が
// 使う test double の残置コピー。解消 = 各残留 domain の移行時。

func strPtr(s string) *string { return &s }

type mockMedicineRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	createFn                  func(ctx context.Context, medicine *model.Medicine) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockMedicineRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockMedicineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockMedicineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn != nil {
		return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
	}
	return 0, nil
}

func (m *mockMedicineRepository) CountUsageByMedicineID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockMedicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {
	return m.createFn(ctx, medicine)
}

func (m *mockMedicineRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockMedicineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMedicineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

type mockProcedureRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.Procedure, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
	createFn                  func(ctx context.Context, procedure *model.Procedure) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	countUsageByProcedureIDFn func(ctx context.Context, clinicID, procedureID uint64) (int64, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockProcedureRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockProcedureRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockProcedureRepository) Create(ctx context.Context, procedure *model.Procedure) error {
	return m.createFn(ctx, procedure)
}

func (m *mockProcedureRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Procedure{ID: id}, nil
}

func (m *mockProcedureRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockProcedureRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockProcedureRepository) CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error) {
	if m.countUsageByProcedureIDFn == nil {
		return 0, nil
	}
	return m.countUsageByProcedureIDFn(ctx, clinicID, procedureID)
}

func (m *mockProcedureRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

type mockConsultationRepository struct {
	findAllFn                    func(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	findByIDFn                   func(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	createFn                     func(ctx context.Context, consultation *model.Consultation) error
	updateFieldsFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error)
	deleteFn                     func(ctx context.Context, clinicID, id uint64) error
	countUsageByConsultationIDFn func(ctx context.Context, clinicID, consultationID uint64) (int64, error)
	countChildrenByParentIDFn    func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	reorderErr                   error
}

func (m *mockConsultationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockConsultationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockConsultationRepository) Create(ctx context.Context, consultation *model.Consultation) error {
	return m.createFn(ctx, consultation)
}

func (m *mockConsultationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockConsultationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockConsultationRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockConsultationRepository) CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error) {
	if m.countUsageByConsultationIDFn == nil {
		return 0, nil
	}
	return m.countUsageByConsultationIDFn(ctx, clinicID, consultationID)
}

func (m *mockConsultationRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

// uint64Ptr は ⑥ 移動ファイル由来 helper の残置コピー（medical_record/owner 系残留テストが使用）。
func uint64Ptr(v uint64) *uint64 { return &v }
