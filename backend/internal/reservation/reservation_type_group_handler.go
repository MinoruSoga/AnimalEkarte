package reservation

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ReservationTypeGroupHandler は予約区分グループの HTTP handler。
type ReservationTypeGroupHandler struct {
	svc ReservationTypeGroupService
}

// NewReservationTypeGroupHandler は ReservationTypeGroupHandler を構築する。
func NewReservationTypeGroupHandler(svc ReservationTypeGroupService) *ReservationTypeGroupHandler {
	return &ReservationTypeGroupHandler{svc: svc}
}

// ListReservationTypeGroups godoc
func (h *ReservationTypeGroupHandler) ListReservationTypeGroups(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	groups, err := h.svc.List(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(groups, toReservationTypeGroupResponse))
}

// GetReservationTypeGroup godoc
func (h *ReservationTypeGroupHandler) GetReservationTypeGroup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	g, err := h.svc.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeGroupResponse(g))
}

// CreateReservationTypeGroup godoc
func (h *ReservationTypeGroupHandler) CreateReservationTypeGroup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createReservationTypeGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	g, err := h.svc.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/reservation-type-groups/%d", g.ID))
	c.JSON(http.StatusCreated, toReservationTypeGroupResponse(g))
}

// UpdateReservationTypeGroup godoc
func (h *ReservationTypeGroupHandler) UpdateReservationTypeGroup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateReservationTypeGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	g, err := h.svc.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationTypeGroupResponse(g))
}

// DeleteReservationTypeGroup godoc
func (h *ReservationTypeGroupHandler) DeleteReservationTypeGroup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), clinicID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderReservationTypeGroups godoc
func (h *ReservationTypeGroupHandler) ReorderReservationTypeGroups(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.svc.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
