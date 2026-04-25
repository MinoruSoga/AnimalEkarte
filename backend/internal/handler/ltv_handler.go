package handler

import (
	"net/http"
	"strconv"
	"strings"

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
	HasLine          bool    `json:"has_line"`
	TotalAmount      int64   `json:"total_amount"`
	TotalFee         int64   `json:"total_fee"` // total_amount の FE エイリアス
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
		HasLine:          item.HasLine,
		TotalAmount:      item.TotalAmount,
		TotalFee:         item.TotalAmount,
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

// resolveLtvSort は FE の sort + order パラメータをリポジトリの sort 定数に変換する。
func resolveLtvSort(sortParam, orderParam string) string {
	order := strings.ToLower(orderParam)
	if order != "asc" {
		order = "desc"
	}
	switch sortParam {
	case "total_fee", "total_amount":
		return "total_amount_" + order
	case "annual_visit_count", "visit_count":
		return "visit_count_" + order
	case "last_visit_date", "last_visit":
		return "last_visit_" + order
	default:
		return "total_amount_desc"
	}
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

	// min_total_fee / max_total_fee は FE エイリアス（min_total_amount 優先）
	var minTotalAmount, maxTotalAmount *int64
	minAmountKey := "min_total_amount"
	if c.Query(minAmountKey) == "" && c.Query("min_total_fee") != "" {
		minAmountKey = "min_total_fee"
	}
	if s := c.Query(minAmountKey); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("min_total_amount / min_total_fee は0以上の整数で指定してください"))
			return
		}
		minTotalAmount = &v
	}
	maxAmountKey := "max_total_amount"
	if c.Query(maxAmountKey) == "" && c.Query("max_total_fee") != "" {
		maxAmountKey = "max_total_fee"
	}
	if s := c.Query(maxAmountKey); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("max_total_amount / max_total_fee は0以上の整数で指定してください"))
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

	// has_line は FE エイリアス（line_linked 優先）
	lineLinked := c.Query("line_linked") == "true" || c.Query("has_line") == "true"

	// sort + order パラメータ（FE: sort=total_fee&order=desc）
	sortStr := resolveLtvSort(c.Query("sort"), c.Query("order"))

	result, err := h.svc.Ltv.ListOwnerLTV(c.Request.Context(), clinicID, service.ListOwnerLTVInput{
		Sort:           sortStr,
		MinTotalAmount: minTotalAmount,
		MaxTotalAmount: maxTotalAmount,
		MinVisitCount:  minVisitCount,
		CPMStageFilter: c.Query("cpm_stage"),
		LineLinked:     lineLinked,
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

// bulkLstepTagRequest は一括タグ付与リクエスト。
type bulkLstepTagRequest struct {
	OwnerIDs []string `json:"owner_ids" binding:"required,min=1"`
	TagName  string   `json:"tag_name"  binding:"required"`
}

// bulkLstepTagResponse は一括タグ付与レスポンス。
type bulkLstepTagResponse struct {
	SyncedCount    int      `json:"synced_count"`
	SkippedCount   int      `json:"skipped_count"`
	FailedOwnerIDs []string `json:"failed_owner_ids"`
}

// BulkLstepTag godoc
// POST /api/v1/owners/lstep/bulk-tags — 選択した飼い主に一括でタグを付与する。
func (h *Handler) BulkLstepTag(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req bulkLstepTagRequest
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
	result, err := h.svc.LstepTag.BulkAddOwnerTag(c.Request.Context(), clinicID, ownerIDs, req.TagName, actorID)
	if err != nil {
		RespondError(c, err)
		return
	}

	failedStrings := make([]string, 0, len(result.FailedOwnerIDs))
	for _, id := range result.FailedOwnerIDs {
		failedStrings = append(failedStrings, strconv.FormatUint(id, 10))
	}

	c.JSON(http.StatusOK, bulkLstepTagResponse{
		SyncedCount:    result.SyncedCount,
		SkippedCount:   result.SkippedCount,
		FailedOwnerIDs: failedStrings,
	})
}

// RegisterLtvRoutes はLTV関連のルートを登録する（BE-010）。
func (h *Handler) RegisterLtvRoutes(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("/ltv", h.ListOwnerLTV)
	owners.POST("/lstep/bulk-tags", h.RequirePermission(string(model.ResourceOwners), "edit"), h.BulkLstepTag)

	lstep := rg.Group("/lstep")
	lstep.POST("/ltv-sync", h.RequirePermission(string(model.ResourceOwners), "edit"), h.SyncLtvTags)

	// ISSUE-005: FE統一エンドポイントのエイリアス
	clinics := rg.Group("/clinics/:clinic_id")
	clinics.GET("/owners/ltv", h.ListOwnerLTV)
	clinics.POST("/owners/lstep/bulk-tags", h.RequirePermission(string(model.ResourceOwners), "edit"), h.BulkLstepTag)
}
