package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// InquiryTemplateHandler serves the InquiryTemplate master HTTP boundary.
type InquiryTemplateHandler struct {
	service InquiryTemplateService
}

// NewInquiryTemplateHandler initializes an InquiryTemplateHandler.
func NewInquiryTemplateHandler(service InquiryTemplateService) *InquiryTemplateHandler {
	return &InquiryTemplateHandler{service: service}
}

// ListInquiryTemplates godoc
func (h *InquiryTemplateHandler) ListInquiryTemplates(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	templates, err := h.service.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(templates, toInquiryTemplateResponse))
}

// GetInquiryTemplate godoc
func (h *InquiryTemplateHandler) GetInquiryTemplate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	tmpl, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInquiryTemplateResponse(tmpl))
}

// CreateInquiryTemplate godoc
func (h *InquiryTemplateHandler) CreateInquiryTemplate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createInquiryTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	tmpl, err := h.service.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/inquiry-templates/%d", tmpl.ID))
	c.JSON(http.StatusCreated, toInquiryTemplateResponse(tmpl))
}

// UpdateInquiryTemplate godoc
func (h *InquiryTemplateHandler) UpdateInquiryTemplate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateInquiryTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	updated, err := h.service.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInquiryTemplateResponse(updated))
}

// DeleteInquiryTemplate godoc
func (h *InquiryTemplateHandler) DeleteInquiryTemplate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderInquiryTemplates godoc
func (h *InquiryTemplateHandler) ReorderInquiryTemplates(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.service.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
