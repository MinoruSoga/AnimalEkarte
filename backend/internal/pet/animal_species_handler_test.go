package pet

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

func setSystemAdminAuth(c *gin.Context) {
	c.Set("is_system_admin", true)
	c.Set("clinic_id", "1")
	c.Set("user_id", "42")
}

// ---- mock AnimalSpeciesService ----

type mockAnimalSpeciesService struct {
	listFn    func(ctx context.Context) ([]model.AnimalSpecies, error)
	getByIDFn func(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	createFn  func(ctx context.Context, input *CreateAnimalSpeciesInput) (*model.AnimalSpecies, error)
	updateFn  func(ctx context.Context, id uint64, input *UpdateAnimalSpeciesInput) (*model.AnimalSpecies, error)
	deleteFn  func(ctx context.Context, id uint64) error
	reorderFn func(ctx context.Context, ids []uint64) error
}

func (m *mockAnimalSpeciesService) List(ctx context.Context) ([]model.AnimalSpecies, error) {
	return m.listFn(ctx)
}

func (m *mockAnimalSpeciesService) GetByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockAnimalSpeciesService) Create(ctx context.Context, input *CreateAnimalSpeciesInput, _ AnimalSpeciesMutationMeta) (*model.AnimalSpecies, error) {
	return m.createFn(ctx, input)
}

func (m *mockAnimalSpeciesService) Update(ctx context.Context, id uint64, input *UpdateAnimalSpeciesInput, _ AnimalSpeciesMutationMeta) (*model.AnimalSpecies, error) {
	return m.updateFn(ctx, id, input)
}

func (m *mockAnimalSpeciesService) Delete(ctx context.Context, id uint64, _ AnimalSpeciesMutationMeta) error {
	return m.deleteFn(ctx, id)
}

func (m *mockAnimalSpeciesService) Reorder(ctx context.Context, ids []uint64, _ AnimalSpeciesMutationMeta) error {
	return m.reorderFn(ctx, ids)
}

// ---- test helper ----

func newHandlerWithAnimalSpeciesSvc(svc AnimalSpeciesService) *Handler {
	return NewHandler(nil, svc, nil, allowAllPermission)
}

// ---- ListAnimalSpecies ----

