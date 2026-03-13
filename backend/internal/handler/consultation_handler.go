// Package handler provides HTTP handler implementations for Consultation entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Consultation ----

// ListConsultations godoc
// @Summary 診察項目一覧取得
// @Description 登録されている診察項目の一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Consultation
// @Failure 500 {object} map[string]string
// @Router /masters/consultations [get]
func (h *Handler) ListConsultations(c *gin.Context) {
	consultations, err := h.svc.Consultation.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, consultations)
}

// CreateConsultation godoc
// @Summary 診察項目作成
// @Description 新しい診察項目を作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.Consultation true "診察項目情報"
// @Success 201 {object} model.Consultation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/consultations [post]
func (h *Handler) CreateConsultation(c *gin.Context) {
	var input model.Consultation
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Consultation.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateConsultation godoc
// @Summary 診察項目更新
// @Description 指定IDの診察項目を更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "診察項目ID"
// @Param request body model.Consultation true "診察項目情報"
// @Success 200 {object} model.Consultation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/consultations/{id} [put]
func (h *Handler) UpdateConsultation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Consultation
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Consultation.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteConsultation godoc
// @Summary 診察項目削除
// @Description 指定IDの診察項目を削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "診察項目ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/consultations/{id} [delete]
func (h *Handler) DeleteConsultation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Consultation.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
