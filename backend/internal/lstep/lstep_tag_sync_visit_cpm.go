package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// SyncCPMStageTag は CPM ステージタグを同期する（BE-011）。
// cpm_version が "v2" のクリニックは syncCPMStageTagV2 に委譲する（Q19）。
func (s *lstepTagSyncService) SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error {
	version, vErr := s.settingsSvc.GetCPMVersion(ctx, clinicID)
	if vErr != nil {
		slog.ErrorContext(ctx, "failed to get cpm_version for CPM dispatch", "error", vErr)
		return apperrors.Wrap(vErr, "failed to get cpm_version")
	}
	if version == "v2" {
		return s.syncCPMStageTagV2(ctx, clinicID, ownerID)
	}

	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "CPM stage")
	if err != nil {
		return err
	}
	if !ok {
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

	// 旧ステージタグをすべて削除してから新ステージを付与
	apiFailed := false
	for _, old := range allCPMStages {
		if old == stage {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, string(old), "CPM stage", "", false); err != nil {
			apiFailed = true
		}
	}

	if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, string(stage), "CPM stage", "", true); err != nil {
		return err
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// syncCPMStageTagV2 は来院回数ベース V2 CPM ステージタグを同期する（FEAT-377）。
func (s *lstepTagSyncService) syncCPMStageTagV2(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "CPM V2 stage")
	if err != nil {
		return err
	}
	if !ok {
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

	apiFailed := false
	for _, old := range allCPMV2Stages {
		if old == stage {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, string(old), "CPM V2 stage", "", false); err != nil {
			apiFailed = true
		}
	}

	if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, string(stage), "CPM V2 stage", "", true); err != nil {
		return err
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
