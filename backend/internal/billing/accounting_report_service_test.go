package billing

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ヘルパー ----

func newAccountingReportService(
	repo *mockAccountingRepository,
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
		getMonthlyReportFn func(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResult, error)
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
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
				return &MonthlyReportResult{
					PaymentRows: []MonthlyPaymentRow{
						{Date: "2026-04-01", PaymentMethodID: nil, Amount: 10000},
					},
					CategoryRows: []MonthlyCategoryRow{
						{Date: "2026-04-01", Category: string(model.ItemCategoryExamination), Amount: 10000},
					},
					DailyBillingCount: map[string]int64{"2026-04-01": 2},
					ClosedAM:          map[string]bool{"2026-04-01": true},
					ClosedPM:          map[string]bool{},
					GrandTotal:        10000,
					BillingCount:      2,
					TaxBreakdown: []TaxBreakdownRow{
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
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
				return &MonthlyReportResult{}, nil
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
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "エラー: FindAll（支払方法）がエラーを返す",
			year:  2026,
			month: 4,
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
				return &MonthlyReportResult{}, nil
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
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
				return &MonthlyReportResult{}, nil
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
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
				return &MonthlyReportResult{
					PaymentRows: []MonthlyPaymentRow{
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
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
				return &MonthlyReportResult{
					TaxBreakdown: []TaxBreakdownRow{
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
			repo := &mockAccountingRepository{
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
	repo := &mockAccountingRepository{
		getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
			return &MonthlyReportResult{
				TaxBreakdown: []TaxBreakdownRow{
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
	repo := &mockAccountingRepository{
		getMonthlyReportByPeriodFn: func(_ context.Context, _ uint64, start, end time.Time) (*MonthlyReportResult, error) {
			gotStart = start
			gotEnd = end
			return &MonthlyReportResult{
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
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, config.JST)
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

// TestAccountingReportService_GetMonthly_ClinicRepoError は buildReportResponse 内の
// clinicRepo.FindByID エラー分岐（税率取得失敗）を固定化する。
func TestAccountingReportService_GetMonthly_ClinicRepoError(t *testing.T) {
	repo := &mockAccountingRepository{}
	svc := NewAccountingReportService(
		repo,
		&mockPaymentMethodMasterRepository{findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{}, nil
		}},
		&mockClinicHolidayRepository{findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
			return []model.ClinicHoliday{}, nil
		}},
		&mockClinicRepository{findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
			return nil, errors.New("clinic not found")
		}},
	)

	got, err := svc.GetMonthly(context.Background(), 1, 2026, 4)

	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestAccountingReportService_GetMonthlyByPeriod_ValidationError(t *testing.T) {
	svc := newAccountingReportService(
		&mockAccountingRepository{},
		&mockPaymentMethodMasterRepository{},
		&mockClinicHolidayRepository{},
	)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, config.JST)
	end := time.Date(2026, 4, 1, 0, 0, 0, 0, config.JST) // start > end

	got, err := svc.GetMonthlyByPeriod(context.Background(), 1, start, end)

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, got)
}

func TestAccountingReportService_GetMonthlyByPeriod_RepoError(t *testing.T) {
	repo := &mockAccountingRepository{
		getMonthlyReportByPeriodFn: func(_ context.Context, _ uint64, _, _ time.Time) (*MonthlyReportResult, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newAccountingReportService(
		repo,
		&mockPaymentMethodMasterRepository{findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
			return []model.PaymentMethodMaster{}, nil
		}},
		&mockClinicHolidayRepository{findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
			return []model.ClinicHoliday{}, nil
		}},
	)
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, config.JST)
	end := time.Date(2026, 4, 2, 0, 0, 0, 0, config.JST)

	got, err := svc.GetMonthlyByPeriod(context.Background(), 1, start, end)

	assert.Error(t, err)
	assert.Nil(t, got)
}

// TestResolvePaymentMethodName は支払方法名解決の全分岐（nil/既知ID/未知ID）を固定化する。
func TestResolvePaymentMethodName(t *testing.T) {
	names := map[uint64]string{1: "クレジット"}
	tests := []struct {
		name string
		id   *uint64
		want string
	}{
		{name: "nil id は現金として扱う", id: nil, want: "現金"},
		{name: "既知の id はマスタ名を返す", id: ptrUint64(1), want: "クレジット"},
		{name: "未知の id はフォールバック表記を返す", id: ptrUint64(99), want: "支払方法(99)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolvePaymentMethodName(tt.id, names)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBuildMonthlyCSV は CSV バイト列の構造（BOM・ヘッダ・日別明細・合計行）を固定化する。
func TestBuildMonthlyCSV(t *testing.T) {
	result := &MonthlyReportResponse{
		Year:      2026,
		Month:     4,
		StartDate: "2026-04-01",
		EndDate:   "2026-04-02",
		Summary: MonthlyReportSummary{
			NetAmount:   3000,
			TotalRefund: 100,
			TaxBreakdown: TaxBreakdownSummary{
				Standard: TaxBreakdownEntry{TaxableAmount: 1000, TaxAmount: 100},
				Reduced:  TaxBreakdownEntry{TaxableAmount: 500, TaxAmount: 40},
			},
		},
		DailyDetails: []DailyReportDetail{
			{
				Date: "2026-04-01", Weekday: "水",
				AMCount: 1, AMNet: 1000, PMCount: 2, PMNet: 2000, DayNet: 3000,
				Refund: 0, AMClosed: true, PMClosed: false, IsHoliday: false,
			},
			{
				Date: "2026-04-02", Weekday: "木", IsHoliday: true,
			},
		},
	}

	got, err := buildMonthlyCSV(result, "test_report.csv")

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "test_report.csv", got.Filename)
	assert.True(t, bytes.HasPrefix(got.Data, []byte("\xEF\xBB\xBF")), "CSV は UTF-8 BOM で始まるべき")

	reader := csv.NewReader(bytes.NewReader(got.Data[3:]))
	records, err := reader.ReadAll()
	assert.NoError(t, err)
	// ヘッダ行 + 日別明細2行 + 合計行
	assert.Len(t, records, 4)
	assert.Equal(t, "日付", records[0][0])
	assert.Equal(t, "2026-04-01", records[1][0])
	assert.Equal(t, "済", records[1][8])  // AM締め
	assert.Equal(t, "未", records[1][9])  // PM締め
	assert.Equal(t, "休", records[2][10]) // 休診
	assert.Equal(t, "合計", records[3][0])
	assert.Equal(t, "3000", records[3][6])
	assert.Equal(t, "100", records[3][7])
}

func TestAccountingReportService_ExportMonthlyCSV(t *testing.T) {
	t.Run("正常: CSV を出力する", func(t *testing.T) {
		repo := &mockAccountingRepository{
			getMonthlyReportFn: func(_ context.Context, _ uint64, _, _ int) (*MonthlyReportResult, error) {
				return &MonthlyReportResult{GrandTotal: 1000, BillingCount: 1}, nil
			},
		}
		svc := newAccountingReportService(
			repo,
			&mockPaymentMethodMasterRepository{findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{}, nil
			}},
			&mockClinicHolidayRepository{findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			}},
		)

		got, err := svc.ExportMonthlyCSV(context.Background(), 1, 2026, 4)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "monthly_report_202604.csv", got.Filename)
		assert.True(t, bytes.HasPrefix(got.Data, []byte("\xEF\xBB\xBF")))
	})

	t.Run("エラー: GetMonthly がエラーを返す", func(t *testing.T) {
		svc := newAccountingReportService(
			&mockAccountingRepository{},
			&mockPaymentMethodMasterRepository{},
			&mockClinicHolidayRepository{},
		)

		// month=0 は validateMonth で拒否される
		got, err := svc.ExportMonthlyCSV(context.Background(), 1, 2026, 0)

		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestAccountingReportService_ExportMonthlyCSVByPeriod(t *testing.T) {
	t.Run("正常: 期間指定で CSV を出力する", func(t *testing.T) {
		repo := &mockAccountingRepository{
			getMonthlyReportByPeriodFn: func(_ context.Context, _ uint64, _, _ time.Time) (*MonthlyReportResult, error) {
				return &MonthlyReportResult{GrandTotal: 500, BillingCount: 1}, nil
			},
		}
		svc := newAccountingReportService(
			repo,
			&mockPaymentMethodMasterRepository{findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return []model.PaymentMethodMaster{}, nil
			}},
			&mockClinicHolidayRepository{findByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			}},
		)
		start := time.Date(2026, 4, 1, 0, 0, 0, 0, config.JST)
		end := time.Date(2026, 4, 2, 0, 0, 0, 0, config.JST)

		got, err := svc.ExportMonthlyCSVByPeriod(context.Background(), 1, start, end)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "monthly_report_20260401_20260402.csv", got.Filename)
	})

	t.Run("エラー: GetMonthlyByPeriod がエラーを返す", func(t *testing.T) {
		svc := newAccountingReportService(
			&mockAccountingRepository{},
			&mockPaymentMethodMasterRepository{},
			&mockClinicHolidayRepository{},
		)
		start := time.Date(2026, 5, 1, 0, 0, 0, 0, config.JST)
		end := time.Date(2026, 4, 1, 0, 0, 0, 0, config.JST) // start > end で validateReportPeriod が拒否

		got, err := svc.ExportMonthlyCSVByPeriod(context.Background(), 1, start, end)

		assert.Error(t, err)
		assert.Nil(t, got)
	})
}
