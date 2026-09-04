package billing

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// GetMonthlyUnpaidCarryover は対象月の未納繰越（前月繰越・当月未払い・次月繰越）を
// 飼主+ペット単位で返す。#114
func (s *accountingService) GetMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, year, month, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
	if month < 1 || month > 12 {
		return nil, 0, MonthlyUnpaidSummary{}, apperrors.WrapInvalidInput("month must be between 1 and 12")
	}
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, config.JST).Format(time.DateOnly)
	lastDay := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, config.JST).Format(time.DateOnly)

	items, total, summary, err := s.repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay, lastDay, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get monthly unpaid carryover", "error", err)
		return nil, 0, summary, apperrors.Wrap(err, "failed to get monthly unpaid carryover")
	}
	return items, total, summary, nil
}

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
func (s *accountingService) ListUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
	result, total, summary, err := s.repo.FindUnpaidByOwner(ctx, clinicID, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list unpaid by owner", "error", err)
		return nil, 0, summary, apperrors.Wrap(err, "failed to list unpaid by owner")
	}
	return result, total, summary, nil
}

// GetOwnerUnpaidBalance は会計画面表示用に飼主の未納残高を返す。#182
func (s *accountingService) GetOwnerUnpaidBalance(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error) {
	if ownerID == 0 {
		return OwnerUnpaidBalance{}, apperrors.WrapInvalidInput("owner_id is required")
	}
	result, err := s.repo.SumUnpaidByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get owner unpaid balance", "clinic_id", clinicID, "owner_id", ownerID, "error", err)
		return OwnerUnpaidBalance{}, apperrors.Wrap(err, "failed to get owner unpaid balance")
	}
	return result, nil
}

// Cancel は会計を論理削除（status=cancelled）する。
// BUG-371 / #118: ハード削除の代替。actorID で監査ログを記録する。
// BE-refactor.md R1-2: 更新+監査を同一 tx で原子化する（fail-closed。#211/refund_service パターン踏襲）。
// 監査書込が失敗すると tx がロールバックし、status=cancelled への更新も無効になる。
func (s *accountingService) Cancel(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find accounting for cancel")
	}
	// 既に cancelled 状態なら二重キャンセル防止（AC-12）
	if existing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict("既にキャンセル済みの会計です")
	}

	billingID := id
	aType := sharedkernel.AuditActorTypeFor(actorID)

	if s.transactor == nil {
		return apperrors.WrapInternalServerError("billing transactor is required for billing cancel")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.lockCloseBoundaries(txCtx, clinicID, existing.ScheduledDate); err != nil {
			return err
		}
		fields := map[string]any{"status": model.BillingStatusCancelled}
		if _, err := s.repo.Update(txCtx, clinicID, id, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to cancel accounting", "error", err, "billing_id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to cancel accounting")
		}

		slog.InfoContext(txCtx, "billing cancelled",
			slog.Uint64("billing_id", id),
			slog.Uint64("clinic_id", clinicID))

		// #118 / BIL-02: cancel audit is fail-closed (missing dependency rejects the write).
		if s.auditTx == nil {
			return apperrors.WrapInternalServerError("billing audit dependency is required for billing cancel")
		}
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicID,
			ActorID:    actorID,
			ActorType:  aType,
			Action:     model.AuditActionBillingCancel,
			Resource:   "billing",
			ResourceID: &billingID,
			OldValue:   map[string]any{"status": string(existing.Status)},
			NewValue:   map[string]any{"status": string(model.BillingStatusCancelled)},
		}); err != nil {
			return apperrors.Wrap(err, "failed to write billing cancel audit log")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to cancel accounting in transaction")
	}
	return nil
}

// parseDailySummaryDate は dateStr（空なら今日）を JST でパースする共通ヘルパー。
func parseDailySummaryDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		dateStr = time.Now().In(time.Local).Format(time.DateOnly)
	}
	date, err := time.ParseInLocation(time.DateOnly, dateStr, config.JST)
	if err != nil {
		return time.Time{}, apperrors.WrapInvalidInput("date must be YYYY-MM-DD")
	}
	return date, nil
}

// GetDailySummary は指定日のレジ締め集計を返す。BUG-368
func (s *accountingService) GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*DailySummaryResult, error) {
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
