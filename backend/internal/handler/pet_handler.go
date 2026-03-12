package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListPets godoc
// @Summary ペット一覧取得
// @Description クリニックに所属するペットの一覧をページネーションで返す
// @Tags Pets
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param search query string false "検索キーワード"
// @Param owner_id query string false "飼主UUID"
// @Success 200 {object} handler.PaginatedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets [get]
func (h *Handler) ListPets(c *gin.Context) {
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

	var ownerID *uuid.UUID
	if ownerIDStr := c.Query("owner_id"); ownerIDStr != "" {
		id, err := uuid.Parse(ownerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
		ownerID = &id
	}

	pets, total, err := h.svc.Pet.List(c.Request.Context(), clinicID, ownerID, page, limit, search)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: pets, Total: total, Page: page, Limit: limit})
}

// GetPet godoc
// @Summary ペット取得
// @Description 指定IDのペットを返す
// @Tags Pets
// @Produce json
// @Security BearerAuth
// @Param id path string true "ペットUUID"
// @Success 200 {object} model.Pet
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{id} [get]
func (h *Handler) GetPet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	pet, err := h.svc.Pet.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pet)
}

// CreatePet godoc
// @Summary ペット作成
// @Description 新規ペットを作成する
// @Tags Pets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.Pet true "ペット情報"
// @Success 201 {object} model.Pet
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets [post]
func (h *Handler) CreatePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input model.Pet
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	input.ClinicID = clinicID
	if err := h.svc.Pet.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdatePet godoc
// @Summary ペット更新
// @Description 指定IDのペット情報を更新する
// @Tags Pets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ペットUUID"
// @Param body body model.Pet true "ペット情報"
// @Success 200 {object} model.Pet
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{id} [put]
func (h *Handler) UpdatePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Pet
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	input.ClinicID = clinicID
	if err := h.svc.Pet.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeletePet godoc
// @Summary ペット削除
// @Description 指定IDのペットを削除する
// @Tags Pets
// @Security BearerAuth
// @Param id path string true "ペットUUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{id} [delete]
func (h *Handler) DeletePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Pet.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
