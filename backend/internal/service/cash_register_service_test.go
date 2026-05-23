package service

import (
	"context"
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
func (m *mockAccountingRepositoryForClose) FindByID(_ context.Context, _, _ uint64) (*model.Billing, error) {
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

func (m *mockAccountingRepositoryForClose) FindUnpaidByBilling(_ context.Context, _ uint64, _ string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForClose) FindUnpaidByOwner(_ context.Context, _ uint64, _ string, _, _ int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	return nil, 0, repository.UnpaidSummary{}, nil
}
func (m *mockAccountingRepositoryForClose) GetDailySummary(_ context.Context, _ uint64, _ time.Time) (*repository.DailySummaryResult, error) {
	return &repository.DailySummaryResult{}, nil
}
func (m *mockAccountingRepositoryForClose) GetCloseAggregate(ctx context.Context, input repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
	if m.getCloseAggregateFn != nil {
		return m.getCloseAggregateFn(ctx, input)
	}
	return &repository.CloseAggregateResult{
		AggregateRows:  []repository.BillingAggregateRow{},
		BillingDetails: []repository.CloseBillingDetail{},
		TaxBreakdown:   []repository.TaxBreakdownRow{},
	}, nil
}
func (m *mockAccountingRepositoryForClose) GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error) {
	if m.getMonthlyReportFn != nil {
		return m.getMonthlyReportFn(ctx, clinicID, year, month)
	}
	return &repository.MonthlyReportResult{Rows: []repository.MonthlyReportRow{}}, nil
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
		AggregateRows:  []repository.BillingAggregateRow{},
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
				return []model.PaymentMethodMaster{}, nil
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
				return []model.PaymentMethodMaster{}, nil
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
			payMethodRepo := &mockPaymentMethodMasterRepository{}
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
