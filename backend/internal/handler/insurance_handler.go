// Package handler provides HTTP handler implementations for Insurance entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Insurance ----

// ListInsurances godoc
// @Summary 保険一覧取得
// @Description クリニックに登録されている保険の一覧を返す。
// @Tags InsuranceMasters
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Insurance
// @Failure 500 {object} map[string]string
// @Router /masters/insurances [get]
func (h *Handler) ListInsurances(c *gin.Context) {
	insurances, err := h.svc.Insurance.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, insurances)
}

// CreateInsurance godoc
// @Summary 保険作成
// @Description 新規保険を登録する。
// @Tags InsuranceMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param insurance body model.Insurance true "保険情報"
// @Success 201 {object} model.Insurance
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/insurances [post]
func (h *Handler) CreateInsurance(c *gin.Context) {
	var input model.Insurance
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Insurance.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateInsurance godoc
// @Summary 保険更新
// @Description 指定IDの保険情報を更新する。
// @Tags InsuranceMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "保険ID（UUID）"
// @Param insurance body model.Insurance true "更新する保険情報"
// @Success 200 {object} model.Insurance
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/insurances/{id} [put]
func (h *Handler) UpdateInsurance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Insurance
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Insurance.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteInsurance godoc
// @Summary 保険削除
// @Description 指定IDの保険を削除する。
// @Tags InsuranceMasters
// @Security BearerAuth
// @Param id path string true "保険ID（UUID）"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/insurances/{id} [delete]
func (h *Handler) DeleteInsurance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Insurance.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
