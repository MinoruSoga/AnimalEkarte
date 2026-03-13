// Package handler provides HTTP handler implementations for TrimmingCourse and TrimmingOption entities.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createTrimmingCourseInput struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	TargetSize  string   `json:"target_size"`
	Duration    string   `json:"duration"`
	SortOrder   int      `json:"sort_order"`
}

type updateTrimmingCourseInput struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	TargetSize  string   `json:"target_size"`
	Duration    string   `json:"duration"`
	SortOrder   int      `json:"sort_order"`
}

type createTrimmingOptionInput struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	Duration    string   `json:"duration"`
	Combinable  bool     `json:"combinable"`
	SortOrder   int      `json:"sort_order"`
}

type updateTrimmingOptionInput struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	Duration    string   `json:"duration"`
	Combinable  *bool    `json:"combinable"`
	SortOrder   int      `json:"sort_order"`
}

// ---- TrimmingCourse ----

// ListTrimmingCourses godoc
func (h *Handler) ListTrimmingCourses(c *gin.Context) {
	courses, err := h.svc.TrimmingCourse.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, courses)
}

// CreateTrimmingCourse godoc
func (h *Handler) CreateTrimmingCourse(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createTrimmingCourseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	course := &model.TrimmingCourse{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Duration:    input.Duration,
		SortOrder:   input.SortOrder,
	}
	if input.TargetSize != "" {
		ts := model.TargetSize(input.TargetSize)
		course.TargetSize = &ts
	}

	if err := h.svc.TrimmingCourse.Create(c.Request.Context(), course); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, course)
}

// UpdateTrimmingCourse godoc
func (h *Handler) UpdateTrimmingCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateTrimmingCourseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	course := &model.TrimmingCourse{
		ID:          id,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		Duration:    input.Duration,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		course.IsActive = *input.IsActive
	}
	if input.TargetSize != "" {
		ts := model.TargetSize(input.TargetSize)
		course.TargetSize = &ts
	}

	if err := h.svc.TrimmingCourse.Update(c.Request.Context(), course); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, course)
}

// DeleteTrimmingCourse godoc
func (h *Handler) DeleteTrimmingCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.TrimmingCourse.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- TrimmingOption ----

// ListTrimmingOptions godoc
func (h *Handler) ListTrimmingOptions(c *gin.Context) {
	options, err := h.svc.TrimmingOption.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, options)
}

// CreateTrimmingOption godoc
func (h *Handler) CreateTrimmingOption(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createTrimmingOptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	option := &model.TrimmingOption{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Duration:    input.Duration,
		Combinable:  input.Combinable,
		SortOrder:   input.SortOrder,
	}

	if err := h.svc.TrimmingOption.Create(c.Request.Context(), option); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, option)
}

// UpdateTrimmingOption godoc
func (h *Handler) UpdateTrimmingOption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateTrimmingOptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	option := &model.TrimmingOption{
		ID:          id,
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		Duration:    input.Duration,
		SortOrder:   input.SortOrder,
	}
	if input.IsActive != nil {
		option.IsActive = *input.IsActive
	}
	if input.Combinable != nil {
		option.Combinable = *input.Combinable
	}

	if err := h.svc.TrimmingOption.Update(c.Request.Context(), option); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, option)
}

// DeleteTrimmingOption godoc
func (h *Handler) DeleteTrimmingOption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.TrimmingOption.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
