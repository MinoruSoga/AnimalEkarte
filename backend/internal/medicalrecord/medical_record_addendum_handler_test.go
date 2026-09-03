package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock MedicalRecordAddendumService ----

type mockMedicalRecordAddendumService struct {
	createFn                func(ctx context.Context, clinicID uint64, input CreateMedicalRecordAddendumInput) (*model.MedicalRecordAddendum, error)
	findByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error)
}

func (m *mockMedicalRecordAddendumService) Create(ctx context.Context, clinicID uint64, input CreateMedicalRecordAddendumInput) (*model.MedicalRecordAddendum, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return nil, nil
}

func (m *mockMedicalRecordAddendumService) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error) {
	if m.findByMedicalRecordIDFn != nil {
		return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
	}
	return nil, nil
}

func newHandlerWithAddendumSvc(svc MedicalRecordAddendumService) *MedicalRecordAddendumHandler {
	return NewMedicalRecordAddendumHandler(svc)
}

// ---- ListMedicalRecordAddenda ----

func TestListMedicalRecordAddenda(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockMedicalRecordAddendumService
		wantStatus int
	}{
		{
			name:     "returns addenda successfully",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordAddendumService{
				findByMedicalRecordIDFn: func(_ context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					return []*model.MedicalRecordAddendum{
						{ID: 1, MedicalRecordID: 1, AfterText: "text", Reason: "reason", CreatedAt: time.Now()},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicalRecordAddendumService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicalRecordAddendumService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when medical record not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordAddendumService{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]*model.MedicalRecordAddendum, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAddendumSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListMedicalRecordAddenda(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- CreateMedicalRecordAddendum ----

func TestCreateMedicalRecordAddendum(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := map[string]any{
		"after_text": "corrected soap note",
		"reason":     "typo correction",
	}

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		body       any
		svc        *mockMedicalRecordAddendumService
		wantStatus int
	}{
		{
			name:    "creates addendum with 201",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("user_id", "10")
			},
			body: validBody,
			svc: &mockMedicalRecordAddendumService{
				createFn: func(_ context.Context, clinicID uint64, input CreateMedicalRecordAddendumInput) (*model.MedicalRecordAddendum, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), input.MedicalRecordID)
					assert.Equal(t, uint64(10), input.AuthorUserID)
					return &model.MedicalRecordAddendum{ID: 42, MedicalRecordID: 1, AfterText: input.AfterText, Reason: input.Reason}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			body:       validBody,
			svc:        &mockMedicalRecordAddendumService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for missing after_text",
			paramID:    "1",
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			body:       map[string]any{"reason": "some reason"},
			svc:        &mockMedicalRecordAddendumService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 when medical record is not finalized",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			body:     validBody,
			svc: &mockMedicalRecordAddendumService{
				createFn: func(_ context.Context, _ uint64, _ CreateMedicalRecordAddendumInput) (*model.MedicalRecordAddendum, error) {
					return nil, apperrors.WrapInvalidInput("addendum can only be added to finalized medical records")
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAddendumSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			bodyBytes, _ := json.Marshal(tt.body)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.CreateMedicalRecordAddendum(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestListMedicalRecordAddendaSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*MedicalRecordAddendumHandler, *gin.Context)
		svc    *mockMedicalRecordAddendumService
	}{
		{
			name: "ListMedicalRecordAddenda returns 403 when selected clinic lacks medical record view grant",
			invoke: func(h *MedicalRecordAddendumHandler, c *gin.Context) {
				h.ListMedicalRecordAddenda(c)
			},
			svc: &mockMedicalRecordAddendumService{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]*model.MedicalRecordAddendum, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAddendumSvc(tt.svc)
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
