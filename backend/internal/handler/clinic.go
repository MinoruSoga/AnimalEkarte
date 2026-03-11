package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllClinics godoc
// @Summary クリニック一覧取得
// @Description 登録されているクリニックの一覧を取得します
// @Tags clinics
// @Accept json
// @Produce json
// @Success 200 {array} model.Clinic
// @Failure 500 {object} map[string]string
// @Router /clinics [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllClinics(c *gin.Context) {
	ctx := c.Request.Context()

	clinics, err := h.svc.GetAllClinics(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinics", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, clinics)
}

// GetClinicByID godoc
// @Summary クリニック詳細取得
// @Description 指定されたIDのクリニック情報を取得します
// @Tags clinics
// @Accept json
// @Produce json
// @Param id path string true "クリニックID (UUID)"
// @Success 200 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{id} [get]
func (h *Handler) GetClinicByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	clinic, err := h.svc.GetClinicByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "clinic", id)
		return
	}

	c.JSON(http.StatusOK, clinic)
}

// CreateClinic godoc
// @Summary クリニック作成
// @Description 新しいクリニックを作成します
// @Tags clinics
// @Accept json
// @Produce json
// @Param clinic body model.CreateClinicInfoRequest true "クリニック情報"
// @Success 201 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics [post]
func (h *Handler) CreateClinic(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateClinicInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	clinic, err := h.svc.CreateClinic(ctx, &req)
	if err != nil {
		h.handleError(c, err, "clinic", "")
		return
	}

	c.JSON(http.StatusCreated, clinic)
}

// UpdateClinic godoc
// @Summary クリニック更新
// @Description 既存のクリニックを更新します
// @Tags clinics
// @Accept json
// @Produce json
// @Param id path string true "クリニックID (UUID)"
// @Param clinic body model.UpdateClinicInfoRequest true "クリニック情報"
// @Success 200 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{id} [put]
func (h *Handler) UpdateClinic(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateClinicInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	clinic, err := h.svc.UpdateClinic(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "clinic", id)
		return
	}

	c.JSON(http.StatusOK, clinic)
}

// DeleteClinic godoc
// @Summary クリニック削除
// @Description 指定されたクリニックを削除します
// @Tags clinics
// @Accept json
// @Produce json
// @Param id path string true "クリニックID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{id} [delete]
func (h *Handler) DeleteClinic(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteClinic(ctx, id); err != nil {
		h.handleError(c, err, "clinic", id)
		return
	}

	c.Status(http.StatusNoContent)
}
