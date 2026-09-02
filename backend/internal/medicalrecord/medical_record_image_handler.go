package medicalrecord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordImageHandler serves the medical-record-image (診療画像) HTTP boundary nested under a
// medical record. It holds the MedicalRecordImageService, a medicalRecordGetter for the
// tenant-ownership guard (pre-move verifyMedicalRecordOwnership), and a fileUploader for the
// multipart upload endpoint (S3 or local storage).
type MedicalRecordImageHandler struct {
	service       MedicalRecordImageService
	medicalRecord medicalRecordGetter
	uploader      fileUploader
	quota         medicalRecordImageUploadQuotaStore
}

// NewMedicalRecordImageHandler initializes a MedicalRecordImageHandler.
// quota is required for uploads; nil fails closed. Optional trailing arg keeps
// route-snapshot call sites compiling during migration.
func NewMedicalRecordImageHandler(
	service MedicalRecordImageService,
	medicalRecord medicalRecordGetter,
	uploader fileUploader,
	quota ...medicalRecordImageUploadQuotaStore,
) *MedicalRecordImageHandler {
	var q medicalRecordImageUploadQuotaStore
	if len(quota) > 0 {
		q = quota[0]
	}
	return &MedicalRecordImageHandler{
		service:       service,
		medicalRecord: medicalRecord,
		uploader:      uploader,
		quota:         q,
	}
}

// verifyOwnership は clinicID + medicalRecordID を検証しテナント分離を保証する（pre-move
// internal/handler.Handler.verifyMedicalRecordOwnership のローカル移植）。検証済みの
// MedicalRecord と成否を返す。
func (h *MedicalRecordImageHandler) verifyOwnership(c *gin.Context, clinicID, medicalRecordID uint64) bool {
	_, err := h.medicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return false
	}
	return true
}

// ListMedicalRecordImages godoc
// GET /medical-records/:id/images
func (h *MedicalRecordImageHandler) ListMedicalRecordImages(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyOwnership(c, clinicID, medicalRecordID) {
		return
	}

	images, err := h.service.List(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	items := make([]medicalRecordImageResponse, 0, len(images))
	for i := range images {
		item, signErr := h.signedMedicalRecordImageResponse(c.Request.Context(), &images[i])
		if signErr != nil {
			httpapi.RespondError(c, apperrors.Wrap(signErr, "failed to sign medical record image url"))
			return
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, items)
}

// CreateMedicalRecordImage godoc
// POST /medical-records/:id/images
func (h *MedicalRecordImageHandler) CreateMedicalRecordImage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyOwnership(c, clinicID, medicalRecordID) {
		return
	}

	var req createMedicalRecordImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	// MRC-09: JSON create must share MIME/size allowlist with upload path.
	if err := req.validateJSONCreate(); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	input := req.toServiceInput()
	signedImageURL, err := h.signStoredMedicalRecordImageURL(c.Request.Context(), medicalRecordID, input.ImageURL)
	if err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to sign medical record image url"))
		return
	}
	signedThumbURL, err := h.signStoredMedicalRecordImageURL(c.Request.Context(), medicalRecordID, input.ThumbnailURL)
	if err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to sign medical record image url"))
		return
	}

	image, err := h.service.Create(c.Request.Context(), clinicID, medicalRecordID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	resp := toMedicalRecordImageResponse(image)
	resp.ImageURL = signedImageURL
	resp.ThumbnailURL = signedThumbURL
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/images/%d", medicalRecordID, image.ID))
	c.JSON(http.StatusCreated, resp)
}

// DeleteMedicalRecordImage godoc
// DELETE /medical-records/:id/images/:imageId
func (h *MedicalRecordImageHandler) DeleteMedicalRecordImage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyOwnership(c, clinicID, medicalRecordID) {
		return
	}

	imageID, ok := httpapi.ParseIDParam(c, "imageId")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), clinicID, medicalRecordID, imageID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadMedicalRecordImage godoc
