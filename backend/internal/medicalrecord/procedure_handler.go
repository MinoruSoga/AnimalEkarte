// Package handler provides HTTP handler implementations for Procedure entity.
package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ProcedureHandler serves the procedure master HTTP boundary. Moved from internal/handler (BE9-2D ⑥ Batch C).
type ProcedureHandler struct {
	service ProcedureService
}

// NewProcedureHandler initializes a ProcedureHandler.
func NewProcedureHandler(service ProcedureService) *ProcedureHandler {
	return &ProcedureHandler{service: service}
}

// ---- Procedure ----

// ListProcedures godoc
func (h *ProcedureHandler) ListProcedures(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	procedures, err := h.service.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(procedures, toProcedureResponse))
}

// GetProcedure godoc
func (h *ProcedureHandler) GetProcedure(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	procedure, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProcedureResponse(procedure))
}

// CreateProcedure godoc
func (h *ProcedureHandler) CreateProcedure(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createProcedureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	procedure, err := h.service.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/procedures/%d", procedure.ID))
	c.JSON(http.StatusCreated, toProcedureResponse(procedure))
}

// UpdateProcedure godoc
func (h *ProcedureHandler) UpdateProcedure(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateProcedureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	procedure, err := h.service.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toProcedureResponse(procedure))
}

// ReorderProcedures godoc
func (h *ProcedureHandler) ReorderProcedures(c *gin.Context) {
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

// DeleteProcedure godoc
func (h *ProcedureHandler) DeleteProcedure(c *gin.Context) {
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
