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
	reorderFn                     func(ctx context.Context, clinicID uint64, ids []uint64) error
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
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
	logEntryErr      error
	logEntryCalled   bool
	logEntryTxInput  *AuditEntry
	logEntryTxCalled bool
	logEntryTxErr    error
	logEntryFn       func(ctx context.Context, entry *AuditEntry) error
	entries          []*AuditEntry
	lastLogEntry     *AuditEntry
}

// LogEntryTx は billingAuditTxLogger 面（fail-closed 経路のテスト共用）。
func (m *mockAuditService) LogEntryTx(ctx context.Context, entry *AuditEntry) error {
	m.logEntryTxInput = entry
	m.logEntryTxCalled = true
	if m.logEntryTxErr != nil {
		return m.logEntryTxErr
	}
	return m.LogEntry(ctx, entry)
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

// mockReservationRepository は billing が消費する owner/pet verifier と会計完了intentだけを実装する。
type mockReservationRepository struct {
	assertOwnerInClinicFn   func(ctx context.Context, clinicID, ownerID uint64) error
	findPetOwnerInClinicFn  func(ctx context.Context, clinicID, petID uint64) (uint64, error)
	completeForAccountingFn func(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error)
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

func (m *mockReservationRepository) FindPetByIDInClinic(_ context.Context, _, petID uint64) (*model.Pet, error) {
	return &model.Pet{ID: petID, Status: model.PetStatusAlive}, nil
}

func (m *mockReservationRepository) CompleteForAccounting(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error) {
	if m.completeForAccountingFn != nil {
		return m.completeForAccountingFn(ctx, clinicID, medicalRecordID, ownerID, petID, scheduledDate)
	}
	return 0, nil
}

type mockAccountingRepository struct {
	findAllFn           func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, search string, page, limit int) ([]model.Billing, int64, error)
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	createFn            func(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	updateFieldsFn      func(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	savePaymentFn       func(ctx context.Context, payment *model.Payment) error
	savePaymentSplitsFn func(ctx context.Context, splits []model.PaymentSplit) error
	getDailySummaryFn   func(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error)
	// #120: start_date/end_date 2引数バリアント
	findUnpaidByBillingFn func(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error)
	findUnpaidByOwnerFn   func(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error)
	// #182: 飼主未納残高
	sumUnpaidByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error)
	// #114: 月次未納繰越集計
	findMonthlyUnpaidCarryoverFn func(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error)
	// 以下4フィールドは F-4 統合で追加（旧 ForReport/ForClose/ForLstepVisit が個別に持っていたフック）。
	// 未設定時は各旧モックのデフォルトと同じ値を返す（挙動不変）。
	getCloseAggregateFn                func(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error)
	getMonthlyReportFn                 func(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResult, error)
	getMonthlyReportByPeriodFn         func(ctx context.Context, clinicID uint64, start, end time.Time) (*MonthlyReportResult, error)
	getCategoryPaymentAllocationDataFn func(ctx context.Context, clinicID uint64, periodStart, periodEnd time.Time) (*CategoryPaymentAllocationData, error)
	sumPaidByOwnerFn                   func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// BUG-018: completion idempotency lookup
	findByCompletionRequestIDFn func(ctx context.Context, clinicID uint64, requestID string) (*model.Billing, error)
}

func (m *mockAccountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, search string, page, limit int) ([]model.Billing, int64, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, search, page, limit)
	}
	return nil, 0, nil
}

func (m *mockAccountingRepository) FindAllForClinics(_ context.Context, _ []uint64, _, _ *uint64, _, _, _ *string, _ string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}

func (m *mockAccountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockAccountingRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Billing, error) {
	return nil, nil
}

func (m *mockAccountingRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockAccountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, accounting)
	}
	return nil
}

func (m *mockAccountingRepository) Update(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, billingID, fields)
	}
	return nil, nil
}

func (m *mockAccountingRepository) SavePayment(ctx context.Context, payment *model.Payment) error {
	if m.savePaymentFn != nil {
		return m.savePaymentFn(ctx, payment)
	}
	return nil
}

func (m *mockAccountingRepository) SavePaymentSplits(ctx context.Context, splits []model.PaymentSplit) error {
	if m.savePaymentSplitsFn != nil {
		return m.savePaymentSplitsFn(ctx, splits)
	}
	return nil
}

func (m *mockAccountingRepository) FindUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error) {
	if m.findUnpaidByBillingFn != nil {
		return m.findUnpaidByBillingFn(ctx, clinicID, startDate, endDate, page, limit)
	}
	return nil, 0, nil
}

