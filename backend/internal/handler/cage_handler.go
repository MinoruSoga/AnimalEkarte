// Package handler provides HTTP handler implementations for Cage entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createCageInput struct {
	Name        string   `json:"name"        binding:"required"`
	CageType    string   `json:"cage_type"   binding:"required"`
	CageSize    string   `json:"cage_size"   binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
}

type updateCageInput struct {
	Name        string   `json:"name"`
	CageType    string   `json:"cage_type"`
	CageSize    string   `json:"cage_size"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
}

// ---- Cage ----

// ListCages godoc
func (h *Handler) ListCages(c *gin.Context) {
	var cageType *string
	if t := c.Query("cage_type"); t != "" {
		cageType = &t
	}
	cages, err := h.svc.Cage.List(c.Request.Context(), cageType)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, cages)
}

// CreateCage godoc
func (h *Handler) CreateCage(c *gin.Context) {
	var input createCageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cage := &model.Cage{
		Name:        input.Name,
		CageType:    model.CageType(input.CageType),
		CageSize:    model.CageSize(input.CageSize),
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}

	if err := h.svc.Cage.Create(c.Request.Context(), cage); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, cage)
}

// UpdateCage godoc
func (h *Handler) UpdateCage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateCageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cage := &model.Cage{
		ID:          id,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if input.CageType != "" {
		cage.CageType = model.CageType(input.CageType)
	}
	if input.CageSize != "" {
		cage.CageSize = model.CageSize(input.CageSize)
	}
	if input.IsActive != nil {
		cage.IsActive = *input.IsActive
	}

	if err := h.svc.Cage.Update(c.Request.Context(), cage); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, cage)
}

// DeleteCage godoc
func (h *Handler) DeleteCage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Cage.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
