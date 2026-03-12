// Package handler provides HTTP handler implementations for Cage entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Cage ----

// ListCages godoc
// @Summary ケージ一覧取得
// @Description クリニックに登録されているケージの一覧を返す。cage_typeクエリで種別絞り込み可能。
// @Tags CageMasters
// @Produce json
// @Security BearerAuth
// @Param cage_type query string false "ケージ種別でフィルタリング（例: icu, dog, cat, general）"
// @Success 200 {array} model.Cage
// @Failure 500 {object} map[string]string
// @Router /masters/cages [get]
func (h *Handler) ListCages(c *gin.Context) {
	var cageType *string
	if t := c.Query("cage_type"); t != "" {
		cageType = &t
	}
	cages, err := h.svc.Cage.List(c.Request.Context(), cageType)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, cages)
}

// CreateCage godoc
// @Summary ケージ作成
// @Description 新規ケージを登録する。
// @Tags CageMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cage body model.Cage true "ケージ情報"
// @Success 201 {object} model.Cage
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/cages [post]
func (h *Handler) CreateCage(c *gin.Context) {
	var input model.Cage
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Cage.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateCage godoc
// @Summary ケージ更新
// @Description 指定IDのケージ情報を更新する。
// @Tags CageMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ケージID（UUID）"
// @Param cage body model.Cage true "更新するケージ情報"
// @Success 200 {object} model.Cage
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/cages/{id} [put]
func (h *Handler) UpdateCage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Cage
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Cage.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteCage godoc
// @Summary ケージ削除
// @Description 指定IDのケージを削除する。
// @Tags CageMasters
// @Security BearerAuth
// @Param id path string true "ケージID（UUID）"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/cages/{id} [delete]
func (h *Handler) DeleteCage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Cage.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
