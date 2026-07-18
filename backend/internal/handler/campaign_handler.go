package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ListCampaigns godoc
// GET /v1/masters/campaigns
func (h *Handler) ListCampaigns(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ms, err := h.svc.Campaign.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(ms, toCampaignResponse))
}

// GetCampaign godoc
// GET /v1/masters/campaigns/:id
func (h *Handler) GetCampaign(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.Campaign.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCampaignResponse(m))
}

// CreateCampaign godoc
// POST /v1/masters/campaigns
func (h *Handler) CreateCampaign(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, err)
		return
	}
	m, err := h.svc.Campaign.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/campaigns/%d", m.ID))
	c.JSON(http.StatusCreated, toCampaignResponse(m))
}

// UpdateCampaign godoc
// PATCH /v1/masters/campaigns/:id
func (h *Handler) UpdateCampaign(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, err)
		return
	}
	m, err := h.svc.Campaign.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCampaignResponse(m))
}

// ReorderCampaigns godoc
// PATCH /v1/masters/campaigns/reorder
func (h *Handler) ReorderCampaigns(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.Campaign.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteCampaign godoc
// DELETE /v1/masters/campaigns/:id
func (h *Handler) DeleteCampaign(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Campaign.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
