package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepTagSyncService) SyncFoodPurchaseTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil || s.billingItemRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, LtvFoodPurchaseTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for food purchase tag", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	// itemCodes が空でも HasFoodPurchaseByOwnerSince は category='food' にフォールバック
	itemCodes := extractTagCodes(mappings, model.CodeTypeMerchandiseItem)

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for food purchase tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	thresholds, err := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get health prevention thresholds for food purchase tag", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get health prevention thresholds")
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)
	hasPurchase, err := s.billingItemRepo.HasFoodPurchaseByOwnerSince(ctx, clinicID, ownerID, since, itemCodes)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check food purchase for tag", "error", err)
		return apperrors.Wrap(err, "failed to check food purchase")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if hasPurchase {
		if addErr := client.AddTag(ctx, lineUserID, LtvFoodPurchaseTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add food purchase tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add food purchase tag")
		}
		if err := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, LtvFoodPurchaseTag, "auto", "購入済"); err != nil {
			slog.WarnContext(ctx, "failed to upsert tag cache (non-fatal)", "tag", LtvFoodPurchaseTag, "owner_id", ownerID, "error", err)
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, LtvFoodPurchaseTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove food purchase tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			if err := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, LtvFoodPurchaseTag); err != nil {
				slog.WarnContext(ctx, "failed to delete tag cache (non-fatal)", "tag", LtvFoodPurchaseTag, "owner_id", ownerID, "error", err)
			}
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncHealthPreventionTagsForClinic は指定クリニックの全飼い主に対して
// 健診・予防・物販タグを一括同期する（FEAT-379 バッチエントリポイント）。
