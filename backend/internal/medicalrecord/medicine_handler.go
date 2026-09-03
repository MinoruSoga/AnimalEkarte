// Package handler provides HTTP handler implementations for Medicine entity.
package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// MedicineHandler serves the medicine master HTTP boundary. Moved from internal/handler (BE9-2D ⑥ Batch C).
type MedicineHandler struct {
	service MedicineService
}

// NewMedicineHandler initializes a MedicineHandler.
func NewMedicineHandler(service MedicineService) *MedicineHandler {
	return &MedicineHandler{service: service}
}

// medicineListMaxLimit はマスタ全件取得の上限定数（他マスタと統一の全件返却）
const medicineListMaxLimit = 10000

// ListMedicines godoc
func (h *MedicineHandler) ListMedicines(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "view") {
		return
	}

	// マスタ系は全件返却（他マスタと統一）。service.List のページネーション引数は固定値で全件取得。
	medicines, _, err := h.service.List(c.Request.Context(), clinicID, 1, medicineListMaxLimit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(medicines, toMedicineResponse))
}

// GetMedicine godoc
func (h *MedicineHandler) GetMedicine(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	medicine, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineResponse(medicine))
}

// CreateMedicine godoc
func (h *MedicineHandler) CreateMedicine(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	createInput := req.toServiceInput()
	createInput.ActorID = httpapi.OptionalStaffID(c) // #201 B-2: per_weight 有効化 audit の実施者
	medicine, err := h.service.Create(c.Request.Context(), clinicID, createInput)
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/medicines/%d", medicine.ID))
	c.JSON(http.StatusCreated, toMedicineResponse(medicine))
}

// UpdateMedicine godoc
func (h *MedicineHandler) UpdateMedicine(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	updateInput := req.toServiceInput()
	updateInput.ActorID = httpapi.OptionalStaffID(c) // #201 B-2: per_weight 有効化 audit の実施者
	medicine, err := h.service.Update(c.Request.Context(), clinicID, id, updateInput)
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineResponse(medicine))
}

// ReorderMedicines godoc
func (h *MedicineHandler) ReorderMedicines(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.service.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteMedicine godoc
func (h *MedicineHandler) DeleteMedicine(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
