// Package handler provides HTTP handler implementations for ChiefComplaintCategory entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- ChiefComplaintCategory ----

// GetChiefComplaint godoc
func (h *Handler) GetChiefComplaint(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	category, err := h.svc.ChiefComplaintCategory.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChiefComplaintResponse(category))
}

// ListChiefComplaints godoc
func (h *Handler) ListChiefComplaints(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	categories, err := h.svc.ChiefComplaintCategory.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChiefComplaintResponseList(categories))
}

// CreateChiefComplaint godoc
func (h *Handler) CreateChiefComplaint(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createChiefComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	category := &model.ChiefComplaintCategory{
		ClinicID:  clinicID,
		Name:      req.Name,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}

	if err := h.svc.ChiefComplaintCategory.Create(c.Request.Context(), category); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toChiefComplaintResponse(category))
}

// UpdateChiefComplaint godoc
func (h *Handler) UpdateChiefComplaint(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req updateChiefComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.UpdateChiefComplaintCategoryInput{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}

	updated, err := h.svc.ChiefComplaintCategory.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChiefComplaintResponse(updated))
}

// DeleteChiefComplaint godoc
func (h *Handler) DeleteChiefComplaint(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.ChiefComplaintCategory.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
