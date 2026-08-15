package lstep

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock LstepTriggerPriorityService ----

type mockLstepTriggerPriorityService struct {
	getByClinicIDFn    func(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error)
	updatePrioritiesFn func(ctx context.Context, clinicID uint64, input UpdateTriggerPrioritiesInput) error
	getPriorityForFn   func(ctx context.Context, clinicID uint64, triggerType string) (int, error)
}

func (m *mockLstepTriggerPriorityService) GetByClinicID(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error) {
	if m.getByClinicIDFn != nil {
		return m.getByClinicIDFn(ctx, clinicID)
	}
	return []model.LstepTriggerPriority{}, nil
}

func (m *mockLstepTriggerPriorityService) UpdatePriorities(ctx context.Context, clinicID uint64, input UpdateTriggerPrioritiesInput) error {
	if m.updatePrioritiesFn != nil {
		return m.updatePrioritiesFn(ctx, clinicID, input)
	}
	return nil
}

func (m *mockLstepTriggerPriorityService) GetPriorityFor(ctx context.Context, clinicID uint64, triggerType string) (int, error) {
	if m.getPriorityForFn != nil {
		return m.getPriorityForFn(ctx, clinicID, triggerType)
	}
	return model.DefaultPriorityFallback, nil
}

// ---- helpers ----

func newHandlerWithTriggerPrioritySvc(svc LstepTriggerPriorityService) *Handler {
	return &Handler{triggerPriority: svc}
}

func newGetTriggerPrioritiesRouter(svc LstepTriggerPriorityService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTriggerPrioritySvc(svc)
	if withClinicID {
		r.GET("/lstep/trigger-priorities", func(c *gin.Context) { setClinicID(c) }, h.GetLstepTriggerPriorities)
	} else {
		r.GET("/lstep/trigger-priorities", h.GetLstepTriggerPriorities)
	}
	return r
}

func newPatchTriggerPrioritiesRouter(svc LstepTriggerPriorityService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTriggerPrioritySvc(svc)
	if withClinicID {
		r.PATCH("/lstep/trigger-priorities", func(c *gin.Context) { setClinicID(c) }, h.UpdateLstepTriggerPriorities)
	} else {
		r.PATCH("/lstep/trigger-priorities", h.UpdateLstepTriggerPriorities)
	}
	return r
}

// ---- tests ----

func TestGetLstepTriggerPriorities(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		svc        *mockLstepTriggerPriorityService
		wantStatus int
	}{
		{
			name: "正常系: 優先順位一覧を返す",
			svc: &mockLstepTriggerPriorityService{
				getByClinicIDFn: func(_ context.Context, _ uint64) ([]model.LstepTriggerPriority, error) {
					return []model.LstepTriggerPriority{
						{ClinicID: 1, TriggerType: model.TriggerTypeDormantPrevention365, Priority: 2},
						{ClinicID: 1, TriggerType: model.TriggerTypeCheckupFollowUp, Priority: 3},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "異常系: clinic_id 未設定 → 401",
			svc:        &mockLstepTriggerPriorityService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "異常系: service エラー → 500",
			svc: &mockLstepTriggerPriorityService{
				getByClinicIDFn: func(_ context.Context, _ uint64) ([]model.LstepTriggerPriority, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withClinicID := tt.wantStatus != http.StatusUnauthorized
			router := newGetTriggerPrioritiesRouter(tt.svc, withClinicID)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/lstep/trigger-priorities", http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var body struct {
					ClinicID string `json:"clinic_id"`
					Items    []struct {
						TriggerType string `json:"trigger_type"`
						Priority    int    `json:"priority"`
					} `json:"items"`
				}
				assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
				assert.Equal(t, "1", body.ClinicID)
				assert.Len(t, body.Items, 2)
				assert.Equal(t, model.TriggerTypeDormantPrevention365, body.Items[0].TriggerType)
				assert.Equal(t, 2, body.Items[0].Priority)
			}
		})
	}
}

func TestUpdateLstepTriggerPriorities(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		body       any
		svc        *mockLstepTriggerPriorityService
		wantStatus int
	}{
		{
			name: "正常系: 優先順位を一括更新",
			body: map[string]any{
				"items": []map[string]any{
					{"trigger_type": model.TriggerTypeDormantPrevention365, "priority": 1},
					{"trigger_type": model.TriggerTypeCheckupFollowUp, "priority": 2},
				},
			},
			svc:        &mockLstepTriggerPriorityService{},
			wantStatus: http.StatusOK,
		},
		{
			name: "異常系: items 空 → 400",
			body: map[string]any{"items": []any{}},
			svc:  &mockLstepTriggerPriorityService{},
			// binding:"min=1" により ShouldBindJSON が 400 を返す
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: priority=0 (min=1 違反) → 400",
			body: map[string]any{
				"items": []map[string]any{
					{"trigger_type": model.TriggerTypeDormantPrevention365, "priority": 0},
				},
			},
			svc:        &mockLstepTriggerPriorityService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: service エラー → 500",
			body: map[string]any{
				"items": []map[string]any{
					{"trigger_type": model.TriggerTypeDormantPrevention365, "priority": 1},
				},
			},
			svc: &mockLstepTriggerPriorityService{
				updatePrioritiesFn: func(_ context.Context, _ uint64, _ UpdateTriggerPrioritiesInput) error {
					return errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "異常系: clinic_id 未設定 → 401",
			body:       map[string]any{"items": []map[string]any{{"trigger_type": "x", "priority": 1}}},
			svc:        &mockLstepTriggerPriorityService{},
			wantStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withClinicID := tt.name != "異常系: clinic_id 未設定 → 401"
			router := newPatchTriggerPrioritiesRouter(tt.svc, withClinicID)
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPatch, "/lstep/trigger-priorities", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestLstepTriggerPriorities_ClinicAliasUsesAuthenticatedClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET", func(t *testing.T) {
		var gotClinicID uint64
		svc := &mockLstepTriggerPriorityService{
			getByClinicIDFn: func(_ context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error) {
				gotClinicID = clinicID
				return []model.LstepTriggerPriority{}, nil
			},
		}
		r := gin.New()
		h := newHandlerWithTriggerPrioritySvc(svc)
		r.GET("/clinics/:clinic_id/lstep/trigger-priorities", func(c *gin.Context) {
			c.Set("clinic_id", "1")
		}, h.GetLstepTriggerPriorities)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/clinics/999/lstep/trigger-priorities", http.NoBody)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uint64(1), gotClinicID)
	})

	t.Run("PATCH", func(t *testing.T) {
		var gotClinicID uint64
		svc := &mockLstepTriggerPriorityService{
			updatePrioritiesFn: func(_ context.Context, clinicID uint64, _ UpdateTriggerPrioritiesInput) error {
				gotClinicID = clinicID
				return nil
			},
		}
		r := gin.New()
		h := newHandlerWithTriggerPrioritySvc(svc)
		r.PATCH("/clinics/:clinic_id/lstep/trigger-priorities", func(c *gin.Context) {
			c.Set("clinic_id", "1")
		}, h.UpdateLstepTriggerPriorities)

		body := bytes.NewBufferString(`{"items":[{"trigger_type":"dormant_365d","priority":1}]}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/clinics/999/lstep/trigger-priorities", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uint64(1), gotClinicID)
	})
}
