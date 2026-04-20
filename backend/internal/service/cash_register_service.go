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

// CashRegisterPreview は締めプレビューの結果
type CashRegisterPreview struct {
	Date            string                           `json:"date"`
	Period          string                           `json:"period"`
	Schedule        *DaySchedule                     `json:"schedule"`
	AggregateRows   []repository.BillingAggregateRow `json:"aggregate_rows"`
	BillingDetails  []repository.CloseBillingDetail  `json:"billing_details"`
	TheoreticalCash int64                            `json:"theoretical_cash"`
}

// CloseRegisterInput はレジ締め実行の入力
type CloseRegisterInput struct {
	Date       time.Time
	Period     string
	ActualCash int64
	Memo       string
	ClosedBy   *uint64
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

func (s *cashRegisterService) GetPreview(ctx context.Context, clinicID uint64, date time.Time, period string) (*CashRegisterPreview, error) {
	if period != "am" && period != "pm" {
		return nil, apperrors.WrapInvalidInput("period は 'am' または 'pm' を指定してください")
	}

	schedule, err := s.closingsSvc.ResolveSchedule(ctx, clinicID, date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to resolve schedule")
	}

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	dateJST := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jst)

	periodStart, periodEnd, err := resolvePeriodRange(dateJST, period, schedule, jst)
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

	// 理論現金（現金支払いのみの純額）
	theoreticalCash := calcTheoreticalCash(aggregate.AggregateRows)

	return &CashRegisterPreview{
		Date:            date.Format("2006-01-02"),
		Period:          period,
		Schedule:        schedule,
		AggregateRows:   aggregate.AggregateRows,
		BillingDetails:  aggregate.BillingDetails,
		TheoreticalCash: theoreticalCash,
	}, nil
}

func (s *cashRegisterService) Close(ctx context.Context, clinicID uint64, input CloseRegisterInput) (*model.CashRegisterClose, error) {
	if input.Period != "am" && input.Period != "pm" {
		return nil, apperrors.WrapInvalidInput("period は 'am' または 'pm' を指定してください")
	}

	// 二重締めチェック
	existing, err := s.closeRepo.FindByDateAndPeriod(ctx, clinicID, input.Date, input.Period)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check existing close")
	}
	if existing != nil {
		return nil, apperrors.WrapConflict("この日時はすでに締め済みです")
	}

	schedule, err := s.closingsSvc.ResolveSchedule(ctx, clinicID, input.Date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to resolve schedule")
	}

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	dateJST := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, jst)

	periodStart, periodEnd, err := resolvePeriodRange(dateJST, input.Period, schedule, jst)
	if err != nil {
		return nil, err
	}

	aggregate, err := s.accountingRepo.GetCloseAggregate(ctx, repository.GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate billings for close")
	}

	theoreticalCash := calcTheoreticalCash(aggregate.AggregateRows)
	cashDifference := input.ActualCash - theoreticalCash

	// category_breakdown JSONB を構築
	breakdownSchema := buildCategoryBreakdown(aggregate.AggregateRows)
	breakdownJSON, err := json.Marshal(breakdownSchema)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to marshal category breakdown")
	}

	record := &model.CashRegisterClose{
		ClinicID:          clinicID,
		CloseDate:         input.Date,
		Period:            input.Period,
		TheoreticalCash:   theoreticalCash,
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
		slog.Int64("theoretical_cash", theoreticalCash),
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
func resolvePeriodRange(dateJST time.Time, period string, schedule *DaySchedule, jst *time.Location) (start, end time.Time, err error) {
	boundaryH, boundaryM, parseErr := parseHHMM(schedule.AmPmBoundary)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("am_pm_boundary の形式が正しくありません")
	}
	pmEndH, pmEndM, parseErr := parseHHMM(schedule.PmEnd)
	if parseErr != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("pm_end の形式が正しくありません")
	}

	boundary := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), boundaryH, boundaryM, 0, 0, jst)
	pmEnd := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), pmEndH, pmEndM, 0, 0, jst)
	dayStart := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, jst)

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
