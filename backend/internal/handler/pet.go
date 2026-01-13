package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetPets godoc
// @Summary ペット一覧取得
// @Description 登録されているペットの一覧を取得します
// @Tags pets
// @Accept json
// @Produce json
// @Success 200 {array} model.Pet
// @Failure 500 {object} map[string]string
// @Router /pets [get]
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
// @Router /pets/{id} [get]
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
// @Summary ペット新規登録
// @Description 新しいペットを登録します
// @Tags pets
// @Accept json
// @Produce json
// @Param pet body model.CreatePetRequest true "ペット情報"
// @Success 201 {object} model.Pet
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets [post]
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
// @Router /pets/{id} [put]
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

// handleError はエラータイプに応じて適切なHTTPレスポンスを返す
// nolint:unparam // resource is kept for future extensibility with other resource types
func (h *Handler) handleError(c *gin.Context, err error, resource, id string) {
	ctx := c.Request.Context()

	switch {
	case apperrors.IsNotFound(err):
		slog.WarnContext(ctx, "resource not found",
			slog.String("resource", resource),
			slog.String("id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

	case apperrors.IsInvalidInput(err):
		slog.WarnContext(ctx, "invalid input",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	default:
		slog.ErrorContext(ctx, "internal error",
			slog.String("error", err.Error()),
			slog.String("resource", resource),
			slog.String("id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
