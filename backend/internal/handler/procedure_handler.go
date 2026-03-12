// Package handler provides HTTP handler implementations for Procedure entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Procedure ----

// ListProcedures godoc
// @Summary 処置項目一覧取得
// @Description 登録されている処置項目の一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Procedure
// @Failure 500 {object} map[string]string
// @Router /masters/procedures [get]
func (h *Handler) ListProcedures(c *gin.Context) {
	procedures, err := h.svc.Procedure.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, procedures)
}

// CreateProcedure godoc
// @Summary 処置項目作成
// @Description 新しい処置項目を作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.Procedure true "処置項目情報"
// @Success 201 {object} model.Procedure
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/procedures [post]
func (h *Handler) CreateProcedure(c *gin.Context) {
	var input model.Procedure
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Procedure.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateProcedure godoc
// @Summary 処置項目更新
// @Description 指定IDの処置項目を更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "処置項目UUID"
// @Param request body model.Procedure true "処置項目情報"
// @Success 200 {object} model.Procedure
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/procedures/{id} [put]
func (h *Handler) UpdateProcedure(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Procedure
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Procedure.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteProcedure godoc
// @Summary 処置項目削除
// @Description 指定IDの処置項目を削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "処置項目UUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/procedures/{id} [delete]
func (h *Handler) DeleteProcedure(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Procedure.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
