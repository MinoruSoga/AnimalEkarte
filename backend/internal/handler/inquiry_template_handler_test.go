package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock InquiryTemplateService ----

type mockInquiryTemplateService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error)
	createFn  func(ctx context.Context, clinicID uint64, input *service.CreateInquiryTemplateInput) (*model.InquiryTemplate, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *service.UpdateInquiryTemplateInput) (*model.InquiryTemplate, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockInquiryTemplateService) List(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return []model.InquiryTemplate{}, nil
}
func (m *mockInquiryTemplateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.InquiryTemplate{ID: id, ClinicID: clinicID}, nil
}
func (m *mockInquiryTemplateService) Create(ctx context.Context, clinicID uint64, input *service.CreateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.InquiryTemplate{ClinicID: clinicID}, nil
}
func (m *mockInquiryTemplateService) Update(ctx context.Context, clinicID, id uint64, input *service.UpdateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.InquiryTemplate{ID: id, ClinicID: clinicID}, nil
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

func newHandlerWithInquiryTemplateSvc(svc service.InquiryTemplateService) *Handler {
	return &Handler{
		svc: &service.Services{InquiryTemplate: svc},
	}
}

// ---- DeleteInquiryTemplate ----

func newDeleteInquiryTemplateRouter(svc service.InquiryTemplateService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithInquiryTemplateSvc(svc)
	r.DELETE("/inquiry-templates/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteInquiryTemplate)
	return r
}

func TestDeleteInquiryTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockInquiryTemplateService
		wantStatus int
	}{
		{
			name:    "deletes inquiry template successfully",
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
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockInquiryTemplateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when template not found",
			paramID: "999",
			svc: &mockInquiryTemplateService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("inquiry_template", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when template is in use",
			paramID: "2",
			svc: &mockInquiryTemplateService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この問診テンプレートは使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
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

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithInquiryTemplateSvc(&mockInquiryTemplateService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteInquiryTemplate(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
