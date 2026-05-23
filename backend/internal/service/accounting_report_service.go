package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// jst は日本標準時（UTC+9）のタイムゾーン定数
var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// ---- レスポンス型（フロントエンド期待形式） ----

// MonthlyReportResponse は月次売上レポートのフロントエンド向けレスポンス
type MonthlyReportResponse struct {
	Year         int                  `json:"year"`
	Month        int                  `json:"month"`
	Summary      MonthlyReportSummary `json:"summary"`
	DailyDetails []DailyReportDetail  `json:"daily_details"`
}

// MonthlyReportSummary は月次サマリー情報
type MonthlyReportSummary struct {
	WorkingDays     int                 `json:"working_days"`
	TotalBillings   int64               `json:"total_billings"`
	TotalAmount     int64               `json:"total_amount"`
	TotalRefund     int64               `json:"total_refund"`
	NetAmount       int64               `json:"net_amount"`
	ByPaymentMethod map[string]int64    `json:"by_payment_method"`
	ByCategory      map[string]int64    `json:"by_category"`
	TaxBreakdown    TaxBreakdownSummary `json:"tax_breakdown"`
}

// TaxBreakdownSummary は標準・軽減税率別サマリー
type TaxBreakdownSummary struct {
	Standard TaxBreakdownEntry `json:"standard"`
	Reduced  TaxBreakdownEntry `json:"reduced"`
}

// TaxBreakdownEntry は税率別の課税対象額・税額
type TaxBreakdownEntry struct {
	TaxableAmount int64 `json:"taxable_amount"`
	TaxAmount     int64 `json:"tax_amount"`
}

// DailyReportDetail は日別明細
type DailyReportDetail struct {
	Date      string `json:"date"`    // YYYY-MM-DD
	Weekday   string `json:"weekday"` // "月","火","水","木","金","土","日"
	AMCount   int64  `json:"am_count"`
	AMNet     int64  `json:"am_net"`
	PMCount   int64  `json:"pm_count"`
	PMNet     int64  `json:"pm_net"`
	DayNet    int64  `json:"day_net"`
	Refund    int64  `json:"refund"`
	AMClosed  bool   `json:"am_closed"`
	PMClosed  bool   `json:"pm_closed"`
	IsHoliday bool   `json:"is_holiday"`
}

// MonthlyCSVResult は月次CSV出力の結果
type MonthlyCSVResult struct {
	Filename string
	Data     []byte
}

