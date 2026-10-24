package lstep

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"log/slog"

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

type nextVisitCleanupIntentFinder interface {
	FindNextVisitCleanupIntent(context.Context, uint64, uint64) (*NextVisitCleanupIntent, error)
}

type nextVisitCleanupStore interface {
	nextVisitSyncRetryQueue
	nextVisitOwnerLocker
	nextVisitCleanupIntentFinder
	FindOwnerForNextVisitCleanup(context.Context, uint64, uint64) (*model.Owner, error)
	CountActiveOwnersWithLineUser(context.Context, uint64, string) (int64, error)
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
	skip, err := s.shouldSkipSync(ctx, clinicID)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}
	finder, ok := s.tagCacheRepo.(nextVisitCleanupIntentFinder)
	if !ok {
		return apperrors.WrapInternalServerError("next visit cleanup intent lookup is not configured")
	}
	intent, err := finder.FindNextVisitCleanupIntent(ctx, clinicID, ownerID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return err
		}
		intent = nil
	}
	if intent != nil {
		return s.syncNextVisitCleanup(ctx, clinicID, ownerID)
	}
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
		s.notifyNextVisitAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
	} else if err == nil && attempted {
		s.notifyNextVisitAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return err
}

func (s *lstepTagSyncService) revalidatedNextVisitTarget(syncCtx context.Context, clinicID, ownerID uint64, lineUserID string) (string, error) {
	target, ready, err := s.resolveSyncTarget(syncCtx, clinicID, ownerID, "next visit")
	if err != nil {
		if apperrors.IsNotFound(err) {
			slog.WarnContext(syncCtx, "next visit notifier skipped: owner missing or deleted", "clinic_id", clinicID, "owner_id", ownerID)
			return "", nil
		}
		return "", err
	}
	if !ready || target != lineUserID {
		return "", nil
	}
	return target, nil
}

func (s *lstepTagSyncService) nextVisitNotifierLock(ctx context.Context, clinicID, ownerID uint64) (nextVisitOwnerLocker, bool) {
	locker, ok := s.tagCacheRepo.(nextVisitOwnerLocker)
	if !ok {
		slog.ErrorContext(ctx, "next visit owner lock is not configured for notifier", "clinic_id", clinicID, "owner_id", ownerID)
		return nil, false
	}
	return locker, true
}

