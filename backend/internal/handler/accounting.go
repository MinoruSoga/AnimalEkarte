package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllAccounting godoc
// @Summary 会計一覧取得
// @Description 登録されている会計の一覧を取得します
// @Tags accountings
// @Accept json
// @Produce json
// @Success 200 {array} model.Accounting
// @Failure 500 {object} map[string]string
// @Router /accountings [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllAccounting(c *gin.Context) {
	ctx := c.Request.Context()

	accs, err := h.svc.GetAllAccounting(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get accountings", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, accs)
}

// GetAccountingByID godoc
// @Summary 会計詳細取得
// @Description 指定されたIDの会計情報を取得します
// @Tags accountings
// @Accept json
// @Produce json
// @Param id path string true "会計ID (UUID)"
// @Success 200 {object} model.Accounting
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings/{id} [get]
func (h *Handler) GetAccountingByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	acc, err := h.svc.GetAccountingByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "accounting", id)
		return
	}

	c.JSON(http.StatusOK, acc)
}

// GetAccountingByPetID godoc
// @Summary ペットの会計一覧取得
// @Description 指定されたペットIDの会計一覧を取得します
// @Tags accountings
// @Accept json
// @Produce json
// @Param petId path string true "ペットID (UUID)"
// @Success 200 {array} model.Accounting
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{petId}/accountings [get]
func (h *Handler) GetAccountingByPetID(c *gin.Context) {
	ctx := c.Request.Context()
	petID := c.Param("petId")

	accs, err := h.svc.GetAccountingByPetID(ctx, petID)
	if err != nil {
		h.handleError(c, err, "accounting", petID)
		return
	}

	c.JSON(http.StatusOK, accs)
}

// GetAccountingByOwnerID godoc
// @Summary 飼い主の会計一覧取得
// @Description 指定された飼い主IDの会計一覧を取得します
// @Tags accountings
// @Accept json
// @Produce json
// @Param ownerId path string true "飼い主ID (UUID)"
// @Success 200 {array} model.Accounting
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{ownerId}/accountings [get]
func (h *Handler) GetAccountingByOwnerID(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := c.Param("ownerId")

	accs, err := h.svc.GetAccountingByOwnerID(ctx, ownerID)
	if err != nil {
		h.handleError(c, err, "accounting", ownerID)
		return
	}

	c.JSON(http.StatusOK, accs)
}

// GetAccountingByStatus godoc
// @Summary ステータス別会計取得
// @Description 指定されたステータスの会計一覧を取得します
// @Tags accountings
// @Accept json
// @Produce json
// @Param status path string true "ステータス (未収,保留,回収済,キャンセル)"
// @Success 200 {array} model.Accounting
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings/status/{status} [get]
func (h *Handler) GetAccountingByStatus(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Param("status")

	accs, err := h.svc.GetAccountingByStatus(ctx, status)
	if err != nil {
		h.handleError(c, err, "accounting", status)
		return
	}

	c.JSON(http.StatusOK, accs)
}

// CreateAccounting godoc
// @Summary 会計作成
// @Description 新しい会計を作成します
// @Tags accountings
// @Accept json
// @Produce json
// @Param acc body model.CreateAccountingRequest true "会計情報"
// @Success 201 {object} model.Accounting
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings [post]
func (h *Handler) CreateAccounting(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateAccountingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	acc, err := h.svc.CreateAccounting(ctx, &req)
	if err != nil {
		h.handleError(c, err, "accounting", "")
		return
	}

	c.JSON(http.StatusCreated, acc)
}

// UpdateAccounting godoc
// @Summary 会計更新
// @Description 既存の会計を更新します
// @Tags accountings
// @Accept json
// @Produce json
// @Param id path string true "会計ID (UUID)"
// @Param acc body model.UpdateAccountingRequest true "会計情報"
// @Success 200 {object} model.Accounting
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings/{id} [put]
func (h *Handler) UpdateAccounting(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateAccountingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	acc, err := h.svc.UpdateAccounting(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "accounting", id)
		return
	}

	c.JSON(http.StatusOK, acc)
}

// DeleteAccounting godoc
// @Summary 会計削除
// @Description 指定された会計を削除します
// @Tags accountings
// @Accept json
// @Produce json
// @Param id path string true "会計ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings/{id} [delete]
func (h *Handler) DeleteAccounting(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteAccounting(ctx, id); err != nil {
		h.handleError(c, err, "accounting", id)
		return
	}

	c.Status(http.StatusNoContent)
}
