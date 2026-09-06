package staff_test

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
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
)

// TestStaffHandlerCompiles verifies staff_handler.go compiles
func TestStaffHandlerCompiles(t *testing.T) {
	// Retained for backwards compatibility. Real tests follow.
	_ = t
}

// ---- mock Service ----

type mockService struct {
	listFn                        func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	getByIDFn                     func(ctx context.Context, id uint64) (*model.Staff, error)
	getByIDInClinicFn             func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	findByAccountIDFn             func(ctx context.Context, accountID uint64) (*model.Staff, error)
	createFn                      func(ctx context.Context, input *staffdomain.CreateStaffInput) (*model.Staff, error)
	createWithAccountFn           func(ctx context.Context, input *staffdomain.CreateStaffWithAccountInput) (*model.Staff, error)
	updatePasswordFn              func(ctx context.Context, accountID uint64, newPassword string) error
	setClinicAssignmentsFn        func(ctx context.Context, input *staffdomain.SetClinicAssignmentsInput) error
	updateFn                      func(ctx context.Context, clinicID, id uint64, input *staffdomain.UpdateStaffInput) (*model.Staff, error)
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	reorderFn                     func(ctx context.Context, clinicID uint64, ids []uint64) error
	getPermissionGroupIDsFn       func(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	setPermissionGroupIDsFn       func(ctx context.Context, staffID uint64, groupIDs []uint64) error
	getExcludedReservationTypesFn func(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	setExcludedReservationTypesFn func(ctx context.Context, staffID uint64, typeIDs []uint64) error
	getCapableReservationTypesFn  func(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	setCapableReservationTypesFn  func(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
	verifyClinicMembershipFn      func(ctx context.Context, staffID, clinicID uint64) error
}

func (m *mockService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID, page, limit)
	}
	return nil, 0, nil
}

func (m *mockService) GetByID(ctx context.Context, id uint64) (*model.Staff, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) GetByIDInClinic(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	if m.getByIDInClinicFn != nil {
		return m.getByIDInClinicFn(ctx, clinicID, id)
	}
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	if m.findByAccountIDFn != nil {
		return m.findByAccountIDFn(ctx, accountID)
	}
	return nil, nil
}

func (m *mockService) Create(ctx context.Context, input *staffdomain.CreateStaffInput) (*model.Staff, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return nil, nil
}

func (m *mockService) CreateWithAccount(ctx context.Context, input *staffdomain.CreateStaffWithAccountInput) (*model.Staff, error) {
	if m.createWithAccountFn != nil {
		return m.createWithAccountFn(ctx, input)
	}
	return nil, nil
}

func (m *mockService) UpdatePassword(ctx context.Context, accountID uint64, newPassword string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, accountID, newPassword)
	}
	return nil
}

func (m *mockService) SetClinicAssignments(ctx context.Context, input *staffdomain.SetClinicAssignmentsInput) error {
	if m.setClinicAssignmentsFn != nil {
		return m.setClinicAssignmentsFn(ctx, input)
	}
	return nil
}

func (m *mockService) Update(ctx context.Context, clinicID, id uint64, input *staffdomain.UpdateStaffInput) (*model.Staff, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return nil, nil
}

func (m *mockService) Delete(ctx context.Context, clinicID, id uint64, _ bool) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func (m *mockService) GetPermissionGroupIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error) {
	if m.getPermissionGroupIDsFn != nil {
		return m.getPermissionGroupIDsFn(ctx, clinicID, staffID)
	}
	return nil, nil
}

func (m *mockService) SetPermissionGroupIDs(ctx context.Context, _, staffID uint64, groupIDs []uint64) error {
	if m.setPermissionGroupIDsFn != nil {
		return m.setPermissionGroupIDsFn(ctx, staffID, groupIDs)
	}
	return nil
}

func (m *mockService) GetExcludedReservationTypeIDs(
	ctx context.Context,
	clinicID, staffID uint64,
) ([]uint64, error) {
	if m.getExcludedReservationTypesFn != nil {
		return m.getExcludedReservationTypesFn(ctx, clinicID, staffID)
	}
	return nil, nil
}

func (m *mockService) SetExcludedReservationTypeIDs(ctx context.Context, _, staffID uint64, typeIDs []uint64) error {
	if m.setExcludedReservationTypesFn != nil {
		return m.setExcludedReservationTypesFn(ctx, staffID, typeIDs)
	}
	return nil
}

func (m *mockService) GetCapableReservationTypeIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error) {
	if m.getCapableReservationTypesFn != nil {
		return m.getCapableReservationTypesFn(ctx, clinicID, staffID)
	}
	return nil, nil
}

func (m *mockService) SetCapableReservationTypeIDs(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if m.setCapableReservationTypesFn != nil {
		return m.setCapableReservationTypesFn(ctx, clinicID, staffID, typeIDs)
	}
	return nil
}

func (m *mockService) VerifyClinicMembership(ctx context.Context, staffID, clinicID uint64) error {
	if m.verifyClinicMembershipFn != nil {
		return m.verifyClinicMembershipFn(ctx, staffID, clinicID)
	}
	return nil
}

// mockStaffClinicAssignmentService は appointment_handler_test.go で定義済み

// ---- test helper ----

