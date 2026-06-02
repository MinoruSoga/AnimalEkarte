package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// verifyMedicalRecordOwnership は clinicID + medicalRecordID の組み合わせを検証し、
// テナント分離を保証するヘルパー。検証済みの MedicalRecord と成否を返す。
func (h *Handler) verifyMedicalRecordOwnership(c *gin.Context, clinicID, medicalRecordID uint64) (*model.MedicalRecord, bool) {
	mr, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return nil, false
	}
	return mr, true
}

// ListMedicalRecordImages godoc
// GET /medical-records/:id/images
func (h *Handler) ListMedicalRecordImages(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	images, err := h.svc.MedicalRecordImage.List(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}

	items := make([]medicalRecordImageResponse, 0, len(images))
	for i := range images {
		items = append(items, toMedicalRecordImageResponse(&images[i]))
	}
	c.JSON(http.StatusOK, items)
}

// CreateMedicalRecordImage godoc
// POST /medical-records/:id/images
func (h *Handler) CreateMedicalRecordImage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	var req createMedicalRecordImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	image, err := h.svc.MedicalRecordImage.Create(c.Request.Context(), clinicID, medicalRecordID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/images/%d", medicalRecordID, image.ID))
	c.JSON(http.StatusCreated, toMedicalRecordImageResponse(image))
}

// DeleteMedicalRecordImage godoc
// DELETE /medical-records/:id/images/:imageId
func (h *Handler) DeleteMedicalRecordImage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	imageID, ok := parseIDParam(c, "imageId")
	if !ok {
		return
	}

	if err := h.svc.MedicalRecordImage.Delete(c.Request.Context(), clinicID, medicalRecordID, imageID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadMedicalRecordImage godoc
// POST /medical-records/:id/images/upload
// Accepts multipart/form-data with field name "file".
func (h *Handler) UploadMedicalRecordImage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("file field is required"))
		return
	}
	defer file.Close() //nolint:errcheck // multipart ファイルのクローズ失敗は復旧不可のため無視

	uploadMeta, err := newMedicalRecordImageUploadRequest(fileHeader).validate()
	if err != nil {
		RespondError(c, err)
		return
	}
	storedName, err := uploadMeta.newStoredName(time.Now())
	if err != nil {
		RespondError(c, err)
		return
	}

	// Upload via FileUploader (S3 or local)
	key := uploadMeta.uploadKey(medicalRecordID, storedName)
	imageURL, err := h.uploader.Upload(c.Request.Context(), key, file, uploadMeta.mimeType)
	if err != nil {
		RespondError(c, apperrors.Wrap(err, "failed to upload file"))
		return
	}

	now := time.Now()
	input := uploadMeta.toUploadedInput(imageURL, now).toServiceInput()

	image, err := h.svc.MedicalRecordImage.Create(c.Request.Context(), clinicID, medicalRecordID, input)
	if err != nil {
		// Clean up uploaded file on service error (best-effort, non-blocking)
		if delErr := h.uploader.Delete(c.Request.Context(), key); delErr != nil {
			slog.WarnContext(c.Request.Context(), "failed to delete uploaded image on service error (best-effort)", "error", delErr, "key", key)
		}
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/images/%d", medicalRecordID, image.ID))
	c.JSON(http.StatusCreated, toMedicalRecordImageResponse(image))
}

// RegisterMedicalRecordImageRoutes は診療画像関連のルートをmedical-recordsグループに登録する
func (h *Handler) RegisterMedicalRecordImageRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/images", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListMedicalRecordImages)
	rg.POST("/:id/images", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateMedicalRecordImage)
	rg.POST("/:id/images/upload", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.UploadMedicalRecordImage)
	rg.DELETE("/:id/images/:imageId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteMedicalRecordImage)
}
