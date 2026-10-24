package billing

import (
	"context"
	"sort"

	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
	// UnclassifiedOtherCount は category=other 明細を1件以上持つ会計の distinct 件数（DEC-40）。
	UnclassifiedOtherCount int64 `json:"unclassified_other_count"`
	// CategoryCounts は部門ごとの会計 distinct 件数（#247 DEC-16⑥ / DEC-40）。
	// key = item_category enum。FE は DISPLAY_CATEGORIES で集約する。
	CategoryCounts map[string]int64 `json:"category_counts"`
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
	// CategoryAmounts は当該 split 支払額の部門別 GROSS 内訳（EMR-354）。repository から透過。
	// nil = 旧バックエンド相当（FE は「記録なし」表示）。
	CategoryAmounts map[string]int64 `json:"category_amounts"`
}

// buildCategoryBreakdown builds CategoryBreakdownSchema from a per-billing allocated matrix
// (#247 DEC-16⑥). Matrix keys are system_key / method_N / cash (NULL legacy).
// Do NOT reintroduce period-wide payment ratio pseudo-allocation.
// unclassifiedOtherCount は DEC-40 の会計 distinct 件数を snapshot に保存する（0 でもポインタで記録）。
func buildCategoryBreakdown(matrix map[string]map[string]int64, taxRows []TaxBreakdownRow, taxRates accountingReportTaxRates, unclassifiedOtherCount int64) model.CategoryBreakdownSchema {
	// Defensive copy so callers cannot mutate snapshot map.
	cats := make(map[string]map[string]int64, len(matrix))
	for cat, byMethod := range matrix {
		dst := make(map[string]int64, len(byMethod))
		for k, v := range byMethod {
			dst[k] = v
		}
		cats[cat] = dst
	}

	tax := buildTaxBreakdown(taxRows, taxRates)
	count := unclassifiedOtherCount
	return model.CategoryBreakdownSchema{
		Categories: cats,
		TaxBreakdown: model.TaxBreakdown{
			Standard: model.TaxBreakdownItem{TaxableAmount: tax.Standard.TaxableAmount, TaxAmount: tax.Standard.TaxAmount},
			Reduced:  model.TaxBreakdownItem{TaxableAmount: tax.Reduced.TaxableAmount, TaxAmount: tax.Reduced.TaxAmount},
		},
		UnclassifiedOtherCount: &count,
	}
}

// buildPreviewCategories validates categories and returns the name-keyed matrix for UI.
func buildPreviewCategories(matrix map[string]map[string]int64) (map[string]map[string]int64, error) {
	for cat := range matrix {
		if err := validateCloseAggregateCategory(cat); err != nil {
			return nil, err
		}
	}
	// Defensive copy
	out := make(map[string]map[string]int64, len(matrix))
	for cat, byMethod := range matrix {
		dst := make(map[string]int64, len(byMethod))
		for k, v := range byMethod {
			dst[k] = v
		}
		out[cat] = dst
	}
	return out, nil
}

// computeCategoryPaymentMatrix allocates payment net onto categories per billing (DEC-16⑥).
func computeCategoryPaymentMatrix(
	data *CategoryPaymentAllocationData,
	methodKeyFn func(*uint64) string,
	refundMethodKeyFn func(*string) string,
) map[string]map[string]int64 {
	billings := BuildAllocationBillings(data, methodKeyFn, refundMethodKeyFn)
	return AggregateCategoryPaymentMatrix(billings)
}

