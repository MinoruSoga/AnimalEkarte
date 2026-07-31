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

// mockMedicalRecordRepository: ⑦で service 側 full 定義へ差し替え（旧 narrow 版はfield互換）。
type mockMedicalRecordRepository struct {
	findAllFn                         func(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error)
	findByIDFn                        func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	lockByIDForUpdateFn               func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	findByAppointmentIDFn             func(ctx context.Context, clinicID, appointmentID uint64) (*model.MedicalRecord, error)
	createFn                          func(ctx context.Context, record *model.MedicalRecord) error
	updateFieldsFn                    func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error)
	deleteFn                          func(ctx context.Context, clinicID, id uint64) error
	countByPetIDFn                    func(ctx context.Context, clinicID, petID uint64) (int64, error)
	findFirstVisitDateByPetIDFn       func(ctx context.Context, clinicID, petID uint64) (*time.Time, error)
	countEstimatesByMedicalRecordIDFn func(ctx context.Context, medicalRecordID uint64) (int64, error)
	findOwnerVisitSummaryFn           func(ctx context.Context, clinicID, ownerID uint64) (*OwnerVisitSummary, error)
	countByOwnerIDFn                  func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	countByOwnerIDCallCount           int
	// 以下4フィールドは F-3 統合で追加（旧 mockMedRecordRepoForDelivery / mockMedRecordRepoForLstepVisit
	// が個別に持っていたフック）。未設定時は旧共有モックと同じ nil,nil を返す（挙動不変）。
	findLatestByOwnerFn         func(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error)
	findOwnersByFirstVisitFn    func(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	findOwnersByLastVisitDaysFn func(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error)
	findOwnersByNextVisitRecFn  func(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	assertOwnerInClinicFn       func(ctx context.Context, clinicID, ownerID uint64) error
	findPetOwnerInClinicFn      func(ctx context.Context, clinicID, petID uint64) (uint64, error)
	assertDoctorInClinicFn      func(ctx context.Context, clinicID, doctorID uint64) error
	withTxFn                    func(ctx context.Context, fn func(context.Context) error) error
}

func (m *mockMedicalRecordRepository) FindAll(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
	if m.findAllFn == nil {
		return []model.MedicalRecord{}, 0, nil
	}
	return m.findAllFn(ctx, clinicIDs, filters, page, limit)
}

func (m *mockMedicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.MedicalRecord, error) {
	if m.findByAppointmentIDFn != nil {
		return m.findByAppointmentIDFn(ctx, clinicID, appointmentID)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepository) Create(ctx context.Context, record *model.MedicalRecord) error {
	return m.createFn(ctx, record)
}

func (m *mockMedicalRecordRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any, _ *int) (*model.MedicalRecord, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.MedicalRecord{}, nil
}

func (m *mockMedicalRecordRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordRepository) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	if m.countByPetIDFn != nil {
		return m.countByPetIDFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockMedicalRecordRepository) FindFirstVisitDateByPetID(ctx context.Context, clinicID, petID uint64) (*time.Time, error) {
	if m.findFirstVisitDateByPetIDFn != nil {
		return m.findFirstVisitDateByPetIDFn(ctx, clinicID, petID)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) CountEstimatesByMedicalRecordID(ctx context.Context, _, medicalRecordID uint64) (int64, error) {
	if m.countEstimatesByMedicalRecordIDFn != nil {
		return m.countEstimatesByMedicalRecordIDFn(ctx, medicalRecordID)
	}
	return 0, nil
}

func (m *mockMedicalRecordRepository) FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*OwnerVisitSummary, error) {
	if m.findOwnerVisitSummaryFn != nil {
		return m.findOwnerVisitSummaryFn(ctx, clinicID, ownerID)
	}
	return &OwnerVisitSummary{}, nil
}

func (m *mockMedicalRecordRepository) FindLatestByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error) {
	if m.findLatestByOwnerFn != nil {
		return m.findLatestByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindDormantOwnerEntries(_ context.Context, _ uint64, _ int) ([]DormantOwnerEntry, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindDormantOwnerEntriesCursor(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]DormantOwnerEntry, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnersByFirstVisitDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	if m.findOwnersByFirstVisitFn != nil {
		return m.findOwnersByFirstVisitFn(ctx, clinicID, targetDate)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnersByLastVisitDays(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error) {
	if m.findOwnersByLastVisitDaysFn != nil {
		return m.findOwnersByLastVisitDaysFn(ctx, clinicID, exactDays, asOf)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnersByNextVisitRecommended(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	if m.findOwnersByNextVisitRecFn != nil {
		return m.findOwnersByNextVisitRecFn(ctx, clinicID, targetDate)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) CountByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	m.countByOwnerIDCallCount++
	if m.countByOwnerIDFn != nil {
		return m.countByOwnerIDFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

func (m *mockMedicalRecordRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	if m.lockByIDForUpdateFn != nil {
		return m.lockByIDForUpdateFn(ctx, clinicID, id)
	}
	record, err := m.FindByID(ctx, clinicID, id)
	if err != nil || record == nil {
		return record, err
	}
	// Legacy service fixtures often specify only the fields relevant to the
	// individual test. The default locker models a correctly scoped repository;
	// tests for malformed/foreign locks use lockByIDForUpdateFn explicitly.
	scoped := *record
	if scoped.ID == 0 {
		scoped.ID = id
	}
	if scoped.ClinicID == 0 {
		scoped.ClinicID = clinicID
	}
	return &scoped, nil
}

func (m *mockMedicalRecordRepository) AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error {
	if m.assertOwnerInClinicFn != nil {
		return m.assertOwnerInClinicFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockMedicalRecordRepository) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	if m.findPetOwnerInClinicFn != nil {
		return m.findPetOwnerInClinicFn(ctx, clinicID, petID)
	}
	return 1, nil
}

func (m *mockMedicalRecordRepository) FindPetByIDInClinic(_ context.Context, _, petID uint64) (*model.Pet, error) {
	return &model.Pet{ID: petID, Status: model.PetStatusAlive}, nil
}


func (m *mockMedicalRecordRepository) AssertMedicalRecordDoctorInClinic(ctx context.Context, clinicID, doctorID uint64) error {
	if m.assertDoctorInClinicFn != nil {
		return m.assertDoctorInClinicFn(ctx, clinicID, doctorID)
	}
	return nil
}

func (m *mockMedicalRecordRepository) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx)
}

// mockLstepTagSyncService satisfies checkupTagSyncer, vaccinationTagSyncer and prescriptionTagSyncer
// (the union of their methods). Only the fields the migrated tests set are declared.
type mockLstepTagSyncService struct {
	syncVaccineTagFn          func(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error
	resyncOwnerVaccineTagsFn  func(ctx context.Context, clinicID, ownerID uint64) error
	syncCheckupTagFn          func(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error
	resyncOwnerCheckupTagsFn  func(ctx context.Context, clinicID, ownerID uint64) error
	syncPrescriptionTagFn     func(ctx context.Context, clinicID, ownerID uint64) error
	syncNextVisitTagFn        func(ctx context.Context, clinicID, ownerID uint64) error
	syncVisitCompletionTagsFn func(ctx context.Context, clinicID, ownerID uint64) error
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

// ⑦統合: medical_record 系テストが使う追加メソッド（no-op）。
func (m *mockLstepTagSyncService) SyncVisitCompletionTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncVisitCompletionTagsFn != nil {
		return m.syncVisitCompletionTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepTagSyncService) SyncOwnerAnimalClassificationTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncPetBasicInfoTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncCPMStageTag(_ context.Context, _, _ uint64) error { return nil }
func (m *mockLstepTagSyncService) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncNextVisitTagFn != nil {
		return m.syncNextVisitTagFn(ctx, clinicID, ownerID)
	}
	return nil
}

// mockAuditService は mrAuditLogger（LogMedicalRecordChange/LogAddendumCreate）の test double
// （⑦移動テスト用・internal/service の同名 mock の該当 subset をfield名互換で再宣言）。
type mockAuditService struct {
	calls                    []string
	logMedicalRecordChangeFn func(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error
	logAddendumCreateFn      func(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error
}

func (m *mockAuditService) LogMedicalRecordChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error {
	m.calls = append(m.calls, action)
	if m.logMedicalRecordChangeFn != nil {
		return m.logMedicalRecordChangeFn(ctx, clinicID, actorID, action, recordID, oldValue, newValue)
	}
	return nil
}

func (m *mockAuditService) LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error {
	m.calls = append(m.calls, "create")
	if m.logAddendumCreateFn != nil {
		return m.logAddendumCreateFn(ctx, clinicID, actorID, addendumID, medicalRecordID, addendum)
	}
	return nil
}
