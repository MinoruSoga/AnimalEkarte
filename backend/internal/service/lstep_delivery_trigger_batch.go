package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepDeliveryTriggerService) runBatch(
	ctx context.Context,
	clinicID uint64,
	ownerIDs []uint64,
	triggerType string,
	tagName string,
	asOf time.Time,
) (int, []error) {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: failed to build lstep client", "clinic_id", clinicID, "trigger", triggerType, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to build lstep client")}
	}
	if client == nil {
		return 0, nil
	}

	var errs []error
	count := 0

	for _, ownerID := range ownerIDs {
		fired, loopErr := s.processSingleOwner(ctx, client, clinicID, ownerID, triggerType, tagName, asOf)
		if loopErr != nil {
			errs = append(errs, loopErr)
			continue
		}
		if fired {
			count++
		}
	}
	return count, errs
}

// processSingleOwner は1飼い主分のトリガー処理を行い、配信実行されたか否かを返す。
func (s *lstepDeliveryTriggerService) processSingleOwner(
	ctx context.Context,
	client lstep.Client,
	clinicID, ownerID uint64,
	triggerType, tagName string,
	asOf time.Time,
) (bool, error) {
	already, err := s.alreadyFiredToday(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: alreadyFiredToday error", "owner_id", ownerID, "trigger", triggerType, "error", err)
		return false, err
	}
	if already {
		return false, nil
	}

	// Q23: 優先順位による配信抑制チェック
	suppressed, err := s.applySuppression(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: applySuppression error", "owner_id", ownerID, "trigger", triggerType, "error", err)
		return false, err
	}
	if suppressed {
		suppressionReason := fmt.Sprintf("owner_id=%d already has higher-priority trigger on %s", ownerID, asOf.Format("2006-01-02"))
		suppressedLog := &model.LstepDeliveryTriggerLog{
			OwnerID:              ownerID,
			ClinicID:             clinicID,
			TriggerType:          triggerType,
			ScheduledAt:          asOf,
			Status:               model.TriggerStatusScheduled,
			SuppressedByPriority: true,
			SuppressionReason:    &suppressionReason,
		}
		if createErr := s.triggerLogRepo.Create(ctx, suppressedLog); createErr != nil {
			slog.ErrorContext(ctx, "delivery trigger: failed to create suppressed log", "owner_id", ownerID, "trigger", triggerType, "error", createErr)
			return false, apperrors.Wrap(createErr, "failed to create suppressed trigger log")
		}
		return false, nil
	}

	excluded, reason, err := s.checkExclusion(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: checkExclusion error", "owner_id", ownerID, "error", err)
		return false, err
	}

	logID, err := s.recordTrigger(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: recordTrigger error", "owner_id", ownerID, "error", err)
		return false, err
	}

	if excluded {
		_ = s.triggerLogRepo.UpdateStatus(ctx, clinicID, logID, string(model.TriggerStatusExcluded), nil, &reason)
		return false, nil
	}

	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: failed to find owner for lineUserID", "owner_id", ownerID, "error", err)
		return false, apperrors.Wrap(err, "failed to find owner")
	}

	if err := s.applyTagAndLog(ctx, clinicID, client, *owner.LineUserID, tagName, logID); err != nil {
		return false, err
	}
	return true, nil
}

// ---- public trigger methods ----
