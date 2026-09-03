package medicalrecord

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

// ---- mock ProcedureService ----

type mockProcedureService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.Procedure, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockProcedureService) List(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	return m.listFn(ctx, clinicID)
}

func (m *mockProcedureService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockProcedureService) Create(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.Procedure{}, nil
}

func (m *mockProcedureService) Update(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockProcedureService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockProcedureService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- test helper ----

func newHandlerWithProcedureSvc(svc ProcedureService) *ProcedureHandler {
	return NewProcedureHandler(svc)
}

// ---- ListProcedures ----

func TestListProcedures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockProcedureService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns all procedures",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.Procedure, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.Procedure{{ID: 1, Name: "去勢手術"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"去勢手術"`,
		},
		{
			name:     "returns empty list when no procedures",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				listFn: func(_ context.Context, _ uint64) ([]model.Procedure, error) {
					return []model.Procedure{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "[]",
		},
		{
			name:     "returns procedures with hierarchy (parent_id)",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				listFn: func(_ context.Context, _ uint64) ([]model.Procedure, error) {
					parentID := uint64(1)
					return []model.Procedure{
						{ID: 1, Name: "外科手術"},
						{ID: 2, Name: "去勢手術", ParentID: &parentID},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"外科手術"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockProcedureService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				listFn: func(_ context.Context, _ uint64) ([]model.Procedure, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithProcedureSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListProcedures(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetProcedure ----

func TestGetProcedure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockProcedureService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns procedure for valid id",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Procedure, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return &model.Procedure{ID: 5, Name: "避妊手術"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"避妊手術"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockProcedureService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when procedure not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Procedure, error) {
					return nil, apperrors.WrapNotFound("procedure", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithProcedureSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetProcedure(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateProcedure ----

func TestCreateProcedure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":       "歯石除去",
			"anesthesia": "none",
			"tax_type":   "excluded",
			"is_active":  true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockProcedureService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates procedure successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "歯石除去", input.Name)
					assert.Equal(t, "none", input.Anesthesia)
					return &model.Procedure{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"歯石除去"`,
		},
		{
			name: "creates procedure with parent_id (hierarchy)",
			body: map[string]any{
				"name":       "去勢手術",
				"anesthesia": "general",
				"tax_type":   "excluded",
				"parent_id":  10,
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				createFn: func(_ context.Context, _ uint64, input *CreateProcedureInput) (*model.Procedure, error) {
					require.NotNil(t, input.ParentID)
					assert.Equal(t, uint64(10), *input.ParentID)
					return &model.Procedure{ID: 2, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockProcedureService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required name is missing",
			body:       map[string]any{"anesthesia": "none", "tax_type": "excluded"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when required anesthesia is missing",
			body:       map[string]any{"name": "テスト処置", "tax_type": "excluded"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when required tax_type is missing",
			body:       map[string]any{"name": "テスト処置", "anesthesia": "none"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when anesthesia value is invalid",
			body:       map[string]any{"name": "テスト", "anesthesia": "invalid", "tax_type": "excluded"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 when procedure name conflicts",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				createFn: func(_ context.Context, _ uint64, _ *CreateProcedureInput) (*model.Procedure, error) {
					return nil, apperrors.WrapAlreadyExists("procedure", "歯石除去")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				createFn: func(_ context.Context, _ uint64, _ *CreateProcedureInput) (*model.Procedure, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithProcedureSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateProcedure(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateProcedure ----

func TestUpdateProcedure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockProcedureService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates procedure name successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新済み処置"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateProcedureInput) (*model.Procedure, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新済み処置", *input.Name)
					return &model.Procedure{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"更新済み処置"`,
		},
		{
			name:    "clears parent_id with clear_parent_id flag",
			paramID: "1",
			body:    map[string]any{"clear_parent_id": true},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
			},
			svc: &mockProcedureService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateProcedureInput) (*model.Procedure, error) {
					assert.True(t, input.ClearParentID)
					return &model.Procedure{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockProcedureService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when procedure not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateProcedureInput) (*model.Procedure, error) {
					return nil, apperrors.WrapNotFound("procedure", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 for empty update body (BUG-397)",
			paramID:  "1",
			body:     map[string]any{},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockProcedureService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateProcedureInput) (*model.Procedure, error) {
					return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid anesthesia value",
			paramID:    "1",
			body:       map[string]any{"anesthesia": "invalid_type"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithProcedureSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateProcedure(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteProcedure ----

func newDeleteProcedureRouter(svc ProcedureService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithProcedureSvc(svc)
	r.DELETE("/procedures/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteProcedure)
	return r
}

func TestDeleteProcedure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockProcedureService
		wantStatus int
	}{
		{
			name:    "deletes procedure successfully",
			paramID: "1",
			svc: &mockProcedureService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when procedure not found",
			paramID: "999",
			svc: &mockProcedureService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("procedure", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when procedure has child procedures (BUG-390)",
			paramID: "5",
			svc: &mockProcedureService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この処置は子処置が存在するため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:    "returns 409 when procedure is used in medical records",
			paramID: "3",
			svc: &mockProcedureService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この診療項目は診療記録で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteProcedureRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/procedures/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithProcedureSvc(&mockProcedureService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteProcedure(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderProcedures ----

func newReorderProceduresRouter(svc ProcedureService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithProcedureSvc(svc)
	r.PUT("/procedures/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderProcedures)
	return r
}

func TestReorderProcedures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		svc        *mockProcedureService
		wantStatus int
	}{
		{
			name: "reorders procedures successfully",
			body: map[string]any{"ids": []int{2, 1, 3}},
			svc: &mockProcedureService{
				reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, []uint64{2, 1, 3}, ids)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for missing ids field",
			body:       map[string]any{},
			svc:        &mockProcedureService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newReorderProceduresRouter(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/procedures/reorder", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithProcedureSvc(&mockProcedureService{})

		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderProcedures(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestProcedureSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*ProcedureHandler, *gin.Context)
		svc    *mockProcedureService
	}{
		{
			name: "ListProcedures returns 403 when selected clinic lacks master-medical view grant",
			invoke: func(h *ProcedureHandler, c *gin.Context) {
				h.ListProcedures(c)
			},
			svc: &mockProcedureService{
				listFn: func(_ context.Context, _ uint64) ([]model.Procedure, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "GetProcedure returns 403 when selected clinic lacks master-medical view grant",
			invoke: func(h *ProcedureHandler, c *gin.Context) {
				h.GetProcedure(c)
			},
			svc: &mockProcedureService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Procedure, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithProcedureSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMasterMedical), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
