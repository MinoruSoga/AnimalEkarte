package billing

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- レスポンス型（フロントエンド期待形式） ----

// MonthlyReportResponse は月次売上レポートのフロントエンド向けレスポンス
type MonthlyReportResponse struct {
	Year         int                  `json:"year"`
	Month        int                  `json:"month"`
	StartDate    string               `json:"start_date"`
	EndDate      string               `json:"end_date"`
	Summary      MonthlyReportSummary `json:"summary"`
	DailyDetails []DailyReportDetail  `json:"daily_details"`
	// CategoryPaymentMatrix は #247 部門×支払方法統合表（画面・印刷同一データ）。
	CategoryPaymentMatrix *CategoryPaymentMatrix `json:"category_payment_matrix"`
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

// CategoryPaymentMatrix は #247 DEC-16⑥ のカテゴリ×支払方法マトリクス。
// 金額は支払実額基準（割引適用後）。件数は会計 distinct。
type CategoryPaymentMatrix struct {
	// PaymentMethods は列順（医院 master 順 + 期間内データ付き inactive/unknown 末尾）。
	PaymentMethods []CategoryPaymentMethodColumn `json:"payment_methods"`
	// Rows は部門行（amount 0 かつ count 0 の行は省略可だが、FE が DISPLAY 集約するため raw category を返す）。
	Rows []CategoryPaymentMatrixRow `json:"rows"`
	// Totals は列合計・総合計。件数合計は会計 distinct 総数（行件数の単純合算ではない）。
	Totals CategoryPaymentMatrixTotals `json:"totals"`
}

// CategoryPaymentMethodColumn はマトリクスの支払方法列。
type CategoryPaymentMethodColumn struct {
	ID       *uint64 `json:"id,omitempty"`
	Name     string  `json:"name"`
	IsActive bool    `json:"is_active"`
}

// CategoryPaymentMatrixRow は1カテゴリ行。
type CategoryPaymentMatrixRow struct {
	Category string           `json:"category"`
	Count    int64            `json:"count"` // 会計 distinct
	ByMethod map[string]int64 `json:"by_method"`
	RowTotal int64            `json:"row_total"`
}

// CategoryPaymentMatrixTotals はフッタ合計。
type CategoryPaymentMatrixTotals struct {
	// Count は期間内完了会計の distinct 総数（split 二重計上なし）。
	Count      int64            `json:"count"`
	ByMethod   map[string]int64 `json:"by_method"`
	GrandTotal int64            `json:"grand_total"`
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

type accountingReportTaxRates struct {
	StandardPercent int64
	ReducedPercent  int64
}

func clinicTaxRates(clinic *model.Clinic) accountingReportTaxRates {
	standard := int64(10)
	reduced := int64(8)
	if clinic != nil {
		if clinic.StandardTaxRate > 0 {
			standard = int64(math.Round(clinic.StandardTaxRate * 100))
		}
		if clinic.ReducedTaxRate > 0 {
			reduced = int64(math.Round(clinic.ReducedTaxRate * 100))
		}
	}
	return accountingReportTaxRates{StandardPercent: standard, ReducedPercent: reduced}
}

func isReducedTaxRate(rowRate int64, rates accountingReportTaxRates) bool {
	return rowRate == rates.ReducedPercent
}

// buildMonthlyReportResponse は生データからフロントエンド向けレスポンスを構築する。
// matrix は #247 配賦済み（nil 可 — その場合 category_payment_matrix は空骨格）。
func buildMonthlyReportResponse(
	year, month int,
	startDate, endDate time.Time,
	raw *MonthlyReportResult,
	payMethodNames map[uint64]string,
	holidaySet map[string]bool,
	taxRates accountingReportTaxRates,
	matrix *CategoryPaymentMatrix,
) *MonthlyReportResponse {
	byPaymentMethod := make(map[string]int64)
	byCategory := make(map[string]int64)
	dailyMap, totalAmount := accumulateMonthlyDailyMap(raw, byPaymentMethod, payMethodNames)
	for _, row := range raw.CategoryRows {
		byCategory[row.Category] += row.Amount
	}
	taxSummary := summarizeMonthlyTax(raw, taxRates)
	dailyDetails, workingDays := buildMonthlyDailyDetails(startDate, endDate, dailyMap, holidaySet)

	summary := MonthlyReportSummary{
		WorkingDays:     workingDays,
		TotalBillings:   raw.BillingCount,
		TotalAmount:     totalAmount,
		TotalRefund:     raw.TotalRefund,
		NetAmount:       raw.GrandTotal,
		ByPaymentMethod: byPaymentMethod,
		ByCategory:      byCategory,
		TaxBreakdown:    taxSummary,
	}

	if matrix == nil {
		matrix = &CategoryPaymentMatrix{
			PaymentMethods: []CategoryPaymentMethodColumn{},
			Rows:           []CategoryPaymentMatrixRow{},
			Totals: CategoryPaymentMatrixTotals{
				Count:    raw.BillingCount,
				ByMethod: map[string]int64{},
			},
		}
	}

	return &MonthlyReportResponse{
		Year:                  year,
		Month:                 month,
		StartDate:             startDate.Format(time.DateOnly),
		EndDate:               endDate.Format(time.DateOnly),
		Summary:               summary,
		DailyDetails:          dailyDetails,
		CategoryPaymentMatrix: matrix,
	}
}

// buildCategoryPaymentMatrix assembles the API matrix from allocation result + masters.
// byCategoryNet is derived from matrix row totals (支払実額) so summary.by_category matches matrix.
func buildCategoryPaymentMatrix(
	matrix map[string]map[string]int64,
	categoryCounts map[string]int64,
	payMethods []model.PaymentMethodMaster,
	billingCount int64,
) *CategoryPaymentMatrix {
	// Order columns: active master order → inactive with data → unknown labels
	ordered := orderPaymentMethodsForMatrix(payMethods, matrix)
	cols := make([]CategoryPaymentMethodColumn, 0, len(ordered))
	for i := range ordered {
		m := ordered[i]
		col := CategoryPaymentMethodColumn{
			Name:     m.Name,
			IsActive: m.IsActive,
		}
		if m.ID != 0 {
			id := m.ID
			col.ID = &id
		}
		cols = append(cols, col)
	}

	// Stable category order = AllItemCategories declaration order
	rows := make([]CategoryPaymentMatrixRow, 0, len(model.AllItemCategories()))
	colTotals := make(map[string]int64)
	var grand int64
	for _, cat := range model.AllItemCategories() {
		key := string(cat)
		byMethod := matrix[key]
		if len(byMethod) == 0 && categoryCounts[key] == 0 {
			continue
		}
		rowByMethod := make(map[string]int64, len(byMethod))
		var rowTotal int64
		for name, amount := range byMethod {
			rowByMethod[name] = amount
			rowTotal += amount
			colTotals[name] += amount
		}
		grand += rowTotal
		rows = append(rows, CategoryPaymentMatrixRow{
			Category: key,
			Count:    categoryCounts[key],
			ByMethod: rowByMethod,
			RowTotal: rowTotal,
		})
	}

	return &CategoryPaymentMatrix{
		PaymentMethods: cols,
		Rows:           rows,
		Totals: CategoryPaymentMatrixTotals{
			Count:      billingCount,
			ByMethod:   colTotals,
			GrandTotal: grand,
		},
	}
}

// buildTaxBreakdown は TaxBreakdownRow スライスを病院マスタ税率（exact-match）で標準/軽減に分類し
// TaxBreakdownSummary に変換する。cash_register_service.go（締めレジ経路）からも利用するため
// パッケージスコープで定義する。M-7(#191): 固定閾値 `>8` を廃止し月次レポート経路と分類規則を統一。
// 軽減税率と一致しない税率（0% 非課税を含む）は標準へ分類する（月次 #191 と同一規則）。
// 保存済みの過去の締め記録（category_breakdown スナップショット）は再計算しない。
func buildTaxBreakdown(rows []TaxBreakdownRow, rates accountingReportTaxRates) TaxBreakdownSummary {
	var summary TaxBreakdownSummary
	for _, tr := range rows {
		if isReducedTaxRate(tr.TaxRate, rates) {
			summary.Reduced.TaxableAmount += tr.TaxableAmount
			summary.Reduced.TaxAmount += tr.TaxAmount
		} else {
			summary.Standard.TaxableAmount += tr.TaxableAmount
			summary.Standard.TaxAmount += tr.TaxAmount
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
	GetMonthlyByPeriod(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyReportResponse, error)
	ExportMonthlyCSV(ctx context.Context, clinicID uint64, year, month int) (*MonthlyCSVResult, error)
	ExportMonthlyCSVByPeriod(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyCSVResult, error)
}

// ---- サービス実装 ----

type accountingReportService struct {
	repo          AccountingRepository
	payMethodRepo PaymentMethodMasterRepository
	holidayRepo   billingHolidayReader
	clinicRepo    billingClinicReader
}

// NewAccountingReportService は AccountingReportService を初期化して返す
func NewAccountingReportService(
	repo AccountingRepository,
	payMethodRepo PaymentMethodMasterRepository,
	holidayRepo billingHolidayReader,
	clinicRepo billingClinicReader,
) AccountingReportService {
	return &accountingReportService{
		repo:          repo,
		payMethodRepo: payMethodRepo,
		holidayRepo:   holidayRepo,
		clinicRepo:    clinicRepo,
	}
}

// validateMonth は month 値（1〜12）のバリデーションを行う
func validateMonth(month int) error {
	if month < 1 || month > 12 {
		return apperrors.WrapInvalidInput("month は 1〜12 の範囲で指定してください")
	}
	return nil
}

func validateReportPeriod(startDate, endDate time.Time) error {
	if startDate.After(endDate) {
		return apperrors.WrapInvalidInput("start_date は end_date 以前の日付を指定してください")
	}
	if endDate.Sub(startDate).Hours()/24 > 366 {
		return apperrors.WrapInvalidInput("集計期間は 367 日以内で指定してください")
	}
	return nil
}

func (s *accountingReportService) GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResponse, error) {
	if err := validateMonth(month); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate month")
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, config.JST)
	endDate := startDate.AddDate(0, 1, -1)
	raw, err := s.repo.GetMonthlyReport(ctx, clinicID, year, month)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get monthly report", "error", err)
		return nil, apperrors.Wrap(err, "failed to get monthly report")
	}
	return s.buildReportResponse(ctx, clinicID, year, month, startDate, endDate, raw)
}

func (s *accountingReportService) GetMonthlyByPeriod(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyReportResponse, error) {
	if err := validateReportPeriod(startDate, endDate); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate report period")
	}

	periodEnd := endDate.AddDate(0, 0, 1)
	raw, err := s.repo.GetMonthlyReportByPeriod(ctx, clinicID, startDate, periodEnd)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get monthly report by period", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get monthly report by period")
	}
	return s.buildReportResponse(ctx, clinicID, startDate.Year(), int(startDate.Month()), startDate, endDate, raw)
}

