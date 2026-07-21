package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockAccountingRepository は repository.AccountingRepository のテスト用モック実装（F-4統合正本）。
// accounting_service_test.go の mockAccountingRepository を正本とし、
// mockAccountingRepositoryForReport（accounting_report_service_test.go）・
// mockAccountingRepositoryForClose（cash_register_service_test.go）・
// mockAccountingRepoForLstepVisit（lstep_tag_sync_visit_test.go）の実際に検証されている
// フックを統合する。FindAll/Create/Update は旧正本ではフック未設定時に nil 関数呼び出しで
// panic する契約だったため、統合を機に nil ガードを追加した（3変種いずれもこの3メソッドを
// 呼ばない対象サービスにのみ使われていたため、この変更で既存テストの挙動は変わらない）。
type mockAccountingRepository struct {
	findAllFn           func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	createFn            func(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	updateFieldsFn      func(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	savePaymentFn       func(ctx context.Context, payment *model.Payment) error
	savePaymentSplitsFn func(ctx context.Context, splits []model.PaymentSplit) error
	completeApptsFn     func(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error)
	getDailySummaryFn   func(ctx context.Context, clinicID uint64, date time.Time) (*repository.DailySummaryResult, error)
	// #120: start_date/end_date 2引数バリアント
	findUnpaidByBillingFn func(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error)
	findUnpaidByOwnerFn   func(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error)
	// #182: 飼主未納残高
	sumUnpaidByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) (repository.OwnerUnpaidBalance, error)
	// #114: 月次未納繰越集計
	findMonthlyUnpaidCarryoverFn func(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error)
	// 以下4フィールドは F-4 統合で追加（旧 ForReport/ForClose/ForLstepVisit が個別に持っていたフック）。
	// 未設定時は各旧モックのデフォルトと同じ値を返す（挙動不変）。
	getCloseAggregateFn        func(ctx context.Context, input repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error)
	getMonthlyReportFn         func(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error)
	getMonthlyReportByPeriodFn func(ctx context.Context, clinicID uint64, start, end time.Time) (*repository.MonthlyReportResult, error)
	sumPaidByOwnerFn           func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

func (m *mockAccountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	}
	return nil, 0, nil
}

func (m *mockAccountingRepository) FindAllForClinics(_ context.Context, _ []uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
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

func (m *mockAccountingRepository) CompleteAccountingAppointments(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error) {
	if m.completeApptsFn != nil {
		return m.completeApptsFn(ctx, clinicID, medicalRecordID, ownerID, petID, scheduledDate)
	}
	return 0, nil
}

// #120: 未納者一覧 repository メソッドの mock（start_date/end_date 2引数）
func (m *mockAccountingRepository) FindUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error) {
	if m.findUnpaidByBillingFn != nil {
		return m.findUnpaidByBillingFn(ctx, clinicID, startDate, endDate, page, limit)
	}
	return nil, 0, nil
}

func (m *mockAccountingRepository) FindUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	if m.findUnpaidByOwnerFn != nil {
		return m.findUnpaidByOwnerFn(ctx, clinicID, startDate, endDate, page, limit)
	}
	return nil, 0, repository.UnpaidSummary{}, nil
}

func (m *mockAccountingRepository) SumUnpaidByOwner(ctx context.Context, clinicID, ownerID uint64) (repository.OwnerUnpaidBalance, error) {
	if m.sumUnpaidByOwnerFn != nil {
		return m.sumUnpaidByOwnerFn(ctx, clinicID, ownerID)
	}
	return repository.OwnerUnpaidBalance{}, nil
}

func (m *mockAccountingRepository) GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*repository.DailySummaryResult, error) {
	if m.getDailySummaryFn != nil {
		return m.getDailySummaryFn(ctx, clinicID, date)
	}
	return &repository.DailySummaryResult{PaymentTotals: []repository.PaymentMethodTotal{}, CategoryTotals: []repository.CategoryTotal{}}, nil
}

// FEAT-368: 集計・締め機能 mock スタブ
func (m *mockAccountingRepository) GetCloseAggregate(ctx context.Context, input repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
	if m.getCloseAggregateFn != nil {
		return m.getCloseAggregateFn(ctx, input)
	}
	return &repository.CloseAggregateResult{
		PaymentRows:    []repository.PaymentAggregateRow{},
		CategoryRows:   []repository.CategoryAggregateRow{},
		BillingDetails: []repository.CloseBillingDetailRow{},
		TaxBreakdown:   []repository.TaxBreakdownRow{},
	}, nil
}

func (m *mockAccountingRepository) GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error) {
	if m.getMonthlyReportFn != nil {
		return m.getMonthlyReportFn(ctx, clinicID, year, month)
	}
	return &repository.MonthlyReportResult{}, nil
}

func (m *mockAccountingRepository) GetMonthlyReportByPeriod(ctx context.Context, clinicID uint64, start, end time.Time) (*repository.MonthlyReportResult, error) {
	if m.getMonthlyReportByPeriodFn != nil {
		return m.getMonthlyReportByPeriodFn(ctx, clinicID, start, end)
	}
	return &repository.MonthlyReportResult{}, nil
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

func (m *mockAccountingRepository) FindOwnersByAnnualRevenue(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
	return nil, nil
}

// #114: 月次未納繰越集計 mock
func (m *mockAccountingRepository) FindMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error) {
	if m.findMonthlyUnpaidCarryoverFn != nil {
		return m.findMonthlyUnpaidCarryoverFn(ctx, clinicID, firstDay, lastDay, page, limit)
	}
	return nil, 0, repository.MonthlyUnpaidSummary{}, nil
}
