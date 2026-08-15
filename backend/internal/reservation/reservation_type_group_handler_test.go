package reservation

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

// ---- mock ReservationTypeGroupService ----

type mockReservationTypeGroupService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockReservationTypeGroupService) List(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockReservationTypeGroupService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeGroup{ID: id}, nil
}

func (m *mockReservationTypeGroupService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.ReservationTypeGroup{ID: 1, Name: input.Name}, nil
}

func (m *mockReservationTypeGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.ReservationTypeGroup{ID: id}, nil
}

func (m *mockReservationTypeGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationTypeGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- helper ----

func newHandlerWithReservationTypeGroupSvc(svc ReservationTypeGroupService) *ReservationTypeGroupHandler {
	return NewReservationTypeGroupHandler(svc)
}

// ---- ListReservationTypeGroups ----

func TestListReservationTypeGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeGroupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of groups",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.ReservationTypeGroup{{ID: 1, Name: "健診グループ"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"健診グループ"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeGroupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				listFn: func(_ context.Context, _ uint64) ([]model.ReservationTypeGroup, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeGroupSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.ListReservationTypeGroups(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetReservationTypeGroup ----

func TestGetReservationTypeGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeGroupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns group for valid id",
			paramID:  "2",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				getByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationTypeGroup, error) {
					return &model.ReservationTypeGroup{ID: id, Name: "ワクチングループ"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"ワクチングループ"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeGroupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeGroupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationTypeGroup, error) {
					return nil, apperrors.WrapNotFound("reservation_type_group", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeGroupSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetReservationTypeGroup(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateReservationTypeGroup ----

func TestCreateReservationTypeGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"name": "新グループ"}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeGroupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates group successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "新グループ", input.Name)
					return &model.ReservationTypeGroup{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"新グループ"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeGroupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"color": "#FF0000"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeGroupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on conflict",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				createFn: func(_ context.Context, _ uint64, _ *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
					return nil, apperrors.WrapConflict("グループ名が重複しています")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				createFn: func(_ context.Context, _ uint64, _ *CreateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeGroupSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateReservationTypeGroup(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateReservationTypeGroup ----

func TestUpdateReservationTypeGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeGroupService
		wantStatus int
	}{
		{
			name:     "updates group successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新済みグループ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				updateFn: func(_ context.Context, _, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新済みグループ", *input.Name)
					return &model.ReservationTypeGroup{ID: id}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeGroupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeGroupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "test"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
					return nil, apperrors.WrapNotFound("reservation_type_group", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 on empty body (no fields to update)",
			paramID:  "1",
			body:     map[string]any{},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeGroupService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
					return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeGroupSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateReservationTypeGroup(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteReservationTypeGroup ----

func newDeleteReservationTypeGroupRouter(svc ReservationTypeGroupService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationTypeGroupSvc(svc)
	r.DELETE("/reservation-type-groups/:id", func(c *gin.Context) { setClinicID(c) }, h.DeleteReservationTypeGroup)
	return r
}

func TestDeleteReservationTypeGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockReservationTypeGroupService
		wantStatus int
	}{
		{
			name:    "deletes group successfully",
			paramID: "1",
			svc: &mockReservationTypeGroupService{
				deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockReservationTypeGroupService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("reservation_type_group", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when group is in use",
			paramID: "2",
			svc: &mockReservationTypeGroupService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このグループは使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteReservationTypeGroupRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/reservation-type-groups/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeGroupSvc(&mockReservationTypeGroupService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteReservationTypeGroup(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderReservationTypeGroups ----

func newReorderReservationTypeGroupsRouter(svc ReservationTypeGroupService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationTypeGroupSvc(svc)
	r.POST("/reservation-type-groups/reorder", func(c *gin.Context) { setClinicID(c) }, h.ReorderReservationTypeGroups)
	return r
}

func TestReorderReservationTypeGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders successfully", func(t *testing.T) {
		svc := &mockReservationTypeGroupService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 1, 3}, ids)
				return nil
			},
		}
		router := newReorderReservationTypeGroupsRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{2, 1, 3}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/reservation-type-groups/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeGroupSvc(&mockReservationTypeGroupService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderReservationTypeGroups(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
