package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
	"github.com/animal-ekarte/backend/internal/timeutil"
)

// CloseRegisterInput はレジ締め実行の入力
type CloseRegisterInput struct {
	Date       time.Time
	Period     string
	ActualCash int64
	Memo       string
	ClosedBy   *uint64
}

// VoidReopenInput は誤作成したレジ締めの特権取消（void/reopen）入力。
type VoidReopenInput struct {
	ID     uint64
	Reason string
	// ActorID は取消実行スタッフ。0 は未認証として fail-closed。
	ActorID uint64
}

// VoidReopenResult は void/reopen の監査付き結果。
type VoidReopenResult struct {
	OriginalCloseID uint64
	ClinicID        uint64
	CloseDate       time.Time
	Period          string
	Reason          string
	VoidedBy        uint64
	VoidedAt        time.Time
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
	// VoidReopen は特権オペレータによる誤作成締めの業務取消。成功後は同一 clinic/date/period を通常 Close できる。
	VoidReopen(ctx context.Context, clinicID uint64, input VoidReopenInput) (*VoidReopenResult, error)
	List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error)
	// #115: 指定日にレジ締めが存在するか確認する（会計の締め後編集チェック用）。
	IsDateClosed(ctx context.Context, clinicID uint64, date time.Time) (bool, error)
}

// periodAggregate は集計処理の共通結果型
type periodAggregate struct {
	Schedule        *sharedkernel.DaySchedule
	PaymentRows     []PaymentAggregateRow
	CategoryRows    []CategoryAggregateRow
	TotalRefund     int64
	BillingDetails  []CloseBillingDetailRow
	TaxBreakdown    []TaxBreakdownRow
	TheoreticalCash int64
	PeriodStart     time.Time
	PeriodEnd       time.Time
	// PaymentMethods は当該 clinic の支払方法マスタ。現金マスタ id 判定（#128）と GetPreview の二重ロード回避に使う。
	PaymentMethods []model.PaymentMethodMaster
	// TaxRates は病院マスタの税率設定。M-7(#191): 締めレジの税率分類を月次レポート経路（exact-match）と統一する。
	TaxRates accountingReportTaxRates
	// UnclassifiedOtherCount は category=other 明細を持つ会計の distinct 件数（DEC-40）。
	UnclassifiedOtherCount int64
	// CategoryPaymentMatrixSystemKey は system_key キーの配賦マトリクス（snapshot 用）。
	CategoryPaymentMatrixSystemKey map[string]map[string]int64
	// CategoryPaymentMatrixNames は表示名キーの配賦マトリクス（preview UI 用）。
	CategoryPaymentMatrixNames map[string]map[string]int64
	// CategoryCounts は部門ごとの会計 distinct 件数。
	CategoryCounts map[string]int64
}

type cashRegisterService struct {
	closeRepo      CashRegisterCloseRepository
	accountingRepo AccountingRepository
	closingsSvc    closingScheduleResolver
	payMethodRepo  PaymentMethodMasterRepository
	clinicRepo     billingClinicReader
}

// NewCashRegisterService は CashRegisterService を初期化して返す
func NewCashRegisterService(
	closeRepo CashRegisterCloseRepository,
	accountingRepo AccountingRepository,
	closingsSvc closingScheduleResolver,
	payMethodRepo PaymentMethodMasterRepository,
	clinicRepo billingClinicReader,
) CashRegisterService {
	return &cashRegisterService{
		closeRepo:      closeRepo,
		accountingRepo: accountingRepo,
		closingsSvc:    closingsSvc,
		payMethodRepo:  payMethodRepo,
		clinicRepo:     clinicRepo,
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

	dateJST := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, config.JST)

	periodStart, periodEnd, err := resolvePeriodRange(dateJST, period, schedule)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch aggregate cash register", "error", err)
		return nil, apperrors.Wrap(err, "failed to fetch aggregate cash register")
	}

	aggregate, err := s.accountingRepo.GetCloseAggregate(ctx, GetCloseAggregateInput{
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

	// M-7(#191): 税率分類は固定閾値ではなく病院マスタ税率（exact-match）で行う（月次レポート経路と統一）。
	clinic, err := s.clinicRepo.FindByID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load clinic for tax rates", "error", err)
		return nil, apperrors.Wrap(err, "failed to load clinic for tax rates")
	}

	// #247 DEC-16⑥: per-billing 配賦（期間全体 payment 比率の擬似按分は禁止）
	allocData, err := s.accountingRepo.GetCategoryPaymentAllocationData(ctx, clinicID, periodStart, periodEnd)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load category payment allocation data", "error", err)
		return nil, apperrors.Wrap(err, "failed to load category payment allocation data")
	}
	sysKeyFn, sysRefundFn := BuildSystemKeyMethodResolvers(payMethods)
	nameFn, nameRefundFn := BuildNameMethodResolvers(payMethods)
	matrixSystemKey := computeCategoryPaymentMatrix(allocData, sysKeyFn, sysRefundFn)
	matrixNames := computeCategoryPaymentMatrix(allocData, nameFn, nameRefundFn)
	categoryCounts := map[string]int64{}
	if allocData != nil && allocData.CategoryCounts != nil {
		categoryCounts = allocData.CategoryCounts
	}

	return &periodAggregate{
		Schedule:                       schedule,
		PaymentRows:                    aggregate.PaymentRows,
		CategoryRows:                   aggregate.CategoryRows,
		TotalRefund:                    aggregate.TotalRefund,
		BillingDetails:                 aggregate.BillingDetails,
		TaxBreakdown:                   aggregate.TaxBreakdown,
		TheoreticalCash:                calcTheoreticalCash(aggregate.PaymentRows, aggregate.TotalRefund, cashMethodID),
		PeriodStart:                    periodStart,
		PeriodEnd:                      periodEnd,
		PaymentMethods:                 payMethods,
		TaxRates:                       clinicTaxRates(clinic),
		UnclassifiedOtherCount:         aggregate.UnclassifiedOtherCount,
		CategoryPaymentMatrixSystemKey: matrixSystemKey,
		CategoryPaymentMatrixNames:     matrixNames,
		CategoryCounts:                 categoryCounts,
	}, nil
}

