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

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, LtvFoodPurchaseTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for food purchase tag", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	// itemCodes が空でも HasFoodPurchaseByOwnerSince は category='food' にフォールバック
	itemCodes := extractTagCodes(mappings, model.CodeTypeMerchandiseItem)

	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, LtvFoodPurchaseTag)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

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
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, LtvFoodPurchaseTag, "food purchase", "購入済", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, LtvFoodPurchaseTag, "food purchase", "購入済", false); err != nil {
			apiFailed = true
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncHealthPreventionTagsForClinic は指定クリニックの全飼い主に対して
// 健診・予防・物販タグを一括同期する（FEAT-379 バッチエントリポイント）。
