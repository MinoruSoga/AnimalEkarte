package lstep

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
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

	if req.MessageType == "text" && len([]rune(req.Text)) > 5000 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("テキストは5000文字以内で入力してください"))
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
//
// 旧 line_send_logs は DROP 済みで、置換先の配送記録（line_dispatch_intents / attempts / results）には
// reader がまだ無い（EMR-504 C）。履歴なしを「送信なし」と見せないため、権限確認後に固定の 422 で
// fail-closed する（成功を装わない）。reader が接続されたら本 handler を差し替える。
func (h *LineSendHandler) GetLineSendLogs(c *gin.Context) {
	if _, ok := httpapi.ExtractClinicID(c); !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceOwners), "view") {
		return
	}
	if _, ok := httpapi.ParseIDParam(c, "id"); !ok {
		return
	}
	httpapi.RespondError(c, apperrors.WrapUnprocessableWithCode(
		apperrors.CodeLineDispatchPending,
		"LINE送信履歴の参照先（配送記録）がまだ接続されていません",
	))
}
