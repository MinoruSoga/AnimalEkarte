package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllHospitalizations godoc
// @Summary 入院/ホテル一覧取得
// @Description 登録されている入院/ホテルの一覧を取得します
// @Tags hospitalizations
// @Accept json
// @Produce json
// @Param status query string false "ステータスでフィルタリング"
// @Success 200 {array} model.Hospitalization
// @Failure 500 {object} map[string]string
// @Router /hospitalizations [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllHospitalizations(c *gin.Context) {
	ctx := c.Request.Context()

	hosps, err := h.svc.GetAllHospitalizations(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get hospitalizations", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, hosps)
}

// GetHospitalizationByID godoc
// @Summary 入院/ホテル詳細取得
// @Description 指定されたIDの入院/ホテル情報を取得します
// @Tags hospitalizations
// @Accept json
// @Produce json
// @Param id path string true "入院/ホテルID (UUID)"
// @Success 200 {object} model.Hospitalization
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /hospitalizations/{id} [get]
func (h *Handler) GetHospitalizationByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	hosp, err := h.svc.GetHospitalizationByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "hospitalization", id)
		return
	}

	c.JSON(http.StatusOK, hosp)
}

// GetHospitalizationsByPetID godoc
// @Summary ペットの入院/ホテル一覧取得
// @Description 指定されたペットIDの入院/ホテル一覧を取得します
// @Tags hospitalizations
// @Accept json
// @Produce json
// @Param petId path string true "ペットID (UUID)"
// @Success 200 {array} model.Hospitalization
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{petId}/hospitalizations [get]
func (h *Handler) GetHospitalizationsByPetID(c *gin.Context) {
	ctx := c.Request.Context()
	petID := c.Param("id")

	hosps, err := h.svc.GetHospitalizationsByPetID(ctx, petID)
	if err != nil {
		h.handleError(c, err, "hospitalization", petID)
		return
	}

	c.JSON(http.StatusOK, hosps)
}

// GetHospitalizationsByOwnerID godoc
// @Summary 飼い主の入院/ホテル一覧取得
// @Description 指定された飼い主IDの入院/ホテル一覧を取得します
// @Tags hospitalizations
// @Accept json
// @Produce json
// @Param ownerId path string true "飼い主ID (UUID)"
// @Success 200 {array} model.Hospitalization
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{ownerId}/hospitalizations [get]
func (h *Handler) GetHospitalizationsByOwnerID(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := c.Param("id")

	hosps, err := h.svc.GetHospitalizationsByOwnerID(ctx, ownerID)
	if err != nil {
		h.handleError(c, err, "hospitalization", ownerID)
		return
	}

	c.JSON(http.StatusOK, hosps)
}

// GetHospitalizationsByStatus godoc
// @Summary ステータス別入院/ホテル取得
// @Description 指定されたステータスの入院/ホテル一覧を取得します
// @Tags hospitalizations
// @Accept json
// @Produce json
// @Param status path string true "ステータス (予約,入院中,退院済,一時帰宅)"
// @Success 200 {array} model.Hospitalization
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /hospitalizations/status/{status} [get]
func (h *Handler) GetHospitalizationsByStatus(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Param("status")

	hosps, err := h.svc.GetHospitalizationsByStatus(ctx, status)
	if err != nil {
		h.handleError(c, err, "hospitalization", status)
		return
	}

	c.JSON(http.StatusOK, hosps)
}

// CreateHospitalization godoc
// @Summary 入院/ホテル作成
// @Description 新しい入院/ホテルを作成します
// @Tags hospitalizations
// @Accept json
// @Produce json
// @Param hosp body model.CreateHospitalizationRequest true "入院/ホテル情報"
// @Success 201 {object} model.Hospitalization
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /hospitalizations [post]
func (h *Handler) CreateHospitalization(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateHospitalizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	hosp, err := h.svc.CreateHospitalization(ctx, &req)
	if err != nil {
		h.handleError(c, err, "hospitalization", "")
		return
	}

	c.JSON(http.StatusCreated, hosp)
}

// UpdateHospitalization godoc
// @Summary 入院/ホテル更新
// @Description 既存の入院/ホテルを更新します
// @Tags hospitalizations
// @Accept json
// @Produce json
// @Param id path string true "入院/ホテルID (UUID)"
// @Param hosp body model.UpdateHospitalizationRequest true "入院/ホテル情報"
// @Success 200 {object} model.Hospitalization
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /hospitalizations/{id} [put]
func (h *Handler) UpdateHospitalization(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateHospitalizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	hosp, err := h.svc.UpdateHospitalization(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "hospitalization", id)
		return
	}

	c.JSON(http.StatusOK, hosp)
}

// DeleteHospitalization godoc
// @Summary 入院/ホテル削除
// @Description 指定された入院/ホテルを削除します
// @Tags hospitalizations
// @Accept json
// @Produce json
// @Param id path string true "入院/ホテルID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /hospitalizations/{id} [delete]
func (h *Handler) DeleteHospitalization(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteHospitalization(ctx, id); err != nil {
		h.handleError(c, err, "hospitalization", id)
		return
	}

	c.Status(http.StatusNoContent)
}
