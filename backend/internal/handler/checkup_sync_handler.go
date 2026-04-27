package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// checkupSyncPreviewOwnerResponse はプレビュー一覧の1件レスポンス。
type checkupSyncPreviewOwnerResponse struct {
	OwnerID       string   `json:"owner_id"`
	OwnerName     string   `json:"owner_name"`
	PetNames      []string `json:"pet_names"`
	LastVisitDate *string  `json:"last_visit_date"`
	HasLine       bool     `json:"has_line"`
	IsOptOut      bool     `json:"is_opt_out"`
	CurrentTags   []string `json:"current_tags"`
}

// checkupSyncPreviewResponse はプレビューレスポンス。
type checkupSyncPreviewResponse struct {
	Owners          []checkupSyncPreviewOwnerResponse `json:"owners"`
	LineLinkedCount int                               `json:"line_linked_count"`
	TotalCount      int                               `json:"total_count"`
}

// checkupSyncRequest は一括タグ付与リクエスト。
type checkupSyncRequest struct {
	CheckupType string   `json:"checkup_type" binding:"required"`
	OwnerIDs    []string `json:"owner_ids"    binding:"required,min=1"`
	TagName     string   `json:"tag_name"     binding:"required"`
}

// checkupSyncResultResponse は一括タグ付与レスポンス。
type checkupSyncResultResponse struct {
	SuccessCount   int      `json:"success_count"`
	SkippedCount   int      `json:"skipped_count"`
	FailedCount    int      `json:"failed_count"`
	FailedOwnerIDs []string `json:"failed_owner_ids"`
}

func toCheckupSyncPreviewOwnerResponse(o *service.CheckupSyncPreviewOwner) checkupSyncPreviewOwnerResponse {
	r := checkupSyncPreviewOwnerResponse{
		OwnerID:     strconv.FormatUint(o.OwnerID, 10),
		OwnerName:   o.OwnerName,
		PetNames:    o.PetNames,
		HasLine:     o.HasLine,
		IsOptOut:    o.IsOptOut,
		CurrentTags: o.CurrentTags,
	}
	if o.LastVisitDate != nil {
		s := o.LastVisitDate.Format("2006-01-02")
		r.LastVisitDate = &s
	}
	return r
}

func toCheckupSyncPreviewResponse(result *service.PreviewCheckupSyncResult) checkupSyncPreviewResponse {
	owners := make([]checkupSyncPreviewOwnerResponse, 0, len(result.Owners))
	for i := range result.Owners {
		owners = append(owners, toCheckupSyncPreviewOwnerResponse(&result.Owners[i]))
	}
	return checkupSyncPreviewResponse{
		Owners:          owners,
		LineLinkedCount: result.LineLinkedCount,
		TotalCount:      result.TotalCount,
	}
}

// GetCheckupSyncPreview godoc
// GET /clinics/:clinic_id/lstep/checkup-sync/preview — 健診対象者プレビューを返す（BE-004）。
func (h *Handler) GetCheckupSyncPreview(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	input := service.PreviewCheckupSyncInput{
		CheckupType: c.Query("checkup_type"),
		Species:     c.Query("species"),
	}

	if s := c.Query("last_visit_before"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("last_visit_before は YYYY-MM-DD 形式で指定してください"))
			return
		}
		input.LastVisitBefore = &t
	}
	if s := c.Query("last_visit_after"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("last_visit_after は YYYY-MM-DD 形式で指定してください"))
			return
		}
		input.LastVisitAfter = &t
	}

	result, err := h.svc.CheckupSync.PreviewCheckupSync(c.Request.Context(), clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupSyncPreviewResponse(result))
}

// CreateCheckupSync godoc
// POST /clinics/:clinic_id/lstep/checkup-sync — 選択した飼い主に健診リマインダータグを一括付与する（BE-004）。
func (h *Handler) CreateCheckupSync(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req checkupSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	if !tagNamePattern.MatchString(req.TagName) {
		RespondError(c, apperrors.WrapInvalidInput("tag_name は英数字・アンダースコア・ハイフンのみ使用可能です（1〜100文字）"))
		return
	}

	ownerIDs := make([]uint64, 0, len(req.OwnerIDs))
	for _, s := range req.OwnerIDs {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("owner_ids の値が不正です: "+s))
			return
		}
		ownerIDs = append(ownerIDs, id)
	}

	var actorID *uint64
	if staffID, ok := extractStaffID(c); ok {
		actorID = &staffID
	}

	result, err := h.svc.CheckupSync.CreateCheckupSync(c.Request.Context(), clinicID, service.CreateCheckupSyncInput{
		CheckupType: req.CheckupType,
		OwnerIDs:    ownerIDs,
		TagName:     req.TagName,
	}, actorID)
	if err != nil {
		RespondError(c, err)
		return
	}

	failedStrings := make([]string, 0, len(result.FailedOwnerIDs))
	for _, id := range result.FailedOwnerIDs {
		failedStrings = append(failedStrings, strconv.FormatUint(id, 10))
	}

	c.JSON(http.StatusOK, checkupSyncResultResponse{
		SuccessCount:   result.SuccessCount,
		SkippedCount:   result.SkippedCount,
		FailedCount:    result.FailedCount,
		FailedOwnerIDs: failedStrings,
	})
}

// RegisterCheckupSyncRoutes は健診同期関連ルートを登録する（BE-004）。
func (h *Handler) RegisterCheckupSyncRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics/:clinic_id")
	clinics.GET("/lstep/checkup-sync/preview", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetCheckupSyncPreview)
	clinics.POST("/lstep/checkup-sync", h.RequirePermission(string(model.ResourceOwners), "edit"), h.CreateCheckupSync)
}
