package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type expiredTentativeTagOwnerReader interface {
	FindOwnerIDsWithExpiredTentativeTagCursor(context.Context, uint64, time.Time, uint64, int) ([]uint64, error)
}

// reconcileExpiredTentativeTags uses immutable scheduled asOf only to select
// expired candidates. SyncNextVisitTag deliberately reads the current wall-clock
// appointment state: a delayed scheduler must not restore a historic next_visit.
// No cursor/checkpoint is persisted; a failed owner remains eligible for retry.
func (s *lstepBatchService) reconcileExpiredTentativeTags(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	repo, ok := s.reservationRepo.(expiredTentativeTagOwnerReader)
	if !ok || repo == nil || s.tagSyncSvc == nil {
		return 0, []error{apperrors.WrapInternalServerError("expired tentative tag sync dependencies are required")}
	}
	const pageSize = 200
	var after uint64
	count := 0
	var errs []error
	for {
		if err := ctx.Err(); err != nil {
			return count, append(errs, err)
		}
		owners, err := repo.FindOwnerIDsWithExpiredTentativeTagCursor(ctx, clinicID, asOf, after, pageSize)
		if err != nil {
			return count, append(errs, apperrors.Wrap(err, "expired tentative tag candidate lookup failed"))
		}
		if len(owners) == 0 {
			return count, errs
		}
		previous := after
		for _, ownerID := range owners {
			// DISTINCT/keyset is the repository contract. Guard duplicate/zero
			// results without an unbounded in-memory seen-owner set.
			if ownerID <= after {
				continue
			}
			after = ownerID
			if err := s.tagSyncSvc.SyncNextVisitTag(ctx, clinicID, ownerID); err != nil {
				slog.ErrorContext(ctx, "expired tentative tag sync failed", "clinic_id", clinicID, "owner_id", ownerID, "error", err)
				errs = append(errs, apperrors.Wrap(err, "expired tentative tag sync failed"))
				continue
			}
			count++
		}
		if after == previous {
			return count, append(errs, apperrors.WrapInternalServerError("expired tentative owner cursor did not advance"))
		}
	}
}
