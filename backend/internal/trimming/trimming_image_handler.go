package trimming

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
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
type ImageUploadQuotaStore interface {
	Acquire(ctx context.Context, clinicID, staffID uint64, declaredBytes int64) (release func(context.Context), err error)
}

// PermissionRequirement identifies one resource/action pair for an OR authorization gate.
type PermissionRequirement struct {
	Resource string
	Action   string
}

// PermissionAnyMiddleware builds a Gin gate that allows any listed requirement.
// The composition root adapts auth's RequirePermissionAny onto this type so the
// package never imports auth internals (lstep precedent).
type PermissionAnyMiddleware func(...PermissionRequirement) gin.HandlerFunc

// Quota error sentinels produced by adapted quota stores so this package can
// classify Acquire failures without importing the shared store's package
// internals. Unclassified errors map to fail-closed 500.
var (
	ErrImageUploadConcurrency      = errors.New("too many concurrent image uploads")
	ErrImageUploadRateLimit        = errors.New("image upload rate limit exceeded")
	ErrImageUploadByteBudget       = errors.New("image upload byte budget exceeded")
	ErrImageUploadQuotaUnavailable = errors.New("upload quota unavailable")
)

// ImageDependencies carries the optional upload/sign/delete collaborators for
// the trimming image slice. RequireAnyPermission gates the unkeyed upload
// route with an any-of trimming create|edit grant — the same endpoint serves
// both the create form and the edit form before a record exists.
type ImageDependencies struct {
	Uploader             FileUploader
	Quota                ImageUploadQuotaStore
	RequireAnyPermission PermissionAnyMiddleware
}

// WithImageSupport attaches the trimming-image upload/sign/delete collaborators.
// The any-of create|edit write gate is resolved eagerly here; without this call
// the upload route still registers but denies (fail-closed) and stored image
// values pass through unsigned on reads.
func (h *Handler) WithImageSupport(deps ImageDependencies) *Handler {
	h.imageUploader = deps.Uploader
	h.imageQuota = deps.Quota
	if deps.RequireAnyPermission != nil {
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
	// Release must not inherit a canceled request context (lease would stick until stale TTL).
	defer release(context.WithoutCancel(c.Request.Context()))

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

	uploadMeta, err := newTrimmingImageUploadRequest(fileHeader).validate()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := uploadMeta.verifySniffedContent(file); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	storedName, err := uploadMeta.newStoredName(time.Now())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// Upload via FileUploader (S3 or local). Persist/return the object key, never a public URL.
	key := uploadMeta.uploadKey(clinicID, storedName)
	if _, err := h.imageUploader.Upload(c.Request.Context(), key, file, uploadMeta.mimeType); err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to upload file"))
		return
	}

	signedURL, err := h.imageUploader.GetSignedURL(c.Request.Context(), key, trimmingImageSignedURLTTL)
	if err != nil {
		deleteCtx := context.WithoutCancel(c.Request.Context())
		if delErr := h.imageUploader.Delete(deleteCtx, key); delErr != nil {
			slog.WarnContext(deleteCtx, "failed to delete uploaded trimming image on sign failure (best-effort)", "error", delErr, "key", key)
		}
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to sign trimming image url"))
		return
	}

	c.JSON(http.StatusCreated, trimmingImageUploadResponse{Key: key, SignedURL: signedURL})
}

// respondTrimmingImageUploadQuotaError maps quota errors to stable HTTP 429 messages
// (or fail-closed 500 when the quota store is unavailable).
func respondTrimmingImageUploadQuotaError(c *gin.Context, err error) {
	if trimmingImageUploadQuotaLimited(err) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "upload quota unavailable"})
}

func trimmingImageUploadQuotaLimited(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrImageUploadConcurrency) ||
		errors.Is(err, ErrImageUploadRateLimit) ||
		errors.Is(err, ErrImageUploadByteBudget) {
		return true
	}
	message := err.Error()
	return strings.Contains(message, "too many concurrent") ||
		strings.Contains(message, "rate limit exceeded") ||
		strings.Contains(message, "byte budget exceeded")
}
