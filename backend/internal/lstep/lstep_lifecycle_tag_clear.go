package lstep

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// removeAllTagsFromLstep preserves the cache-based all-tags cleanup policy.
// Share next-visit serialization so a pre-death sync cannot add a stale tag
// after the all-pets-dead cleanup has completed.
func (s *lstepLifecycleService) removeAllTagsFromLstep(ctx context.Context, clinicID, ownerID uint64, lineUserID string) error {
	ctx, cancel := context.WithTimeout(persistence.WithTxValue(ctx, nil), 30*time.Second)
	defer cancel()
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	return s.clearTagsWithOwnerLock(ctx, client, clinicID, ownerID, lineUserID)
}

func (s *lstepLifecycleService) clearTagsWithOwnerLock(ctx context.Context, client lstepapi.Client, clinicID, ownerID uint64, lineUserID string) error {
	locker, ok := s.tagCacheRepo.(nextVisitOwnerLocker)
	if !ok {
		return apperrors.WrapInternalServerError("next visit owner lock is not configured")
	}
	return locker.WithNextVisitOwnerLock(ctx, clinicID, ownerID, func(ctx context.Context) error {
		// The lock fixes the current owner and its current LINE recipient. Never
		// operate on an earlier snapshot: that recipient may now belong to someone else.
		if s.ownerRepo == nil {
			return apperrors.WrapInternalServerError("cleanup owner repository is not configured")
		}
		owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
		if err != nil {
			return apperrors.Wrap(err, "failed to revalidate cleanup owner")
		}
		if owner == nil {
			return apperrors.WrapInternalServerError("cleanup owner lookup returned no owner")
		}
		if owner.LineUserID == nil || *owner.LineUserID == "" || *owner.LineUserID != lineUserID {
			return nil
		}
		tags, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
		if err != nil {
			return apperrors.Wrap(err, "failed to find tag cache")
		}

		if client == nil {
			// 同期無効・APIキー未設定: リモート解除不能なのに cache を消さない（LSB-02）
			slog.WarnContext(ctx, "lstep client unavailable; keeping tag cache for retry", "owner_id", ownerID, "tag_count", len(tags))
			return nil
		}

		// Cache writes can fail after a provider Add succeeds. Include only our
		// derived next_visit tags from the provider; other remote categories keep
		// the existing cache-based cleanup policy.
		var failures []error
		remote, err := client.GetUserTags(ctx, lineUserID)
		if err != nil {
			// A discovery failure must not block best-effort removal of known
			// cache tags. Retain the error so the transaction keeps retry state.
			failures = append(failures, apperrors.Wrap(err, "failed to read next visit tags before cleanup"))
		}
		tagSet := make(map[string]struct{}, len(tags))
		for _, tag := range tags {
			tagSet[tag.TagName] = struct{}{}
		}
		for _, tag := range remote {
			if strings.HasPrefix(tag, tagPrefixNextVisit) {
				tagSet[tag] = struct{}{}
			}
		}
		names := make([]string, 0, len(tagSet))
		for tag := range tagSet {
			names = append(names, tag)
		}
		sort.Strings(names)
		for _, tag := range names {
			if err := client.RemoveTag(ctx, lineUserID, tag); err != nil {
				failures = append(failures, apperrors.Wrap(err, "failed to remove tag on cleanup"))
				continue
			}
			if err := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, tag); err != nil {
				failures = append(failures, apperrors.Wrap(err, "failed to delete tag cache after remote remove"))
			}
		}
		// The caller keeps the committed lifecycle operation successful, logs this
		// secondary error, and retries from provider/cache state on the next event.
		return errors.Join(failures...)
	})
}
