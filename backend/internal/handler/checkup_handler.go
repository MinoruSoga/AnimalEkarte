package handler

import (
	"net/http"
	"strconv"

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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical_record id"))
		return
	}

	var req createCheckupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	checkup, err := h.svc.Checkup.Create(c.Request.Context(), id, &service.CreateCheckupInput{
		CheckupTypeID: req.CheckupTypeID,
		PetID:         req.PetID,
		Date:          req.Date,
		NextDate:      req.NextDate,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	checkup, err := h.svc.Checkup.Update(c.Request.Context(), medicalRecordID, checkupID, &service.UpdateCheckupInput{
		CheckupTypeID: req.CheckupTypeID,
		PetID:         req.PetID,
		Date:          req.Date,
		NextDate:      req.NextDate,
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

// RegisterCheckupRoutes は健診記録関連のルートを登録する
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/checkups", h.ListCheckups)
	rg.POST("/:id/checkups", h.CreateCheckup)
	rg.PATCH("/:id/checkups/:checkupId", h.UpdateCheckup)
	rg.DELETE("/:id/checkups/:checkupId", h.DeleteCheckup)
}
