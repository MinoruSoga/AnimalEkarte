package lstep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock LstepAnalyticsService ----

type mockLstepAnalyticsService struct {
	getMonthlyDeliveryStatsFn   func(ctx context.Context, clinicID uint64, yearMonth string) (*MonthlyDeliveryStats, error)
	getVisitConversionFn        func(ctx context.Context, clinicID uint64, yearMonth string, days int) (*VisitConversionSummary, error)
	getLatestFriendAttributesFn func(ctx context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error)
}

func (m *mockLstepAnalyticsService) GetMonthlyDeliveryStats(ctx context.Context, clinicID uint64, yearMonth string) (*MonthlyDeliveryStats, error) {
	if m.getMonthlyDeliveryStatsFn != nil {
		return m.getMonthlyDeliveryStatsFn(ctx, clinicID, yearMonth)
	}
	return &MonthlyDeliveryStats{YearMonth: yearMonth}, nil
}

func (m *mockLstepAnalyticsService) GetVisitConversionSummary(ctx context.Context, clinicID uint64, yearMonth string, days int) (*VisitConversionSummary, error) {
	if m.getVisitConversionFn != nil {
		return m.getVisitConversionFn(ctx, clinicID, yearMonth, days)
	}
	return &VisitConversionSummary{YearMonth: yearMonth, Days: days}, nil
}

func (m *mockLstepAnalyticsService) GetLatestFriendAttributes(ctx context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error) {
	if m.getLatestFriendAttributesFn != nil {
		return m.getLatestFriendAttributesFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

// ---- router helpers ----

func newGetDeliveryStatsRouter(analyticsSvc LstepAnalyticsService, _ any, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{analytics: analyticsSvc, requirePermission: testPermissionMiddleware}
	r := gin.New()
	r.GET("/clinics/:clinic_id/lstep/analytics/delivery-stats",
		setupCtx,
		h.requirePermission(string(model.ResourceLstepAnalytics), "view"),
		h.GetLstepMonthlyDeliveryStats,
	)
	return r
}

func newGetOwnerFriendAttributesRouter(analyticsSvc LstepAnalyticsService, _ any, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{analytics: analyticsSvc, requirePermission: testPermissionMiddleware}
	r := gin.New()
	r.GET("/clinics/:clinic_id/owners/:id/lstep/friend-attributes",
		setupCtx,
		h.requirePermission(string(model.ResourceLstepAnalytics), "view"),
		h.GetLstepOwnerFriendAttributes,
	)
	return r
}

func newGetVisitConversionRouter(analyticsSvc LstepAnalyticsService, _ any, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{analytics: analyticsSvc, requirePermission: testPermissionMiddleware}
	r := gin.New()
	r.GET("/clinics/:clinic_id/lstep/analytics/visit-conversion",
		setupCtx,
		h.requirePermission(string(model.ResourceLstepAnalytics), "view"),
		h.GetLstepVisitConversionSummary,
	)
	return r
}

// ---- Case F: GET delivery-stats 403 — 権限なし ----

func TestGetLstepMonthlyDeliveryStats_F_403_NoPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newGetDeliveryStatsRouter(
		&mockLstepAnalyticsService{},
		&mockEffectivePermissionService{},
		func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/analytics/delivery-stats?year_month=2026-04", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---- Case G: GET delivery-stats 200 — 成功 ----

func TestGetLstepMonthlyDeliveryStats_G_200_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newGetDeliveryStatsRouter(
		&mockLstepAnalyticsService{},
		nil,
		func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/analytics/delivery-stats?year_month=2026-04", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- Case H: GET delivery-stats 400 — year_month パラメータなし ----

func TestGetLstepMonthlyDeliveryStats_H_400_MissingYearMonth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newGetDeliveryStatsRouter(
		&mockLstepAnalyticsService{},
		nil,
		func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/analytics/delivery-stats", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetLstepVisitConversionSummary_L_403_NoPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newGetVisitConversionRouter(
		&mockLstepAnalyticsService{},
		&mockEffectivePermissionService{},
		func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/analytics/visit-conversion?year_month=2026-04", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetLstepVisitConversionSummary_M_200_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	analyticsSvc := &mockLstepAnalyticsService{
		getVisitConversionFn: func(_ context.Context, clinicID uint64, yearMonth string, days int) (*VisitConversionSummary, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, "2026-04", yearMonth)
			assert.Equal(t, 30, days)
			return &VisitConversionSummary{YearMonth: yearMonth, Days: days}, nil
		},
	}
	r := newGetVisitConversionRouter(analyticsSvc, nil, func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) })
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/analytics/visit-conversion?year_month=2026-04", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetLstepVisitConversionSummary_N_400_InvalidDays(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newGetVisitConversionRouter(
		&mockLstepAnalyticsService{},
		nil,
		func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/analytics/visit-conversion?year_month=2026-04&days=0", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---- Case I: GET owner friend-attrs 403 — 権限なし ----

func TestGetLstepOwnerFriendAttributes_I_403_NoPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newGetOwnerFriendAttributesRouter(
		&mockLstepAnalyticsService{},
		&mockEffectivePermissionService{},
		func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/owners/42/lstep/friend-attributes", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---- Case J: GET owner friend-attrs 200 — 成功 ----

func TestGetLstepOwnerFriendAttributes_J_200_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	analyticsSvc := &mockLstepAnalyticsService{
		getLatestFriendAttributesFn: func(_ context.Context, _, _ uint64) (*model.LstepFriendAttributeSnapshot, error) {
			return &model.LstepFriendAttributeSnapshot{
				ID:              1,
				ClinicID:        1,
				LineUserID:      "Uabc123",
				SnapshotTakenAt: time.Now(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}, nil
		},
	}
	r := newGetOwnerFriendAttributesRouter(analyticsSvc, nil, func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) })
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/owners/42/lstep/friend-attributes", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- Case K: GET owner friend-attrs 404 — 飼主なし ----

func TestGetLstepOwnerFriendAttributes_K_404_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	analyticsSvc := &mockLstepAnalyticsService{
		getLatestFriendAttributesFn: func(_ context.Context, _, _ uint64) (*model.LstepFriendAttributeSnapshot, error) {
			return nil, apperrors.WrapNotFound("owner", "42")
		},
	}
	r := newGetOwnerFriendAttributesRouter(analyticsSvc, nil, func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) })
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/owners/42/lstep/friend-attributes", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
