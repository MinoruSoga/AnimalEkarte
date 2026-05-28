package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

func (s *accountingService) ListUnpaidByBilling(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindUnpaidByBilling(ctx, clinicID, baseDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list unpaid billings", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list unpaid billings")
	}
	return result, total, nil
}

// BUG-370: 月末未納者一覧（飼主単位集約）
func (s *accountingService) ListUnpaidByOwner(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	result, total, summary, err := s.repo.FindUnpaidByOwner(ctx, clinicID, baseDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list unpaid by owner", "error", err)
		return nil, 0, summary, apperrors.Wrap(err, "failed to list unpaid by owner")
	}
	return result, total, summary, nil
}

// Cancel は会計を論理削除（status=cancelled）する。
// BUG-371: ハード削除の代替。監査性のため物理削除しない。
func (s *accountingService) Cancel(ctx context.Context, clinicID, id uint64) error {
	// 既存値取得
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find accounting for cancel")
	}
	// 既に cancelled 状態なら二重キャンセル防止（AC-12）
	if existing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict("既にキャンセル済みの会計です")
	}

	fields := map[string]any{
		"status": model.BillingStatusCancelled,
	}
	if _, err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return apperrors.Wrap(err, "failed to cancel accounting")
	}

	slog.InfoContext(ctx, "billing cancelled",
		slog.Uint64("billing_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}

// GetDailySummary は指定日のレジ締め集計を返す。BUG-368
func (s *accountingService) GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*repository.DailySummaryResult, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	date, err := time.ParseInLocation("2006-01-02", dateStr, jst)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD")
	}
	result, err := s.repo.GetDailySummary(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get daily summary", "error", err)
		return nil, apperrors.Wrap(err, "failed to get daily summary")
	}
	return result, nil
}
