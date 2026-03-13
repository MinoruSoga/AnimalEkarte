// Package handler provides HTTP handler implementations for Medicine entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createMedicineInput struct {
	Name            string   `json:"name"             binding:"required"`
	Price           *float64 `json:"price"`
	IsActive        bool     `json:"is_active"`
	Description     string   `json:"description"`
	DosageForm      string   `json:"dosage_form"`
	MedicineUnit    string   `json:"medicine_unit"`
	InventoryID     *uint64  `json:"inventory_id"`
	DefaultQuantity int      `json:"default_quantity"`
	SortOrder       int      `json:"sort_order"`
}

type updateMedicineInput struct {
	Name            string   `json:"name"`
	Price           *float64 `json:"price"`
	IsActive        *bool    `json:"is_active"`
	Description     string   `json:"description"`
	DosageForm      string   `json:"dosage_form"`
	MedicineUnit    string   `json:"medicine_unit"`
	InventoryID     *uint64  `json:"inventory_id"`
	DefaultQuantity int      `json:"default_quantity"`
	SortOrder       int      `json:"sort_order"`
}

// ---- Medicine ----

// ListMedicines godoc
func (h *Handler) ListMedicines(c *gin.Context) {
	medicines, err := h.svc.Medicine.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, medicines)
}

// CreateMedicine godoc
func (h *Handler) CreateMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createMedicineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	medicine := &model.Medicine{
		ClinicID:        clinicID,
		Name:            input.Name,
		Price:           input.Price,
		IsActive:        input.IsActive,
		Description:     input.Description,
		InventoryID:     input.InventoryID,
		DefaultQuantity: input.DefaultQuantity,
		SortOrder:       input.SortOrder,
	}
	if input.DosageForm != "" {
		df := model.DosageForm(input.DosageForm)
		medicine.DosageForm = &df
	}
	if input.MedicineUnit != "" {
		mu := model.MedicineUnit(input.MedicineUnit)
		medicine.MedicineUnit = &mu
	}

	if err := h.svc.Medicine.Create(c.Request.Context(), medicine); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, medicine)
}

// UpdateMedicine godoc
func (h *Handler) UpdateMedicine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateMedicineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	medicine := &model.Medicine{
		ID:              id,
		Name:            input.Name,
		Price:           input.Price,
		Description:     input.Description,
		InventoryID:     input.InventoryID,
		DefaultQuantity: input.DefaultQuantity,
		SortOrder:       input.SortOrder,
	}
	if input.IsActive != nil {
		medicine.IsActive = *input.IsActive
	}
	if input.DosageForm != "" {
		df := model.DosageForm(input.DosageForm)
		medicine.DosageForm = &df
	}
	if input.MedicineUnit != "" {
		mu := model.MedicineUnit(input.MedicineUnit)
		medicine.MedicineUnit = &mu
	}

	if err := h.svc.Medicine.Update(c.Request.Context(), medicine); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, medicine)
}

// DeleteMedicine godoc
func (h *Handler) DeleteMedicine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Medicine.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
