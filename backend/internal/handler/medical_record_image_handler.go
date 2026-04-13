package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

const (
	maxUploadSize = 10 * 1024 * 1024 // 10MB
)

var allowedMIMETypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"application/pdf": true,
}

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
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	images, err := h.svc.MedicalRecordImage.List(c.Request.Context(), medicalRecordID)
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
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
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

	imageType := model.MedicalImageType(req.ImageType)
	if imageType == "" {
		imageType = model.MedicalImageTypeOther
	}

	input := &service.CreateMedicalRecordImageInput{
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

	image, err := h.svc.MedicalRecordImage.Create(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toMedicalRecordImageResponse(image))
}

// DeleteMedicalRecordImage godoc
// DELETE /medical-records/:id/images/:imageId
func (h *Handler) DeleteMedicalRecordImage(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid image id"))
		return
	}

	if err := h.svc.MedicalRecordImage.Delete(c.Request.Context(), medicalRecordID, imageID); err != nil {
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
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
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

	// Validate file size
	if fileHeader.Size > maxUploadSize {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("file size exceeds limit of %dMB", maxUploadSize/1024/1024)))
		return
	}

	// Detect MIME type from Content-Type header or file extension
	fileExt := strings.ToLower(filepath.Ext(fileHeader.Filename))
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" || !allowedMIMETypes[mimeType] {
		// Fall back to extension-based detection
		switch fileExt {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".pdf":
			mimeType = "application/pdf"
		default:
			RespondError(c, apperrors.WrapInvalidInput("unsupported file type; allowed: jpeg, png, gif, pdf"))
			return
		}
	}
	if !allowedMIMETypes[mimeType] {
		RespondError(c, apperrors.WrapInvalidInput("unsupported MIME type: "+mimeType))
		return
	}

	// Generate unique filename to avoid collisions
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		RespondError(c, apperrors.Wrap(err, "failed to generate unique filename"))
		return
	}
	storedName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes), fileExt)

	// Upload via FileUploader (S3 or local)
	key := fmt.Sprintf("medical-records/%d/%s", medicalRecordID, storedName)
	imageURL, err := h.uploader.Upload(c.Request.Context(), key, file, mimeType)
	if err != nil {
		RespondError(c, apperrors.Wrap(err, "failed to upload file"))
		return
	}

	now := time.Now()
	input := &service.CreateMedicalRecordImageInput{
		ImageURL:  imageURL,
		FileName:  fileHeader.Filename,
		FileSize:  fileHeader.Size,
		MimeType:  mimeType,
		ImageType: model.MedicalImageTypeOther,
		TakenAt:   &now,
	}

	image, err := h.svc.MedicalRecordImage.Create(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		// Clean up uploaded file on service error (non-fatal)
		if removeErr := h.uploader.Delete(c.Request.Context(), key); removeErr != nil {
			slog.WarnContext(c.Request.Context(), "failed to clean up uploaded file", "key", key, "error", removeErr)
		}
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toMedicalRecordImageResponse(image))
}

// RegisterMedicalRecordImageRoutes は診療画像関連のルートをmedical-recordsグループに登録する
func (h *Handler) RegisterMedicalRecordImageRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/images", h.ListMedicalRecordImages)
	rg.POST("/:id/images", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateMedicalRecordImage)
	rg.POST("/:id/images/upload", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.UploadMedicalRecordImage)
	rg.DELETE("/:id/images/:imageId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteMedicalRecordImage)
}
