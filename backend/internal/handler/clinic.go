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
// @Param clinic body model.CreateClinicRequest true "クリニック情報"
// @Success 201 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics [post]
func (h *Handler) CreateClinic(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateClinicRequest
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
// @Param clinic body model.UpdateClinicRequest true "クリニック情報"
// @Success 200 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{id} [put]
func (h *Handler) UpdateClinic(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateClinicRequest
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

// GetAllStaff godoc
// @Summary スタッフ一覧取得
// @Description 登録されているスタッフの一覧を取得します
// @Tags staffs
// @Accept json
// @Produce json
// @Success 200 {array} model.Staff
// @Failure 500 {object} map[string]string
// @Router /staffs [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllStaff(c *gin.Context) {
	ctx := c.Request.Context()

	staffs, err := h.svc.GetAllStaff(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get staffs", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, staffs)
}

// GetStaffByID godoc
// @Summary スタッフ詳細取得
// @Description 指定されたIDのスタッフ情報を取得します
// @Tags staffs
// @Accept json
// @Produce json
// @Param id path string true "スタッフID (UUID)"
// @Success 200 {object} model.Staff
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /staffs/{id} [get]
func (h *Handler) GetStaffByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	staff, err := h.svc.GetStaffByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "staff", id)
		return
	}

	c.JSON(http.StatusOK, staff)
}

// GetStaffByClinicID godoc
// @Summary クリニックのスタッフ一覧取得
// @Description 指定されたクリニックIDのスタッフ一覧を取得します
// @Tags staffs
// @Accept json
// @Produce json
// @Param clinicId path string true "クリニックID (UUID)"
// @Success 200 {array} model.Staff
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{clinicId}/staffs [get]
func (h *Handler) GetStaffByClinicID(c *gin.Context) {
	ctx := c.Request.Context()
	clinicID := c.Param("clinicId")

	staffs, err := h.svc.GetStaffByClinicID(ctx, clinicID)
	if err != nil {
		h.handleError(c, err, "staff", clinicID)
		return
	}

	c.JSON(http.StatusOK, staffs)
}

// CreateStaff godoc
// @Summary スタッフ作成
// @Description 新しいスタッフを作成します
// @Tags staffs
// @Accept json
// @Produce json
// @Param staff body model.CreateStaffRequest true "スタッフ情報"
// @Success 201 {object} model.Staff
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /staffs [post]
func (h *Handler) CreateStaff(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	staff, err := h.svc.CreateStaff(ctx, &req)
	if err != nil {
		h.handleError(c, err, "staff", "")
		return
	}

	c.JSON(http.StatusCreated, staff)
}

// UpdateStaff godoc
// @Summary スタッフ更新
// @Description 既存のスタッフを更新します
// @Tags staffs
// @Accept json
// @Produce json
// @Param id path string true "スタッフID (UUID)"
// @Param staff body model.UpdateStaffRequest true "スタッフ情報"
// @Success 200 {object} model.Staff
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /staffs/{id} [put]
func (h *Handler) UpdateStaff(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	staff, err := h.svc.UpdateStaff(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "staff", id)
		return
	}

	c.JSON(http.StatusOK, staff)
}

// DeleteStaff godoc
// @Summary スタッフ削除
// @Description 指定されたスタッフを削除します
// @Tags staffs
// @Accept json
// @Produce json
// @Param id path string true "スタッフID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /staffs/{id} [delete]
func (h *Handler) DeleteStaff(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteStaff(ctx, id); err != nil {
		h.handleError(c, err, "staff", id)
		return
	}

	c.Status(http.StatusNoContent)
}