func (m *mockAccountingRepository) FindUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
	if m.findUnpaidByOwnerFn != nil {
		return m.findUnpaidByOwnerFn(ctx, clinicID, startDate, endDate, page, limit)
	}
	return nil, 0, UnpaidSummary{}, nil
}

func (m *mockAccountingRepository) SumUnpaidByOwner(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error) {
	if m.sumUnpaidByOwnerFn != nil {
		return m.sumUnpaidByOwnerFn(ctx, clinicID, ownerID)
	}
	return OwnerUnpaidBalance{}, nil
}

func (m *mockAccountingRepository) GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error) {
	if m.getDailySummaryFn != nil {
		return m.getDailySummaryFn(ctx, clinicID, date)
	}
	return &DailySummaryResult{PaymentTotals: []PaymentMethodTotal{}, CategoryTotals: []CategoryTotal{}}, nil
}

func (m *mockAccountingRepository) GetCloseAggregate(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error) {
	if m.getCloseAggregateFn != nil {
		return m.getCloseAggregateFn(ctx, input)
	}
	return &CloseAggregateResult{
		PaymentRows:    []PaymentAggregateRow{},
		CategoryRows:   []CategoryAggregateRow{},
		BillingDetails: []CloseBillingDetailRow{},
		TaxBreakdown:   []TaxBreakdownRow{},
	}, nil
}

func (m *mockAccountingRepository) GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResult, error) {
	if m.getMonthlyReportFn != nil {
		return m.getMonthlyReportFn(ctx, clinicID, year, month)
	}
	return &MonthlyReportResult{}, nil
}

func (m *mockAccountingRepository) GetMonthlyReportByPeriod(ctx context.Context, clinicID uint64, start, end time.Time) (*MonthlyReportResult, error) {
	if m.getMonthlyReportByPeriodFn != nil {
		return m.getMonthlyReportByPeriodFn(ctx, clinicID, start, end)
	}
	return &MonthlyReportResult{}, nil
}

func (m *mockAccountingRepository) GetCategoryPaymentAllocationData(ctx context.Context, clinicID uint64, periodStart, periodEnd time.Time) (*CategoryPaymentAllocationData, error) {
	if m.getCategoryPaymentAllocationDataFn != nil {
		return m.getCategoryPaymentAllocationDataFn(ctx, clinicID, periodStart, periodEnd)
	}
	return &CategoryPaymentAllocationData{CategoryCounts: map[string]int64{}}, nil
}

func (m *mockAccountingRepository) SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.sumPaidByOwnerFn != nil {
		return m.sumPaidByOwnerFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

func (m *mockAccountingRepository) MaxSingleVisitAmountByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepository) FindOwnersByAnnualRevenue(_ context.Context, _ uint64) ([]OwnerAnnualRevenue, error) {
	return nil, nil
}

func (m *mockAccountingRepository) FindMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
	if m.findMonthlyUnpaidCarryoverFn != nil {
		return m.findMonthlyUnpaidCarryoverFn(ctx, clinicID, firstDay, lastDay, page, limit)
	}
	return nil, 0, MonthlyUnpaidSummary{}, nil
}

func (m *mockAccountingRepository) FindByCompletionRequestID(ctx context.Context, clinicID uint64, requestID string) (*model.Billing, error) {
	if m.findByCompletionRequestIDFn != nil {
		return m.findByCompletionRequestIDFn(ctx, clinicID, requestID)
	}
	return nil, nil
}

// （def残存=accounting系はB④・再宣言規約）。

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

// mockHospitalizationRepository — billingHospitalizationFinder view の最小モック（AUD-002）。
type mockHospitalizationRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
}

func (m *mockHospitalizationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Hospitalization{ID: id, ClinicID: clinicID}, nil
}

// mockLstepTagSyncService — cpmTagSyncer view の最小モック（best-effort CPM同期）。
type mockLstepTagSyncService struct {
	syncCPMStageTagFn func(ctx context.Context, clinicID, ownerID uint64) error
}

func (m *mockLstepTagSyncService) SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncCPMStageTagFn != nil {
		return m.syncCPMStageTagFn(ctx, clinicID, ownerID)
	}
	return nil
}

// mockClinicHolidayRepository — billingHolidayReader view（FindAllByYearMonth のみ）の最小モック。
type mockClinicHolidayRepository struct {
	findByYearMonthFn func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
}

func (m *mockClinicHolidayRepository) FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	return m.findByYearMonthFn(ctx, clinicID, yearMonth)
}

// mockClinicRepository — billingClinicReader view（FindByID のみ）の最小モック。
type mockClinicRepository struct {
	findByIDFn func(ctx context.Context, id uint64) (*model.Clinic, error)
}

func (m *mockClinicRepository) FindByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &model.Clinic{ID: id}, nil
}
