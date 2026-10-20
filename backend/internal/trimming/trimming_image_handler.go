package trimming

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/imageupload"
	"github.com/animal-ekarte/backend/internal/model"
)

// FileUploader is trimming's consumer-side view of infra.FileUploader (S3 or local
// storage). Declaring it here keeps internal/trimming free of an internal/infra
// import; the composition root passes the concrete implementation by structural typing.
type FileUploader interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
	GetSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// ImageUploadQuotaStore is trimming's consumer-side view of the shared
// clinic/staff image-upload quota store. The composition root supplies the
// shared PostgreSQL-backed store used by the medical-record image pipeline.
// release must be called when the request finishes (success or failure).
type ImageUploadQuotaStore = imageupload.QuotaStore

// PermissionRequirement identifies one resource/action pair for an OR authorization gate.
type PermissionRequirement struct {
	Resource string
	Action   string
}

// PermissionAnyMiddleware builds a Gin gate that allows any listed requirement.
// The composition root adapts auth's RequirePermissionAny onto this type so the
// package never imports auth internals (lstep precedent).
type PermissionAnyMiddleware func(...PermissionRequirement) gin.HandlerFunc

// ImageDependencies carries the optional upload/sign/delete collaborators for
// the trimming image slice. RequireAnyPermission gates the unkeyed upload and
// orphan-cleanup routes with an any-of trimming create|edit grant — the same
// endpoints serve both the create form and the edit form before a record exists.
type ImageDependencies struct {
	Uploader             FileUploader
	Quota                ImageUploadQuotaStore
	CourseImages         CourseImageDrafts
	RequireAnyPermission PermissionAnyMiddleware
}

// WithImageSupport attaches the trimming-image upload/sign/delete collaborators.
// The any-of create|edit write gate is resolved eagerly here; without this call
// the image routes still register but deny (fail-closed) and stored image
// values pass through unsigned on reads.
func (h *Handler) WithImageSupport(deps ImageDependencies) *Handler {
	h.imageUploader = deps.Uploader
	h.imageQuota = deps.Quota
	h.courseImageDrafts = deps.CourseImages
	if deps.RequireAnyPermission != nil {
		h.courseImageWritePermission = deps.RequireAnyPermission(
			PermissionRequirement{Resource: string(model.ResourceMasterTrimming), Action: "create"},
			PermissionRequirement{Resource: string(model.ResourceMasterTrimming), Action: "edit"},
		)
		h.imageWritePermission = deps.RequireAnyPermission(
			PermissionRequirement{Resource: string(model.ResourceTrimming), Action: "create"},
			PermissionRequirement{Resource: string(model.ResourceTrimming), Action: "edit"},
		)
	}
	return h
}

// trimmingImageUploadPermission returns the resolved any-of create|edit gate.
// Fail-closed when the gate was never wired.
func (h *Handler) trimmingImageUploadPermission() gin.HandlerFunc {
	if h.imageWritePermission != nil {
		return h.imageWritePermission
	}
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "trimming image upload is not configured"})
	}
}

// UploadTrimmingImage godoc
// POST /trimmings/images/upload
// multipart/form-data のフォームフィールド名は `file`。レコード未存在でも使える
// 無キー endpoint として create|edit の any-of 権限でゲートされる（routes.go 参照）。
func (h *Handler) UploadCourseImage(c *gin.Context) {
	if _, ok := httpapi.ExtractClinicID(c); !ok {
		return
	}
	if _, ok := httpapi.ExtractStaffID(c); !ok {
		return
	}
	if h.courseImageDrafts == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("course image provenance unavailable"))
		return
	}
	c.Set("courseImageDraftUpload", true)
	h.UploadTrimmingImage(c)
}

