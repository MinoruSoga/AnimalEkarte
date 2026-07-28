package medicalrecord

// medicine_dose_param_handler_test.go — #201 B-2c: dose パラメータ authoring API の handler 配線テスト。
// status code + DTO contract + path param 検証 + ルートツリー共存（:id と :id/dose-params/:species）を固定する。

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock MedicineDoseParamService ----

type mockMedicineDoseParamService struct {
	listFn   func(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error)
	upsertFn func(ctx context.Context, clinicID, medicineID uint64, input *MedicineDoseParamInput, actorID *uint64) (*model.MedicineDoseParam, error)
	deleteFn func(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies, actorID *uint64) error
}

func (m *mockMedicineDoseParamService) List(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error) {
	if m.listFn == nil {
		return nil, nil
	}
	return m.listFn(ctx, clinicID, medicineID)
}
func (m *mockMedicineDoseParamService) Upsert(ctx context.Context, clinicID, medicineID uint64, input *MedicineDoseParamInput, actorID *uint64) (*model.MedicineDoseParam, error) {
	if m.upsertFn == nil {
		return nil, nil
	}
	return m.upsertFn(ctx, clinicID, medicineID, input, actorID)
}
func (m *mockMedicineDoseParamService) Delete(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies, actorID *uint64) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, clinicID, medicineID, species, actorID)
}

func newHandlerWithDoseParamSvc(svc MedicineDoseParamService) *MedicineDoseParamHandler {
	return NewMedicineDoseParamHandler(svc)
}

// ---- ListMedicineDoseParams ----

func TestListMedicineDoseParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockMedicineDoseParamService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns dose params",
			paramID:  "50",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineDoseParamService{
				listFn: func(_ context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(50), medicineID)
					return []model.MedicineDoseParam{{ID: 1, MedicineID: 50, Species: model.MedicineDoseSpeciesDog, DosePerKg: 5}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"species":"dog"`,
		},
		{
			name:       "401 when clinic missing",
			paramID:    "50",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicineDoseParamService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicineDoseParamService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "404 when medicine not owned",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineDoseParamService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.MedicineDoseParam, error) {
					return nil, apperrors.WrapNotFound("medicine", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDoseParamSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ListMedicineDoseParams(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpsertMedicineDoseParam ----

func TestUpsertMedicineDoseParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"dose_per_kg": 5.0, "max_mg_per_kg": 10.0}
	}

	tests := []struct {
		name       string
		paramID    string
		species    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockMedicineDoseParamService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "upserts dog param",
			paramID:  "50",
			species:  "dog",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineDoseParamService{
				upsertFn: func(_ context.Context, clinicID, medicineID uint64, input *MedicineDoseParamInput, _ *uint64) (*model.MedicineDoseParam, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(50), medicineID)
					assert.Equal(t, model.MedicineDoseSpeciesDog, input.Species)
					assert.Equal(t, 5.0, input.DosePerKg)
					return &model.MedicineDoseParam{ID: 7, MedicineID: 50, Species: input.Species, DosePerKg: input.DosePerKg, MaxMgPerKg: input.MaxMgPerKg}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"species":"dog"`,
		},
		{
			name:       "401 when clinic missing",
			paramID:    "50",
			species:    "dog",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicineDoseParamService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400 for non-numeric id",
			paramID:    "abc",
			species:    "dog",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicineDoseParamService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 for invalid species",
			paramID:    "50",
			species:    "rabbit",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicineDoseParamService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 when dose_per_kg missing (binding)",
			paramID:    "50",
			species:    "dog",
			body:       map[string]any{"max_mg_per_kg": 10.0},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicineDoseParamService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "400 when service rejects (validation)",
			paramID:  "50",
			species:  "dog",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineDoseParamService{
				upsertFn: func(_ context.Context, _, _ uint64, _ *MedicineDoseParamInput, _ *uint64) (*model.MedicineDoseParam, error) {
					return nil, apperrors.WrapInvalidInput("この薬剤は per_weight 計算ではないため投与量パラメータを設定できません")
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "404 when medicine not owned",
			paramID:  "999",
			species:  "dog",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineDoseParamService{
				upsertFn: func(_ context.Context, _, _ uint64, _ *MedicineDoseParamInput, _ *uint64) (*model.MedicineDoseParam, error) {
					return nil, apperrors.WrapNotFound("medicine", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDoseParamSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}, {Key: "species", Value: tt.species}}
			tt.setupCtx(c)

			h.UpsertMedicineDoseParam(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteMedicineDoseParam (route-based: path params + tree coexistence) ----

func newDoseParamRouter(svc MedicineDoseParamService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithDoseParamSvc(svc)
	setCtx := func(c *gin.Context) { setClinicID(c) }
	// 実 master_routes と同じツリー形状で登録し、:id と :id/dose-params/:species の共存を検証する。
	r.GET("/medicines/:id", setCtx, func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/medicines/:id/dose-params", setCtx, h.ListMedicineDoseParams)
	r.PUT("/medicines/:id/dose-params/:species", setCtx, h.UpsertMedicineDoseParam)
	r.DELETE("/medicines/:id/dose-params/:species", setCtx, h.DeleteMedicineDoseParam)
	return r
}

func TestDeleteMedicineDoseParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		svc        *mockMedicineDoseParamService
		wantStatus int
	}{
		{
			name: "deletes successfully",
			path: "/medicines/50/dose-params/cat",
			svc: &mockMedicineDoseParamService{
				deleteFn: func(_ context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies, _ *uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(50), medicineID)
					assert.Equal(t, model.MedicineDoseSpeciesCat, species)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "400 for invalid species",
			path:       "/medicines/50/dose-params/rabbit",
			svc:        &mockMedicineDoseParamService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "404 when not found",
			path: "/medicines/50/dose-params/dog",
			svc: &mockMedicineDoseParamService{
				deleteFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies, _ *uint64) error {
					return apperrors.WrapNotFound("medicine_dose_param", "")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDoseParamRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, tt.path, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestMedicineDoseParamRouteTreeCoexistence は :id と :id/dose-params/:species が
// 同一 Gin ツリーで panic なく共存し、正しいハンドラへ解決されることを検証する。
func TestMedicineDoseParamRouteTreeCoexistence(t *testing.T) {
	gin.SetMode(gin.TestMode)

	listCalled := false
	svc := &mockMedicineDoseParamService{
		listFn: func(_ context.Context, _, medicineID uint64) ([]model.MedicineDoseParam, error) {
			listCalled = true
			assert.Equal(t, uint64(50), medicineID)
			return []model.MedicineDoseParam{}, nil
		},
	}
	router := newDoseParamRouter(svc)

	// GET /medicines/:id は別ハンドラ（200・list 未呼び出し）。
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/medicines/50", http.NoBody))
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.False(t, listCalled, "/medicines/:id は dose-params list を呼ばない")

	// GET /medicines/:id/dose-params は list ハンドラへ解決。
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/medicines/%d/dose-params", 50), http.NoBody))
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.True(t, listCalled, "/medicines/:id/dose-params は list ハンドラへ解決される")
}
