package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncFilariaTagWithMappings は事前取得済み mappings/thresholds を使って処理する（PERF-M2 N+1 解消用）。
func (s *lstepTagSyncService) SyncFilariaTagWithMappings(ctx context.Context, clinicID, ownerID uint64, cachedMappings []*model.LstepTagCodeMapping, cachedThresholds *model.HealthPreventionThresholds) error {
	if s.tagCodeRepo == nil {
		return nil
	}

	// PERF-M2: cachedMappings が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	mappings, err := s.mappingsFor(ctx, clinicID, PrevFilariaTag, "filaria tag", cachedMappings)
	if err != nil {
		return err
	}
	testCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	rxCodes := extractTagCodes(mappings, model.CodeTypePrescription)
	if len(testCodes) == 0 && len(rxCodes) == 0 {
		return nil
	}

	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, PrevFilariaTag)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// 犬を保有する飼い主のみ対象
	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pets for filaria tag", "error", err)
		return apperrors.Wrap(err, "failed to find pets")
	}
	hasDog := false
	for i := range pets {
		if pets[i].AnimalSpecies != nil && isDogSpeciesName(pets[i].AnimalSpecies.Name) {
			hasDog = true
			break
		}
	}
	if !hasDog {
		return nil
	}

	// PERF-M2: cachedThresholds が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	thresholds, err := s.thresholdsFor(ctx, clinicID, "filaria tag", cachedThresholds)
	if err != nil {
		return err
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)

	testDone := false
	if len(testCodes) > 0 {
		checkups, chkErr := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
		if chkErr != nil {
			slog.ErrorContext(ctx, "failed to find checkups for filaria tag", "error", chkErr)
			return apperrors.Wrap(chkErr, "failed to find checkups")
		}
		codeSet := strSet(testCodes)
		for i := range checkups {
			if checkups[i].Date.Before(since) {
				continue
			}
			if checkups[i].CheckupType == nil {
				continue
			}
			if _, ok := codeSet[checkups[i].CheckupType.Name]; ok {
				testDone = true
				break
			}
		}
	}

	rxDone := false
	if len(rxCodes) > 0 && s.billingItemRepo != nil {
		var rxErr error
		rxDone, rxErr = s.billingItemRepo.HasItemByOwnerSince(ctx, clinicID, ownerID, since, rxCodes)
		if rxErr != nil {
			slog.ErrorContext(ctx, "failed to check prescription for filaria tag", "error", rxErr)
			return apperrors.Wrap(rxErr, "failed to check prescription")
		}
	}

	// 検査・処方どちらか完了していれば「完了」→ タグ解除
	incomplete := !testDone && !rxDone

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	// DEC-35 / G2B-01: Remove 失敗も Add 同様に err 伝播し、BatchRunResult.Failed に計上する。
	if incomplete {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevFilariaTag, "filaria", "未処方", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevFilariaTag, "filaria", "未処方", false); err != nil {
			return err
		}
	}
	s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	return nil
}

// SyncFleaTickTagWithMappings はノミ・マダニ駆除薬処方に基づき PREV_ノミダニ対象 を同期する（FEAT-379）。
// 処方実績がなければタグを付与し、あれば解除する。事前取得済み mappings/thresholds を使って処理する（PERF-M2 N+1 解消用）。
func (s *lstepTagSyncService) SyncFleaTickTagWithMappings(ctx context.Context, clinicID, ownerID uint64, cachedMappings []*model.LstepTagCodeMapping, cachedThresholds *model.HealthPreventionThresholds) error {
	if s.tagCodeRepo == nil || s.billingItemRepo == nil {
		return nil
	}

	// PERF-M2: cachedMappings が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	mappings, err := s.mappingsFor(ctx, clinicID, PrevFleaTickTag, "flea tick tag", cachedMappings)
	if err != nil {
		return err
	}
	rxCodes := extractTagCodes(mappings, model.CodeTypePrescription)
	if len(rxCodes) == 0 {
		return nil
	}

	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, PrevFleaTickTag)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// PERF-M2: cachedThresholds が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	thresholds, err := s.thresholdsFor(ctx, clinicID, "flea tick tag", cachedThresholds)
	if err != nil {
		return err
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)
	hasRx, err := s.billingItemRepo.HasItemByOwnerSince(ctx, clinicID, ownerID, since, rxCodes)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check billing items for flea tick tag", "error", err)
		return apperrors.Wrap(err, "failed to check billing items")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	// DEC-35 / G2B-01: Remove 失敗も Add 同様に err 伝播し、BatchRunResult.Failed に計上する。
	if !hasRx {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevFleaTickTag, "flea tick", "未処方", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevFleaTickTag, "flea tick", "未処方", false); err != nil {
			return err
		}
	}
	s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	return nil
}

// SyncFoodPurchaseTag はフード購入履歴に基づき LTV_フード購入あり を同期する（FEAT-379）。
// codes が空の場合は category='food' でフォールバック検索する。
