package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- モック: CashRegisterCloseRepository ----

type mockCashRegisterCloseRepository struct {
	createFn              func(ctx context.Context, c *model.CashRegisterClose) error
	findAllFn             func(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error)
	findByIDFn            func(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error)
	findByDateAndPeriodFn func(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error)
	hasCloseOnDateFn      func(ctx context.Context, clinicID uint64, date time.Time) (bool, error)
}

func (m *mockCashRegisterCloseRepository) Create(ctx context.Context, c *model.CashRegisterClose) error {
	if m.createFn != nil {
		return m.createFn(ctx, c)
	}
	return nil
}

func (m *mockCashRegisterCloseRepository) FindAll(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, startDate, endDate, page, limit)
	}
	return nil, 0, nil
}

func (m *mockCashRegisterCloseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockCashRegisterCloseRepository) FindByDateAndPeriod(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error) {
	if m.findByDateAndPeriodFn != nil {
		return m.findByDateAndPeriodFn(ctx, clinicID, date, period)
	}
	return nil, nil
}

func (m *mockCashRegisterCloseRepository) HasCloseOnDate(ctx context.Context, clinicID uint64, date time.Time) (bool, error) {
	if m.hasCloseOnDateFn != nil {
		return m.hasCloseOnDateFn(ctx, clinicID, date)
	}
	return false, nil
}

// ---- モック: AccountingRepository（GetCloseAggregate 用追加スタブ） ----
// accounting_service_test.go の mockAccountingRepository にスタブが既に定義されているため、
// ここでは cash_register テスト専用の関数フィールド付きモックを別名で定義する。

type mockAccountingRepositoryForClose struct {
	getCloseAggregateFn func(ctx context.Context, input repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error)
	getMonthlyReportFn  func(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error)
}

func (m *mockAccountingRepositoryForClose) FindAll(_ context.Context, _ uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForClose) FindAllForClinics(_ context.Context, _ []uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForClose) FindByID(_ context.Context, _, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForClose) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForClose) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForClose) Create(_ context.Context, _ uint64, _ *model.Billing) error {
	return nil
}
func (m *mockAccountingRepositoryForClose) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForClose) UpsertPayment(_ context.Context, _ *model.Payment) error {
	return nil
}
func (m *mockAccountingRepositoryForClose) SavePayment(_ context.Context, _ *model.Payment) error {
	return nil
}

func (m *mockAccountingRepositoryForClose) SavePaymentSplits(_ context.Context, _ []model.PaymentSplit) error {
	return nil
}

func (m *mockAccountingRepositoryForClose) CompleteAccountingAppointments(_ context.Context, _ uint64, _, _, _ *uint64, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepositoryForClose) FindUnpaidByBilling(_ context.Context, _ uint64, _, _ string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForClose) FindUnpaidByOwner(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	return nil, 0, repository.UnpaidSummary{}, nil
}
func (m *mockAccountingRepositoryForClose) SumUnpaidByOwner(_ context.Context, _, _ uint64) (repository.OwnerUnpaidBalance, error) {
	return repository.OwnerUnpaidBalance{}, nil
}
func (m *mockAccountingRepositoryForClose) FindMonthlyUnpaidCarryover(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error) {
	return nil, 0, repository.MonthlyUnpaidSummary{}, nil
}
func (m *mockAccountingRepositoryForClose) GetDailySummary(_ context.Context, _ uint64, _ time.Time) (*repository.DailySummaryResult, error) {
	return &repository.DailySummaryResult{}, nil
}
func (m *mockAccountingRepositoryForClose) GetCloseAggregate(ctx context.Context, input repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
	if m.getCloseAggregateFn != nil {
		return m.getCloseAggregateFn(ctx, input)
	}
	return &repository.CloseAggregateResult{
		PaymentRows:    []repository.PaymentAggregateRow{},
		CategoryRows:   []repository.CategoryAggregateRow{},
		BillingDetails: []repository.CloseBillingDetail{},
		TaxBreakdown:   []repository.TaxBreakdownRow{},
	}, nil
}
func (m *mockAccountingRepositoryForClose) GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error) {
	if m.getMonthlyReportFn != nil {
		return m.getMonthlyReportFn(ctx, clinicID, year, month)
	}
	return &repository.MonthlyReportResult{}, nil
}

