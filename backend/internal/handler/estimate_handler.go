package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListEstimates godoc
func (h *Handler) ListEstimates(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid owner_id"))
			return
		}
		ownerID = &id
	}

	var medicalRecordID *uint64
	if s := c.Query("medical_record_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid medical_record_id"))
			return
		}
		medicalRecordID = &id
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	estimates, total, err := h.svc.Estimate.List(c.Request.Context(), clinicID, ownerID, medicalRecordID, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toEstimateResponseList(estimates), total, page, limit))
}

// GetEstimate godoc
func (h *Handler) GetEstimate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	estimate, err := h.svc.Estimate.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEstimateResponse(estimate))
}

// CreateEstimate godoc
func (h *Handler) CreateEstimate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.CreateEstimateInput{
		MedicalRecordID: req.MedicalRecordID,
		Title:           req.Title,
		OwnerID:         req.OwnerID,
		Subtotal:        req.Subtotal,
		TaxTotal:        req.TaxTotal,
		TotalAmount:     req.TotalAmount,
		InsuranceAmount: req.InsuranceAmount,
		DiscountAmount:  req.DiscountAmount,
		ValidUntil:      req.ValidUntil,
		Comment:         req.Comment,
		Notes:           req.Notes,
		CreatedBy:       req.CreatedBy,
	}
	if req.Status != "" {
		input.Status = model.EstimateStatus(req.Status)
	}

	ctx := c.Request.Context()
	estimate, err := h.svc.Estimate.Create(ctx, clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toEstimateResponse(estimate))
}

// UpdateEstimate godoc
func (h *Handler) UpdateEstimate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req updateEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.UpdateEstimateInput{
		Title:           req.Title,
		Subtotal:        req.Subtotal,
		TaxTotal:        req.TaxTotal,
		TotalAmount:     req.TotalAmount,
		InsuranceAmount: req.InsuranceAmount,
		DiscountAmount:  req.DiscountAmount,
		ValidUntil:      req.ValidUntil,
		ClearValidUntil: req.ClearValidUntil,
		Comment:         req.Comment,
		Notes:           req.Notes,
	}
	if req.Status != nil {
		s := model.EstimateStatus(*req.Status)
		input.Status = &s
	}

	ctx := c.Request.Context()
	estimate, err := h.svc.Estimate.Update(ctx, clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEstimateResponse(estimate))
}

// DeleteEstimate godoc
func (h *Handler) DeleteEstimate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Estimate.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

