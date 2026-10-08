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
	fallbackClinicID uint64,
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
	h.deleteSupersededTrimmingImages(ctx, clinicID, updated.ID, replaced.Style, replaced.Completed, newStyle, newCompleted)
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
	clinicID, exceptAppointmentID uint64,
	oldStyle, oldCompleted, newStyle, newCompleted string,
) {
	h.bestEffortDeleteTrimmingImageKeys(ctx, clinicID, exceptAppointmentID, supersededTrimmingImageKeys(
		clinicID, oldStyle, oldCompleted, newStyle, newCompleted,
	)...)
}

// supersededImageCleanupTimeout bounds post-commit object cleanup so a stalled
// S3/R2 delete cannot hold the mutation response open (client timeout/retry
// while the record is already committed). 削除失敗は warn のみで孤立扱い。
const supersededImageCleanupTimeout = 10 * time.Second

func (h *Handler) bestEffortDeleteTrimmingImageKeys(ctx context.Context, clinicID, exceptAppointmentID uint64, keys ...string) {
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
		// 他レコードが同一キーを共有していれば保持する（create/update は医院
		// prefix のみ検査するため、同一キー参照は起こり得る）。確認できない
		// 場合も消さず warn — 参照中オブジェクトの損失より孤立の方が安全。
		if !h.imageKeyFreeToDelete(deleteCtx, clinicID, key, exceptAppointmentID) {
			continue
		}
		if err := h.imageUploader.Delete(deleteCtx, key); err != nil {
			slog.WarnContext(deleteCtx, "failed to delete superseded trimming image",
				slog.String("error", err.Error()),
				slog.String("key", key))
		}
	}
}

// imageKeyFreeToDelete は key が exceptAppointmentID 以外のどのレコードからも
// 参照されていない場合のみ true を返す。参照確認の経路が無い/失敗した場合は
// false（保持側 fail-closed）。
func (h *Handler) imageKeyFreeToDelete(ctx context.Context, clinicID uint64, key string, exceptAppointmentID uint64) bool {
	if h.svc == nil || h.svc.Trimming == nil {
		slog.WarnContext(ctx, "skipped superseded trimming image delete; reference check is not configured",
			slog.String("key", key))
		return false
	}
	inUse, err := h.svc.Trimming.ImageKeyInUse(ctx, clinicID, key, exceptAppointmentID)
	if err != nil {
		slog.WarnContext(ctx, "skipped superseded trimming image delete; reference check failed",
			slog.String("error", err.Error()),
			slog.String("key", key))
		return false
	}
	if inUse {
		slog.InfoContext(ctx, "retained trimming image object; still referenced by another record",
			slog.String("key", key))
		return false
	}
	return true
}
