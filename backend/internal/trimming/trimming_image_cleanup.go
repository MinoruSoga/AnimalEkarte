package trimming

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// deleteReplacedTrimmingImages は Update commit 後の置き換え済み画像を掃除する。
// replaced は update tx 内で捕捉した直前値（Service.Update が返す）— tx 前の
// スナップショットより正確で、同時 PATCH が互いの新値を掃除候補から漏らさない。
func (h *Handler) deleteReplacedTrimmingImages(
	ctx context.Context,
	fallbackClinicID, actorID uint64,
	replaced *TrimmingReplacedImages,
	updated *model.Reservation,
) {
	if replaced == nil {
		return
	}
	newStyle, newCompleted, ok := trimmingStoredImages(updated)
	if !ok {
		return
	}
	clinicID := trimmingRecordClinicID(updated, fallbackClinicID)
	h.deleteSupersededTrimmingImages(ctx, clinicID, 0, actorID, replaced.Style, replaced.Completed, newStyle, newCompleted)
}

func trimmingImageObjectKey(stored string, clinicID uint64) (string, bool) {
	key, isObject, err := resolveTrimmingImageStorageKey(stored)
	if err != nil || !isObject || !trimmingImageKeyBelongsToClinic(key, clinicID) {
		return "", false
	}
	return key, true
}

func supersededTrimmingImageKeys(clinicID uint64, oldStyle, oldCompleted, newStyle, newCompleted string) []string {
	keep := map[string]struct{}{}
	for _, value := range []string{newStyle, newCompleted} {
		key, ok := trimmingImageObjectKey(value, clinicID)
		if ok {
			keep[key] = struct{}{}
		}
	}
	var dropped []string
	seen := map[string]struct{}{}
	for _, value := range []string{oldStyle, oldCompleted} {
		key, ok := trimmingImageObjectKey(value, clinicID)
		if !ok {
			continue
		}
		if _, retained := keep[key]; retained {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		dropped = append(dropped, key)
	}
	return dropped
}

func (h *Handler) deleteSupersededTrimmingImages(
	ctx context.Context,
	clinicID, exceptAppointmentID, actorID uint64,
	oldStyle, oldCompleted, newStyle, newCompleted string,
) {
	h.bestEffortDeleteTrimmingImageKeys(ctx, clinicID, exceptAppointmentID, actorID, supersededTrimmingImageKeys(
		clinicID, oldStyle, oldCompleted, newStyle, newCompleted,
	)...)
}

// supersededImageCleanupTimeout bounds post-commit object cleanup so a stalled
// S3/R2 delete cannot hold the mutation response open (client timeout/retry
// while the record is already committed). 削除失敗は warn のみで孤立扱い。
const supersededImageCleanupTimeout = 10 * time.Second

func (h *Handler) bestEffortDeleteTrimmingImageKeys(ctx context.Context, clinicID, exceptAppointmentID, actorID uint64, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if h == nil || h.imageUploader == nil {
		slog.WarnContext(ctx, "skipped superseded trimming image delete; uploader is not configured")
		return
	}
	deleteCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), supersededImageCleanupTimeout)
	defer cancel()
	for _, key := range keys {
		if key == "" {
			continue
		}
		// Reference check + durable tombstone commit precede the external delete.
		// PATCH includes the current record, so a later X→Y→X is protected.
		// DELETE excludes only its already-deleted appointment; shared keys stay.
		if !h.reserveTrimmingImageDeletion(deleteCtx, clinicID, key, exceptAppointmentID, actorID) {
			continue
		}
		if err := h.imageUploader.Delete(deleteCtx, key); err != nil {
			slog.WarnContext(deleteCtx, "failed to delete superseded trimming image",
				slog.String("error", err.Error()),
				slog.String("key", key))
		}
	}
}

// reserveTrimmingImageDeletion commits a key tombstone under the same locks as
// reference writes. A failed reservation retains the object (fail-closed).
func (h *Handler) reserveTrimmingImageDeletion(ctx context.Context, clinicID uint64, key string, exceptAppointmentID, actorID uint64) bool {
	if h.svc == nil || h.svc.Trimming == nil {
		slog.WarnContext(ctx, "skipped trimming image delete; lifecycle is not configured")
		return false
	}
	reserved, err := h.svc.Trimming.ReserveImageDeletion(ctx, clinicID, key, exceptAppointmentID, actorID)
	if err != nil {
		slog.WarnContext(ctx, "skipped trimming image delete; reservation failed", slog.String("error", err.Error()))
		return false
	}
	return reserved
}