// orderPaymentMethodsForMatrix returns payment method columns:
// active masters in display_order, then inactive masters that have period data, then
// unknown method labels present in the matrix (DEC-16⑥ 末尾表示).
func orderPaymentMethodsForMatrix(masters []model.PaymentMethodMaster, matrix map[string]map[string]int64) []model.PaymentMethodMaster {
	usedNames := make(map[string]struct{})
	for _, byMethod := range matrix {
		for name := range byMethod {
			usedNames[name] = struct{}{}
		}
	}

	out := make([]model.PaymentMethodMaster, 0, len(masters)+len(usedNames))
	seenName := make(map[string]struct{}, len(masters))
	// Pass 1: active masters (already ordered by FindAll display_order)
	for i := range masters {
		m := masters[i]
		if !m.IsActive {
			continue
		}
		out = append(out, m)
		seenName[m.Name] = struct{}{}
		delete(usedNames, m.Name)
	}
	// Pass 2: inactive masters with data
	for i := range masters {
		m := masters[i]
		if m.IsActive {
			continue
		}
		if _, ok := usedNames[m.Name]; !ok {
			continue
		}
		out = append(out, m)
		seenName[m.Name] = struct{}{}
		delete(usedNames, m.Name)
	}
	// Pass 3: unknown / deleted labels still holding amounts
	unknown := make([]string, 0, len(usedNames))
	for name := range usedNames {
		if _, ok := seenName[name]; ok {
			continue
		}
		unknown = append(unknown, name)
	}
	sort.Strings(unknown)
	for _, name := range unknown {
		out = append(out, model.PaymentMethodMaster{
			Name:     name,
			IsActive: false,
		})
	}
	return out
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

type cashRegisterService struct {
	closeRepo      CashRegisterCloseRepository
	accountingRepo AccountingRepository
	closingsSvc    closingScheduleResolver
	payMethodRepo  PaymentMethodMasterRepository
	clinicRepo     billingClinicReader
	transactor     Transactor
}

// NewCashRegisterService は CashRegisterService を初期化して返す
func NewCashRegisterService(
	closeRepo CashRegisterCloseRepository,
	accountingRepo AccountingRepository,
	closingsSvc closingScheduleResolver,
	payMethodRepo PaymentMethodMasterRepository,
	clinicRepo billingClinicReader,
	transactor Transactor,
) CashRegisterService {
	return &cashRegisterService{
		closeRepo:      closeRepo,
		accountingRepo: accountingRepo,
		closingsSvc:    closingsSvc,
		payMethodRepo:  payMethodRepo,
		clinicRepo:     clinicRepo,
		transactor:     transactor,
	}
}

// validatePeriod は period 値（"am"/"pm"/"emg"）のバリデーションを行う
func validatePeriod(period string) error {
	if period != "am" && period != "pm" && period != "emg" {
		return apperrors.WrapInvalidInput("period は 'am'、'pm'、'emg' のいずれかを指定してください")
	}
	return nil
}

// レジ締めのプレビューと締めの作成は、held for EMR-509。
// 旧実装は completed_at 基準の会計集計（payments / payment_splits / billing_refunds）から理論現金を算出し、
// cash_difference を保存していたが、いずれも廃止された（001_init.sql:10899,10918-10920）。後継は締め境界の版
// （closing_boundary_revisions）・開始残高（cash_opening_balances）・現金の補充回収（cash_adjustment_events）・
// 受領と返金の記録（finance_source_events）を束ねる締め文脈（cash_close_contexts。D12）で、その reader / writer が未接続のため、
// 入力の形の検証の後に固定エラー（422）で拒否する。理論現金を 0 などで返さず、締めレコードを作らない。
// 締め済みの一覧・詳細・締め済み判定の読取は残す。
// 診療終了の判定などを含む旧集計の純粋な helper（resolvePeriodRange・calcTheoreticalCash・buildCategoryBreakdown 等）は 509 が再利用する。
const (
	cashRegisterPreviewPendingMessage = "レジ締めのプレビューは、受領・返金・確定の記録と締め境界の版の読取が未接続のため利用できません"
	cashRegisterClosePendingMessage   = "レジ締めの実行は、締め文脈（開始残高・補充回収・受領と返金の記録）の保存が未接続のため利用できません"
)

func (*cashRegisterService) GetPreview(_ context.Context, _ uint64, dateStr, period string) (*CashRegisterPreview, error) {
	if dateStr == "" {
		return nil, apperrors.WrapInvalidInput("date クエリパラメータは必須です")
	}
	if _, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local); err != nil {
		return nil, apperrors.WrapInvalidInput("date は YYYY-MM-DD 形式で指定してください")
	}
	if period == "" {
		return nil, apperrors.WrapInvalidInput("period クエリパラメータは必須です")
	}
	if err := validatePeriod(period); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate period")
	}
	return nil, apperrors.WrapFinanceConsumerPending(cashRegisterPreviewPendingMessage)
}

