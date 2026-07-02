package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CloseRegisterInput はレジ締め実行の入力
type CloseRegisterInput struct {
	Date       time.Time
	Period     string
	ActualCash int64
	Memo       string
	ClosedBy   *uint64
}

// CashRegisterPreview は締めプレビューのフロントエンド向けレスポンス
type CashRegisterPreview struct {
	Date            string                `json:"date"`
	Period          string                `json:"period"`
	PeriodStart     string                `json:"period_start"`
	PeriodEnd       string                `json:"period_end"`
	IsAlreadyClosed bool                  `json:"is_already_closed"`
	IsHoliday       bool                  `json:"is_holiday"`
	Aggregate       CloseAggregateSummary `json:"aggregate"`
	BillingDetails  []CloseBillingDetail  `json:"billing_details"`
}

// CloseAggregateSummary は締めプレビューの集計サマリー
type CloseAggregateSummary struct {
	Categories      map[string]map[string]int64 `json:"categories"` // category → {payment_method_name → amount}
	PaymentMethods  []model.PaymentMethodMaster `json:"payment_methods"`
	TheoreticalCash int64                       `json:"theoretical_cash"`
	TaxBreakdown    TaxBreakdownSummary         `json:"tax_breakdown"`
}

// CloseBillingDetail は個別会計一覧の1レコード（フロントエンド向け）
type CloseBillingDetail struct {
	BillingID         uint64  `json:"billing_id"`
	PaidAt            string  `json:"paid_at"` // ISO8601
	OwnerName         string  `json:"owner_name"`
	PetName           string  `json:"pet_name"`
	IsHospitalization bool    `json:"is_hospitalization"`
	Category          string  `json:"category"`
	PaymentMethodID   *uint64 `json:"payment_method_id,omitempty"`
	PaymentMethodName string  `json:"payment_method_name"`
	BillingAmount     int64   `json:"billing_amount"`
	RefundAmount      int64   `json:"refund_amount"`
	NetAmount         int64   `json:"net_amount"`
}

// buildCategoryBreakdown はカテゴリ行と支払方法行から CategoryBreakdownSchema を構築する。
// 混在支払いの場合、カテゴリ金額を支払方法比率で按分する。
// 支払方法キーは system_key（例: "cash", "credit_card"）を使用。未登録 id は "method_N" にフォールバック。
func buildCategoryBreakdown(payRows []repository.PaymentAggregateRow, catRows []repository.CategoryAggregateRow, taxRows []repository.TaxBreakdownRow, payMethods []model.PaymentMethodMaster) model.CategoryBreakdownSchema {
	idToKey := make(map[uint64]string, len(payMethods))
	for i := range payMethods {
		m := &payMethods[i]
		if m.SystemKey != nil {
			idToKey[m.ID] = *m.SystemKey
		} else {
			idToKey[m.ID] = fmt.Sprintf("method_%d", m.ID)
		}
	}

	var totalPayment int64
	for _, pm := range payRows {
		totalPayment += pm.Amount
	}

	cats := make(map[string]map[string]int64)
	for _, cat := range catRows {
		cats[cat.Category] = make(map[string]int64)
		for _, pm := range payRows {
			if pm.PaymentMethodID == nil {
				continue // #128 hotfix 後は payment_method_id=NULL は発生しない
			}
			key, ok := idToKey[*pm.PaymentMethodID]
			if !ok {
				key = fmt.Sprintf("method_%d", *pm.PaymentMethodID)
			}
			if totalPayment > 0 {
				cats[cat.Category][key] += cat.Amount * pm.Amount / totalPayment
			}
		}
	}

	tax := buildTaxBreakdown(taxRows)
	return model.CategoryBreakdownSchema{
		Categories: cats,
		TaxBreakdown: model.TaxBreakdown{
			Standard: model.TaxBreakdownItem{TaxableAmount: tax.Standard.TaxableAmount, TaxAmount: tax.Standard.TaxAmount},
			Reduced:  model.TaxBreakdownItem{TaxableAmount: tax.Reduced.TaxableAmount, TaxAmount: tax.Reduced.TaxAmount},
		},
	}
}

// CashRegisterService はレジ締めのビジネスロジックインターフェース
type CashRegisterService interface {
	GetPreview(ctx context.Context, clinicID uint64, dateStr, period string) (*CashRegisterPreview, error)
	Close(ctx context.Context, clinicID uint64, input CloseRegisterInput) (*model.CashRegisterClose, error)
	List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error)
	// #115: 指定日にレジ締めが存在するか確認する（会計の締め後編集チェック用）。
	IsDateClosed(ctx context.Context, clinicID uint64, date time.Time) (bool, error)
}