func (s *cashRegisterService) GetPreview(ctx context.Context, clinicID uint64, dateStr, period string) (*CashRegisterPreview, error) {
	if dateStr == "" {
		return nil, apperrors.WrapInvalidInput("date クエリパラメータは必須です")
	}
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
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

	// #247: per-billing 配賦済みマトリクス（表示名キー）
	categories, err := buildPreviewCategories(agg.CategoryPaymentMatrixNames)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to build preview categories")
	}

	// 個別会計一覧を変換
	details := make([]CloseBillingDetail, 0, len(agg.BillingDetails))
	for _, d := range agg.BillingDetails {
		pmName := paymentMethodNameForClose(d.PaymentMethodID, payMethodNames)
		details = append(details, CloseBillingDetail{
			BillingID:         d.BillingID,
			PaidAt:            timeutil.LocalRFC3339(d.PaidAt),
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
	taxSummary := buildTaxBreakdown(agg.TaxBreakdown, agg.TaxRates)

	return &CashRegisterPreview{
		Date:            date.In(time.Local).Format(time.DateOnly),
		Period:          period,
		PeriodStart:     timeutil.LocalRFC3339(agg.PeriodStart),
		PeriodEnd:       timeutil.LocalRFC3339(agg.PeriodEnd),
		IsAlreadyClosed: isAlreadyClosed,
		IsHoliday:       agg.Schedule != nil && agg.Schedule.IsHoliday,
		Aggregate: CloseAggregateSummary{
			Categories:             categories,
			PaymentMethods:         orderPaymentMethodsForMatrix(payMethods, categories),
			TheoreticalCash:        agg.TheoreticalCash,
			TaxBreakdown:           taxSummary,
			UnclassifiedOtherCount: agg.UnclassifiedOtherCount,
			CategoryCounts:         agg.CategoryCounts,
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

	if agg.Schedule != nil && agg.Schedule.IsHoliday {
		return nil, apperrors.WrapInvalidInput("休診日は締め処理できません")
	}
	if input.ActualCash < 0 {
		return nil, apperrors.WrapInvalidInput("actual_cash は 0 以上で指定してください")
	}

	cashDifference := input.ActualCash - agg.TheoreticalCash

	// category_breakdown JSONB を構築（#247 per-billing 配賦・消費税内訳・未分類件数 snapshot）
	breakdownSchema := buildCategoryBreakdown(agg.CategoryPaymentMatrixSystemKey, agg.TaxBreakdown, agg.TaxRates, agg.UnclassifiedOtherCount)
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
		slog.String("date", input.Date.Format(time.DateOnly)),
		slog.String("period", input.Period),
		slog.Int64("theoretical_cash", agg.TheoreticalCash),
		slog.Int64("actual_cash", input.ActualCash))

	return record, nil
}

func (s *cashRegisterService) VoidReopen(ctx context.Context, clinicID uint64, input VoidReopenInput) (*VoidReopenResult, error) {
	if input.ActorID == 0 {
		return nil, apperrors.WrapUnauthorized("authenticated staff is required to void cash register close")
	}
	if input.ID == 0 {
		return nil, apperrors.WrapInvalidInput("id は必須です")
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, apperrors.WrapInvalidInput("reason は必須です")
	}

	existing, err := s.closeRepo.FindByID(ctx, clinicID, input.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load cash register close for void", "error", err, "id", input.ID)
		return nil, apperrors.Wrap(err, "failed to load cash register close for void")
	}
	if existing == nil {
		return nil, apperrors.WrapNotFound("cash_register_close", fmt.Sprintf("%d", input.ID))
	}

	if err := s.closeRepo.Void(ctx, clinicID, input.ID); err != nil {
		slog.ErrorContext(ctx, "failed to void cash register close", "error", err, "id", input.ID)
		return nil, apperrors.Wrap(err, "failed to void cash register close")
	}

	voidedAt := time.Now()
	result := &VoidReopenResult{
		OriginalCloseID: existing.ID,
		ClinicID:        existing.ClinicID,
		CloseDate:       existing.CloseDate,
		Period:          existing.Period,
		Reason:          reason,
		VoidedBy:        input.ActorID,
		VoidedAt:        voidedAt,
	}

	// 監査: who / why / original id（構造化ログ。永続調整台帳は billing 紐付け必須のためログを正とする）
	slog.InfoContext(ctx, "cash register close voided",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("original_close_id", existing.ID),
		slog.String("close_date", existing.CloseDate.Format(time.DateOnly)),
		slog.String("period", existing.Period),
		slog.Uint64("voided_by", input.ActorID),
		slog.String("reason", reason),
		slog.Time("voided_at", voidedAt),
	)

	return result, nil
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
