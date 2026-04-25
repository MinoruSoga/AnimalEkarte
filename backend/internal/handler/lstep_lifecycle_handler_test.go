package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock LstepLifecycleService ----

type mockLstepLifecycleService struct {
	handlePetDeathFn    func(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error
	handlePetRevivalFn  func(ctx context.Context, clinicID, petID uint64) error
	handleOwnerOptOutFn func(ctx context.Context, clinicID, ownerID uint64, reason string) error
	handleOwnerOptInFn  func(ctx context.Context, clinicID, ownerID uint64) error
	handleOwnerDeleteFn func(ctx context.Context, clinicID, ownerID uint64) error
}

func (m *mockLstepLifecycleService) HandlePetDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error {
	if m.handlePetDeathFn != nil {
		return m.handlePetDeathFn(ctx, clinicID, petID, deceasedAt, reason)
	}
	return nil
}
func (m *mockLstepLifecycleService) HandlePetRevival(ctx context.Context, clinicID, petID uint64) error {
	if m.handlePetRevivalFn != nil {
		return m.handlePetRevivalFn(ctx, clinicID, petID)
	}
	return nil
}
func (m *mockLstepLifecycleService) HandleOwnerOptOut(ctx context.Context, clinicID, ownerID uint64, reason string) error {
	if m.handleOwnerOptOutFn != nil {
		return m.handleOwnerOptOutFn(ctx, clinicID, ownerID, reason)
	}
	return nil
}
func (m *mockLstepLifecycleService) HandleOwnerOptIn(ctx context.Context, clinicID, ownerID uint64) error {
	if m.handleOwnerOptInFn != nil {
		return m.handleOwnerOptInFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepLifecycleService) HandleOwnerDeletion(ctx context.Context, clinicID, ownerID uint64) error {
	if m.handleOwnerDeleteFn != nil {
		return m.handleOwnerDeleteFn(ctx, clinicID, ownerID)
	}
	return nil
}

// ---- helpers ----

func newHandlerWithLstepLifecycleSvc(lc service.LstepLifecycleService, ow service.OwnerService) *Handler {
	if ow == nil {
		ow = &mockOwnerService{}
	}
	return &Handler{
		svc: &service.Services{
			LstepLifecycle: lc,
			Owner:          ow,
		},
	}
}

// ---- PatchPetDeath ----

func TestPatchPetDeath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		body       map[string]any
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:     "returns 204 on success",
			paramID:  "10",
			setupCtx: setClinicID,
			body:     map[string]any{"deceased_at": "2026-04-01", "reason": "自然死"},
			svc: &mockLstepLifecycleService{
				handlePetDeathFn: func(_ context.Context, clinicID, petID uint64, _ time.Time, reason string) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), petID)
					assert.Equal(t, "自然死", reason)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "10",
			setupCtx:   func(_ *gin.Context) {},
			body:       map[string]any{"deceased_at": "2026-04-01"},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when deceased_at missing",
			paramID:    "10",
			setupCtx:   setClinicID,
			body:       map[string]any{"reason": "外傷"},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when pet id is not numeric",
			paramID:    "abc",
			setupCtx:   setClinicID,
			body:       map[string]any{"deceased_at": "2026-04-01"},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "10",
			setupCtx: setClinicID,
			body:     map[string]any{"deceased_at": "2026-04-01"},
			svc: &mockLstepLifecycleService{
				handlePetDeathFn: func(_ context.Context, _, _ uint64, _ time.Time, _ string) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLstepLifecycleSvc(tt.svc, nil)

			raw, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(raw))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.PatchPetDeath(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeletePetDeath ----

func TestDeletePetDeath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:     "returns 204 on success",
			paramID:  "5",
			setupCtx: setClinicID,
			svc: &mockLstepLifecycleService{
				handlePetRevivalFn: func(_ context.Context, clinicID, petID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), petID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "5",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is not numeric",
			paramID:    "bad",
			setupCtx:   setClinicID,
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 on not-found error",
			paramID:  "99",
			setupCtx: setClinicID,
			svc: &mockLstepLifecycleService{
				handlePetRevivalFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("pet", "99")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLstepLifecycleSvc(tt.svc, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.DeletePetDeath(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- PostOwnerLstepOptOut ----

func TestPostOwnerLstepOptOut(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		body       map[string]any
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:     "returns 204 with reason",
			paramID:  "3",
			setupCtx: setClinicID,
			body:     map[string]any{"reason": "クレーム"},
			svc: &mockLstepLifecycleService{
				handleOwnerOptOutFn: func(_ context.Context, clinicID, ownerID uint64, reason string) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), ownerID)
					assert.Equal(t, "クレーム", reason)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 204 with empty body (reason optional)",
			paramID:    "3",
			setupCtx:   setClinicID,
			body:       map[string]any{},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "3",
			setupCtx:   func(_ *gin.Context) {},
			body:       map[string]any{"reason": "test"},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is not numeric",
			paramID:    "xyz",
			setupCtx:   setClinicID,
			body:       map[string]any{},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLstepLifecycleSvc(tt.svc, nil)

			raw, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(raw))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.PostOwnerLstepOptOut(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteOwnerLstepOptOut ----

func TestDeleteOwnerLstepOptOut(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:     "returns 204 on success (opt-in)",
			paramID:  "7",
			setupCtx: setClinicID,
			svc: &mockLstepLifecycleService{
				handleOwnerOptInFn: func(_ context.Context, clinicID, ownerID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), ownerID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "7",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is not numeric",
			paramID:    "nope",
			setupCtx:   setClinicID,
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLstepLifecycleSvc(tt.svc, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.DeleteOwnerLstepOptOut(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- PatchOwnerLstepOptOut (unified) ----

func TestPatchOwnerLstepOptOut(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		body       map[string]any
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:     "opt_out=true calls HandleOwnerOptOut",
			paramID:  "2",
			setupCtx: setClinicID,
			body:     map[string]any{"opt_out": true, "reason": "解約"},
			svc: &mockLstepLifecycleService{
				handleOwnerOptOutFn: func(_ context.Context, _, _ uint64, reason string) error {
					assert.Equal(t, "解約", reason)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:     "opt_out=false calls HandleOwnerOptIn",
			paramID:  "2",
			setupCtx: setClinicID,
			body:     map[string]any{"opt_out": false},
			svc: &mockLstepLifecycleService{
				handleOwnerOptInFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "2",
			setupCtx:   func(_ *gin.Context) {},
			body:       map[string]any{"opt_out": true},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLstepLifecycleSvc(tt.svc, nil)

			raw, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(raw))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.PatchOwnerLstepOptOut(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteOwnerLine ----

func TestDeleteOwnerLine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		ownerSvc   *mockOwnerService
		wantStatus int
	}{
		{
			name:     "returns 204 on success (unlinks LINE user)",
			paramID:  "4",
			setupCtx: setClinicID,
			ownerSvc: &mockOwnerService{
				linkLineUserIDFn: func(_ context.Context, clinicID, id uint64, lineUserID *string) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(4), id)
					assert.Nil(t, lineUserID, "unlink must pass nil")
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "4",
			setupCtx:   func(_ *gin.Context) {},
			ownerSvc:   &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is not numeric",
			paramID:    "bad",
			setupCtx:   setClinicID,
			ownerSvc:   &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when owner not found",
			paramID:  "999",
			setupCtx: setClinicID,
			ownerSvc: &mockOwnerService{
				linkLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
					return apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLstepLifecycleSvc(&mockLstepLifecycleService{}, tt.ownerSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.DeleteOwnerLine(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
