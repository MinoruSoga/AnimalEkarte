package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock LstepAnalyticsService ----

type mockLstepAnalyticsService struct {
	getDeliveryHistoryByOwnerFn func(ctx context.Context, clinicID, ownerID uint64, from, to time.Time) ([]model.LstepDeliveryTriggerLog, error)
	getMonthlyDeliveryStatsFn   func(ctx context.Context, clinicID uint64, yearMonth string) (*service.MonthlyDeliveryStats, error)
	getLatestFriendAttributesFn func(ctx context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error)
}

func (m *mockLstepAnalyticsService) GetDeliveryHistoryByOwner(ctx context.Context, clinicID, ownerID uint64, from, to time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	if m.getDeliveryHistoryByOwnerFn != nil {
		return m.getDeliveryHistoryByOwnerFn(ctx, clinicID, ownerID, from, to)
	}
	return nil, nil
}

func (m *mockLstepAnalyticsService) GetMonthlyDeliveryStats(ctx context.Context, clinicID uint64, yearMonth string) (*service.MonthlyDeliveryStats, error) {
	if m.getMonthlyDeliveryStatsFn != nil {
		return m.getMonthlyDeliveryStatsFn(ctx, clinicID, yearMonth)
	}
	return &service.MonthlyDeliveryStats{YearMonth: yearMonth}, nil
}

func (m *mockLstepAnalyticsService) GetLatestFriendAttributes(ctx context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error) {
	if m.getLatestFriendAttributesFn != nil {
		return m.getLatestFriendAttributesFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

// ---- router helpers ----

func newGetDeliveryStatsRouter(analyticsSvc service.LstepAnalyticsService, permSvc service.EffectivePermissionService, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{svc: &service.Services{LstepAnalytics: analyticsSvc, EffectivePermission: permSvc}}
	r := gin.New()
	r.GET("/clinics/:clinic_id/lstep/analytics/delivery-stats",
		setupCtx,
		h.RequirePermission(string(model.ResourceLstepAnalytics), "view"),
		h.GetLstepMonthlyDeliveryStats,
	)
	return r
}

func newGetOwnerFriendAttributesRouter(analyticsSvc service.LstepAnalyticsService, permSvc service.EffectivePermissionService, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{svc: &service.Services{LstepAnalytics: analyticsSvc, EffectivePermission: permSvc}}
	r := gin.New()
	r.GET("/clinics/:clinic_id/owners/:id/lstep/friend-attributes",
		setupCtx,
		h.RequirePermission(string(model.ResourceLstepAnalytics), "view"),
		h.GetLstepOwnerFriendAttributes,
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
