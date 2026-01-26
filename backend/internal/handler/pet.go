package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetPets godoc
// @Summary ペット一覧取得
// @Description 登録されているペットの一覧を取得します
// @Tags pets
// @Accept json
// @Produce json
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "1ページあたりの件数" default(10)
// @Param search query string false "検索キーワード（名前、種別、品種）"
// @Success 200 {array} model.Pet
// @Failure 500 {object} map[string]string
// @Router /pets [get]
// @Security ApiKeyAuth
func (h *Handler) GetPets(c *gin.Context) {
	ctx := c.Request.Context()

	pets, err := h.svc.GetAllPets(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get pets", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, pets)
}

// GetPet godoc
// @Summary ペット詳細取得
// @Description 指定されたIDのペット情報を取得します
// @Tags pets
// @Accept json
// @Produce json
// @Param id path string true "ペットID (UUID)"
// @Success 200 {object} model.Pet
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetPet(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	pet, err := h.svc.GetPetByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "pet", id)
		return
	}
	c.JSON(http.StatusOK, pet)
}

// CreatePet godoc
// @Summary ペット作成
// @Description 新しいペットを登録します
// @Tags pets
// @Accept json
// @Produce json
// @Param pet body model.CreatePetRequest true "ペット情報"
// @Success 201 {object} model.Pet
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets [post]
// @Security ApiKeyAuth
func (h *Handler) CreatePet(c *gin.Context) {
	ctx := c.Request.Context()

	var req model.CreatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	pet, err := h.svc.CreatePet(ctx, &req)
	if err != nil {
		h.handleError(c, err, "pet", "")
		return
	}

	slog.InfoContext(ctx, "pet created", slog.String("pet_id", pet.ID.String()))
	c.JSON(http.StatusCreated, pet)
}

// UpdatePet godoc
// @Summary ペット情報更新
// @Description 指定されたIDのペット情報を更新します
// @Tags pets
// @Accept json
// @Produce json
// @Param id path string true "ペットID (UUID)"
// @Param pet body model.UpdatePetRequest true "更新するペット情報"
// @Success 200 {object} model.Pet
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{id} [put]
// @Security ApiKeyAuth
func (h *Handler) UpdatePet(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	pet, err := h.svc.UpdatePet(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "pet", id)
		return
	}

	slog.InfoContext(ctx, "pet updated", slog.String("pet_id", id))
	c.JSON(http.StatusOK, pet)
}

// DeletePet godoc
// @Summary ペット削除
// @Description 指定されたIDのペットを削除します
// @Tags pets
// @Accept json
// @Produce json
// @Param id path string true "ペットID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /pets/{id} [delete]
func (h *Handler) DeletePet(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeletePet(ctx, id); err != nil {
		h.handleError(c, err, "pet", id)
		return
	}

	slog.InfoContext(ctx, "pet deleted", slog.String("pet_id", id))
	c.JSON(http.StatusOK, gin.H{"message": "pet deleted"})
}