// periodAggregate は集計処理の共通結果型
type periodAggregate struct {
	Schedule        *DaySchedule
	PaymentRows     []repository.PaymentAggregateRow
	CategoryRows    []repository.CategoryAggregateRow
	TotalRefund     int64
	BillingDetails  []repository.CloseBillingDetail
	TaxBreakdown    []repository.TaxBreakdownRow
	TheoreticalCash int64
	PeriodStart     time.Time
	PeriodEnd       time.Time
	// PaymentMethods は当該 clinic の支払方法マスタ。現金マスタ id 判定（#128）と GetPreview の二重ロード回避に使う。
	PaymentMethods []model.PaymentMethodMaster
}

type cashRegisterService struct {
	closeRepo      repository.CashRegisterCloseRepository
	accountingRepo repository.AccountingRepository
	closingsSvc    ClosingSettingsService
	payMethodRepo  repository.PaymentMethodMasterRepository
}

// NewCashRegisterService は CashRegisterService を初期化して返す
func NewCashRegisterService(
	closeRepo repository.CashRegisterCloseRepository,
	accountingRepo repository.AccountingRepository,
	closingsSvc ClosingSettingsService,
	payMethodRepo repository.PaymentMethodMasterRepository,
) CashRegisterService {
	return &cashRegisterService{
		closeRepo:      closeRepo,
		accountingRepo: accountingRepo,
		closingsSvc:    closingsSvc,
		payMethodRepo:  payMethodRepo,
	}
}

// validatePeriod は period 値（"am"/"pm"/"emg"）のバリデーションを行う
func validatePeriod(period string) error {
	if period != "am" && period != "pm" && period != "emg" {
		return apperrors.WrapInvalidInput("period は 'am'、'pm'、'emg' のいずれかを指定してください")
	}
	return nil
}

// fetchAggregate は date/period から集計処理（スケジュール取得・期間算出・会計集計）を実行する共通ロジック
func (s *cashRegisterService) fetchAggregate(ctx context.Context, clinicID uint64, date time.Time, period string) (*periodAggregate, error) {
	schedule, err := s.closingsSvc.ResolveSchedule(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to resolve schedule", "error", err)
		return nil, apperrors.Wrap(err, "failed to resolve schedule")
	}

	dateJST := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jstLocation)

	periodStart, periodEnd, err := resolvePeriodRange(dateJST, period, schedule)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch aggregate cash register", "error", err)
		return nil, apperrors.Wrap(err, "failed to fetch aggregate cash register")
	}

	aggregate, err := s.accountingRepo.GetCloseAggregate(ctx, repository.GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to aggregate billings", "error", err)
		return nil, apperrors.Wrap(err, "failed to aggregate billings")
	}

	// #128: 現金マスタ id を解決し、payment_method_id=現金id の split を理論現金に含める。
	payMethods, err := s.payMethodRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load payment methods for aggregate", "error", err)
		return nil, apperrors.Wrap(err, "failed to load payment methods for aggregate")
	}
	cashMethodID, err := findCashMethodID(payMethods)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find cash payment method", "error", err)
		return nil, apperrors.Wrap(err, "failed to find cash payment method")
	}

	return &periodAggregate{
		Schedule:        schedule,
		PaymentRows:     aggregate.PaymentRows,
		CategoryRows:    aggregate.CategoryRows,
		TotalRefund:     aggregate.TotalRefund,
		BillingDetails:  aggregate.BillingDetails,
		TaxBreakdown:    aggregate.TaxBreakdown,
		TheoreticalCash: calcTheoreticalCash(aggregate.PaymentRows, aggregate.TotalRefund, cashMethodID),
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		PaymentMethods:  payMethods,
	}, nil
}