func (s *accountingReportService) buildReportResponse(
	ctx context.Context,
	clinicID uint64,
	year, month int,
	startDate, endDate time.Time,
	raw *MonthlyReportResult,
) (*MonthlyReportResponse, error) {
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
	holidaySet, err := s.buildHolidaySet(ctx, clinicID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	clinic, err := s.clinicRepo.FindByID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic tax settings", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get clinic tax settings")
	}

	// #247: category×payment matrix（期間は [startDate, endDate+1day)）
	periodEnd := endDate.AddDate(0, 0, 1)
	allocData, err := s.repo.GetCategoryPaymentAllocationData(ctx, clinicID, startDate, periodEnd)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load category payment allocation data", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to load category payment allocation data")
	}
	nameFn, nameRefundFn := BuildNameMethodResolvers(payMethods)
	allocated := computeCategoryPaymentMatrix(allocData, nameFn, nameRefundFn)
	counts := map[string]int64{}
	if allocData != nil {
		counts = allocData.CategoryCounts
	}
	matrix := buildCategoryPaymentMatrix(allocated, counts, payMethods, raw.BillingCount)

	// by_category を支払実額基準の行合計で上書き（summary と matrix の一致）
	byCategoryNet := make(map[string]int64, len(matrix.Rows))
	for _, row := range matrix.Rows {
		byCategoryNet[row.Category] = row.RowTotal
	}
	resp := buildMonthlyReportResponse(year, month, startDate, endDate, raw, payMethodNames, holidaySet, clinicTaxRates(clinic), matrix)
	if len(byCategoryNet) > 0 {
		resp.Summary.ByCategory = byCategoryNet
	}
	return resp, nil
}

