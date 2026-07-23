package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

func (s *lstepBatchService) detectDormantOwners(ctx context.Context, clinicID uint64) (int, []error) {
	return s.detectDormantOwnersWithFinder(
		ctx,
		clinicID,
		s.medRecordRepo.FindDormantOwnerEntriesCursor,
	)
}

func (s *lstepBatchService) detectDormantOwnersAt(
	ctx context.Context,
	clinicID uint64,
	evaluatedAt time.Time,
) (int, []error) {
	repositoryAt, ok := s.medRecordRepo.(lstepBatchMedicalRecordRepositoryAt)
	if !ok {
		err := apperrors.WrapInternalServerError("dormant scheduled-time repository is not configured")
		slog.ErrorContext(ctx, "dormant batch: scheduled-time repository is not configured", "clinic_id", clinicID)
		return 0, []error{err}
	}
	return s.detectDormantOwnersWithFinder(
		ctx,
		clinicID,
		func(
			ctx context.Context,
			clinicID uint64,
			minDaysSince int,
			afterOwnerID uint64,
			limit int,
		) ([]medicalrecord.DormantOwnerEntry, error) {
			return repositoryAt.FindDormantOwnerEntriesCursorAt(
				ctx,
				clinicID,
				minDaysSince,
				afterOwnerID,
				limit,
				evaluatedAt,
			)
		},
	)
}

type dormantOwnerPageFinder func(
	ctx context.Context,
	clinicID uint64,
	minDaysSince int,
	afterOwnerID uint64,
	limit int,
) ([]medicalrecord.DormantOwnerEntry, error)

func (s *lstepBatchService) detectDormantOwnersWithFinder(
	ctx context.Context,
	clinicID uint64,
	findPage dormantOwnerPageFinder,
) (int, []error) {
	const minDaysSince = 180
	if s.settingsSvc == nil {
		err := apperrors.WrapInternalServerError("LSTEP settings service is not configured")
		slog.ErrorContext(ctx, "dormant batch: settings service is not configured", "clinic_id", clinicID)
		return 0, []error{err}
	}

	// PERF-FOLLOWUP-02: 無制限全件取得を避けるため、先頭ページのみ先に取得する。
	firstPage, err := findPage(ctx, clinicID, minDaysSince, 0, lstepDormantBatchPageSize)
	if err != nil {
		slog.ErrorContext(ctx, "dormant batch: failed to find dormant owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find dormant owners")}
	}

	// PERF-2: 閾値を clinic 単位で 1 回だけ取得し、ループ内で再利用する（N+1 解消）。
	thresholds, tErr := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if tErr != nil {
		slog.ErrorContext(ctx, "dormant batch: failed to get dormant thresholds", "clinic_id", clinicID, "error", tErr)
		return 0, []error{apperrors.Wrap(tErr, "failed to get dormant thresholds")}
	}

	var errs []error
	count := 0
	syncEntries := func(entries []medicalrecord.DormantOwnerEntry) {
		for _, entry := range entries {
			if tagErr := s.tagSyncSvc.SyncDormantTagsWithThresholds(ctx, clinicID, entry.OwnerID, entry.DaysSince, thresholds); tagErr != nil {
				slog.ErrorContext(ctx, "dormant batch: failed to sync dormant tag", "clinic_id", clinicID, "owner_id", entry.OwnerID, "error", tagErr)
				errs = append(errs, apperrors.Wrap(tagErr, "failed to sync dormant tag"))
				continue
			}
			count++
		}
	}

	// PERF-FOLLOWUP-02: カーソルページネーションで休眠飼い主を取得しながら処理する。
	// 直前のページが pageSize ちょうどの場合のみ次ページを取得する（最後の空ページ取得を 1 回に抑える）。
	page := firstPage
	afterOwnerID := uint64(0)
	for len(page) > 0 {
		syncEntries(page)
		afterOwnerID = page[len(page)-1].OwnerID
		if len(page) < lstepDormantBatchPageSize {
			break
		}
		var pageErr error
		page, pageErr = findPage(ctx, clinicID, minDaysSince, afterOwnerID, lstepDormantBatchPageSize)
		if pageErr != nil {
			slog.ErrorContext(ctx, "dormant batch: failed to find dormant owners (next page)",
				"clinic_id", clinicID, "after_owner_id", afterOwnerID, "error", pageErr)
			errs = append(errs, apperrors.Wrap(pageErr, "failed to find dormant owners"))
			break
		}
	}
	return count, errs
}

// RunLTVTopPercentSyncAllClinics は全クリニックに対して LTV 上位 20% タグを同期する（FEAT-377）。

// RunDormantDetectionAllClinics は全クリニックに対して休眠飼い主検知を実行する（02:00 JST cron）。
// runBatchAllClinics に処理を委譲する。ISSUE-010: 処理件数・エラー件数・閾値(min_days_since)を
// audit メタデータとして永続化し、閾値は後から判定基準を再現できるよう含める。
func (s *lstepBatchService) RunDormantDetectionAllClinics(ctx context.Context) error {
	return s.runBatchAllClinics(ctx,
		"dormant batch", "dormant batch", "synced dormant tags", "batch_dormant_detect",
		map[string]any{"min_days_since": 180},
		s.detectDormantOwners,
	)
}

// RunDormantDetectionAllClinicsAt uses the immutable scheduled instant for
// both the dormant cutoff and the reported days-since value.
func (s *lstepBatchService) RunDormantDetectionAllClinicsAt(
	ctx context.Context,
	scheduledAt time.Time,
	runID string,
) BatchRunResult {
	scheduledAt = scheduledAt.UTC()
	metadata := scheduledBatchMetadata(scheduledAt, runID)
	metadata["min_days_since"] = 180
	result, _ := s.runBatchAllClinicsWithResult(
		ctx,
		"dormant batch",
		"dormant batch",
		"synced dormant tags",
		"batch_dormant_detect",
		metadata,
		func(ctx context.Context, clinicID uint64) (int, []error) {
			return s.detectDormantOwnersAt(ctx, clinicID, scheduledAt)
		},
	)
	return result
}
