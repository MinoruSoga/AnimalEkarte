package owner

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *ownerService) UpdateDeliveryExclusion(ctx context.Context, clinicID, id uint64, input UpdateDeliveryExclusionInput) (*model.Owner, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}

	var reason *string
	if input.Excluded && input.Reason != nil {
		normalized, err := normalizeOwnerReason(*input.Reason, colDeliveryExcludedReason)
		if err != nil {
			return nil, err
		}
		reason = normalized
	}

	var optOutAt any
	if input.Excluded {
		optOutAt = time.Now()
	}
	fields := map[string]any{
		colDeliveryExcluded:       input.Excluded,
		colDeliveryExcludedReason: reason,
		colLstepOptOut:            input.Excluded,
		colLstepOptOutAt:          optOutAt,
		colLstepOptOutReason:      reason,
	}
	updated, err := s.updateOwnerAndFind(
		ctx,
		clinicID,
		id,
		fields,
		"failed to update delivery exclusion",
		"failed to update delivery exclusion",
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "delivery exclusion updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("excluded", input.Excluded))
	if s.tagNotifier != nil {
		if err := s.tagNotifier.SyncExclusionTags(ctx, clinicID, id); err != nil {
			slog.ErrorContext(ctx, "failed to sync exclusion tag after delivery exclusion update", "error", err, "id", id, "clinic_id", clinicID)
		}
	}

	return updated, nil
}

func (s *ownerService) UpdateDeliveryCaution(ctx context.Context, clinicID, id uint64, input UpdateDeliveryCautionInput) (*model.Owner, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}

	var reason *string
	if input.Caution {
		normalized, err := normalizeOwnerReason(input.Reason, colDeliveryCautionReason)
		if err != nil {
			return nil, err
		}
		reason = normalized
	}

	fields := map[string]any{
		colDeliveryCaution:       input.Caution,
		colDeliveryCautionReason: reason,
	}
	updated, err := s.updateOwnerAndFind(
		ctx,
		clinicID,
		id,
		fields,
		"failed to update delivery caution",
		"failed to update delivery caution",
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "delivery caution updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("caution", input.Caution))
	if s.tagNotifier != nil {
		if err := s.tagNotifier.SyncExclusionTags(ctx, clinicID, id); err != nil {
			slog.WarnContext(ctx, "failed to sync exclusion tag after delivery caution update", "error", err, "id", id, "clinic_id", clinicID)
		}
	}

	return updated, nil
}

func (s *ownerService) UpdateTransferStatus(ctx context.Context, clinicID, id uint64, input UpdateTransferStatusInput) (*model.Owner, error) {
	owner, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	var transferAt any
	if input.IsTransferred {
		now := time.Now()
		transferAt = now
	}
	fields := map[string]any{
		colIsTransferred: input.IsTransferred,
		colTransferAt:    transferAt,
	}
	if input.IsTransferred {
		fields[colMembershipType] = model.MembershipTypeTransferred
	} else if owner.MembershipType == model.MembershipTypeTransferred {
		fields[colMembershipType] = model.MembershipTypeNonMember
	}
	updated, err := s.updateOwnerAndFind(
		ctx,
		clinicID,
		id,
		fields,
		"failed to update transfer status",
		"failed to update transfer status",
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "transfer status updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("is_transferred", input.IsTransferred))
	if s.tagNotifier != nil {
		if err := s.tagNotifier.SyncExclusionTags(ctx, clinicID, id); err != nil {
			slog.ErrorContext(ctx, "failed to sync exclusion tag after transfer status update", "error", err, "id", id, "clinic_id", clinicID)
		}
	}

	return updated, nil
}
