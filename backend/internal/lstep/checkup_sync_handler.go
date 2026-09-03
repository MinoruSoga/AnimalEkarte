package lstep

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// GetCheckupSyncPreview godoc
// GET /clinics/:clinic_id/lstep/checkup-sync/preview — 健診対象者プレビューを返す（BE-004 / ISSUE-009）。
func (h *Handler) GetCheckupSyncPreview(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	q := newCheckupSyncPreviewQuery(c.Request.URL.Query())
	input, err := q.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// ISSUE-010: 抽出メタデータを audit_logs に永続化するため actorID を service に渡す。
	actorID := &staffID

	result, err := h.checkupSync.PreviewCheckupSync(c.Request.Context(), clinicID, input, actorID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupSyncPreviewResponse(result))
}

// CreateCheckupSync godoc
// POST /clinics/:clinic_id/lstep/checkup-sync — 選択した飼い主に健診リマインダータグを一括付与する（BE-004）。
func (h *Handler) CreateCheckupSync(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req checkupSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	actorID := &staffID

	result, err := h.checkupSync.CreateCheckupSync(c.Request.Context(), clinicID, input, actorID)
	if err != nil {
		httpapi.RespondError(c, err)
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
