package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllExaminations godoc
// @Summary 検査記録一覧取得
// @Description 登録されている検査記録の一覧を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Success 200 {array} model.ExaminationRecord
// @Failure 500 {object} map[string]string
// @Router /examinations [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllExaminations(c *gin.Context) {
	ctx := c.Request.Context()

	exams, err := h.svc.GetAllExaminationRecords(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get examination records", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, exams)
}

// GetExaminationByID godoc
// @Summary 検査記録詳細取得
// @Description 指定されたIDの検査記録情報を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Param id path string true "検査記録ID (UUID)"
// @Success 200 {object} model.ExaminationRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [get]
func (h *Handler) GetExaminationByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	exam, err := h.svc.GetExaminationRecordByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "examination", id)
		return
	}

	c.JSON(http.StatusOK, exam)
}

// GetExaminationsByPetID godoc
// @Summary ペットの検査記録一覧取得
// @Description 指定されたペットIDの検査記録一覧を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Param petId path string true "ペットID (UUID)"
// @Success 200 {array} model.ExaminationRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{petId}/examinations [get]
func (h *Handler) GetExaminationsByPetID(c *gin.Context) {
	ctx := c.Request.Context()
	petID := c.Param("id")

	exams, err := h.svc.GetExaminationRecordsByPetID(ctx, petID)
	if err != nil {
		h.handleError(c, err, "examination", petID)
		return
	}

	c.JSON(http.StatusOK, exams)
}

// GetExaminationsByOwnerID godoc
// @Summary 飼い主の検査記録一覧取得
// @Description 指定された飼い主IDに関連する検査記録一覧を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Param ownerId path string true "飼い主ID (UUID)"
// @Success 200 {array} model.ExaminationRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{ownerId}/examinations [get]
func (h *Handler) GetExaminationsByOwnerID(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := c.Param("id")

	exams, err := h.svc.GetExaminationRecordsByOwnerID(ctx, ownerID)
	if err != nil {
		h.handleError(c, err, "examination", ownerID)
		return
	}

	c.JSON(http.StatusOK, exams)
}

// GetExaminationsByStatus godoc
// @Summary ステータス別検査記録取得
// @Description 指定されたステータスの検査記録一覧を取得します
// @Tags examinations
// @Accept json
// @Produce json
// @Param status path string true "ステータス (依頼中,検査中,完了)"
// @Success 200 {array} model.ExaminationRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/status/{status} [get]
func (h *Handler) GetExaminationsByStatus(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Param("status")

	exams, err := h.svc.GetAllExaminationRecords(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get examination records", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Filter by status in memory (examination_records has no direct status filter in repo)
	var filtered []model.ExaminationRecord
	for _, e := range exams {
		if e.Status == status {
			filtered = append(filtered, e)
		}
	}
	if filtered == nil {
		filtered = []model.ExaminationRecord{}
	}
	c.JSON(http.StatusOK, filtered)
}

// CreateExamination godoc
// @Summary 検査記録作成
// @Description 新しい検査記録を作成します
// @Tags examinations
// @Accept json
// @Produce json
// @Param exam body model.CreateExaminationRecordRequest true "検査記録情報"
// @Success 201 {object} model.ExaminationRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations [post]
func (h *Handler) CreateExamination(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateExaminationRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	exam, err := h.svc.CreateExaminationRecord(ctx, &req)
	if err != nil {
		h.handleError(c, err, "examination", "")
		return
	}

	c.JSON(http.StatusCreated, exam)
}

// UpdateExamination godoc
// @Summary 検査記録更新
// @Description 既存の検査記録を更新します
// @Tags examinations
// @Accept json
// @Produce json
// @Param id path string true "検査記録ID (UUID)"
// @Param exam body model.UpdateExaminationRecordRequest true "検査記録情報"
// @Success 200 {object} model.ExaminationRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [put]
func (h *Handler) UpdateExamination(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateExaminationRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	exam, err := h.svc.UpdateExaminationRecord(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "examination", id)
		return
	}

	c.JSON(http.StatusOK, exam)
}

// DeleteExamination godoc
// @Summary 検査記録削除
// @Description 指定された検査記録を削除します
// @Tags examinations
// @Accept json
// @Produce json
// @Param id path string true "検査記録ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [delete]
func (h *Handler) DeleteExamination(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteExaminationRecord(ctx, id); err != nil {
		h.handleError(c, err, "examination", id)
		return
	}

	c.Status(http.StatusNoContent)
}
