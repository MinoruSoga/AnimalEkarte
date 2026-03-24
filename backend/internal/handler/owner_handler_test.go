package handler

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

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock OwnerService ----

type mockOwnerService struct {
	listFn           func(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
	getByIDFn        func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	createWithPetsFn func(ctx context.Context, clinicID uint64, input *service.CreateOwnerInput) (*model.Owner, error)
	updateFn         func(ctx context.Context, clinicID, id uint64, input *service.UpdateOwnerInput) (*model.Owner, error)
	deleteFn         func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockOwnerService) List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	return m.listFn(ctx, clinicID, page, limit, search)
}

func (m *mockOwnerService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockOwnerService) CreateWithPets(ctx context.Context, clinicID uint64, input *service.CreateOwnerInput) (*model.Owner, error) {
	return m.createWithPetsFn(ctx, clinicID, input)
}

func (m *mockOwnerService) Update(ctx context.Context, clinicID, id uint64, input *service.UpdateOwnerInput) (*model.Owner, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockOwnerService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

// ---- test helpers ----

func newHandlerWithOwnerSvc(svc service.OwnerService) *Handler {
	return &Handler{
		svc: &service.Services{Owner: svc},
	}
}

// setClinicID は gin.Context に clinic_id を設定するヘルパー
func setClinicID(c *gin.Context, id string) {
	c.Set("clinic_id", id)
}

// ---- ListOwners ----

func TestListOwners(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockOwnerService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated owners",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
					return []model.Owner{{ID: 1, OwnerName: "田中太郎"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_name":"田中太郎"`,
		},
		{
			name:     "returns empty list when no owners",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, _ uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					return []model.Owner{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"total":0`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c, "1") },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "passes search param to service",
			query:    "page=1&limit=10&search=田中",
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, _ uint64, _, _ int, search string) ([]model.Owner, int64, error) {
					assert.Equal(t, "田中", search)
					return []model.Owner{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, _ uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					return nil, 0, fmt.Errorf("unexpected db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOwnerSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListOwners(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetOwner ----

func TestGetOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockOwnerService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns owner for valid id",
			paramID:  "42",
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Owner, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(42), id)
					return &model.Owner{ID: 42, OwnerName: "佐藤花子"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_name":"佐藤花子"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c, "1") },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when owner not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOwnerSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetOwner(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateOwner ----

func TestCreateOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"owner_name":      "山田一郎",
			"owner_name_kana": "ヤマダイチロウ",
			"email":           "yamada@example.com",
			"phone":           "090-1234-5678",
			"membership_type": "general",
			"pets":            []any{},
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockOwnerService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates owner successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				createWithPetsFn: func(_ context.Context, _ uint64, input *service.CreateOwnerInput) (*model.Owner, error) {
					return &model.Owner{ID: 1, OwnerName: input.OwnerName}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"owner_name":"山田一郎"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "invalid json",
			setupCtx:   func(c *gin.Context) { setClinicID(c, "1") },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on duplicate email",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				createWithPetsFn: func(_ context.Context, _ uint64, _ *service.CreateOwnerInput) (*model.Owner, error) {
					return nil, apperrors.WrapAlreadyExists("owner", "yamada@example.com")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOwnerSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateOwner(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateOwner ----

func TestUpdateOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ptrStr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockOwnerService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates owner successfully",
			paramID:  "1",
			body:     map[string]any{"owner_name": "田中次郎"},
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				updateFn: func(_ context.Context, _, _ uint64, input *service.UpdateOwnerInput) (*model.Owner, error) {
					return &model.Owner{ID: 1, OwnerName: *input.OwnerName}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_name":"田中次郎"`,
		},
		{
			name:     "partial update with only phone",
			paramID:  "1",
			body:     map[string]any{"phone": "080-9999-0000"},
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				updateFn: func(_ context.Context, _, _ uint64, input *service.UpdateOwnerInput) (*model.Owner, error) {
					assert.Equal(t, "080-9999-0000", *input.Phone)
					assert.Nil(t, input.OwnerName)
					return &model.Owner{ID: 1, Phone: *input.Phone}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"owner_name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c, "1") },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when owner not found",
			paramID:  "999",
			body:     map[string]any{"owner_name": ptrStr("テスト")},
			setupCtx: func(c *gin.Context) { setClinicID(c, "1") },
			svc: &mockOwnerService{
				updateFn: func(_ context.Context, _, _ uint64, _ *service.UpdateOwnerInput) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOwnerSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateOwner(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteOwner ----
//
// c.Status(http.StatusNoContent) は Gin の ResponseWriter にステータスをバッファするだけで
// httptest.ResponseRecorder には即時書き込まれない。
// そのため NoContent 系レスポンスは gin.Engine 経由でリクエストを送る。

func newDeleteOwnerRouter(svc service.OwnerService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOwnerSvc(svc)
	r.DELETE("/owners/:id", func(c *gin.Context) {
		setClinicID(c, "1")
	}, h.DeleteOwner)
	return r
}

func TestDeleteOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockOwnerService
		wantStatus int
	}{
		{
			name:    "deletes owner successfully",
			paramID: "1",
			svc: &mockOwnerService{
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
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when owner not found",
			paramID: "999",
			svc: &mockOwnerService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteOwnerRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/owners/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	// clinic_id なしのケースは CreateTestContext で直接テスト
	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithOwnerSvc(&mockOwnerService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteOwner(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
