// Package handler provides HTTP handler implementations for Vaccine entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Vaccine ----

// ListVaccines godoc
// @Summary ワクチン一覧取得
// @Description クリニックに登録されているワクチンの一覧を返す。speciesクエリで対象種絞り込み可能。
// @Tags VaccineMasters
// @Produce json
// @Security BearerAuth
// @Param species query string false "対象種でフィルタリング（例: dog, cat, both）"
// @Success 200 {array} model.Vaccine
// @Failure 500 {object} map[string]string
// @Router /masters/vaccines [get]
func (h *Handler) ListVaccines(c *gin.Context) {
	var species *string
	if s := c.Query("species"); s != "" {
		species = &s
	}
	vaccines, err := h.svc.Vaccine.List(c.Request.Context(), species)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, vaccines)
}

// CreateVaccine godoc
// @Summary ワクチン作成
// @Description 新規ワクチンを登録する。
// @Tags VaccineMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vaccine body model.Vaccine true "ワクチン情報"
// @Success 201 {object} model.Vaccine
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/vaccines [post]
func (h *Handler) CreateVaccine(c *gin.Context) {
	var input model.Vaccine
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Vaccine.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateVaccine godoc
// @Summary ワクチン更新
// @Description 指定IDのワクチン情報を更新する。
// @Tags VaccineMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ワクチンID（UUID）"
// @Param vaccine body model.Vaccine true "更新するワクチン情報"
// @Success 200 {object} model.Vaccine
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/vaccines/{id} [put]
func (h *Handler) UpdateVaccine(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Vaccine
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Vaccine.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteVaccine godoc
// @Summary ワクチン削除
// @Description 指定IDのワクチンを削除する。
// @Tags VaccineMasters
// @Security BearerAuth
// @Param id path string true "ワクチンID（UUID）"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/vaccines/{id} [delete]
func (h *Handler) DeleteVaccine(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Vaccine.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
