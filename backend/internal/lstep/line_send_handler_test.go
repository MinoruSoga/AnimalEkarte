package lstep

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ---- mock LineSendService ----

type mockLineSendService struct {
	sendFn func(ctx context.Context, clinicID uint64, input *SendLineMessageInput) (*SendLineMessageResult, error)
}

func (m *mockLineSendService) Send(ctx context.Context, clinicID uint64, input *SendLineMessageInput) (*SendLineMessageResult, error) {
	if m.sendFn != nil {
		return m.sendFn(ctx, clinicID, input)
	}
	return &SendLineMessageResult{SentAt: time.Now()}, nil
}

func newHandlerWithLineSendSvc(ls LineSendService) *LineSendHandler {
	return NewLineSendHandler(ls, func(_, _ string) gin.HandlerFunc { return func(_ *gin.Context) {} })
}

// ---- TestPostLineSend ----

func TestPostLineSend(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type bodyMap = map[string]any

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(*gin.Context)
		body       any
		svc        *mockLineSendService
		wantStatus int
	}{
		{
			name:    "returns 200 on success (text)",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			body:       bodyMap{"message_type": "text", "text": "hello"},
			svc:        &mockLineSendService{},
			wantStatus: http.StatusOK,
		},
		{
			name:    "returns 401 when clinic_id missing",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setStaffID(c)
			},
			body:       bodyMap{"message_type": "text", "text": "hello"},
			svc:        &mockLineSendService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "returns 401 when staff_id missing",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
			},
			body:       bodyMap{"message_type": "text", "text": "hello"},
			svc:        &mockLineSendService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "returns 400 on invalid owner id",
			paramID: "abc",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			body:       bodyMap{"message_type": "text", "text": "hello"},
			svc:        &mockLineSendService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 on malformed JSON",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			body:       nil, // will be sent as raw invalid bytes
			svc:        &mockLineSendService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 when message_type missing",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			body:       bodyMap{"text": "hello"},
			svc:        &mockLineSendService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 200 when text is LINE max 5000 chars",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			body: bodyMap{
				"message_type": "text",
				"text":         strings.Repeat("a", 5000),
			},
			svc:        &mockLineSendService{},
			wantStatus: http.StatusOK,
		},
		{
			name:    "returns 400 when text exceeds 5000 chars",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			body: bodyMap{
				"message_type": "text",
				"text":         strings.Repeat("a", 5001),
			},
			svc:        &mockLineSendService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 500 on service error",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				setStaffID(c)
			},
			body: bodyMap{"message_type": "text", "text": "hello"},
			svc: &mockLineSendService{
				sendFn: func(_ context.Context, _ uint64, _ *SendLineMessageInput) (*SendLineMessageResult, error) {
					return nil, apperrors.Wrap(assert.AnError, "send failed")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLineSendSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var bodyBytes []byte
			if tt.name == "returns 400 on malformed JSON" {
				bodyBytes = []byte(`{invalid json`)
			} else if tt.body != nil {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req, err := http.NewRequest(http.MethodPost, "/owners/"+tt.paramID+"/line/send", bytes.NewReader(bodyBytes))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			if tt.setupCtx != nil {
				tt.setupCtx(c)
			}

			h.SendLineMessage(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- TestGetLineSendLogs ----

// EMR-504: line_send_logs は DROP 済みで置換先の reader が未接続のため、履歴 GET は権限・入力検証の後
// 固定の 422 (line_dispatch_pending) で fail-closed する。空の一覧（=送信なし）は返さない。
func TestGetLineSendLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(*gin.Context)
		wantStatus int
		wantCode   string
	}{
		{
			name:       "returns 422 line_dispatch_pending for a valid owner",
			paramID:    "1",
			setupCtx:   setClinicID,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   apperrors.CodeLineDispatchPending,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			setupCtx:   nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid owner id",
			paramID:    "xyz",
			setupCtx:   setClinicID,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithLineSendSvc(&mockLineSendService{})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodGet, "/owners/"+tt.paramID+"/line/send-logs", http.NoBody)
			require.NoError(t, err)
			c.Request = req

			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			if tt.setupCtx != nil {
				tt.setupCtx(c)
			}

			h.GetLineSendLogs(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCode != "" {
				var body struct {
					Code string `json:"code"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				assert.Equal(t, tt.wantCode, body.Code)
			}
		})
	}
}
