package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetLstepMonthlyDeliveryStats godoc
// GET /api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats?year_month=YYYY-MM — 月次配信集計（FEAT-385）。
func (h *Handler) GetLstepMonthlyDeliveryStats(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	query, err := newLstepMonthlyDeliveryStatsQuery(c.Request.URL.Query())
	if err != nil {
		RespondError(c, err)
		return
	}
	stats, err := h.svc.LstepAnalytics.GetMonthlyDeliveryStats(c.Request.Context(), clinicID, query.YearMonth)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepMonthlyDeliveryStatsResponse(stats))
}

// GetLstepVisitConversionSummary godoc
// GET /api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion?year_month=YYYY-MM&days=30 — 月次配信後来院率集計。
func (h *Handler) GetLstepVisitConversionSummary(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	query, err := newLstepVisitConversionQuery(c.Request.URL.Query())
	if err != nil {
		RespondError(c, err)
		return
	}
	days, err := query.toDays()
	if err != nil {
		RespondError(c, err)
		return
	}
	stats, err := h.svc.LstepAnalytics.GetVisitConversionSummary(c.Request.Context(), clinicID, query.YearMonth, days)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepVisitConversionResponse(stats))
}

// GetLstepOwnerFriendAttributes godoc
// GET /api/v1/clinics/:clinic_id/owners/:id/lstep/friend-attributes — 飼主の最新 Lステップ友だち属性（FEAT-385）。
func (h *Handler) GetLstepOwnerFriendAttributes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	snapshot, err := h.svc.LstepAnalytics.GetLatestFriendAttributes(c.Request.Context(), clinicID, ownerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepFriendAttributeResponse(snapshot))
}

// RegisterLstepAnalyticsRoutes は FEAT-385 分析のルートを登録する。
func (h *Handler) RegisterLstepAnalyticsRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/clinics/:clinic_id/lstep/analytics")
	g.GET("/delivery-stats", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepMonthlyDeliveryStats)
	g.GET("/visit-conversion", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepVisitConversionSummary)
}