func (s *cashRegisterService) GetPreview(ctx context.Context, clinicID uint64, dateStr, period string) (*CashRegisterPreview, error) {
	if dateStr == "" {
		return nil, apperrors.WrapInvalidInput("date クエリパラメータは必須です")
	}
	date, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("date は YYYY-MM-DD 形式で指定してください")
	}
	if period == "" {
		return nil, apperrors.WrapInvalidInput("period クエリパラメータは必須です")
	}
	if err := validatePeriod(period); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate period")
	}

	agg, err := s.fetchAggregate(ctx, clinicID, date, period)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch aggregate cash register", "error", err)
		return nil, apperrors.Wrap(err, "failed to fetch aggregate cash register")
	}

	// 二重締め確認
	existing, err := s.closeRepo.FindByDateAndPeriod(ctx, clinicID, date, period)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check existing close", "error", err)
		return nil, apperrors.Wrap(err, "failed to check existing close")
	}
	isAlreadyClosed := existing != nil

	// 支払方法マスタは fetchAggregate で取得済みのものを再利用する（#128: 二重ロード回避）
	payMethods := agg.PaymentMethods
	payMethodNames := buildPayMethodNameMap(payMethods)

	// カテゴリ別×支払方法名別集計マップを構築（混在支払いは比率按分）
	var totalPayment int64
	for _, pm := range agg.PaymentRows {
		totalPayment += pm.Amount
	}
	categories := make(map[string]map[string]int64)
	for _, cat := range agg.CategoryRows {
		categories[cat.Category] = make(map[string]int64)
		for _, pm := range agg.PaymentRows {
			pmName := paymentMethodNameForClose(pm.PaymentMethodID, payMethodNames)
			if totalPayment > 0 {
				categories[cat.Category][pmName] += cat.Amount * pm.Amount / totalPayment
			}
		}
	}

	// 個別会計一覧を変換
	details := make([]CloseBillingDetail, 0, len(agg.BillingDetails))
	for _, d := range agg.BillingDetails {
		pmName := paymentMethodNameForClose(d.PaymentMethodID, payMethodNames)
		details = append(details, CloseBillingDetail{
			BillingID:         d.BillingID,
			PaidAt:            d.PaidAt.In(time.Local).Format(time.RFC3339),
			OwnerName:         d.OwnerName,
			PetName:           d.PetName,
			IsHospitalization: d.IsHospitalization,
			Category:          d.Category,
			PaymentMethodID:   d.PaymentMethodID,
			PaymentMethodName: pmName,
			BillingAmount:     d.BillingAmount,
			RefundAmount:      d.RefundAmount,
			NetAmount:         d.NetAmount,
		})
	}

	// 税率別集計
	taxSummary := buildTaxBreakdown(agg.TaxBreakdown)

	return &CashRegisterPreview{
		Date:            date.In(time.Local).Format("2006-01-02"),
		Period:          period,
		PeriodStart:     agg.PeriodStart.In(time.Local).Format(time.RFC3339),
		PeriodEnd:       agg.PeriodEnd.In(time.Local).Format(time.RFC3339),
		IsAlreadyClosed: isAlreadyClosed,
		IsHoliday:       agg.Schedule != nil && agg.Schedule.IsHoliday,
		Aggregate: CloseAggregateSummary{
			Categories:      categories,
			PaymentMethods:  payMethods,
			TheoreticalCash: agg.TheoreticalCash,
			TaxBreakdown:    taxSummary,
		},
		BillingDetails: details,
	}, nil
}

func (s *cashRegisterService) Close(ctx context.Context, clinicID uint64, input CloseRegisterInput) (*model.CashRegisterClose, error) {
	if err := validatePeriod(input.Period); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate period")
	}

	// 二重締めチェック
	existing, err := s.closeRepo.FindByDateAndPeriod(ctx, clinicID, input.Date, input.Period)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check existing close", "error", err)
		return nil, apperrors.Wrap(err, "failed to check existing close")
	}
	if existing != nil {
		return nil, apperrors.WrapConflict("この日時はすでに締め済みです")
	}

	agg, err := s.fetchAggregate(ctx, clinicID, input.Date, input.Period)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch aggregate cash register", "error", err)
		return nil, apperrors.Wrap(err, "failed to fetch aggregate cash register")
	}

	cashDifference := input.ActualCash - agg.TheoreticalCash

	// category_breakdown JSONB を構築（消費税内訳を含む）
	breakdownSchema := buildCategoryBreakdown(agg.PaymentRows, agg.CategoryRows, agg.TaxBreakdown, agg.PaymentMethods)
	breakdownJSON, err := json.Marshal(breakdownSchema)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal category breakdown", "error", err)
		return nil, apperrors.Wrap(err, "failed to marshal category breakdown")
	}

	record := &model.CashRegisterClose{
		ClinicID:          clinicID,
		CloseDate:         input.Date,
		Period:            input.Period,
		TheoreticalCash:   agg.TheoreticalCash,
		ActualCash:        input.ActualCash,
		CashDifference:    cashDifference,
		CategoryBreakdown: breakdownJSON,
		Memo:              input.Memo,
		ClosedBy:          input.ClosedBy,
		ClosedAt:          time.Now(),
	}

	if err := s.closeRepo.Create(ctx, record); err != nil {
		slog.ErrorContext(ctx, "failed to create cash register close", "error", err)
		return nil, apperrors.Wrap(err, "failed to create cash register close")
	}

	slog.InfoContext(ctx, "cash register closed",
		slog.Uint64("clinic_id", clinicID),
		slog.String("date", input.Date.Format("2006-01-02")),
		slog.String("period", input.Period),
		slog.Int64("theoretical_cash", agg.TheoreticalCash),
		slog.Int64("actual_cash", input.ActualCash))

	return record, nil
}

