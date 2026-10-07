package trimming

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type loadedTrimmingImages struct {
	clinicID  uint64
	style     string
	completed string
	loaded    bool
}

func (h *Handler) previousTrimmingImages(
	c *gin.Context,
	clinicID, id uint64,
	input *UpdateTrimmingInput,
) (loadedTrimmingImages, bool) {
	if input == nil || (input.StyleImage == nil && input.CompletedImage == nil) {
		return loadedTrimmingImages{}, true
	}
	existing, err := h.svc.Trimming.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return loadedTrimmingImages{}, false
	}
	style, completed, _ := trimmingStoredImages(existing)
	return loadedTrimmingImages{
		clinicID:  trimmingRecordClinicID(existing, clinicID),
		style:     style,
		completed: completed,
		loaded:    true,
	}, true
}

func (h *Handler) deleteReplacedTrimmingImages(
	ctx context.Context,
	fallbackClinicID uint64,
	previous loadedTrimmingImages,
	updated *model.Reservation,
) {
	if !previous.loaded {
		return
	}
	newStyle, newCompleted, ok := trimmingStoredImages(updated)
	if !ok {
		return
	}
	clinicID := previous.clinicID
	if clinicID == 0 {
		clinicID = trimmingRecordClinicID(updated, fallbackClinicID)
	}
	h.deleteSupersededTrimmingImages(ctx, clinicID, previous.style, previous.completed, newStyle, newCompleted)
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
	h.bestEffortDeleteTrimmingImageKeys(ctx, supersededTrimmingImageKeys(
		clinicID, oldStyle, oldCompleted, newStyle, newCompleted,
	)...)
}

func (h *Handler) bestEffortDeleteTrimmingImageKeys(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if h == nil || h.imageUploader == nil {
		slog.WarnContext(ctx, "skipped superseded trimming image delete; uploader is not configured")
		return
	}
	deleteCtx := context.WithoutCancel(ctx)
	for _, key := range keys {
		if key == "" {
			continue
		}
		if err := h.imageUploader.Delete(deleteCtx, key); err != nil {
			slog.WarnContext(deleteCtx, "failed to delete superseded trimming image",
				slog.String("error", err.Error()),
				slog.String("key", key))
		}
	}
}
