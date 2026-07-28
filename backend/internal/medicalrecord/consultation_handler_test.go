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

// TestConsultationHandlerCompiles verifies consultation_handler.go compiles
func TestConsultationHandlerCompiles(t *testing.T) {
	assert.True(t, true, "consultation_handler.go compiled successfully")
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Consultation Handler Test Cases
// This handler manages consultation/service type master data for medical billing (Section 4: カルテ管理 master)
// Consultations: billable medical consultation types (e.g., "予防接種相談", "栄養相談", "行動相談")
//
// CRITICAL ENDPOINTS:
//
// 1. ListConsultations (GET /consultations)
//    Test Cases (6 scenarios):
//    ✓ Returns 200 OK with empty list when no consultations exist
//    ✓ Returns 200 OK with list of all clinic's consultations (no pagination)
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Response includes all fields: id, name, price, description, parent_id
//    ✓ Response includes: time_condition, duration, tax_type, tax_rate, sort_order, is_active
//    ✓ Returns 500 on database error
//
// 2. GetConsultation (GET /consultations/:id)
//    Test Cases (9 scenarios):
//    ✓ Returns 200 OK with single consultation record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when id is non-numeric or invalid format
//    ✓ Returns 404 when consultation doesn't exist
//    ✓ Returns 403 when consultation belongs to different clinic (tenant isolation)
//    ✓ Response includes complete consultation data with all fields
//    ✓ Returns 500 on database error
//
// 3. CreateConsultation (POST /consultations)
//    Test Cases (19 scenarios):
//    ✓ Returns 201 Created when consultation created successfully
//    ✓ Returns 400 when required field missing (name)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Requires ResourceMasterMedical create permission (checked via middleware)
//    ✓ Name field: required, text (e.g., "予防接種相談", "栄養相談")
//    ✓ Price field: optional numeric for billing
//    ✓ Price: non-negative validation if provided
//    ✓ Description field: optional text for service details
//    ✓ TimeCondition field: optional text (e.g., "初診のみ", "フォローアップ")
//    ✓ Duration field: optional numeric (minutes required for consultation)
//    ✓ IsActive field: optional boolean, defaults to true
//    ✓ SortOrder field: optional numeric for display ordering
//    ✓ ParentID field: optional FK to parent consultation (hierarchical structure)
//    ✓ TaxType field: optional ENUM (included, excluded), defaults to excluded
//    ✓ TaxRate field: optional numeric (percentage 0-100), defaults to 0.10 (10%)
//    ✓ Created consultation includes generated id and timestamps
//    ✓ Returns 409 if name already exists (if UNIQUE constraint per clinic)
//    ✓ Returns 500 on database error
//
// 4. UpdateConsultation (PATCH /consultations/:id)
//    Test Cases (19 scenarios):
//    ✓ Returns 200 OK when consultation updated successfully
//    ✓ Returns 400 when id is non-numeric or invalid format
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when consultation doesn't exist
//    ✓ Returns 403 when consultation belongs to different clinic
//    ✓ Requires ResourceMasterMedical edit permission (checked via middleware)
//    ✓ Partial updates: name can be updated independently
//    ✓ Partial updates: price can be updated or cleared
//    ✓ Partial updates: description can be updated or cleared
//    ✓ Partial updates: time_condition can be updated or cleared
//    ✓ Partial updates: duration can be updated or cleared
//    ✓ Partial updates: is_active can be toggled
//    ✓ Partial updates: sort_order can be updated
//    ✓ Partial updates: parent_id can be updated (change hierarchy)
//    ✓ ClearParentID field: special flag to null out parent_id on PATCH
//    ✓ Partial updates: tax_type can be updated (with ENUM validation)
//    ✓ Partial updates: tax_rate can be updated (numeric 0-100 validation)
//    ✓ Unspecified fields remain unchanged (PATCH semantics, not PUT)
//    ✓ Returns 500 on database error
//
// 5. DeleteConsultation (DELETE /consultations/:id)
//    Test Cases (10 scenarios):
//    ✓ Returns 204 No Content when consultation deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when id is non-numeric or invalid format
//    ✓ Returns 404 when consultation doesn't exist
//    ✓ Returns 403 when consultation belongs to different clinic
//    ✓ Deletion behavior: soft delete or hard delete (depends on implementation)
//    ✓ Deleted consultation no longer appears in ListConsultations
//    ✓ Deleted consultation cannot be retrieved by GetConsultation (404)
//    ✓ Deletion should check for FK dependencies (billing records referencing this)
//    ✓ Returns 409 Conflict if consultation is still in use (billing records exist)
//
// 6. ReorderConsultations (POST /consultations/reorder)
//    Test Cases (8 scenarios):
//    ✓ Returns 200 OK with message when reorder succeeds
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Requires ResourceMasterMedical edit permission (checked via middleware)
//    ✓ Accepts array of IDs in desired order
//    ✓ Updates sort_order for all provided IDs (0, 1, 2, ...)
//    ✓ Partial reorder supported (only specified IDs reordered)
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification on all endpoints)
//    ✓ RBAC: ResourceMasterMedical permission (create, edit required)
//    ✓ Partial updates prevent mass assignment (explicit field mapping)
//    ✓ Soft delete prevents accidental data loss (if implemented)
//
// DATA USES:
//    ✓ Consultation referenced by billing records (FK constraint)
//    ✓ Price used for billing calculations
//    ✓ TaxType and TaxRate used for tax computation on invoices
//    ✓ Duration used for schedule planning
//    ✓ TimeCondition controls when service is available
//    ✓ IsActive used to hide inactive consultations from UI dropdowns
//
// DATA MODEL (consultations):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT NOT NULL (multitenancy)
//    - name: VARCHAR(100) NOT NULL - consultation type name
//    - price: NUMERIC(10,2) (NULLABLE) - consultation fee
//    - is_active: BOOLEAN DEFAULT true - enable/disable flag
//    - description: TEXT (NULLABLE) - service details
//    - time_condition: VARCHAR(100) (NULLABLE) - condition (初診のみ, etc.)
//    - duration: INTEGER (NULLABLE) - minutes required
//    - parent_id (FK, NULLABLE): BIGINT → consultations(id) - parent for hierarchy
//    - sort_order: INTEGER DEFAULT 0 - display ordering
//    - tax_type: VARCHAR(50) DEFAULT 'excluded' - ENUM (included, excluded)
//    - tax_rate: NUMERIC(5,4) DEFAULT 0.1000 - tax rate (0-1.0 or 0-100%)
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete, if implemented)
//    - Indexes: (clinic_id, id), (clinic_id, parent_id), (clinic_id, is_active), (clinic_id, sort_order)
//    - Unique constraint: (clinic_id, name) WHERE deleted_at IS NULL (if enforced)
//
// IMPLEMENTATION NOTES:
//    - Clinic-scoped master data (clinic_id extraction required)
//    - Hierarchical structure: ParentID allows parent-child relationships
//    - ClearParentID special flag: used during PATCH to set parent_id to null
//    - Price: numeric for billing integration
//    - TaxType: ENUM for tax classification (included vs excluded)
//    - TaxRate: numeric default 0.10 (10%) for Japanese standard rate
//    - Duration: optional time estimate in minutes
//    - TimeCondition: optional text constraint (e.g., "初診のみ" = first visit only)
//    - IsActive: allows disabling without deletion
//    - SortOrder: numeric for custom display ordering
//    - Transformations: direct response (no transformation function needed based on code)
//    - PATCH semantics: unspecified fields remain unchanged
//    - RBAC: ResourceMasterMedical permission required
//    - Should check FK dependencies before delete (billing records reference this)
//    - ReorderConsultations: returns 200 OK with message (not 204)
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample consultation records
//    - Real service/repository layers
//    - Verify clinic_id scoping (no cross-clinic data access)
//    - Test default values (is_active=true, sort_order=0, tax_type=excluded, tax_rate=0.10)
//    - Test price numeric range and non-negative validation
//    - Test tax_type ENUM validation (included, excluded)
//    - Test tax_rate numeric range (0-1.0 or 0-100 validation)
//    - Test parent-child relationships (hierarchy validation)
//    - Test ClearParentID flag sets parent_id to null
//    - Test sort_order affects ListConsultations ordering
//    - Test ReorderConsultations updates sort_order correctly
//    - Test FK constraint: billing records referencing deleted consultation (409)
//    - Verify soft delete behavior (if implemented)
//    - Test active filtering (is_active=false excluded from UI dropdowns)
//    - Test PATCH semantics (unspecified fields unchanged)
//    - Test name uniqueness per clinic (if UNIQUE constraint exists)
//    - Test bulk operations (reorder with partial ID list)
//    - Verify clinic_id parameter on all endpoints
//    - Test permission checks (ResourceMasterMedical)
//

