package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepDeliveryTriggerService) checkExclusion(ctx context.Context, clinicID, ownerID uint64) (excluded bool, reason string, err error) {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return false, "", apperrors.Wrap(err, "failed to find owner")
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

// recordTrigger はトリガーログを作成し生成された ID を返す。
func (s *lstepDeliveryTriggerService) recordTrigger(ctx context.Context, clinicID, ownerID uint64, triggerType string, asOf time.Time) (uint64, error) {
	log := &model.LstepDeliveryTriggerLog{
		OwnerID:     ownerID,
		ClinicID:    clinicID,
		TriggerType: triggerType,
		ScheduledAt: asOf,
		Status:      model.TriggerStatusScheduled,
	}
	if err := s.triggerLogRepo.Create(ctx, log); err != nil {
		return 0, apperrors.Wrap(err, "failed to create trigger log")
	}
	return log.ID, nil
}

// applyTagAndLog は L ステップへタグを付与しログを fired 状態に更新する。
