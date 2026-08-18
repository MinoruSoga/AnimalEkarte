package medicalrecord

import (
	"fmt"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// CageHandler serves the cage master HTTP boundary. Moved from internal/handler (BE9-2D ⑥ Batch C).
type CageHandler struct {
	service CageService
}

// NewCageHandler initializes a CageHandler.
func NewCageHandler(service CageService) *CageHandler {
	return &CageHandler{service: service}
}

// ---- Cage ----

// ListCages godoc
func (h *CageHandler) ListCages(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	cageType := newListCageQuery(c.Request.URL.Query()).toServiceFilter()
	cages, err := h.service.List(c.Request.Context(), clinicID, cageType)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(cages, toCageResponse))
}

// GetCage godoc
func (h *CageHandler) GetCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	cage, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCageResponse(cage))
}

// CreateCage godoc
func (h *CageHandler) CreateCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createCageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	cage, err := h.service.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/cages/%d", cage.ID))
	c.JSON(http.StatusCreated, toCageResponse(cage))
}

// UpdateCage godoc
func (h *CageHandler) UpdateCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateCageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	cage, err := h.service.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toCageResponse(cage))
}

// ReorderCages godoc
func (h *CageHandler) ReorderCages(c *gin.Context) {
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

// DeleteCage godoc
func (h *CageHandler) DeleteCage(c *gin.Context) {
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