func (m *mockAccountingRepositoryForClose) GetMonthlyReportByPeriod(_ context.Context, _ uint64, _, _ time.Time) (*repository.MonthlyReportResult, error) {
	return &repository.MonthlyReportResult{}, nil
}

func (m *mockAccountingRepositoryForClose) SumPaidByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepositoryForClose) MaxSingleVisitAmountByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepositoryForClose) FindOwnersByAnnualRevenue(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
	return nil, nil
}

// ---- モック: ClosingSettingsService（ResolveSchedule のみ） ----

type mockClosingSettingsService struct {
	resolveScheduleFn func(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error)
}

func (m *mockClosingSettingsService) Get(_ context.Context, _ uint64) (*ClosingSettingsResponse, error) {
	return nil, nil
}
func (m *mockClosingSettingsService) ListSpecialPeriods(_ context.Context, _ uint64) ([]model.ClosingSpecialPeriod, error) {
	return nil, nil
}
func (m *mockClosingSettingsService) UpdateStandard(_ context.Context, _ uint64, _ UpdateClinicSettingsInput) (*model.ClinicSettings, error) {
	return nil, nil
}
func (m *mockClosingSettingsService) CreateSpecialPeriod(_ context.Context, _ uint64, _ *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	return nil, nil
}
func (m *mockClosingSettingsService) UpdateSpecialPeriod(_ context.Context, _, _ uint64, _ UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	return nil, nil
}
func (m *mockClosingSettingsService) DeleteSpecialPeriod(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockClosingSettingsService) ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error) {
	if m.resolveScheduleFn != nil {
		return m.resolveScheduleFn(ctx, clinicID, date)
	}
	return &DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30", IsHoliday: false}, nil
}

// ---- ヘルパー ----

func defaultSchedule() *DaySchedule {
	return &DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30", IsHoliday: false}
}

func emptyAggregateResult() *repository.CloseAggregateResult {
	return &repository.CloseAggregateResult{
		PaymentRows:    []repository.PaymentAggregateRow{},
		CategoryRows:   []repository.CategoryAggregateRow{},
		BillingDetails: []repository.CloseBillingDetail{},
		TaxBreakdown:   []repository.TaxBreakdownRow{},
	}
}

func newCashRegisterService(
	closeRepo *mockCashRegisterCloseRepository,
	accountingRepo *mockAccountingRepositoryForClose,
	closingsSvc *mockClosingSettingsService,
	payMethodRepo *mockPaymentMethodMasterRepository,
) CashRegisterService {
	return NewCashRegisterService(closeRepo, accountingRepo, closingsSvc, payMethodRepo)
}

// ---- テスト ----

