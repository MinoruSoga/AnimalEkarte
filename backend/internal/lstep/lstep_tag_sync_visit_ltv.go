package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncLTVTopPercent は LTV 上位 20% の飼い主に LTV_上位20 タグを付与し、
// それ以外から解除する（FEAT-377 / G2F-03）。
//
// Ranking contract:
//   - accountRepo.FindOwnersByAnnualRevenue は DB 側で exact top-20% を確定した
//     bounded 集合を返す（全 clinic owner 売上を Go へ materialize しない）。
//   - 返却集合の全 owner が top 対象。Go 側で再スライスしない。
//
// Tag application:
//   - LINE 連携 owner は FindAllWithLineUserIDCursor でページングし、
//     ページ単位で tag cache をバッチ取得する（全件 materialize を避ける）。
func (s *lstepTagSyncService) SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{err}
	} else if skip {
		return 0, nil
	}

	// Bounded top-percent set from SQL (exact ceil(N*20/100), deterministic ties).
	// Do not recompute topN from len(revenues) — revenues is already the top set.
	topRevenues, err := s.accountRepo.FindOwnersByAnnualRevenue(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find revenues", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by annual revenue")}
	}

	topOwnerIDs := make(map[uint64]struct{}, len(topRevenues))
	for _, rev := range topRevenues {
		topOwnerIDs[rev.OwnerID] = struct{}{}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return 0, []error{err}
	}
	if client == nil {
		return 0, nil
	}

	var errs []error
	count := 0
	afterID := uint64(0)
	for {
		owners, pageErr := s.ownerRepo.FindAllWithLineUserIDCursor(ctx, clinicID, afterID, lstepBatchPageSize)
		if pageErr != nil {
			slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find owners page",
				"clinic_id", clinicID, "after_id", afterID, "error", pageErr)
			return count, append(errs, apperrors.Wrap(pageErr, "failed to find owners with line user id"))
		}
		if len(owners) == 0 {
			break
		}

		// Page-scoped non-top owner IDs for batch tag-cache load (PERF-FOLLOWUP-08 / G2F-03).
		nonTopOwnerIDs := make([]uint64, 0, len(owners))
		for i := range owners {
			owner := &owners[i]
			if owner.LineUserID == nil || *owner.LineUserID == "" {
				continue
			}
			if _, isTop := topOwnerIDs[owner.ID]; isTop {
				continue
			}
			nonTopOwnerIDs = append(nonTopOwnerIDs, owner.ID)
		}
		var tagCacheByOwner map[uint64][]*model.LstepTagCache
		if len(nonTopOwnerIDs) > 0 {
			tagCacheByOwner, err = s.tagCacheRepo.FindByOwners(ctx, clinicID, nonTopOwnerIDs)
			if err != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to batch load tag cache", "clinic_id", clinicID, "error", err)
				return count, append(errs, apperrors.Wrap(err, "failed to batch load tag cache"))
			}
		}

		for i := range owners {
			owner := &owners[i]
			if owner.LineUserID == nil || *owner.LineUserID == "" {
				continue
			}
			lineUserID := *owner.LineUserID
			_, isTop := topOwnerIDs[owner.ID]
			if isTop {
				if tagErr := s.applyTagState(ctx, client, clinicID, owner.ID, lineUserID, ltvTop20Tag, "LTV top20", "", true); tagErr != nil {
					errs = append(errs, tagErr)
					continue
				}
			} else {
				cached := tagCacheByOwner[owner.ID]
				hasTag := false
				for _, c := range cached {
					if c.TagName == ltvTop20Tag {
						hasTag = true
						break
					}
				}
				if !hasTag {
					s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
					count++
					continue
				}
				if tagErr := s.applyTagState(ctx, client, clinicID, owner.ID, lineUserID, ltvTop20Tag, "LTV top20", "", false); tagErr != nil {
					errs = append(errs, tagErr)
					continue
				}
			}
			s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
			count++
		}

		afterID = owners[len(owners)-1].ID
		if len(owners) < lstepBatchPageSize {
			break
		}
	}
	return count, errs
}
