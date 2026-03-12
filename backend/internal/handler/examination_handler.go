package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListExaminations godoc
// @Summary 検査一覧取得
// @Description 検査記録の一覧をページネーション付きで取得する
// @Tags Examinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param pet_id query string false "ペットIDフィルター"
// @Param owner_id query string false "飼主IDフィルター"
// @Param status query string false "ステータスフィルター"
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations [get]
func (h *Handler) ListExaminations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var petID *uuid.UUID
	if s := c.Query("pet_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id"})
			return
		}
		petID = &id
	}

	var ownerID *uuid.UUID
	if s := c.Query("owner_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
		ownerID = &id
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	exams, total, err := h.svc.Examination.List(c.Request.Context(), clinicID, petID, ownerID, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: exams, Total: total, Page: page, Limit: limit})
}

// GetExamination godoc
// @Summary 検査詳細取得
// @Description 指定IDの検査記録を取得する
// @Tags Examinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "検査ID"
// @Success 200 {object} model.Exam
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [get]
func (h *Handler) GetExamination(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	exam, err := h.svc.Examination.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, exam)
}

// CreateExamination godoc
// @Summary 検査作成
// @Description 新しい検査記録を作成する
// @Tags Examinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body model.Exam true "検査情報"
// @Success 201 {object} model.Exam
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations [post]
func (h *Handler) CreateExamination(c *gin.Context) {
	var input model.Exam
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Examination.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateExamination godoc
// @Summary 検査更新
// @Description 指定IDの検査記録を更新する
// @Tags Examinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "検査ID"
// @Param input body model.Exam true "更新する検査情報"
// @Success 200 {object} model.Exam
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [patch]
func (h *Handler) UpdateExamination(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Exam
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Examination.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteExamination godoc
// @Summary 検査削除
// @Description 指定IDの検査記録を削除する
// @Tags Examinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "検査ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /examinations/{id} [delete]
func (h *Handler) DeleteExamination(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Examination.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
