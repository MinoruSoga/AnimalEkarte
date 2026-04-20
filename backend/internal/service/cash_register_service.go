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

// CashRegisterService はレジ締めのビジネスロジックインターフェース
type CashRegisterService interface {
	GetPreview(ctx context.Context, clinicID uint64, date time.Time, period string) (*CashRegisterPreview, error)
	Close(ctx context.Context, clinicID uint64, input CloseRegisterInput) (*model.CashRegisterClose, error)
	List(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error)
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

// CloseRegisterInput はレジ締め実行の入力
type CloseRegisterInput struct {
	Date       time.Time
	Period     string
	ActualCash int64
	Memo       string
	ClosedBy   *uint64
}

// periodAggregate は集計処理の共通結果型
type periodAggregate struct {
	Schedule        *DaySchedule
	AggregateRows   []repository.BillingAggregateRow
	BillingDetails  []repository.CloseBillingDetail
	TaxBreakdown    []repository.TaxBreakdownRow
	TheoreticalCash int64
	PeriodStart     time.Time
	PeriodEnd       time.Time
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

// validatePeriod は period 値（"am"/"pm"）のバリデーションを行う
func validatePeriod(period string) error {
	if period != "am" && period != "pm" {
		return apperrors.WrapInvalidInput("period は 'am' または 'pm' を指定してください")
	}
	return nil
}

// fetchAggregate は date/period から集計処理（スケジュール取得・期間算出・会計集計）を実行する共通ロジック
func (s *cashRegisterService) fetchAggregate(ctx context.Context, clinicID uint64, date time.Time, period string) (*periodAggregate, error) {
	schedule, err := s.closingsSvc.ResolveSchedule(ctx, clinicID, date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to resolve schedule")
	}

	dateJST := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jstLocation())

	periodStart, periodEnd, err := resolvePeriodRange(dateJST, period, schedule)
	if err != nil {
		return nil, err
	}

	aggregate, err := s.accountingRepo.GetCloseAggregate(ctx, repository.GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate billings")
	}

	return &periodAggregate{
		Schedule:        schedule,
		AggregateRows:   aggregate.AggregateRows,
		BillingDetails:  aggregate.BillingDetails,
		TaxBreakdown:    aggregate.TaxBreakdown,
		TheoreticalCash: calcTheoreticalCash(aggregate.AggregateRows),
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
	}, nil
}

func (s *cashRegisterService) GetPreview(ctx context.Context, clinicID uint64, date time.Time, period string) (*CashRegisterPreview, error) {
	if err := validatePeriod(period); err != nil {
		return nil, err
	}

	agg, err := s.fetchAggregate(ctx, clinicID, date, period)
	if err != nil {
		return nil, err
	}

	// 二重締め確認
	existing, err := s.closeRepo.FindByDateAndPeriod(ctx, clinicID, date, period)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check existing close")
	}
	isAlreadyClosed := existing != nil

	// 支払方法マスタを取得
	payMethods, err := s.payMethodRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get payment methods")
	}
	payMethodNames := buildPayMethodNameMap(payMethods)

	// カテゴリ別×支払方法名別集計マップを構築
	categories := make(map[string]map[string]int64)
	for _, row := range agg.AggregateRows {
		if categories[row.Category] == nil {
			categories[row.Category] = make(map[string]int64)
		}
		pmName := paymentMethodNameForClose(row.PaymentMethodID, payMethodNames)
		categories[row.Category][pmName] += row.NetAmount
	}

	// 個別会計一覧を変換
	details := make([]CloseBillingDetail, 0, len(agg.BillingDetails))
	for _, d := range agg.BillingDetails {
		pmName := paymentMethodNameForClose(d.PaymentMethodID, payMethodNames)
		details = append(details, CloseBillingDetail{
			BillingID:         d.BillingID,
			PaidAt:            d.PaidAt.Format(time.RFC3339),
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
		Date:            date.Format("2006-01-02"),
		Period:          period,
		PeriodStart:     agg.PeriodStart.Format(time.RFC3339),
		PeriodEnd:       agg.PeriodEnd.Format(time.RFC3339),
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
		return nil, err
	}

	// 二重締めチェック
	existing, err := s.closeRepo.FindByDateAndPeriod(ctx, clinicID, input.Date, input.Period)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check existing close")
	}
	if existing != nil {
		return nil, apperrors.WrapConflict("この日時はすでに締め済みです")
	}

	agg, err := s.fetchAggregate(ctx, clinicID, input.Date, input.Period)
	if err != nil {
		return nil, err
	}

	cashDifference := input.ActualCash - agg.TheoreticalCash

	// category_breakdown JSONB を構築
	breakdownSchema := buildCategoryBreakdown(agg.AggregateRows)
	breakdownJSON, err := json.Marshal(breakdownSchema)
	if err != nil {
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
	return s.closeRepo.FindAll(ctx, clinicID, startDate, endDate, page, limit)
}

func (s *cashRegisterService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
	return s.closeRepo.FindByID(ctx, clinicID, id)
}

// resolvePeriodRange は period（"am"/"pm"）と DaySchedule から集計期間（JST）を返す
func resolvePeriodRange(dateJST time.Time, period string, schedule *DaySchedule) (start, end time.Time, err error) {
	boundaryH, boundaryM, parseErr := parseHHMM(schedule.AmPmBoundary)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("am_pm_boundary の形式が正しくありません")
	}
	pmEndH, pmEndM, parseErr := parseHHMM(schedule.PmEnd)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("pm_end の形式が正しくありません")
	}

	boundary := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), boundaryH, boundaryM, 0, 0, jstLocation())
	pmEnd := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), pmEndH, pmEndM, 0, 0, jstLocation())
	dayStart := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, jstLocation())

	switch period {
	case "am":
		return dayStart, boundary, nil
	case "pm":
		return boundary, pmEnd, nil
	default:
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("period は 'am' または 'pm' を指定してください")
	}
}

// parseHHMM は "HH:MM" 形式の時刻文字列を時・分に分解する
func parseHHMM(s string) (h, m int, err error) {
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

// calcTheoreticalCash は集計行から現金扱い（payment_method_id が nil）の純額を合計する
// payment_method_id が nil のものを現金として扱う（旧来の method="cash" 相当）
func calcTheoreticalCash(rows []repository.BillingAggregateRow) int64 {
	var total int64
	for _, r := range rows {
		if r.PaymentMethodID == nil {
			total += r.NetAmount
		}
	}
	return total
}

// buildCategoryBreakdown は集計行から CategoryBreakdownSchema を構築する
func buildCategoryBreakdown(rows []repository.BillingAggregateRow) model.CategoryBreakdownSchema {
	cats := make(map[string]map[string]int64)
	for _, r := range rows {
		if cats[r.Category] == nil {
			cats[r.Category] = make(map[string]int64)
		}
		key := "cash"
		if r.PaymentMethodID != nil {
			key = fmt.Sprintf("method_%d", *r.PaymentMethodID)
		}
		cats[r.Category][key] += r.NetAmount
	}
	return model.CategoryBreakdownSchema{
		Categories: cats,
	}
}