func TestCashRegisterService_GetPreview(t *testing.T) {
	targetDateStr := "2026-04-20"

	tests := []struct {
		name                  string
		dateStr               string
		period                string
		resolveScheduleFn     func(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error)
		getCloseAggregateFn   func(ctx context.Context, input repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error)
		findByDateAndPeriodFn func(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error)
		findAllPayMethodFn    func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
		wantErr               bool
		wantErrIs             error
		checkResult           func(t *testing.T, got *CashRegisterPreview)
	}{
		{
			name:    "正常: スケジュール解決成功 → CashRegisterPreview を返す（未締め）",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil // 未締め
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "現金", SystemKey: ptrString("cash")},
				}, nil
			},
			checkResult: func(t *testing.T, got *CashRegisterPreview) {
				assert.Equal(t, "2026-04-20", got.Date)
				assert.Equal(t, "am", got.Period)
				assert.False(t, got.IsAlreadyClosed)
				assert.False(t, got.IsHoliday)
				assert.Empty(t, got.BillingDetails)
			},
		},
		{
			name:    "正常: すでに締め済みの日時 → IsAlreadyClosed=true",
			dateStr: targetDateStr,
			period:  "pm",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return &model.CashRegisterClose{ID: 99, Period: "pm"}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "現金", SystemKey: ptrString("cash")},
				}, nil
			},
			checkResult: func(t *testing.T, got *CashRegisterPreview) {
				assert.True(t, got.IsAlreadyClosed)
			},
		},
		{
			name:      "エラー: date が空 → ErrInvalidInput",
			dateStr:   "",
			period:    "am",
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:      "エラー: date のフォーマットが不正 → ErrInvalidInput",
			dateStr:   "20260420",
			period:    "am",
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:      "エラー: period が空 → ErrInvalidInput",
			dateStr:   targetDateStr,
			period:    "",
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:      "エラー: period が不正 → ErrInvalidInput",
			dateStr:   targetDateStr,
			period:    "noon",
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:    "正常: period=emg → CashRegisterPreview を返す（#150）",
			dateStr: targetDateStr,
			period:  "emg",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "現金", SystemKey: ptrString("cash")},
				}, nil
			},
			checkResult: func(t *testing.T, got *CashRegisterPreview) {
				assert.Equal(t, "emg", got.Period)
				assert.False(t, got.IsAlreadyClosed)
			},
		},
		{
			name:    "正常: 混在会計 → TheoreticalCash は現金分のみ・カテゴリ按分を確認",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
				cashID := uint64(2)
				creditID := uint64(1)
				return &repository.CloseAggregateResult{
					PaymentRows: []repository.PaymentAggregateRow{
						{PaymentMethodID: &cashID, Amount: 3000},   // 現金
						{PaymentMethodID: &creditID, Amount: 7000}, // クレジット
					},
					CategoryRows: []repository.CategoryAggregateRow{
						{Category: "診察", Amount: 10000},
					},
					TotalRefund:    0,
					BillingDetails: []repository.CloseBillingDetail{},
					TaxBreakdown:   []repository.TaxBreakdownRow{},
				}, nil
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "クレジット", SystemKey: ptrString("credit_card")},
					{ID: 2, Name: "現金", SystemKey: ptrString("cash")},
				}, nil
			},
			checkResult: func(t *testing.T, got *CashRegisterPreview) {
				assert.Equal(t, int64(3000), got.Aggregate.TheoreticalCash)
				// カテゴリ別按分: 診察 10000 を現金3000/クレジット7000 で按分
				diagCats := got.Aggregate.Categories["診察"]
				assert.Equal(t, int64(3000), diagCats["現金"])
				assert.Equal(t, int64(7000), diagCats["クレジット"])
			},
		},
		{
			name:    "エラー: ResolveSchedule がエラーを返す",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return nil, errors.New("schedule error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			closeRepo := &mockCashRegisterCloseRepository{
				findByDateAndPeriodFn: tt.findByDateAndPeriodFn,
			}
			accountingRepo := &mockAccountingRepositoryForClose{
				getCloseAggregateFn: tt.getCloseAggregateFn,
			}
			closingsSvc := &mockClosingSettingsService{
				resolveScheduleFn: tt.resolveScheduleFn,
			}
			payMethodRepo := &mockPaymentMethodMasterRepository{
				findAllFn: tt.findAllPayMethodFn,
			}
			svc := newCashRegisterService(closeRepo, accountingRepo, closingsSvc, payMethodRepo)

			// Act
			got, err := svc.GetPreview(context.Background(), 1, tt.dateStr, tt.period)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "want errors.Is(%v), got %v", tt.wantErrIs, err)
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}

