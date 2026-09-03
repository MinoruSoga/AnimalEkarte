package pet

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// chronicConditionResponse は慢性疾患フラグのレスポンス型（BE-012）。
type chronicConditionResponse struct {
	ID            uint64  `json:"id"`
	ClinicID      uint64  `json:"clinic_id"`
	PetID         uint64  `json:"pet_id"`
	ConditionCode string  `json:"condition_code"`
	ConditionName string  `json:"condition_name"`
	DiagnosedAt   string  `json:"diagnosed_at"`
	Notes         *string `json:"notes,omitempty"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func toChronicConditionResponse(r *model.PetChronicCondition) chronicConditionResponse {
	return chronicConditionResponse{
		ID:            r.ID,
		ClinicID:      r.ClinicID,
		PetID:         r.PetID,
		ConditionCode: r.ConditionCode,
		ConditionName: r.ConditionName,
		DiagnosedAt:   r.DiagnosedAt.In(time.Local).Format(time.DateOnly),
		Notes:         r.Notes,
		IsActive:      r.IsActive,
		CreatedAt:     httpapi.LocalTimeRFC3339(r.CreatedAt),
		UpdatedAt:     httpapi.LocalTimeRFC3339(r.UpdatedAt),
	}
}

func toChronicConditionListResponse(records []model.PetChronicCondition) []chronicConditionResponse {
	out := make([]chronicConditionResponse, len(records))
	for i := range records {
		out[i] = toChronicConditionResponse(&records[i])
	}
	return out
}

// ListChronicConditions GET /pets/:id/chronic-conditions
func (h *Handler) ListChronicConditions(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	records, err := h.chronicConditions.List(c.Request.Context(), clinicID, petID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChronicConditionListResponse(records))
}

func chronicConditionInputError(err error) error {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return err
	}
	return apperrors.WrapInvalidInput("日時の形式が正しくありません")
}

// CreateChronicCondition POST /pets/:id/chronic-conditions
func (h *Handler) CreateChronicCondition(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req createChronicConditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, chronicConditionInputError(err))
		return
	}

	record, err := h.chronicConditions.Create(c.Request.Context(), clinicID, petID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/pets/%d/chronic-conditions/%d", petID, record.ID))
	c.JSON(http.StatusCreated, toChronicConditionResponse(record))
}

// UpdateChronicCondition PATCH /pets/:id/chronic-conditions/:cc_id
func (h *Handler) UpdateChronicCondition(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	ccID, ok := httpapi.ParseIDParam(c, "cc_id")
	if !ok {
		return
	}

	var req updateChronicConditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, chronicConditionInputError(err))
		return
	}

	record, err := h.chronicConditions.Update(c.Request.Context(), clinicID, petID, ccID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChronicConditionResponse(record))
}

// DeleteChronicCondition DELETE /pets/:id/chronic-conditions/:cc_id
func (h *Handler) DeleteChronicCondition(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	ccID, ok := httpapi.ParseIDParam(c, "cc_id")
	if !ok {
		return
	}

	if err := h.chronicConditions.Delete(c.Request.Context(), clinicID, petID, ccID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
