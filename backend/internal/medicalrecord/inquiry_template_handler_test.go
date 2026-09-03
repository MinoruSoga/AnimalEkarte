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

// ---- mock InquiryTemplateService ----

type mockInquiryTemplateService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateInquiryTemplateInput) (*model.InquiryTemplate, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockInquiryTemplateService) List(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockInquiryTemplateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.InquiryTemplate{ID: id}, nil
}

func (m *mockInquiryTemplateService) Create(ctx context.Context, clinicID uint64, input *CreateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.InquiryTemplate{ID: 1, Title: input.Title}, nil
}

func (m *mockInquiryTemplateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.InquiryTemplate{ID: id}, nil
}

func (m *mockInquiryTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockInquiryTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- helper ----

func newHandlerWithInquiryTemplateSvc(svc InquiryTemplateService) *InquiryTemplateHandler {
	return NewInquiryTemplateHandler(svc)
}

// ---- TestInquiryTemplateHandler_List ----

func TestInquiryTemplateHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockInquiryTemplateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "正常: 一覧を返す",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInquiryTemplateService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.InquiryTemplate{
						{ID: 1, ClinicID: 1, Title: "問診テンプレート1"},
						{ID: 2, ClinicID: 1, Title: "問診テンプレート2"},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"問診テンプレート1"`,
		},
		{
			name:       "401: clinic_id なし",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "500: サービスエラー",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInquiryTemplateService{
				listFn: func(_ context.Context, _ uint64) ([]model.InquiryTemplate, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInquiryTemplateSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.ListInquiryTemplates(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestInquiryTemplateHandler_GetByID ----

func TestInquiryTemplateHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockInquiryTemplateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "正常: テンプレートを返す",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInquiryTemplateService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					return &model.InquiryTemplate{ID: id, Title: "問診テンプレート"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"問診テンプレート"`,
		},
		{
			name:       "401: clinic_id なし",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400: 不正なID",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "404: 存在しないID",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInquiryTemplateService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.InquiryTemplate, error) {
					return nil, apperrors.WrapNotFound("inquiry_template", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInquiryTemplateSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetInquiryTemplate(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestInquiryTemplateHandler_Create ----

func TestInquiryTemplateHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"title":    "新規テンプレート",
			"category": "pre-visit",
			"content":  "問診内容",
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockInquiryTemplateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "201: 正常作成",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInquiryTemplateService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateInquiryTemplateInput) (*model.InquiryTemplate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "新規テンプレート", input.Title)
					return &model.InquiryTemplate{ID: 5, ClinicID: clinicID, Title: input.Title}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"title":"新規テンプレート"`,
		},
		{
			name:       "401: clinic_id なし",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400: title なし（バリデーションエラー）",
			body:       map[string]any{"category": "pre-visit"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400: 不正なJSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "500: サービスエラー",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInquiryTemplateService{
				createFn: func(_ context.Context, _ uint64, _ *CreateInquiryTemplateInput) (*model.InquiryTemplate, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInquiryTemplateSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateInquiryTemplate(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusCreated {
				assert.Contains(t, w.Header().Get("Location"), "/inquiry-templates/")
			}
		})
	}
}

// ---- TestInquiryTemplateHandler_Update ----

func TestInquiryTemplateHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockInquiryTemplateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "200: 正常更新",
			paramID:  "1",
			body:     map[string]any{"title": "更新されたタイトル"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInquiryTemplateService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Title)
					assert.Equal(t, "更新されたタイトル", *input.Title)
					return &model.InquiryTemplate{ID: id, Title: *input.Title}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"更新されたタイトル"`,
		},
		{
			name:       "401: clinic_id なし",
			paramID:    "1",
			body:       map[string]any{"title": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400: 不正なID",
			paramID:    "xyz",
			body:       map[string]any{"title": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400: 不正なJSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "404: 存在しないID",
			paramID:  "999",
			body:     map[string]any{"title": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInquiryTemplateService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error) {
					return nil, apperrors.WrapNotFound("inquiry_template", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInquiryTemplateSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateInquiryTemplate(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestInquiryTemplateHandler_Delete ----

func newDeleteInquiryTemplateRouter(svc InquiryTemplateService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithInquiryTemplateSvc(svc)
	r.DELETE("/inquiry-templates/:id", func(c *gin.Context) {
		setClinicID(c)
		h.DeleteInquiryTemplate(c)
	})
	return r
}

func TestInquiryTemplateHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockInquiryTemplateService
		wantStatus int
	}{
		{
			name:    "204: 正常削除",
			paramID: "1",
			svc: &mockInquiryTemplateService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "409: 使用中のため削除不可",
			paramID: "2",
			svc: &mockInquiryTemplateService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この問診定型文は使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:    "404: 存在しないID",
			paramID: "999",
			svc: &mockInquiryTemplateService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("inquiry_template", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteInquiryTemplateRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/inquiry-templates/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("401: clinic_id なし", func(t *testing.T) {
		h := newHandlerWithInquiryTemplateSvc(&mockInquiryTemplateService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteInquiryTemplate(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- TestInquiryTemplateHandler_Reorder ----

func newReorderInquiryTemplatesRouter(svc InquiryTemplateService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithInquiryTemplateSvc(svc)
	r.POST("/inquiry-templates/reorder", func(c *gin.Context) {
		setClinicID(c)
		h.ReorderInquiryTemplates(c)
	})
	return r
}

func TestInquiryTemplateHandler_Reorder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200: 正常並び替え", func(t *testing.T) {
		svc := &mockInquiryTemplateService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderInquiryTemplatesRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/inquiry-templates/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("400: 空のIDsエラー", func(t *testing.T) {
		svc := &mockInquiryTemplateService{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return apperrors.WrapInvalidInput("IDsは1件以上必要です")
			},
		}
		router := newReorderInquiryTemplatesRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/inquiry-templates/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("401: clinic_id なし", func(t *testing.T) {
		h := newHandlerWithInquiryTemplateSvc(&mockInquiryTemplateService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderInquiryTemplates(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestInquiryTemplateSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*InquiryTemplateHandler, *gin.Context)
		svc    *mockInquiryTemplateService
	}{
		{
			name: "ListInquiryTemplates returns 403 when selected clinic lacks master-medical view grant",
			invoke: func(h *InquiryTemplateHandler, c *gin.Context) {
				h.ListInquiryTemplates(c)
			},
			svc: &mockInquiryTemplateService{
				listFn: func(_ context.Context, _ uint64) ([]model.InquiryTemplate, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "GetInquiryTemplate returns 403 when selected clinic lacks master-medical view grant",
			invoke: func(h *InquiryTemplateHandler, c *gin.Context) {
				h.GetInquiryTemplate(c)
			},
			svc: &mockInquiryTemplateService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.InquiryTemplate, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInquiryTemplateSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMasterMedical), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
