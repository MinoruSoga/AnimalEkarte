package billing

// service_mocks_test.go — def残存（inventory系mock）→移動先で再宣言する規約の複製（B①）。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockMerchandiseItemRepository は MerchandiseItemRepository のテスト用モック実装
type mockMerchandiseItemRepository struct {
	findAllFn                     func(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	countUsageByMerchandiseItemFn func(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error)
	createFn                      func(ctx context.Context, item *model.MerchandiseItem) error
	updateFieldsFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error)
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	reorderFn                     func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockMerchandiseItemRepository) FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, category)
	}
	return nil, nil
}

func (m *mockMerchandiseItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.MerchandiseItem{ID: id, ClinicID: clinicID}, nil
}

func (m *mockMerchandiseItemRepository) CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error) {
	if m.countUsageByMerchandiseItemFn != nil {
		return m.countUsageByMerchandiseItemFn(ctx, clinicID, merchandiseItemID)
	}
	return 0, nil
}

func (m *mockMerchandiseItemRepository) Create(ctx context.Context, item *model.MerchandiseItem) error {
	return m.createFn(ctx, item)
}

func (m *mockMerchandiseItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockMerchandiseItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMerchandiseItemRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// mockMedicalRecordRepository — billingMedicalRecordLocker（sharedkernel.MedicalRecordLocker面）の
// 最小モック（旧 service 側 full mock の view 版・def残存→再宣言規約）。
type mockMedicalRecordRepository struct {
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	lockByIDForUpdateFn func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

func (m *mockMedicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
}

func (m *mockMedicalRecordRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	if m.lockByIDForUpdateFn != nil {
		return m.lockByIDForUpdateFn(ctx, clinicID, id)
	}
	// 旧 full mock と同じく、findByIDFn が設定されていれば同一の行を返す（Lock 系ガードの検証用）
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
}

// mockAuditService — billingAuditLogger view の最小モック。
type mockAuditService struct {
	logEntryErr    error
	logEntryCalled bool
	logEntryFn     func(ctx context.Context, entry *AuditEntry) error
	entries        []*AuditEntry
	lastLogEntry   *AuditEntry
}

func (m *mockAuditService) LogEntry(ctx context.Context, entry *AuditEntry) error {
	m.entries = append(m.entries, entry)
	m.lastLogEntry = entry
	m.logEntryCalled = true
	if m.logEntryFn != nil {
		return m.logEntryFn(ctx, entry)
	}
	return m.logEntryErr
}

// noopTransactor — service 側同名テストヘルパの複製（WithTx を素通しする）。
type noopTransactor struct{}

func (noopTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// mockReservationRepository — reservation側test定義（package跨ぎimport不能）の複製（AUD-005 estimate fk test用）。
type mockReservationRepository struct {
	findAllFn                          func(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	findByIDFn                         func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	lockAndFindByIDFn                  func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	createFn                           func(ctx context.Context, reservation *model.Reservation) error
	updateFieldsFn                     func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
	deleteFn                           func(ctx context.Context, clinicID, id uint64) error
	countMedicalRecordsByReservationID func(ctx context.Context, reservationID uint64) (int64, error)
	countOnDutyDoctorsFn               func(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	countConflictsFn                   func(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
	countByTypeAndStartTimeFn          func(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
	assertOwnerInClinicFn              func(ctx context.Context, clinicID, ownerID uint64) error
	findPetOwnerInClinicFn             func(ctx context.Context, clinicID, petID uint64) (uint64, error)
	assertLineCustomerInClinicFn       func(ctx context.Context, clinicID, lineCustomerID uint64) error
}

func (m *mockReservationRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	return m.findAllFn(ctx, clinicIDs, page, limit, date, startDate, endDate, status, source, petID, ownerID)
}

func (m *mockReservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Reservation, error) {
	return nil, nil
}

func (m *mockReservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	return m.createFn(ctx, reservation)
}

func (m *mockReservationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockReservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationRepository) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, _, reservationID uint64) (int64, error) {
	if m.countMedicalRecordsByReservationID != nil {
		return m.countMedicalRecordsByReservationID(ctx, reservationID)
	}
	return 0, nil
}

func (m *mockReservationRepository) AcquireBookingLock(_ context.Context, _ uint64) error {
	return nil
}

func (m *mockReservationRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.lockAndFindByIDFn != nil {
		return m.lockAndFindByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationRepository) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error) {
	if m.countOnDutyDoctorsFn != nil {
		return m.countOnDutyDoctorsFn(ctx, clinicID, date)
	}
	return 1, nil
}

func (m *mockReservationRepository) CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error) {
	if m.countConflictsFn != nil {
		return m.countConflictsFn(ctx, clinicID, start, end, excludeID)
	}
	return 0, nil
}

func (m *mockReservationRepository) CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
	if m.countByTypeAndStartTimeFn != nil {
		return m.countByTypeAndStartTimeFn(ctx, clinicID, reservationTypeID, startTime, excludeID)
	}
	return 0, nil
}

func (m *mockReservationRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockReservationRepository) AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error {
	if m.assertOwnerInClinicFn != nil {
		return m.assertOwnerInClinicFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockReservationRepository) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	if m.findPetOwnerInClinicFn != nil {
		return m.findPetOwnerInClinicFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockReservationRepository) AssertLineCustomerInClinic(ctx context.Context, clinicID, lineCustomerID uint64) error {
	if m.assertLineCustomerInClinicFn != nil {
		return m.assertLineCustomerInClinicFn(ctx, clinicID, lineCustomerID)
	}
	return nil
}

func (m *mockReservationRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}

// mockAccountingRepository — accountingBillingView（FindByID/LockAndFindByID）の最小view mock
// （def残存=accounting系はB④・再宣言規約）。
type mockAccountingRepository struct {
	findByIDFn        func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	lockAndFindByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
}

func (m *mockAccountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Billing{ID: id, ClinicID: clinicID}, nil
}

func (m *mockAccountingRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	if m.lockAndFindByIDFn != nil {
		return m.lockAndFindByIDFn(ctx, clinicID, id)
	}
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Billing{ID: id, ClinicID: clinicID}, nil
}

// mockTransactor / okTrimming* — service/reservation 側同名テストヘルパの複製（def残存→再宣言規約）。
type mockTransactor struct {
	withTxErr error
	withTxFn  func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	if m.withTxErr != nil {
		return m.withTxErr
	}
	return fn(ctx)
}

type mockTrimmingCourseFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
}

func (m *mockTrimmingCourseFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingCourseFinder) FindAll(_ context.Context, _ uint64) ([]model.TrimmingCourse, error) {
	return nil, nil
}

type mockTrimmingOptionFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
}

func (m *mockTrimmingOptionFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingOptionFinder) FindAll(_ context.Context, _ uint64) ([]model.TrimmingOption, error) {
	return nil, nil
}

func okTrimmingCourseRepo() trimmingCourseFinder {
	return &mockTrimmingCourseFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		return &model.TrimmingCourse{ID: id, IsActive: true}, nil
	}}
}

func okTrimmingOptionRepo() trimmingOptionFinder {
	return &mockTrimmingOptionFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		return &model.TrimmingOption{ID: id, IsActive: true}, nil
	}}
}

// mockOwnerRepository — billingOwnerReader（FindByID のみ）の最小view mock（#81 段階2b）。
type mockOwnerRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
}

func (m *mockOwnerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Owner{ID: id, ClinicID: clinicID}, nil
}

// reject系builder — service側同名のview型版複製。
func rejectTrimmingCourseRepo(ownedID uint64) trimmingCourseFinder {
	return &mockTrimmingCourseFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_course", "foreign")
		}
		return &model.TrimmingCourse{ID: id, IsActive: true}, nil
	}}
}

func rejectTrimmingOptionRepo(ownedID uint64) trimmingOptionFinder {
	return &mockTrimmingOptionFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_option", "foreign")
		}
		return &model.TrimmingOption{ID: id, IsActive: true}, nil
	}}
}