func TestListAnimalSpecies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		svc        *mockAnimalSpeciesService
		wantStatus int
		wantBody   string
	}{
		{
			name: "returns list of animal species",
			svc: &mockAnimalSpeciesService{
				listFn: func(_ context.Context) ([]model.AnimalSpecies, error) {
					return []model.AnimalSpecies{{ID: 1, Name: "犬"}, {ID: 2, Name: "猫"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"犬"`,
		},
		{
			name: "returns empty list when no species exist",
			svc: &mockAnimalSpeciesService{
				listFn: func(_ context.Context) ([]model.AnimalSpecies, error) {
					return []model.AnimalSpecies{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name: "returns 500 on service error",
			svc: &mockAnimalSpeciesService{
				listFn: func(_ context.Context) ([]model.AnimalSpecies, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAnimalSpeciesSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)

			h.ListAnimalSpecies(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetAnimalSpecies ----

func TestGetAnimalSpecies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockAnimalSpeciesService
		wantStatus int
		wantBody   string
	}{
		{
			name:    "returns animal species for valid id",
			paramID: "3",
			svc: &mockAnimalSpeciesService{
				getByIDFn: func(_ context.Context, id uint64) (*model.AnimalSpecies, error) {
					assert.Equal(t, uint64(3), id)
					return &model.AnimalSpecies{ID: 3, Name: "ウサギ"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"ウサギ"`,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockAnimalSpeciesService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when species not found",
			paramID: "999",
			svc: &mockAnimalSpeciesService{
				getByIDFn: func(_ context.Context, _ uint64) (*model.AnimalSpecies, error) {
					return nil, apperrors.WrapNotFound("animal_species", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 500 on service error",
			paramID: "1",
			svc: &mockAnimalSpeciesService{
				getByIDFn: func(_ context.Context, _ uint64) (*model.AnimalSpecies, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAnimalSpeciesSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}

			h.GetAnimalSpecies(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateAnimalSpecies ----

func TestCreateAnimalSpecies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":       "鳥",
			"is_active":  true,
			"sort_order": 5,
		}
	}

	tests := []struct {
		name         string
		body         any
		svc          *mockAnimalSpeciesService
		wantStatus   int
		wantBody     string
		wantLocation string
	}{
		{
			name: "creates animal species successfully",
			body: validBody(),
			svc: &mockAnimalSpeciesService{
				createFn: func(_ context.Context, input *CreateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
					assert.Equal(t, "鳥", input.Name)
					assert.True(t, input.IsActive)
					assert.Equal(t, 5, input.SortOrder)
					return &model.AnimalSpecies{ID: 10, Name: input.Name, IsActive: input.IsActive, SortOrder: input.SortOrder}, nil
				},
			},
			wantStatus:   http.StatusCreated,
			wantBody:     `"name":"鳥"`,
			wantLocation: "/api/v1/masters/animal-species/10",
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			svc:        &mockAnimalSpeciesService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			svc:        &mockAnimalSpeciesService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: validBody(),
			svc: &mockAnimalSpeciesService{
				createFn: func(_ context.Context, _ *CreateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAnimalSpeciesSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			setSystemAdminAuth(c)

			h.CreateAnimalSpecies(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantLocation != "" {
				assert.Equal(t, tt.wantLocation, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateAnimalSpecies ----

func TestUpdateAnimalSpecies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		svc        *mockAnimalSpeciesService
		wantStatus int
		wantBody   string
	}{
		{
			name:    "updates animal species successfully",
			paramID: "1",
			body:    map[string]any{"name": "爬虫類"},
			svc: &mockAnimalSpeciesService{
				updateFn: func(_ context.Context, id uint64, input *UpdateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "爬虫類", *input.Name)
					return &model.AnimalSpecies{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"爬虫類"`,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			svc:        &mockAnimalSpeciesService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when species not found",
			paramID: "999",
			body:    map[string]any{"name": "テスト"},
			svc: &mockAnimalSpeciesService{
				updateFn: func(_ context.Context, _ uint64, _ *UpdateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
					return nil, apperrors.WrapNotFound("animal_species", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 500 on service error",
			paramID: "1",
			body:    map[string]any{"name": "テスト"},
			svc: &mockAnimalSpeciesService{
				updateFn: func(_ context.Context, _ uint64, _ *UpdateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAnimalSpeciesSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			setSystemAdminAuth(c)

			h.UpdateAnimalSpecies(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteAnimalSpecies ----

func newDeleteAnimalSpeciesRouter(svc AnimalSpeciesService) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) { setSystemAdminAuth(c); c.Next() })
	h := newHandlerWithAnimalSpeciesSvc(svc)
	r.DELETE("/animal-species/:id", h.DeleteAnimalSpecies)
	return r
}

func TestDeleteAnimalSpecies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockAnimalSpeciesService
		wantStatus int
	}{
		{
			name:    "deletes animal species successfully",
			paramID: "2",
			svc: &mockAnimalSpeciesService{
				deleteFn: func(_ context.Context, id uint64) error {
					assert.Equal(t, uint64(2), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockAnimalSpeciesService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when species not found",
			paramID: "999",
			svc: &mockAnimalSpeciesService{
				deleteFn: func(_ context.Context, _ uint64) error {
					return apperrors.WrapNotFound("animal_species", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when species is in use",
			paramID: "1",
			svc: &mockAnimalSpeciesService{
				deleteFn: func(_ context.Context, _ uint64) error {
					return apperrors.WrapConflict("この動物種はペット情報で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteAnimalSpeciesRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/animal-species/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- ReorderAnimalSpecies ----

func newReorderAnimalSpeciesRouter(svc AnimalSpeciesService) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) { setSystemAdminAuth(c); c.Next() })
	h := newHandlerWithAnimalSpeciesSvc(svc)
	r.POST("/animal-species/reorder", h.ReorderAnimalSpecies)
	return r
}

func TestReorderAnimalSpecies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		svc        *mockAnimalSpeciesService
		wantStatus int
	}{
		{
			name: "reorders animal species successfully",
			body: map[string]any{"ids": []int{3, 1, 2}},
			svc: &mockAnimalSpeciesService{
				reorderFn: func(_ context.Context, ids []uint64) error {
					assert.Equal(t, []uint64{3, 1, 2}, ids)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			svc:        &mockAnimalSpeciesService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: map[string]any{"ids": []int{1, 2}},
			svc: &mockAnimalSpeciesService{
				reorderFn: func(_ context.Context, _ []uint64) error {
					return fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newReorderAnimalSpeciesRouter(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/animal-species/reorder", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestCreateAnimalSpecies_RejectsNonSystemAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithAnimalSpeciesSvc(&mockAnimalSpeciesService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(map[string]any{"name": "x", "is_active": true, "sort_order": 1})
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("is_system_admin", false)
	c.Set("clinic_id", "1")
	c.Set("user_id", "42")
	h.CreateAnimalSpecies(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
