// Package handler provides HTTP handler implementations for Medicine entity.
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListMedicines godoc
func (h *Handler) ListMedicines(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	medicines, total, err := h.svc.Medicine.List(c.Request.Context(), clinicID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toMedicineResponseList(medicines), total, page, limit))
}

// GetMedicine godoc
func (h *Handler) GetMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	medicine, err := h.svc.Medicine.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineResponse(medicine))
}

// CreateMedicine godoc
func (h *Handler) CreateMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := service.CreateMedicineInput{
		Name:            req.Name,
		ParentID:        req.ParentID,
		Price:           req.Price,
		IsActive:        req.IsActive,
		Description:     req.Description,
		DosageForm:      req.DosageForm,
		MedicineUnit:    req.MedicineUnit,
		InventoryID:     req.InventoryID,
		DefaultQuantity: req.DefaultQuantity,
		SortOrder:       req.SortOrder,
		TaxType:         req.TaxType,
		TaxRate:         req.TaxRate,
	}

	medicine, err := h.svc.Medicine.Create(c.Request.Context(), clinicID, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/medicines/%d", medicine.ID))
	c.JSON(http.StatusCreated, toMedicineResponse(medicine))
}

// UpdateMedicine godoc
func (h *Handler) UpdateMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req updateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := service.UpdateMedicineInput{
		Name:            req.Name,
		ParentID:        req.ParentID,
		ClearParentID:   req.ClearParentID,
		Price:           req.Price,
		IsActive:        req.IsActive,
		Description:     req.Description,
		DosageForm:      req.DosageForm,
		MedicineUnit:    req.MedicineUnit,
		InventoryID:     req.InventoryID,
		DefaultQuantity: req.DefaultQuantity,
		SortOrder:       req.SortOrder,
		TaxType:         req.TaxType,
		TaxRate:         req.TaxRate,
	}

	medicine, err := h.svc.Medicine.Update(c.Request.Context(), clinicID, id, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineResponse(medicine))
}

// ReorderMedicines godoc
func (h *Handler) ReorderMedicines(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.Medicine.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// DeleteMedicine godoc
func (h *Handler) DeleteMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Medicine.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
