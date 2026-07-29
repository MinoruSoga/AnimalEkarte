package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func (s *lstepBatchService) RunLTVTopPercentSyncAllClinics(ctx context.Context) error {
	return s.runBatchAllClinics(ctx,
		"ltv-top-percent batch", "ltv top percent batch", "synced ltv tags", "batch_ltv_top_percent",
		nil,
		s.tagSyncSvc.SyncLTVTopPercent,
	)
}

// syncVisitDormantForClinic は 1 クリニック分の VISIT_* タグ同期を行う（G3-2 切り出し）。
// entries 取得失敗は (0, nil) で握り潰さず error として返し、audit / Failed に計上する（LSA-03 / DEC-35）。
// G2F-04: FindDormantOwnerEntries（unbounded）ではなく Cursor でページングする。
func (s *lstepBatchService) syncVisitDormantForClinic(ctx context.Context, clinicID uint64) (int, []error) {
	const minDaysSince = 120
	count := 0
	var errs []error
	afterOwnerID := uint64(0)
	for {
		entries, findErr := s.medRecordRepo.FindDormantOwnerEntriesCursor(ctx, clinicID, minDaysSince, afterOwnerID, lstepBatchPageSize)
		if findErr != nil {
			slog.ErrorContext(ctx, "visit-dormant batch: failed to find entries", "clinic_id", clinicID, "error", findErr)
			return count, append(errs, apperrors.Wrap(findErr, "failed to find visit dormant entries"))
		}
		if len(entries) == 0 {
			break
		}
		for _, entry := range entries {
			if tagErr := s.tagSyncSvc.SyncVisitDormantTags(ctx, clinicID, entry.OwnerID, entry.DaysSince); tagErr != nil {
				slog.ErrorContext(ctx, "visit-dormant batch: failed to sync tag", "clinic_id", clinicID, "owner_id", entry.OwnerID, "error", tagErr)
				errs = append(errs, apperrors.Wrap(tagErr, "failed to sync visit dormant tag"))
				continue
			}
			count++
		}
		afterOwnerID = entries[len(entries)-1].OwnerID
		if len(entries) < lstepBatchPageSize {
			break
		}
	}
	return count, errs
}

// RunVisitDormantSyncAllClinics は全クリニックに対して VISIT_* タグを同期する（FEAT-377）。
func (s *lstepBatchService) RunVisitDormantSyncAllClinics(ctx context.Context) error {
	return s.runBatchAllClinics(ctx,
		"visit-dormant batch", "visit dormant batch", "synced visit tags", "batch_visit_dormant",
		nil,
		s.syncVisitDormantForClinic,
	)
}

// RunHealthPreventionTagSyncAllClinics は全クリニックに対して健診・予防・物販タグを同期する（FEAT-379）。
func (s *lstepBatchService) RunHealthPreventionTagSyncAllClinics(ctx context.Context) error {
	return s.runBatchAllClinics(ctx,
		"health-prevention batch", "health prevention batch", "synced tags", "batch_health_prevention",
		nil,
		s.tagSyncSvc.SyncHealthPreventionTagsForClinic,
	)
}

// runDeliveryTriggersForClinic は 1 クリニック分の全配信トリガーバッチを実行する（FEAT-383）。
