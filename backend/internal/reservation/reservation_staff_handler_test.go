package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestReservationStaffHandlerCompiles verifies reservation_staff_handler.go compiles
func TestReservationStaffHandlerCompiles(t *testing.T) {
	// Retained for backwards compatibility. Real tests follow.
	_ = t
}

// ---- mock ReservationStaffService ----

type mockReservationStaffService struct {
	listFn                   func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	getByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	createFn                 func(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error)
	updateFn                 func(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error)
	deleteFn                 func(ctx context.Context, clinicID, id uint64) error
	patchStatusFn            func(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, []model.StaffReservationExclusion, error)
	patchSortOrderFn         func(ctx context.Context, clinicID, id uint64, direction string) error
	getExcludedFn            func(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error)
	listExcludedByStaffIDsFn func(ctx context.Context, clinicID uint64, staffIDs []uint64) (map[uint64][]model.StaffReservationExclusion, error)
}

func (m *mockReservationStaffService) List(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockReservationStaffService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationStaffService) Create(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return nil, nil, nil
}

func (m *mockReservationStaffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return nil, nil, nil
}

func (m *mockReservationStaffService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationStaffService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, []model.StaffReservationExclusion, error) {
	if m.patchStatusFn != nil {
		return m.patchStatusFn(ctx, clinicID, id, isActive)
	}
	return nil, nil, nil
}

func (m *mockReservationStaffService) PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if m.patchSortOrderFn != nil {
		return m.patchSortOrderFn(ctx, clinicID, id, direction)
	}
	return nil
}

func (m *mockReservationStaffService) GetExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
	if m.getExcludedFn != nil {
		return m.getExcludedFn(ctx, staffID)
	}
	return nil, nil
}

func (m *mockReservationStaffService) ListExcludedByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) (map[uint64][]model.StaffReservationExclusion, error) {
	if m.listExcludedByStaffIDsFn != nil {
		return m.listExcludedByStaffIDsFn(ctx, clinicID, staffIDs)
	}
	return map[uint64][]model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffService) ListCapableByStaffIDs(_ context.Context, _ uint64, _ []uint64) (map[uint64][]model.StaffReservationCapability, error) {
	return map[uint64][]model.StaffReservationCapability{}, nil
}

// ---- test helper ----

func newHandlerWithReservationStaffSvc(svc ReservationStaffService) *ReservationStaffHandler {
	return NewReservationStaffHandler(svc)
}

// ---- ListReservationStaffs ----

