package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock LstepSettingsService ----

type mockLstepSettingsService struct {
	getSettingsFn    func(ctx context.Context, clinicID uint64) (*service.LstepSettingsResponse, error)
	updateSettingsFn func(ctx context.Context, clinicID uint64, input service.UpdateLstepSettingsInput, actorID *uint64) (*service.LstepSettingsResponse, error)
	deleteSettingsFn func(ctx context.Context, clinicID uint64, actorID *uint64) error
	testConnectionFn func(ctx context.Context, clinicID uint64) (*service.LstepConnectionTestResult, error)
}

func (m *mockLstepSettingsService) GetSettings(ctx context.Context, clinicID uint64) (*service.LstepSettingsResponse, error) {
	if m.getSettingsFn != nil {
		return m.getSettingsFn(ctx, clinicID)
	}
	return &service.LstepSettingsResponse{}, nil
}
func (m *mockLstepSettingsService) UpdateSettings(ctx context.Context, clinicID uint64, input service.UpdateLstepSettingsInput, actorID *uint64) (*service.LstepSettingsResponse, error) {
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
func (m *mockLstepSettingsService) GetRawCredentials(_ context.Context, _ uint64) (string, string, string, error) {
	return "", "", "", nil
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
			req := httptest.NewRequest(http.MethodGet, "/lstep-settings", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
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
				updateSettingsFn: func(_ context.Context, _ uint64, _ service.UpdateLstepSettingsInput, _ *uint64) (*service.LstepSettingsResponse, error) {
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
			req := httptest.NewRequest(http.MethodDelete, "/lstep-settings", nil)
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
			req := httptest.NewRequest(http.MethodPost, "/lstep-settings/test-connection", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
