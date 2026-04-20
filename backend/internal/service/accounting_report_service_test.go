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

// mockAccountingRepositoryForReport は AccountingReportService 専用のモック実装。
// accounting_service_test.go の mockAccountingRepository とは別型として定義し、
// GetMonthlyReport の関数フィールドを制御可能にする。
type mockAccountingRepositoryForReport struct {
	getMonthlyReportFn func(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error)
}

func (m *mockAccountingRepositoryForReport) FindAll(_ context.Context, _ uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForReport) FindByID(_ context.Context, _, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForReport) Create(_ context.Context, _ uint64, _ *model.Billing) error {
	return nil
}
func (m *mockAccountingRepositoryForReport) UpdateFields(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForReport) UpsertPayment(_ context.Context, _ *model.Payment) error {
	return nil
}
func (m *mockAccountingRepositoryForReport) FindUnpaidByBilling(_ context.Context, _ uint64, _ string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForReport) FindUnpaidByOwner(_ context.Context, _ uint64, _ string, _, _ int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	return nil, 0, repository.UnpaidSummary{}, nil
}
func (m *mockAccountingRepositoryForReport) GetDailySummary(_ context.Context, _ uint64, _ time.Time) (*repository.DailySummaryResult, error) {
	return &repository.DailySummaryResult{}, nil
}
func (m *mockAccountingRepositoryForReport) GetCloseAggregate(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
	return &repository.CloseAggregateResult{}, nil
}
func (m *mockAccountingRepositoryForReport) GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error) {
	if m.getMonthlyReportFn != nil {
		return m.getMonthlyReportFn(ctx, clinicID, year, month)
	}
	return &repository.MonthlyReportResult{Rows: []repository.MonthlyReportRow{}}, nil
}

// ---- ヘルパー ----

func newAccountingReportService(
	repo *mockAccountingRepositoryForReport,
	payMethodRepo *mockPaymentMethodMasterRepository,
	holidayRepo *mockClinicHolidayRepository,
) AccountingReportService {
	return NewAccountingReportService(repo, payMethodRepo, holidayRepo)
}

// ---- テスト ----

func TestAccountingReportService_GetMonthly(t *testing.T) {
	tests := []struct {
		name               string
		year               int
		month              int
		getMonthlyReportFn func(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error)
		findAllPayMethodFn func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
		findByYearMonthFn  func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
		wantErr            bool
		wantErrIs          error
		checkResult        func(t *testing.T, got *MonthlyReportResponse)
	}{
		{
			name:      "エラー: month=0 → ErrInvalidInput",
			year:      2026,
			month:     0,
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:      "エラー: month=13 → ErrInvalidInput",
			year:      2026,
			month:     13,
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:  "正常: モックデータで MonthlyReportResponse の shape を確認",
			year:  2026,
			month: 4,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return &repository.MonthlyReportResult{
					Rows: []repository.MonthlyReportRow{
						{
							Date:         "2026-04-01",
							Category:     "診察",
							NetAmount:    10000,
							BillingCount: 2,
							AMClosed:     true,
							PMClosed:     false,
						},
					},
					GrandTotal:   10000,
					BillingCount: 2,
					TaxBreakdown: []repository.TaxBreakdownRow{
						{TaxRate: 10, TaxableAmount: 9090, TaxAmount: 910},
					},
				}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "クレジット"},
				}, nil
			},
			findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			},
			checkResult: func(t *testing.T, got *MonthlyReportResponse) {
				assert.Equal(t, 2026, got.Year)
				assert.Equal(t, 4, got.Month)
				// 4月は30日
				assert.Len(t, got.DailyDetails, 30)
				// 最初の日（2026-04-01）の集計を確認
				assert.Equal(t, "2026-04-01", got.DailyDetails[0].Date)
				assert.True(t, got.DailyDetails[0].AMClosed)
				assert.False(t, got.DailyDetails[0].PMClosed)
				// サマリー確認
				assert.Equal(t, int64(10000), got.Summary.NetAmount)
				assert.Equal(t, int64(2), got.Summary.TotalBillings)
				// 税率別サマリー（10% → Standard）
				assert.Equal(t, int64(9090), got.Summary.TaxBreakdown.Standard.TaxableAmount)
				assert.Equal(t, int64(910), got.Summary.TaxBreakdown.Standard.TaxAmount)
			},
		},
		{
			name:  "正常: 空データ → 全日付分の DailyDetails を返す",
			year:  2026,
			month: 2,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return &repository.MonthlyReportResult{
					Rows:         []repository.MonthlyReportRow{},
					GrandTotal:   0,
					BillingCount: 0,
				}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{}, nil
			},
			findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			},
			checkResult: func(t *testing.T, got *MonthlyReportResponse) {
				assert.Equal(t, 2026, got.Year)
				assert.Equal(t, 2, got.Month)
				// 2026年2月は28日
				assert.Len(t, got.DailyDetails, 28)
				assert.Equal(t, int64(0), got.Summary.NetAmount)
			},
		},
		{
			name:  "エラー: GetMonthlyReport がエラーを返す",
			year:  2026,
			month: 4,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "エラー: FindAll（支払方法）がエラーを返す",
			year:  2026,
			month: 4,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return &repository.MonthlyReportResult{Rows: []repository.MonthlyReportRow{}}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "エラー: FindByYearMonth（休診日）がエラーを返す",
			year:  2026,
			month: 4,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return &repository.MonthlyReportResult{Rows: []repository.MonthlyReportRow{}}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{}, nil
			},
			findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "正常: 軽減税率（8%）データ → Reduced に集計される",
			year:  2026,
			month: 4,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return &repository.MonthlyReportResult{
					Rows: []repository.MonthlyReportRow{},
					TaxBreakdown: []repository.TaxBreakdownRow{
						{TaxRate: 8, TaxableAmount: 5000, TaxAmount: 400},
					},
				}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{}, nil
			},
			findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			},
			checkResult: func(t *testing.T, got *MonthlyReportResponse) {
				assert.Equal(t, int64(5000), got.Summary.TaxBreakdown.Reduced.TaxableAmount)
				assert.Equal(t, int64(400), got.Summary.TaxBreakdown.Reduced.TaxAmount)
				assert.Equal(t, int64(0), got.Summary.TaxBreakdown.Standard.TaxableAmount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockAccountingRepositoryForReport{
				getMonthlyReportFn: tt.getMonthlyReportFn,
			}
			payMethodRepo := &mockPaymentMethodMasterRepository{
				findAllFn: tt.findAllPayMethodFn,
			}
			holidayRepo := &mockClinicHolidayRepository{
				findByYearMonthFn: tt.findByYearMonthFn,
			}
			svc := newAccountingReportService(repo, payMethodRepo, holidayRepo)

			// Act
			got, err := svc.GetMonthly(context.Background(), 1, tt.year, tt.month)

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
