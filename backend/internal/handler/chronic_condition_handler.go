package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// chronicConditionResponse は慢性疾患フラグのレスポンス型（BE-012）。
type chronicConditionResponse struct {
	ID            uint64    `json:"id"`
	ClinicID      uint64    `json:"clinic_id"`
	PetID         uint64    `json:"pet_id"`
	ConditionCode string    `json:"condition_code"`
	ConditionName string    `json:"condition_name"`
	DiagnosedAt   string    `json:"diagnosed_at"`
	Notes         *string   `json:"notes,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toChronicConditionResponse(r *model.PetChronicCondition) chronicConditionResponse {
	return chronicConditionResponse{
		ID:            r.ID,
		ClinicID:      r.ClinicID,
		PetID:         r.PetID,
		ConditionCode: r.ConditionCode,
		ConditionName: r.ConditionName,
		DiagnosedAt:   r.DiagnosedAt.Format("2006-01-02"),
		Notes:         r.Notes,
		IsActive:      r.IsActive,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func toChronicConditionListResponse(records []model.PetChronicCondition) []chronicConditionResponse {
	out := make([]chronicConditionResponse, len(records))
	for i := range records {
		out[i] = toChronicConditionResponse(&records[i])
	}
	return out
}

// RegisterChronicConditionRoutes はペットIDネストの慢性疾患ルートを登録する（BE-012）。
func (h *Handler) RegisterChronicConditionRoutes(rg *gin.RouterGroup) {
	cc := rg.Group("/:id/chronic-conditions")
	cc.GET("", h.RequirePermission(string(model.ResourceOwners), "view"), h.ListChronicConditions)
	cc.POST("", h.RequirePermission(string(model.ResourceOwners), "create"), h.CreateChronicCondition)
	cc.PATCH("/:cc_id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateChronicCondition)
	cc.DELETE("/:cc_id", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteChronicCondition)
}

// ListChronicConditions GET /pets/:id/chronic-conditions
func (h *Handler) ListChronicConditions(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid pet id"))
		return
	}

	records, err := h.svc.ChronicCondition.List(c.Request.Context(), clinicID, petID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChronicConditionListResponse(records))
}

// CreateChronicCondition POST /pets/:id/chronic-conditions
func (h *Handler) CreateChronicCondition(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid pet id"))
		return
	}

	var req struct {
		ConditionCode string  `json:"condition_code" binding:"required"`
		ConditionName string  `json:"condition_name" binding:"required"`
		DiagnosedAt   string  `json:"diagnosed_at"   binding:"required"`
		Notes         *string `json:"notes"`
		IsActive      *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	diagnosedAt, err := time.Parse("2006-01-02", req.DiagnosedAt)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("diagnosed_at must be YYYY-MM-DD"))
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	input := service.CreateChronicConditionInput{
		ConditionCode: req.ConditionCode,
		ConditionName: req.ConditionName,
		DiagnosedAt:   diagnosedAt,
		Notes:         req.Notes,
		IsActive:      isActive,
	}

	record, err := h.svc.ChronicCondition.Create(c.Request.Context(), clinicID, petID, input)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/pets/%d/chronic-conditions/%d", petID, record.ID))
	c.JSON(http.StatusCreated, toChronicConditionResponse(record))
}

// UpdateChronicCondition PATCH /pets/:id/chronic-conditions/:cc_id
func (h *Handler) UpdateChronicCondition(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid pet id"))
		return
	}
	ccID, err := strconv.ParseUint(c.Param("cc_id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid cc_id"))
		return
	}

	var req struct {
		ConditionCode *string `json:"condition_code"`
		ConditionName *string `json:"condition_name"`
		DiagnosedAt   *string `json:"diagnosed_at"`
		Notes         *string `json:"notes"`
		IsActive      *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := service.UpdateChronicConditionInput{
		ConditionCode: req.ConditionCode,
		ConditionName: req.ConditionName,
		Notes:         req.Notes,
		IsActive:      req.IsActive,
	}
	if req.DiagnosedAt != nil {
		t, parseErr := time.Parse("2006-01-02", *req.DiagnosedAt)
		if parseErr != nil {
			RespondError(c, apperrors.WrapInvalidInput("diagnosed_at must be YYYY-MM-DD"))
			return
		}
		input.DiagnosedAt = &t
	}

	record, err := h.svc.ChronicCondition.Update(c.Request.Context(), clinicID, petID, ccID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toChronicConditionResponse(record))
}

// DeleteChronicCondition DELETE /pets/:id/chronic-conditions/:cc_id
func (h *Handler) DeleteChronicCondition(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid pet id"))
		return
	}
	ccID, err := strconv.ParseUint(c.Param("cc_id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid cc_id"))
		return
	}

	if err := h.svc.ChronicCondition.Delete(c.Request.Context(), clinicID, petID, ccID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
