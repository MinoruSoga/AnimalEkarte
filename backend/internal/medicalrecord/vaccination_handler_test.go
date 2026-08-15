package medicalrecord

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

// ---- mock VaccinationService ----

type mockVaccinationService struct {
	listFn    func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateVaccinationInput) (*model.Vaccination, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateVaccinationInput) (*model.Vaccination, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockVaccinationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error) {
	return m.listFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockVaccinationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockVaccinationService) Create(ctx context.Context, clinicID uint64, input *CreateVaccinationInput) (*model.Vaccination, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.Vaccination{}, nil
}

func (m *mockVaccinationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccinationInput) (*model.Vaccination, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.Vaccination{}, nil
}

func (m *mockVaccinationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

// The pre-move handler test built a full *service.Services (Vaccination + Pet +
// LstepTagSync) because the handler was a method on *handler.Handler; the moved
// VaccinationHandler only depends on the VaccinationService, so the mockPetService /
// mockLstepTagSyncService fillers are dropped (they exercised no handler code path).
func newHandlerWithVaccinationSvc(svc VaccinationService) *VaccinationHandler {
	return NewVaccinationHandler(svc)
}

// ---- ListVaccinations ----

func TestListVaccinations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockVaccinationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated vaccinations",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				listFn: func(_ context.Context, _ uint64, petID, ownerID *uint64, _, _ *string, _, _ int) ([]model.Vaccination, int64, error) {
					assert.Nil(t, petID)
					assert.Nil(t, ownerID)
					return []model.Vaccination{{ID: 1, Remarks: "混合ワクチン"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"remarks":"混合ワクチン"`,
		},
		{
			name:     "passes pet_id filter to service",
			query:    "page=1&limit=10&pet_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				listFn: func(_ context.Context, _ uint64, petID, _ *uint64, _, _ *string, _, _ int) ([]model.Vaccination, int64, error) {
					require.NotNil(t, petID)
					assert.Equal(t, uint64(5), *petID)
					return []model.Vaccination{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid pet_id",
			query:      "page=1&limit=10&pet_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid pagination",
			query:      "page=0&limit=10",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "page=1&limit=10",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				listFn: func(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.Vaccination, int64, error) {
					return nil, 0, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccinationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListVaccinations(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetVaccination ----

func TestGetVaccination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockVaccinationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns vaccination for valid id",
			paramID:  "10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Vaccination, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					return &model.Vaccination{ID: 10, Remarks: "狂犬病"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"remarks":"狂犬病"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return nil, apperrors.WrapNotFound("vaccination", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccinationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetVaccination(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateVaccination ----

func TestCreateVaccination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	vaccineID := uint64(1)
	validBody := func() map[string]any {
		return map[string]any{
			"vaccine_id": vaccineID,
			"date":       time.Now().Format(time.RFC3339),
			"remarks":    "定期接種",
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockVaccinationService
		wantStatus int
	}{
		{
			name:     "creates vaccination successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				createFn: func(_ context.Context, _ uint64, input *CreateVaccinationInput) (*model.Vaccination, error) {
					assert.Equal(t, "定期接種", input.Remarks)
					return &model.Vaccination{}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required fields are missing",
			body:       map[string]any{"remarks": "テスト"}, // vaccine_id and date missing
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 when date is invalid",
			body: map[string]any{
				"vaccine_id": vaccineID,
				"date":       "not-a-date",
			},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				createFn: func(_ context.Context, _ uint64, _ *CreateVaccinationInput) (*model.Vaccination, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccinationSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateVaccination(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateVaccination ----

func TestUpdateVaccination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockVaccinationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates vaccination successfully",
			paramID:  "10",
			body:     map[string]any{"remarks": "更新済み"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateVaccinationInput) (*model.Vaccination, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					require.NotNil(t, input.Remarks)
					assert.Equal(t, "更新済み", *input.Remarks)
					return &model.Vaccination{ID: 10, Remarks: *input.Remarks}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"remarks":"更新済み"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "10",
			body:       map[string]any{"remarks": "更新済み"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       map[string]any{"remarks": "更新済み"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for malformed json body",
			paramID:    "10",
			body:       `{"remarks":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when date is invalid",
			paramID:    "10",
			body:       map[string]any{"date": "not-a-date"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"remarks": "更新済み"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockVaccinationService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateVaccinationInput) (*model.Vaccination, error) {
					return nil, apperrors.WrapNotFound("vaccination", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVaccinationSvc(tt.svc)
			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateVaccination(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteVaccination ----

func newDeleteVaccinationRouter(svc VaccinationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithVaccinationSvc(svc)
	r.DELETE("/vaccinations/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteVaccination)
	return r
}

func TestDeleteVaccination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockVaccinationService
		wantStatus int
	}{
		{
			name:    "deletes vaccination successfully",
			paramID: "1",
			svc: &mockVaccinationService{
				deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockVaccinationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockVaccinationService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("vaccination", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteVaccinationRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/vaccinations/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithVaccinationSvc(&mockVaccinationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteVaccination(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
