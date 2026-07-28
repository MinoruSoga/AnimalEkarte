package pet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestUpdatePet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockPetServiceHandler
		wantStatus int
	}{
		{
			name:     "updates pet successfully",
			paramID:  "1",
			body:     map[string]any{"name": "タマ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdatePetInput) (*model.Pet, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "タマ", *input.Name)
					return &model.Pet{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "タマ"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "タマ"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPetServiceHandler{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "ignores status in PATCH body (status write is not part of generic update, BUG-415)",
			paramID:  "1",
			body:     map[string]any{"name": "タマ", "status": "deceased"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdatePetInput) (*model.Pet, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "タマ", *input.Name)
					return &model.Pet{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns 404 when pet not found",
			paramID:  "999",
			body:     map[string]any{"name": "タマ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPetServiceHandler{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdatePetInput) (*model.Pet, error) {
					return nil, apperrors.WrapNotFound("pet", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPetSvcHandler(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdatePet(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
