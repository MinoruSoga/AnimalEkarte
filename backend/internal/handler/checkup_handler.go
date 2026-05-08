package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
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

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
		return
	}
	var nextDate *time.Time
	if req.NextDate != nil && *req.NextDate != "" {
		nd, err2 := time.Parse("2006-01-02", *req.NextDate)
		if err2 != nil {
			RespondError(c, apperrors.WrapInvalidInput("next_date must be YYYY-MM-DD format"))
			return
		}
		nextDate = &nd
	}

	checkup, err := h.svc.Checkup.Create(c.Request.Context(), id, &service.CreateCheckupInput{
		ClinicID:      clinicID,
		CheckupTypeID: req.CheckupTypeID,
		PetID:         req.PetID,
		Date:          date,
		NextDate:      nextDate,
		DoctorID:      req.DoctorID,
		Result:        req.Result,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-008: 健診タグ同期（best-effort）
	if checkup.MedicalRecord != nil && checkup.MedicalRecord.OwnerID != nil {
		_ = h.svc.LstepTagSync.SyncCheckupTag(c.Request.Context(), clinicID, *checkup.MedicalRecord.OwnerID, checkup.CheckupTypeID, checkup.Date, checkup.NextDate)
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

	var updateDate *time.Time
	if req.Date != nil && *req.Date != "" {
		d, err2 := time.Parse("2006-01-02", *req.Date)
		if err2 != nil {
			RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
			return
		}
		updateDate = &d
	}
	var updateNextDate *time.Time
	if req.NextDate != nil && *req.NextDate != "" {
		nd, err2 := time.Parse("2006-01-02", *req.NextDate)
		if err2 != nil {
			RespondError(c, apperrors.WrapInvalidInput("next_date must be YYYY-MM-DD format"))
			return
		}
		updateNextDate = &nd
	}

	checkup, err := h.svc.Checkup.Update(c.Request.Context(), clinicID, medicalRecordID, checkupID, &service.UpdateCheckupInput{
		CheckupTypeID: req.CheckupTypeID,
		PetID:         req.PetID,
		Date:          updateDate,
		NextDate:      updateNextDate,
		DoctorID:      req.DoctorID,
		Result:        req.Result,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	// ISSUE-004: 更新後は DB 全体から checkup_done_*/next_checkup_* を再構築（best-effort）。
	// type_id 変更時に旧 typeID のタグが残らない。
	if checkup.MedicalRecord != nil && checkup.MedicalRecord.OwnerID != nil {
		_ = h.svc.LstepTagSync.ResyncOwnerCheckupTags(c.Request.Context(), clinicID, *checkup.MedicalRecord.OwnerID)
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

	// ISSUE-004: Delete 前に checkup を取得して owner_id を確保。削除後は medical_record 経由で取れない。
	checkup, err := h.svc.Checkup.GetByID(c.Request.Context(), clinicID, medicalRecordID, checkupID)
	if err != nil {
		RespondError(c, err)
		return
	}
	var ownerID uint64
	if checkup.MedicalRecord != nil && checkup.MedicalRecord.OwnerID != nil {
		ownerID = *checkup.MedicalRecord.OwnerID
	}

	if err := h.svc.Checkup.Delete(c.Request.Context(), clinicID, medicalRecordID, checkupID); err != nil {
		RespondError(c, err)
		return
	}
	// ISSUE-004: 削除後の DB 状態から checkup_done_*/next_checkup_* を再構築（best-effort）。
	// 削除済みレコードのタグが再付与されない。同種別に他レコードが残れば最新日のタグを保持。
	if ownerID != 0 {
		_ = h.svc.LstepTagSync.ResyncOwnerCheckupTags(c.Request.Context(), clinicID, ownerID)
	}
	c.Status(http.StatusNoContent)
}

// ListGlobalCheckups は GET /v1/checkups — クリニック横断の健診記録一覧を返す
func (h *Handler) ListGlobalCheckups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	input := service.ListCheckupsByClinicInput{
		ClinicID: clinicID,
	}
	if v := c.Query("start_date"); v != "" {
		input.StartDate = &v
	}
	if v := c.Query("end_date"); v != "" {
		input.EndDate = &v
	}
	if v := c.Query("next_start_date"); v != "" {
		input.NextStartDate = &v
	}
	if v := c.Query("next_end_date"); v != "" {
		input.NextEndDate = &v
	}

	checkups, err := h.svc.Checkup.ListByClinic(c.Request.Context(), input)
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
	withinDays := 30
	if v := c.Query("within_days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("within_days must be integer"))
			return
		}
		withinDays = n
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
}

// RegisterCheckupRoutes は健診記録関連のルートを登録する
// RegisterCheckupRoutes はカルテ内の定期健診サブリソースルートを登録する。
// 子リソースは親（medical-records）の権限に従う（BUG-133: vitals/treatments 等と統一）。
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/checkups", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListCheckups)
	rg.POST("/:id/checkups", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateCheckup)
	rg.PATCH("/:id/checkups/:checkupId", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateCheckup)
	rg.DELETE("/:id/checkups/:checkupId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteCheckup)
}
