// Package handler provides HTTP handler implementations for Medicine entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// medicineListMaxLimit はマスタ全件取得の上限定数（他マスタと統一の全件返却）
const medicineListMaxLimit = 10000

// ListMedicines godoc
func (h *Handler) ListMedicines(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	// マスタ系は全件返却（他マスタと統一）。service.List のページネーション引数は固定値で全件取得。
	medicines, _, err := h.svc.Medicine.List(c.Request.Context(), clinicID, 1, medicineListMaxLimit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(medicines, toMedicineResponse))
}

// GetMedicine godoc
func (h *Handler) GetMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
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

	medicine, err := h.svc.Medicine.Create(c.Request.Context(), clinicID, &service.CreateMedicineInput{
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
		IsNonInsurance:  req.IsNonInsurance,
	})
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
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	medicine, err := h.svc.Medicine.Update(c.Request.Context(), clinicID, id, &service.UpdateMedicineInput{
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
		IsNonInsurance:  req.IsNonInsurance,
	})
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
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.Medicine.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteMedicine godoc
func (h *Handler) DeleteMedicine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Medicine.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
