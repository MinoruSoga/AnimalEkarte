package lstep

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

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock LstepSettingsService ----

type mockLstepSettingsService struct {
	getSettingsFn                   func(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error)
	updateSettingsFn                func(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput, actorID *uint64) (*LstepSettingsResponse, error)
	deleteSettingsFn                func(ctx context.Context, clinicID uint64, actorID *uint64) error
	testConnectionFn                func(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error)
	isSyncEnabledFn                 func(ctx context.Context, clinicID uint64) (bool, error)
	getDormantThresholdsFn          func(ctx context.Context, clinicID uint64) (model.DormantThresholds, error)
	getHealthPreventionThresholdsFn func(ctx context.Context, clinicID uint64) (model.HealthPreventionThresholds, error)
	getCPMV1ThresholdsFn            func(ctx context.Context, clinicID uint64) (model.CPMV1Thresholds, error)
	// getRawCredentialsFn は line_send_service_test の credential 失敗系で使う（BE9-2C L②）。
	getRawCredentialsFn func(ctx context.Context, clinicID uint64) (string, string, string, error)
}

func (m *mockLstepSettingsService) GetSettings(ctx context.Context, clinicID uint64) (*LstepSettingsResponse, error) {
	if m.getSettingsFn != nil {
		return m.getSettingsFn(ctx, clinicID)
	}
	return &LstepSettingsResponse{}, nil
}
func (m *mockLstepSettingsService) UpdateSettings(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput, actorID *uint64) (*LstepSettingsResponse, error) {
	if m.updateSettingsFn != nil {
		return m.updateSettingsFn(ctx, clinicID, input, actorID)
	}
	return &LstepSettingsResponse{}, nil
}
func (m *mockLstepSettingsService) DeleteSettings(ctx context.Context, clinicID uint64, actorID *uint64) error {
	if m.deleteSettingsFn != nil {
		return m.deleteSettingsFn(ctx, clinicID, actorID)
	}
	return nil
}
func (m *mockLstepSettingsService) TestConnection(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error) {
	if m.testConnectionFn != nil {
		return m.testConnectionFn(ctx, clinicID)
	}
	return &LstepConnectionTestResult{LstepOK: true, LineOK: true}, nil
}
func (m *mockLstepSettingsService) GetRawCredentials(ctx context.Context, clinicID uint64) (apiKey, baseURL, lineToken string, err error) {
	if m.getRawCredentialsFn != nil {
		return m.getRawCredentialsFn(ctx, clinicID)
	}
	return "", "", "", nil
}
func (m *mockLstepSettingsService) IsSyncEnabled(ctx context.Context, clinicID uint64) (bool, error) {
	if m.isSyncEnabledFn != nil {
		return m.isSyncEnabledFn(ctx, clinicID)
	}
	return true, nil
}
func (m *mockLstepSettingsService) GetCPMVersion(_ context.Context, _ uint64) (string, error) {
	return "v1", nil
}
func (m *mockLstepSettingsService) GetDormantThresholds(ctx context.Context, clinicID uint64) (model.DormantThresholds, error) {
	if m.getDormantThresholdsFn != nil {
		return m.getDormantThresholdsFn(ctx, clinicID)
	}
	return model.DormantThresholds{}.WithDefaults(), nil
}
func (m *mockLstepSettingsService) GetCPMV2Thresholds(_ context.Context, _ uint64) (model.CPMV2Thresholds, error) {
	return model.CPMV2Thresholds{}.WithDefaults(), nil
}
func (m *mockLstepSettingsService) GetCPMV1Thresholds(ctx context.Context, clinicID uint64) (model.CPMV1Thresholds, error) {
	if m.getCPMV1ThresholdsFn != nil {
		return m.getCPMV1ThresholdsFn(ctx, clinicID)
	}
	return model.CPMV1Thresholds{}.WithDefaults(), nil
}
func (m *mockLstepSettingsService) GetHealthPreventionThresholds(ctx context.Context, clinicID uint64) (model.HealthPreventionThresholds, error) {
	if m.getHealthPreventionThresholdsFn != nil {
		return m.getHealthPreventionThresholdsFn(ctx, clinicID)
	}
	return model.HealthPreventionThresholds{}.WithDefaults(), nil
}

