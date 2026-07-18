// Package handler provides HTTP handler implementations for InquiryTemplate entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ---- InquiryTemplate ----

// ListInquiryTemplates godoc
func (h *Handler) ListInquiryTemplates(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	templates, err := h.svc.InquiryTemplate.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(templates, toInquiryTemplateResponse))
}

// GetInquiryTemplate godoc
func (h *Handler) GetInquiryTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	tmpl, err := h.svc.InquiryTemplate.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInquiryTemplateResponse(tmpl))
}

// CreateInquiryTemplate godoc
func (h *Handler) CreateInquiryTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createInquiryTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	tmpl, err := h.svc.InquiryTemplate.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/inquiry-templates/%d", tmpl.ID))
	c.JSON(http.StatusCreated, toInquiryTemplateResponse(tmpl))
}

// UpdateInquiryTemplate godoc
func (h *Handler) UpdateInquiryTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateInquiryTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	updated, err := h.svc.InquiryTemplate.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInquiryTemplateResponse(updated))
}

// DeleteInquiryTemplate godoc
func (h *Handler) DeleteInquiryTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.InquiryTemplate.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderInquiryTemplates godoc
func (h *Handler) ReorderInquiryTemplates(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.InquiryTemplate.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
