package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// dormantBatchPageSize は PERF-FOLLOWUP-02 のカーソルページネーション 1 ページあたりの件数。
const dormantBatchPageSize = 500

func (s *lstepBatchService) DetectDormantOwners(ctx context.Context, clinicID uint64) (int, []error) {
	const minDaysSince = 180

	// PERF-FOLLOWUP-02: 無制限全件取得を避けるため、先頭ページのみ先に取得する。
	firstPage, err := s.medRecordRepo.FindDormantOwnerEntriesCursor(ctx, clinicID, minDaysSince, 0, dormantBatchPageSize)
	if err != nil {
		slog.ErrorContext(ctx, "dormant batch: failed to find dormant owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find dormant owners")}
	}

	// PERF-2: 閾値を clinic 単位で 1 回だけ取得し、ループ内で再利用する（N+1 解消）。
	// settingsSvc が nil の場合（旧コード互換 / テスト環境）はデフォルト閾値で継続する。
	var thresholds model.DormantThresholds
	if s.settingsSvc != nil {
		var tErr error
		thresholds, tErr = s.settingsSvc.GetDormantThresholds(ctx, clinicID)
		if tErr != nil {
			slog.ErrorContext(ctx, "dormant batch: failed to get dormant thresholds", "clinic_id", clinicID, "error", tErr)
			return 0, []error{apperrors.Wrap(tErr, "failed to get dormant thresholds")}
		}
	} else {
		thresholds = model.DormantThresholds{}.WithDefaults()
	}

	var errs []error
	count := 0
	syncEntries := func(entries []repository.DormantOwnerEntry) {
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
	for {
		if len(page) == 0 {
			break
		}
		syncEntries(page)
		afterOwnerID = page[len(page)-1].OwnerID
		if len(page) < dormantBatchPageSize {
			break
		}
		var pageErr error
		page, pageErr = s.medRecordRepo.FindDormantOwnerEntriesCursor(ctx, clinicID, minDaysSince, afterOwnerID, dormantBatchPageSize)
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
		s.DetectDormantOwners,
	)
}
