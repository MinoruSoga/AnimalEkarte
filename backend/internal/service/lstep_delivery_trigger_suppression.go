package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepDeliveryTriggerService) applySuppression(
	ctx context.Context,
	clinicID, ownerID uint64,
	triggerType string,
	asOf time.Time,
) (suppressed bool, err error) {
	if s.prioritySvc == nil {
		return false, nil
	}
	existing, err := s.triggerLogRepo.FindByOwnerAndDate(ctx, clinicID, ownerID, asOf)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to find existing trigger logs")
	}
	// SuppressedByPriority=true は既に抑制済みなので除外
	active := make([]model.LstepDeliveryTriggerLog, 0, len(existing))
	for i := range existing {
		if !existing[i].SuppressedByPriority {
			active = append(active, existing[i])
		}
	}
	if len(active) == 0 {
		return false, nil
	}

	currentPri, err := s.prioritySvc.GetPriorityFor(ctx, clinicID, triggerType)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to get priority for current trigger")
	}

	// 既存の中で最高優先度（最小値）を持つログを探す
	bestPri := int(^uint(0) >> 1) // MaxInt
	var bestLog *model.LstepDeliveryTriggerLog
	for i := range active {
		l := &active[i]
		p, pErr := s.prioritySvc.GetPriorityFor(ctx, clinicID, l.TriggerType)
		if pErr != nil {
			return false, apperrors.Wrap(pErr, "failed to get priority for existing trigger")
		}
		if p < bestPri {
			bestPri = p
			bestLog = l
		}
	}

	if currentPri > bestPri {
		// 現在のトリガーは既存より低優先度 → 抑制
		return true, nil
	}
	if currentPri < bestPri {
		// 現在のトリガーは既存より高優先度 → 全既存を降格
		for i := range active {
			l := &active[i]
			reason := fmt.Sprintf("superseded by %s (priority %d < %d)", triggerType, currentPri, bestPri)
			if suppErr := s.triggerLogRepo.UpdateSuppressed(ctx, l.ID, reason); suppErr != nil {
				slog.ErrorContext(ctx, "delivery trigger: failed to suppress existing log", "log_id", l.ID, "error", suppErr)
				return false, apperrors.Wrap(suppErr, "failed to suppress existing trigger log")
			}
		}
		_ = bestLog // suppress "declared but not used" — bestLog is used only for logging context above
	}
	// 同優先度または降格完了 → 現在のトリガーは通常通り発火
	return false, nil
}

// runBatch は ownerID リストに対して除外チェック・重複チェック・タグ付与・ログ記録を行う汎用バッチ実行。
