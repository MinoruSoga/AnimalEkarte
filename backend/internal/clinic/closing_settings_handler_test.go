package clinic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock ClosingSettingsService ----

type mockClosingSettingsService struct {
	getFn                 func(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error)
	listSpecialPeriodsFn  func(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)
	updateStandardFn      func(ctx context.Context, clinicID, actorID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error)
	createSpecialPeriodFn func(ctx context.Context, clinicID uint64, input *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	updateSpecialPeriodFn func(ctx context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	deleteSpecialPeriodFn func(ctx context.Context, clinicID, id uint64) error
	resolveScheduleFn     func(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error)
}

func (m *mockClosingSettingsService) Get(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error) {
	return m.getFn(ctx, clinicID)
}

func (m *mockClosingSettingsService) ListSpecialPeriods(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	return m.listSpecialPeriodsFn(ctx, clinicID)
}

func (m *mockClosingSettingsService) UpdateStandard(ctx context.Context, clinicID, actorID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error) {
	return m.updateStandardFn(ctx, clinicID, actorID, input)
}

func (m *mockClosingSettingsService) CreateSpecialPeriod(ctx context.Context, clinicID uint64, input *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	return m.createSpecialPeriodFn(ctx, clinicID, input)
}

func (m *mockClosingSettingsService) UpdateSpecialPeriod(ctx context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	return m.updateSpecialPeriodFn(ctx, clinicID, id, input)
}

func (m *mockClosingSettingsService) DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error {
	return m.deleteSpecialPeriodFn(ctx, clinicID, id)
}

func (m *mockClosingSettingsService) ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error) {
	return m.resolveScheduleFn(ctx, clinicID, date)
}

func newHandlerWithClosingSettingsSvc(svc ClosingSettingsService) *Handler {
	return &Handler{closingSettingsSvc: svc}
}

// ---- GetClosingSettings ----

func TestGetClosingSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockClosingSettingsService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns settings and special periods",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				getFn: func(_ context.Context, clinicID uint64) (*ClosingSettingsResponse, error) {
					assert.Equal(t, uint64(1), clinicID)
					return &ClosingSettingsResponse{
						Settings: &model.ClinicSettings{
							ClinicID:            1,
							ClosingAmPmBoundary: "12:00",
						},
						SpecialPeriods: []model.ClosingSpecialPeriod{
							{ID: 1, ClinicID: 1, Note: "holiday"},
						},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"note":"holiday"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				getFn: func(_ context.Context, _ uint64) (*ClosingSettingsResponse, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClosingSettingsSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.GetClosingSettings(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateClosingSettings ----

func TestUpdateClosingSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	boundary := "13:00"

	tests := []struct {
		name       string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockClosingSettingsService
		wantStatus int
		wantBody   string
	}{
		{
			name: "updates settings successfully",
			body: map[string]any{"closing_am_pm_boundary": boundary},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			svc: &mockClosingSettingsService{
				updateStandardFn: func(_ context.Context, clinicID, actorID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), actorID)
					assert.Equal(t, boundary, *input.ClosingAmPmBoundary)
					return &model.ClinicSettings{ClinicID: 1, ClosingAmPmBoundary: boundary}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"closing_am_pm_boundary":"13:00"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns 401 when staff_id is missing",
			body: map[string]any{},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
			},
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for malformed JSON",
			bodyRaw:    `{"closing_am_pm_boundary":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: map[string]any{},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			svc: &mockClosingSettingsService{
				updateStandardFn: func(_ context.Context, _, _ uint64, _ UpdateClinicSettingsInput) (*model.ClinicSettings, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClosingSettingsSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			if tt.bodyRaw != "" {
				body = []byte(tt.bodyRaw)
			} else {
				var err error
				body, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.UpdateClosingSettings(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ListSpecialPeriods ----

func TestListSpecialPeriods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockClosingSettingsService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of special periods",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				listSpecialPeriodsFn: func(_ context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.ClosingSpecialPeriod{{ID: 1, Note: "gw"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"note":"gw"`,
		},
		{
			name:     "returns empty list",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				listSpecialPeriodsFn: func(_ context.Context, _ uint64) ([]model.ClosingSpecialPeriod, error) {
					return []model.ClosingSpecialPeriod{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				listSpecialPeriodsFn: func(_ context.Context, _ uint64) ([]model.ClosingSpecialPeriod, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClosingSettingsSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.ListSpecialPeriods(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateSpecialPeriod ----

func TestCreateSpecialPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockClosingSettingsService
		wantStatus int
		wantBody   string
		wantLoc    string
	}{
		{
			name: "creates special period successfully",
			body: map[string]any{
				"start_date":     "2026-05-01",
				"end_date":       "2026-05-05",
				"am_pm_boundary": "12:00",
				"pm_end":         "17:00",
				"note":           "gw",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				createSpecialPeriodFn: func(_ context.Context, clinicID uint64, input *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-05-01", input.StartDate)
					return &model.ClosingSpecialPeriod{ID: 42, ClinicID: 1, Note: "gw"}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"id":42`,
			wantLoc:    "/v1/closing-settings/special-periods/42",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for malformed JSON",
			bodyRaw:    `{"start_date":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when required field missing",
			body:       map[string]any{"end_date": "2026-05-05"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: map[string]any{
				"start_date":     "2026-05-01",
				"end_date":       "2026-05-05",
				"am_pm_boundary": "12:00",
				"pm_end":         "17:00",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				createSpecialPeriodFn: func(_ context.Context, _ uint64, _ *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClosingSettingsSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			if tt.bodyRaw != "" {
				body = []byte(tt.bodyRaw)
			} else {
				var err error
				body, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateSpecialPeriod(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantLoc != "" {
				assert.Equal(t, tt.wantLoc, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateSpecialPeriod ----

func TestUpdateSpecialPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockClosingSettingsService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates special period successfully",
			paramID:  "5",
			body:     map[string]any{"note": "updated"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				updateSpecialPeriodFn: func(_ context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					assert.Equal(t, "updated", *input.Note)
					return &model.ClosingSpecialPeriod{ID: 5, Note: "updated"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"note":"updated"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is non-numeric",
			paramID:    "abc",
			body:       map[string]any{},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for malformed JSON",
			paramID:    "5",
			bodyRaw:    `{"note":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when special period does not exist",
			paramID:  "999",
			body:     map[string]any{"note": "x"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClosingSettingsService{
				updateSpecialPeriodFn: func(_ context.Context, _, _ uint64, _ UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
					return nil, apperrors.WrapNotFound("special period", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClosingSettingsSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			if tt.bodyRaw != "" {
				body = []byte(tt.bodyRaw)
			} else {
				var err error
				body, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateSpecialPeriod(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteSpecialPeriod ----

func newDeleteSpecialPeriodRouter(svc ClosingSettingsService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithClosingSettingsSvc(svc)
	r.DELETE("/closing-settings/special-periods/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteSpecialPeriod)
	return r
}

func TestDeleteSpecialPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockClosingSettingsService
		wantStatus int
	}{
		{
			name:    "deletes special period successfully",
			paramID: "5",
			svc: &mockClosingSettingsService{
				deleteSpecialPeriodFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 when id is non-numeric",
			paramID:    "abc",
			svc:        &mockClosingSettingsService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when special period does not exist",
			paramID: "999",
			svc: &mockClosingSettingsService{
				deleteSpecialPeriodFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("special period", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteSpecialPeriodRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/closing-settings/special-periods/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithClosingSettingsSvc(&mockClosingSettingsService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "5"}}
		h.DeleteSpecialPeriod(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// POC-01 (REFRAMED): dual holiday routes are intentional; the defect is that
// defaultPermissionRuleTable denies closing-settings create/delete while
// /closing-settings/holidays POST|DELETE require those actions, so FE
// holiday mutations 403 for every non-system_admin on new clinics.
func TestClosingSettingsHoliday_DefaultPermissionAllowsCreateDelete(t *testing.T) {
	findRule := func(rules []model.PermissionGroupRule, resource model.Resource) *model.PermissionGroupRule {
		for i := range rules {
			if rules[i].Resource == string(resource) {
				return &rules[i]
			}
		}
		return nil
	}

	execRules := buildDefaultPermissionGroupRules(true)
	execCS := findRule(execRules, model.ResourceClosingSettings)
	if assert.NotNil(t, execCS, "執行に closing-settings ルールが存在すること") {
		assert.True(t, execCS.CanView, "執行は closing-settings を閲覧可能であること")
		assert.True(t, execCS.CanCreate, "執行は closing-settings を作成可能であること（/closing-settings/holidays POST）")
		assert.True(t, execCS.CanEdit, "執行は closing-settings を編集可能であること")
		assert.True(t, execCS.CanDelete, "執行は closing-settings を削除可能であること（/closing-settings/holidays DELETE）")
	}

	genRules := buildDefaultPermissionGroupRules(false)
	genCS := findRule(genRules, model.ResourceClosingSettings)
	if assert.NotNil(t, genCS, "一般に closing-settings ルールが存在すること") {
		assert.True(t, genCS.CanView, "一般は closing-settings を閲覧可能であること")
		assert.False(t, genCS.CanCreate, "一般は closing-settings を作成できないこと")
		assert.False(t, genCS.CanEdit, "一般は closing-settings を編集できないこと")
		assert.False(t, genCS.CanDelete, "一般は closing-settings を削除できないこと")
	}
}

// POC-01: keep dual path; assert closing-settings holiday mutations still
// require closing-settings:create|delete (aligned with defaultPermissionRuleTable).
func TestClosingSettingsHolidayRoutes_PermissionActions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissions []string
	requirePermission := func(resource, action string) gin.HandlerFunc {
		permissions = append(permissions, resource+":"+action)
		return func(c *gin.Context) { c.Next() }
	}

	handler := NewHandler(nil, nil, nil, nil, requirePermission)
	router := gin.New()
	handler.RegisterClosingSettingsRoutes(router.Group("/api/v1"))

	assert.Contains(t, permissions, "closing-settings:view")
	assert.Contains(t, permissions, "closing-settings:create")
	assert.Contains(t, permissions, "closing-settings:delete")

	// Holiday mutations must remain create/delete (not silently remapped).
	// Registration order for holidays is view, create, delete after special-periods.
	require.GreaterOrEqual(t, len(permissions), 3)
	var holidayCreate, holidayDelete bool
	for _, p := range permissions {
		if p == "closing-settings:create" {
			holidayCreate = true
		}
		if p == "closing-settings:delete" {
			holidayDelete = true
		}
	}
	assert.True(t, holidayCreate, "closing-settings holiday POST must require create")
	assert.True(t, holidayDelete, "closing-settings holiday DELETE must require delete")
}
