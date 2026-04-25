package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ltvOwnerResponse はLTV一覧の1件レスポンス。
type ltvOwnerResponse struct {
	OwnerID          string  `json:"owner_id"`
	OwnerName        string  `json:"owner_name"`
	LineUserIDMasked *string `json:"line_user_id_masked,omitempty"`
	TotalAmount      int64   `json:"total_amount"`
	TotalVisitCount  int64   `json:"total_visit_count"`
	AnnualVisitCount int64   `json:"annual_visit_count"`
	LastVisitDate    *string `json:"last_visit_date,omitempty"`
	FirstVisitDate   *string `json:"first_visit_date,omitempty"`
	CPMStage         string  `json:"cpm_stage"`
}

// ltvListResponse はLTV一覧レスポンス。
type ltvListResponse struct {
	Total   int                `json:"total"`
	Page    int                `json:"page"`
	PerPage int                `json:"per_page"`
	Owners  []ltvOwnerResponse `json:"owners"`
}

// syncLtvTagsRequest はLTV一括タグ同期リクエスト。
type syncLtvTagsRequest struct {
	TagName        string `json:"tag_name"         binding:"required"`
	MinTotalAmount *int64 `json:"min_total_amount"`
	DryRun         bool   `json:"dry_run"`
}

// syncLtvTagsResponse はLTV一括タグ同期レスポンス。
type syncLtvTagsResponse struct {
	DryRun  bool `json:"dry_run"`
	Total   int  `json:"total"`
	Synced  int  `json:"synced"`
	Skipped int  `json:"skipped"`
}

func toLtvOwnerResponse(item service.OwnerLTVItem) ltvOwnerResponse {
	r := ltvOwnerResponse{
		OwnerID:          strconv.FormatUint(item.OwnerID, 10),
		OwnerName:        item.OwnerName,
		LineUserIDMasked: item.LineUserIDMasked,
		TotalAmount:      item.TotalAmount,
		TotalVisitCount:  item.TotalVisitCount,
		AnnualVisitCount: item.AnnualVisitCount,
		CPMStage:         item.CPMStageAPI,
	}
	if item.LastVisitDate != nil {
		s := item.LastVisitDate.Format("2006-01-02")
		r.LastVisitDate = &s
	}
	if item.FirstVisitDate != nil {
		s := item.FirstVisitDate.Format("2006-01-02")
		r.FirstVisitDate = &s
	}
	return r
}

func toLtvListResponse(result *service.ListOwnerLTVResult) ltvListResponse {
	owners := make([]ltvOwnerResponse, 0, len(result.Items))
	for _, item := range result.Items {
		owners = append(owners, toLtvOwnerResponse(item))
	}
	return ltvListResponse{
		Total:   result.Total,
		Page:    result.Page,
		PerPage: result.PerPage,
		Owners:  owners,
	}
}

// ListOwnerLTV godoc
// GET /api/v1/owners/ltv — 累計診療費・来院回数でソート・フィルタ可能なLTV一覧を返す（BE-010）。
func (h *Handler) ListOwnerLTV(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, pageErr := strconv.Atoi(c.DefaultQuery("page", "1"))
	if pageErr != nil || page < 1 {
		RespondError(c, apperrors.WrapInvalidInput("page は1以上の整数で指定してください"))
		return
	}
	perPage, ppErr := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if ppErr != nil || perPage < 1 || perPage > 200 {
		RespondError(c, apperrors.WrapInvalidInput("per_page は1〜200の範囲で指定してください"))
		return
	}

	var minTotalAmount, maxTotalAmount *int64
	if s := c.Query("min_total_amount"); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("min_total_amount は0以上の整数で指定してください"))
			return
		}
		minTotalAmount = &v
	}
	if s := c.Query("max_total_amount"); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("max_total_amount は0以上の整数で指定してください"))
			return
		}
		maxTotalAmount = &v
	}

	var minVisitCount *int64
	if s := c.Query("min_visit_count"); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("min_visit_count は0以上の整数で指定してください"))
			return
		}
		minVisitCount = &v
	}

	result, err := h.svc.Ltv.ListOwnerLTV(c.Request.Context(), clinicID, service.ListOwnerLTVInput{
		Sort:           c.DefaultQuery("sort", "total_amount_desc"),
		MinTotalAmount: minTotalAmount,
		MaxTotalAmount: maxTotalAmount,
		MinVisitCount:  minVisitCount,
		CPMStageFilter: c.Query("cpm_stage"),
		LineLinked:     c.Query("line_linked") == "true",
		Page:           page,
		PerPage:        perPage,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLtvListResponse(result))
}

// SyncLtvTags godoc
// POST /api/v1/lstep/ltv-sync — LTVベースのタグを一括付与する（BE-010）。
func (h *Handler) SyncLtvTags(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req syncLtvTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	if !tagNamePattern.MatchString(req.TagName) {
		RespondError(c, apperrors.WrapInvalidInput("tag_name は英数字・アンダースコア・ハイフンのみ使用可能です（1〜100文字）"))
		return
	}

	result, err := h.svc.Ltv.SyncLtvTags(c.Request.Context(), clinicID, service.SyncLtvTagsInput{
		TagName:        req.TagName,
		MinTotalAmount: req.MinTotalAmount,
		DryRun:         req.DryRun,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, syncLtvTagsResponse{
		DryRun:  result.DryRun,
		Total:   result.Total,
		Synced:  result.Synced,
		Skipped: result.Skipped,
	})
}

// RegisterLtvRoutes はLTV関連のルートを登録する（BE-010）。
func (h *Handler) RegisterLtvRoutes(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("/ltv", h.ListOwnerLTV)

	lstep := rg.Group("/lstep")
	lstep.POST("/ltv-sync", h.RequirePermission(string(model.ResourceOwners), "edit"), h.SyncLtvTags)
}