func TestListReservationStaffs_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationStaffService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with reservation staff list",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.Staff, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.Staff{{ID: 5, Name: "予約スタッフA", StaffType: model.StaffTypeDoctor}}, nil
				},
				listExcludedByStaffIDsFn: func(_ context.Context, clinicID uint64, staffIDs []uint64) (map[uint64][]model.StaffReservationExclusion, error) {
					assert.Equal(t, uint64(1), clinicID)
					return map[uint64][]model.StaffReservationExclusion{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"予約スタッフA"`,
		},
		{
			name:     "returns 200 with empty list when no reservation staff",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				listFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{}, nil
				},
				listExcludedByStaffIDsFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]model.StaffReservationExclusion, error) {
					return map[uint64][]model.StaffReservationExclusion{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				listFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 when ListExcludedByStaffIDs fails",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				listFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return []model.Staff{{ID: 5, Name: "予約スタッフA"}}, nil
				},
				listExcludedByStaffIDsFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]model.StaffReservationExclusion, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationStaffSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListReservationStaffs(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateReservationStaff ----

func TestCreateReservationStaff_Valid_Returns201(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationStaffService
		wantStatus int
		wantBody   string
	}{
		{
			name: "creates reservation staff successfully returns 201",
			body: map[string]any{
				"name":                "新予約スタッフ",
				"staff_type":          "doctor",
				"reservation_visible": true,
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "新予約スタッフ", input.Name)
					return &model.Staff{ID: 30, Name: input.Name, StaffType: model.StaffTypeDoctor}, nil, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"新予約スタッフ"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     map[string]any{"name": "エラースタッフ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				createFn: func(_ context.Context, _ uint64, _ *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
					return nil, nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationStaffSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateReservationStaff(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteReservationStaff ----

func newDeleteReservationStaffRouter(svc ReservationStaffService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationStaffSvc(svc)
	r.DELETE("/reservation-staffs/:staffId", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteReservationStaff)
	return r
}

func TestDeleteReservationStaff_NotFound_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockReservationStaffService
		wantStatus int
	}{
		{
			name:    "deletes reservation staff successfully returns 204",
			paramID: "1",
			svc: &mockReservationStaffService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 404 when reservation staff not found",
			paramID: "999",
			svc: &mockReservationStaffService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("staff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteReservationStaffRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/reservation-staffs/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithReservationStaffSvc(&mockReservationStaffService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "staffId", Value: "1"}}
		h.DeleteReservationStaff(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- UpdateReservationStaff ----

func TestUpdateReservationStaff(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nameVal := "更新スタッフ名"

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationStaffService
		wantStatus int
	}{
		{
			name:     "updates reservation staff successfully returns 200",
			paramID:  "3",
			body:     map[string]any{"name": nameVal},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, nameVal, *input.Name)
					return &model.Staff{ID: 3, Name: *input.Name}, nil, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "3",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 404 when reservation staff not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
					return nil, nil, apperrors.WrapNotFound("staff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "3",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationStaffSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "staffId", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateReservationStaff(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateReservationStaffStatus ----

func TestPatchReservationStaffStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationStaffService
		wantStatus int
	}{
		{
			name:     "activates reservation staff successfully",
			paramID:  "3",
			body:     map[string]any{"is_active": true},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				patchStatusFn: func(_ context.Context, clinicID, id uint64, isActive bool) (*model.Staff, []model.StaffReservationExclusion, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					assert.True(t, isActive)
					return &model.Staff{ID: 3, IsActive: isActive}, nil, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "3",
			body:       map[string]any{"is_active": false},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "3",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when reservation staff not found",
			paramID:  "999",
			body:     map[string]any{"is_active": true},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationStaffService{
				patchStatusFn: func(_ context.Context, _, _ uint64, _ bool) (*model.Staff, []model.StaffReservationExclusion, error) {
					return nil, nil, apperrors.WrapNotFound("staff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationStaffSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "staffId", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateReservationStaffStatus(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateReservationStaffSortOrder ----
//
// c.Status(http.StatusNoContent) は httptest.ResponseRecorder に即時反映されないため
// gin.Engine 経由でリクエストを送る。

func newPatchSortOrderRouter(svc ReservationStaffService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationStaffSvc(svc)
	r.PATCH("/reservation-staffs/:staffId/sort-order", func(c *gin.Context) {
		setClinicID(c)
	}, h.UpdateReservationStaffSortOrder)
	return r
}

func TestPatchReservationStaffSortOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       string
		svc        *mockReservationStaffService
		wantStatus int
	}{
		{
			name:    "moves sort order up successfully",
			paramID: "3",
			body:    `{"direction":"up"}`,
			svc: &mockReservationStaffService{
				patchSortOrderFn: func(_ context.Context, clinicID, id uint64, direction string) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					assert.Equal(t, "up", direction)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "moves sort order down successfully",
			paramID: "3",
			body:    `{"direction":"down"}`,
			svc: &mockReservationStaffService{
				patchSortOrderFn: func(_ context.Context, _, _ uint64, direction string) error {
					assert.Equal(t, "down", direction)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for invalid direction",
			paramID:    "3",
			body:       `{"direction":"sideways"}`,
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "3",
			body:       `{invalid}`,
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       `{"direction":"up"}`,
			svc:        &mockReservationStaffService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when reservation staff not found",
			paramID: "999",
			body:    `{"direction":"up"}`,
			svc: &mockReservationStaffService{
				patchSortOrderFn: func(_ context.Context, _, _ uint64, _ string) error {
					return apperrors.WrapNotFound("staff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchSortOrderRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/reservation-staffs/"+tt.paramID+"/sort-order", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithReservationStaffSvc(&mockReservationStaffService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"direction":"up"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "staffId", Value: "3"}}
		h.UpdateReservationStaffSortOrder(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- UploadReservationStaffImage ----
//
// v2 スコープ：未実装（RespondError で 501 Not Implemented を返すのみ）。

func TestUploadReservationStaffImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newHandlerWithReservationStaffSvc(&mockReservationStaffService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)

	h.UploadReservationStaffImage(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}
