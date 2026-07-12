package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncFoodPurchaseTagWithMappings はフード購入履歴に基づき LTV_フード購入あり タグを同期する（FEAT-379）。
// 事前取得済み mappings/thresholds を使って処理する（PERF-M2 N+1 解消用）。
func (s *lstepTagSyncService) SyncFoodPurchaseTagWithMappings(ctx context.Context, clinicID, ownerID uint64, cachedMappings []*model.LstepTagCodeMapping, cachedThresholds *model.HealthPreventionThresholds) error {
	if s.tagCodeRepo == nil || s.billingItemRepo == nil {
		return nil
	}

	// PERF-M2: cachedMappings が提供されている場合は再取得しない（batch からの hoist）。
	var mappings []*model.LstepTagCodeMapping
	if cachedMappings != nil {
		mappings = cachedMappings
	} else {
		var err error
		mappings, err = s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, LtvFoodPurchaseTag)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find tag code mappings for food purchase tag", "error", err)
			return apperrors.Wrap(err, "failed to find tag code mappings")
		}
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

	// PERF-M2: cachedThresholds が提供されている場合は再取得しない（batch からの hoist）。
	var thresholds model.HealthPreventionThresholds
	if cachedThresholds != nil {
		thresholds = *cachedThresholds
	} else {
		var tErr error
		thresholds, tErr = s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
		if tErr != nil {
			slog.ErrorContext(ctx, "failed to get health prevention thresholds for food purchase tag", "error", tErr, "clinic_id", clinicID)
			return apperrors.Wrap(tErr, "failed to get health prevention thresholds")
		}
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
