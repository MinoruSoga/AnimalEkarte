package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncHealthcheckTagsWithMappings は健診履歴に基づき HLTH_健診あり / HLTH_健診未受診 を同期する（FEAT-379）。
func (s *lstepTagSyncService) SyncHealthcheckTagsWithMappings(ctx context.Context, clinicID, ownerID uint64, cachedMappings []*model.LstepTagCodeMapping, cachedThresholds *model.HealthPreventionThresholds) error {
	return s.syncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, cachedMappings, cachedThresholds, nil)
}

// syncHealthcheckTagsWithMappings は preloadedCheckups が非 nil なら再取得せずその値を使う（G2F-02 page bulk）。
func (s *lstepTagSyncService) syncHealthcheckTagsWithMappings(
	ctx context.Context,
	clinicID, ownerID uint64,
	cachedMappings []*model.LstepTagCodeMapping,
	cachedThresholds *model.HealthPreventionThresholds,
	preloadedCheckups *[]model.Checkup,
) error {
	if s.tagCodeRepo == nil {
		return nil
	}

	// PERF-03: Use cached mappings if provided, otherwise fetch (fallback)（BE-refactor.md E-7）。
	mappings, err := s.mappingsFor(ctx, clinicID, HlthHealthcheckDoneTag, "healthcheck", cachedMappings)
	if err != nil {
		return err
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

	// PERF-1: cachedThresholds が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	thresholds, err := s.thresholdsFor(ctx, clinicID, "healthcheck tags", cachedThresholds)
	if err != nil {
		return err
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)

	checkups, err := s.checkupsForOwner(ctx, clinicID, ownerID, preloadedCheckups, "healthcheck tags")
	if err != nil {
		return err
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

	// DEC-35 / G2B-01: Remove 失敗も err 伝播（silent success 禁止）。
	if hasHealthcheck {
		doneReason := fmt.Sprintf("最終健診: %s", lastCheckupDate.Format(time.DateOnly))
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthHealthcheckDoneTag, "healthcheck done", doneReason, true); err != nil {
			return err
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthHealthcheckNeverTag, "healthcheck never", "", false); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthHealthcheckNeverTag, "healthcheck never", "", true); err != nil {
			return err
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthHealthcheckDoneTag, "healthcheck done", "", false); err != nil {
			return err
		}
	}
	s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	return nil
}

// SyncAnnual4CheckupTagWithMappings は年2回以上来院かつ健診履歴がある飼い主に HLTH_年4回候補 を付与する
// （FEAT-379）。事前取得済み mappings/thresholds を使って処理する（PERF-M1 N+1 解消用）。
func (s *lstepTagSyncService) SyncAnnual4CheckupTagWithMappings(ctx context.Context, clinicID, ownerID uint64, cachedMappings []*model.LstepTagCodeMapping, cachedThresholds *model.HealthPreventionThresholds) error {
	return s.syncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, cachedMappings, cachedThresholds, nil, nil)
}

// syncAnnual4CheckupTagWithMappings は preloaded 入力が非 nil なら再取得しない（G2F-02 page bulk）。
func (s *lstepTagSyncService) syncAnnual4CheckupTagWithMappings(
	ctx context.Context,
	clinicID, ownerID uint64,
	cachedMappings []*model.LstepTagCodeMapping,
	cachedThresholds *model.HealthPreventionThresholds,
	preloadedCheckups *[]model.Checkup,
	preloadedVisitSummary **medicalrecord.OwnerVisitSummary,
) error {
	if s.tagCodeRepo == nil {
		return nil
	}

	// PERF-M1: cachedMappings が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	mappings, err := s.mappingsFor(ctx, clinicID, HlthHealthcheckDoneTag, "annual4checkup", cachedMappings)
	if err != nil {
		return err
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

	// PERF-M1: cachedThresholds が提供されている場合は再取得しない（batch からの hoist）（BE-refactor.md E-7）。
	thresholds, err := s.thresholdsFor(ctx, clinicID, "annual4checkup tag", cachedThresholds)
	if err != nil {
		return err
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)

	checkups, err := s.checkupsForOwner(ctx, clinicID, ownerID, preloadedCheckups, "annual4checkup tag")
	if err != nil {
		return err
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

	visitSummary, err := s.visitSummaryForOwner(ctx, clinicID, ownerID, preloadedVisitSummary)
	if err != nil {
		return err
	}

	qualified := hasHealthcheck && visitSummary.AnnualCount >= 2

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	// DEC-35 / G2B-01: Remove 失敗も err 伝播（silent success 禁止）。
	if qualified {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthAnnual4CheckupTag, "annual4checkup", "", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, HlthAnnual4CheckupTag, "annual4checkup", "", false); err != nil {
			return err
		}
	}
	s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	return nil
}

// checkupsForOwner は preloaded が非 nil ならそれを返し、nil なら repo から取得する。
func (s *lstepTagSyncService) checkupsForOwner(
	ctx context.Context,
	clinicID, ownerID uint64,
	preloaded *[]model.Checkup,
	label string,
) ([]model.Checkup, error) {
	if preloaded != nil {
		return *preloaded, nil
	}
	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for "+label, "error", err)
		return nil, apperrors.Wrap(err, "failed to find checkups")
	}
	return checkups, nil
}

// visitSummaryForOwner は preloaded が非 nil ならそれを返し、nil なら repo から取得する。
// preloaded の内側が nil の場合は空サマリ（来院 0）として扱う。
func (s *lstepTagSyncService) visitSummaryForOwner(
	ctx context.Context,
	clinicID, ownerID uint64,
	preloaded **medicalrecord.OwnerVisitSummary,
) (*medicalrecord.OwnerVisitSummary, error) {
	if preloaded != nil {
		if *preloaded != nil {
			return *preloaded, nil
		}
		return &medicalrecord.OwnerVisitSummary{}, nil
	}
	visitSummary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for annual4checkup tag", "error", err)
		return nil, apperrors.Wrap(err, "failed to find visit summary")
	}
	return visitSummary, nil
}

// SyncVaccineDeadlineTag はワクチン次回予定日が VaccineDeadlineDays 以内なら
// PREV_ワクチン期限 を付与し、それ以外は解除する（FEAT-379）。
