package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

func (h *Handler) RegisterMedicalRecordAddendumRoutes(records *gin.RouterGroup) {
	records.GET("/:id/addenda", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListMedicalRecordAddenda)
	records.POST("/:id/addenda", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.CreateMedicalRecordAddendum)
}

func (h *Handler) ListMedicalRecordAddenda(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}

	addenda, svcErr := h.svc.MedicalRecordAddendum.FindByMedicalRecordID(c.Request.Context(), clinicID, medicalRecordID)
	if svcErr != nil {
		RespondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, toMedicalRecordAddendumListResponse(addenda))
}

func (h *Handler) CreateMedicalRecordAddendum(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}

	var req CreateMedicalRecordAddendumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := service.CreateMedicalRecordAddendumInput{
		MedicalRecordID: medicalRecordID,
		AuthorUserID:    staffID,
		AfterText:       req.AfterText,
		Reason:          req.Reason,
	}

	addendum, svcErr := h.svc.MedicalRecordAddendum.Create(c.Request.Context(), clinicID, input)
	if svcErr != nil {
		RespondError(c, svcErr)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/addenda/%d", medicalRecordID, addendum.ID))
	c.JSON(http.StatusCreated, toMedicalRecordAddendumResponse(addendum))
}
