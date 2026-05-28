package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// SyncCPMStageTag は CPM ステージタグを同期する（BE-011）。
// cpm_version が "v2" のクリニックは SyncCPMStageTagV2 に委譲する（Q19）。
func (s *lstepTagSyncService) SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error {
	version, vErr := s.settingsSvc.GetCPMVersion(ctx, clinicID)
	if vErr != nil {
		slog.ErrorContext(ctx, "failed to get cpm_version for CPM dispatch", "error", vErr)
		return apperrors.Wrap(vErr, "failed to get cpm_version")
	}
	if version == "v2" {
		return s.SyncCPMStageTagV2(ctx, clinicID, ownerID)
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for CPM tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	summary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for CPM", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	ltv, err := s.accountRepo.SumPaidByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to sum paid amount for CPM", "error", err)
		return apperrors.Wrap(err, "failed to sum paid amount")
	}

	maxSingle, err := s.accountRepo.MaxSingleVisitAmountByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get max single visit amount for CPM", "error", err)
		return apperrors.Wrap(err, "failed to get max single visit amount")
	}

	thresholds, err := s.settingsSvc.GetCPMV1Thresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cpm v1 thresholds", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get cpm v1 thresholds")
	}

	daysSince := -1
	if summary.LastVisitAt != nil {
		daysSince = int(time.Since(*summary.LastVisitAt).Hours() / 24)
	}
	firstVisitDaysSince := 0
	if summary.FirstVisitAt != nil {
		firstVisitDaysSince = int(time.Since(*summary.FirstVisitAt).Hours() / 24)
	}

	stage := CalculateCPMStage(CPMData{
		TotalVisitCount:      summary.TotalCount,
		AnnualVisitCount:     summary.AnnualCount,
		DaysSinceVisit:       daysSince,
		LTVAmount:            ltv,
		FirstVisitDaysSince:  firstVisitDaysSince,
		MaxSingleVisitAmount: maxSingle,
		Thresholds:           thresholds,
	})

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID

	// 旧ステージタグをすべて削除してから新ステージを付与
	apiFailed := false
	for _, old := range allCPMStages {
		if string(old) == string(stage) {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, string(old)); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old CPM stage tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, string(old))
		}
	}

	if addErr := client.AddTag(ctx, lineUserID, string(stage)); addErr != nil {
		slog.ErrorContext(ctx, "failed to add CPM stage tag", "error", addErr, "stage", stage)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add CPM stage tag")
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, string(stage), "auto", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert CPM stage tag cache", "error", cacheErr)
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncCPMStageTagV2 は来院回数ベース V2 CPM ステージタグを同期する（FEAT-377）。
func (s *lstepTagSyncService) SyncCPMStageTagV2(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for CPM V2 tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	summary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for CPM V2", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	v2Thresholds, err := s.settingsSvc.GetCPMV2Thresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cpm v2 thresholds", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get cpm v2 thresholds")
	}

	stage := CalculateCPMStageV2(CPMStageV2Input{
		TotalVisitCount: summary.TotalCount,
		CPMV2Thresholds: v2Thresholds,
	})

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID
	apiFailed := false
	for _, old := range allCPMV2Stages {
		if old == stage {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, string(old)); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old CPM V2 stage tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, string(old))
		}
	}

	if addErr := client.AddTag(ctx, lineUserID, string(stage)); addErr != nil {
		slog.ErrorContext(ctx, "failed to add CPM V2 stage tag", "error", addErr, "stage", stage)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add CPM V2 stage tag")
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, string(stage), "auto", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert CPM V2 stage tag cache", "error", cacheErr)
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
