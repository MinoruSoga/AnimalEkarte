package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createTrimmingInput struct {
	Date           time.Time `json:"date"            binding:"required"`
	PetID          *uint64   `json:"pet_id"`
	StaffID        uint64    `json:"staff_id"        binding:"required"`
	CourseID       uint64    `json:"course_id"       binding:"required"`
	Weight         string    `json:"weight"`
	Status         string    `json:"status"`
	StyleRequest   string    `json:"style_request"`
	BW             string    `json:"bw"`
	BWUnit         string    `json:"bw_unit"`
	BT             string    `json:"bt"`
	UsedShampoo    string    `json:"used_shampoo"`
	UsedRibbon     string    `json:"used_ribbon"`
	Remarks        string    `json:"remarks"`
	StyleImage     string    `json:"style_image"`
	CompletedImage string    `json:"completed_image"`
}

type updateTrimmingInput struct {
	Date           *time.Time `json:"date"`
	PetID          *uint64    `json:"pet_id"`
	StaffID        uint64     `json:"staff_id"`
	CourseID       uint64     `json:"course_id"`
	Weight         string     `json:"weight"`
	Status         string     `json:"status"`
	StyleRequest   string     `json:"style_request"`
	BW             string     `json:"bw"`
	BWUnit         string     `json:"bw_unit"`
	BT             string     `json:"bt"`
	UsedShampoo    string     `json:"used_shampoo"`
	UsedRibbon     string     `json:"used_ribbon"`
	Remarks        string     `json:"remarks"`
	StyleImage     string     `json:"style_image"`
	CompletedImage string     `json:"completed_image"`
}

// ListTrimmings godoc
func (h *Handler) ListTrimmings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var petID *uint64
	if petIDStr := c.Query("pet_id"); petIDStr != "" {
		id, err := strconv.ParseUint(petIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id"})
			return
		}
		petID = &id
	}

	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
		ownerID = &id
	}

	trimmings, total, err := h.svc.Trimming.List(c.Request.Context(), clinicID, petID, ownerID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(trimmings, total, page, limit))
}

// GetTrimming godoc
func (h *Handler) GetTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	trimming, err := h.svc.Trimming.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, trimming)
}

// CreateTrimming godoc
func (h *Handler) CreateTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createTrimmingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trimming := &model.TrimmingRecord{
		ClinicID:       clinicID,
		Date:           input.Date,
		PetID:          input.PetID,
		StaffID:        input.StaffID,
		CourseID:       input.CourseID,
		Weight:         input.Weight,
		StyleRequest:   input.StyleRequest,
		BW:             input.BW,
		BT:             input.BT,
		UsedShampoo:    input.UsedShampoo,
		UsedRibbon:     input.UsedRibbon,
		Remarks:        input.Remarks,
		StyleImage:     input.StyleImage,
		CompletedImage: input.CompletedImage,
	}
	if input.Status != "" {
		trimming.Status = model.TrimmingStatus(input.Status)
	}
	if input.BWUnit != "" {
		trimming.BWUnit = model.BodyWeightUnit(input.BWUnit)
	}

	if err := h.svc.Trimming.Create(c.Request.Context(), clinicID, trimming); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, trimming)
}

// UpdateTrimming godoc
func (h *Handler) UpdateTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateTrimmingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trimming := &model.TrimmingRecord{
		ID:             id,
		ClinicID:       clinicID,
		PetID:          input.PetID,
		StaffID:        input.StaffID,
		CourseID:       input.CourseID,
		Weight:         input.Weight,
		StyleRequest:   input.StyleRequest,
		BW:             input.BW,
		BT:             input.BT,
		UsedShampoo:    input.UsedShampoo,
		UsedRibbon:     input.UsedRibbon,
		Remarks:        input.Remarks,
		StyleImage:     input.StyleImage,
		CompletedImage: input.CompletedImage,
	}
	if input.Date != nil {
		trimming.Date = *input.Date
	}
	if input.Status != "" {
		trimming.Status = model.TrimmingStatus(input.Status)
	}
	if input.BWUnit != "" {
		trimming.BWUnit = model.BodyWeightUnit(input.BWUnit)
	}

	if err := h.svc.Trimming.Update(c.Request.Context(), clinicID, trimming); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, trimming)
}

func (h *Handler) DeleteTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Trimming.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterTrimmingRoutes はトリミング関連のルートを登録する
func (h *Handler) RegisterTrimmingRoutes(rg *gin.RouterGroup) {
	trimmings := rg.Group("/trimmings")
	trimmings.GET("", h.ListTrimmings)
	trimmings.POST("", h.CreateTrimming)
	trimmings.GET("/:id", h.GetTrimming)
	trimmings.PATCH("/:id", h.UpdateTrimming)
	trimmings.DELETE("/:id", h.DeleteTrimming)
}
