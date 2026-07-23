package medicalrecord

import (
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
}

// NewMedicalRecordImageHandler initializes a MedicalRecordImageHandler.
func NewMedicalRecordImageHandler(service MedicalRecordImageService, medicalRecord medicalRecordGetter, uploader fileUploader) *MedicalRecordImageHandler {
	return &MedicalRecordImageHandler{service: service, medicalRecord: medicalRecord, uploader: uploader}
}

// verifyOwnership は clinicID + medicalRecordID を検証しテナント分離を保証する（pre-move
// internal/handler.Handler.verifyMedicalRecordOwnership のローカル移植）。検証済みの
// MedicalRecord と成否を返す。
func (h *MedicalRecordImageHandler) verifyOwnership(c *gin.Context, clinicID, medicalRecordID uint64) (*model.MedicalRecord, bool) {
	mr, err := h.medicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return nil, false
	}
	return mr, true
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
	if _, ok := h.verifyOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	images, err := h.service.List(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	items := httpapi.MapSlice(images, toMedicalRecordImageResponse)
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
	if _, ok := h.verifyOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	var req createMedicalRecordImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	image, err := h.service.Create(c.Request.Context(), clinicID, medicalRecordID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/images/%d", medicalRecordID, image.ID))
	c.JSON(http.StatusCreated, toMedicalRecordImageResponse(image))
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
	if _, ok := h.verifyOwnership(c, clinicID, medicalRecordID); !ok {
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
	if _, ok := h.verifyOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	if c.Request.ContentLength > medicalRecordImageMaxRequestSize {
		httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("medical record image upload request exceeds size limit"))
		return
	}
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

	// Upload via FileUploader (S3 or local)
	key := uploadMeta.uploadKey(medicalRecordID, storedName)
	imageURL, err := h.uploader.Upload(c.Request.Context(), key, file, uploadMeta.mimeType)
	if err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to upload file"))
		return
	}

	now := time.Now()
	input := uploadMeta.toUploadedInput(imageURL, now).toServiceInput()

	image, err := h.service.Create(c.Request.Context(), clinicID, medicalRecordID, input)
	if err != nil {
		// Clean up uploaded file on service error (best-effort, non-blocking)
		if delErr := h.uploader.Delete(c.Request.Context(), key); delErr != nil {
			slog.WarnContext(c.Request.Context(), "failed to delete uploaded image on service error (best-effort)", "error", delErr, "key", key)
		}
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/images/%d", medicalRecordID, image.ID))
	c.JSON(http.StatusCreated, toMedicalRecordImageResponse(image))
}
