package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListCheckups は指定カルテに紐づく健診記録の一覧を返す
func (h *Handler) ListCheckups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	checkups, err := h.svc.Checkup.List(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(checkups, toCheckupResponse))
}

// CreateCheckup は指定カルテに健診記録を作成する
func (h *Handler) CreateCheckup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req createCheckupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput(clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}

	checkup, err := h.svc.Checkup.Create(c.Request.Context(), id, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/checkups/%d", id, checkup.ID))
	c.JSON(http.StatusCreated, toCheckupResponse(checkup))
}

// UpdateCheckup は健診記録を部分更新する
func (h *Handler) UpdateCheckup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	checkupID, ok := parseIDParam(c, "checkupId")
	if !ok {
		return
	}

	var req updateCheckupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		RespondError(c, err)
		return
	}

	checkup, err := h.svc.Checkup.Update(c.Request.Context(), clinicID, medicalRecordID, checkupID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupResponse(checkup))
}

// DeleteCheckup は健診記録を soft delete する
func (h *Handler) DeleteCheckup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	checkupID, ok := parseIDParam(c, "checkupId")
	if !ok {
		return
	}

	if err := h.svc.Checkup.Delete(c.Request.Context(), clinicID, medicalRecordID, checkupID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListGlobalCheckups は GET /v1/checkups — クリニック横断の健診記録一覧を返す
func (h *Handler) ListGlobalCheckups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	query := newListGlobalCheckupsQuery(clinicID, c.Request.URL.Query())

	checkups, err := h.svc.Checkup.ListByClinic(c.Request.Context(), query.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	responses := mapSlice(checkups, toCheckupGlobalResponse)
	c.JSON(http.StatusOK, newPaginatedResponse(responses, int64(len(responses)), 1, len(responses)))
}

// GetCheckupAlerts は GET /v1/checkups/alerts?within_days=30 — 健診期限アラート集計
func (h *Handler) GetCheckupAlerts(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	withinDays, err := newCheckupAlertsQuery(c.Request.URL.Query()).toWithinDays()
	if err != nil {
		RespondError(c, err)
		return
	}
	result, err := h.svc.Checkup.GetAlerts(c.Request.Context(), clinicID, withinDays)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupAlertsResponse(result))
}

// RegisterGlobalCheckupRoutes は /checkups トップレベルルートを登録する
func (h *Handler) RegisterGlobalCheckupRoutes(rg *gin.RouterGroup) {
	checkups := rg.Group("/checkups")
	checkups.GET("", h.RequirePermission(string(model.ResourceCheckups), "view"), h.ListGlobalCheckups)
	checkups.GET("/alerts", h.RequirePermission(string(model.ResourceCheckups), "view"), h.GetCheckupAlerts)
	// #211: pet 単位の健診結果（飼い主レポート用）。
	checkups.GET("/field-results", h.RequirePermission(string(model.ResourceCheckups), "view"), h.ListPetCheckupResults)
}

// RegisterCheckupRoutes は健診記録関連のルートを登録する
// RegisterCheckupRoutes はカルテ内の定期健診サブリソースルートを登録する。
// 子リソースは親（medical-records）の権限に従う（BUG-133: vitals/treatments 等と統一）。
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/checkups", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListCheckups)
	rg.POST("/:id/checkups", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateCheckup)
	rg.PATCH("/:id/checkups/:checkupId", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateCheckup)
	rg.DELETE("/:id/checkups/:checkupId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteCheckup)
	// #211: 健診パッケージの型付き結果値（サブリソース。親 medical-records 権限に従う）。
	rg.GET("/:id/checkups/:checkupId/field-results", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListCheckupFieldResults)
	rg.PUT("/:id/checkups/:checkupId/field-results", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.ReplaceCheckupFieldResults)
}
