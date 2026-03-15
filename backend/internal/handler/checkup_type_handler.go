// Package handler provides HTTP handler implementations for CheckupType entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createCheckupTypeInput struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	Interval    string   `json:"interval"`
	TargetAge   string   `json:"target_age"`
	ParentID    *uint64  `json:"parent_id"`
	SortOrder   int      `json:"sort_order"`
}

type updateCheckupTypeInput struct {
	Name          string   `json:"name"`
	Price         *float64 `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   string   `json:"description"`
	Interval      string   `json:"interval"`
	TargetAge     string   `json:"target_age"`
	ParentID      *uint64  `json:"parent_id"`
	ClearParentID bool     `json:"clear_parent_id"`
	SortOrder     int      `json:"sort_order"`
}

// ---- CheckupType ----

// ListCheckupTypes godoc
func (h *Handler) ListCheckupTypes(c *gin.Context) {
	checkupTypes, err := h.svc.CheckupType.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, checkupTypes)
}

// CreateCheckupType godoc
func (h *Handler) CreateCheckupType(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createCheckupTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checkupType := &model.CheckupType{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Interval:    input.Interval,
		TargetAge:   input.TargetAge,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
	}

	if err := h.svc.CheckupType.Create(c.Request.Context(), checkupType); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, checkupType)
}

// UpdateCheckupType godoc
func (h *Handler) UpdateCheckupType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateCheckupTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checkupType := &model.CheckupType{
		ID:          id,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		Interval:    input.Interval,
		TargetAge:   input.TargetAge,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		checkupType.IsActive = *input.IsActive
	}
	if input.ClearParentID {
		checkupType.ParentID = nil
	} else if input.ParentID != nil {
		checkupType.ParentID = input.ParentID
	}

	if err := h.svc.CheckupType.Update(c.Request.Context(), checkupType); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, checkupType)
}

// ReorderCheckupTypes godoc
func (h *Handler) ReorderCheckupTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderCheckupTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.CheckupType.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// DeleteCheckupType godoc
func (h *Handler) DeleteCheckupType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.CheckupType.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
