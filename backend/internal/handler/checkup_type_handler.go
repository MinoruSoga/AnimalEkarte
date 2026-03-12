// Package handler provides HTTP handler implementations for CheckupType entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CheckupType ----

// ListCheckupTypes godoc
// @Summary 健診種別一覧取得
// @Description 登録されている健診種別の一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.CheckupType
// @Failure 500 {object} map[string]string
// @Router /masters/checkup-types [get]
func (h *Handler) ListCheckupTypes(c *gin.Context) {
	checkupTypes, err := h.svc.CheckupType.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, checkupTypes)
}

// CreateCheckupType godoc
// @Summary 健診種別作成
// @Description 新しい健診種別を作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.CheckupType true "健診種別情報"
// @Success 201 {object} model.CheckupType
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/checkup-types [post]
func (h *Handler) CreateCheckupType(c *gin.Context) {
	var input model.CheckupType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.CheckupType.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateCheckupType godoc
// @Summary 健診種別更新
// @Description 指定IDの健診種別を更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "健診種別UUID"
// @Param request body model.CheckupType true "健診種別情報"
// @Success 200 {object} model.CheckupType
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/checkup-types/{id} [put]
func (h *Handler) UpdateCheckupType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.CheckupType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.CheckupType.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteCheckupType godoc
// @Summary 健診種別削除
// @Description 指定IDの健診種別を削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "健診種別UUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/checkup-types/{id} [delete]
func (h *Handler) DeleteCheckupType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.CheckupType.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