// POST /medical-records/:id/images/upload
// Accepts multipart/form-data with field name "file".
func (h *MedicalRecordImageHandler) UploadMedicalRecordImage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	if !h.verifyOwnership(c, clinicID, medicalRecordID) {
		return
	}

	if c.Request.ContentLength > medicalRecordImageMaxRequestSize {
		httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("medical record image upload request exceeds size limit"))
		return
	}

	// SEC-CS-F08-R1: acquire shared quota BEFORE FormFile / MaxBytesReader body work.
	if h.quota == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "upload quota unavailable"})
		return
	}
	declaredBytes := c.Request.ContentLength
	if declaredBytes <= 0 {
		// Chunked / unknown length: reserve per-file upper bound conservatively.
		declaredBytes = medicalRecordImageMaxUploadSize
	}
	release, err := h.quota.Acquire(c.Request.Context(), clinicID, staffID, declaredBytes)
	if err != nil {
		respondMedicalRecordImageUploadQuotaError(c, err)
		return
	}
	// Release must not inherit a canceled request context (lease would stick until stale TTL).
	defer release(context.WithoutCancel(c.Request.Context()))

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, medicalRecordImageMaxRequestSize)
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("medical record image upload request exceeds size limit"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapInvalidInput("file field is required"))
		return
	}
	defer file.Close() //nolint:errcheck // multipart ファイルのクローズ失敗は復旧不可のため無視

	uploadMeta, err := newMedicalRecordImageUploadRequest(fileHeader).validate()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	storedName, err := uploadMeta.newStoredName(time.Now())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// Upload via FileUploader (S3 or local). Persist the object key, never a public URL.
	key := uploadMeta.uploadKey(medicalRecordID, storedName)
	if _, err := h.uploader.Upload(c.Request.Context(), key, file, uploadMeta.mimeType); err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to upload file"))
		return
	}

	cleanupUpload := func() {
		if delErr := h.uploader.Delete(c.Request.Context(), key); delErr != nil {
			slog.WarnContext(c.Request.Context(), "failed to delete uploaded image on service error (best-effort)", "error", delErr, "key", key)
		}
	}

	signedImageURL, err := h.signStoredMedicalRecordImageURL(c.Request.Context(), medicalRecordID, key)
	if err != nil {
		cleanupUpload()
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to sign medical record image url"))
		return
	}

	now := time.Now()
	input := uploadMeta.toUploadedInput(key, now).toServiceInput()

	image, err := h.service.Create(c.Request.Context(), clinicID, medicalRecordID, input)
	if err != nil {
		cleanupUpload()
		httpapi.RespondError(c, err)
		return
	}
	resp := toMedicalRecordImageResponse(image)
	resp.ImageURL = signedImageURL
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/images/%d", medicalRecordID, image.ID))
	c.JSON(http.StatusCreated, resp)
}

func (h *MedicalRecordImageHandler) signedMedicalRecordImageResponse(ctx context.Context, img *model.MedicalRecordImage) (medicalRecordImageResponse, error) {
	resp := toMedicalRecordImageResponse(img)
	imageURL, err := h.signStoredMedicalRecordImageURL(ctx, img.MedicalRecordID, resp.ImageURL)
	if err != nil {
		return medicalRecordImageResponse{}, err
	}
	resp.ImageURL = imageURL
	thumbURL, err := h.signStoredMedicalRecordImageURL(ctx, img.MedicalRecordID, resp.ThumbnailURL)
	if err != nil {
		return medicalRecordImageResponse{}, err
	}
	resp.ThumbnailURL = thumbURL
	return resp, nil
}

func (h *MedicalRecordImageHandler) signStoredMedicalRecordImageURL(ctx context.Context, medicalRecordID uint64, stored string) (string, error) {
	key, isObject, err := resolveMedicalRecordImageStorageKey(stored)
	if err != nil {
		return "", err
	}
	if !isObject {
		return stored, nil
	}
	if !medicalRecordImageKeyBelongsToRecord(key, medicalRecordID) {
		if medicalRecordImageHasHTTPScheme(stored) {
			return stored, nil
		}
		return "", errMedicalRecordImageStorageKeyUnresolved
	}
	if h.uploader == nil {
		return "", fmt.Errorf("file uploader is not configured")
	}
	signed, err := h.uploader.GetSignedURL(ctx, key, medicalRecordImageSignedURLTTL)
	if err != nil {
		return "", err
	}
	if signed == "" {
		return "", fmt.Errorf("signed url is empty")
	}
	return signed, nil
}

// respondMedicalRecordImageUploadQuotaError maps quota errors to stable HTTP 429 messages
// (or fail-closed 500 when the quota store is unavailable).
func respondMedicalRecordImageUploadQuotaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errMedicalRecordImageUploadConcurrency):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": errMedicalRecordImageUploadConcurrency.Error(),
		})
	case errors.Is(err, errMedicalRecordImageUploadRateLimit):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": errMedicalRecordImageUploadRateLimit.Error(),
		})
	case errors.Is(err, errMedicalRecordImageUploadByteBudget):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": errMedicalRecordImageUploadByteBudget.Error(),
		})
	default:
		// Infrastructure failure → fail-closed (do not allow unlimited upload).
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "upload quota unavailable",
		})
	}
}
