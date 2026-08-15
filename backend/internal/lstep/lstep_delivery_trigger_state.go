package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepDeliveryTriggerService) checkExclusion(ctx context.Context, clinicID, ownerID uint64, owner *model.Owner) (excluded bool, reason string, err error) {
	// Fail-closed: nil owner is never delivery-eligible. Can reach here when
	// FindByID returns (nil, nil) (mock or contract violation) after bulk cache
	// miss/fallback (see processSingleOwner). Do not treat as "not excluded".
	if owner == nil {
		return true, "owner_missing", nil
	}
	// LSB-01 / LSA-02 / DEC-38: LstepOptOut is fail-closed delivery exclusion evidence
	// independent of delivery_excluded and EXCL cache tags (which opt-out may clear).
	if owner.LstepOptOut {
		return true, "lstep_opt_out", nil
	}
	if owner.DeliveryExcluded {
		return true, "delivery_excluded_flag", nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return true, "no_line_user_id", nil
	}
	tags, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		return false, "", apperrors.Wrap(err, "failed to find owner tags")
	}
	for _, t := range tags {
		if t.TagName == exclTagDeliveryStop {
			return true, "excl_tag_delivery_stop", nil
		}
	}
	return false, "", nil
}

// alreadyFiredToday は当日同一トリガーが既に発火済みかを確認する（二重発火防止）。
func (s *lstepDeliveryTriggerService) alreadyFiredToday(ctx context.Context, clinicID, ownerID uint64, triggerType string, asOf time.Time) (bool, error) {
	exists, err := s.triggerLogRepo.ExistsTodayByOwnerAndType(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to check trigger log")
	}
	return exists, nil
}

// recordTrigger claims the day/type slot under advisory lock (LSA-15) and returns the log ID.
// created=false means another worker already claimed the slot — caller must treat as no-op.
func (s *lstepDeliveryTriggerService) recordTrigger(ctx context.Context, clinicID, ownerID uint64, triggerType string, asOf time.Time) (id uint64, created bool, err error) {
	log := &model.LstepDeliveryTriggerLog{
		OwnerID:     ownerID,
		ClinicID:    clinicID,
		TriggerType: triggerType,
		ScheduledAt: asOf,
		Status:      model.TriggerStatusScheduled,
	}
	created, err = s.triggerLogRepo.CreateIfAbsentToday(ctx, log)
	if err != nil {
		return 0, false, apperrors.Wrap(err, "failed to create trigger log")
	}
	if !created {
		return 0, false, nil
	}
	return log.ID, true, nil
}

// applyTagAndLog は L ステップへタグを付与しログを fired 状態に更新する。
