package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

const maxVisitConversionDays = 365

// MonthlyDeliveryStats はクリニック単位の月次配信集計 DTO。
type MonthlyDeliveryStats struct {
	YearMonth string             `json:"year_month"`
	Rows      []DeliveryStatsRow `json:"rows"`
}

// VisitConversionSummary は月次の配信後来院率集計 DTO。
type VisitConversionSummary struct {
	YearMonth      string                  `json:"year_month"`
	Days           int                     `json:"days"`
	DeliveredCount int64                   `json:"delivered_count"`
	VisitedCount   int64                   `json:"visited_count"`
	VisitRate      float64                 `json:"visit_rate"`
	Rows           []VisitConversionByType `json:"rows"`
}

// VisitConversionByType はトリガー種別ごとの来院率行。
type VisitConversionByType struct {
	TriggerType    string  `json:"trigger_type"`
	DeliveredCount int64   `json:"delivered_count"`
	VisitedCount   int64   `json:"visited_count"`
	VisitRate      float64 `json:"visit_rate"`
}

// LstepAnalyticsService は Lステップ分析データ取得サービス（FEAT-385）。
type LstepAnalyticsService interface {
	// GetMonthlyDeliveryStats は月次でトリガー種別 × ステータス別集計を返す（クリニック全体）。yearMonth は "2006-01" 形式。
	GetMonthlyDeliveryStats(ctx context.Context, clinicID uint64, yearMonth string) (*MonthlyDeliveryStats, error)
	// GetVisitConversionSummary は月次で配信後来院率を返す。days は配信後来院の判定窓（日数）。
	GetVisitConversionSummary(ctx context.Context, clinicID uint64, yearMonth string, days int) (*VisitConversionSummary, error)
	// GetLatestFriendAttributes は飼主 ID から最新の Lステップ友だち属性スナップショットを返す。
	GetLatestFriendAttributes(ctx context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error)
}

type lstepAnalyticsService struct {
	ownerRepo      analyticsOwnerRepository
	triggerLogRepo analyticsTriggerLogRepository
	snapshotRepo   analyticsSnapshotRepository
}

type analyticsOwnerRepository interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

type analyticsTriggerLogRepository interface {
	CountByTypeAndStatus(ctx context.Context, clinicID uint64, from, to time.Time) ([]DeliveryStatsRow, error)
	CountVisitConversionsByType(ctx context.Context, clinicID uint64, from, to time.Time, days int) ([]VisitConversionRow, error)
}

type analyticsSnapshotRepository interface {
	FindLatestByOwner(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error)
}

// NewLstepAnalyticsService は LstepAnalyticsService を初期化して返す。
func NewLstepAnalyticsService(
	ownerRepo analyticsOwnerRepository,
	triggerLogRepo analyticsTriggerLogRepository,
	snapshotRepo analyticsSnapshotRepository,
) LstepAnalyticsService {
	return &lstepAnalyticsService{
		ownerRepo:      ownerRepo,
		triggerLogRepo: triggerLogRepo,
		snapshotRepo:   snapshotRepo,
	}
}

func (s *lstepAnalyticsService) GetMonthlyDeliveryStats(ctx context.Context, clinicID uint64, yearMonth string) (*MonthlyDeliveryStats, error) {
	t, err := time.ParseInLocation("2006-01", yearMonth, config.JST)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid year_month format: " + yearMonth)
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, config.JST)
	until := from.AddDate(0, 1, 0)

	rows, err := s.triggerLogRepo.CountByTypeAndStatus(ctx, clinicID, from, until)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count monthly delivery stats", "error", err, "year_month", yearMonth)
		return nil, apperrors.Wrap(err, "failed to get monthly delivery stats")
	}
	return &MonthlyDeliveryStats{YearMonth: yearMonth, Rows: rows}, nil
}

func (s *lstepAnalyticsService) GetVisitConversionSummary(ctx context.Context, clinicID uint64, yearMonth string, days int) (*VisitConversionSummary, error) {
	if days < 1 || days > maxVisitConversionDays {
		return nil, apperrors.WrapInvalidInput("days は 1 以上 365 以下で指定してください")
	}
	t, err := time.ParseInLocation("2006-01", yearMonth, config.JST)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid year_month format: " + yearMonth)
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, config.JST)
	until := from.AddDate(0, 1, 0)

	rows, err := s.triggerLogRepo.CountVisitConversionsByType(ctx, clinicID, from, until, days)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count monthly visit conversion summary", "error", err, "year_month", yearMonth, "days", days)
		return nil, apperrors.Wrap(err, "failed to get visit conversion summary")
	}

	result := &VisitConversionSummary{
		YearMonth: yearMonth,
		Days:      days,
		Rows:      make([]VisitConversionByType, len(rows)),
	}
	for i, row := range rows {
		result.DeliveredCount += row.DeliveredCount
		result.VisitedCount += row.VisitedCount
		result.Rows[i] = VisitConversionByType{
			TriggerType:    row.TriggerType,
			DeliveredCount: row.DeliveredCount,
			VisitedCount:   row.VisitedCount,
			VisitRate:      calculateRate(row.VisitedCount, row.DeliveredCount),
		}
	}
	result.VisitRate = calculateRate(result.VisitedCount, result.DeliveredCount)
	return result, nil
}

func (s *lstepAnalyticsService) GetLatestFriendAttributes(ctx context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error) {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil, apperrors.WrapNotFound("owner_line_user_id", fmt.Sprintf("%d", ownerID))
	}
	snapshot, err := s.snapshotRepo.FindLatestByOwner(ctx, clinicID, *owner.LineUserID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find latest friend attributes")
	}
	return snapshot, nil
}

func calculateRate(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
