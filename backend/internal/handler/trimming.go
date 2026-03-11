package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllTrimmings godoc
// @Summary トリミング一覧取得
// @Description 登録されているトリミングの一覧を取得します
// @Tags trimmings
// @Accept json
// @Produce json
// @Success 200 {array} model.TrimmingRecord
// @Failure 500 {object} map[string]string
// @Router /trimmings [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllTrimmings(c *gin.Context) {
	ctx := c.Request.Context()

	trims, err := h.svc.GetAllTrimmings(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get trimmings", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, trims)
}

// GetTrimmingByID godoc
// @Summary トリミング詳細取得
// @Description 指定されたIDのトリミング情報を取得します
// @Tags trimmings
// @Accept json
// @Produce json
// @Param id path string true "トリミングID (UUID)"
// @Success 200 {object} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings/{id} [get]
func (h *Handler) GetTrimmingByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	trim, err := h.svc.GetTrimmingByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "trimming", id)
		return
	}

	c.JSON(http.StatusOK, trim)
}

// GetTrimmingsByPetID godoc
// @Summary ペットのトリミング一覧取得
// @Description 指定されたペットIDのトリミング一覧を取得します
// @Tags trimmings
// @Accept json
// @Produce json
// @Param petId path string true "ペットID (UUID)"
// @Success 200 {array} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{petId}/trimmings [get]
func (h *Handler) GetTrimmingsByPetID(c *gin.Context) {
	ctx := c.Request.Context()
	petID := c.Param("petId")

	trims, err := h.svc.GetTrimmingsByPetID(ctx, petID)
	if err != nil {
		h.handleError(c, err, "trimming", petID)
		return
	}

	c.JSON(http.StatusOK, trims)
}

// GetTrimmingsByOwnerID godoc
// @Summary 飼い主のトリミング一覧取得
// @Description 指定された飼い主IDのトリミング一覧を取得します
// @Tags trimmings
// @Accept json
// @Produce json
// @Param ownerId path string true "飼い主ID (UUID)"
// @Success 200 {array} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{ownerId}/trimmings [get]
func (h *Handler) GetTrimmingsByOwnerID(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := c.Param("ownerId")

	trims, err := h.svc.GetTrimmingsByOwnerID(ctx, ownerID)
	if err != nil {
		h.handleError(c, err, "trimming", ownerID)
		return
	}

	c.JSON(http.StatusOK, trims)
}

// GetTrimmingsByStatus godoc
// @Summary ステータス別トリミング取得
// @Description 指定されたステータスのトリミング一覧を取得します
// @Tags trimmings
// @Accept json
// @Produce json
// @Param status path string true "ステータス (予約,進行中,完了)"
// @Success 200 {array} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings/status/{status} [get]
func (h *Handler) GetTrimmingsByStatus(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Param("status")

	trims, err := h.svc.GetTrimmingsByStatus(ctx, status)
	if err != nil {
		h.handleError(c, err, "trimming", status)
		return
	}

	c.JSON(http.StatusOK, trims)
}

// CreateTrimming godoc
// @Summary トリミング作成
// @Description 新しいトリミングを作成します
// @Tags trimmings
// @Accept json
// @Produce json
// @Param trim body model.CreateTrimmingRequest true "トリミング情報"
// @Success 201 {object} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings [post]
func (h *Handler) CreateTrimming(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateTrimmingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	trim, err := h.svc.CreateTrimming(ctx, &req)
	if err != nil {
		h.handleError(c, err, "trimming", "")
		return
	}

	c.JSON(http.StatusCreated, trim)
}

// UpdateTrimming godoc
// @Summary トリミング更新
// @Description 既存のトリミングを更新します
// @Tags trimmings
// @Accept json
// @Produce json
// @Param id path string true "トリミングID (UUID)"
// @Param trim body model.UpdateTrimmingRequest true "トリミング情報"
// @Success 200 {object} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings/{id} [put]
func (h *Handler) UpdateTrimming(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateTrimmingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	trim, err := h.svc.UpdateTrimming(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "trimming", id)
		return
	}

	c.JSON(http.StatusOK, trim)
}

// DeleteTrimming godoc
// @Summary トリミング削除
// @Description 指定されたトリミングを削除します
// @Tags trimmings
// @Accept json
// @Produce json
// @Param id path string true "トリミングID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings/{id} [delete]
func (h *Handler) DeleteTrimming(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteTrimming(ctx, id); err != nil {
		h.handleError(c, err, "trimming", id)
		return
	}

	c.Status(http.StatusNoContent)
}