func (h *Handler) UploadTrimmingImage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	if c.Request.ContentLength > trimmingImageMaxRequestSize {
		httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("trimming image upload request exceeds size limit"))
		return
	}

	// Fail closed when collaborators are not wired (mirrors the medical-record posture).
	if h.imageUploader == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "image upload not configured"})
		return
	}
	if h.imageQuota == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "upload quota unavailable"})
		return
	}

	// Acquire shared quota BEFORE FormFile / MaxBytesReader body work.
	declaredBytes := c.Request.ContentLength
	if declaredBytes <= 0 {
		// Chunked / unknown length: reserve per-file upper bound conservatively.
		declaredBytes = trimmingImageMaxUploadSize
	}
	release, err := h.imageQuota.Acquire(c.Request.Context(), clinicID, staffID, declaredBytes)
	if err != nil {
		respondTrimmingImageUploadQuotaError(c, err)
		return
	}
	if release == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "upload quota unavailable"})
		return
	}
	// Release must not inherit a canceled request context (lease would stick
	// until stale TTL); bound it so a stalled quota store cannot hold the
	// upload response open past the detached-cleanup deadline.
	defer releaseTrimmingImageQuota(c.Request.Context(), release)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, trimmingImageMaxRequestSize)
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("trimming image upload request exceeds size limit"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapInvalidInput("file field is required"))
		return
	}
	defer file.Close() //nolint:errcheck // multipart ファイルのクローズ失敗は復旧不可のため無視

	uploadMeta, err := trimmingImageConstraints.ValidateFileHeader(fileHeader)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := imageupload.VerifySniffedContent(file, uploadMeta.MIMEType); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	storedName, err := trimmingImageConstraints.NewStoredName(uploadMeta.MIMEType, time.Now())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// Upload via FileUploader (S3 or local). Persist/return the object key, never a public URL.
	key := imageupload.StoredKey(trimmingImageStorageKeyPrefix, clinicID, storedName)
	courseUpload := c.GetBool("courseImageDraftUpload")
	if courseUpload {
		if err := h.courseImageDrafts.ReserveUpload(c.Request.Context(), clinicID, staffID, key); err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	if _, err := h.imageUploader.Upload(c.Request.Context(), key, file, uploadMeta.MIMEType); err != nil {
		if courseUpload {
			err = errors.Join(err, h.protectUnknownCourseUpload(c.Request.Context(), clinicID, staffID, key))
		}
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to upload file"))
		return
	}

	if courseUpload {
		if err := h.courseImageDrafts.CompleteUpload(c.Request.Context(), clinicID, staffID, key); err != nil {
			err = errors.Join(err, h.protectUnknownCourseUpload(c.Request.Context(), clinicID, staffID, key))
			httpapi.RespondError(c, err)
			return
		}
	}
	signedURL, err := h.imageUploader.GetSignedURL(c.Request.Context(), key, trimmingImageSignedURLTTL)
	if err != nil {
		if courseUpload {
			err = errors.Join(err, h.protectUnknownCourseUpload(c.Request.Context(), clinicID, staffID, key))
		}
		if !courseUpload {
			h.bestEffortDeleteTrimmingImageKeys(c.Request.Context(), clinicID, 0, staffID, key)
		}
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to sign trimming image url"))
		return
	}

	c.JSON(http.StatusCreated, trimmingImageUploadResponse{Key: key, SignedURL: signedURL})
}

// The deadline starts when cleanup runs, after slow body/storage work finishes.
// Starting it at Acquire would expire the release lease during a valid upload.
func releaseTrimmingImageQuota(ctx context.Context, release func(context.Context)) {
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), supersededImageCleanupTimeout)
	defer cancel()
	release(releaseCtx)
}

// respondTrimmingImageUploadQuotaError maps quota errors to stable HTTP 429
// messages via typed sentinels (or fail-closed 500 when the quota store is
// unavailable). Internal error details are never echoed to the client.
func respondTrimmingImageUploadQuotaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, imageupload.ErrQuotaConcurrency):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": imageupload.ErrQuotaConcurrency.Error(),
		})
	case errors.Is(err, imageupload.ErrQuotaRateLimit):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": imageupload.ErrQuotaRateLimit.Error(),
		})
	case errors.Is(err, imageupload.ErrQuotaByteBudget):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": imageupload.ErrQuotaByteBudget.Error(),
		})
	default:
		// Infrastructure failure → fail-closed (do not allow unlimited upload).
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "upload quota unavailable",
		})
	}
}

// trimmingImageDeleteRequest は DELETE /trimmings/images の要求。
// フォーム放棄・置換で未永続化になったアップロード済みオブジェクトの掃除に使う。
type trimmingImageDeleteRequest struct {
	Key string `json:"key"`
}

// DeleteTrimmingImage godoc
// DELETE /trimmings/images — body {"key": "trimming-images/<clinic>/<name>"}。
// 自医院スコープのオブジェクトキーのみ削除を許可する。さらにレコード参照が
// 残るキーは 409 で拒否する（署名URLからキーを得た同権限スタッフが保存済み
// 画像を消せないよう fail-closed）。この endpoint は未永続化アップロードの
// best-effort 掃除専用で、永続化済みフィールドの削除は PATCH/DELETE の
// superseded 処理が担う。
func (h *Handler) DeleteTrimmingImage(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unproven image cleanup is deferred"})
}

// courseImagePermission is restricted to the approved master create/edit scope.
func (h *Handler) courseImagePermission() gin.HandlerFunc {
	if h.courseImageWritePermission != nil {
		return h.courseImageWritePermission
	}
	return func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) }
}

// DiscardCourseImage permits only an owned, completed and never-submitted draft.
func (h *Handler) DiscardCourseImage(c *gin.Context) {
	clinic, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	actor, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	if h.courseImageDrafts == nil || h.imageUploader == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("course image provenance unavailable"))
		return
	}
	var req trimmingImageDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("key required"))
		return
	}
	reserved, err := h.courseImageDrafts.ReserveDiscard(c.Request.Context(), clinic, actor, req.Key)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if !reserved {
		httpapi.RespondError(c, apperrors.WrapConflict("course image is protected"))
		return
	}
	// The permanent tombstone has committed before this external operation.
	if err := h.imageUploader.Delete(c.Request.Context(), req.Key); err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "course image delete pending retry"))
		return
	}
	c.Status(http.StatusNoContent)
}

// Upload/sign/response preparation failures do not establish an absent object.
// Preserve the key even when the request has already been cancelled.
func (h *Handler) protectUnknownCourseUpload(ctx context.Context, clinic, actor uint64, key string) error {
	protectCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), supersededImageCleanupTimeout)
	defer cancel()
	err := h.courseImageDrafts.ProtectSubmission(protectCtx, clinic, actor, key)
	if errors.Is(err, apperrors.ErrConflict) {
		return nil
	}
	return err
}
