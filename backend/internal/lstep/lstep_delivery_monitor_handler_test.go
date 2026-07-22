package lstep

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---- mock LstepDeliveryMonitorService ----

type mockLstepDeliveryMonitorService struct {
	getSummaryFn func(ctx context.Context, input GetDeliveryMonitorSummaryInput) (DeliveryTriggerSummary, error)
	getLogsFn    func(ctx context.Context, input *GetDeliveryMonitorLogsInput) (DeliveryTriggerLogsPage, error)
}

func (m *mockLstepDeliveryMonitorService) GetSummary(ctx context.Context, input GetDeliveryMonitorSummaryInput) (DeliveryTriggerSummary, error) {
	return m.getSummaryFn(ctx, input)
}

func (m *mockLstepDeliveryMonitorService) GetLogs(ctx context.Context, input *GetDeliveryMonitorLogsInput) (DeliveryTriggerLogsPage, error) {
	return m.getLogsFn(ctx, input)
}

func newHandlerWithLstepDeliveryMonitorSvc(svc LstepDeliveryMonitorService) *Handler {
	return &Handler{deliveryMonitor: svc}
}

// ---- conversion function tests ----

func TestToDeliveryTriggerSummaryResponse(t *testing.T) {
	tests := []struct {
		name  string
		input DeliveryTriggerSummary
		want  map[string]int64
	}{
		{
			name: "keeps existing breakdown map",
			input: DeliveryTriggerSummary{
				Scheduled:               10,
				Fired:                   5,
				Excluded:                2,
				Failed:                  1,
				SuppressedByPriority:    3,
				ExcludedReasonBreakdown: map[string]int64{"blocked": 2},
			},
			want: map[string]int64{"blocked": 2},
		},
		{
			name: "replaces nil breakdown with empty map",
			input: DeliveryTriggerSummary{
				Scheduled: 1,
			},
			want: map[string]int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDeliveryTriggerSummaryResponse(tt.input)
			assert.Equal(t, tt.input.Scheduled, got.Scheduled)
			assert.Equal(t, tt.input.Fired, got.Fired)
			assert.Equal(t, tt.input.Excluded, got.Excluded)
			assert.Equal(t, tt.input.Failed, got.Failed)
			assert.Equal(t, tt.input.SuppressedByPriority, got.SuppressedByPriority)
			assert.Equal(t, tt.want, got.ExcludedReasonBreakdown)
		})
	}
}

func TestToDeliveryTriggerLogItemResponse(t *testing.T) {
	firedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	excludedReason := "dormant"

	tests := []struct {
		name string
		item *DeliveryTriggerLogItem
	}{
		{
			name: "converts item with fired_at and excluded_reason set",
			item: &DeliveryTriggerLogItem{
				ID:             1,
				OwnerID:        2,
				OwnerName:      "田中太郎",
				TriggerType:    "birthday",
				ScheduledAt:    time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC),
				Status:         "fired",
				FiredAt:        &firedAt,
				ExcludedReason: &excludedReason,
			},
		},
		{
			name: "converts item with nil fired_at and excluded_reason",
			item: &DeliveryTriggerLogItem{
				ID:          3,
				OwnerID:     4,
				OwnerName:   "山田花子",
				TriggerType: "vaccine",
				ScheduledAt: time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC),
				Status:      "scheduled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDeliveryTriggerLogItemResponse(tt.item)
			assert.Equal(t, fmt.Sprintf("%d", tt.item.ID), got.ID)
			assert.Equal(t, fmt.Sprintf("%d", tt.item.OwnerID), got.OwnerID)
			assert.Equal(t, tt.item.OwnerName, got.OwnerName)
			assert.Equal(t, tt.item.TriggerType, got.TriggerType)
			assert.Equal(t, tt.item.Status, got.Status)
			if tt.item.FiredAt == nil {
				assert.Nil(t, got.FiredAt)
			} else {
				wantFiredAt := tt.item.FiredAt.In(time.Local).Format(time.RFC3339)
				assert.Equal(t, wantFiredAt, *got.FiredAt)
			}
			assert.Equal(t, tt.item.ExcludedReason, got.ExcludedReason)
		})
	}
}

