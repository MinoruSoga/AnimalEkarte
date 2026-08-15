package lstep

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListSharedFiles godoc
// GET /api/v1/shared-files
func (h *Handler) ListSharedFiles(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	files, err := h.sharedFile.FindAll(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
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
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	// サイズ制限を適用（LINE Messaging API 上限 10MB）
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, model.MaxSharedFileSizeBytes)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("file field is required"))
		return
	}
	defer file.Close() //nolint:errcheck // multipart file close error is not actionable

	meta, err := newSharedFileUploadMeta(header)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	var req uploadSharedFileRequest
	if err := c.ShouldBind(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput(file, meta)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	resp, err := h.sharedFile.Upload(c.Request.Context(), clinicID, staffID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/shared-files/%d", resp.ID))
	c.JSON(http.StatusCreated, toSharedFileResponse(resp))
}

// GetSharedFileSignedURL godoc
// GET /api/v1/shared-files/:id/signed-url
func (h *Handler) GetSharedFileSignedURL(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	url, err := h.sharedFile.GetSignedURL(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, sharedFileSignedURLResponse{SignedURL: url})
}

// DeleteSharedFile godoc
// DELETE /api/v1/shared-files/:id
func (h *Handler) DeleteSharedFile(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.sharedFile.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterSharedFileRoutes はファイル共有のルートを登録する。
// POST は個別送信用ファイルアップロードをサポートするため、owners:edit または medical-records:create/edit 権限で実行可能。
// GET list/signed-url は管理用途のため hospital-settings:view 権限で保護。
// DELETE は管理用途のため hospital-settings:delete 権限で保護。
func (h *Handler) RegisterSharedFileRoutes(rg *gin.RouterGroup) {
	sf := rg.Group("/shared-files")
	sf.GET("", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.ListSharedFiles)
	sf.POST("", h.requireAnyPermission(
		PermissionRequirement{Resource: string(model.ResourceOwners), Action: "edit"},
		PermissionRequirement{Resource: string(model.ResourceMedicalRecords), Action: "create"},
		PermissionRequirement{Resource: string(model.ResourceMedicalRecords), Action: "edit"},
	), h.UploadSharedFile)
	sf.GET("/:id/signed-url", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.GetSharedFileSignedURL)
	sf.DELETE("/:id", h.requirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteSharedFile)
}