func (s *lstepTagSyncService) notifyNextVisitAPIFailure(ctx context.Context, client lstepapi.Client, clinicID, ownerID uint64, lineUserID string) {
	if s.errorCounterRepo == nil {
		return
	}
	newCount, err := s.errorCounterRepo.IncrementFailure(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to increment lstep error counter", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if newCount < lstepSyncErrorThreshold {
		return
	}
	locker, ok := s.nextVisitNotifierLock(ctx, clinicID, ownerID)
	if !ok {
		return
	}
	err = locker.WithNextVisitOwnerLock(ctx, clinicID, ownerID, func(syncCtx context.Context) error {
		target, err := s.revalidatedNextVisitTarget(syncCtx, clinicID, ownerID, lineUserID)
		if err != nil || target == "" {
			return err
		}
		if addErr := client.AddTag(syncCtx, target, lstepErrorTag); addErr != nil {
			return apperrors.Wrap(addErr, "failed to add EXCL error tag")
		}
		return s.tagCacheRepo.UpsertTag(syncCtx, clinicID, ownerID, lstepErrorTag, "auto", "")
	})
	if err != nil {
		slog.ErrorContext(ctx, "next visit failure notifier transaction failed", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
	}
}

func (s *lstepTagSyncService) notifyNextVisitAPISuccess(ctx context.Context, client lstepapi.Client, clinicID, ownerID uint64, lineUserID string) {
	if s.errorCounterRepo == nil {
		return
	}
	counter, err := s.errorCounterRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return
		}
		slog.ErrorContext(ctx, "failed to find lstep error counter", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if counter.FailureCount == 0 {
		return
	}
	locker, ok := s.nextVisitNotifierLock(ctx, clinicID, ownerID)
	if !ok {
		return
	}
	attempted := false
	err = locker.WithNextVisitOwnerLock(ctx, clinicID, ownerID, func(syncCtx context.Context) error {
		target, err := s.revalidatedNextVisitTarget(syncCtx, clinicID, ownerID, lineUserID)
		if err != nil || target == "" {
			return err
		}
		attempted = true
		if removeErr := client.RemoveTag(syncCtx, target, lstepErrorTag); removeErr != nil {
			return apperrors.Wrap(removeErr, "failed to remove EXCL error tag")
		}
		return s.tagCacheRepo.DeleteTag(syncCtx, clinicID, ownerID, lstepErrorTag)
	})
	if err != nil {
		slog.ErrorContext(ctx, "next visit success notifier transaction failed", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if !attempted {
		return
	}
	if resetErr := s.errorCounterRepo.ResetFailure(ctx, clinicID, ownerID); resetErr != nil {
		slog.ErrorContext(ctx, "failed to reset lstep error counter", "error", resetErr, "clinic_id", clinicID, "owner_id", ownerID)
	}
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

func (s *lstepTagSyncService) EnqueueNextVisitCleanup(ctx context.Context, clinicID, ownerID uint64, lineUserID string) error {
	queue, ok := s.tagCacheRepo.(nextVisitCleanupQueue)
	if !ok {
		return apperrors.WrapInternalServerError("next visit cleanup queue is not configured")
	}
	return queue.EnqueueNextVisitCleanup(ctx, clinicID, ownerID, lineUserID)
}

func (s *lstepTagSyncService) syncNextVisitCleanup(ctx context.Context, clinicID, ownerID uint64) error {
	cleanup, ok := s.tagCacheRepo.(nextVisitCleanupStore)
	if !ok {
		return apperrors.WrapInternalServerError("next visit cleanup store is not configured")
	}
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}
	return cleanup.WithNextVisitOwnerLock(ctx, clinicID, ownerID, func(syncCtx context.Context) error {
		// Settings may have changed while waiting for the shared recipient lock.
		// A stopped clinic retains its intent and revision for a later retry.
		if skip, err := s.shouldSkipSync(syncCtx, clinicID); err != nil {
			return err
		} else if skip {
			return nil
		}
		intent, err := cleanup.FindNextVisitCleanupIntent(syncCtx, clinicID, ownerID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if intent == nil {
			return apperrors.WrapInternalServerError("next visit cleanup intent lookup returned no intent")
		}
		owner, err := cleanup.FindOwnerForNextVisitCleanup(syncCtx, clinicID, ownerID)
		if err != nil {
			return err
		}
		if owner == nil {
			return apperrors.WrapInternalServerError("next visit cleanup owner lookup returned no owner")
		}
		if !owner.DeletedAt.Valid {
			return apperrors.WrapConflict("next visit cleanup intent requires a deleted owner")
		}
		currentRecipient, recipientErr := linkedLineUserID(syncCtx, s.recipients, clinicID, ownerID)
		if recipientErr != nil {
			return recipientErr
		}
		if currentRecipient == "" || currentRecipient != intent.LineUserID {
			return apperrors.WrapConflict("next visit cleanup target does not match the owner recipient")
		}
		if owner.LstepOptOut {
			return nil
		}
		relinked, err := cleanup.CountActiveOwnersWithLineUser(syncCtx, clinicID, intent.LineUserID)
		if err != nil {
			return err
		}
		if relinked > 0 {
			return apperrors.WrapConflict("next visit cleanup recipient was relinked to an active owner")
		}
		remote, err := client.GetUserTags(syncCtx, intent.LineUserID)
		if err != nil {
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
		cached, err := s.tagCacheRepo.FindByOwner(syncCtx, clinicID, ownerID)
		if err != nil {
			return apperrors.Wrap(err, "failed to load tag cache")
		}
		for _, tag := range cached {
			if strings.HasPrefix(tag.TagName, tagPrefixNextVisit) {
				staleTags[tag.TagName] = struct{}{}
			}
		}
		tags := make([]string, 0, len(staleTags))
		for tag := range staleTags {
			tags = append(tags, tag)
		}
		sort.Strings(tags)
		var removalErrs []error
		for _, tag := range tags {
			if _, present := remoteTags[tag]; present {
				if err := client.RemoveTag(syncCtx, intent.LineUserID, tag); err != nil {
					removalErrs = append(removalErrs, apperrors.Wrap(err, "failed to remove next visit tag"))
					continue
				}
			}
			if err := s.tagCacheRepo.DeleteTag(syncCtx, clinicID, ownerID, tag); err != nil {
				removalErrs = append(removalErrs, apperrors.Wrap(err, "failed to delete next visit tag cache"))
			}
		}
		if err := errors.Join(removalErrs...); err != nil {
			return err
		}
		return cleanup.AcknowledgeNextVisitSync(syncCtx, clinicID, ownerID, intent.Revision)
	})
}