func (s *accountingReportService) buildHolidaySet(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (map[string]bool, error) {
	holidaySet := make(map[string]bool)
	startStr := startDate.Format(time.DateOnly)
	endStr := endDate.Format(time.DateOnly)
	for cursor := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, config.JST); !cursor.After(endDate); cursor = cursor.AddDate(0, 1, 0) {
		yearMonth := fmt.Sprintf("%04d-%02d", cursor.Year(), cursor.Month())
		holidays, err := s.holidayRepo.FindAllByYearMonth(ctx, clinicID, yearMonth)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get clinic holidays", "error", err, "clinic_id", clinicID, "year_month", yearMonth)
			return nil, apperrors.Wrap(err, "failed to get clinic holidays")
		}
		for _, h := range holidays {
			date := h.Date.Format(time.DateOnly)
			if date >= startStr && date <= endStr {
				holidaySet[date] = true
			}
		}
	}
	return holidaySet, nil
}

// ExportMonthlyCSV は月次売上レポートの CSV バイト列とファイル名を返す
func (s *accountingReportService) ExportMonthlyCSV(ctx context.Context, clinicID uint64, year, month int) (*MonthlyCSVResult, error) {
	result, err := s.GetMonthly(ctx, clinicID, year, month)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get monthly accounting report", "error", err)
		return nil, apperrors.Wrap(err, "failed to get monthly accounting report")
	}
	return buildMonthlyCSV(result, fmt.Sprintf("monthly_report_%04d%02d.csv", year, month))
}

