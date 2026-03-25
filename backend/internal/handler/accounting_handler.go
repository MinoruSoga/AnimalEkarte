package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListAccountings godoc
func (h *Handler) ListAccountings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var petID *uint64
	if s := c.Query("pet_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid pet_id"))
			return
		}
		petID = &id
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

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	startDate, err := parseDateQuery(c, "start_date")
	if err != nil {
		RespondError(c, err)
		return
	}
	endDate, err := parseDateQuery(c, "end_date")
	if err != nil {
		RespondError(c, err)
		return
	}

	accountings, total, err := h.svc.Accounting.List(c.Request.Context(), clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toAccountingResponseList(accountings), total, page, limit))
}

// GetAccounting godoc
func (h *Handler) GetAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	accounting, err := h.svc.Accounting.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountingResponse(accounting))
}

// CreateAccounting godoc
func (h *Handler) CreateAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createAccountingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	createInput := service.CreateAccountingInput{
		ClinicID:          clinicID,
		MedicalRecordID:   input.MedicalRecordID,
		HospitalizationID: input.HospitalizationID,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		Subtotal:          input.Subtotal,
		TaxTotal:          input.TaxTotal,
		TotalAmount:       input.TotalAmount,
		HasInsurance:      input.HasInsurance,
		ScheduledDate:     input.ScheduledDate,
		CompletedAt:       input.CompletedAt,
		Memo:              input.Memo,
	}
	if input.Status != "" {
		createInput.Status = model.BillingStatus(input.Status)
	}

	ctx := c.Request.Context()
	created, err := h.svc.Accounting.Create(ctx, createInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toAccountingResponse(created))
}

// UpdateAccounting godoc
func (h *Handler) UpdateAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var input updateAccountingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	updateInput := service.UpdateAccountingInput{
		ID:                id,
		ClinicID:          clinicID,
		MedicalRecordID:   input.MedicalRecordID,
		HospitalizationID: input.HospitalizationID,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		Subtotal:          input.Subtotal,
		TaxTotal:          input.TaxTotal,
		TotalAmount:       input.TotalAmount,
		HasInsurance:      input.HasInsurance,
		ScheduledDate:     input.ScheduledDate,
		CompletedAt:       input.CompletedAt,
		Memo:              input.Memo,
	}
	if input.Status != nil {
		s := model.BillingStatus(*input.Status)
		updateInput.Status = &s
	}

	ctx := c.Request.Context()
	updated, err := h.svc.Accounting.Update(ctx, updateInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountingResponse(updated))
}

func (h *Handler) DeleteAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Accounting.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterAccountingRoutes は会計関連のルートを登録する
func (h *Handler) RegisterAccountingRoutes(rg *gin.RouterGroup) {
	accountings := rg.Group("/accountings")
	accountings.GET("", h.ListAccountings)
	accountings.POST("", h.CreateAccounting)
	accountings.GET("/:id", h.GetAccounting)
	accountings.PATCH("/:id", h.UpdateAccounting)
	accountings.DELETE("/:id", h.DeleteAccounting)
	accountings.GET("/:id/refunds", h.ListRefunds)
	accountings.POST("/:id/refunds", h.CreateRefund)
}
