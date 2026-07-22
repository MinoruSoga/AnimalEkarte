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
)

// ---- mock LstepTagService ----

type mockLstepTagService struct {
	getOwnerTagsFn   func(ctx context.Context, clinicID, ownerID uint64) (*OwnerTagsResult, error)
	addOwnerTagFn    func(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error
	removeOwnerTagFn func(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error
}

func (m *mockLstepTagService) GetOwnerTags(ctx context.Context, clinicID, ownerID uint64) (*OwnerTagsResult, error) {
	if m.getOwnerTagsFn != nil {
		return m.getOwnerTagsFn(ctx, clinicID, ownerID)
	}
	return &OwnerTagsResult{Tags: []string{}}, nil
}
func (m *mockLstepTagService) AddOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error {
	if m.addOwnerTagFn != nil {
		return m.addOwnerTagFn(ctx, clinicID, ownerID, tagName, actorID)
	}
	return nil
}
func (m *mockLstepTagService) RemoveOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error {
	if m.removeOwnerTagFn != nil {
		return m.removeOwnerTagFn(ctx, clinicID, ownerID, tagName, actorID)
	}
	return nil
}

// ---- helpers ----

func newHandlerWithLstepTagSvc(svc LstepTagService) *Handler {
	return &Handler{tag: svc}
}

func newGetOwnerLstepTagsRouter(svc LstepTagService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepTagSvc(svc)
	if withClinicID {
		r.GET("/owners/:id/lstep/tags", func(c *gin.Context) { setClinicID(c) }, h.GetOwnerLstepTags)
	} else {
		r.GET("/owners/:id/lstep/tags", h.GetOwnerLstepTags)
	}
	return r
}

func newPostOwnerLstepTagRouter(svc LstepTagService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepTagSvc(svc)
	if withClinicID {
		r.POST("/owners/:id/lstep/tags", func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") }, h.AddOwnerLstepTag)
	} else {
		r.POST("/owners/:id/lstep/tags", h.AddOwnerLstepTag)
	}
	return r
}

func newDeleteOwnerLstepTagRouter(svc LstepTagService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepTagSvc(svc)
	if withClinicID {
		r.DELETE("/owners/:id/lstep/tags/:tag_name", func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") }, h.DeleteOwnerLstepTag)
	} else {
		r.DELETE("/owners/:id/lstep/tags/:tag_name", h.DeleteOwnerLstepTag)
	}
	return r
}

func TestGetOwnerLstepTags_ClinicAliasUsesAuthenticatedClinic(t *testing.T) {
	var gotClinicID uint64
	svc := &mockLstepTagService{
		getOwnerTagsFn: func(_ context.Context, clinicID, _ uint64) (*OwnerTagsResult, error) {
			gotClinicID = clinicID
			return &OwnerTagsResult{Tags: []string{}}, nil
		},
	}
	r := gin.New()
	h := newHandlerWithLstepTagSvc(svc)
	r.GET("/clinics/:clinic_id/owners/:id/lstep/tags", func(c *gin.Context) {
		c.Set("clinic_id", "1")
	}, h.GetOwnerLstepTags)

	req := httptest.NewRequest(http.MethodGet, "/clinics/999/owners/10/lstep/tags", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint64(1), gotClinicID)
}

// ---- tests ----

func TestGetOwnerLstepTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		ownerID    string
		svc        *mockLstepTagService
		wantStatus int
	}{
		{
			name:       "200 linked owner",
			ownerID:    "1",
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusOK,
		},
		{
			name:    "200 unlinked returns empty tags",
			ownerID: "1",
			svc: &mockLstepTagService{
				getOwnerTagsFn: func(_ context.Context, _, _ uint64) (*OwnerTagsResult, error) {
					return &OwnerTagsResult{IsLinked: false, Tags: []string{}}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 no clinic",
			ownerID:    "1",
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400 bad owner ID",
			ownerID:    "abc",
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "500 service error",
			ownerID: "1",
			svc: &mockLstepTagService{
				getOwnerTagsFn: func(_ context.Context, _, _ uint64) (*OwnerTagsResult, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newGetOwnerLstepTagsRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodGet, "/owners/"+tt.ownerID+"/lstep/tags", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestPostOwnerLstepTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validBody, _ := json.Marshal(map[string]string{"tag_name": "my_tag"})
	missingTagBody, _ := json.Marshal(map[string]string{"reason": "test"})
	invalidCharsBody, _ := json.Marshal(map[string]string{"tag_name": "invalid tag!"})

	tests := []struct {
		name       string
		ownerID    string
		body       []byte
		svc        *mockLstepTagService
		wantStatus int
	}{
		{
			name:       "200 success",
			ownerID:    "1",
			body:       validBody,
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 no clinic",
			ownerID:    "1",
			body:       validBody,
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400 bad owner ID",
			ownerID:    "abc",
			body:       validBody,
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 missing tag_name",
			ownerID:    "1",
			body:       missingTagBody,
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 invalid chars in tag_name",
			ownerID:    "1",
			body:       invalidCharsBody,
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "500 service error",
			ownerID: "1",
			body:    validBody,
			svc: &mockLstepTagService{
				addOwnerTagFn: func(_ context.Context, _, _ uint64, _ string, _ *uint64) error {
					return errors.New("lstep error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPostOwnerLstepTagRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodPost, "/owners/"+tt.ownerID+"/lstep/tags", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeleteOwnerLstepTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		ownerID    string
		tagName    string
		svc        *mockLstepTagService
		wantStatus int
	}{
		{
			name:       "204 success",
			ownerID:    "1",
			tagName:    "my_tag",
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "401 no clinic",
			ownerID:    "1",
			tagName:    "my_tag",
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400 bad owner ID",
			ownerID:    "abc",
			tagName:    "my_tag",
			svc:        &mockLstepTagService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "204 idempotent tag not found",
			ownerID: "1",
			tagName: "nonexistent",
			svc: &mockLstepTagService{
				removeOwnerTagFn: func(_ context.Context, _, _ uint64, _ string, _ *uint64) error {
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "500 service error",
			ownerID: "1",
			tagName: "my_tag",
			svc: &mockLstepTagService{
				removeOwnerTagFn: func(_ context.Context, _, _ uint64, _ string, _ *uint64) error {
					return errors.New("lstep error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteOwnerLstepTagRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodDelete, "/owners/"+tt.ownerID+"/lstep/tags/"+tt.tagName, http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
