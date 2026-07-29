package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncFoodPurchaseTagWithMappings はフード購入履歴に基づき LTV_フード購入あり タグを同期する（FEAT-379）。
// 事前取得済み mappings/thresholds を使って処理する（PERF-M2 N+1 解消用）。
func (s *lstepTagSyncService) SyncFoodPurchaseTagWithMappings(ctx context.Context, clinicID, ownerID uint64, cachedMappings []*model.LstepTagCodeMapping, cachedThresholds *model.HealthPreventionThresholds) error {
	if s.tagCodeRepo == nil || s.billingItemRepo == nil {
		return nil
	}

	// PERF-M2: cachedMappings が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	mappings, err := s.mappingsFor(ctx, clinicID, LtvFoodPurchaseTag, "food purchase tag", cachedMappings)
	if err != nil {
		return err
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

	// PERF-M2: cachedThresholds が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	thresholds, err := s.thresholdsFor(ctx, clinicID, "food purchase tag", cachedThresholds)
	if err != nil {
		return err
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

	// DEC-35 / G2B-01: Remove 失敗も Add 同様に err 伝播し、BatchRunResult.Failed に計上する。
	if hasPurchase {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, LtvFoodPurchaseTag, "food purchase", "購入済", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, LtvFoodPurchaseTag, "food purchase", "購入済", false); err != nil {
			return err
		}
	}
	s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	return nil
}

// SyncHealthPreventionTagsForClinic は指定クリニックの全飼い主に対して
// 健診・予防・物販タグを一括同期する（FEAT-379 バッチエントリポイント）。
