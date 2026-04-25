package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListSharedFiles godoc
// GET /api/v1/shared-files
func (h *Handler) ListSharedFiles(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	files, err := h.svc.SharedFile.FindAll(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := make([]sharedFileResponse, 0, len(files))
	for _, f := range files {
		resp = append(resp, toSharedFileResponse(f))
	}
	c.JSON(http.StatusOK, resp)
}

// UploadSharedFile godoc
// POST /api/v1/shared-files
// Content-Type: multipart/form-data
func (h *Handler) UploadSharedFile(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	// サイズ制限を適用（LINE Messaging API 上限 10MB）
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, model.MaxSharedFileSizeBytes)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("file field is required"))
		return
	}
	defer file.Close() //nolint:errcheck

	// 拡張子チェック（.pdf, .jpg, .jpeg, .png のみ許可）
	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentType, allowed := model.AllowedFileExtensions[ext]
	if !allowed {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("file type %q is not allowed", ext)))
		return
	}
	fileType := model.AllowedFileTypes[ext]

	var req uploadSharedFileRequest
	if err := c.ShouldBind(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if req.Purpose == "" {
		req.Purpose = model.SharedFilePurposeOther
	}

	resp, err := h.svc.SharedFile.Upload(c.Request.Context(), clinicID, staffID, service.UploadSharedFileInput{
		Content:     file,
		FileName:    header.Filename,
		ContentType: contentType,
		FileType:    fileType,
		FileSize:    header.Size,
		Purpose:     req.Purpose,
		OwnerID:     req.OwnerID,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/shared-files/%d", resp.ID))
	c.JSON(http.StatusCreated, toSharedFileResponse(resp))
}

// GetSharedFileSignedURL godoc
// GET /api/v1/shared-files/:id/signed-url
func (h *Handler) GetSharedFileSignedURL(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	url, err := h.svc.SharedFile.GetSignedURL(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, sharedFileSignedURLResponse{SignedURL: url})
}

// DeleteSharedFile godoc
// DELETE /api/v1/shared-files/:id
func (h *Handler) DeleteSharedFile(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.SharedFile.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterSharedFileRoutes はファイル共有のルートを登録する。
func (h *Handler) RegisterSharedFileRoutes(rg *gin.RouterGroup) {
	sf := rg.Group("/shared-files")
	sf.GET("", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.ListSharedFiles)
	sf.POST("", h.RequirePermission(string(model.ResourceHospitalSettings), "create"), h.UploadSharedFile)
	sf.GET("/:id/signed-url", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.GetSharedFileSignedURL)
	sf.DELETE("/:id", h.RequirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteSharedFile)
}
