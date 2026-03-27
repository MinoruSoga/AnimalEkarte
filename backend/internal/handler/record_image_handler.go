package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// verifyMedicalRecordOwnership は clinicID + medicalRecordID の組み合わせを検証し、
// テナント分離を保証するヘルパー。
func (h *Handler) verifyMedicalRecordOwnership(c *gin.Context, clinicID, medicalRecordID uint64) bool {
	if _, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID); err != nil {
		RespondError(c, err)
		return false
	}
	return true
}

// ListRecordImages godoc
// GET /medical-records/:id/images
func (h *Handler) ListRecordImages(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID) {
		return
	}

	images, err := h.svc.RecordImage.List(c.Request.Context(), medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}

	items := make([]recordImageResponse, 0, len(images))
	for i := range images {
		items = append(items, toRecordImageResponse(&images[i]))
	}
	c.JSON(http.StatusOK, items)
}

// CreateRecordImage godoc
// POST /medical-records/:id/images
func (h *Handler) CreateRecordImage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID) {
		return
	}

	var req createRecordImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	imageType := model.MedicalImageType(req.ImageType)
	if imageType == "" {
		imageType = model.MedicalImageTypeOther
	}

	input := &service.CreateRecordImageInput{
		ImageURL:     req.ImageURL,
		ThumbnailURL: req.ThumbnailURL,
		FileName:     req.FileName,
		FileSize:     req.FileSize,
		MimeType:     req.MimeType,
		ImageType:    imageType,
		Description:  req.Description,
		TakenAt:      req.TakenAt,
		ExamID:       req.ExamID,
		StaffID:      req.StaffID,
		SortOrder:    req.SortOrder,
	}

	image, err := h.svc.RecordImage.Create(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toRecordImageResponse(image))
}

// DeleteRecordImage godoc
// DELETE /medical-records/:id/images/:imageId
func (h *Handler) DeleteRecordImage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID) {
		return
	}

	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid image id"))
		return
	}

	if err := h.svc.RecordImage.Delete(c.Request.Context(), medicalRecordID, imageID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterRecordImageRoutes は診療画像関連のルートをmedical-recordsグループに登録する
func (h *Handler) RegisterRecordImageRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/images", h.ListRecordImages)
	rg.POST("/:id/images", h.CreateRecordImage)
	rg.DELETE("/:id/images/:imageId", h.DeleteRecordImage)
}
