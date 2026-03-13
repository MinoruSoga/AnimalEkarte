package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createAccountingInput struct {
	MedicalRecordID   *uint64              `json:"medical_record_id"`
	HospitalizationID *uint64              `json:"hospitalization_id"`
	OwnerID           *uint64              `json:"owner_id"`
	PetID             *uint64              `json:"pet_id"`
	Subtotal          int                  `json:"subtotal"`
	TaxTotal          int                  `json:"tax_total"`
	TotalAmount       int                  `json:"total_amount"`
	HasInsurance      bool                 `json:"has_insurance"`
	Status            string               `json:"status"`
	ScheduledDate     time.Time            `json:"scheduled_date" binding:"required"`
	CompletedAt       *time.Time           `json:"completed_at"`
	Memo              string               `json:"memo"`
}

type updateAccountingInput struct {
	MedicalRecordID   *uint64              `json:"medical_record_id"`
	HospitalizationID *uint64              `json:"hospitalization_id"`
	OwnerID           *uint64              `json:"owner_id"`
	PetID             *uint64              `json:"pet_id"`
	Subtotal          int                  `json:"subtotal"`
	TaxTotal          int                  `json:"tax_total"`
	TotalAmount       int                  `json:"total_amount"`
	HasInsurance      bool                 `json:"has_insurance"`
	Status            string               `json:"status"`
	ScheduledDate     *time.Time           `json:"scheduled_date"`
	CompletedAt       *time.Time           `json:"completed_at"`
	Memo              string               `json:"memo"`
}

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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id"})
			return
		}
		petID = &id
	}
	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
		ownerID = &id
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	accountings, total, err := h.svc.Accounting.List(c.Request.Context(), clinicID, petID, ownerID, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: accountings, Total: total, Page: page, Limit: limit})
}

// GetAccounting godoc
func (h *Handler) GetAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	accounting, err := h.svc.Accounting.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, accounting)
}

// CreateAccounting godoc
func (h *Handler) CreateAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createAccountingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	billing := &model.Billing{
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
		billing.Status = model.BillingStatus(input.Status)
	}

	if err := h.svc.Accounting.Create(c.Request.Context(), clinicID, billing); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, billing)
}

// UpdateAccounting godoc
func (h *Handler) UpdateAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateAccountingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	billing := &model.Billing{
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
		CompletedAt:       input.CompletedAt,
		Memo:              input.Memo,
	}
	if input.ScheduledDate != nil {
		billing.ScheduledDate = *input.ScheduledDate
	}
	if input.Status != "" {
		billing.Status = model.BillingStatus(input.Status)
	}

	if err := h.svc.Accounting.Update(c.Request.Context(), clinicID, billing); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, billing)
}

func (h *Handler) DeleteAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Accounting.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
