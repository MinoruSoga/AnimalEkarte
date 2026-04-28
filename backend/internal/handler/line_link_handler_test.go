package handler

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

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock LineLinkService ----

type mockLineLinkService struct {
	generateLinkTokenFn func(ctx context.Context, clinicID, ownerID uint64) (*service.LinkTokenResult, error)
	linkAccountFn       func(ctx context.Context, clinicID uint64, input service.LinkAccountInput) (*model.Owner, error)
	handleWebhookFn     func(ctx context.Context, body []byte, signature string) error
}

func (m *mockLineLinkService) GenerateLinkToken(ctx context.Context, clinicID, ownerID uint64) (*service.LinkTokenResult, error) {
	if m.generateLinkTokenFn != nil {
		return m.generateLinkTokenFn(ctx, clinicID, ownerID)
	}
	return &service.LinkTokenResult{Token: "tok123", LiffURL: "https://liff.example.com"}, nil
}
func (m *mockLineLinkService) LinkAccount(ctx context.Context, clinicID uint64, input service.LinkAccountInput) (*model.Owner, error) {
	if m.linkAccountFn != nil {
		return m.linkAccountFn(ctx, clinicID, input)
	}
	return &model.Owner{ID: 1}, nil
}
func (m *mockLineLinkService) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	if m.handleWebhookFn != nil {
		return m.handleWebhookFn(ctx, body, signature)
	}
	return nil
}

// ---- helpers ----

func newHandlerWithLineLinkSvc(svc service.LineLinkService) *Handler {
	return &Handler{svc: &service.Services{LineLink: svc}}
}

func newPostGenerateLineLinkTokenRouter(svc service.LineLinkService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLineLinkSvc(svc)
	if withClinicID {
		r.POST("/owners/:id/line/link-token", func(c *gin.Context) { setClinicID(c) }, h.GenerateLineLinkToken)
	} else {
		r.POST("/owners/:id/line/link-token", h.GenerateLineLinkToken)
	}
	return r
}

func newPostLiffLinkAccountRouter(svc service.LineLinkService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLineLinkSvc(svc)
	r.POST("/liff/:clinicId/link", h.LinkLiffAccount)
	return r
}

func newPostReceiveLineWebhookRouter(svc service.LineLinkService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLineLinkSvc(svc)
	r.POST("/line/webhook", h.ReceiveLineWebhook)
	return r
}

// ---- tests ----

func TestPostGenerateLineLinkToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		ownerID    string
		svc        *mockLineLinkService
		wantStatus int
	}{
		{
			name:       "201 success",
			ownerID:    "1",
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "401 no clinic",
			ownerID:    "1",
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400 bad owner ID",
			ownerID:    "abc",
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "404 owner not found",
			ownerID: "99",
			svc: &mockLineLinkService{
				generateLinkTokenFn: func(_ context.Context, _, _ uint64) (*service.LinkTokenResult, error) {
					return nil, apperrors.WrapNotFound("owner", "99")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "500 service error",
			ownerID: "1",
			svc: &mockLineLinkService{
				generateLinkTokenFn: func(_ context.Context, _, _ uint64) (*service.LinkTokenResult, error) {
					return nil, errors.New("unexpected error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPostGenerateLineLinkTokenRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodPost, "/owners/"+tt.ownerID+"/line/link-token", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestPostLiffLinkAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validBody, _ := json.Marshal(map[string]string{
		"link_token":    "tok123",
		"line_id_token": "eyJhbGci...",
	})
	noLinkTokenBody, _ := json.Marshal(map[string]string{"line_id_token": "eyJhbGci..."})
	noLineIDTokenBody, _ := json.Marshal(map[string]string{"link_token": "tok123"})

	tests := []struct {
		name       string
		clinicID   string
		body       []byte
		svc        *mockLineLinkService
		wantStatus int
	}{
		{
			name:       "200 success",
			clinicID:   "1",
			body:       validBody,
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 bad clinic ID",
			clinicID:   "abc",
			body:       validBody,
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 missing link_token",
			clinicID:   "1",
			body:       noLinkTokenBody,
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 missing line_id_token",
			clinicID:   "1",
			body:       noLineIDTokenBody,
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "404 owner not found",
			clinicID: "1",
			body:     validBody,
			svc: &mockLineLinkService{
				linkAccountFn: func(_ context.Context, _ uint64, _ service.LinkAccountInput) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "link_token")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "500 service error",
			clinicID: "1",
			body:     validBody,
			svc: &mockLineLinkService{
				linkAccountFn: func(_ context.Context, _ uint64, _ service.LinkAccountInput) (*model.Owner, error) {
					return nil, errors.New("unexpected error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPostLiffLinkAccountRouter(tt.svc)
			req := httptest.NewRequest(http.MethodPost, "/liff/"+tt.clinicID+"/link", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestPostReceiveLineWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		signature  string
		body       []byte
		svc        *mockLineLinkService
		wantStatus int
	}{
		{
			name:       "200 success",
			signature:  "sha256=abcdef1234",
			body:       []byte(`{"destination":"U123","events":[]}`),
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 missing X-Line-Signature",
			signature:  "",
			body:       []byte(`{"destination":"U123","events":[]}`),
			svc:        &mockLineLinkService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "500 service error",
			signature: "sha256=abcdef1234",
			body:      []byte(`{"destination":"U123","events":[]}`),
			svc: &mockLineLinkService{
				handleWebhookFn: func(_ context.Context, _ []byte, _ string) error {
					return errors.New("signature mismatch")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPostReceiveLineWebhookRouter(tt.svc)
			req := httptest.NewRequest(http.MethodPost, "/line/webhook", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.signature != "" {
				req.Header.Set("X-Line-Signature", tt.signature)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