func newHandlerWithStaffSvc(staffSvc staffdomain.Service) *staffdomain.Handler {
	return staffdomain.NewHandler(
		staffSvc,
		&mockStaffClinicAssignmentService{},
		nil,
		nil,
		nil,
		nil,
	)
}

// ---- ListStaffs ----

func TestListStaffs_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with staff list",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				listFn: func(_ context.Context, clinicID uint64, _, _ int) ([]model.Staff, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.Staff{{ID: 10, Name: "田中太郎", StaffType: model.StaffTypeDoctor}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"田中太郎"`,
		},
		{
			name:     "returns 200 with empty list when no staff",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				listFn: func(_ context.Context, _ uint64, _, _ int) ([]model.Staff, int64, error) {
					return []model.Staff{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				listFn: func(_ context.Context, _ uint64, _, _ int) ([]model.Staff, int64, error) {
					return nil, 0, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithStaffSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListStaffs(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateStaff ----

func TestCreateStaff_Valid_Returns201(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockService
		wantStatus int
		wantBody   string
	}{
		{
			name: "creates staff without account returns 201",
			body: map[string]any{
				"name":       "山田花子",
				"staff_type": "nurse",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				createFn: func(_ context.Context, input *staffdomain.CreateStaffInput) (*model.Staff, error) {
					assert.Equal(t, "山田花子", input.Name)
					assert.Equal(t, uint64(1), input.ClinicID)
					return &model.Staff{ID: 20, Name: input.Name, StaffType: model.StaffTypeNurse}, nil
				},
				getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, Name: "山田花子", StaffType: model.StaffTypeNurse}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"山田花子"`,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"staff_type": "doctor"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     map[string]any{"name": "エラースタッフ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				createFn: func(_ context.Context, _ *staffdomain.CreateStaffInput) (*model.Staff, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithStaffSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateStaff(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteStaff ----

func newDeleteStaffRouter(staffSvc staffdomain.Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithStaffSvc(staffSvc)
	r.DELETE("/staffs/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteStaff)
	return r
}

func newDeleteStaffRouterWithActor(staffSvc staffdomain.Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithStaffSvc(staffSvc)
	r.DELETE("/staffs/:id", func(c *gin.Context) {
		setClinicID(c)
		setStaffID(c)
	}, h.DeleteStaff)
	return r
}

func TestDeleteStaff_NotFound_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockService
		wantStatus int
	}{
		{
			name:    "deletes staff successfully returns 204",
			paramID: "1",
			svc: &mockService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 404 when staff not found",
			paramID: "999",
			svc: &mockService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("staff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when staff is in use",
			paramID: "2",
			svc: &mockService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteStaffRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/staffs/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithStaffSvc(&mockService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteStaff(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 when deleting current staff", func(t *testing.T) {
		router := newDeleteStaffRouterWithActor(&mockService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/staffs/1", http.NoBody)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ---- UpdateStaff ----

func TestUpdateStaff(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockService
		wantStatus int
	}{
		{
			name:     "updates staff successfully returns 200",
			paramID:  "5",
			body:     map[string]any{"name": "更新スタッフ"},
			setupCtx: setStaffEditorContext,
			svc: &mockService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *staffdomain.UpdateStaffInput) (*model.Staff, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新スタッフ", *input.Name)
					assert.Equal(t, []uint64{1}, input.AuthorizedClinicIDs)
					assert.False(t, input.IsSystemAdmin)
					return &model.Staff{ID: 5, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 404 when staff not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: setStaffEditorContext,
			svc: &mockService{
				updateFn: func(_ context.Context, _, _ uint64, _ *staffdomain.UpdateStaffInput) (*model.Staff, error) {
					return nil, apperrors.WrapNotFound("staff", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 401 when system admin context is missing",
			paramID: "5",
			body:    map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_ids", []uint64{1})
			},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "returns 401 when non-admin clinic ids are missing",
			paramID: "5",
			body:    map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", false)
			},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "system administrator propagates trusted active clinic ids",
			paramID: "5",
			body:    map[string]any{"name": "更新スタッフ"},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockService{
				updateFn: func(_ context.Context, _, _ uint64, input *staffdomain.UpdateStaffInput) (*model.Staff, error) {
					assert.True(t, input.IsSystemAdmin)
					assert.Equal(t, []uint64{1, 2}, input.AuthorizedClinicIDs)
					return &model.Staff{ID: 5, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON body",
			paramID:    "5",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			body:     map[string]any{"name": "エラースタッフ"},
			setupCtx: setStaffEditorContext,
			svc: &mockService{
				updateFn: func(_ context.Context, _, _ uint64, _ *staffdomain.UpdateStaffInput) (*model.Staff, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithStaffSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateStaff(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- ReorderStaffs ----

func TestReorderStaffs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockService
		wantStatus int
	}{
		{
			name:     "reorders successfully and returns 204",
			body:     map[string]any{"ids": []uint64{3, 1, 2}},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, []uint64{3, 1, 2}, ids)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{"ids": []uint64{1}},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when ids is empty",
			body:       map[string]any{"ids": []uint64{}},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON body",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     map[string]any{"ids": []uint64{1, 2}},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockService{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithStaffSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.ReorderStaffs(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
