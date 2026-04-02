package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListCheckups は指定カルテに紐づく健診記録の一覧を返す
func (h *Handler) ListCheckups(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical_record id"))
		return
	}

	checkups, err := h.svc.Checkup.List(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupResponseList(checkups))
}

// CreateCheckup は指定カルテに健診記録を作成する
func (h *Handler) CreateCheckup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical_record id"))
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
	c.JSON(http.StatusCreated, toCheckupResponse(checkup))
}

// UpdateCheckup は健診記録を部分更新する
func (h *Handler) UpdateCheckup(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical_record id"))
		return
	}

	checkupID, err := strconv.ParseUint(c.Param("checkupId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid checkup id"))
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

	checkup, err := h.svc.Checkup.Update(c.Request.Context(), medicalRecordID, checkupID, &service.UpdateCheckupInput{
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
	c.JSON(http.StatusOK, toCheckupResponse(checkup))
}

// DeleteCheckup は健診記録を soft delete する
func (h *Handler) DeleteCheckup(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical_record id"))
		return
	}

	checkupID, err := strconv.ParseUint(c.Param("checkupId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid checkup id"))
		return
	}

	if err := h.svc.Checkup.Delete(c.Request.Context(), medicalRecordID, checkupID); err != nil {
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
	c.JSON(http.StatusOK, gin.H{"data": toCheckupGlobalResponseList(checkups)})
}

// RegisterGlobalCheckupRoutes は /checkups トップレベルルートを登録する
func (h *Handler) RegisterGlobalCheckupRoutes(rg *gin.RouterGroup) {
	checkups := rg.Group("/checkups")
	checkups.GET("", h.ListGlobalCheckups)
}

// RegisterCheckupRoutes は健診記録関連のルートを登録する
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/checkups", h.ListCheckups)
	rg.POST("/:id/checkups", h.CreateCheckup)
	rg.PATCH("/:id/checkups/:checkupId", h.UpdateCheckup)
	rg.DELETE("/:id/checkups/:checkupId", h.DeleteCheckup)
}
