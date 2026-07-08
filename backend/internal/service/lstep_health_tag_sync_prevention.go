package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepTagSyncService) SyncFilariaTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, PrevFilariaTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for filaria tag", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	testCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	rxCodes := extractTagCodes(mappings, model.CodeTypePrescription)
	if len(testCodes) == 0 && len(rxCodes) == 0 {
		return nil
	}

	lineUserID, ok, err := s.resolveSyncTargetOwner(ctx, clinicID, ownerID, PrevFilariaTag)
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
		if pets[i].AnimalSpecies != nil && strings.Contains(pets[i].AnimalSpecies.Name, "犬") {
			hasDog = true
			break
		}
	}
	if !hasDog {
		return nil
	}

	thresholds, err := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get health prevention thresholds for filaria tag", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get health prevention thresholds")
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

	apiFailed := false
	if incomplete {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevFilariaTag, "filaria", "未処方", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevFilariaTag, "filaria", "未処方", false); err != nil {
			apiFailed = true
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncFleaTickTag はノミ・マダニ駆除薬処方に基づき PREV_ノミダニ対象 を同期する（FEAT-379）。
// 処方実績がなければタグを付与し、あれば解除する。
func (s *lstepTagSyncService) SyncFleaTickTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil || s.billingItemRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, PrevFleaTickTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for flea tick tag", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	rxCodes := extractTagCodes(mappings, model.CodeTypePrescription)
	if len(rxCodes) == 0 {
		return nil
	}

	lineUserID, ok, err := s.resolveSyncTargetOwner(ctx, clinicID, ownerID, PrevFleaTickTag)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	thresholds, err := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get health prevention thresholds for flea tick tag", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get health prevention thresholds")
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

	apiFailed := false
	if !hasRx {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevFleaTickTag, "flea tick", "未処方", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevFleaTickTag, "flea tick", "未処方", false); err != nil {
			apiFailed = true
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncFoodPurchaseTag はフード購入履歴に基づき LTV_フード購入あり を同期する（FEAT-379）。
// codes が空の場合は category='food' でフォールバック検索する。