func TestCashRegisterService_Close(t *testing.T) {
	targetDate := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	validInput := CloseRegisterInput{
		Date:       targetDate,
		Period:     "am",
		ActualCash: 50000,
		Memo:       "通常締め",
	}

	tests := []struct {
		name                  string
		input                 CloseRegisterInput
		findByDateAndPeriodFn func(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error)
		resolveScheduleFn     func(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error)
		getCloseAggregateFn   func(ctx context.Context, input repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error)
		findAllPayMethodFn    func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
		createFn              func(ctx context.Context, c *model.CashRegisterClose) error
		wantErr               bool
		wantErrIs             error
		checkResult           func(t *testing.T, got *model.CashRegisterClose)
	}{
		{
			name:  "正常: CashRegisterClose を返す",
			input: validInput,
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil // 未締め
			},
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			createFn: func(_ context.Context, _ *model.CashRegisterClose) error {
				return nil
			},
			checkResult: func(t *testing.T, got *model.CashRegisterClose) {
				assert.Equal(t, uint64(1), got.ClinicID)
				assert.Equal(t, "am", got.Period)
				assert.Equal(t, int64(50000), got.ActualCash)
				assert.Equal(t, "通常締め", got.Memo)
				// 集計が空なので TheoreticalCash=0, CashDifference=50000
				assert.Equal(t, int64(0), got.TheoreticalCash)
				assert.Equal(t, int64(50000), got.CashDifference)
			},
		},
		{
			name: "正常: 混在会計 + 返金 → TheoreticalCash = 現金 - 返金、CashDifference = 実際 - 理論",
			input: CloseRegisterInput{
				Date:       targetDate,
				Period:     "am",
				ActualCash: 4000,
				Memo:       "",
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
				cashID := uint64(2)
				creditID := uint64(1)
				return &repository.CloseAggregateResult{
					PaymentRows: []repository.PaymentAggregateRow{
						{PaymentMethodID: &cashID, Amount: 3000},   // 現金
						{PaymentMethodID: &creditID, Amount: 7000}, // クレジット
					},
					CategoryRows:   []repository.CategoryAggregateRow{},
					TotalRefund:    500,
					BillingDetails: []repository.CloseBillingDetail{},
					TaxBreakdown:   []repository.TaxBreakdownRow{},
				}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "クレジットカード", SystemKey: ptrString("credit_card")},
					{ID: 2, Name: "現金", SystemKey: ptrString("cash")},
				}, nil
			},
			createFn: func(_ context.Context, _ *model.CashRegisterClose) error {
				return nil
			},
			checkResult: func(t *testing.T, got *model.CashRegisterClose) {
				// TheoreticalCash = 現金(3000) - 返金(500) = 2500
				assert.Equal(t, int64(2500), got.TheoreticalCash)
				// CashDifference = ActualCash(4000) - TheoreticalCash(2500) = 1500
				assert.Equal(t, int64(1500), got.CashDifference)
			},
		},
		{
			name: "正常: period=emg の締めを作成できる（#150）",
			input: CloseRegisterInput{
				Date:       targetDate,
				Period:     "emg",
				ActualCash: 10000,
				Memo:       "夜間緊急",
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			createFn: func(_ context.Context, _ *model.CashRegisterClose) error {
				return nil
			},
			checkResult: func(t *testing.T, got *model.CashRegisterClose) {
				assert.Equal(t, "emg", got.Period)
				assert.Equal(t, int64(10000), got.ActualCash)
			},
		},
		{
			name:  "エラー: 二重締め（FindByDateAndPeriod が既存レコードを返す） → ErrConflict",
			input: validInput,
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return &model.CashRegisterClose{ID: 1, Period: "am"}, nil
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrConflict,
		},
		{
			name: "エラー: period が不正 → ErrInvalidInput",
			input: CloseRegisterInput{
				Date:   targetDate,
				Period: "invalid",
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:  "エラー: ResolveSchedule がエラーを返す",
			input: validInput,
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return nil, errors.New("schedule error")
			},
			wantErr: true,
		},
		{
			name:  "エラー: Create がエラーを返す",
			input: validInput,
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			createFn: func(_ context.Context, _ *model.CashRegisterClose) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			closeRepo := &mockCashRegisterCloseRepository{
				findByDateAndPeriodFn: tt.findByDateAndPeriodFn,
				createFn:              tt.createFn,
			}
			accountingRepo := &mockAccountingRepositoryForClose{
				getCloseAggregateFn: tt.getCloseAggregateFn,
			}
			closingsSvc := &mockClosingSettingsService{
				resolveScheduleFn: tt.resolveScheduleFn,
			}
			findAllFn := tt.findAllPayMethodFn
			if findAllFn == nil {
				findAllFn = func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
					return []model.PaymentMethodMaster{
						{ID: 1, Name: "現金", SystemKey: ptrString("cash")},
					}, nil
				}
			}
			payMethodRepo := &mockPaymentMethodMasterRepository{findAllFn: findAllFn}
			svc := newCashRegisterService(closeRepo, accountingRepo, closingsSvc, payMethodRepo)

			// Act
			got, err := svc.Close(context.Background(), 1, tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "want errors.Is(%v), got %v", tt.wantErrIs, err)
				}
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}

