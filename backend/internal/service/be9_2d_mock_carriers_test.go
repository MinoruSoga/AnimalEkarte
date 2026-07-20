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
	"errors"
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

// ── BE9-2D sub-batch④b carriers ──
// dose kernel（dose_calc/dose_revalidation/dose_validators/treatment_dose_save）の medicalrecord
// 移動後も、残留する medicine_dose_config_test.go / medicine_dose_param_service_test.go が使う
// helper/mock/センチネルの残置コピー。carrier 解消は medicine domain の BE9-2C/2D 時。

// fptr は旧 dose_calc_test.go の同名 helper の残置コピー。
func fptr(v float64) *float64 { return &v }

// errAuditWriteFailed は旧 treatment_dose_save_test.go の同名センチネルの残置コピー
// （mockAuditService.logEntryTxErr にセットして使う）。
var errAuditWriteFailed = errors.New("audit write failed")

// mockMedicineDoseParamRepository は旧 treatment_dose_save_test.go の同名 mock の残置コピー
// （repository.MedicineDoseParamRepository 全メソッド実装）。
type mockMedicineDoseParamRepository struct {
	findByMedicineIDFn         func(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error)
	findByMedicineAndSpeciesFn func(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies) (*model.MedicineDoseParam, error)
	createFn                   func(ctx context.Context, clinicID uint64, param *model.MedicineDoseParam) error
	updateFn                   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicineDoseParam, error)
	deleteFn                   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockMedicineDoseParamRepository) FindByMedicineID(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error) {
	if m.findByMedicineIDFn == nil {
		return nil, nil
	}
	return m.findByMedicineIDFn(ctx, clinicID, medicineID)
}
func (m *mockMedicineDoseParamRepository) FindByMedicineAndSpecies(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
	if m.findByMedicineAndSpeciesFn == nil {
		return nil, apperrors.WrapNotFound("medicine_dose_param", "")
	}
	return m.findByMedicineAndSpeciesFn(ctx, clinicID, medicineID, species)
}
func (m *mockMedicineDoseParamRepository) Create(ctx context.Context, clinicID uint64, param *model.MedicineDoseParam) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, clinicID, param)
}
func (m *mockMedicineDoseParamRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicineDoseParam, error) {
	if m.updateFn == nil {
		return nil, nil
	}
	return m.updateFn(ctx, clinicID, id, fields)
}
func (m *mockMedicineDoseParamRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, clinicID, id)
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
