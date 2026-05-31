package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// checkupSyncPreviewOwnerResponse はプレビュー一覧の1件レスポンス（ISSUE-005: 除外理由対応 / ISSUE-009: 追加フィルタ表示）。
type checkupSyncPreviewOwnerResponse struct {
	OwnerID         string   `json:"owner_id"`
	OwnerName       string   `json:"owner_name"`
	PetNames        []string `json:"pet_names"`
	LastVisitDate   *string  `json:"last_visit_date"`
	HasLine         bool     `json:"has_line"`
	IsOptOut        bool     `json:"is_opt_out"`
	HasLivingPet    bool     `json:"has_living_pet"`
	ExclusionReason *string  `json:"exclusion_reason"`
	CurrentTags     []string `json:"current_tags"`

	// ISSUE-009: フィルタ確認用のメタ情報（additive）
	MinPetAgeYears      *int    `json:"min_pet_age_years"`
	MaxPetAgeYears      *int    `json:"max_pet_age_years"`
	HasChronicCondition bool    `json:"has_chronic_condition"`
	CPMStage            string  `json:"cpm_stage"`
	TotalAmount         int64   `json:"total_amount"`
	AnnualVisitCount    int64   `json:"annual_visit_count"`
	LastCheckupDate     *string `json:"last_checkup_date"`
}

// checkupSyncPreviewResponse はプレビューレスポンス（ISSUE-005: 除外サマリー対応）。
type checkupSyncPreviewResponse struct {
	Owners           []checkupSyncPreviewOwnerResponse `json:"owners"`
	TotalCount       int                               `json:"total_count"`
	EligibleCount    int                               `json:"eligible_count"`
	LineLinkedCount  int                               `json:"line_linked_count"`
	OptOutCount      int                               `json:"opt_out_count"`
	NoLivingPetCount int                               `json:"no_living_pet_count"`
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
		OwnerID:             strconv.FormatUint(o.OwnerID, 10),
		OwnerName:           o.OwnerName,
		PetNames:            o.PetNames,
		HasLine:             o.HasLine,
		IsOptOut:            o.IsOptOut,
		HasLivingPet:        o.HasLivingPet,
		ExclusionReason:     o.ExclusionReason,
		CurrentTags:         o.CurrentTags,
		MinPetAgeYears:      o.MinPetAgeYears,
		MaxPetAgeYears:      o.MaxPetAgeYears,
		HasChronicCondition: o.HasChronicCondition,
		CPMStage:            o.CPMStage,
		TotalAmount:         o.TotalAmount,
		AnnualVisitCount:    o.AnnualVisitCount,
	}
	if o.LastVisitDate != nil {
		s := o.LastVisitDate.Format("2006-01-02")
		r.LastVisitDate = &s
	}
	if o.LastCheckupDate != nil {
		s := o.LastCheckupDate.Format("2006-01-02")
		r.LastCheckupDate = &s
	}
	return r
}

func toCheckupSyncPreviewResponse(result *service.PreviewCheckupSyncResult) checkupSyncPreviewResponse {
	owners := make([]checkupSyncPreviewOwnerResponse, 0, len(result.Owners))
	for i := range result.Owners {
		owners = append(owners, toCheckupSyncPreviewOwnerResponse(&result.Owners[i]))
	}
	return checkupSyncPreviewResponse{
		Owners:           owners,
		TotalCount:       result.TotalCount,
		EligibleCount:    result.EligibleCount,
		LineLinkedCount:  result.LineLinkedCount,
		OptOutCount:      result.OptOutCount,
		NoLivingPetCount: result.NoLivingPetCount,
	}
}

// GetCheckupSyncPreview godoc
// GET /clinics/:clinic_id/lstep/checkup-sync/preview — 健診対象者プレビューを返す（BE-004 / ISSUE-009）。
func (h *Handler) GetCheckupSyncPreview(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	q := newCheckupSyncPreviewQuery(c.Request.URL.Query())
	input, err := q.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	// ISSUE-010: 抽出メタデータを audit_logs に永続化するため actorID を service に渡す。
	var actorID *uint64
	if staffID, ok := extractStaffID(c); ok {
		actorID = &staffID
	}

	result, err := h.svc.CheckupSync.PreviewCheckupSync(c.Request.Context(), clinicID, input, actorID)
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

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	var actorID *uint64
	if staffID, ok := extractStaffID(c); ok {
		actorID = &staffID
	}

	result, err := h.svc.CheckupSync.CreateCheckupSync(c.Request.Context(), clinicID, input, actorID)
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
