// Package handler provides HTTP handler implementations for ServiceType entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ServiceType ----

// ListServiceTypes godoc
// @Summary サービス種別一覧取得
// @Description 登録されているサービス種別の一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.ServiceType
// @Failure 500 {object} map[string]string
// @Router /masters/service-types [get]
func (h *Handler) ListServiceTypes(c *gin.Context) {
	serviceTypes, err := h.svc.ServiceType.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, serviceTypes)
}

// CreateServiceType godoc
// @Summary サービス種別作成
// @Description 新しいサービス種別を作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.ServiceType true "サービス種別情報"
// @Success 201 {object} model.ServiceType
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/service-types [post]
func (h *Handler) CreateServiceType(c *gin.Context) {
	var input model.ServiceType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.ServiceType.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateServiceType godoc
// @Summary サービス種別更新
// @Description 指定IDのサービス種別を更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "サービス種別ID"
// @Param request body model.ServiceType true "サービス種別情報"
// @Success 200 {object} model.ServiceType
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/service-types/{id} [put]
func (h *Handler) UpdateServiceType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.ServiceType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.ServiceType.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteServiceType godoc
// @Summary サービス種別削除
// @Description 指定IDのサービス種別を削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "サービス種別ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/service-types/{id} [delete]
func (h *Handler) DeleteServiceType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.ServiceType.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
