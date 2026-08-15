package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
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
	for i := range active {
		l := &active[i]
		p, pErr := s.prioritySvc.GetPriorityFor(ctx, clinicID, l.TriggerType)
		if pErr != nil {
			return false, apperrors.Wrap(pErr, "failed to get priority for existing trigger")
		}
		if p < bestPri {
			bestPri = p
		}
	}

	if currentPri > bestPri {
		// 現在のトリガーは既存より低優先度 → 抑制
		return true, nil
	}
	if currentPri < bestPri {
		// 現在のトリガーは既存より高優先度 → 低優先ログを降格し、既付与タグを取り消す (G2B-03)。
		// tagName は発火時と同じく trigger_type 文字列（runBatch が同一値を AddTag に渡す）。
		client, lineUserID, remErr := s.demoteRemoveTarget(ctx, clinicID, ownerID)
		if remErr != nil {
			return false, remErr
		}
		for i := range active {
			l := &active[i]
			if client != nil && lineUserID != "" {
				if removeErr := client.RemoveTag(ctx, lineUserID, l.TriggerType); removeErr != nil {
					slog.ErrorContext(ctx, "delivery trigger: failed to remove demoted tag",
						"log_id", l.ID, "trigger", l.TriggerType, "error", removeErr)
					// fail-closed: do not fire the higher-priority trigger if exclusivity cannot be restored
					return false, apperrors.Wrap(removeErr, "failed to remove demoted trigger tag")
				}
			}
			reason := fmt.Sprintf("superseded by %s (priority %d < %d)", triggerType, currentPri, bestPri)
			if suppErr := s.triggerLogRepo.UpdateSuppressed(ctx, clinicID, l.ID, reason); suppErr != nil {
				slog.ErrorContext(ctx, "delivery trigger: failed to suppress existing log", "log_id", l.ID, "error", suppErr)
				return false, apperrors.Wrap(suppErr, "failed to suppress existing trigger log")
			}
		}
	}
	// 同優先度または降格完了 → 現在のトリガーは通常通り発火
	return false, nil
}

// demoteRemoveTarget resolves LSTEP client + line user for exclusivity tag revoke on demote.
// nil client means LSTEP write path is disabled (sync off / no key) — log demote only.
func (s *lstepDeliveryTriggerService) demoteRemoveTarget(
	ctx context.Context,
	clinicID, ownerID uint64,
) (client lstep.Client, lineUserID string, err error) {
	if s.ownerRepo == nil {
		return nil, "", nil
	}
	owner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if findErr != nil {
		return nil, "", apperrors.Wrap(findErr, "failed to find owner for demote tag remove")
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil, "", nil
	}
	client, buildErr := s.buildClient(ctx, clinicID)
	if buildErr != nil {
		return nil, "", buildErr
	}
	return client, *owner.LineUserID, nil
}

// runBatch は ownerID リストに対して除外チェック・重複チェック・タグ付与・ログ記録を行う汎用バッチ実行。
