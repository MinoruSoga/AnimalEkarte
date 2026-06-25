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
	getMonthlyReportFn         func(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error)
	getMonthlyReportByPeriodFn func(ctx context.Context, clinicID uint64, start, end time.Time) (*repository.MonthlyReportResult, error)
}

func (m *mockAccountingRepositoryForReport) FindAll(_ context.Context, _ uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForReport) FindByID(_ context.Context, _, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForReport) FindAllForClinics(_ context.Context, _ []uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForReport) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForReport) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForReport) Create(_ context.Context, _ uint64, _ *model.Billing) error {
	return nil
}
func (m *mockAccountingRepositoryForReport) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepositoryForReport) SavePayment(_ context.Context, _ *model.Payment) error {
	return nil
}

func (m *mockAccountingRepositoryForReport) SavePaymentSplits(_ context.Context, _ []model.PaymentSplit) error {
	return nil
}

func (m *mockAccountingRepositoryForReport) CompleteAccountingAppointments(_ context.Context, _ uint64, _, _, _ *uint64, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepositoryForReport) FindUnpaidByBilling(_ context.Context, _ uint64, _, _ string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepositoryForReport) FindUnpaidByOwner(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	return nil, 0, repository.UnpaidSummary{}, nil
}
func (m *mockAccountingRepositoryForReport) SumUnpaidByOwner(_ context.Context, _, _ uint64) (repository.OwnerUnpaidBalance, error) {
	return repository.OwnerUnpaidBalance{}, nil
}
func (m *mockAccountingRepositoryForReport) FindMonthlyUnpaidCarryover(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error) {
	return nil, 0, repository.MonthlyUnpaidSummary{}, nil
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
	return &repository.MonthlyReportResult{}, nil
}
func (m *mockAccountingRepositoryForReport) GetMonthlyReportByPeriod(ctx context.Context, clinicID uint64, start, end time.Time) (*repository.MonthlyReportResult, error) {
	if m.getMonthlyReportByPeriodFn != nil {
		return m.getMonthlyReportByPeriodFn(ctx, clinicID, start, end)
	}
	return &repository.MonthlyReportResult{}, nil
}

func (m *mockAccountingRepositoryForReport) SumPaidByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepositoryForReport) MaxSingleVisitAmountByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepositoryForReport) FindOwnersByAnnualRevenue(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
	return nil, nil
}

// ---- ヘルパー ----

func newAccountingReportService(
	repo *mockAccountingRepositoryForReport,
	payMethodRepo *mockPaymentMethodMasterRepository,
	holidayRepo *mockClinicHolidayRepository,
) AccountingReportService {
	return NewAccountingReportService(repo, payMethodRepo, holidayRepo, &mockClinicRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
			return &model.Clinic{ID: id, StandardTaxRate: 0.10, ReducedTaxRate: 0.08}, nil
		},
	})
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
					PaymentRows: []repository.MonthlyPaymentRow{
						{Date: "2026-04-01", PaymentMethodID: nil, Amount: 10000},
					},
					CategoryRows: []repository.MonthlyCategoryRow{
						{Date: "2026-04-01", Category: "診察", Amount: 10000},
					},
					DailyBillingCount: map[string]int64{"2026-04-01": 2},
					ClosedAM:          map[string]bool{"2026-04-01": true},
					ClosedPM:          map[string]bool{},
					GrandTotal:        10000,
					BillingCount:      2,
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
				return &repository.MonthlyReportResult{}, nil
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
				return &repository.MonthlyReportResult{}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "エラー: FindAllByYearMonth（休診日）がエラーを返す",
			year:  2026,
			month: 4,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return &repository.MonthlyReportResult{}, nil
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
			name:  "正常: 3種混在支払い → ByPaymentMethod が支払方法別に正しく集計される",
			year:  2026,
			month: 5,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return &repository.MonthlyReportResult{
					PaymentRows: []repository.MonthlyPaymentRow{
						{Date: "2026-05-01", PaymentMethodID: nil, Amount: 5000},
						{Date: "2026-05-01", PaymentMethodID: ptrUint64(1), Amount: 3000},
						{Date: "2026-05-02", PaymentMethodID: ptrUint64(2), Amount: 2000},
					},
					GrandTotal:   10000,
					BillingCount: 3,
				}, nil
			},
			findAllPayMethodFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{
					{ID: 1, Name: "クレジット"},
					{ID: 2, Name: "電子マネー"},
				}, nil
			},
			findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			},
			checkResult: func(t *testing.T, got *MonthlyReportResponse) {
				assert.Equal(t, int64(5000), got.Summary.ByPaymentMethod["現金"])
				assert.Equal(t, int64(3000), got.Summary.ByPaymentMethod["クレジット"])
				assert.Equal(t, int64(2000), got.Summary.ByPaymentMethod["電子マネー"])
				assert.Equal(t, int64(10000), got.Summary.NetAmount)
				assert.Equal(t, int64(3), got.Summary.TotalBillings)
			},
		},
		{
			name:  "正常: 軽減税率（8%）データ → Reduced に集計される",
			year:  2026,
			month: 4,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
				return &repository.MonthlyReportResult{
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

func TestAccountingReportService_GetMonthly_UsesClinicTaxRates(t *testing.T) {
	repo := &mockAccountingRepositoryForReport{
		getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
			return &repository.MonthlyReportResult{
				TaxBreakdown: []repository.TaxBreakdownRow{
					{TaxRate: 12, TaxableAmount: 12000, TaxAmount: 1440},
					{TaxRate: 10, TaxableAmount: 10000, TaxAmount: 1000},
				},
			}, nil
		},
	}
	svc := NewAccountingReportService(
		repo,
		&mockPaymentMethodMasterRepository{findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{}, nil
		}},
		&mockClinicHolidayRepository{findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
			return []model.ClinicHoliday{}, nil
		}},
		&mockClinicRepository{findByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
			return &model.Clinic{ID: id, StandardTaxRate: 0.12, ReducedTaxRate: 0.10}, nil
		}},
	)

	got, err := svc.GetMonthly(context.Background(), 1, 2026, 4)

	assert.NoError(t, err)
	assert.Equal(t, int64(12000), got.Summary.TaxBreakdown.Standard.TaxableAmount)
	assert.Equal(t, int64(1440), got.Summary.TaxBreakdown.Standard.TaxAmount)
	assert.Equal(t, int64(10000), got.Summary.TaxBreakdown.Reduced.TaxableAmount)
	assert.Equal(t, int64(1000), got.Summary.TaxBreakdown.Reduced.TaxAmount)
}

