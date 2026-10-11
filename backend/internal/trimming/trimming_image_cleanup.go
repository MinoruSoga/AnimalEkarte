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
	h.deleteSupersededTrimmingImages(ctx, clinicID, replaced.Style, replaced.Completed, newStyle, newCompleted)
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
	clinicID uint64,
	oldStyle, oldCompleted, newStyle, newCompleted string,
) {
	h.bestEffortDeleteTrimmingImageKeys(ctx, clinicID, supersededTrimmingImageKeys(
		clinicID, oldStyle, oldCompleted, newStyle, newCompleted,
	)...)
}

// supersededImageCleanupTimeout bounds post-commit object cleanup so a stalled
// S3/R2 delete cannot hold the mutation response open (client timeout/retry
// while the record is already committed). 削除失敗は warn のみで孤立扱い。
const supersededImageCleanupTimeout = 10 * time.Second

func (h *Handler) bestEffortDeleteTrimmingImageKeys(ctx context.Context, clinicID uint64, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if h == nil || h.imageUploader == nil {
		slog.WarnContext(ctx, "skipped superseded trimming image delete; uploader is not configured")
		return
	}
	if h.svc == nil || h.svc.Trimming == nil {
		slog.WarnContext(ctx, "skipped superseded trimming image delete; reference check is not configured")
		return
	}
	deleteCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), supersededImageCleanupTimeout)
	defer cancel()
	for _, key := range keys {
		if key == "" {
			continue
		}
		// セッションスコープ advisory lock 下で「参照再確認→オブジェクト削除
		// →unlock」を不可分に実行する。durable tombstone はオブジェクト自体
		// の不在 — ロック解放後に進んだ writer は Exists=false を見て拒否
		// される。確認不能・参照あり・削除無効（rollout gate）はすべて保持側
		// fail-closed — 参照中オブジェクトの損失より孤立の方が安全。
		deleted, err := h.svc.Trimming.DeleteUnreferencedImageObject(
			deleteCtx, clinicID, key, h.imageUploader.Delete)
		if err != nil {
			slog.WarnContext(deleteCtx, "skipped superseded trimming image delete; lifecycle guard failed",
				slog.String("error", err.Error()),
				slog.String("key", key))
			continue
		}
		if !deleted {
			slog.InfoContext(deleteCtx,
				"retained trimming image object; still referenced or deletion not enabled",
				slog.String("key", key))
		}
	}
}
