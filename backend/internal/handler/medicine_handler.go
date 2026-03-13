// Package handler provides HTTP handler implementations for Medicine entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Medicine ----

// ListMedicines godoc
// @Summary 薬剤一覧取得
// @Description クリニックに登録されている薬剤の一覧を返す。
// @Tags MedicineMasters
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Medicine
// @Failure 500 {object} map[string]string
// @Router /masters/medicines [get]
func (h *Handler) ListMedicines(c *gin.Context) {
	medicines, err := h.svc.Medicine.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, medicines)
}

// CreateMedicine godoc
// @Summary 薬剤作成
// @Description 新規薬剤を登録する。
// @Tags MedicineMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param medicine body model.Medicine true "薬剤情報"
// @Success 201 {object} model.Medicine
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/medicines [post]
func (h *Handler) CreateMedicine(c *gin.Context) {
	var input model.Medicine
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Medicine.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateMedicine godoc
// @Summary 薬剤更新
// @Description 指定IDの薬剤情報を更新する。
// @Tags MedicineMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "薬剤ID"
// @Param medicine body model.Medicine true "更新する薬剤情報"
// @Success 200 {object} model.Medicine
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/medicines/{id} [put]
func (h *Handler) UpdateMedicine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Medicine
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Medicine.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteMedicine godoc
// @Summary 薬剤削除
// @Description 指定IDの薬剤を削除する。
// @Tags MedicineMasters
// @Security BearerAuth
// @Param id path integer true "薬剤ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/medicines/{id} [delete]
func (h *Handler) DeleteMedicine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Medicine.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