func TestAccountingReportService_GetMonthlyByPeriod(t *testing.T) {
	var gotStart time.Time
	var gotEnd time.Time
	repo := &mockAccountingRepositoryForReport{
		getMonthlyReportByPeriodFn: func(_ context.Context, _ uint64, start, end time.Time) (*repository.MonthlyReportResult, error) {
			gotStart = start
			gotEnd = end
			return &repository.MonthlyReportResult{
				DailyBillingCount: map[string]int64{"2026-04-30": 1, "2026-05-01": 2},
				GrandTotal:        3000,
				BillingCount:      3,
			}, nil
		},
	}
	svc := NewAccountingReportService(
		repo,
		&mockPaymentMethodMasterRepository{findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{}, nil
		}},
		&mockClinicHolidayRepository{findByYearMonthFn: func(_ context.Context, _ uint64, yearMonth string) ([]model.ClinicHoliday, error) {
			if yearMonth == "2026-05" {
				return []model.ClinicHoliday{{Date: time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)}}, nil
			}
			return []model.ClinicHoliday{}, nil
		}},
		&mockClinicRepository{findByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
			return &model.Clinic{ID: id, StandardTaxRate: 0.10, ReducedTaxRate: 0.08}, nil
		}},
	)
	start := time.Date(2026, 4, 30, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)

	got, err := svc.GetMonthlyByPeriod(context.Background(), 1, start, end)

	assert.NoError(t, err)
	assert.Equal(t, "2026-04-30", gotStart.Format("2006-01-02"))
	assert.Equal(t, "2026-05-02", gotEnd.Format("2006-01-02"))
	assert.Equal(t, "2026-04-30", got.StartDate)
	assert.Equal(t, "2026-05-01", got.EndDate)
	assert.Len(t, got.DailyDetails, 2)
	assert.False(t, got.DailyDetails[0].IsHoliday)
	assert.True(t, got.DailyDetails[1].IsHoliday)
}

// TestValidateReportPeriod は #192 期間集計のガード契約を固定化する。
// start>end の拒否と 367 日上限（無制限スキャン防止）が回帰しないことを保証する。
// 366 日差はちょうど境界で許可、367 日差で拒否となる。
func TestValidateReportPeriod(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, jst)
	cases := []struct {
		name      string
		start     time.Time
		end       time.Time
		wantError bool
	}{
		{"同日は許可", base, base, false},
		{"通常の昇順期間は許可", base, base.AddDate(0, 0, 30), false},
		{"上限ちょうど(366日差)は許可", base, base.AddDate(0, 0, 366), false},
		{"上限超過(367日差)は拒否", base, base.AddDate(0, 0, 367), true},
		{"start が end より後は拒否", base.AddDate(0, 0, 1), base, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReportPeriod(tc.start, tc.end)
			if tc.wantError {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err), "期間バリデーションは invalid input 種別であるべき")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
