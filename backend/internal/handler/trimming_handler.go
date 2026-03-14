package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

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
	c.JSON(http.StatusOK, toTrimmingResponse(trimming))
}

// CreateTrimming godoc
func (h *Handler) CreateTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createTrimmingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.CreateTrimmingInput{
		PetID:          req.PetID,
		StaffID:        req.StaffID,
		CourseID:       req.CourseID,
		Weight:         req.Weight,
		StyleRequest:   req.StyleRequest,
		BW:             req.BW,
		BT:             req.BT,
		UsedShampoo:    req.UsedShampoo,
		UsedRibbon:     req.UsedRibbon,
		Remarks:        req.Remarks,
		StyleImage:     req.StyleImage,
		CompletedImage: req.CompletedImage,
		OptionIDs:      req.OptionIDs,
	}
	if req.Date != nil {
		input.Date = *req.Date
	}
	if req.Status != "" {
		input.Status = model.TrimmingStatus(req.Status)
	}
	if req.BWUnit != "" {
		input.BWUnit = model.BodyWeightUnit(req.BWUnit)
	}

	record, err := h.svc.Trimming.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTrimmingResponse(record))
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
	var req updateTrimmingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.UpdateTrimmingInput{
		Date:           req.Date,
		PetID:          req.PetID,
		StaffID:        req.StaffID,
		CourseID:       req.CourseID,
		Weight:         req.Weight,
		StyleRequest:   req.StyleRequest,
		BW:             req.BW,
		BT:             req.BT,
		UsedShampoo:    req.UsedShampoo,
		UsedRibbon:     req.UsedRibbon,
		Remarks:        req.Remarks,
		StyleImage:     req.StyleImage,
		CompletedImage: req.CompletedImage,
		OptionIDs:      req.OptionIDs,
	}
	if req.Status != nil {
		status := model.TrimmingStatus(*req.Status)
		input.Status = &status
	}
	if req.BWUnit != nil {
		unit := model.BodyWeightUnit(*req.BWUnit)
		input.BWUnit = &unit
	}

	if err := h.svc.Trimming.Update(c.Request.Context(), clinicID, id, input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
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
