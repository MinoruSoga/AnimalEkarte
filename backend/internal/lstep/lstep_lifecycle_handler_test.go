package lstep

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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

// ---- mock LstepLifecycleService ----

type mockLstepLifecycleService struct {
	handlePetDeathFn    func(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string, actorID *uint64) error
	handlePetRevivalFn  func(ctx context.Context, clinicID, petID uint64, actorID *uint64) error
	handleOwnerOptOutFn func(ctx context.Context, clinicID, ownerID uint64, reason string) error
	handleOwnerOptInFn  func(ctx context.Context, clinicID, ownerID uint64) error
	handleOwnerDeleteFn func(ctx context.Context, clinicID, ownerID uint64) error
}

func (m *mockLstepLifecycleService) HandlePetDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string, actorID *uint64) error {
	if m.handlePetDeathFn != nil {
		return m.handlePetDeathFn(ctx, clinicID, petID, deceasedAt, reason, actorID)
	}
	return nil
}
func (m *mockLstepLifecycleService) HandlePetRevival(ctx context.Context, clinicID, petID uint64, actorID *uint64) error {
	if m.handlePetRevivalFn != nil {
		return m.handlePetRevivalFn(ctx, clinicID, petID, actorID)
	}
	return nil
}
func (m *mockLstepLifecycleService) HandleOwnerOptOut(ctx context.Context, clinicID, ownerID uint64, reason string, actorID *uint64) error {
	if m.handleOwnerOptOutFn != nil {
		return m.handleOwnerOptOutFn(ctx, clinicID, ownerID, reason)
	}
	return nil
}
func (m *mockLstepLifecycleService) HandleOwnerOptIn(ctx context.Context, clinicID, ownerID uint64, actorID *uint64) error {
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

type mockOwnerLineLinker struct {
	linkLineUserIDFn func(context.Context, uint64, uint64, *string, *uint64) error
}

func (m *mockOwnerLineLinker) LinkLineUserID(ctx context.Context, clinicID, ownerID uint64, lineUserID *string, actorID *uint64) error {
	if m.linkLineUserIDFn != nil {
		return m.linkLineUserIDFn(ctx, clinicID, ownerID, lineUserID, actorID)
	}
	return nil
}

func newHandlerWithLstepLifecycleSvc(lc LstepLifecycleService, ow OwnerLineLinker) *Handler {
	if ow == nil {
		ow = &mockOwnerLineLinker{}
	}
	return &Handler{
		lifecycle:       lc,
		ownerLineLinker: ow,
	}
}

func newPatchPetDeathRouter(lc LstepLifecycleService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepLifecycleSvc(lc, nil)
	if withClinicID {
		r.PATCH("/pets/:id/death", func(c *gin.Context) { setClinicID(c) }, h.UpdatePetDeath)
	} else {
		r.PATCH("/pets/:id/death", h.UpdatePetDeath)
	}
	return r
}

func newDeletePetDeathRouter(lc LstepLifecycleService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepLifecycleSvc(lc, nil)
	if withClinicID {
		r.DELETE("/pets/:id/death", func(c *gin.Context) { setClinicID(c) }, h.DeletePetDeath)
	} else {
		r.DELETE("/pets/:id/death", h.DeletePetDeath)
	}
	return r
}

func newPostOwnerLstepOptOutRouter(lc LstepLifecycleService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepLifecycleSvc(lc, nil)
	if withClinicID {
		r.POST("/owners/:id/lstep-opt-out", func(c *gin.Context) { setClinicID(c) }, h.UpdateOwnerLstepOptOut)
	} else {
		r.POST("/owners/:id/lstep-opt-out", h.UpdateOwnerLstepOptOut)
	}
	return r
}

func newPatchOwnerLstepOptOutRouter(lc LstepLifecycleService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepLifecycleSvc(lc, nil)
	if withClinicID {
		r.PATCH("/owners/:id/lstep/opt-out", func(c *gin.Context) { setClinicID(c) }, h.PatchOwnerLstepOptOut)
	} else {
		r.PATCH("/owners/:id/lstep/opt-out", h.PatchOwnerLstepOptOut)
	}
	return r
}

func newDeleteOwnerLineRouter(ow OwnerLineLinker, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepLifecycleSvc(&mockLstepLifecycleService{}, ow)
	if withClinicID {
		r.DELETE("/owners/:id/line", func(c *gin.Context) { setClinicID(c) }, h.DeleteOwnerLine)
	} else {
		r.DELETE("/owners/:id/line", h.DeleteOwnerLine)
	}
	return r
}

// ---- UpdatePetDeath ----

func TestPatchPetDeath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       map[string]any
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:    "returns 204 on success",
			paramID: "10",
			body:    map[string]any{"deceased_at": "2026-04-01", "reason": "自然死"},
			svc: &mockLstepLifecycleService{
				handlePetDeathFn: func(_ context.Context, clinicID, petID uint64, _ time.Time, reason string, _ *uint64) error {
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
			body:       map[string]any{"deceased_at": "2026-04-01"},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when deceased_at missing",
			paramID:    "10",
			body:       map[string]any{"reason": "外傷"},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when pet id is not numeric",
			paramID:    "abc",
			body:       map[string]any{"deceased_at": "2026-04-01"},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 500 on service error",
			paramID: "10",
			body:    map[string]any{"deceased_at": "2026-04-01"},
			svc: &mockLstepLifecycleService{
				handlePetDeathFn: func(_ context.Context, _, _ uint64, _ time.Time, _ string, _ *uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "returns 409 on conflict error",
			paramID: "10",
			body:    map[string]any{"deceased_at": "2026-04-01"},
			svc: &mockLstepLifecycleService{
				handlePetDeathFn: func(_ context.Context, _, _ uint64, _ time.Time, _ string, _ *uint64) error {
					return apperrors.WrapConflict("死亡記録は既に登録されています")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:    "returns 404 on not-found error",
			paramID: "99",
			body:    map[string]any{"deceased_at": "2026-04-01"},
			svc: &mockLstepLifecycleService{
				handlePetDeathFn: func(_ context.Context, _, _ uint64, _ time.Time, _ string, _ *uint64) error {
					return apperrors.WrapNotFound("pet", "99")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchPetDeathRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			raw, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/pets/"+tt.paramID+"/death", bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestPatchPetDeathDateValidationJST(t *testing.T) {
	gin.SetMode(gin.TestMode)

	todayJST := time.Now().In(config.JST)
	tests := []struct {
		name             string
		body             map[string]any
		wantStatus       int
		wantServiceCalls int
	}{
		{name: "missing", body: map[string]any{"reason": "任意"}, wantStatus: http.StatusBadRequest},
		{name: "empty", body: map[string]any{"deceased_at": ""}, wantStatus: http.StatusBadRequest},
		{name: "invalid calendar date", body: map[string]any{"deceased_at": "2026-02-30"}, wantStatus: http.StatusBadRequest},
		{name: "invalid timezone", body: map[string]any{"deceased_at": "2026-08-03T00:00:00+25:00"}, wantStatus: http.StatusBadRequest},
		{name: "future in JST", body: map[string]any{"deceased_at": todayJST.AddDate(0, 0, 2).Format(time.DateOnly)}, wantStatus: http.StatusBadRequest},
		{name: "today in JST", body: map[string]any{"deceased_at": todayJST.Format(time.DateOnly)}, wantStatus: http.StatusNoContent, wantServiceCalls: 1},
		{name: "past in JST", body: map[string]any{"deceased_at": todayJST.AddDate(0, 0, -1).Format(time.DateOnly)}, wantStatus: http.StatusNoContent, wantServiceCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalls := 0
			svc := &mockLstepLifecycleService{
				handlePetDeathFn: func(_ context.Context, _, _ uint64, _ time.Time, _ string, _ *uint64) error {
					serviceCalls++
					return nil
				},
			}
			router := newPatchPetDeathRouter(svc, true)
			raw, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/pets/10/death", bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantServiceCalls, serviceCalls)
		})
	}
}

// ---- DeletePetDeath ----

func TestDeletePetDeath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:    "returns 204 on success",
			paramID: "5",
			svc: &mockLstepLifecycleService{
				handlePetRevivalFn: func(_ context.Context, clinicID, petID uint64, _ *uint64) error {
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
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is not numeric",
			paramID:    "bad",
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 on not-found error",
			paramID: "99",
			svc: &mockLstepLifecycleService{
				handlePetRevivalFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					return apperrors.WrapNotFound("pet", "99")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 on conflict error",
			paramID: "5",
			svc: &mockLstepLifecycleService{
				handlePetRevivalFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					return apperrors.WrapConflict("死亡記録が登録されていないため解除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeletePetDeathRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/pets/"+tt.paramID+"/death", http.NoBody)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeletePetDeath_ClinicAliasUsesAuthenticatedClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	svc := &mockLstepLifecycleService{
		handlePetRevivalFn: func(_ context.Context, clinicID, petID uint64, _ *uint64) error {
			called = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(22), petID)
			return nil
		},
	}
	h := newHandlerWithLstepLifecycleSvc(svc, nil)
	r := gin.New()
	r.DELETE("/clinics/:clinic_id/pets/:id/death", func(c *gin.Context) {
		c.Set("clinic_id", "1")
	}, h.DeletePetDeath)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/clinics/999/pets/22/death", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, called)
}

// ---- UpdateOwnerLstepOptOut ----

func TestPostOwnerLstepOptOut(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       map[string]any
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:    "returns 204 with reason",
			paramID: "3",
			body:    map[string]any{"reason": "クレーム"},
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
			body:       map[string]any{},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "3",
			body:       map[string]any{"reason": "test"},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is not numeric",
			paramID:    "xyz",
			body:       map[string]any{},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPostOwnerLstepOptOutRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)

			raw, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/owners/"+tt.paramID+"/lstep-opt-out", bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

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
		body       map[string]any
		svc        *mockLstepLifecycleService
		wantStatus int
	}{
		{
			name:    "opt_out=true calls HandleOwnerOptOut",
			paramID: "2",
			body:    map[string]any{"opt_out": true, "reason": "解約"},
			svc: &mockLstepLifecycleService{
				handleOwnerOptOutFn: func(_ context.Context, _, _ uint64, reason string) error {
					assert.Equal(t, "解約", reason)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "opt_out=false calls HandleOwnerOptIn",
			paramID: "2",
			body:    map[string]any{"opt_out": false},
			svc: &mockLstepLifecycleService{
				handleOwnerOptInFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "2",
			body:       map[string]any{"opt_out": true},
			svc:        &mockLstepLifecycleService{},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchOwnerLstepOptOutRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)

			raw, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/owners/"+tt.paramID+"/lstep/opt-out", bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

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
		ownerSvc   *mockOwnerLineLinker
		wantStatus int
	}{
		{
			name:    "returns 204 on success (unlinks LINE user)",
			paramID: "4",
			ownerSvc: &mockOwnerLineLinker{
				linkLineUserIDFn: func(_ context.Context, clinicID, id uint64, lineUserID *string, _ *uint64) error {
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
			ownerSvc:   &mockOwnerLineLinker{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is not numeric",
			paramID:    "bad",
			ownerSvc:   &mockOwnerLineLinker{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when owner not found",
			paramID: "999",
			ownerSvc: &mockOwnerLineLinker{
				linkLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string, _ *uint64) error {
					return apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteOwnerLineRouter(tt.ownerSvc, tt.wantStatus != http.StatusUnauthorized)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/owners/"+tt.paramID+"/line", http.NoBody)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
