package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllMedicalRecords godoc
// @Summary カルテ一覧取得
// @Description 登録されているカルテの一覧を取得します
// @Tags medical-records
// @Accept json
// @Produce json
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "1ページあたりの件数" default(10)
// @Param pet_id query string false "ペットIDでフィルタリング"
// @Param owner_id query string false "飼い主IDでフィルタリング"
// @Param visit_type query string false "診察タイプでフィルタリング（初診/再診）"
// @Param status query string false "ステータスでフィルタリング（作成中/確定済）"
// @Param date_from query string false "診察日（開始）でフィルタリング（YYYY-MM-DD）"
// @Param date_to query string false "診察日（終了）でフィルタリング（YYYY-MM-DD）"
// @Success 200 {array} model.MedicalRecord
// @Failure 500 {object} map[string]string
// @Router /medical-records [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllMedicalRecords(c *gin.Context) {
	ctx := c.Request.Context()

	records, err := h.svc.GetAllMedicalRecords(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get medical records", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, records)
}

// GetMedicalRecord godoc
// @Summary カルテ詳細取得
// @Description 指定されたIDのカルテ情報を取得します
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "カルテID (UUID)"
// @Success 200 {object} model.MedicalRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records/{id} [get]
func (h *Handler) GetMedicalRecord(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	record, err := h.svc.GetMedicalRecordByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "medical_record", id)
		return
	}

	c.JSON(http.StatusOK, record)
}

// GetMedicalRecordsByPetID godoc
// @Summary ペットのカルテ一覧取得
// @Description 指定されたペットIDのカルテ一覧を取得します
// @Tags medical-records
// @Accept json
// @Produce json
// @Param petId path string true "ペットID (UUID)"
// @Success 200 {array} model.MedicalRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{petId}/medical-records [get]
func (h *Handler) GetMedicalRecordsByPetID(c *gin.Context) {
	ctx := c.Request.Context()
	petID := c.Param("petId")

	records, err := h.svc.GetMedicalRecordsByPetID(ctx, petID)
	if err != nil {
		h.handleError(c, err, "medical_record", petID)
		return
	}

	c.JSON(http.StatusOK, records)
}

// GetMedicalRecordsByOwnerID godoc
// @Summary 飼い主のカルテ一覧取得
// @Description 指定された飼い主IDのカルテ一覧を取得します
// @Tags medical-records
// @Accept json
// @Produce json
// @Param ownerId path string true "飼い主ID (UUID)"
// @Success 200 {array} model.MedicalRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{ownerId}/medical-records [get]
func (h *Handler) GetMedicalRecordsByOwnerID(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := c.Param("ownerId")

	records, err := h.svc.GetMedicalRecordsByOwnerID(ctx, ownerID)
	if err != nil {
		h.handleError(c, err, "medical_record", ownerID)
		return
	}

	c.JSON(http.StatusOK, records)
}

// CreateMedicalRecord godoc
// @Summary カルテ作成
// @Description 新しいカルテを作成します
// @Tags medical-records
// @Accept json
// @Produce json
// @Param medicalRecord body model.CreateMedicalRecordRequest true "カルテ情報"
// @Success 201 {object} model.MedicalRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records [post]
func (h *Handler) CreateMedicalRecord(c *gin.Context) {
	ctx := c.Request.Context()

	var req model.CreateMedicalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	record, err := h.svc.CreateMedicalRecord(ctx, &req)
	if err != nil {
		h.handleError(c, err, "medical_record", "")
		return
	}

	slog.InfoContext(ctx, "medical record created", slog.String("record_id", record.ID.String()))
	c.JSON(http.StatusCreated, record)
}

// UpdateMedicalRecord godoc
// @Summary カルテ情報更新
// @Description 指定されたIDのカルテ情報を更新します
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "カルテID (UUID)"
// @Param medicalRecord body model.UpdateMedicalRecordRequest true "更新するカルテ情報"
// @Success 200 {object} model.MedicalRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records/{id} [put]
func (h *Handler) UpdateMedicalRecord(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateMedicalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	record, err := h.svc.UpdateMedicalRecord(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "medical_record", id)
		return
	}

	slog.InfoContext(ctx, "medical record updated", slog.String("record_id", record.ID.String()))
	c.JSON(http.StatusOK, record)
}

// DeleteMedicalRecord godoc
// @Summary カルテ削除
// @Description 指定されたIDのカルテを削除します
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "カルテID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /medical-records/{id} [delete]
func (h *Handler) DeleteMedicalRecord(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	err := h.svc.DeleteMedicalRecord(ctx, id)
	if err != nil {
		h.handleError(c, err, "medical_record", id)
		return
	}

	slog.InfoContext(ctx, "medical record deleted", slog.String("record_id", id))
	c.JSON(http.StatusOK, gin.H{"message": "medical record deleted"})
}

// GetMedicalRecordsWithPagination godoc
// @Summary カルテ一覧取得（ページング）
// @Description カルテ一覧をページングして取得します。フィルタリング機能も利用できます。
// @Tags medical-records
// @Accept json
// @Produce json
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "1ページあたりの件数" default(10)
// @Param pet_id query string false "ペットIDでフィルタリング"
// @Param owner_id query string false "飼い主IDでフィルタリング"
// @Param visit_type query string false "診察タイプでフィルタリング（初診/再診）"
// @Param status query string false "ステータスでフィルタリング（作成中/確定済）"
// @Param date_from query string false "診察日（開始）でフィルタリング（YYYY-MM-DD）"
// @Param date_to query string false "診察日（終了）でフィルタリング（YYYY-MM-DD）"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /medical-records/paginated [get]
// @Security ApiKeyAuth
func (h *Handler) GetMedicalRecordsWithPagination(c *gin.Context) {
	ctx := c.Request.Context()

	// クエリパラメータを取得
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	// 全件取得
	allRecords, err := h.svc.GetAllMedicalRecords(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get medical records", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// ページング計算
	total := len(allRecords)
	totalPages := (total + limit - 1) / limit
	start := (page - 1) * limit
	end := start + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var records []model.MedicalRecord
	if start < total {
		records = allRecords[start:end]
	}

	response := map[string]interface{}{
		"records":      records,
		"current_page": page,
		"per_page":     limit,
		"total":        total,
		"total_pages":  totalPages,
		"has_next":     page < totalPages,
		"has_prev":     page > 1,
	}

	c.JSON(http.StatusOK, response)
}
