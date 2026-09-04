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
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock LineSendService ----

type mockLineSendService struct {
	sendFn    func(ctx context.Context, clinicID uint64, input *SendLineMessageInput) (*SendLineMessageResult, error)
	getLogsFn func(ctx context.Context, clinicID, ownerID uint64) ([]*model.LineSendLog, error)
}

func (m *mockLineSendService) Send(ctx context.Context, clinicID uint64, input *SendLineMessageInput) (*SendLineMessageResult, error) {
	if m.sendFn != nil {
		return m.sendFn(ctx, clinicID, input)
	}
	return &SendLineMessageResult{SentAt: time.Now()}, nil
}

func (m *mockLineSendService) GetSendLogs(ctx context.Context, clinicID, ownerID uint64) ([]*model.LineSendLog, error) {
	if m.getLogsFn != nil {
		return m.getLogsFn(ctx, clinicID, ownerID)
	}
	return nil, nil
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

func TestGetLineSendLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fixedAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	errMsg := "LINE send failed"

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(*gin.Context)
		svc        *mockLineSendService
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:     "returns 200 with items list",
			paramID:  "1",
			setupCtx: setClinicID,
			svc: &mockLineSendService{
				getLogsFn: func(_ context.Context, _, _ uint64) ([]*model.LineSendLog, error) {
					return []*model.LineSendLog{
						{
							ID:             1,
							MessageType:    "text",
							ContentSummary: "hello",
							Status:         "sent",
							ErrorMessage:   nil,
							SentAt:         fixedAt,
						},
						{
							ID:             2,
							MessageType:    "pdf_url",
							ContentSummary: "report.pdf",
							Status:         "failed",
							ErrorMessage:   &errMsg,
							SentAt:         fixedAt,
						},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp lineSendLogListResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.Items, 2)
				assert.Equal(t, uint64(1), resp.Items[0].ID)
				assert.Equal(t, "text", resp.Items[0].MessageType)
				assert.Equal(t, "sent", resp.Items[0].Status)
				assert.Nil(t, resp.Items[0].ErrorMessage)
				assert.Equal(t, uint64(2), resp.Items[1].ID)
				assert.Equal(t, "failed", resp.Items[1].Status)
				assert.NotNil(t, resp.Items[1].ErrorMessage)
				assert.Equal(t, "2026-01-15T19:00:00+09:00", resp.Items[0].SentAt)
			},
		},
		{
			name:     "returns 200 with empty items list",
			paramID:  "1",
			setupCtx: setClinicID,
			svc: &mockLineSendService{
				getLogsFn: func(_ context.Context, _, _ uint64) ([]*model.LineSendLog, error) {
					return nil, nil
				},
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp lineSendLogListResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.NotNil(t, resp.Items)
				assert.Len(t, resp.Items, 0)
			},
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			setupCtx:   nil,
			svc:        &mockLineSendService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid owner id",
			paramID:    "xyz",
			setupCtx:   setClinicID,
			svc:        &mockLineSendService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: setClinicID,
			svc: &mockLineSendService{
				getLogsFn: func(_ context.Context, _, _ uint64) ([]*model.LineSendLog, error) {
					return nil, apperrors.Wrap(assert.AnError, "db error")
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

			req, err := http.NewRequest(http.MethodGet, "/owners/"+tt.paramID+"/line/send-logs", http.NoBody)
			require.NoError(t, err)
			c.Request = req

			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			if tt.setupCtx != nil {
				tt.setupCtx(c)
			}

			h.GetLineSendLogs(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.Bytes())
			}
		})
	}
}
