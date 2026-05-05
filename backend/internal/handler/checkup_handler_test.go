package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock CheckupService ----

type mockCheckupService struct {
	listFn         func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	listByClinicFn func(ctx context.Context, input service.ListCheckupsByClinicInput) ([]model.Checkup, error)
	getByIDFn      func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) (*model.Checkup, error)
	createFn       func(ctx context.Context, medicalRecordID uint64, input *service.CreateCheckupInput) (*model.Checkup, error)
	updateFn       func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, input *service.UpdateCheckupInput) (*model.Checkup, error)
	deleteFn       func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) error
	getAlertsFn    func(ctx context.Context, clinicID uint64, withinDays int) (*service.CheckupAlertsResult, error)
}

func (m *mockCheckupService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID, medicalRecordID)
	}
	return nil, nil
}

func (m *mockCheckupService) ListByClinic(ctx context.Context, input service.ListCheckupsByClinicInput) ([]model.Checkup, error) {
	if m.listByClinicFn != nil {
		return m.listByClinicFn(ctx, input)
	}
	return nil, nil
}

func (m *mockCheckupService) GetByID(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) (*model.Checkup, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, medicalRecordID, checkupID)
	}
	return nil, nil
}

func (m *mockCheckupService) Create(ctx context.Context, medicalRecordID uint64, input *service.CreateCheckupInput) (*model.Checkup, error) {
	if m.createFn != nil {
		return m.createFn(ctx, medicalRecordID, input)
	}
	return nil, nil
}

func (m *mockCheckupService) Update(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, input *service.UpdateCheckupInput) (*model.Checkup, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, medicalRecordID, checkupID, input)
	}
	return nil, nil
}

func (m *mockCheckupService) Delete(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, medicalRecordID, checkupID)
	}
	return nil
}

func (m *mockCheckupService) GetAlerts(ctx context.Context, clinicID uint64, withinDays int) (*service.CheckupAlertsResult, error) {
	if m.getAlertsFn != nil {
		return m.getAlertsFn(ctx, clinicID, withinDays)
	}
	return &service.CheckupAlertsResult{
		Overdue:  []model.Checkup{},
		Upcoming: []model.Checkup{},
	}, nil
}

// ---- helpers ----

func newHandlerWithCheckupSvc(svc service.CheckupService) *Handler {
	return &Handler{svc: &service.Services{Checkup: svc}}
}

// ---- GetCheckupAlerts tests ----

func TestGetCheckupAlerts_DefaultWithinDays(t *testing.T) {
	gin.SetMode(gin.TestMode)

	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)

	svc := &mockCheckupService{
		getAlertsFn: func(_ context.Context, _ uint64, _ int) (*service.CheckupAlertsResult, error) {
			return &service.CheckupAlertsResult{
				Overdue:  []model.Checkup{{ID: 1, NextDate: &yesterday}},
				Upcoming: []model.Checkup{{ID: 2, NextDate: &tomorrow}},
			}, nil
		},
	}

	h := newHandlerWithCheckupSvc(svc)
	r := gin.New()
	r.GET("/checkups/alerts", func(c *gin.Context) { setClinicID(c) }, h.GetCheckupAlerts)

	req := httptest.NewRequest(http.MethodGet, "/checkups/alerts", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	overdue, _ := body["overdue"].(map[string]any)
	upcoming, _ := body["upcoming"].(map[string]any)
	assert.Equal(t, float64(1), overdue["count"])
	assert.Equal(t, float64(1), upcoming["count"])
}

func TestGetCheckupAlerts_CustomWithinDays(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedDays int
	svc := &mockCheckupService{
		getAlertsFn: func(_ context.Context, _ uint64, days int) (*service.CheckupAlertsResult, error) {
			capturedDays = days
			return &service.CheckupAlertsResult{
				Overdue:  []model.Checkup{},
				Upcoming: []model.Checkup{},
			}, nil
		},
	}

	h := newHandlerWithCheckupSvc(svc)
	r := gin.New()
	r.GET("/checkups/alerts", func(c *gin.Context) { setClinicID(c) }, h.GetCheckupAlerts)

	req := httptest.NewRequest(http.MethodGet, "/checkups/alerts?within_days=60", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 60, capturedDays)
}

func TestGetCheckupAlerts_InvalidWithinDays(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newHandlerWithCheckupSvc(&mockCheckupService{})
	r := gin.New()
	r.GET("/checkups/alerts", func(c *gin.Context) { setClinicID(c) }, h.GetCheckupAlerts)

	req := httptest.NewRequest(http.MethodGet, "/checkups/alerts?within_days=abc", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCheckupAlerts_AuthorizationGate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// non-system-admin with no granted permissions → RequirePermission returns 403
	h := &Handler{svc: &service.Services{
		Checkup:             &mockCheckupService{},
		EffectivePermission: &mockEffectivePermissionService{},
	}}
	r := gin.New()
	r.GET("/checkups/alerts",
		func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) },
		h.RequirePermission(string(model.ResourceCheckups), "view"),
		h.GetCheckupAlerts,
	)

	req := httptest.NewRequest(http.MethodGet, "/checkups/alerts", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
