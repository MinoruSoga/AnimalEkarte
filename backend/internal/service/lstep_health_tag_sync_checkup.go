package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncHealthcheckTags は健診履歴に基づき HLTH_健診あり / HLTH_健診未受診 を同期する（FEAT-379）。
func (s *lstepTagSyncService) SyncHealthcheckTags(ctx context.Context, clinicID, ownerID uint64) error {
	return s.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
}

func (s *lstepTagSyncService) SyncHealthcheckTagsWithMappings(ctx context.Context, clinicID, ownerID uint64, cachedMappings []*model.LstepTagCodeMapping, cachedThresholds *model.HealthPreventionThresholds) error {
	if s.tagCodeRepo == nil {
		return nil
	}

	// PERF-03: Use cached mappings if provided, otherwise fetch (fallback)
	var mappings []*model.LstepTagCodeMapping
	if cachedMappings != nil {
		mappings = cachedMappings
	} else {
		var err error
		mappings, err = s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, HlthHealthcheckDoneTag)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find tag code mappings for healthcheck", "error", err)
			return apperrors.Wrap(err, "failed to find tag code mappings")
		}
	}
	checkupCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	if len(checkupCodes) == 0 {
		return nil
	}

	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, HlthHealthcheckDoneTag)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// PERF-1: cachedThresholds が提供されている場合は再取得しない（batch からの hoist）。
	var thresholds model.HealthPreventionThresholds
	if cachedThresholds != nil {
		thresholds = *cachedThresholds
	} else {
		var tErr error
		thresholds, tErr = s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
		if tErr != nil {
			slog.ErrorContext(ctx, "failed to get health prevention thresholds for healthcheck tags", "error", tErr, "clinic_id", clinicID)
			return apperrors.Wrap(tErr, "failed to get health prevention thresholds")
		}
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)
	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for healthcheck tags", "error", err)
		return apperrors.Wrap(err, "failed to find checkups")
	}

	codeSet := strSet(checkupCodes)
	hasHealthcheck := false
	var lastCheckupDate time.Time
	for i := range checkups {
		if checkups[i].Date.Before(since) {
			continue
		}
		if checkups[i].CheckupType == nil {
			continue
		}
		if _, ok := codeSet[checkups[i].CheckupType.Name]; ok {
			hasHealthcheck = true
			if checkups[i].Date.After(lastCheckupDate) {
				lastCheckupDate = checkups[i].Date
			}
		}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if hasHealthcheck {
		doneReason := fmt.Sprintf("最終健診: %s", lastCheckupDate.Format(time.DateOnly))
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthHealthcheckDoneTag, "healthcheck done", doneReason, true); err != nil {
			return err
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthHealthcheckNeverTag, "healthcheck never", "", false); err != nil {
			apiFailed = true
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthHealthcheckNeverTag, "healthcheck never", "", true); err != nil {
			return err
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthHealthcheckDoneTag, "healthcheck done", "", false); err != nil {
			apiFailed = true
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncAnnual4CheckupTag は年2回以上来院かつ健診履歴がある飼い主に HLTH_年4回候補 を付与する（FEAT-379）。
func (s *lstepTagSyncService) SyncAnnual4CheckupTag(ctx context.Context, clinicID, ownerID uint64) error {
	return s.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil)
}

// SyncAnnual4CheckupTagWithMappings は事前取得済み mappings/thresholds を使って処理する（PERF-M1 N+1 解消用）。
func (s *lstepTagSyncService) SyncAnnual4CheckupTagWithMappings(ctx context.Context, clinicID, ownerID uint64, cachedMappings []*model.LstepTagCodeMapping, cachedThresholds *model.HealthPreventionThresholds) error {
	if s.tagCodeRepo == nil {
		return nil
	}

	// PERF-M1: cachedMappings が提供されている場合は再取得しない（batch からの hoist）。
	var mappings []*model.LstepTagCodeMapping
	if cachedMappings != nil {
		mappings = cachedMappings
	} else {
		var err error
		mappings, err = s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, HlthHealthcheckDoneTag)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find tag code mappings for annual4checkup", "error", err)
			return apperrors.Wrap(err, "failed to find tag code mappings")
		}
	}
	checkupCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	if len(checkupCodes) == 0 {
		return nil
	}

	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, HlthAnnual4CheckupTag)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// PERF-M1: cachedThresholds が提供されている場合は再取得しない（batch からの hoist）。
	var thresholds model.HealthPreventionThresholds
	if cachedThresholds != nil {
		thresholds = *cachedThresholds
	} else {
		var tErr error
		thresholds, tErr = s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
		if tErr != nil {
			slog.ErrorContext(ctx, "failed to get health prevention thresholds for annual4checkup tag", "error", tErr, "clinic_id", clinicID)
			return apperrors.Wrap(tErr, "failed to get health prevention thresholds")
		}
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)
	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for annual4checkup tag", "error", err)
		return apperrors.Wrap(err, "failed to find checkups")
	}

	codeSet := strSet(checkupCodes)
	hasHealthcheck := false
	for i := range checkups {
		if checkups[i].Date.Before(since) {
			continue
		}
		if checkups[i].CheckupType == nil {
			continue
		}
		if _, ok := codeSet[checkups[i].CheckupType.Name]; ok {
			hasHealthcheck = true
			break
		}
	}

	visitSummary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for annual4checkup tag", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	qualified := hasHealthcheck && visitSummary.AnnualCount >= 2

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if qualified {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthAnnual4CheckupTag, "annual4checkup", "", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthAnnual4CheckupTag, "annual4checkup", "", false); err != nil {
			apiFailed = true
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncVaccineDeadlineTag はワクチン次回予定日が VaccineDeadlineDays 以内なら
// PREV_ワクチン期限 を付与し、それ以外は解除する（FEAT-379）。
