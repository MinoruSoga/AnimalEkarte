package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// tagSummaryItemResponse はタグ集計1件のJSONレスポンス。
type tagSummaryItemResponse struct {
	TagName    string `json:"tag_name"`
	OwnerCount int64  `json:"owner_count"`
	Category   string `json:"category"`
}

// tagSummaryResponse は GET /lstep/tag-summary のJSONレスポンス。
type tagSummaryResponse struct {
	Tags                 []tagSummaryItemResponse `json:"tags"`
	TotalOwnersWithLstep int64                    `json:"total_owners_with_lstep"`
	AsOf                 string                   `json:"as_of"`
}

// tagOwnerItemResponse は GET /lstep/owners 1件のJSONレスポンス。
type tagOwnerItemResponse struct {
	OwnerID          string   `json:"owner_id"`
	OwnerName        string   `json:"owner_name"`
	LineUserIDMasked *string  `json:"line_user_id_masked,omitempty"`
	LastVisitDate    *string  `json:"last_visit_date,omitempty"`
	AllTags          []string `json:"all_tags"`
	Reason           *string  `json:"reason,omitempty"`
}

// tagOwnerListResponse は GET /lstep/owners のJSONレスポンス。
type tagOwnerListResponse struct {
	Owners  []tagOwnerItemResponse `json:"owners"`
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	PerPage int                    `json:"per_page"`
}

func toTagSummaryResponse(r service.TagSummaryResponse) tagSummaryResponse {
	items := make([]tagSummaryItemResponse, len(r.Tags))
	for i, t := range r.Tags {
		items[i] = tagSummaryItemResponse{TagName: t.TagName, OwnerCount: t.OwnerCount, Category: t.Category}
	}
	return tagSummaryResponse{
		Tags:                 items,
		TotalOwnersWithLstep: r.TotalOwnersWithLstep,
		AsOf:                 r.AsOf.In(time.Local).Format(time.RFC3339),
	}
}

func toTagOwnerListResponse(r service.TagOwnerListResponse) tagOwnerListResponse {
	owners := make([]tagOwnerItemResponse, len(r.Owners))
	for i, o := range r.Owners {
		tags := o.AllTags
		if tags == nil {
			tags = []string{}
		}
		owners[i] = tagOwnerItemResponse{
			OwnerID:       strconv.FormatUint(o.OwnerID, 10),
			OwnerName:     o.OwnerName,
			LastVisitDate: o.LastVisitDate,
			AllTags:       tags,
			Reason:        o.Reason,
		}
	}
	return tagOwnerListResponse{Owners: owners, Total: r.Total, Page: r.Page, PerPage: r.PerPage}
}

// GetLstepTagSummary godoc
// GET /api/clinics/:clinic_id/lstep/tag-summary — タグ別飼い主数集計を返す（BE-020）。
func (h *Handler) GetLstepTagSummary(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	result, err := h.svc.LstepTagSummary.GetTagSummary(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTagSummaryResponse(result))
}

// SearchLstepOwnersByTag godoc
// GET /api/clinics/:clinic_id/lstep/owners — タグで飼い主を絞り込む（BE-020）。
// ?format=csv でCSVダウンロード。
func (h *Handler) SearchLstepOwnersByTag(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	query, err := newLstepOwnersByTagQuery(c.Request.URL.Query())
	if err != nil {
		RespondError(c, err)
		return
	}
	if query.isCSV() {
		date := time.Now().In(time.Local).Format(time.DateOnly)
		filename := fmt.Sprintf("lstep-%s-%s.csv", query.TagName, date)
		c.Header("Content-Type", "text/csv; charset=utf-8")
		// RFC 5987: 日本語タグ名を含むファイル名の文字化けを防ぐ（#179 ③）。
		// 非対応クライアント用に ASCII フォールバック(filename=)、対応クライアント用に
		// UTF-8 パーセントエンコード版(filename*=UTF-8'') を併記する。
		asciiFallback := fmt.Sprintf("lstep-%s.csv", date)
		// filename= は HTTP quoted-string。asciiFallback は純 ASCII（日付のみ）なので
		// %q による Go クオートは二重引用符の付与と等価で、HTTP quoted-string として安全。
		c.Header("Content-Disposition", fmt.Sprintf(
			"attachment; filename=%q; filename*=UTF-8''%s",
			asciiFallback, url.PathEscape(filename),
		))
		if err := h.svc.LstepTagSummary.ExportOwnersByTagCSV(c.Request.Context(), clinicID, query.TagName, query.NameQuery, c.Writer); err != nil {
			RespondError(c, err)
		}
		return
	}

	result, err := h.svc.LstepTagSummary.ListOwnersByTag(c.Request.Context(), clinicID, query.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTagOwnerListResponse(result))
}

// RegisterLstepTagSummaryRoutes は BE-020 のルートを登録する。
func (h *Handler) RegisterLstepTagSummaryRoutes(rg *gin.RouterGroup) {
	lstep := rg.Group("/lstep")
	lstep.GET("/tag-summary", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepTagSummary)
	lstep.GET("/owners", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.SearchLstepOwnersByTag)

	// ISSUE-020: FE が /clinics/:clinic_id/lstep/... で呼ぶエイリアス
	clinicLstep := rg.Group("/clinics/:clinic_id/lstep")
	clinicLstep.GET("/tag-summary", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepTagSummary)
	clinicLstep.GET("/owners", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.SearchLstepOwnersByTag)
}
