// Package handler provides HTTP handler implementations for Vaccine entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createVaccineInput struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	Species     string   `json:"species"`
	Interval    string   `json:"interval"`
	SortOrder   int      `json:"sort_order"`
}

type updateVaccineInput struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	Species     string   `json:"species"`
	Interval    string   `json:"interval"`
	SortOrder   int      `json:"sort_order"`
}

// ---- Vaccine ----

// ListVaccines godoc
func (h *Handler) ListVaccines(c *gin.Context) {
	var species *string
	if s := c.Query("species"); s != "" {
		species = &s
	}
	vaccines, err := h.svc.Vaccine.List(c.Request.Context(), species)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, vaccines)
}

// CreateVaccine godoc
func (h *Handler) CreateVaccine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createVaccineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vaccine := &model.Vaccine{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Interval:    input.Interval,
		SortOrder:   input.SortOrder,
	}
	if input.Species != "" {
		s := model.VaccineSpecies(input.Species)
		vaccine.Species = &s
	}

	if err := h.svc.Vaccine.Create(c.Request.Context(), vaccine); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, vaccine)
}

// UpdateVaccine godoc
func (h *Handler) UpdateVaccine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateVaccineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vaccine := &model.Vaccine{
		ID:          id,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		Interval:    input.Interval,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		vaccine.IsActive = *input.IsActive
	}
	if input.Species != "" {
		s := model.VaccineSpecies(input.Species)
		vaccine.Species = &s
	}

	if err := h.svc.Vaccine.Update(c.Request.Context(), vaccine); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, vaccine)
}

// DeleteVaccine godoc
func (h *Handler) DeleteVaccine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Vaccine.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
