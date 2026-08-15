package billing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// CampaignHandler は CampaignService の HTTP handler。
type CampaignHandler struct {
	svc CampaignService
}

// NewCampaignHandler は CampaignHandler を構築する。
func NewCampaignHandler(svc CampaignService) *CampaignHandler {
	return &CampaignHandler{svc: svc}
}

// ListCampaigns godoc
// GET /v1/masters/campaigns
func (h *CampaignHandler) ListCampaigns(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ms, err := h.svc.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(ms, toCampaignResponse))
}

// GetCampaign godoc
// GET /v1/masters/campaigns/:id
func (h *CampaignHandler) GetCampaign(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCampaignResponse(m))
}

// CreateCampaign godoc
// POST /v1/masters/campaigns
func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	m, err := h.svc.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/campaigns/%d", m.ID))
	c.JSON(http.StatusCreated, toCampaignResponse(m))
}

// UpdateCampaign godoc
// PATCH /v1/masters/campaigns/:id
func (h *CampaignHandler) UpdateCampaign(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	m, err := h.svc.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCampaignResponse(m))
}

// ReorderCampaigns godoc
// PATCH /v1/masters/campaigns/reorder
func (h *CampaignHandler) ReorderCampaigns(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.svc.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteCampaign godoc
// DELETE /v1/masters/campaigns/:id
func (h *CampaignHandler) DeleteCampaign(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
