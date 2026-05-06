package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// GetDeliveryMonitorSummaryInput は GetSummary の入力パラメータ。
type GetDeliveryMonitorSummaryInput struct {
	ClinicID    uint64
	From        time.Time
	To          time.Time
	TriggerType string
}

// GetDeliveryMonitorLogsInput は GetLogs の入力パラメータ。
type GetDeliveryMonitorLogsInput struct {
	ClinicID    uint64
	From        time.Time
	To          time.Time
	TriggerType string
	Status      string
	Page        int
	PerPage     int
}

// DeliveryTriggerSummary は配信トリガーのステータス別集計結果。
type DeliveryTriggerSummary struct {
	Scheduled               int64
	Fired                   int64
	Excluded                int64
	Failed                  int64
	ExcludedReasonBreakdown map[string]int64
}

// DeliveryTriggerLogItem は配信トリガーログの1件。
type DeliveryTriggerLogItem struct {
	ID             uint64
	OwnerID        uint64
	OwnerName      string
	TriggerType    string
	ScheduledAt    time.Time
	Status         string
	FiredAt        *time.Time
	ExcludedReason *string
}

// DeliveryTriggerLogsPage は配信トリガーログのページネーション結果。
type DeliveryTriggerLogsPage struct {
	Items   []DeliveryTriggerLogItem
	Total   int64
	Page    int
	PerPage int
}

// LstepDeliveryMonitorService は自動配信トリガー監視のサービスインターフェース。
type LstepDeliveryMonitorService interface {
	GetSummary(ctx context.Context, input GetDeliveryMonitorSummaryInput) (DeliveryTriggerSummary, error)
	GetLogs(ctx context.Context, input *GetDeliveryMonitorLogsInput) (DeliveryTriggerLogsPage, error)
}

type lstepDeliveryMonitorService struct {
	triggerLog repository.LstepDeliveryTriggerLogRepository
}

// NewLstepDeliveryMonitorService は LstepDeliveryMonitorService を初期化して返す。
func NewLstepDeliveryMonitorService(triggerLog repository.LstepDeliveryTriggerLogRepository) LstepDeliveryMonitorService {
	return &lstepDeliveryMonitorService{triggerLog: triggerLog}
}

func (s *lstepDeliveryMonitorService) GetSummary(ctx context.Context, input GetDeliveryMonitorSummaryInput) (DeliveryTriggerSummary, error) {
	statusCounts, err := s.triggerLog.CountByStatusAndDateRange(ctx, input.ClinicID, input.From, input.To, input.TriggerType)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count delivery trigger log by status", "clinic_id", input.ClinicID, "error", err)
		return DeliveryTriggerSummary{}, apperrors.Wrap(err, "failed to count delivery trigger log by status")
	}
	excludedReasons, err := s.triggerLog.CountExcludedReasonByDateRange(ctx, input.ClinicID, input.From, input.To, input.TriggerType)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count delivery trigger excluded reasons", "clinic_id", input.ClinicID, "error", err)
		return DeliveryTriggerSummary{}, apperrors.Wrap(err, "failed to count delivery trigger excluded reasons")
	}
	return DeliveryTriggerSummary{
		Scheduled:               statusCounts[model.TriggerStatusScheduled],
		Fired:                   statusCounts[model.TriggerStatusFired],
		Excluded:                statusCounts[model.TriggerStatusExcluded],
		Failed:                  statusCounts[model.TriggerStatusFailed],
		ExcludedReasonBreakdown: excludedReasons,
	}, nil
}

func (s *lstepDeliveryMonitorService) GetLogs(ctx context.Context, input *GetDeliveryMonitorLogsInput) (DeliveryTriggerLogsPage, error) {
	if input == nil {
		return DeliveryTriggerLogsPage{}, apperrors.WrapInvalidInput("input is nil")
	}
	perPage := input.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	page := input.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	rows, total, err := s.triggerLog.FindByDateRangeWithFilters(ctx, input.ClinicID, input.From, input.To, input.TriggerType, input.Status, perPage, offset)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find delivery trigger logs", "clinic_id", input.ClinicID, "error", err)
		return DeliveryTriggerLogsPage{}, apperrors.Wrap(err, "failed to find delivery trigger logs")
	}

	items := make([]DeliveryTriggerLogItem, len(rows))
	for i := range rows {
		r := &rows[i]
		items[i] = DeliveryTriggerLogItem{
			ID:             r.ID,
			OwnerID:        r.OwnerID,
			OwnerName:      r.OwnerName,
			TriggerType:    r.TriggerType,
			ScheduledAt:    r.ScheduledAt,
			Status:         r.Status,
			FiredAt:        r.FiredAt,
			ExcludedReason: r.ExcludedReason,
		}
	}
	return DeliveryTriggerLogsPage{Items: items, Total: total, Page: page, PerPage: perPage}, nil
}