// buildMonthlyReportResponse は生データからフロントエンド向けレスポンスを構築する
func buildMonthlyReportResponse(
	year, month int,
	raw *repository.MonthlyReportResult,
	payMethodNames map[uint64]string,
	holidaySet map[string]bool,
) *MonthlyReportResponse {
	// 日付別の集計マップを構築
	type dailyAgg struct {
		amNet    int64
		pmNet    int64
		amCount  int64
		pmCount  int64
		refund   int64
		amClosed bool
		pmClosed bool
	}
	dailyMap := make(map[string]*dailyAgg)

	byPaymentMethod := make(map[string]int64)
	byCategory := make(map[string]int64)
	var totalAmount int64
	totalRefund := raw.TotalRefund

	// 支払方法別集計 + 日別 net 合算
	for _, row := range raw.PaymentRows {
		d, ok := dailyMap[row.Date]
		if !ok {
			d = &dailyAgg{
				amClosed: raw.ClosedAM[row.Date],
				pmClosed: raw.ClosedPM[row.Date],
			}
			dailyMap[row.Date] = d
		}
		d.pmNet += row.Amount
		totalAmount += row.Amount
		pmName := resolvePaymentMethodName(row.PaymentMethodID, payMethodNames)
		byPaymentMethod[pmName] += row.Amount
	}

	// 日別件数を DailyBillingCount から注入
	for date, count := range raw.DailyBillingCount {
		d, ok := dailyMap[date]
		if !ok {
			d = &dailyAgg{
				amClosed: raw.ClosedAM[date],
				pmClosed: raw.ClosedPM[date],
			}
			dailyMap[date] = d
		}
		d.pmCount = count
	}

	// カテゴリ別集計
	for _, row := range raw.CategoryRows {
		byCategory[row.Category] += row.Amount
	}

	// 締め状態を dailyMap に反映（PaymentRows が空の日付用）
	for date, closed := range raw.ClosedAM {
		if _, ok := dailyMap[date]; !ok {
			dailyMap[date] = &dailyAgg{}
		}
		if closed {
			dailyMap[date].amClosed = true
		}
	}
	for date, closed := range raw.ClosedPM {
		if _, ok := dailyMap[date]; !ok {
			dailyMap[date] = &dailyAgg{}
		}
		if closed {
			dailyMap[date].pmClosed = true
		}
	}

	// 税率別集計を TaxBreakdownSummary に変換
	var taxSummary TaxBreakdownSummary
	for _, tr := range raw.TaxBreakdown {
		if tr.TaxRate > 8 {
			// 標準税率（10%）
			taxSummary.Standard.TaxableAmount += tr.TaxableAmount
			taxSummary.Standard.TaxAmount += tr.TaxAmount
		} else {
			// 軽減税率（8%以下）
			taxSummary.Reduced.TaxableAmount += tr.TaxableAmount
			taxSummary.Reduced.TaxAmount += tr.TaxAmount
		}
	}

	// 日別明細スライスを日付昇順で構築
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, jst)
	daysInMonth := startDate.AddDate(0, 1, 0).AddDate(0, 0, -1).Day()

	weekdayJP := [7]string{"日", "月", "火", "水", "木", "金", "土"}
	dailyDetails := make([]DailyReportDetail, 0, daysInMonth)
	workingDays := 0

	for day := 1; day <= daysInMonth; day++ {
		d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, jst)
		dateStr := d.Format("2006-01-02")
		isHoliday := holidaySet[dateStr]

		agg := dailyMap[dateStr]
		detail := DailyReportDetail{
			Date:      dateStr,
			Weekday:   weekdayJP[d.Weekday()],
			IsHoliday: isHoliday,
		}
		if agg != nil {
			detail.AMCount = agg.amCount
			detail.AMNet = agg.amNet
			detail.PMCount = agg.pmCount
			detail.PMNet = agg.pmNet
			detail.DayNet = agg.amNet + agg.pmNet
			detail.Refund = agg.refund
			detail.AMClosed = agg.amClosed
			detail.PMClosed = agg.pmClosed
		}

		if !isHoliday {
			workingDays++
		}

		dailyDetails = append(dailyDetails, detail)
	}

	summary := MonthlyReportSummary{
		WorkingDays:     workingDays,
		TotalBillings:   raw.BillingCount,
		TotalAmount:     totalAmount,
		TotalRefund:     totalRefund,
		NetAmount:       raw.GrandTotal,
		ByPaymentMethod: byPaymentMethod,
		ByCategory:      byCategory,
		TaxBreakdown:    taxSummary,
	}

	return &MonthlyReportResponse{
		Year:         year,
		Month:        month,
		Summary:      summary,
		DailyDetails: dailyDetails,
	}
}

// buildTaxBreakdown は TaxBreakdownRow スライスを TaxBreakdownSummary に変換する
// cash_register_service.go からも利用するためパッケージスコープで定義する
func buildTaxBreakdown(rows []repository.TaxBreakdownRow) TaxBreakdownSummary {
	var summary TaxBreakdownSummary
	for _, tr := range rows {
		if tr.TaxRate > 8 {
			summary.Standard.TaxableAmount += tr.TaxableAmount
			summary.Standard.TaxAmount += tr.TaxAmount
		} else {
			summary.Reduced.TaxableAmount += tr.TaxableAmount
			summary.Reduced.TaxAmount += tr.TaxAmount
		}
	}
	return summary
}

// buildPayMethodNameMap は PaymentMethodMaster スライスから ID→名前マップを構築する
func buildPayMethodNameMap(methods []model.PaymentMethodMaster) map[uint64]string {
	m := make(map[uint64]string, len(methods))
	for i := range methods {
		m[methods[i].ID] = methods[i].Name
	}
	return m
}

// ---- インターフェース ----

// AccountingReportService は月次売上レポートのビジネスロジックインターフェース
type AccountingReportService interface {
	GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResponse, error)
	ExportMonthlyCSV(ctx context.Context, clinicID uint64, year, month int) (*MonthlyCSVResult, error)
}

// ---- サービス実装 ----

type accountingReportService struct {
	repo          repository.AccountingRepository
	payMethodRepo repository.PaymentMethodMasterRepository
	holidayRepo   repository.ClinicHolidayRepository
}

// NewAccountingReportService は AccountingReportService を初期化して返す
func NewAccountingReportService(
	repo repository.AccountingRepository,
	payMethodRepo repository.PaymentMethodMasterRepository,
	holidayRepo repository.ClinicHolidayRepository,
) AccountingReportService {
	return &accountingReportService{
		repo:          repo,
		payMethodRepo: payMethodRepo,
		holidayRepo:   holidayRepo,
	}
}

// validateMonth は month 値（1〜12）のバリデーションを行う
func validateMonth(month int) error {
	if month < 1 || month > 12 {
		return apperrors.WrapInvalidInput("month は 1〜12 の範囲で指定してください")
	}
	return nil
}

