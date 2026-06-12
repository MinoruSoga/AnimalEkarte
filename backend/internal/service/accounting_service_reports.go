package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// #120: start_date/end_date 2引数
func (s *accountingService) ListUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindUnpaidByBilling(ctx, clinicID, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list unpaid billings", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list unpaid billings")
	}
	return result, total, nil
}

// #120: start_date/end_date 2引数（飼主単位集約）
func (s *accountingService) ListUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	result, total, summary, err := s.repo.FindUnpaidByOwner(ctx, clinicID, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list unpaid by owner", "error", err)
		return nil, 0, summary, apperrors.Wrap(err, "failed to list unpaid by owner")
	}
	return result, total, summary, nil
}

// Cancel は会計を論理削除（status=cancelled）する。
// BUG-371 / #118: ハード削除の代替。actorID で監査ログを記録する。
func (s *accountingService) Cancel(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find accounting for cancel")
	}
	// 既に cancelled 状態なら二重キャンセル防止（AC-12）
	if existing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict("既にキャンセル済みの会計です")
	}

	fields := map[string]any{"status": model.BillingStatusCancelled}
	if _, err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return apperrors.Wrap(err, "failed to cancel accounting")
	}

	slog.InfoContext(ctx, "billing cancelled",
		slog.Uint64("billing_id", id),
		slog.Uint64("clinic_id", clinicID))

	// #118: 監査ログ（ベストエフォート）
	if s.auditSvc != nil {
		billingID := id
		aType := "system"
		if actorID != nil {
			aType = "staff"
		}
		if logErr := s.auditSvc.LogEntry(ctx, &AuditLogInput{
			ClinicID:   &clinicID,
			ActorID:    actorID,
			ActorType:  aType,
			Action:     "cancel",
			Resource:   "billing",
			ResourceID: &billingID,
		}); logErr != nil {
			slog.ErrorContext(ctx, "audit log failed for billing cancel", "error", logErr, "billing_id", id)
		}
	}

	return nil
}

// parseDailySummaryDate は dateStr（空なら今日）を JST でパースする共通ヘルパー。
func parseDailySummaryDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		dateStr = time.Now().In(time.Local).Format("2006-01-02")
	}
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	date, err := time.ParseInLocation("2006-01-02", dateStr, jst)
	if err != nil {
		return time.Time{}, apperrors.WrapInvalidInput("date must be YYYY-MM-DD")
	}
	return date, nil
}

// GetDailySummary は指定日のレジ締め集計を返す。BUG-368
func (s *accountingService) GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*repository.DailySummaryResult, error) {
	date, err := parseDailySummaryDate(dateStr)
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetDailySummary(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get daily summary", "error", err)
		return nil, apperrors.Wrap(err, "failed to get daily summary")
	}
	return result, nil
}

// GetDailySummaryForClinics は複数医院の拠点別日次集計を返す (#86 段階3 論点4=2)。
func (s *accountingService) GetDailySummaryForClinics(ctx context.Context, clinicIDs []uint64, dateStr string) ([]ClinicDailySummary, error) {
	date, err := parseDailySummaryDate(dateStr)
	if err != nil {
		return nil, err
	}
	results := make([]ClinicDailySummary, 0, len(clinicIDs))
	for _, clinicID := range clinicIDs {
		r, rerr := s.repo.GetDailySummary(ctx, clinicID, date)
		if rerr != nil {
			slog.ErrorContext(ctx, "failed to get daily summary for clinic", "clinic_id", clinicID, "error", rerr)
			return nil, apperrors.Wrap(rerr, "failed to get daily summary for clinics")
		}
		results = append(results, ClinicDailySummary{ClinicID: clinicID, Summary: r})
	}
	return results, nil
}