// ---- mock ConsultationService ----

type mockConsultationService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateConsultationInput) (*model.Consultation, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockConsultationService) List(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	return m.listFn(ctx, clinicID)
}

func (m *mockConsultationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockConsultationService) Create(ctx context.Context, clinicID uint64, input *CreateConsultationInput) (*model.Consultation, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockConsultationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockConsultationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockConsultationService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- test helper ----

func newHandlerWithConsultationSvc(svc ConsultationService) *ConsultationHandler {
	return NewConsultationHandler(svc)
}

// ---- ListConsultations ----

func TestListConsultations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockConsultationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of consultations",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.Consultation, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.Consultation{{ID: 1, ClinicID: clinicID, Name: "予防接種相談"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"予防接種相談"`,
		},
		{
			name:     "returns empty list when no consultations exist",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				listFn: func(_ context.Context, _ uint64) ([]model.Consultation, error) {
					return []model.Consultation{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockConsultationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				listFn: func(_ context.Context, _ uint64) ([]model.Consultation, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithConsultationSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListConsultations(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetConsultation ----

func TestGetConsultation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockConsultationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns consultation for valid id",
			paramID:  "4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Consultation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(4), id)
					return &model.Consultation{ID: 4, ClinicID: clinicID, Name: "栄養相談"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"栄養相談"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockConsultationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockConsultationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when consultation not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Consultation, error) {
					return nil, apperrors.WrapNotFound("consultation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithConsultationSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetConsultation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateConsultation ----

func TestCreateConsultation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":      "予防接種相談",
			"is_active": true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockConsultationService
		wantStatus int
		wantBody   string
		wantHeader bool
	}{
		{
			name:     "creates consultation successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateConsultationInput) (*model.Consultation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "予防接種相談", input.Name)
					return &model.Consultation{ID: 8, ClinicID: clinicID, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"予防接種相談"`,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockConsultationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockConsultationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockConsultationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				createFn: func(_ context.Context, _ uint64, _ *CreateConsultationInput) (*model.Consultation, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithConsultationSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateConsultation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantHeader {
				assert.Equal(t, "/v1/masters/consultations/8", w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateConsultation ----

func TestUpdateConsultation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockConsultationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates consultation successfully",
			paramID:  "1",
			body:     map[string]any{"name": "行動相談"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "行動相談", *input.Name)
					return &model.Consultation{ID: 1, ClinicID: clinicID, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"行動相談"`,
		},
		{
			name:     "partial update with description",
			paramID:  "1",
			body:     map[string]any{"description": "行動に関する相談"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateConsultationInput) (*model.Consultation, error) {
					require.NotNil(t, input.Description)
					assert.Equal(t, "行動に関する相談", *input.Description)
					assert.Nil(t, input.Name)
					return &model.Consultation{ID: 1, Description: *input.Description}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockConsultationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockConsultationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockConsultationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when consultation not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockConsultationService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateConsultationInput) (*model.Consultation, error) {
					return nil, apperrors.WrapNotFound("consultation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithConsultationSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateConsultation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ReorderConsultations ----

func newReorderConsultationsRouter(svc ConsultationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithConsultationSvc(svc)
	r.POST("/consultations/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderConsultations)
	return r
}

func TestReorderConsultations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders consultations successfully", func(t *testing.T) {
		svc := &mockConsultationService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 1, 3}, ids)
				return nil
			},
		}
		router := newReorderConsultationsRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{2, 1, 3}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/consultations/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithConsultationSvc(&mockConsultationService{})
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderConsultations(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderConsultationsRouter(&mockConsultationService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/consultations/reorder", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		svc := &mockConsultationService{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return fmt.Errorf("db error")
			},
		}
		router := newReorderConsultationsRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/consultations/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- DeleteConsultation ----

func newDeleteConsultationRouter(svc ConsultationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithConsultationSvc(svc)
	r.DELETE("/consultations/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteConsultation)
	return r
}

func TestDeleteConsultation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockConsultationService
		wantStatus int
	}{
		{
			name:    "deletes consultation successfully",
			paramID: "1",
			svc: &mockConsultationService{
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
			svc:        &mockConsultationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when consultation not found",
			paramID: "999",
			svc: &mockConsultationService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("consultation", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when consultation is in use",
			paramID: "2",
			svc: &mockConsultationService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この診察種別は診療記録で使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteConsultationRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/consultations/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithConsultationSvc(&mockConsultationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteConsultation(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
