package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// MonthlyDeliveryStats はクリニック単位の月次配信集計 DTO。
type MonthlyDeliveryStats struct {
	YearMonth string                        `json:"year_month"`
	Rows      []repository.DeliveryStatsRow `json:"rows"`
}

// LstepAnalyticsService は Lステップ分析データ取得サービス（FEAT-385）。
type LstepAnalyticsService interface {
	// GetDeliveryHistoryByOwner は飼主単位で期間内の配信トリガーログ一覧を返す。
	GetDeliveryHistoryByOwner(ctx context.Context, clinicID, ownerID uint64, from, to time.Time) ([]model.LstepDeliveryTriggerLog, error)
	// GetMonthlyDeliveryStats は月次でトリガー種別 × ステータス別集計を返す（クリニック全体）。yearMonth は "2006-01" 形式。
	GetMonthlyDeliveryStats(ctx context.Context, clinicID uint64, yearMonth string) (*MonthlyDeliveryStats, error)
	// GetLatestFriendAttributes は飼主 ID から最新の Lステップ友だち属性スナップショットを返す。
	GetLatestFriendAttributes(ctx context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error)
}

type lstepAnalyticsService struct {
	ownerRepo      repository.OwnerRepository
	triggerLogRepo repository.LstepDeliveryTriggerLogRepository
	snapshotRepo   repository.LstepFriendAttributeSnapshotRepository
}

// NewLstepAnalyticsService は LstepAnalyticsService を初期化して返す。
func NewLstepAnalyticsService(
	ownerRepo repository.OwnerRepository,
	triggerLogRepo repository.LstepDeliveryTriggerLogRepository,
	snapshotRepo repository.LstepFriendAttributeSnapshotRepository,
) LstepAnalyticsService {
	return &lstepAnalyticsService{
		ownerRepo:      ownerRepo,
		triggerLogRepo: triggerLogRepo,
		snapshotRepo:   snapshotRepo,
	}
}

func (s *lstepAnalyticsService) GetDeliveryHistoryByOwner(ctx context.Context, clinicID, ownerID uint64, from, to time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	logs, err := s.triggerLogRepo.ListByOwnerAndDateRange(ctx, clinicID, ownerID, from, to)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list delivery trigger logs by owner", "error", err, "owner_id", ownerID)
		return nil, apperrors.Wrap(err, "failed to list delivery history by owner")
	}
	return logs, nil
}

func (s *lstepAnalyticsService) GetMonthlyDeliveryStats(ctx context.Context, clinicID uint64, yearMonth string) (*MonthlyDeliveryStats, error) {
	t, err := time.ParseInLocation("2006-01", yearMonth, jst)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid year_month format: " + yearMonth)
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, jst)
	until := from.AddDate(0, 1, 0)

	rows, err := s.triggerLogRepo.CountByTypeAndStatus(ctx, clinicID, from, until)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count monthly delivery stats", "error", err, "year_month", yearMonth)
		return nil, apperrors.Wrap(err, "failed to get monthly delivery stats")
	}
	return &MonthlyDeliveryStats{YearMonth: yearMonth, Rows: rows}, nil
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
