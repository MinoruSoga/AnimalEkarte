package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

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
