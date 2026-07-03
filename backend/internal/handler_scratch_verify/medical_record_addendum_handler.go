package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
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

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
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

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req CreateMedicalRecordAddendumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	addendum, svcErr := h.svc.MedicalRecordAddendum.Create(c.Request.Context(), clinicID, req.toServiceInput(medicalRecordID, staffID))
	if svcErr != nil {
		RespondError(c, svcErr)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/addenda/%d", medicalRecordID, addendum.ID))
	c.JSON(http.StatusCreated, toMedicalRecordAddendumResponse(addendum))
}
