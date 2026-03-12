package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListOwners godoc
// @Summary 飼主一覧取得
// @Description クリニックに所属する飼主の一覧をページネーションで返す
// @Tags Owners
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param search query string false "検索キーワード"
// @Success 200 {object} handler.PaginatedResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners [get]
func (h *Handler) ListOwners(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}
	search := c.Query("search")

	owners, total, err := h.svc.Owner.List(c.Request.Context(), clinicID, page, limit, search)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: owners, Total: total, Page: page, Limit: limit})
}

// GetOwner godoc
// @Summary 飼主取得
// @Description 指定IDの飼主を返す
// @Tags Owners
// @Produce json
// @Security BearerAuth
// @Param id path string true "飼主UUID"
// @Success 200 {object} model.Owner
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{id} [get]
func (h *Handler) GetOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	owner, err := h.svc.Owner.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, owner)
}

// CreateOwner godoc
// @Summary 飼主作成
// @Description 新規飼主を作成する
// @Tags Owners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.Owner true "飼主情報"
// @Success 201 {object} model.Owner
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners [post]
func (h *Handler) CreateOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input model.Owner
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	input.ClinicID = clinicID
	if err := h.svc.Owner.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateOwner godoc
// @Summary 飼主更新
// @Description 指定IDの飼主情報を更新する
// @Tags Owners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "飼主UUID"
// @Param body body model.Owner true "飼主情報"
// @Success 200 {object} model.Owner
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{id} [put]
func (h *Handler) UpdateOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Owner
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	input.ClinicID = clinicID
	if err := h.svc.Owner.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteOwner godoc
// @Summary 飼主削除
// @Description 指定IDの飼主を削除する
// @Tags Owners
// @Security BearerAuth
// @Param id path string true "飼主UUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{id} [delete]
func (h *Handler) DeleteOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Owner.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
