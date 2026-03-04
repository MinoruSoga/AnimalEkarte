package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllExaminations godoc
// @Summary 検査一覧取得
// @Description 登録されている検査の一覧を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Success 200 {array} model.Examination
// @Failure 500 {object} map[string]string
// @Router /examinations [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllExaminations(c *gin.Context) {
	ctx := c.Request.Context()

	exams, err := h.svc.GetAllExaminations(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get examinations", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, exams)
}

// GetExaminationByID godoc
// @Summary 検査詳細取得
// @Description 指定されたIDの検査情報を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Param id path string true "検査ID (UUID)"
// @Success 200 {object} model.Examination
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [get]
func (h *Handler) GetExaminationByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	exam, err := h.svc.GetExaminationByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "examination", id)
		return
	}

	c.JSON(http.StatusOK, exam)
}

// GetExaminationsByPetID godoc
// @Summary ペットの検査一覧取得
// @Description 指定されたペットIDの検査一覧を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Param petId path string true "ペットID (UUID)"
// @Success 200 {array} model.Examination
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{petId}/examinations [get]
func (h *Handler) GetExaminationsByPetID(c *gin.Context) {
	ctx := c.Request.Context()
	petID := c.Param("petId")

	exams, err := h.svc.GetExaminationsByPetID(ctx, petID)
	if err != nil {
		h.handleError(c, err, "examination", petID)
		return
	}

	c.JSON(http.StatusOK, exams)
}

// GetExaminationsByOwnerID godoc
// @Summary 飼い主の検査一覧取得
// @Description 指定された飼い主IDの検査一覧を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Param ownerId path string true "飼い主ID (UUID)"
// @Success 200 {array} model.Examination
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{ownerId}/examinations [get]
func (h *Handler) GetExaminationsByOwnerID(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := c.Param("ownerId")

	exams, err := h.svc.GetExaminationsByOwnerID(ctx, ownerID)
	if err != nil {
		h.handleError(c, err, "examination", ownerID)
		return
	}

	c.JSON(http.StatusOK, exams)
}

// GetExaminationsByStatus godoc
// @Summary ステータス別検査取得
// @Description 指定されたステータスの検査一覧を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Param status path string true "ステータス (依頼中,検査中,完了)"
// @Success 200 {array} model.Examination
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/status/{status} [get]
func (h *Handler) GetExaminationsByStatus(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Param("status")

	exams, err := h.svc.GetExaminationsByStatus(ctx, status)
	if err != nil {
		h.handleError(c, err, "examination", status)
		return
	}

	c.JSON(http.StatusOK, exams)
}

// CreateExamination godoc
// @Summary 検査作成
// @Description 新しい検査を作成します
// @Tags examinations
// @Accept json
// @Produce json
// @Param exam body model.CreateExaminationRequest true "検査情報"
// @Success 201 {object} model.Examination
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations [post]
func (h *Handler) CreateExamination(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateExaminationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	exam, err := h.svc.CreateExamination(ctx, &req)
	if err != nil {
		h.handleError(c, err, "examination", "")
		return
	}

	c.JSON(http.StatusCreated, exam)
}

// UpdateExamination godoc
// @Summary 検査更新
// @Description 既存の検査を更新します
// @Tags examinations
// @Accept json
// @Produce json
// @Param id path string true "検査ID (UUID)"
// @Param exam body model.UpdateExaminationRequest true "検査情報"
// @Success 200 {object} model.Examination
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [put]
func (h *Handler) UpdateExamination(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateExaminationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	exam, err := h.svc.UpdateExamination(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "examination", id)
		return
	}

	c.JSON(http.StatusOK, exam)
}

// DeleteExamination godoc
// @Summary 検査削除
// @Description 指定された検査を削除します
// @Tags examinations
// @Accept json
// @Produce json
// @Param id path string true "検査ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [delete]
func (h *Handler) DeleteExamination(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteExamination(ctx, id); err != nil {
		h.handleError(c, err, "examination", id)
		return
	}

	c.Status(http.StatusNoContent)
}
