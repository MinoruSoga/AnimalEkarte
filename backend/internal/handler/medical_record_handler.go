package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListMedicalRecords godoc
// @Summary 診療記録一覧取得
// @Description クリニックの診療記録一覧をページネーション付きで取得する
// @Tags MedicalRecords
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param pet_id query string false "ペットUUID"
// @Success 200 {object} handler.PaginatedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records [get]
func (h *Handler) ListMedicalRecords(c *gin.Context) {
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
	if petIDStr := c.Query("pet_id"); petIDStr != "" {
		id, err := uuid.Parse(petIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id"})
			return
		}
		petID = &id
	}

	records, total, err := h.svc.MedicalRecord.List(c.Request.Context(), clinicID, petID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: records, Total: total, Page: page, Limit: limit})
}

// GetMedicalRecord godoc
// @Summary 診療記録取得
// @Description 指定IDの診療記録を取得する
// @Tags MedicalRecords
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "診療記録UUID"
// @Success 200 {object} model.MedicalRecord
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records/{id} [get]
func (h *Handler) GetMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	record, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

// CreateMedicalRecord godoc
// @Summary 診療記録作成
// @Description 新しい診療記録を作成する
// @Tags MedicalRecords
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.MedicalRecord true "診療記録情報"
// @Success 201 {object} model.MedicalRecord
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records [post]
func (h *Handler) CreateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input model.MedicalRecord
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	input.ClinicID = clinicID
	if err := h.svc.MedicalRecord.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateMedicalRecord godoc
// @Summary 診療記録更新
// @Description 指定IDの診療記録を更新する
// @Tags MedicalRecords
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "診療記録UUID"
// @Param body body model.MedicalRecord true "診療記録情報"
// @Success 200 {object} model.MedicalRecord
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records/{id} [put]
func (h *Handler) UpdateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.MedicalRecord
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	input.ClinicID = clinicID
	if err := h.svc.MedicalRecord.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteMedicalRecord godoc
// @Summary 診療記録削除
// @Description 指定IDの診療記録を削除する
// @Tags MedicalRecords
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "診療記録UUID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records/{id} [delete]
func (h *Handler) DeleteMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.MedicalRecord.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