// TestCashRegisterService_IsDateClosed は #115 締め後編集チェックをテストする。
func TestCashRegisterService_IsDateClosed(t *testing.T) {
	date := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		hasCloseOnDateFn func(ctx context.Context, clinicID uint64, d time.Time) (bool, error)
		wantClosed       bool
		wantErr          bool
	}{
		{
			name: "正常: 締め済み → true",
			hasCloseOnDateFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) {
				return true, nil
			},
			wantClosed: true,
		},
		{
			name: "正常: 未締め → false",
			hasCloseOnDateFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) {
				return false, nil
			},
			wantClosed: false,
		},
		{
			name: "エラー: リポジトリエラー → wrappedエラーを返す",
			hasCloseOnDateFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) {
				return false, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closeRepo := &mockCashRegisterCloseRepository{
				hasCloseOnDateFn: tt.hasCloseOnDateFn,
			}
			svc := newCashRegisterService(closeRepo, &mockAccountingRepositoryForClose{}, &mockClosingSettingsService{}, &mockPaymentMethodMasterRepository{})

			got, err := svc.IsDateClosed(context.Background(), 1, date)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantClosed, got)
		})
	}
}

// TestResolvePeriodRange は period→集計期間（JST）の算出を検証する。
// #150 で追加した emg は PM 締め終了〜翌日0時を対象とする。
func TestResolvePeriodRange(t *testing.T) {
	dateJST := time.Date(2026, 4, 20, 0, 0, 0, 0, jstLocation)
	schedule := &DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30"}

	tests := []struct {
		name      string
		period    string
		wantStart time.Time
		wantEnd   time.Time
		wantErr   bool
	}{
		{
			name:      "am: 0時〜AM/PM境界",
			period:    "am",
			wantStart: time.Date(2026, 4, 20, 0, 0, 0, 0, jstLocation),
			wantEnd:   time.Date(2026, 4, 20, 14, 0, 0, 0, jstLocation),
		},
		{
			name:      "pm: 境界〜PM締め終了",
			period:    "pm",
			wantStart: time.Date(2026, 4, 20, 14, 0, 0, 0, jstLocation),
			wantEnd:   time.Date(2026, 4, 20, 18, 30, 0, 0, jstLocation),
		},
		{
			name:      "emg: PM締め終了〜翌日0時（PM終了と連続・非破壊）",
			period:    "emg",
			wantStart: time.Date(2026, 4, 20, 18, 30, 0, 0, jstLocation),
			wantEnd:   time.Date(2026, 4, 21, 0, 0, 0, 0, jstLocation),
		},
		{
			name:    "エラー: 不正な period",
			period:  "noon",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := resolvePeriodRange(dateJST, tt.period, schedule)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, tt.wantStart.Equal(start), "start: want %v, got %v", tt.wantStart, start)
			assert.True(t, tt.wantEnd.Equal(end), "end: want %v, got %v", tt.wantEnd, end)
		})
	}
}

