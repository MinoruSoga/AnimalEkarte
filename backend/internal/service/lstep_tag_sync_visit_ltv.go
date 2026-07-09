package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// SyncLTVTopPercent は LTV 上位 20% の飼い主に LTV_上位20 タグを付与し、
// それ以外から解除する（FEAT-377）。
func (s *lstepTagSyncService) SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{err}
	} else if skip {
		return 0, nil
	}

	revenues, err := s.accountRepo.FindOwnersByAnnualRevenue(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find revenues", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by annual revenue")}
	}

	topN := 0
	if len(revenues) > 0 {
		topN = (len(revenues)*20 + 99) / 100
	}
	topOwnerIDs := make(map[uint64]struct{}, topN)
	for i := 0; i < topN; i++ {
		topOwnerIDs[revenues[i].OwnerID] = struct{}{}
	}

	owners, err := s.ownerRepo.FindAllWithLineUserID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners with line user id")}
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
	for i := range owners {
		owner := &owners[i]
		if owner.LineUserID == nil || *owner.LineUserID == "" {
			continue
		}
		lineUserID := *owner.LineUserID
		_, isTop := topOwnerIDs[owner.ID]
		if isTop {
			if err := s.applyTagState(ctx, client, clinicID, owner.ID, lineUserID, ltvTop20Tag, "LTV top20", "", true); err != nil {
				errs = append(errs, err)
				continue
			}
		} else {
			cached, cacheErr := s.tagCacheRepo.FindByOwner(ctx, clinicID, owner.ID)
			if cacheErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to load cache", "owner_id", owner.ID, "error", cacheErr)
				errs = append(errs, apperrors.Wrap(cacheErr, fmt.Sprintf("failed to load tag cache for owner %d", owner.ID)))
				continue
			}
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
			if err := s.applyTagState(ctx, client, clinicID, owner.ID, lineUserID, ltvTop20Tag, "LTV top20", "", false); err != nil {
				errs = append(errs, err)
				continue
			}
		}
		s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
		count++
	}
	return count, errs
}
