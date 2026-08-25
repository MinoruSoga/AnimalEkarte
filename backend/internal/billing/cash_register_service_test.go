package billing

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ---- モック: CashRegisterCloseRepository ----

type mockCashRegisterCloseRepository struct {
	createFn              func(ctx context.Context, c *model.CashRegisterClose) error
	createAdjustmentFn    func(ctx context.Context, adj *model.CashRegisterCloseAdjustment) error
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

func (m *mockCashRegisterCloseRepository) CreateAdjustment(ctx context.Context, adj *model.CashRegisterCloseAdjustment) error {
	if m.createAdjustmentFn != nil {
		return m.createAdjustmentFn(ctx, adj)
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

// ---- モック: closingScheduleResolver（ResolveSchedule のみ） ----

type mockClosingSettingsService struct {
	resolveScheduleFn func(ctx context.Context, clinicID uint64, date time.Time) (*sharedkernel.DaySchedule, error)
}

func (m *mockClosingSettingsService) ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*sharedkernel.DaySchedule, error) {
	if m.resolveScheduleFn != nil {
		return m.resolveScheduleFn(ctx, clinicID, date)
	}
	return &sharedkernel.DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30", IsHoliday: false}, nil
}

// ---- ヘルパー ----

func defaultSchedule() *sharedkernel.DaySchedule {
	return &sharedkernel.DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30", IsHoliday: false}
}

func emptyAggregateResult() *CloseAggregateResult {
	return &CloseAggregateResult{
		PaymentRows:    []PaymentAggregateRow{},
		CategoryRows:   []CategoryAggregateRow{},
		BillingDetails: []CloseBillingDetailRow{},
		TaxBreakdown:   []TaxBreakdownRow{},
	}
}

// allocationDataFromAggregate synthesizes per-billing allocation inputs from a period
// aggregate (unit tests treat the period as one synthetic billing). This keeps existing
// cash-register service tests working after #247 switched off period-wide ratio allocation.
func allocationDataFromAggregate(agg *CloseAggregateResult) *CategoryPaymentAllocationData {
	if agg == nil {
		return &CategoryPaymentAllocationData{CategoryCounts: map[string]int64{}}
	}
	const billingID = uint64(1)
	weights := make([]allocationCategoryWeightRow, 0, len(agg.CategoryRows))
	counts := make(map[string]int64, len(agg.CategoryRows))
	for _, c := range agg.CategoryRows {
		weights = append(weights, allocationCategoryWeightRow{
			BillingID: billingID,
			Category:  c.Category,
			Amount:    c.Amount,
		})
		if c.Amount > 0 {
			counts[c.Category] = 1
		}
	}
	payments := make([]allocationPaymentRow, 0, len(agg.PaymentRows))
	for _, p := range agg.PaymentRows {
		payments = append(payments, allocationPaymentRow{
			BillingID:       billingID,
			PaymentMethodID: p.PaymentMethodID,
			Amount:          p.Amount,
		})
	}
	return &CategoryPaymentAllocationData{
		Weights:        weights,
		Payments:       payments,
		CategoryCounts: counts,
	}
}

func newCashRegisterService(
	closeRepo *mockCashRegisterCloseRepository,
	accountingRepo *mockAccountingRepository,
	closingsSvc *mockClosingSettingsService,
	payMethodRepo *mockPaymentMethodMasterRepository,
) CashRegisterService {
	// clinicRepo の zero-value mock は nil clinic を返し、税率は既定値（標準10%/軽減8%）になる（M-7）。
	return NewCashRegisterService(closeRepo, accountingRepo, closingsSvc, payMethodRepo, &mockClinicRepository{})
}

// ---- テスト ----

func TestCashRegisterService_GetPreview(t *testing.T) {
	targetDateStr := "2026-04-20"

	tests := []struct {
		name                  string
		dateStr               string
		period                string
		resolveScheduleFn     func(ctx context.Context, clinicID uint64, date time.Time) (*sharedkernel.DaySchedule, error)
		getCloseAggregateFn   func(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error)
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				cashID := uint64(2)
				creditID := uint64(1)
				return &CloseAggregateResult{
					PaymentRows: []PaymentAggregateRow{
						{PaymentMethodID: &cashID, Amount: 3000},   // 現金
						{PaymentMethodID: &creditID, Amount: 7000}, // クレジット
					},
					CategoryRows: []CategoryAggregateRow{
						{Category: string(model.ItemCategoryExamination), Amount: 10000},
					},
					TotalRefund:    0,
					BillingDetails: []CloseBillingDetailRow{},
					TaxBreakdown:   []TaxBreakdownRow{},
				}, nil
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "クレジット", SystemKey: ptrString("credit_card"), IsActive: true},
					{ID: 2, Name: "現金", SystemKey: ptrString("cash"), IsActive: true},
				}, nil
			},
			checkResult: func(t *testing.T, got *CashRegisterPreview) {
				assert.Equal(t, int64(3000), got.Aggregate.TheoreticalCash)
				// #247: 支払実額基準の per-billing 配賦
				diagCats := got.Aggregate.Categories[string(model.ItemCategoryExamination)]
				assert.Equal(t, int64(3000), diagCats["現金"])
				assert.Equal(t, int64(7000), diagCats["クレジット"])
			},
		},
		{
			name:    "正常: 個別会計一覧（BillingDetails）が正しく変換される",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				cashID := uint64(2)
				return &CloseAggregateResult{
					PaymentRows:  []PaymentAggregateRow{},
					CategoryRows: []CategoryAggregateRow{},
					BillingDetails: []CloseBillingDetailRow{
						{
							BillingID:         500,
							PaidAt:            time.Date(2026, 4, 20, 10, 30, 0, 0, time.UTC),
							OwnerName:         "田中太郎",
							PetName:           "ポチ",
							IsHospitalization: false,
							Category:          string(model.ItemCategoryExamination),
							PaymentMethodID:   &cashID,
							BillingAmount:     5000,
							RefundAmount:      0,
							NetAmount:         5000,
						},
					},
					TaxBreakdown: []TaxBreakdownRow{},
				}, nil
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 2, Name: "現金", SystemKey: ptrString("cash")},
				}, nil
			},
			checkResult: func(t *testing.T, got *CashRegisterPreview) {
				if assert.Len(t, got.BillingDetails, 1) {
					d := got.BillingDetails[0]
					assert.Equal(t, uint64(500), d.BillingID)
					assert.Equal(t, "田中太郎", d.OwnerName)
					assert.Equal(t, "ポチ", d.PetName)
					assert.Equal(t, "現金", d.PaymentMethodName)
					assert.Equal(t, int64(5000), d.BillingAmount)
					assert.Equal(t, int64(5000), d.NetAmount)
				}
			},
		},
		{
			name:    "エラー: ResolveSchedule がエラーを返す",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return nil, errors.New("schedule error")
			},
			wantErr: true,
		},
		{
			name:    "エラー: resolvePeriodRange が失敗する（スケジュールの時刻形式が不正）",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return &sharedkernel.DaySchedule{AmPmBoundary: "invalid", PmEnd: "18:30", AmStart: "09:00"}, nil
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:    "エラー: GetCloseAggregate がエラーを返す",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				return nil, errors.New("aggregate error")
			},
			wantErr: true,
		},
		{
			name:    "エラー: 支払方法マスタ取得に失敗する",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return nil, errors.New("pay method error")
			},
			wantErr: true,
		},
		{
			name:    "エラー: 現金マスタ（system_key='cash'）が存在しない",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "クレジットカード", SystemKey: ptrString("credit_card")},
				}, nil
			},
			wantErr: true,
		},
		{
			name:    "エラー: 二重締め確認（FindByDateAndPeriod）がエラーを返す",
			dateStr: targetDateStr,
			period:  "am",
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "現金", SystemKey: ptrString("cash")},
				}, nil
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, errors.New("find close error")
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
			accountingRepo := &mockAccountingRepository{
				getCloseAggregateFn: tt.getCloseAggregateFn,
				getCategoryPaymentAllocationDataFn: func(ctx context.Context, clinicID uint64, periodStart, periodEnd time.Time) (*CategoryPaymentAllocationData, error) {
					if tt.getCloseAggregateFn == nil {
						return allocationDataFromAggregate(emptyAggregateResult()), nil
					}
					agg, err := tt.getCloseAggregateFn(ctx, GetCloseAggregateInput{
						ClinicID: clinicID, PeriodStart: periodStart, PeriodEnd: periodEnd,
					})
					if err != nil {
						return nil, err
					}
					return allocationDataFromAggregate(agg), nil
				},
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

	holidayCreateCalled := &atomic.Bool{}
	negativeCreateCalled := &atomic.Bool{}

	tests := []struct {
		name                  string
		input                 CloseRegisterInput
		findByDateAndPeriodFn func(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error)
		resolveScheduleFn     func(ctx context.Context, clinicID uint64, date time.Time) (*sharedkernel.DaySchedule, error)
		getCloseAggregateFn   func(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error)
		findAllPayMethodFn    func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
		createFn              func(ctx context.Context, c *model.CashRegisterClose) error
		createCalled          *atomic.Bool
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				cashID := uint64(2)
				creditID := uint64(1)
				return &CloseAggregateResult{
					PaymentRows: []PaymentAggregateRow{
						{PaymentMethodID: &cashID, Amount: 3000},   // 現金
						{PaymentMethodID: &creditID, Amount: 7000}, // クレジット
					},
					CategoryRows:   []CategoryAggregateRow{},
					TotalRefund:    500,
					BillingDetails: []CloseBillingDetailRow{},
					TaxBreakdown:   []TaxBreakdownRow{},
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
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
			name:  "エラー: 休診日は締め処理できない → ErrInvalidInput",
			input: validInput,
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				s := defaultSchedule()
				return &sharedkernel.DaySchedule{
					AmPmBoundary: s.AmPmBoundary,
					PmEnd:        s.PmEnd,
					IsHoliday:    true,
				}, nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			createFn: func(_ context.Context, _ *model.CashRegisterClose) error {
				holidayCreateCalled.Store(true)
				return nil
			},
			createCalled: holidayCreateCalled,
			wantErr:      true,
			wantErrIs:    apperrors.ErrInvalidInput,
		},
		{
			name: "エラー: ActualCash が負 → ErrInvalidInput",
			input: CloseRegisterInput{
				Date:       validInput.Date,
				Period:     validInput.Period,
				ActualCash: -1,
				Memo:       validInput.Memo,
				ClosedBy:   validInput.ClosedBy,
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, nil
			},
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
				return emptyAggregateResult(), nil
			},
			createFn: func(_ context.Context, _ *model.CashRegisterClose) error {
				negativeCreateCalled.Store(true)
				return nil
			},
			createCalled: negativeCreateCalled,
			wantErr:      true,
			wantErrIs:    apperrors.ErrInvalidInput,
		},
		{
			name:  "エラー: 二重締め確認（FindByDateAndPeriod）がリポジトリエラーを返す",
			input: validInput,
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
				return nil, errors.New("find close error")
			},
			wantErr: true,
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
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
			resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
				return defaultSchedule(), nil
			},
			getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
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
			accountingRepo := &mockAccountingRepository{
				getCloseAggregateFn: tt.getCloseAggregateFn,
				getCategoryPaymentAllocationDataFn: func(ctx context.Context, clinicID uint64, periodStart, periodEnd time.Time) (*CategoryPaymentAllocationData, error) {
					if tt.getCloseAggregateFn == nil {
						return allocationDataFromAggregate(emptyAggregateResult()), nil
					}
					agg, err := tt.getCloseAggregateFn(ctx, GetCloseAggregateInput{
						ClinicID: clinicID, PeriodStart: periodStart, PeriodEnd: periodEnd,
					})
					if err != nil {
						return nil, err
					}
					return allocationDataFromAggregate(agg), nil
				},
			}
			closingsSvc := &mockClosingSettingsService{
				resolveScheduleFn: tt.resolveScheduleFn,
			}
			findAllFn := tt.findAllPayMethodFn
			if findAllFn == nil {
				findAllFn = func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
					return []model.PaymentMethodMaster{
						{ID: 1, Name: "現金", SystemKey: ptrString("cash"), IsActive: true},
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
				if tt.createCalled != nil {
					assert.False(t, tt.createCalled.Load(), "closeRepo.Create must not be called")
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
			svc := newCashRegisterService(closeRepo, &mockAccountingRepository{}, &mockClosingSettingsService{}, &mockPaymentMethodMasterRepository{})

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
	dateJST := time.Date(2026, 4, 20, 0, 0, 0, 0, config.JST)
	schedule := &sharedkernel.DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30", AmStart: "09:00"}

	tests := []struct {
		name      string
		period    string
		wantStart time.Time
		wantEnd   time.Time
		wantErr   bool
	}{
		{
			name:      "am: am_start(9:00)〜AM/PM境界（#215: 0時開始ではない）",
			period:    "am",
			wantStart: time.Date(2026, 4, 20, 9, 0, 0, 0, config.JST),
			wantEnd:   time.Date(2026, 4, 20, 14, 0, 0, 0, config.JST),
		},
		{
			name:      "pm: 境界〜PM締め終了",
			period:    "pm",
			wantStart: time.Date(2026, 4, 20, 14, 0, 0, 0, config.JST),
			wantEnd:   time.Date(2026, 4, 20, 18, 30, 0, 0, config.JST),
		},
		{
			name:      "emg: PM締め終了〜翌日am_start（#215 越日レンジ・PM終了と連続）",
			period:    "emg",
			wantStart: time.Date(2026, 4, 20, 18, 30, 0, 0, config.JST),
			wantEnd:   time.Date(2026, 4, 21, 9, 0, 0, 0, config.JST),
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

// TestResolvePeriodRange_CrossMidnightEMG は越日 EMG の帰属を検証する（#215）。
// 深夜 0:00〜am_start の緊急会計は「前日の EMG」に、am_start 以降は「当日の AM」に計上される。
func TestResolvePeriodRange_CrossMidnightEMG(t *testing.T) {
	schedule := &sharedkernel.DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30", AmStart: "09:00"}
	day := time.Date(2026, 4, 20, 0, 0, 0, 0, config.JST)

	// 集計クエリは completed_at >= start AND < end の終端排他（GetCloseAggregate）
	contains := func(start, end, at time.Time) bool {
		return !at.Before(start) && at.Before(end)
	}

	emgStart, emgEnd, err := resolvePeriodRange(day, "emg", schedule)
	assert.NoError(t, err)
	nextAmStart, nextAmEnd, err := resolvePeriodRange(day.AddDate(0, 0, 1), "am", schedule)
	assert.NoError(t, err)

	t.Run("当日23:00の会計は当日EMGに計上される", func(t *testing.T) {
		at := time.Date(2026, 4, 20, 23, 0, 0, 0, config.JST)
		assert.True(t, contains(emgStart, emgEnd, at))
	})

	t.Run("翌朝7:00の会計は前日EMGに計上される（翌日AMに流入しない）", func(t *testing.T) {
		at := time.Date(2026, 4, 21, 7, 0, 0, 0, config.JST)
		assert.True(t, contains(emgStart, emgEnd, at), "前日EMGレンジに含まれる")
		assert.False(t, contains(nextAmStart, nextAmEnd, at), "翌日AMレンジには含まれない")
	})

	t.Run("翌朝9:00ちょうどは翌日AM（EMG終端は排他）", func(t *testing.T) {
		at := time.Date(2026, 4, 21, 9, 0, 0, 0, config.JST)
		assert.False(t, contains(emgStart, emgEnd, at))
		assert.True(t, contains(nextAmStart, nextAmEnd, at))
	})

	t.Run("前日EMGと当日AMの間に隙間・重複が無い", func(t *testing.T) {
		assert.True(t, emgEnd.Equal(nextAmStart))
	})

	t.Run("AmStart未設定は既定9:00にフォールバックする（後方互換）", func(t *testing.T) {
		legacy := &sharedkernel.DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30"}
		s, e, err := resolvePeriodRange(day, "emg", legacy)
		assert.NoError(t, err)
		assert.True(t, s.Equal(emgStart), "start が AmStart 指定時と一致")
		assert.True(t, e.Equal(emgEnd), "end が AmStart 指定時と一致")
	})

	t.Run("am_start >= boundary は設定不正としてエラー", func(t *testing.T) {
		bad := &sharedkernel.DaySchedule{AmPmBoundary: "09:00", PmEnd: "18:30", AmStart: "09:00"}
		_, _, err := resolvePeriodRange(day, "am", bad)
		assert.Error(t, err)
	})

	t.Run("am_start の形式不正はエラー", func(t *testing.T) {
		bad := &sharedkernel.DaySchedule{AmPmBoundary: "14:00", PmEnd: "18:30", AmStart: "9時"}
		_, _, err := resolvePeriodRange(day, "am", bad)
		assert.Error(t, err)
	})

	t.Run("pm_end の形式不正はエラー", func(t *testing.T) {
		bad := &sharedkernel.DaySchedule{AmPmBoundary: "14:00", PmEnd: "invalid", AmStart: "09:00"}
		_, _, err := resolvePeriodRange(day, "pm", bad)
		assert.Error(t, err)
	})
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
		rows         []PaymentAggregateRow
		totalRefund  int64
		cashMethodID uint64
		want         int64
	}{
		{
			name: "現金マスタ id の行を現金として加算する",
			rows: []PaymentAggregateRow{
				{PaymentMethodID: &cashID, Amount: 5000},
				{PaymentMethodID: &creditID, Amount: 3000},
			},
			cashMethodID: cashID,
			want:         5000,
		},
		{
			name: "payment_method_id=NULL の行も現金として加算する（レガシー seed 後方互換, bug.md H-3）",
			rows: []PaymentAggregateRow{
				{PaymentMethodID: nil, Amount: 4000},
				{PaymentMethodID: &creditID, Amount: 1000},
			},
			cashMethodID: cashID,
			want:         4000,
		},
		{
			name: "現金マスタ id と NULL が混在しても両方加算する",
			rows: []PaymentAggregateRow{
				{PaymentMethodID: &cashID, Amount: 5000},
				{PaymentMethodID: nil, Amount: 2000},
				{PaymentMethodID: &creditID, Amount: 3000},
			},
			cashMethodID: cashID,
			want:         7000,
		},
		{
			name: "現金マスタ id がどの行にも一致しなくても NULL 行は現金として加算する（NULL 後方互換は id 一致に非依存）",
			rows: []PaymentAggregateRow{
				{PaymentMethodID: nil, Amount: 4000},
				{PaymentMethodID: &cashID, Amount: 5000},
			},
			cashMethodID: unknownID,
			want:         4000,
		},
		{
			name: "非現金（credit_card）のみは現金 0",
			rows: []PaymentAggregateRow{
				{PaymentMethodID: &creditID, Amount: 3000},
			},
			cashMethodID: cashID,
			want:         0,
		},
		{
			name: "返金合計を理論現金から差し引く",
			rows: []PaymentAggregateRow{
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

// TestCashRegisterService_UnclassifiedOtherCountSnapshot DEC-40:
// preview に件数を載せ、Close の category_breakdown snapshot にポインタで保存する。
func TestCashRegisterService_UnclassifiedOtherCountSnapshot(t *testing.T) {
	targetDate := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	targetDateStr := "2026-04-20"
	const wantCount int64 = 2

	cashID := uint64(1)
	var created *model.CashRegisterClose
	closeRepo := &mockCashRegisterCloseRepository{
		findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, _ string) (*model.CashRegisterClose, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, c *model.CashRegisterClose) error {
			created = c
			return nil
		},
	}
	aggResult := &CloseAggregateResult{
		PaymentRows: []PaymentAggregateRow{
			{PaymentMethodID: &cashID, Amount: 1000},
		},
		CategoryRows: []CategoryAggregateRow{
			{Category: string(model.ItemCategoryOther), Amount: 1000},
		},
		TotalRefund:            0,
		BillingDetails:         []CloseBillingDetailRow{},
		TaxBreakdown:           []TaxBreakdownRow{},
		UnclassifiedOtherCount: wantCount,
	}
	accountingRepo := &mockAccountingRepository{
		getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
			return aggResult, nil
		},
		getCategoryPaymentAllocationDataFn: func(_ context.Context, _ uint64, _, _ time.Time) (*CategoryPaymentAllocationData, error) {
			return allocationDataFromAggregate(aggResult), nil
		},
	}
	closingsSvc := &mockClosingSettingsService{
		resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
			return defaultSchedule(), nil
		},
	}
	payMethodRepo := &mockPaymentMethodMasterRepository{
		findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{
				{ID: 1, Name: "現金", SystemKey: ptrString("cash"), IsActive: true},
			}, nil
		},
	}
	svc := newCashRegisterService(closeRepo, accountingRepo, closingsSvc, payMethodRepo)

	preview, err := svc.GetPreview(context.Background(), 1, targetDateStr, "am")
	require.NoError(t, err)
	assert.Equal(t, wantCount, preview.Aggregate.UnclassifiedOtherCount)

	got, err := svc.Close(context.Background(), 1, CloseRegisterInput{
		Date: targetDate, Period: "am", ActualCash: 1000,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, created)

	var schema model.CategoryBreakdownSchema
	require.NoError(t, json.Unmarshal(created.CategoryBreakdown, &schema))
	require.NotNil(t, schema.UnclassifiedOtherCount, "new snapshot must record the count (including via pointer)")
	assert.Equal(t, wantCount, *schema.UnclassifiedOtherCount)

	// 旧 snapshot（フィールド欠落）は nil のまま（FE は「記録なし」）
	var oldSchema model.CategoryBreakdownSchema
	require.NoError(t, json.Unmarshal([]byte(`{"categories":{"other":{"cash":1000}},"tax_breakdown":{"standard":{"taxable_amount":0,"tax_amount":0},"reduced":{"taxable_amount":0,"tax_amount":0}}}`), &oldSchema))
	assert.Nil(t, oldSchema.UnclassifiedOtherCount)
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
	aggResult := &CloseAggregateResult{
		PaymentRows: []PaymentAggregateRow{
			{PaymentMethodID: &cashID, Amount: cashBilling},     // 現金
			{PaymentMethodID: &creditID, Amount: creditBilling}, // クレジット
		},
		CategoryRows: []CategoryAggregateRow{
			{Category: string(model.ItemCategoryExamination), Amount: medicalRecordTotal},
		},
		TotalRefund:    0,
		BillingDetails: []CloseBillingDetailRow{},
		TaxBreakdown:   []TaxBreakdownRow{},
	}
	accountingRepo := &mockAccountingRepository{
		getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
			return aggResult, nil
		},
		getCategoryPaymentAllocationDataFn: func(_ context.Context, _ uint64, _, _ time.Time) (*CategoryPaymentAllocationData, error) {
			return allocationDataFromAggregate(aggResult), nil
		},
	}
	closingsSvc := &mockClosingSettingsService{
		resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
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

	diag := schema.Categories[string(model.ItemCategoryExamination)]
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

// TestCashRegisterService_List は締め履歴一覧取得を検証する。
func TestCashRegisterService_List(t *testing.T) {
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		findAllFn func(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error)
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "正常: 締め履歴一覧を返す",
			findAllFn: func(_ context.Context, _ uint64, _, _ *time.Time, _, _ int) ([]model.CashRegisterClose, int64, error) {
				return []model.CashRegisterClose{{ID: 1}, {ID: 2}}, 2, nil
			},
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name: "正常: 該当なし → 空リスト",
			findAllFn: func(_ context.Context, _ uint64, _, _ *time.Time, _, _ int) ([]model.CashRegisterClose, int64, error) {
				return []model.CashRegisterClose{}, 0, nil
			},
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name: "エラー: リポジトリエラーを wrap して返す",
			findAllFn: func(_ context.Context, _ uint64, _, _ *time.Time, _, _ int) ([]model.CashRegisterClose, int64, error) {
				return nil, 0, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closeRepo := &mockCashRegisterCloseRepository{findAllFn: tt.findAllFn}
			svc := newCashRegisterService(closeRepo, &mockAccountingRepository{}, &mockClosingSettingsService{}, &mockPaymentMethodMasterRepository{})

			got, total, err := svc.List(context.Background(), 1, &start, &end, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Zero(t, total)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

// TestCashRegisterService_GetByID は締め詳細の単体取得を検証する。
func TestCashRegisterService_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error)
		wantErr    bool
	}{
		{
			name: "正常: 締め記録を返す",
			findByIDFn: func(_ context.Context, _, id uint64) (*model.CashRegisterClose, error) {
				return &model.CashRegisterClose{ID: id}, nil
			},
		},
		{
			name: "エラー: リポジトリがエラーを返す",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.CashRegisterClose, error) {
				return nil, apperrors.WrapNotFound("cash_register_close", "999")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closeRepo := &mockCashRegisterCloseRepository{findByIDFn: tt.findByIDFn}
			svc := newCashRegisterService(closeRepo, &mockAccountingRepository{}, &mockClosingSettingsService{}, &mockPaymentMethodMasterRepository{})

			got, err := svc.GetByID(context.Background(), 1, 5)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, uint64(5), got.ID)
		})
	}
}

// TestCashRegisterService_GetPreview_ClinicRepoError は税率解決のための病院マスタ取得失敗が
// fetchAggregate 経由で GetPreview のエラーとして伝播することを検証する（M-7 #191）。
func TestCashRegisterService_GetPreview_ClinicRepoError(t *testing.T) {
	closeRepo := &mockCashRegisterCloseRepository{}
	accountingRepo := &mockAccountingRepository{
		getCloseAggregateFn: func(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
			return emptyAggregateResult(), nil
		},
	}
	closingsSvc := &mockClosingSettingsService{
		resolveScheduleFn: func(_ context.Context, _ uint64, _ time.Time) (*sharedkernel.DaySchedule, error) {
			return defaultSchedule(), nil
		},
	}
	payMethodRepo := &mockPaymentMethodMasterRepository{
		findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{{ID: 1, Name: "現金", SystemKey: ptrString("cash")}}, nil
		},
	}
	clinicRepo := &mockClinicRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
			return nil, errors.New("clinic lookup error")
		},
	}
	svc := NewCashRegisterService(closeRepo, accountingRepo, closingsSvc, payMethodRepo, clinicRepo)

	got, err := svc.GetPreview(context.Background(), 1, "2026-04-20", "am")

	assert.Error(t, err)
	assert.Nil(t, got)
}

// TestBuildCategoryBreakdown は #247 per-billing 配賦マトリクス → snapshot の各分岐を検証する。
func TestBuildCategoryBreakdown(t *testing.T) {
	rates := accountingReportTaxRates{StandardPercent: 10, ReducedPercent: 8}

	t.Run("SystemKey が nil のマスタは method_N にフォールバックする", func(t *testing.T) {
		matrix := map[string]map[string]int64{
			string(model.ItemCategoryExamination): {"method_5": 1000},
		}
		got := buildCategoryBreakdown(matrix, nil, rates, 0)
		assert.Equal(t, int64(1000), got.Categories[string(model.ItemCategoryExamination)]["method_5"])
	})

	t.Run("payRows の PaymentMethodID がマスタに存在しない場合も method_N にフォールバックする", func(t *testing.T) {
		matrix := map[string]map[string]int64{
			string(model.ItemCategoryTest): {"method_99": 500},
		}
		got := buildCategoryBreakdown(matrix, nil, rates, 0)
		assert.Equal(t, int64(500), got.Categories[string(model.ItemCategoryTest)]["method_99"])
	})

	t.Run("PaymentMethodID=nil の行は現金 system_key に計上される（#128 後方互換・レガシー seed 対応）", func(t *testing.T) {
		matrix := map[string]map[string]int64{
			string(model.ItemCategoryExamination): {"cash": 1000},
		}
		got := buildCategoryBreakdown(matrix, nil, rates, 0)
		assert.Equal(t, int64(1000), got.Categories[string(model.ItemCategoryExamination)]["cash"], "PaymentMethodID=nil の支払方法は現金として按分される")
	})

	// P2-13 / #247: 配賦マトリクス合計が支払実額と一致する（conservation）。
	t.Run("NULL 行を含む混在支払いで category_breakdown の合計が totalPayment と一致する", func(t *testing.T) {
		const totalPayment = int64(1000 + 2000)
		matrix := map[string]map[string]int64{
			string(model.ItemCategoryExamination): {
				"cash":        1000,
				"credit_card": 2000,
			},
		}
		got := buildCategoryBreakdown(matrix, nil, rates, 0)

		assert.Equal(t, int64(1000), got.Categories[string(model.ItemCategoryExamination)]["cash"], "NULL 行は現金に計上される")
		assert.Equal(t, int64(2000), got.Categories[string(model.ItemCategoryExamination)]["credit_card"], "非 NULL 行は対応する system_key に計上される")

		var sum int64
		for _, byMethod := range got.Categories {
			for _, amount := range byMethod {
				sum += amount
			}
		}
		assert.Equal(t, totalPayment, sum, "category_breakdown の合計は totalPayment と一致しなければならない")
	})

	t.Run("空マトリクスの場合はカテゴリが無い", func(t *testing.T) {
		got := buildCategoryBreakdown(nil, nil, rates, 0)
		assert.Empty(t, got.Categories)
	})

	t.Run("税率内訳を標準/軽減に分類する", func(t *testing.T) {
		taxRows := []TaxBreakdownRow{
			{TaxRate: 10, TaxableAmount: 1000, TaxAmount: 100},
			{TaxRate: 8, TaxableAmount: 500, TaxAmount: 40},
		}
		got := buildCategoryBreakdown(nil, taxRows, rates, 0)
		assert.Equal(t, int64(1000), got.TaxBreakdown.Standard.TaxableAmount)
		assert.Equal(t, int64(100), got.TaxBreakdown.Standard.TaxAmount)
		assert.Equal(t, int64(500), got.TaxBreakdown.Reduced.TaxableAmount)
		assert.Equal(t, int64(40), got.TaxBreakdown.Reduced.TaxAmount)
	})

	// DEC-40: snapshot に会計 distinct 件数を保存し、0 件でも欠落と区別する。
	t.Run("UnclassifiedOtherCount を snapshot にポインタで保存する", func(t *testing.T) {
		got := buildCategoryBreakdown(nil, nil, rates, 3)
		require.NotNil(t, got.UnclassifiedOtherCount)
		assert.Equal(t, int64(3), *got.UnclassifiedOtherCount)

		gotZero := buildCategoryBreakdown(nil, nil, rates, 0)
		require.NotNil(t, gotZero.UnclassifiedOtherCount, "0 件でもポインタを立てて欠落と区別する")
		assert.Equal(t, int64(0), *gotZero.UnclassifiedOtherCount)
	})
}

// TestParseHHMM は "HH:MM" / "HH:MM:SS"（PostgreSQL time 型）形式のパースを検証する。
func TestParseHHMM(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantH   int
		wantM   int
		wantErr bool
	}{
		{name: "HH:MM 形式を正しくパースする", input: "09:05", wantH: 9, wantM: 5},
		{name: "PostgreSQL time 型 HH:MM:SS 形式は秒部分を除去する", input: "18:30:00", wantH: 18, wantM: 30},
		{name: "空文字はエラー", input: "", wantErr: true},
		{name: "5文字だがコロン位置が不正な場合はエラー", input: "09-05", wantErr: true},
		{name: "4文字（ゼロ埋めなし）はエラー", input: "9:05", wantErr: true},
		{name: "数値でない HH:MM は Sscanf 失敗でエラー", input: "ab:cd", wantErr: true},
		{name: "8文字だが PostgreSQL 形式でない場合はエラー", input: "12345678", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, m, err := sharedkernel.ParseHHMM(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantH, h)
			assert.Equal(t, tt.wantM, m)
		})
	}
}
