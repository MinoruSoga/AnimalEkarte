// Package handler provides HTTP handler implementations for Medicine entity.
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/service"
)

// ListMedicines godoc
func (h *Handler) ListMedicines(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicines, err := h.svc.Medicine.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, medicines)
}

// GetMedicine godoc
func (h *Handler) GetMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	medicine, err := h.svc.Medicine.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, medicine)
}

// CreateMedicine godoc
func (h *Handler) CreateMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.CreateMedicineInput{
		Name:            req.Name,
		Price:           req.Price,
		IsActive:        req.IsActive,
		Description:     req.Description,
		DosageForm:      req.DosageForm,
		MedicineUnit:    req.MedicineUnit,
		InventoryID:     req.InventoryID,
		DefaultQuantity: req.DefaultQuantity,
		SortOrder:       req.SortOrder,
	}

	medicine, err := h.svc.Medicine.Create(c.Request.Context(), clinicID, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/medicines/%d", medicine.ID))
	c.JSON(http.StatusCreated, medicine)
}

// UpdateMedicine godoc
func (h *Handler) UpdateMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.UpdateMedicineInput{
		Name:            req.Name,
		Price:           req.Price,
		IsActive:        req.IsActive,
		Description:     req.Description,
		DosageForm:      req.DosageForm,
		MedicineUnit:    req.MedicineUnit,
		InventoryID:     req.InventoryID,
		DefaultQuantity: req.DefaultQuantity,
		SortOrder:       req.SortOrder,
	}

	medicine, err := h.svc.Medicine.Update(c.Request.Context(), clinicID, id, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, medicine)
}

// DeleteMedicine godoc
func (h *Handler) DeleteMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Medicine.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