// TestFindCashMethodID は payment_methods マスタ群から現金マスタ id を system_key 一致で解決することを検証する（#197）。
func TestFindCashMethodID(t *testing.T) {
	ptr := func(s string) *string { return &s }

	t.Run("system_key='cash' の行が正しく解決される", func(t *testing.T) {
		methods := []model.PaymentMethodMaster{
			{ID: 101, Name: "現金", SystemKey: ptr("cash")},
			{ID: 102, Name: "クレジットカード", SystemKey: ptr("credit_card")},
			{ID: 104, Name: "銀行振込", SystemKey: ptr("bank_transfer")},
		}
		got, err := findCashMethodID(methods)
		assert.NoError(t, err)
		assert.Equal(t, uint64(101), got)
	})

	t.Run("rename 耐性: name='お金' に改名しても system_key='cash' で正しく解決される", func(t *testing.T) {
		renamed := []model.PaymentMethodMaster{
			{ID: 101, Name: "お金", SystemKey: ptr("cash")},
		}
		got, err := findCashMethodID(renamed)
		assert.NoError(t, err)
		assert.Equal(t, uint64(101), got)
	})

	t.Run("現金マスタが存在しない場合はエラーを返す", func(t *testing.T) {
		_, err := findCashMethodID([]model.PaymentMethodMaster{{ID: 102, Name: "クレジット", SystemKey: ptr("credit_card")}})
		assert.Error(t, err)
	})
}