// Close は transaction・境界ロックを開く前に拒否する（失敗のためだけに advisory lock を取らない）。
func (*cashRegisterService) Close(_ context.Context, _ uint64, input CloseRegisterInput) (*model.CashRegisterClose, error) {
	if err := validatePeriod(input.Period); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate period")
	}
	if input.ActualCash < 0 {
		return nil, apperrors.WrapInvalidInput("actual_cash は 0 以上で指定してください")
	}
	return nil, apperrors.WrapFinanceConsumerPending(cashRegisterClosePendingMessage)
}

func (s *cashRegisterService) List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
	records, total, err := s.closeRepo.FindAll(ctx, clinicID, startDate, endDate, page, limit)
	if err != nil {
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
		return false, apperrors.Wrap(err, "failed to check if date is closed")
	}
	return closed, nil
}

// resolvePeriodRange は period（"am"/"pm"/"emg"）と sharedkernel.DaySchedule から集計期間（JST）を返す。
// 集計クエリ（GetCloseAggregate）は completed_at >= start AND < end の終端排他なので、
// AM=[am_start, boundary) / PM=[boundary, pmEnd) / EMG=[pmEnd, 翌日 am_start) は
// 連続・非重複で24時間を被覆する（#215: 深夜 0:00〜am_start の会計は前日 EMG に帰属）。
func resolvePeriodRange(dateJST time.Time, period string, schedule *sharedkernel.DaySchedule) (start, end time.Time, err error) {
	boundaryH, boundaryM, parseErr := sharedkernel.ParseHHMM(schedule.AmPmBoundary)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("am_pm_boundary の形式が正しくありません")
	}
	pmEndH, pmEndM, parseErr := sharedkernel.ParseHHMM(schedule.PmEnd)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("pm_end の形式が正しくありません")
	}
	amStartStr := schedule.AmStart
	if amStartStr == "" {
		// migration 011 以前のデータ・旧呼び出し元は既定 09:00 として扱う（#215 後方互換）
		amStartStr = defaultClosingAmStart
	}
	amStartH, amStartM, parseErr := sharedkernel.ParseHHMM(amStartStr)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("am_start の形式が正しくありません")
	}

	boundary := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), boundaryH, boundaryM, 0, 0, config.JST)
	pmEnd := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), pmEndH, pmEndM, 0, 0, config.JST)
	amStart := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), amStartH, amStartM, 0, 0, config.JST)
	if !amStart.Before(boundary) {
		// 逆転設定は空レンジ集計を silent に返すより設定不正として fail-loud にする
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("am_start は am_pm_boundary より前に設定してください")
	}

	switch period {
	case "am":
		return amStart, boundary, nil
	case "pm":
		return boundary, pmEnd, nil
	case "emg":
		// #215: EMG は当日 pmEnd 〜 翌日 am_start の越日レンジ。am_start は標準設定由来で日別に
		// 変わらない（特別期間も標準設定を継承する）ため、「翌日の am_start」は当日値 + 1日 と一致する。
		return pmEnd, amStart.AddDate(0, 0, 1), nil
	default:
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("period は 'am'、'pm'、'emg' のいずれかを指定してください")
	}
}

// findCashMethodID は payment_methods マスタ群から現金マスタ（system_key='cash'）の id を返す（#197）。
// system_key 一致で検索するため、クリニックが「現金」等の name を改名しても正しく現金マスタを識別できる。
// 現金マスタが存在しない場合は内部エラーを返す（migration 009 で必ず backfill 済み）。
func findCashMethodID(methods []model.PaymentMethodMaster) (uint64, error) {
	cashKey := PaymentMethodSystemKeys[model.PaymentMethodCash]
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
func calcTheoreticalCash(payRows []PaymentAggregateRow, totalRefund int64, cashMethodID uint64) int64 {
	var cashTotal int64
	for _, r := range payRows {
		if r.PaymentMethodID == nil || *r.PaymentMethodID == cashMethodID {
			cashTotal += r.Amount
		}
	}
	return cashTotal - totalRefund
}

// defaultClosingAmStart は AM 開始時刻の既定値（#215・service/closing_settings_service.go の
// 同名 const の複製。migration 011 の DB default と一致）。
const defaultClosingAmStart = "09:00"
