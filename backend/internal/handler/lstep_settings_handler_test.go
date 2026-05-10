package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock LstepSettingsService ----

type mockLstepSettingsService struct {
	getSettingsFn    func(ctx context.Context, clinicID uint64) (*service.LstepSettingsResponse, error)
	updateSettingsFn func(ctx context.Context, clinicID uint64, input *service.UpdateLstepSettingsInput, actorID *uint64) (*service.LstepSettingsResponse, error)
	deleteSettingsFn func(ctx context.Context, clinicID uint64, actorID *uint64) error
	testConnectionFn func(ctx context.Context, clinicID uint64) (*service.LstepConnectionTestResult, error)
}

func (m *mockLstepSettingsService) GetSettings(ctx context.Context, clinicID uint64) (*service.LstepSettingsResponse, error) {
	if m.getSettingsFn != nil {
		return m.getSettingsFn(ctx, clinicID)
	}
	return &service.LstepSettingsResponse{}, nil
}
func (m *mockLstepSettingsService) UpdateSettings(ctx context.Context, clinicID uint64, input *service.UpdateLstepSettingsInput, actorID *uint64) (*service.LstepSettingsResponse, error) {
	if m.updateSettingsFn != nil {
		return m.updateSettingsFn(ctx, clinicID, input, actorID)
	}
	return &service.LstepSettingsResponse{}, nil
}
func (m *mockLstepSettingsService) DeleteSettings(ctx context.Context, clinicID uint64, actorID *uint64) error {
	if m.deleteSettingsFn != nil {
		return m.deleteSettingsFn(ctx, clinicID, actorID)
	}
	return nil
}
func (m *mockLstepSettingsService) TestConnection(ctx context.Context, clinicID uint64) (*service.LstepConnectionTestResult, error) {
	if m.testConnectionFn != nil {
		return m.testConnectionFn(ctx, clinicID)
	}
	return &service.LstepConnectionTestResult{LstepOK: true, LineOK: true}, nil
}
func (m *mockLstepSettingsService) GetRawCredentials(_ context.Context, _ uint64) (apiKey, baseURL, lineToken string, err error) {
	return "", "", "", nil
}
func (m *mockLstepSettingsService) IsSyncEnabled(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}
func (m *mockLstepSettingsService) GetCPMVersion(_ context.Context, _ uint64) (string, error) {
	return "v1", nil
}

// ---- helpers ----

func newHandlerWithLstepSettingsSvc(svc service.LstepSettingsService) *Handler {
	return &Handler{svc: &service.Services{LstepSettings: svc}}
}

func newGetLstepSettingsRouter(svc service.LstepSettingsService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepSettingsSvc(svc)
	if withClinicID {
		r.GET("/lstep-settings", func(c *gin.Context) { setClinicID(c) }, h.GetLstepSettings)
	} else {
		r.GET("/lstep-settings", h.GetLstepSettings)
	}
	return r
}