// ---- helpers ----

func newHandlerWithLstepSettingsSvc(svc LstepSettingsService) *SettingsHandler {
	return NewSettingsHandler(svc, func(_, _ string) gin.HandlerFunc { return func(_ *gin.Context) {} })
}

func newGetLstepSettingsRouter(svc LstepSettingsService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepSettingsSvc(svc)
	if withClinicID {
		r.GET("/lstep-settings", func(c *gin.Context) { setClinicID(c) }, h.GetLstepSettings)
	} else {
		r.GET("/lstep-settings", h.GetLstepSettings)
	}
	return r
}

func newPatchLstepSettingsRouter(svc LstepSettingsService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepSettingsSvc(svc)
	if withClinicID {
		r.PATCH("/lstep-settings", func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") }, h.UpdateLstepSettings)
	} else {
		r.PATCH("/lstep-settings", h.UpdateLstepSettings)
	}
	return r
}

func newDeleteLstepSettingsRouter(svc LstepSettingsService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepSettingsSvc(svc)
	if withClinicID {
		r.DELETE("/lstep-settings", func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") }, h.DeleteLstepSettings)
	} else {
		r.DELETE("/lstep-settings", h.DeleteLstepSettings)
	}
	return r
}

func newPostLstepTestConnectionRouter(svc LstepSettingsService, withClinicID bool) *gin.Engine {
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
				getSettingsFn: func(_ context.Context, _ uint64) (*LstepSettingsResponse, error) {
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
		getSettingsFn: func(_ context.Context, _ uint64) (*LstepSettingsResponse, error) {
			return &LstepSettingsResponse{
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
				updateSettingsFn: func(_ context.Context, _ uint64, _ *UpdateLstepSettingsInput, _ *uint64) (*LstepSettingsResponse, error) {
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
		updateSettingsFn: func(_ context.Context, _ uint64, input *UpdateLstepSettingsInput, _ *uint64) (*LstepSettingsResponse, error) {
			got = input.IsSyncEnabled
			return &LstepSettingsResponse{IsSyncEnabled: true}, nil
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

func TestPatchLstepSettingsPassesCPMVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got *string
	svc := &mockLstepSettingsService{
		updateSettingsFn: func(_ context.Context, _ uint64, input *UpdateLstepSettingsInput, _ *uint64) (*LstepSettingsResponse, error) {
			got = input.CPMVersion
			return &LstepSettingsResponse{
				LstepAPIKeyMasked:            "",
				LstepBaseURL:                 "",
				LineChannelAccessTokenMasked: "",
				LineChannelSecretMasked:      "",
				LiffID:                       "",
				LineAccountName:              "",
				IsConfigured:                 false,
				LastUpdatedAt:                nil,
				IsSyncEnabled:                false,
				SyncEnabledAt:                nil,
				CPMVersion:                   "v2",
				DormantPrevention180Days:     180,
				DormantPrevention210Days:     210,
				DormantPrevention240Days:     240,
				DormantPrevention365Days:     365,
			}, nil
		},
	}

	body, _ := json.Marshal(map[string]any{"cpm_version": "v2"})
	router := newPatchLstepSettingsRouter(svc, true)
	req := httptest.NewRequest(http.MethodPatch, "/lstep-settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	if assert.NotNil(t, got) {
		assert.Equal(t, "v2", *got)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "v2", resp["cpm_version"])
}

func TestPatchLstepSettingsPassesDormantThresholds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got180, got210, got240, got365 *int
	svc := &mockLstepSettingsService{
		updateSettingsFn: func(_ context.Context, _ uint64, input *UpdateLstepSettingsInput, _ *uint64) (*LstepSettingsResponse, error) {
			got180 = input.DormantPrevention180Days
			got210 = input.DormantPrevention210Days
			got240 = input.DormantPrevention240Days
			got365 = input.DormantPrevention365Days
			return &LstepSettingsResponse{
				LstepAPIKeyMasked:            "",
				LstepBaseURL:                 "",
				LineChannelAccessTokenMasked: "",
				LineChannelSecretMasked:      "",
				LiffID:                       "",
				LineAccountName:              "",
				IsConfigured:                 false,
				LastUpdatedAt:                nil,
				IsSyncEnabled:                false,
				SyncEnabledAt:                nil,
				CPMVersion:                   "v1",
				DormantPrevention180Days:     180,
				DormantPrevention210Days:     210,
				DormantPrevention240Days:     240,
				DormantPrevention365Days:     365,
			}, nil
		},
	}

	body, _ := json.Marshal(map[string]any{
		"dormant_prevention_180_days": 180,
		"dormant_prevention_210_days": 210,
		"dormant_prevention_240_days": 240,
		"dormant_prevention_365_days": 365,
	})
	router := newPatchLstepSettingsRouter(svc, true)
	req := httptest.NewRequest(http.MethodPatch, "/lstep-settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	if assert.NotNil(t, got180) {
		assert.Equal(t, 180, *got180)
	}
	if assert.NotNil(t, got210) {
		assert.Equal(t, 210, *got210)
	}
	if assert.NotNil(t, got240) {
		assert.Equal(t, 240, *got240)
	}
	if assert.NotNil(t, got365) {
		assert.Equal(t, 365, *got365)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(180), resp["dormant_prevention_180_days"])
	assert.Equal(t, float64(210), resp["dormant_prevention_210_days"])
	assert.Equal(t, float64(240), resp["dormant_prevention_240_days"])
	assert.Equal(t, float64(365), resp["dormant_prevention_365_days"])
}

func TestGetLstepSettingsIncludesNewFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockLstepSettingsService{
		getSettingsFn: func(_ context.Context, _ uint64) (*LstepSettingsResponse, error) {
			return &LstepSettingsResponse{
				LstepAPIKeyMasked:            "",
				LstepBaseURL:                 "",
				LineChannelAccessTokenMasked: "",
				LineChannelSecretMasked:      "",
				LiffID:                       "",
				LineAccountName:              "",
				IsConfigured:                 false,
				LastUpdatedAt:                nil,
				IsSyncEnabled:                false,
				SyncEnabledAt:                nil,
				CPMVersion:                   "v2",
				DormantPrevention180Days:     180,
				DormantPrevention210Days:     210,
				DormantPrevention240Days:     240,
				DormantPrevention365Days:     365,
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
	assert.Equal(t, "v2", body["cpm_version"])
	assert.Equal(t, float64(180), body["dormant_prevention_180_days"])
	assert.Equal(t, float64(210), body["dormant_prevention_210_days"])
	assert.Equal(t, float64(240), body["dormant_prevention_240_days"])
	assert.Equal(t, float64(365), body["dormant_prevention_365_days"])
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
				testConnectionFn: func(_ context.Context, _ uint64) (*LstepConnectionTestResult, error) {
					return &LstepConnectionTestResult{LstepOK: false, LstepError: "invalid key"}, nil
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
				testConnectionFn: func(_ context.Context, _ uint64) (*LstepConnectionTestResult, error) {
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

func TestPatchLstepSettingsPassesCPMV1Thresholds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var gotDormant *int
	var gotNoahLTV *int64
	svc := &mockLstepSettingsService{
		updateSettingsFn: func(_ context.Context, _ uint64, input *UpdateLstepSettingsInput, _ *uint64) (*LstepSettingsResponse, error) {
			gotDormant = input.CPMV1DormantDays
			gotNoahLTV = input.CPMV1NoahLTV
			return &LstepSettingsResponse{
				CPMV1DormantDays: 300,
				CPMV1NoahLTV:     90000,
			}, nil
		},
	}

	body, _ := json.Marshal(map[string]any{
		"cpm_v1_dormant_days": 300,
		"cpm_v1_noah_ltv":     90000,
	})
	router := newPatchLstepSettingsRouter(svc, true)
	req := httptest.NewRequest(http.MethodPatch, "/lstep-settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	if assert.NotNil(t, gotDormant) {
		assert.Equal(t, 300, *gotDormant)
	}
	if assert.NotNil(t, gotNoahLTV) {
		assert.Equal(t, int64(90000), *gotNoahLTV)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(300), resp["cpm_v1_dormant_days"])
	assert.Equal(t, float64(90000), resp["cpm_v1_noah_ltv"])
}

func TestGetLstepSettingsIncludesHealthPreventionFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockLstepSettingsService{
		getSettingsFn: func(_ context.Context, _ uint64) (*LstepSettingsResponse, error) {
			return &LstepSettingsResponse{
				HealthPreventionLookbackDays: 365,
				VaccineDeadlineDays:          60,
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
	assert.Equal(t, float64(365), body["health_prevention_lookback_days"])
	assert.Equal(t, float64(60), body["vaccine_deadline_days"])
}

func TestPatchLstepSettingsPassesHealthPreventionThresholds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var gotLookback *int
	var gotVaccine *int
	svc := &mockLstepSettingsService{
		updateSettingsFn: func(_ context.Context, _ uint64, input *UpdateLstepSettingsInput, _ *uint64) (*LstepSettingsResponse, error) {
			gotLookback = input.HealthPreventionLookbackDays
			gotVaccine = input.VaccineDeadlineDays
			return &LstepSettingsResponse{
				HealthPreventionLookbackDays: 180,
				VaccineDeadlineDays:          30,
			}, nil
		},
	}

	body, _ := json.Marshal(map[string]any{
		"health_prevention_lookback_days": 180,
		"vaccine_deadline_days":           30,
	})
	router := newPatchLstepSettingsRouter(svc, true)
	req := httptest.NewRequest(http.MethodPatch, "/lstep-settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	if assert.NotNil(t, gotLookback) {
		assert.Equal(t, 180, *gotLookback)
	}
	if assert.NotNil(t, gotVaccine) {
		assert.Equal(t, 30, *gotVaccine)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(180), resp["health_prevention_lookback_days"])
	assert.Equal(t, float64(30), resp["vaccine_deadline_days"])
}

func TestGetLstepSettingsIncludesCPMV1Fields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockLstepSettingsService{
		getSettingsFn: func(_ context.Context, _ uint64) (*LstepSettingsResponse, error) {
			return &LstepSettingsResponse{
				CPMV1DormantDays:      240,
				CPMV1NoahDays:         365,
				CPMV1NoahAnnualVisits: 3,
				CPMV1NoahLTV:          80000,
				CPMV1CoreDays:         180,
				CPMV1CoreAnnualVisits: 2,
				CPMV1CoreLTV:          50000,
				CPMV1SpotMinAmount:    30000,
				CPMV1SpotInactiveDays: 90,
				CPMV1GrowingMaxDays:   90,
				CPMV1GrowingMinVisits: 2,
				CPMV1GrowingMaxVisits: 3,
				CPMV1LTVBreakLow:      20000,
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
	assert.Equal(t, float64(240), body["cpm_v1_dormant_days"])
	assert.Equal(t, float64(365), body["cpm_v1_noah_days"])
	assert.Equal(t, float64(3), body["cpm_v1_noah_annual_visits"])
	assert.Equal(t, float64(80000), body["cpm_v1_noah_ltv"])
	assert.Equal(t, float64(180), body["cpm_v1_core_days"])
	assert.Equal(t, float64(2), body["cpm_v1_core_annual_visits"])
	assert.Equal(t, float64(50000), body["cpm_v1_core_ltv"])
	assert.Equal(t, float64(30000), body["cpm_v1_spot_min_amount"])
	assert.Equal(t, float64(90), body["cpm_v1_spot_inactive_days"])
	assert.Equal(t, float64(90), body["cpm_v1_growing_max_days"])
	assert.Equal(t, float64(2), body["cpm_v1_growing_min_visits"])
	assert.Equal(t, float64(3), body["cpm_v1_growing_max_visits"])
	assert.Equal(t, float64(20000), body["cpm_v1_ltv_break_low"])
}
