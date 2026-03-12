// Package handler provides HTTP handler implementations for HospitalizationPlan entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- HospitalizationPlan ----

// ListHospitalizationPlans godoc
// @Summary 入院プラン一覧取得
// @Description 登録されている入院プランの一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.HospitalizationPlan
// @Failure 500 {object} map[string]string
// @Router /masters/hospitalization-plans [get]
func (h *Handler) ListHospitalizationPlans(c *gin.Context) {
	plans, err := h.svc.HospitalizationPlan.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, plans)
}

// CreateHospitalizationPlan godoc
// @Summary 入院プラン作成
// @Description 新しい入院プランを作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.HospitalizationPlan true "入院プラン情報"
// @Success 201 {object} model.HospitalizationPlan
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/hospitalization-plans [post]
func (h *Handler) CreateHospitalizationPlan(c *gin.Context) {
	var input model.HospitalizationPlan
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.HospitalizationPlan.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateHospitalizationPlan godoc
// @Summary 入院プラン更新
// @Description 指定IDの入院プランを更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "入院プランUUID"
// @Param request body model.HospitalizationPlan true "入院プラン情報"
// @Success 200 {object} model.HospitalizationPlan
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/hospitalization-plans/{id} [put]
func (h *Handler) UpdateHospitalizationPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.HospitalizationPlan
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.HospitalizationPlan.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteHospitalizationPlan godoc
// @Summary 入院プラン削除
// @Description 指定IDの入院プランを削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "入院プランUUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/hospitalization-plans/{id} [delete]
func (h *Handler) DeleteHospitalizationPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.HospitalizationPlan.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