func TestToDeliveryTriggerLogsPageResponse(t *testing.T) {
	tests := []struct {
		name string
		page DeliveryTriggerLogsPage
	}{
		{
			name: "converts page with items",
			page: DeliveryTriggerLogsPage{
				Items: []DeliveryTriggerLogItem{
					{ID: 1, OwnerID: 2, ScheduledAt: time.Now()},
					{ID: 2, OwnerID: 3, ScheduledAt: time.Now()},
				},
				Total:   2,
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name: "converts page with empty items",
			page: DeliveryTriggerLogsPage{
				Items:   []DeliveryTriggerLogItem{},
				Total:   0,
				Page:    1,
				PerPage: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDeliveryTriggerLogsPageResponse(tt.page)
			assert.Len(t, got.Items, len(tt.page.Items))
			assert.Equal(t, tt.page.Total, got.Total)
			assert.Equal(t, tt.page.Page, got.Page)
			assert.Equal(t, tt.page.PerPage, got.PerPage)
		})
	}
}

// ---- GetLstepDeliveryTriggerSummary ----

func TestGetLstepDeliveryTriggerSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockLstepDeliveryMonitorService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns summary",
			query:    "from=2026-05-01&to=2026-05-31",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockLstepDeliveryMonitorService{
				getSummaryFn: func(_ context.Context, input GetDeliveryMonitorSummaryInput) (DeliveryTriggerSummary, error) {
					assert.Equal(t, uint64(1), input.ClinicID)
					return DeliveryTriggerSummary{Scheduled: 5, Fired: 3}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"scheduled":5`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockLstepDeliveryMonitorService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid from date",
			query:      "from=invalid",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockLstepDeliveryMonitorService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "from=2026-05-01&to=2026-05-31",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockLstepDeliveryMonitorService{
				getSummaryFn: func(_ context.Context, _ GetDeliveryMonitorSummaryInput) (DeliveryTriggerSummary, error) {
					return DeliveryTriggerSummary{}, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLstepDeliveryMonitorSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetLstepDeliveryTriggerSummary(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestGetLstepDeliveryTriggerSummary_ClinicAliasUsesAuthenticatedClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	svc := &mockLstepDeliveryMonitorService{
		getSummaryFn: func(_ context.Context, input GetDeliveryMonitorSummaryInput) (DeliveryTriggerSummary, error) {
			called = true
			assert.Equal(t, uint64(1), input.ClinicID)
			return DeliveryTriggerSummary{}, nil
		},
	}
	h := newHandlerWithLstepDeliveryMonitorSvc(svc)
	r := gin.New()
	r.GET("/clinics/:clinic_id/lstep/delivery-monitor/summary", func(c *gin.Context) {
		c.Set("clinic_id", "1")
	}, h.GetLstepDeliveryTriggerSummary)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clinics/999/lstep/delivery-monitor/summary?from=2026-05-01&to=2026-05-31", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
}

// ---- GetLstepDeliveryTriggerLogs ----

func TestGetLstepDeliveryTriggerLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockLstepDeliveryMonitorService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns logs page",
			query:    "from=2026-05-01&to=2026-05-31&page=1&per_page=20",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockLstepDeliveryMonitorService{
				getLogsFn: func(_ context.Context, input *GetDeliveryMonitorLogsInput) (DeliveryTriggerLogsPage, error) {
					assert.Equal(t, uint64(1), input.ClinicID)
					return DeliveryTriggerLogsPage{
						Items:   []DeliveryTriggerLogItem{{ID: 1, OwnerID: 2, ScheduledAt: time.Now()}},
						Total:   1,
						Page:    1,
						PerPage: 20,
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"total":1`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockLstepDeliveryMonitorService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid page",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockLstepDeliveryMonitorService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "from=2026-05-01&to=2026-05-31",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockLstepDeliveryMonitorService{
				getLogsFn: func(_ context.Context, _ *GetDeliveryMonitorLogsInput) (DeliveryTriggerLogsPage, error) {
					return DeliveryTriggerLogsPage{}, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLstepDeliveryMonitorSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetLstepDeliveryTriggerLogs(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
