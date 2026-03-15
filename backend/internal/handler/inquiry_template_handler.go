// Package handler provides HTTP handler implementations for InquiryTemplate entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
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
	c.JSON(http.StatusOK, toInquiryTemplateResponseList(templates))
}

// GetInquiryTemplate godoc
func (h *Handler) GetInquiryTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	tmpl, err := h.svc.InquiryTemplate.GetByID(c.Request.Context(), id)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	tmpl := &model.InquiryTemplate{
		ClinicID:  clinicID,
		Category:  req.Category,
		Title:     req.Title,
		Content:   req.Content,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}

	if err := h.svc.InquiryTemplate.Create(c.Request.Context(), tmpl); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toInquiryTemplateResponse(tmpl))
}

// UpdateInquiryTemplate godoc
func (h *Handler) UpdateInquiryTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req updateInquiryTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.UpdateInquiryTemplateInput{
		Category:  req.Category,
		Title:     req.Title,
		Content:   req.Content,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}

	updated, err := h.svc.InquiryTemplate.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInquiryTemplateResponse(updated))
}

// DeleteInquiryTemplate godoc
func (h *Handler) DeleteInquiryTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.InquiryTemplate.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