func (s *accountingReportService) ExportMonthlyCSVByPeriod(ctx context.Context, clinicID uint64, startDate, endDate time.Time) (*MonthlyCSVResult, error) {
	result, err := s.GetMonthlyByPeriod(ctx, clinicID, startDate, endDate)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get accounting report by period", "error", err)
		return nil, apperrors.Wrap(err, "failed to get accounting report by period")
	}
	return buildMonthlyCSV(result, fmt.Sprintf("monthly_report_%s_%s.csv", startDate.Format("20060102"), endDate.Format("20060102")))
}

func buildMonthlyCSV(result *MonthlyReportResponse, filename string) (*MonthlyCSVResult, error) {
	var buf bytes.Buffer
	// BOM（Excel の UTF-8 認識用）
	buf.Write([]byte("\xEF\xBB\xBF"))

	w := csv.NewWriter(&buf)
	if err := w.Write([]string{
		"日付", "曜日", "AM件数", "AM純売上(円)", "PM件数", "PM純売上(円)", "日計(円)", "返金(円)", "AM締め", "PM締め", "休診",
		"標準税率課税対象額(円)", "標準税率消費税額(円)", "軽減税率課税対象額(円)", "軽減税率消費税額(円)",
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to write CSV header")
	}

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
		if err := w.Write([]string{
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
		}); err != nil {
			return nil, apperrors.Wrap(err, "failed to write CSV row")
		}
	}

	tb := result.Summary.TaxBreakdown
	if err := w.Write([]string{
		"合計", "", "", "", "", "",
		fmt.Sprintf("%d", result.Summary.NetAmount),
		fmt.Sprintf("%d", result.Summary.TotalRefund),
		"", "", "",
		fmt.Sprintf("%d", tb.Standard.TaxableAmount),
		fmt.Sprintf("%d", tb.Standard.TaxAmount),
		fmt.Sprintf("%d", tb.Reduced.TaxableAmount),
		fmt.Sprintf("%d", tb.Reduced.TaxAmount),
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to write CSV summary")
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return nil, apperrors.Wrap(err, "CSV writer error")
	}

	return &MonthlyCSVResult{
		Filename: filename,
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
