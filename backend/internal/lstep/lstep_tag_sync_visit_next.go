package lstep

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// nextVisitOwnerLocker owns a short transaction shared by source/cache reads
// and cache writes. Its advisory lock serializes HTTP and batch syncs across processes.
type nextVisitOwnerLocker interface {
	WithNextVisitOwnerLock(context.Context, uint64, uint64, func(context.Context) error) error
}

// SyncNextVisitTag は次回来院タグを同期する（BE-006 / EMR-308）。
// 飼い主の次回 active appointment を参照し、古い next_visit_* タグを差し替える。
// tentative は intended_date、それ以外は start_time の JST 暦日で
// next_visit_YYYY-MM-DD を生成する。次回来院がない場合はすべての next_visit_* タグを削除する。
// appointment finder 未注入は構成不全 —— タグ除去へフォールバックすると
// 誤配線の呼び出し元（cmd/lstep-migrate 等）が既存タグを黙って剥奪するため、
// タグに触れずエラーを返す（fail-closed）。
func (s *lstepTagSyncService) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	// Remove only the caller's transaction: cancellation and any shorter batch
	// deadline survive. Post-commit callers own request cancellation detachment.
	ctx, cancel := context.WithTimeout(persistence.WithTxValue(ctx, nil), 30*time.Second)
	defer cancel()
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "next visit")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}
	if s.nextVisitAppointmentRepo == nil {
		return apperrors.WrapInternalServerError("next visit appointment repository is not configured")
	}
	locker, ok := s.tagCacheRepo.(nextVisitOwnerLocker)
	if !ok {
		return apperrors.WrapInternalServerError("next visit owner lock is not configured")
	}
	retries, ok := s.tagCacheRepo.(nextVisitSyncRetryQueue)
	if !ok {
		return apperrors.WrapInternalServerError("next visit retry queue is not configured")
	}
	// Durable before provider side effects; a failed/ambiguous cache commit must
	// remain discoverable even when this owner has never had a cached tag.
	revision, err := retries.EnqueueNextVisitSync(ctx, clinicID, ownerID)
	if err != nil {
		return err
	}
	apiFailed := false
	attempted := false
	err = locker.WithNextVisitOwnerLock(ctx, clinicID, ownerID, func(syncCtx context.Context) error {
		// The fast gate may precede a queued opt-out/link/settings commit.
		// Re-read under the sync-owned transaction; both repositories use
		// DBOrTx, so this does not borrow a second connection from the pool.
		target, ready, targetErr := s.resolveSyncTarget(syncCtx, clinicID, ownerID, "next visit")
		if targetErr != nil {
			return targetErr
		}
		if !ready {
			return nil
		}
		lineUserID = target
		attempted = true
		if syncErr := s.reconcileNextVisitTag(syncCtx, client, clinicID, ownerID, lineUserID, &apiFailed); syncErr != nil {
			return syncErr
		}
		return retries.AcknowledgeNextVisitSync(syncCtx, clinicID, ownerID, revision)
	})
	// Counter persistence must happen after the lock transaction releases its
	// connection: otherwise rollback loses accounting, and a pool of one deadlocks.
	if apiFailed {
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
	} else if err == nil && attempted {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return err
}

func (s *lstepTagSyncService) reconcileNextVisitTag(ctx context.Context, client lstepapi.Client, clinicID, ownerID uint64, lineUserID string, apiFailed *bool) error {
	next, err := s.nextVisitAppointmentRepo.FindNextAppointmentForOwner(ctx, clinicID, ownerID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find next appointment")
	}
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	remote, err := client.GetUserTags(ctx, lineUserID)
	if err != nil {
		*apiFailed = true
		return apperrors.Wrap(err, "failed to read next visit tags")
	}
	remoteTags := make(map[string]struct{}, len(remote))
	staleTags := make(map[string]struct{})
	for _, tag := range remote {
		remoteTags[tag] = struct{}{}
		if strings.HasPrefix(tag, tagPrefixNextVisit) {
			staleTags[tag] = struct{}{}
		}
	}
	for _, tag := range cached {
		if strings.HasPrefix(tag.TagName, tagPrefixNextVisit) {
			staleTags[tag.TagName] = struct{}{}
		}
	}
	newTag := ""
	if date := nextVisitDateFromAppointment(next); date != nil {
		newTag = tagPrefixNextVisit + date.In(config.JST).Format(time.DateOnly)
		// Cache presence is never proof of provider presence. A remote success
		// followed by a cache/commit failure converges on the next provider read.
		if _, present := remoteTags[newTag]; !present {
			if err := client.AddTag(ctx, lineUserID, newTag); err != nil {
				*apiFailed = true
				return apperrors.Wrap(err, "failed to add next visit tag")
			}
		}
		if err := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto", ""); err != nil {
			return apperrors.Wrap(err, "failed to persist next visit tag")
		}
	}
	// Add the current tag before retiring old tags; retain failed cache entries
	// for the bounded daily retry. Other tag categories are untouched.
	tags := make([]string, 0, len(staleTags))
	for tag := range staleTags {
		if tag != newTag {
			tags = append(tags, tag)
		}
	}
	sort.Strings(tags)
	var removalErrs []error
	for _, tag := range tags {
		if _, present := remoteTags[tag]; present {
			if err := client.RemoveTag(ctx, lineUserID, tag); err != nil {
				*apiFailed = true
				removalErrs = append(removalErrs, apperrors.Wrap(err, "failed to remove next visit tag"))
				continue
			}
		}
		if err := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, tag); err != nil {
			removalErrs = append(removalErrs, apperrors.Wrap(err, "failed to delete next visit tag cache"))
		}
	}
	return errors.Join(removalErrs...)
}

// nextVisitDateFromAppointment は appointment からタグ用の意味上の予定日を返す。
// tentative（予定・未確定）は intended_date、それ以外の active 予約は start_time。
// 該当なし・tentative で intended_date 未設定の場合は nil（＝ next_visit_* 除去）。
func nextVisitDateFromAppointment(appointment *model.Reservation) *time.Time {
	if appointment == nil {
		return nil
	}
	if appointment.Status == model.ReservationStatusTentative {
		return appointment.IntendedDate
	}
	return &appointment.StartTime
}
