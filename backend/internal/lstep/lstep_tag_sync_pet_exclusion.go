package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncExclusionTags は配信停止条件に基づき EXCL_配信停止 タグを同期する（FEAT-377/FEAT-381）。
// 配信停止条件: LstepOptOut=true / DeliveryExcluded=true / IsTransferred=true /
//
//	会員ステータスが deceased / transferred / 全ペット死亡。
//
// 注: checkOptOut は呼ばない（このメソッド自体が opt-out 判定の実装）。
func (s *lstepTagSyncService) SyncExclusionTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	// 全ペット死亡判定
	totalPets, err := s.petRepo.CountByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count pets for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to count pets")
	}
	livingPets, err := s.petRepo.CountLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count living pets for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to count living pets")
	}
	allPetsDead := totalPets > 0 && livingPets == 0

	shouldExclude := owner.LstepOptOut ||
		owner.DeliveryExcluded ||
		owner.IsTransferred ||
		owner.MembershipType == model.MembershipTypeDeceased ||
		owner.MembershipType == model.MembershipTypeTransferred ||
		allPetsDead

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if shouldExclude {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, exclTagDeliveryStop, "EXCL exclusion", "", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, exclTagDeliveryStop, "EXCL exclusion", "", false); err != nil {
			apiFailed = true
		}
	}

	// FEAT-381-2: EXCL_配信注意 タグは delivery_caution フラグのみで独立して制御する。
	if owner.DeliveryCaution {
		if addErr := client.AddTag(ctx, lineUserID, exclTagDeliveryCaution); addErr != nil {
			slog.ErrorContext(ctx, "failed to add EXCL caution tag", "error", addErr)
			apiFailed = true
		} else if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, exclTagDeliveryCaution, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert EXCL caution tag cache", "error", cacheErr)
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, exclTagDeliveryCaution); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove EXCL caution tag", "error", delErr)
			apiFailed = true
		} else if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, exclTagDeliveryCaution); delCacheErr != nil {
			slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", exclTagDeliveryCaution)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