func (s *accountingReportService) GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResponse, error) {
	if err := validateMonth(month); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate month")
	}

	raw, err := s.repo.GetMonthlyReport(ctx, clinicID, year, month)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get monthly report", "error", err)
		return nil, apperrors.Wrap(err, "failed to get monthly report")
	}

	// 支払方法マスタを取得してID→名前マップを構築
	payMethods, err := s.payMethodRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get payment methods", "error", err)
		return nil, apperrors.Wrap(err, "failed to get payment methods")
	}
	payMethodNames := make(map[uint64]string, len(payMethods))
	for i := range payMethods {
		payMethodNames[payMethods[i].ID] = payMethods[i].Name
	}

	// 休診日マスタを取得
	yearMonth := fmt.Sprintf("%04d-%02d", year, month)
	holidays, err := s.holidayRepo.FindAllByYearMonth(ctx, clinicID, yearMonth)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic holidays", "error", err)
		return nil, apperrors.Wrap(err, "failed to get clinic holidays")
	}
	holidaySet := make(map[string]bool, len(holidays))
	for _, h := range holidays {
		holidaySet[h.Date.Format("2006-01-02")] = true
	}

	return buildMonthlyReportResponse(year, month, raw, payMethodNames, holidaySet), nil
}

// ExportMonthlyCSV は月次売上レポートの CSV バイト列とファイル名を返す
func (s *accountingReportService) ExportMonthlyCSV(ctx context.Context, clinicID uint64, year, month int) (*MonthlyCSVResult, error) {
	result, err := s.GetMonthly(ctx, clinicID, year, month)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get monthly accounting report", "error", err)
		return nil, apperrors.Wrap(err, "failed to get monthly accounting report")
	}

	var buf bytes.Buffer
	// BOM（Excel の UTF-8 認識用）
	buf.Write([]byte("\xEF\xBB\xBF"))

	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"日付", "曜日", "AM件数", "AM純売上(円)", "PM件数", "PM純売上(円)", "日計(円)", "返金(円)", "AM締め", "PM締め", "休診",
		"標準税率課税対象額(円)", "標準税率消費税額(円)", "軽減税率課税対象額(円)", "軽減税率消費税額(円)",
	})

	for _, d := range result.DailyDetails {
		amClosed := "未"
		if d.AMClosed {
			amClosed = "済"
		}
		pmClosed := "未"
		if d.PMClosed {
			pmClosed = "済"
		}
		holiday := ""
		if d.IsHoliday {
			holiday = "休"
		}
		_ = w.Write([]string{
			d.Date,
			d.Weekday,
			fmt.Sprintf("%d", d.AMCount),
			fmt.Sprintf("%d", d.AMNet),
			fmt.Sprintf("%d", d.PMCount),
			fmt.Sprintf("%d", d.PMNet),
			fmt.Sprintf("%d", d.DayNet),
			fmt.Sprintf("%d", d.Refund),
			amClosed,
			pmClosed,
			holiday,
			"", "", "", "", // 税率別内訳は月次サマリのみ（日別内訳なし）
		})
	}

	tb := result.Summary.TaxBreakdown
	_ = w.Write([]string{
		"合計", "", "", "", "", "",
		fmt.Sprintf("%d", result.Summary.NetAmount),
		fmt.Sprintf("%d", result.Summary.TotalRefund),
		"", "", "",
		fmt.Sprintf("%d", tb.Standard.TaxableAmount),
		fmt.Sprintf("%d", tb.Standard.TaxAmount),
		fmt.Sprintf("%d", tb.Reduced.TaxableAmount),
		fmt.Sprintf("%d", tb.Reduced.TaxAmount),
	})
	w.Flush()

	return &MonthlyCSVResult{
		Filename: fmt.Sprintf("monthly_report_%04d%02d.csv", year, month),
		Data:     buf.Bytes(),
	}, nil
}

// resolvePaymentMethodName は支払方法IDを名前に変換する。nil の場合は "現金" を返す
func resolvePaymentMethodName(id *uint64, names map[uint64]string) string {
	if id == nil {
		return "現金"
	}
	if name, ok := names[*id]; ok {
		return name
	}
	return fmt.Sprintf("支払方法(%d)", *id)
}

// paymentMethodNameForClose は ClosePreview 用の支払方法名解決ヘルパー
// cash_register_service.go からも利用するためパッケージスコープで定義する
func paymentMethodNameForClose(id *uint64, names map[uint64]string) string {
	return resolvePaymentMethodName(id, names)
}