// TestCalcTheoreticalCash は理論現金の現金判定を検証する（#128/#197）。
// 現金 = payment_method_id が現金マスタ id と一致する split（#197 の新データ）、
// または payment_method_id=NULL の split（旧 seed/レガシーデータの現金: #128 後方互換, bug.md H-3）。
func TestCalcTheoreticalCash(t *testing.T) {
	cashID := uint64(101)
	creditID := uint64(102)
	unknownID := uint64(999) // どの行の payment_method_id とも一致しない現金マスタ id

	tests := []struct {
		name         string
		rows         []repository.PaymentAggregateRow
		totalRefund  int64
		cashMethodID uint64
		want         int64
	}{
		{
			name: "現金マスタ id の行を現金として加算する",
			rows: []repository.PaymentAggregateRow{
				{PaymentMethodID: &cashID, Amount: 5000},
				{PaymentMethodID: &creditID, Amount: 3000},
			},
			cashMethodID: cashID,
			want:         5000,
		},
		{
			name: "payment_method_id=NULL の行も現金として加算する（レガシー seed 後方互換, bug.md H-3）",
			rows: []repository.PaymentAggregateRow{
				{PaymentMethodID: nil, Amount: 4000},
				{PaymentMethodID: &creditID, Amount: 1000},
			},
			cashMethodID: cashID,
			want:         4000,
		},
		{
			name: "現金マスタ id と NULL が混在しても両方加算する",
			rows: []repository.PaymentAggregateRow{
				{PaymentMethodID: &cashID, Amount: 5000},
				{PaymentMethodID: nil, Amount: 2000},
				{PaymentMethodID: &creditID, Amount: 3000},
			},
			cashMethodID: cashID,
			want:         7000,
		},
		{
			name: "現金マスタ id がどの行にも一致しなくても NULL 行は現金として加算する（NULL 後方互換は id 一致に非依存）",
			rows: []repository.PaymentAggregateRow{
				{PaymentMethodID: nil, Amount: 4000},
				{PaymentMethodID: &cashID, Amount: 5000},
			},
			cashMethodID: unknownID,
			want:         4000,
		},
		{
			name: "非現金（credit_card）のみは現金 0",
			rows: []repository.PaymentAggregateRow{
				{PaymentMethodID: &creditID, Amount: 3000},
			},
			cashMethodID: cashID,
			want:         0,
		},
		{
			name: "返金合計を理論現金から差し引く",
			rows: []repository.PaymentAggregateRow{
				{PaymentMethodID: &cashID, Amount: 5000},
			},
			totalRefund:  1200,
			cashMethodID: cashID,
			want:         3800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcTheoreticalCash(tt.rows, tt.totalRefund, tt.cashMethodID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestCashRegisterService_Close_ExcludesActualCashFromAggregation は #152 の不変条件を固定する。
// 締め集計サマリー（永続化される category_breakdown JSONB と理論現金）は医療記録ベースの金額のみで構成され、
// レジ実査入力 actual_cash / cash_difference は一切混入しないことを検証する。
// 現コードは既にこの分離を満たしており（production バグは不在）、本テストは将来の回帰
// （actual_cash を buildCategoryBreakdown / calcTheoreticalCash に取り込む変更）を検出する特性化テストである。
func TestCashRegisterService_Close_ExcludesActualCashFromAggregation(t *testing.T) {
	targetDate := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	// 医療記録の合計は 10000（現金3000 + クレジット7000）。実査現金は意図的に大きく乖離させる。
	const (
		medicalRecordTotal = int64(10000)
		actualCash         = int64(99999)
		cashBilling        = int64(3000)
		creditBilling      = int64(7000)
	)

	cashID := uint64(2)
	creditID := uint64(1)

	var created *model.CashRegisterClose
	closeRepo := &mockCashRegisterCloseRepository{
		findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
			return nil, nil // 未締め
		},
		createFn: func(_ context.Context, c *model.CashRegisterClose) error {
			created = c
			return nil
		},
	}
	accountingRepo := &mockAccountingRepositoryForClose{
		getCloseAggregateFn: func(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
			return &repository.CloseAggregateResult{
				PaymentRows: []repository.PaymentAggregateRow{
					{PaymentMethodID: &cashID, Amount: cashBilling},     // 現金
					{PaymentMethodID: &creditID, Amount: creditBilling}, // クレジット
				},
				CategoryRows: []repository.CategoryAggregateRow{
					{Category: "診察", Amount: medicalRecordTotal},
				},
				TotalRefund:    0,
				BillingDetails: []repository.CloseBillingDetail{},
				TaxBreakdown:   []repository.TaxBreakdownRow{},
			}, nil
		},
	}
	closingsSvc := &mockClosingSettingsService{
		resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*DaySchedule, error) {
			return defaultSchedule(), nil
		},
	}
	payMethodRepo := &mockPaymentMethodMasterRepository{
		findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{
				{ID: 1, Name: "クレジットカード", SystemKey: ptrString("credit_card")},
				{ID: 2, Name: "現金", SystemKey: ptrString("cash")},
			}, nil
		},
	}
	svc := newCashRegisterService(closeRepo, accountingRepo, closingsSvc, payMethodRepo)

	// Act
	got, err := svc.Close(context.Background(), 1, CloseRegisterInput{
		Date:       targetDate,
		Period:     "am",
		ActualCash: actualCash,
		Memo:       "実査過剰",
	})

	// Assert: 理論現金は請求の現金分のみ（実査現金 99999 ではない）
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, cashBilling, got.TheoreticalCash, "理論現金は請求の現金分(3000)のみで、actual_cash は含まない")
	assert.Equal(t, actualCash, got.ActualCash, "actual_cash は照合用に別フィールドへ保存される")
	assert.Equal(t, actualCash-cashBilling, got.CashDifference, "cash_difference は照合差異であり集計サマリーとは別")

	// Assert: 永続化された category_breakdown JSONB は医療記録金額のみ。actual_cash は混入しない。
	assert.NotNil(t, created)
	var schema model.CategoryBreakdownSchema
	if !assert.NoError(t, json.Unmarshal(created.CategoryBreakdown, &schema)) {
		return
	}

	diag := schema.Categories["診察"]
	assert.Equal(t, cashBilling, diag["cash"], "診察カテゴリの現金按分は請求現金(3000)")
	assert.Equal(t, creditBilling, diag["credit_card"], "診察カテゴリのクレジット按分は請求クレジット(7000)")

	// カテゴリ内訳の総額は医療記録合計(10000)に一致し、actual_cash(99999) で汚染されていない。
	var breakdownTotal int64
	for _, byMethod := range schema.Categories {
		for _, amount := range byMethod {
			breakdownTotal += amount
			assert.NotEqual(t, actualCash, amount, "category_breakdown のいかなる金額も actual_cash と一致してはならない")
		}
	}
	assert.Equal(t, medicalRecordTotal, breakdownTotal, "category_breakdown の総額は医療記録合計のみ（actual_cash 不含）")
}
