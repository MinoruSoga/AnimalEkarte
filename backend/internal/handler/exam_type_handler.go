// Package handler provides HTTP handler implementations for ExaminationType entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ExaminationType ----

// ListExaminationTypes godoc
// @Summary 検査種別一覧取得
// @Description 登録されている検査種別の一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.ExamType
// @Failure 500 {object} map[string]string
// @Router /masters/examination-types [get]
func (h *Handler) ListExaminationTypes(c *gin.Context) {
	exTypes, err := h.svc.ExaminationType.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, exTypes)
}

// CreateExaminationType godoc
// @Summary 検査種別作成
// @Description 新しい検査種別を作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.ExamType true "検査種別情報"
// @Success 201 {object} model.ExamType
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/examination-types [post]
func (h *Handler) CreateExaminationType(c *gin.Context) {
	var input model.ExamType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.ExaminationType.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateExaminationType godoc
// @Summary 検査種別更新
// @Description 指定IDの検査種別を更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "検査種別ID"
// @Param request body model.ExamType true "検査種別情報"
// @Success 200 {object} model.ExamType
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/examination-types/{id} [put]
func (h *Handler) UpdateExaminationType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.ExamType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.ExaminationType.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteExaminationType godoc
// @Summary 検査種別削除
// @Description 指定IDの検査種別を削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "検査種別ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/examination-types/{id} [delete]
func (h *Handler) DeleteExaminationType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.ExaminationType.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