func newPatchLstepSettingsRouter(svc service.LstepSettingsService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepSettingsSvc(svc)
	if withClinicID {
		r.PATCH("/lstep-settings", func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") }, h.UpdateLstepSettings)
	} else {
		r.PATCH("/lstep-settings", h.UpdateLstepSettings)
	}
	return r
}

func newDeleteLstepSettingsRouter(svc service.LstepSettingsService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepSettingsSvc(svc)
	if withClinicID {
		r.DELETE("/lstep-settings", func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") }, h.DeleteLstepSettings)
	} else {
		r.DELETE("/lstep-settings", h.DeleteLstepSettings)
	}
	return r
}

func newPostLstepTestConnectionRouter(svc service.LstepSettingsService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepSettingsSvc(svc)
	if withClinicID {
		r.POST("/lstep-settings/test-connection", func(c *gin.Context) { setClinicID(c) }, h.TestLstepConnection)
	} else {
		r.POST("/lstep-settings/test-connection", h.TestLstepConnection)
	}
	return r
}

// ---- tests ----

func TestGetLstepSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		svc        *mockLstepSettingsService
		wantStatus int
	}{
		{
			name:       "200 success",
			svc:        &mockLstepSettingsService{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 no clinic",
			svc:        &mockLstepSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "500 service error",
			svc: &mockLstepSettingsService{
				getSettingsFn: func(_ context.Context, _ uint64) (*service.LstepSettingsResponse, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newGetLstepSettingsRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodGet, "/lstep-settings", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetLstepSettingsIncludesSyncFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	syncEnabledAt := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	svc := &mockLstepSettingsService{
		getSettingsFn: func(_ context.Context, _ uint64) (*service.LstepSettingsResponse, error) {
			return &service.LstepSettingsResponse{
				IsSyncEnabled: true,
				SyncEnabledAt: &syncEnabledAt,
			}, nil
		},
	}

	router := newGetLstepSettingsRouter(svc, true)
	req := httptest.NewRequest(http.MethodGet, "/lstep-settings", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, true, body["is_sync_enabled"])
	assert.NotEmpty(t, body["sync_enabled_at"])
}

func TestPatchLstepSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validBody, _ := json.Marshal(map[string]string{"lstep_api_key": "key123"})
	tests := []struct {
		name       string
		svc        *mockLstepSettingsService
		body       []byte
		wantStatus int
	}{
		{
			name:       "200 success",
			svc:        &mockLstepSettingsService{},
			body:       validBody,
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 no clinic",
			svc:        &mockLstepSettingsService{},
			body:       validBody,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400 invalid JSON",
			svc:        &mockLstepSettingsService{},
			body:       []byte(`{invalid`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "500 service error",
			svc: &mockLstepSettingsService{
				updateSettingsFn: func(_ context.Context, _ uint64, _ *service.UpdateLstepSettingsInput, _ *uint64) (*service.LstepSettingsResponse, error) {
					return nil, errors.New("db error")
				},
			},
			body:       validBody,
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchLstepSettingsRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodPatch, "/lstep-settings", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestPatchLstepSettingsPassesSyncEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got *bool
	svc := &mockLstepSettingsService{
		updateSettingsFn: func(_ context.Context, _ uint64, input *service.UpdateLstepSettingsInput, _ *uint64) (*service.LstepSettingsResponse, error) {
			got = input.IsSyncEnabled
			return &service.LstepSettingsResponse{IsSyncEnabled: true}, nil
		},
	}

	body, _ := json.Marshal(map[string]any{"is_sync_enabled": true})
	router := newPatchLstepSettingsRouter(svc, true)
	req := httptest.NewRequest(http.MethodPatch, "/lstep-settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	if assert.NotNil(t, got) {
		assert.True(t, *got)
	}
}

func TestPatchLstepSettingsPassesFireHourJST(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got *int
	svc := &mockLstepSettingsService{
		updateSettingsFn: func(_ context.Context, _ uint64, input *service.UpdateLstepSettingsInput, _ *uint64) (*service.LstepSettingsResponse, error) {
			got = input.FireHourJST
			return &service.LstepSettingsResponse{FireHourJST: 9}, nil
		},
	}

	body, _ := json.Marshal(map[string]any{"fire_hour_jst": 9})
	router := newPatchLstepSettingsRouter(svc, true)
	req := httptest.NewRequest(http.MethodPatch, "/lstep-settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	if assert.NotNil(t, got) {
		assert.Equal(t, 9, *got)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(9), resp["fire_hour_jst"])
}

func TestDeleteLstepSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		svc        *mockLstepSettingsService
		wantStatus int
	}{
		{
			name:       "204 success",
			svc:        &mockLstepSettingsService{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "401 no clinic",
			svc:        &mockLstepSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "500 service error",
			svc: &mockLstepSettingsService{
				deleteSettingsFn: func(_ context.Context, _ uint64, _ *uint64) error {
					return errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteLstepSettingsRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodDelete, "/lstep-settings", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestPostLstepTestConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		svc        *mockLstepSettingsService
		wantStatus int
	}{
		{
			name:       "200 both ok",
			svc:        &mockLstepSettingsService{},
			wantStatus: http.StatusOK,
		},
		{
			name: "200 partial failure still returns 200",
			svc: &mockLstepSettingsService{
				testConnectionFn: func(_ context.Context, _ uint64) (*service.LstepConnectionTestResult, error) {
					return &service.LstepConnectionTestResult{LstepOK: false, LstepError: "invalid key"}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 no clinic",
			svc:        &mockLstepSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "500 service error",
			svc: &mockLstepSettingsService{
				testConnectionFn: func(_ context.Context, _ uint64) (*service.LstepConnectionTestResult, error) {
					return nil, errors.New("network error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPostLstepTestConnectionRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodPost, "/lstep-settings/test-connection", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
