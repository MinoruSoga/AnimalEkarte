package lstep

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// LineSendHandler は LineSendService の HTTP handler。
type LineSendHandler struct {
	svc               LineSendService
	requirePermission PermissionMiddleware
}

// NewLineSendHandler は LineSendHandler を構築する。
func NewLineSendHandler(svc LineSendService, requirePermission PermissionMiddleware) *LineSendHandler {
	return &LineSendHandler{svc: svc, requirePermission: requirePermission}
}

// SendLineMessage godoc
func (h *LineSendHandler) SendLineMessage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req lineSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	if req.MessageType == "text" && len([]rune(req.Text)) > 500 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("テキストは500文字以内で入力してください"))
		return
	}

	result, err := h.svc.Send(c.Request.Context(), clinicID, req.toServiceInput(ownerID, staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, lineSendResponse{
		Sent:     true,
		SentAt:   httpapi.LocalTimeRFC3339(result.SentAt),
		TagAdded: result.TagAdded,
	})
}

// GetLineSendLogs godoc
func (h *LineSendHandler) GetLineSendLogs(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	logs, err := h.svc.GetSendLogs(c.Request.Context(), clinicID, ownerID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, lineSendLogListResponse{Items: toLineSendLogListResponse(logs)})
}
