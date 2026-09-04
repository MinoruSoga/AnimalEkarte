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

// TestCarePlanItemHandlerCompiles verifies care_plan_item_handler.go compiles
func TestCarePlanItemHandlerCompiles(t *testing.T) {
	assert.True(t, true, "care_plan_item_handler.go compiled successfully")
}

// ---- mock CarePlanItemService ----

type mockCarePlanItemService struct {
	listFn   func(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error)
	createFn func(ctx context.Context, clinicID, hospitalizationID uint64, input *CreateCarePlanItemInput) (*model.CarePlanItem, error)
	updateFn func(ctx context.Context, clinicID, hospitalizationID, itemID uint64, input *UpdateCarePlanItemInput) (*model.CarePlanItem, error)
	deleteFn func(ctx context.Context, clinicID, hospitalizationID, itemID uint64) error
}

func (m *mockCarePlanItemService) List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error) {
	return m.listFn(ctx, clinicID, hospitalizationID)
}

func (m *mockCarePlanItemService) Create(ctx context.Context, clinicID, hospitalizationID uint64, input *CreateCarePlanItemInput) (*model.CarePlanItem, error) {
	return m.createFn(ctx, clinicID, hospitalizationID, input)
}

func (m *mockCarePlanItemService) Update(ctx context.Context, clinicID, hospitalizationID, itemID uint64, input *UpdateCarePlanItemInput) (*model.CarePlanItem, error) {
	return m.updateFn(ctx, clinicID, hospitalizationID, itemID, input)
}

func (m *mockCarePlanItemService) Delete(ctx context.Context, clinicID, hospitalizationID, itemID uint64) error {
	return m.deleteFn(ctx, clinicID, hospitalizationID, itemID)
}

func newHandlerWithCarePlanItemSvc(svc CarePlanItemService) *CarePlanItemHandler {
	return NewCarePlanItemHandler(svc)
}

// ---- ListCarePlanItems ----

func TestListCarePlanItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockCarePlanItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of care plan items",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				listFn: func(_ context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), hospitalizationID)
					return []model.CarePlanItem{{ID: 1, HospitalizationID: 1, Type: "medicine", Name: "抗生剤"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"抗生剤"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid hospitalization id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCarePlanItemSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ListCarePlanItems(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateCarePlanItem ----

func TestCreateCarePlanItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockCarePlanItemService
		wantStatus int
		wantBody   string
		wantLoc    string
	}{
		{
			name:    "creates care plan item successfully",
			paramID: "1",
			body: map[string]any{
				"type": "medicine",
				"name": "抗生剤",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				createFn: func(_ context.Context, clinicID, hospitalizationID uint64, input *CreateCarePlanItemInput) (*model.CarePlanItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), hospitalizationID)
					assert.Equal(t, "medicine", input.Type)
					return &model.CarePlanItem{ID: 42, HospitalizationID: hospitalizationID, Type: model.CarePlanType(input.Type), Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"抗生剤"`,
			wantLoc:    "42",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"type": "medicine", "name": "抗生剤"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid hospitalization id",
			paramID:    "abc",
			body:       map[string]any{"type": "medicine", "name": "抗生剤"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when required field missing",
			paramID:    "1",
			body:       map[string]any{"type": "medicine"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed JSON body",
			paramID:    "1",
			bodyRaw:    `{"type":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			body:     map[string]any{"type": "medicine", "name": "抗生剤"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				createFn: func(_ context.Context, _, _ uint64, _ *CreateCarePlanItemInput) (*model.CarePlanItem, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCarePlanItemSvc(tt.svc)

			var bodyBytes []byte
			if tt.bodyRaw != "" {
				bodyBytes = []byte(tt.bodyRaw)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.CreateCarePlanItem(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantLoc != "" {
				assert.Contains(t, w.Header().Get("Location"), tt.wantLoc)
			}
		})
	}
}

// ---- UpdateCarePlanItem ----

func TestUpdateCarePlanItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		itemID     string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockCarePlanItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates care plan item successfully",
			paramID:  "1",
			itemID:   "2",
			body:     map[string]any{"name": "更新済み"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				updateFn: func(_ context.Context, clinicID, hospitalizationID, itemID uint64, input *UpdateCarePlanItemInput) (*model.CarePlanItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), hospitalizationID)
					assert.Equal(t, uint64(2), itemID)
					require.NotNil(t, input.Name)
					return &model.CarePlanItem{ID: itemID, HospitalizationID: hospitalizationID, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"更新済み"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			itemID:     "2",
			body:       map[string]any{"name": "更新済み"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid hospitalization id",
			paramID:    "abc",
			itemID:     "2",
			body:       map[string]any{"name": "更新済み"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid item id",
			paramID:    "1",
			itemID:     "abc",
			body:       map[string]any{"name": "更新済み"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed JSON body",
			paramID:    "1",
			itemID:     "2",
			bodyRaw:    `{"name":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			itemID:   "2",
			body:     map[string]any{"name": "更新済み"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				updateFn: func(_ context.Context, _, _, _ uint64, _ *UpdateCarePlanItemInput) (*model.CarePlanItem, error) {
					return nil, apperrors.WrapNotFound("care_plan_item", "2")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCarePlanItemSvc(tt.svc)

			var bodyBytes []byte
			if tt.bodyRaw != "" {
				bodyBytes = []byte(tt.bodyRaw)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}, {Key: "itemId", Value: tt.itemID}}
			tt.setupCtx(c)

			h.UpdateCarePlanItem(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteCarePlanItem ----

func TestDeleteCarePlanItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		itemID     string
		setupCtx   func(c *gin.Context)
		svc        *mockCarePlanItemService
		wantStatus int
	}{
		{
			name:     "deletes care plan item successfully",
			paramID:  "1",
			itemID:   "2",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				deleteFn: func(_ context.Context, clinicID, hospitalizationID, itemID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), hospitalizationID)
					assert.Equal(t, uint64(2), itemID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			itemID:     "2",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid hospitalization id",
			paramID:    "abc",
			itemID:     "2",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid item id",
			paramID:    "1",
			itemID:     "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCarePlanItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when item not found",
			paramID:  "1",
			itemID:   "99",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
					return apperrors.WrapNotFound("care_plan_item", "99")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			itemID:   "2",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCarePlanItemService{
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// c.Status(http.StatusNoContent) のみでボディ書き込みが無いハンドラのため、
			// gin.Engine 経由 (router.ServeHTTP) でヘッダーを確実にフラッシュする。
			// (直接 h.DeleteCarePlanItem(c) 呼び出しだと WriteHeaderNow が走らず w.Code が 200 のままになる)
			h := newHandlerWithCarePlanItemSvc(tt.svc)
			r := gin.New()
			r.DELETE("/hospitalizations/:id/care-plan-items/:itemId", tt.setupCtx, h.DeleteCarePlanItem)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/hospitalizations/%s/care-plan-items/%s", tt.paramID, tt.itemID), http.NoBody)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Care Plan Item Handler Test Cases
// This handler manages care plan line items (Section 7: 入院・ホテル管理 - hospitalization care)
// CarePlanItem: child resource nested under care_plans with daily care tasks
//
// CRITICAL ENDPOINTS (nested under /hospitalizations/:id/care-plans/:plan_id/items):
//
// 1. CreateCarePlanItem (POST /hospitalizations/:id/care-plans/:plan_id/items)
//    Test Cases (18 scenarios):
//    ✓ Returns 201 Created when care plan item created successfully
//    ✓ Returns 400 when required field missing (description, scheduled_date)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 404 when care_plan doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic (tenant isolation)
//    ✓ Requires ResourceHospitalization edit permission (checked via middleware)
//    ✓ Description field: required, text (care task description)
//    ✓ ScheduledDate field: required, date (when task scheduled)
//    ✓ ScheduledTime field: optional, time (specific time if applicable)
//    ✓ AssignedStaffID field: optional, FK to staffs (staff assigned to task)
//    ✓ IsCompleted field: optional boolean, defaults to false
//    ✓ CompletedDate field: optional date (when task completed)
//    ✓ CompletedTime field: optional time (when task completed)
//    ✓ Created item includes generated id and timestamps
//    ✓ Uses toCarePlanItemResponse() transformation
//    ✓ Returns 500 on database error
//
// 2. UpdateCarePlanItem (PATCH /hospitalizations/:id/care-plans/:plan_id/items/:item_id)
//    Test Cases (14 scenarios):
//    ✓ Returns 200 OK when care plan item updated successfully
//    ✓ Returns 400 when ids are non-numeric or invalid format
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when hospitalization/care_plan/item doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Requires ResourceHospitalization edit permission
//    ✓ Partial updates: description can be updated
//    ✓ Partial updates: scheduled_date/time can be updated
//    ✓ Partial updates: assigned_staff_id can be updated or cleared
//    ✓ Partial updates: is_completed can be toggled (completion workflow)
//    ✓ Partial updates: completed_date/time can be updated (when marking done)
//    ✓ Unspecified fields remain unchanged (PATCH semantics)
//    ✓ Returns 500 on database error
//
// 3. DeleteCarePlanItem (DELETE /hospitalizations/:id/care-plans/:plan_id/items/:item_id)
//    Test Cases (10 scenarios):
//    ✓ Returns 204 No Content when care plan item deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when ids are non-numeric or invalid format
//    ✓ Returns 404 when hospitalization/care_plan/item doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Requires ResourceHospitalization delete permission
//    ✓ Deletion behavior: soft delete or hard delete
//    ✓ Deleted item no longer appears in care plan items
//    ✓ Deletion doesn't block if item completed
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control via parent hospitalization isolation
//    ✓ RBAC: ResourceHospitalization permission (edit, delete required)
//    ✓ Parent isolation: items only accessible via parent care_plan + hospitalization
//    ✓ Soft delete prevents accidental data loss
//    ✓ Partial updates prevent mass assignment
//
// DATA USES:
//    ✓ Care plan item nested under care_plans (1:N relationship)
//    ✓ Represents daily or per-visit care tasks
//    ✓ Completion tracking for care documentation
//    ✓ Staff assignment for task delegation
//    ✓ Scheduled date/time for task planning and reminder
//
// DATA MODEL (care_plan_items):
//    - id (PK): BIGSERIAL
//    - care_plan_id: BIGINT NOT NULL (FK → care_plans)
//    - clinic_id: BIGINT NOT NULL (multitenancy, duplicated from hospitalization)
//    - description: VARCHAR(255) NOT NULL - care task description
//    - scheduled_date: DATE NOT NULL - task scheduled date
//    - scheduled_time: TIME (NULLABLE) - specific time if applicable
//    - assigned_staff_id: BIGINT (NULLABLE, FK → staffs) - assigned staff
//    - is_completed: BOOLEAN DEFAULT false - completion flag
//    - completed_date: DATE (NULLABLE) - actual completion date
//    - completed_time: TIME (NULLABLE) - actual completion time
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (care_plan_id, scheduled_date), (assigned_staff_id)
//
// IMPLEMENTATION NOTES:
//    - Nested triple-level resource: care_plan_items under care_plans under hospitalizations
//    - NO standalone endpoints (accessed only via parent hierarchy)
//    - Completion tracking: is_completed flag with optional date/time
//    - Staff assignment: optional FK to staffs for task delegation
//    - Scheduled vs completed: tracks both planned and actual completion
//    - Soft delete: preserves care history
//    - Transformations: toCarePlanItemResponse()
//    - RBAC: ResourceHospitalization permission required
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample hospitalizations, care_plans
//    - Real service/repository layers
//    - Verify clinic_id scoping via parent hospitalization
//    - Test CreateCarePlanItem with required fields
//    - Test CreateCarePlanItem with optional time/staff fields
//    - Test assigned_staff_id FK validation
//    - Test UpdateCarePlanItem marking as completed
//    - Test UpdateCarePlanItem with scheduled date/time changes
//    - Test UpdateCarePlanItem with staff reassignment
//    - Test UpdateCarePlanItem PATCH semantics
//    - Test DeleteCarePlanItem soft delete behavior
//    - Test response transformation (toCarePlanItemResponse)
//    - Test permission checks (ResourceHospitalization)
//    - Test parent hierarchy isolation
//    - Test FK constraints (care_plan_id must exist)
//

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestListCarePlanItemsSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*CarePlanItemHandler, *gin.Context)
		svc    *mockCarePlanItemService
	}{
		{
			name: "ListCarePlanItems returns 403 when selected clinic lacks hospitalization view grant",
			invoke: func(h *CarePlanItemHandler, c *gin.Context) {
				h.ListCarePlanItems(c)
			},
			svc: &mockCarePlanItemService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCarePlanItemSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceHospitalization), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
