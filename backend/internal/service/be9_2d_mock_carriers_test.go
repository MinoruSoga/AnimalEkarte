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
