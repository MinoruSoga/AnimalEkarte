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

// ---- mock PrescriptionService ----

type mockPrescriptionService struct {
	listFn    func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error)
	getByIDFn func(ctx context.Context, clinicID, prescriptionID uint64) (*model.Prescription, error)
	createFn  func(ctx context.Context, clinicID, medicalRecordID uint64, input *CreatePrescriptionInput) (*model.Prescription, error)
	updateFn  func(ctx context.Context, clinicID, medicalRecordID, prescriptionID uint64, input *UpdatePrescriptionInput) (*model.Prescription, error)
	deleteFn  func(ctx context.Context, clinicID, medicalRecordID, prescriptionID uint64) error
}

func (m *mockPrescriptionService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
	return m.listFn(ctx, clinicID, medicalRecordID)
}

func (m *mockPrescriptionService) GetByID(ctx context.Context, clinicID, prescriptionID uint64) (*model.Prescription, error) {
	return m.getByIDFn(ctx, clinicID, prescriptionID)
}

func (m *mockPrescriptionService) Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreatePrescriptionInput) (*model.Prescription, error) {
	return m.createFn(ctx, clinicID, medicalRecordID, input)
}

func (m *mockPrescriptionService) Update(ctx context.Context, clinicID, medicalRecordID, prescriptionID uint64, input *UpdatePrescriptionInput) (*model.Prescription, error) {
	return m.updateFn(ctx, clinicID, medicalRecordID, prescriptionID, input)
}

func (m *mockPrescriptionService) Delete(ctx context.Context, clinicID, medicalRecordID, prescriptionID uint64) error {
	return m.deleteFn(ctx, clinicID, medicalRecordID, prescriptionID)
}

func newHandlerWithPrescriptionSvc(svc PrescriptionService) *PrescriptionHandler {
	return NewPrescriptionHandler(svc)
}

// ---- ListPrescriptions ----

func TestListPrescriptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockPrescriptionService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of prescriptions",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPrescriptionService{
				listFn: func(_ context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					return []model.Prescription{{ID: 1, ClinicID: clinicID, OwnerID: 5, DurationDays: 7}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"duration_days":7`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPrescriptionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPrescriptionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPrescriptionService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPrescriptionSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ListPrescriptions(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreatePrescription ----

func TestCreatePrescription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockPrescriptionService
		wantStatus int
		wantBody   string
		wantHeader string
	}{
		{
			name:     "creates prescription successfully",
			paramID:  "1",
			body:     map[string]any{"date": "2026-05-28", "duration_days": 7},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPrescriptionService{
				createFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *CreatePrescriptionInput) (*model.Prescription, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, 7, input.DurationDays)
					return &model.Prescription{ID: 9, ClinicID: clinicID, OwnerID: 3, DurationDays: input.DurationDays, PrescribedAt: input.PrescribedAt}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"duration_days":7`,
			wantHeader: "/api/v1/medical-records/1/prescriptions/9",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"date": "2026-05-28", "duration_days": 7},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPrescriptionService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"date": "2026-05-28", "duration_days": 7},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPrescriptionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when duration_days is missing",
			paramID:    "1",
			body:       map[string]any{"date": "2026-05-28"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPrescriptionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date format",
			paramID:    "1",
			body:       map[string]any{"date": "2026/05/28", "duration_days": 7},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPrescriptionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPrescriptionService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			body:     map[string]any{"date": "2026-05-28", "duration_days": 7},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPrescriptionService{
				createFn: func(_ context.Context, _, _ uint64, _ *CreatePrescriptionInput) (*model.Prescription, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPrescriptionSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.CreatePrescription(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdatePrescription ----

func TestUpdatePrescription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		paramID           string
		paramPrescription string
		body              any
		setupCtx          func(c *gin.Context)
		svc               *mockPrescriptionService
		wantStatus        int
		wantBody          string
	}{
		{
			name:              "updates prescription successfully",
			paramID:           "1",
			paramPrescription: "9",
			body:              map[string]any{"duration_days": 14},
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc: &mockPrescriptionService{
				updateFn: func(_ context.Context, clinicID, medicalRecordID, prescriptionID uint64, input *UpdatePrescriptionInput) (*model.Prescription, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(9), prescriptionID)
					require.NotNil(t, input.DurationDays)
					return &model.Prescription{ID: prescriptionID, ClinicID: clinicID, OwnerID: 3, DurationDays: *input.DurationDays}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"duration_days":14`,
		},
		{
			name:              "returns 401 when clinic_id is missing",
			paramID:           "1",
			paramPrescription: "9",
			body:              map[string]any{"duration_days": 14},
			setupCtx:          func(_ *gin.Context) {},
			svc:               &mockPrescriptionService{},
			wantStatus:        http.StatusUnauthorized,
		},
		{
			name:              "returns 400 for non-numeric medical record id",
			paramID:           "xyz",
			paramPrescription: "9",
			body:              map[string]any{"duration_days": 14},
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc:               &mockPrescriptionService{},
			wantStatus:        http.StatusBadRequest,
		},
		{
			name:              "returns 400 for non-numeric prescription id",
			paramID:           "1",
			paramPrescription: "xyz",
			body:              map[string]any{"duration_days": 14},
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc:               &mockPrescriptionService{},
			wantStatus:        http.StatusBadRequest,
		},
		{
			name:              "returns 400 for invalid JSON",
			paramID:           "1",
			paramPrescription: "9",
			body:              "not-json",
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc:               &mockPrescriptionService{},
			wantStatus:        http.StatusBadRequest,
		},
		{
			name:              "returns 400 for invalid date format",
			paramID:           "1",
			paramPrescription: "9",
			body:              map[string]any{"date": "not-a-date"},
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc:               &mockPrescriptionService{},
			wantStatus:        http.StatusBadRequest,
		},
		{
			name:              "returns 404 when prescription not found",
			paramID:           "1",
			paramPrescription: "999",
			body:              map[string]any{"duration_days": 14},
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc: &mockPrescriptionService{
				updateFn: func(_ context.Context, _, _, _ uint64, _ *UpdatePrescriptionInput) (*model.Prescription, error) {
					return nil, apperrors.WrapNotFound("prescription", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPrescriptionSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{
				{Key: "id", Value: tt.paramID},
				{Key: "prescriptionId", Value: tt.paramPrescription},
			}
			tt.setupCtx(c)

			h.UpdatePrescription(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeletePrescription ----

func TestDeletePrescription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		paramID           string
		paramPrescription string
		setupCtx          func(c *gin.Context)
		svc               *mockPrescriptionService
		wantStatus        int
	}{
		{
			name:              "deletes prescription successfully",
			paramID:           "1",
			paramPrescription: "9",
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc: &mockPrescriptionService{
				deleteFn: func(_ context.Context, clinicID, medicalRecordID, prescriptionID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(9), prescriptionID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:              "returns 401 when clinic_id is missing",
			paramID:           "1",
			paramPrescription: "9",
			setupCtx:          func(_ *gin.Context) {},
			svc:               &mockPrescriptionService{},
			wantStatus:        http.StatusUnauthorized,
		},
		{
			name:              "returns 400 for non-numeric medical record id",
			paramID:           "xyz",
			paramPrescription: "9",
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc:               &mockPrescriptionService{},
			wantStatus:        http.StatusBadRequest,
		},
		{
			name:              "returns 400 for non-numeric prescription id",
			paramID:           "1",
			paramPrescription: "xyz",
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc:               &mockPrescriptionService{},
			wantStatus:        http.StatusBadRequest,
		},
		{
			name:              "returns 500 on service error",
			paramID:           "1",
			paramPrescription: "9",
			setupCtx:          func(c *gin.Context) { setClinicID(c) },
			svc: &mockPrescriptionService{
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPrescriptionSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{
				{Key: "id", Value: tt.paramID},
				{Key: "prescriptionId", Value: tt.paramPrescription},
			}
			tt.setupCtx(c)

			h.DeletePrescription(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestListPrescriptionsSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*PrescriptionHandler, *gin.Context)
		svc    *mockPrescriptionService
	}{
		{
			name: "ListPrescriptions returns 403 when selected clinic lacks medical record view grant",
			invoke: func(h *PrescriptionHandler, c *gin.Context) {
				h.ListPrescriptions(c)
			},
			svc: &mockPrescriptionService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPrescriptionSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMedicalRecords), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
