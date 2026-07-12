package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetCheckupSyncPreview godoc
// GET /clinics/:clinic_id/lstep/checkup-sync/preview — 健診対象者プレビューを返す（BE-004 / ISSUE-009）。
func (h *Handler) GetCheckupSyncPreview(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	q := newCheckupSyncPreviewQuery(c.Request.URL.Query())
	input, err := q.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	// ISSUE-010: 抽出メタデータを audit_logs に永続化するため actorID を service に渡す。
	var actorID *uint64
	if staffID, ok := extractStaffID(c); ok {
		actorID = &staffID
	}

	result, err := h.svc.CheckupSync.PreviewCheckupSync(c.Request.Context(), clinicID, input, actorID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupSyncPreviewResponse(result))
}

// CreateCheckupSync godoc
// POST /clinics/:clinic_id/lstep/checkup-sync — 選択した飼い主に健診リマインダータグを一括付与する（BE-004）。
func (h *Handler) CreateCheckupSync(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req checkupSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	var actorID *uint64
	if staffID, ok := extractStaffID(c); ok {
		actorID = &staffID
	}

	result, err := h.svc.CheckupSync.CreateCheckupSync(c.Request.Context(), clinicID, input, actorID)
	if err != nil {
		RespondError(c, err)
		return
	}

	failedStrings := make([]string, 0, len(result.FailedOwnerIDs))
	for _, id := range result.FailedOwnerIDs {
		failedStrings = append(failedStrings, strconv.FormatUint(id, 10))
	}

	c.JSON(http.StatusOK, checkupSyncResultResponse{
		SuccessCount:   result.SuccessCount,
		SkippedCount:   result.SkippedCount,
		FailedCount:    result.FailedCount,
		FailedOwnerIDs: failedStrings,
	})
}

// RegisterCheckupSyncRoutes は健診同期関連ルートを登録する（BE-004）。
func (h *Handler) RegisterCheckupSyncRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics/:clinic_id")
	clinics.GET("/lstep/checkup-sync/preview", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetCheckupSyncPreview)
	clinics.POST("/lstep/checkup-sync", h.RequirePermission(string(model.ResourceOwners), "edit"), h.CreateCheckupSync)
}