func (s *cashRegisterService) List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
	records, total, err := s.closeRepo.FindAll(ctx, clinicID, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list cash register closes", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list cash register closes")
	}
	return records, total, nil
}

func (s *cashRegisterService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
	return s.closeRepo.FindByID(ctx, clinicID, id)
}

func (s *cashRegisterService) IsDateClosed(ctx context.Context, clinicID uint64, date time.Time) (bool, error) {
	closed, err := s.closeRepo.HasCloseOnDate(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check if date is closed", "error", err, "clinic_id", clinicID)
		return false, apperrors.Wrap(err, "failed to check if date is closed")
	}
	return closed, nil
}

// resolvePeriodRange は period（"am"/"pm"/"emg"）と DaySchedule から集計期間（JST）を返す
func resolvePeriodRange(dateJST time.Time, period string, schedule *DaySchedule) (start, end time.Time, err error) {
	boundaryH, boundaryM, parseErr := parseHHMM(schedule.AmPmBoundary)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("am_pm_boundary の形式が正しくありません")
	}
	pmEndH, pmEndM, parseErr := parseHHMM(schedule.PmEnd)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("pm_end の形式が正しくありません")
	}

	boundary := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), boundaryH, boundaryM, 0, 0, jstLocation)
	pmEnd := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), pmEndH, pmEndM, 0, 0, jstLocation)
	dayStart := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, jstLocation)

	switch period {
	case "am":
		return dayStart, boundary, nil
	case "pm":
		return boundary, pmEnd, nil
	case "emg":
		// 越日EMG（18:30〜翌8:59）は未実装。現行は同日内（pmEnd〜24:00）で集計する。
		// 追跡Issueは未起票（bug.md M-3）。
		nextDayStart := dayStart.AddDate(0, 0, 1)
		return pmEnd, nextDayStart, nil
	default:
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("period は 'am'、'pm'、'emg' のいずれかを指定してください")
	}
}

// parseHHMM は "HH:MM" または "HH:MM:SS"（PostgreSQL time 型）形式の時刻文字列を時・分に分解する
func parseHHMM(s string) (h, m int, err error) {
	// PostgreSQL time 型は "HH:MM:SS" で返るので秒部分を除去する
	if len(s) == 8 && s[2] == ':' && s[5] == ':' {
		s = s[:5]
	}
	if len(s) != 5 || s[2] != ':' {
		return 0, 0, apperrors.WrapInvalidInput("時刻は HH:MM 形式で指定してください")
	}
	var hh, mm int
	_, parseErr := fmt.Sscanf(s, "%d:%d", &hh, &mm)
	if parseErr != nil {
		return 0, 0, apperrors.WrapInvalidInput("時刻の解析に失敗しました")
	}
	return hh, mm, nil
}

// findCashMethodID は payment_methods マスタ群から現金マスタ（system_key='cash'）の id を返す（#197）。
// system_key 一致で検索するため、クリニックが「現金」等の name を改名しても正しく現金マスタを識別できる。
// 現金マスタが存在しない場合は内部エラーを返す（migration 009 で必ず backfill 済み）。
func findCashMethodID(methods []model.PaymentMethodMaster) (uint64, error) {
	cashKey := paymentMethodSystemKeys[model.PaymentMethodCash]
	for i := range methods {
		if methods[i].SystemKey != nil && *methods[i].SystemKey == cashKey {
			return methods[i].ID, nil
		}
	}
	return 0, apperrors.WrapInternalServerError("現金マスタが見つかりません (system_key='cash')")
}

// calcTheoreticalCash は支払方法別集計行から現金合計を返し、返金合計を差し引いて理論現金残高を算出する。
// 現金判定: payment_method_id が現金マスタ id と一致する split（#197 の新データ）に加え、
// payment_method_id=NULL の split（旧 seed/レガシーデータの現金 split）も現金として集計する（#128 後方互換）。
// 集計クエリ（accounting_repository_reports_close.go）は payment_method_id で GROUP BY し method ENUM を
// 射影しないため、NULL 行の元 method は判定できない。月次レポート側 resolvePaymentMethodName も NULL→"現金"
// 扱いのため、締め理論現金と月次レポート現金を一致させるにはここでも NULL を現金に含める（bug.md H-3）。
func calcTheoreticalCash(payRows []repository.PaymentAggregateRow, totalRefund int64, cashMethodID uint64) int64 {
	var cashTotal int64
	for _, r := range payRows {
		if r.PaymentMethodID == nil || *r.PaymentMethodID == cashMethodID {
			cashTotal += r.Amount
		}
	}
	return cashTotal - totalRefund
}
