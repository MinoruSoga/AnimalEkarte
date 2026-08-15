package lstep

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestRegisterRoutes_L5UsesAuthenticatedClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	type call struct {
		name     string
		clinicID uint64
	}
	var calls []call

	aggregation := &mockAggregationService{listFn: func(_ context.Context, clinicID uint64, _ *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
		calls = append(calls, call{name: "aggregation", clinicID: clinicID})
		return &ListOwnerAggregationResult{Page: 1, PerPage: 50}, nil
	}}
	analytics := &mockLstepAnalyticsService{
		getMonthlyDeliveryStatsFn: func(_ context.Context, clinicID uint64, yearMonth string) (*MonthlyDeliveryStats, error) {
			calls = append(calls, call{name: "delivery_stats", clinicID: clinicID})
			return &MonthlyDeliveryStats{YearMonth: yearMonth}, nil
		},
		getVisitConversionFn: func(_ context.Context, clinicID uint64, yearMonth string, days int) (*VisitConversionSummary, error) {
			calls = append(calls, call{name: "visit_conversion", clinicID: clinicID})
			return &VisitConversionSummary{YearMonth: yearMonth, Days: days}, nil
		},
		getLatestFriendAttributesFn: func(_ context.Context, clinicID, ownerID uint64) (*model.LstepFriendAttributeSnapshot, error) {
			assert.Equal(t, uint64(42), ownerID)
			calls = append(calls, call{name: "friend_attributes", clinicID: clinicID})
			now := time.Now()
			return &model.LstepFriendAttributeSnapshot{ID: 1, ClinicID: clinicID, LineUserID: "U1", SnapshotTakenAt: now, CreatedAt: now, UpdatedAt: now}, nil
		},
	}
	csvImport := &mockLstepCsvImportService{
		importFriendAttributesFn: func(_ context.Context, clinicID uint64, fileName string, _ io.Reader, actorID uint64) (*model.LstepCsvImport, error) {
			assert.Equal(t, "test.csv", fileName)
			assert.Equal(t, uint64(1), actorID)
			calls = append(calls, call{name: "csv_post", clinicID: clinicID})
			return &model.LstepCsvImport{ID: uuid.New(), ClinicID: clinicID, CreatedAt: time.Now()}, nil
		},
		listByClinicFn: func(_ context.Context, clinicID uint64, _ int) ([]*model.LstepCsvImport, error) {
			calls = append(calls, call{name: "csv_get", clinicID: clinicID})
			return nil, nil
		},
	}

	h := &Handler{
		aggregation: aggregation,
		analytics:   analytics,
		csvImport:   csvImport,
		requirePermission: func(_, _ string) gin.HandlerFunc {
			return func(c *gin.Context) { c.Next() }
		},
		requireAnyPermission: noopPermissionAny,
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		setClinicID(c)
		setSystemAdmin(c)
		c.Next()
	})
	h.RegisterRoutes(r.Group("/api/v1"))

	requests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/clinics/999/owners/aggregations"},
		{method: http.MethodGet, path: "/api/v1/clinics/999/lstep/analytics/delivery-stats?year_month=2026-07"},
		{method: http.MethodGet, path: "/api/v1/clinics/999/lstep/analytics/visit-conversion?year_month=2026-07"},
		{method: http.MethodGet, path: "/api/v1/clinics/999/owners/42/lstep/friend-attributes"},
		{method: http.MethodGet, path: "/api/v1/clinics/999/lstep/csv-imports"},
	}
	for _, request := range requests {
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, httptest.NewRequest(request.method, request.path, http.NoBody))
		assert.Less(t, recorder.Code, http.StatusBadRequest, request.path)
	}

	body, contentType := buildCSVMultipart(t, "line_user_id\nU1")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/clinics/999/lstep/csv-imports/friend-attributes", body)
	request.Header.Set("Content-Type", contentType)
	r.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusCreated, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Location"), "/clinics/1/")

	require.Len(t, calls, 6)
	for _, got := range calls {
		assert.Equal(t, uint64(1), got.clinicID, got.name)
	}
}
