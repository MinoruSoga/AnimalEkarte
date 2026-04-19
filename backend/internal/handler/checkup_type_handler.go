// Package handler provides HTTP handler implementations for CheckupType entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- CheckupType ----

// GetCheckupType godoc
func (h *Handler) GetCheckupType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	checkupType, err := h.svc.CheckupType.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupTypeResponse(checkupType))
}

// ListCheckupTypes godoc
func (h *Handler) ListCheckupTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	checkupTypes, err := h.svc.CheckupType.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(checkupTypes, toCheckupTypeResponse))
}

// CreateCheckupType godoc
func (h *Handler) CreateCheckupType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createCheckupTypeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := &service.CreateCheckupTypeInput{
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Interval:    input.Interval,
		TargetAge:   input.TargetAge,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
	}

	checkupType, err := h.svc.CheckupType.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toCheckupTypeResponse(checkupType))
}

// UpdateCheckupType godoc
func (h *Handler) UpdateCheckupType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateCheckupTypeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateCheckupTypeInput{
		Name:          input.Name,
		Price:         input.Price,
		IsActive:      input.IsActive,
		Description:   input.Description,
		Interval:      input.Interval,
		TargetAge:     input.TargetAge,
		ParentID:      input.ParentID,
		ClearParentID: input.ClearParentID,
		SortOrder:     input.SortOrder,
	}

	checkupType, err := h.svc.CheckupType.Update(c.Request.Context(), clinicID, id, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupTypeResponse(checkupType))
}

// ReorderCheckupTypes godoc
func (h *Handler) ReorderCheckupTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.CheckupType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteCheckupType godoc
func (h *Handler) DeleteCheckupType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.CheckupType.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
