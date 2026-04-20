package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- request types ----

type createPaymentMethodRequest struct {
	Name         string `json:"name"          binding:"required"`
	DisplayOrder int    `json:"display_order"`
}

type updatePaymentMethodRequest struct {
	Name         *string `json:"name"`
	DisplayOrder *int    `json:"display_order"`
	IsActive     *bool   `json:"is_active"`
}

// ---- handlers ----

// ListPaymentMethods godoc
// GET /v1/payment-methods
func (h *Handler) ListPaymentMethods(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ms, err := h.svc.PaymentMethodMaster.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ms)
}

// CreatePaymentMethod godoc
// POST /v1/payment-methods
func (h *Handler) CreatePaymentMethod(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	m, err := h.svc.PaymentMethodMaster.Create(c.Request.Context(), clinicID, service.CreatePaymentMethodInput{
		Name:         req.Name,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/payment-methods/%d", m.ID))
	c.JSON(http.StatusCreated, m)
}

// UpdatePaymentMethod godoc
// PATCH /v1/payment-methods/:id
func (h *Handler) UpdatePaymentMethod(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	m, err := h.svc.PaymentMethodMaster.Update(c.Request.Context(), clinicID, id, service.UpdatePaymentMethodInput{
		Name:         req.Name,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

// DeletePaymentMethod godoc
// DELETE /v1/payment-methods/:id
func (h *Handler) DeletePaymentMethod(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.PaymentMethodMaster.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterPaymentMethodMasterRoutes は支払方法マスタ関連のルートを登録する
func (h *Handler) RegisterPaymentMethodMasterRoutes(rg *gin.RouterGroup) {
	pm := rg.Group("/payment-methods")
	pm.GET("", h.RequirePermission(string(model.ResourceClosingSettings), "view"), h.ListPaymentMethods)
	pm.POST("", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.CreatePaymentMethod)
	pm.PATCH("/:id", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.UpdatePaymentMethod)
	pm.DELETE("/:id", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.DeletePaymentMethod)
}
