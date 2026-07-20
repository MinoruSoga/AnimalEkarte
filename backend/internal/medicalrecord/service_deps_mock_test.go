package medicalrecord

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// Test doubles for the consumer-side dependency views declared in service_deps.go. These are
// local, minimal copies of the shared mocks the moved service tests used while they lived in
// internal/service (mocks_medical_record_test.go's mockMedicalRecordRepository and
// lstep_lifecycle_service_test.go's mockLstepTagSyncService) — internal/medicalrecord cannot
// import those _test.go definitions, so BE9-2D re-declares the subset the migrated tests need
// against the narrow interfaces (medicalRecordFinder/medicalRecordLocker and the *TagSyncer views).

// mockMedicalRecordRepository satisfies medicalRecordFinder (FindByID) and medicalRecordLocker
// (LockByIDForUpdate delegates to FindByID, mirroring the residual mock so X-11 finalize-lock
// tests keep passing findByIDFn only).
type mockMedicalRecordRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

func (m *mockMedicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.FindByID(ctx, clinicID, id)
}

// mockLstepTagSyncService satisfies checkupTagSyncer, vaccinationTagSyncer and prescriptionTagSyncer
// (the union of their methods). Only the fields the migrated tests set are declared.
type mockLstepTagSyncService struct {
	syncVaccineTagFn         func(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error
	resyncOwnerVaccineTagsFn func(ctx context.Context, clinicID, ownerID uint64) error
	syncCheckupTagFn         func(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error
	resyncOwnerCheckupTagsFn func(ctx context.Context, clinicID, ownerID uint64) error
	syncPrescriptionTagFn    func(ctx context.Context, clinicID, ownerID uint64) error
}

func (m *mockLstepTagSyncService) SyncVaccineTag(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error {
	if m.syncVaccineTagFn != nil {
		return m.syncVaccineTagFn(ctx, clinicID, ownerID, vaccinationID)
	}
	return nil
}

func (m *mockLstepTagSyncService) ResyncOwnerVaccineTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.resyncOwnerVaccineTagsFn != nil {
		return m.resyncOwnerVaccineTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagSyncService) SyncCheckupTag(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error {
	if m.syncCheckupTagFn != nil {
		return m.syncCheckupTagFn(ctx, clinicID, ownerID, checkupTypeID, checkupDate, nextDate)
	}
	return nil
}

func (m *mockLstepTagSyncService) ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.resyncOwnerCheckupTagsFn != nil {
		return m.resyncOwnerCheckupTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagSyncService) SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncPrescriptionTagFn != nil {
		return m.syncPrescriptionTagFn(ctx, clinicID, ownerID)
	}
	return nil
}

// fptr mirrors internal/service's dose_calc_test.go fptr helper (a *float64 literal builder),
// re-declared here for the migrated checkup_field_result_service_test.go.
func fptr(v float64) *float64 { return &v }

// ptrUint64 mirrors internal/service's pet_service_test.go ptrUint64 helper (a *uint64 literal
// builder), re-declared here for the migrated vaccination_service_test.go.
func ptrUint64(v uint64) *uint64 { return &v }
