package manualarticle

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

// ---- mock ManualArticleService ----

type mockManualArticleServiceForHandler struct {
	findAllFn               func(ctx context.Context) ([]model.ManualArticle, error)
	findByCategoryAndSlugFn func(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	upsertFn                func(ctx context.Context, input *UpsertManualArticleInput, editorStaffID *uint64) (*model.ManualArticle, error)
	deleteFn                func(ctx context.Context, category model.ManualCategory, slug string) error
	findVersionsByArticleFn func(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
}

func (m *mockManualArticleServiceForHandler) FindAll(ctx context.Context) ([]model.ManualArticle, error) {
	return m.findAllFn(ctx)
}

func (m *mockManualArticleServiceForHandler) FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
	return m.findByCategoryAndSlugFn(ctx, category, slug)
}

func (m *mockManualArticleServiceForHandler) Upsert(ctx context.Context, input *UpsertManualArticleInput, editorStaffID *uint64) (*model.ManualArticle, error) {
	return m.upsertFn(ctx, input, editorStaffID)
}

func (m *mockManualArticleServiceForHandler) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	return m.deleteFn(ctx, category, slug)
}

func (m *mockManualArticleServiceForHandler) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	return m.findVersionsByArticleFn(ctx, articleID)
}

// ---- mock AuditLogger ----
// A small local mock of manualarticle's own AuditLogger interface (2c/BE9-2B design: no
// dependency on internal/service's much larger AuditService mock).

type mockAuditLogger struct {
	lastLogEntry *AuditEntry
	logEntryErr  error
}

func (m *mockAuditLogger) LogEntry(_ context.Context, entry AuditEntry) error {
	m.lastLogEntry = &entry
	return m.logEntryErr
}

// ---- test helpers ----

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

func newHandlerWithManualArticleSvc(svc ManualArticleService) *Handler {
	return NewHandler(svc, &mockAuditLogger{}, nil)
}

func newHandlerWithManualArticleAndAuditSvc(svc ManualArticleService, audit AuditLogger) *Handler {
	return NewHandler(svc, audit, nil)
}

// ---- ListManualArticles ----

func TestListManualArticles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		svc        *mockManualArticleServiceForHandler
		wantStatus int
		wantBody   string
	}{
		{
			name: "returns list of manual articles",
			svc: &mockManualArticleServiceForHandler{
				findAllFn: func(_ context.Context) ([]model.ManualArticle, error) {
					return []model.ManualArticle{{ID: 1, Category: "general", Slug: "intro", Title: "はじめに"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"はじめに"`,
		},
		{
			name: "returns 500 on service error",
			svc: &mockManualArticleServiceForHandler{
				findAllFn: func(_ context.Context) ([]model.ManualArticle, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithManualArticleSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)

			h.ListManualArticles(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetManualArticle ----

func TestGetManualArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		paramCategory string
		paramSlug     string
		svc           *mockManualArticleServiceForHandler
		wantStatus    int
		wantBody      string
	}{
		{
			name:          "returns manual article",
			paramCategory: "general",
			paramSlug:     "intro",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
					assert.Equal(t, model.ManualCategory("general"), category)
					assert.Equal(t, "intro", slug)
					return &model.ManualArticle{ID: 1, Category: category, Slug: slug, Title: "はじめに"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"はじめに"`,
		},
		{
			name:          "returns 400 when slug is empty",
			paramCategory: "general",
			paramSlug:     "",
			svc:           &mockManualArticleServiceForHandler{},
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "returns 404 when article not found",
			paramCategory: "general",
			paramSlug:     "missing",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, _ model.ManualCategory, _ string) (*model.ManualArticle, error) {
					return nil, apperrors.WrapNotFound("manual_article", "missing")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on service error",
			paramCategory: "general",
			paramSlug:     "intro",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, _ model.ManualCategory, _ string) (*model.ManualArticle, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithManualArticleSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{
				{Key: "category", Value: tt.paramCategory},
				{Key: "slug", Value: tt.paramSlug},
			}

			h.GetManualArticle(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpsertManualArticle ----

func TestUpsertManualArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"title":         "はじめに",
			"order_value":   1.0,
			"section":       "基本",
			"body_markdown": "# はじめに",
		}
	}

	tests := []struct {
		name          string
		setupCtx      func(c *gin.Context)
		paramCategory string
		paramSlug     string
		body          any
		svc           *mockManualArticleServiceForHandler
		wantStatus    int
		wantBody      string
	}{
		{
			name:          "upserts manual article successfully",
			setupCtx:      func(c *gin.Context) { setClinicID(c); c.Set("user_id", "5") },
			paramCategory: "general",
			paramSlug:     "intro",
			body:          validBody(),
			svc: &mockManualArticleServiceForHandler{
				upsertFn: func(_ context.Context, input *UpsertManualArticleInput, editorStaffID *uint64) (*model.ManualArticle, error) {
					assert.Equal(t, model.ManualCategory("general"), input.Category)
					assert.Equal(t, "intro", input.Slug)
					assert.Equal(t, "はじめに", input.Title)
					require.NotNil(t, editorStaffID)
					assert.Equal(t, uint64(5), *editorStaffID)
					return &model.ManualArticle{ID: 1, Category: input.Category, Slug: input.Slug, Title: input.Title}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"はじめに"`,
		},
		{
			name:          "returns 401 when clinic_id is missing",
			setupCtx:      func(_ *gin.Context) {},
			paramCategory: "general",
			paramSlug:     "intro",
			body:          validBody(),
			svc:           &mockManualArticleServiceForHandler{},
			wantStatus:    http.StatusUnauthorized,
		},
		{
			name:          "returns 400 when slug is empty",
			setupCtx:      func(c *gin.Context) { setClinicID(c) },
			paramCategory: "general",
			paramSlug:     "",
			body:          validBody(),
			svc:           &mockManualArticleServiceForHandler{},
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "returns 400 when required field missing",
			setupCtx:      func(c *gin.Context) { setClinicID(c) },
			paramCategory: "general",
			paramSlug:     "intro",
			body:          map[string]any{"section": "基本"},
			svc:           &mockManualArticleServiceForHandler{},
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "returns 400 for invalid JSON",
			setupCtx:      func(c *gin.Context) { setClinicID(c) },
			paramCategory: "general",
			paramSlug:     "intro",
			body:          "not-json",
			svc:           &mockManualArticleServiceForHandler{},
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "returns 500 on service error",
			setupCtx:      func(c *gin.Context) { setClinicID(c); c.Set("user_id", "5") },
			paramCategory: "general",
			paramSlug:     "intro",
			body:          validBody(),
			svc: &mockManualArticleServiceForHandler{
				upsertFn: func(_ context.Context, _ *UpsertManualArticleInput, _ *uint64) (*model.ManualArticle, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithManualArticleSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{
				{Key: "category", Value: tt.paramCategory},
				{Key: "slug", Value: tt.paramSlug},
			}
			tt.setupCtx(c)

			h.UpsertManualArticle(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// TestUpsertManualArticle_AuditLogged verifies the audit log entry is written with expected fields
// and that the handler still succeeds even if the audit write itself fails (best-effort).
func TestUpsertManualArticle_AuditLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockManualArticleServiceForHandler{
		upsertFn: func(_ context.Context, input *UpsertManualArticleInput, _ *uint64) (*model.ManualArticle, error) {
			return &model.ManualArticle{ID: 42, Category: input.Category, Slug: input.Slug, Title: input.Title}, nil
		},
	}
	audit := &mockAuditLogger{logEntryErr: fmt.Errorf("audit write failed")}
	h := newHandlerWithManualArticleAndAuditSvc(svc, audit)

	body := map[string]any{
		"title":         "はじめに",
		"order_value":   1.0,
		"section":       "基本",
		"body_markdown": "# はじめに",
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "category", Value: "general"},
		{Key: "slug", Value: "intro"},
	}
	setClinicID(c)
	c.Set("user_id", "5")

	h.UpsertManualArticle(c)

	// Audit failure is best-effort; handler still returns 200.
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, audit.lastLogEntry)
	assert.Equal(t, model.AuditActionManualArticleUpsert, audit.lastLogEntry.Action)
	assert.Equal(t, "manual_article", audit.lastLogEntry.Resource)
	require.NotNil(t, audit.lastLogEntry.ResourceID)
	assert.Equal(t, uint64(42), *audit.lastLogEntry.ResourceID)
}

// TestDeleteManualArticle_AuditLogged verifies the audit log entry is written with expected fields.
// TRM-02: audit failure is fail-closed (delete already applied, but 204 must not be returned).
func TestDeleteManualArticle_AuditLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockManualArticleServiceForHandler{
		findByCategoryAndSlugFn: func(_ context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
			return &model.ManualArticle{ID: 42, Category: category, Slug: slug}, nil
		},
		deleteFn: func(_ context.Context, _ model.ManualCategory, _ string) error {
			return nil
		},
	}
	audit := &mockAuditLogger{logEntryErr: fmt.Errorf("audit write failed")}
	h := newHandlerWithManualArticleAndAuditSvc(svc, audit)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
	c.Params = gin.Params{
		{Key: "category", Value: "general"},
		{Key: "slug", Value: "intro"},
	}
	setClinicID(c)
	c.Set("user_id", "5")

	h.DeleteManualArticle(c)

	assert.NotEqual(t, http.StatusNoContent, w.Code, "audit failure must not return 204")
	require.NotNil(t, audit.lastLogEntry)
	assert.Equal(t, model.AuditActionManualArticleDelete, audit.lastLogEntry.Action)
	assert.Equal(t, "manual_article", audit.lastLogEntry.Resource)
	require.NotNil(t, audit.lastLogEntry.ResourceID)
	assert.Equal(t, uint64(42), *audit.lastLogEntry.ResourceID)
}

// ---- DeleteManualArticle ----

// newDeleteManualArticleRouter builds a router with clinic_id/user_id injected via middleware.
// A full router is required so gin flushes the response header for the 204 No Content success path.
func newDeleteManualArticleRouter(svc ManualArticleService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithManualArticleSvc(svc)
	r.DELETE("/manual/articles/:category/:slug", func(c *gin.Context) {
		setClinicID(c)
		c.Set("user_id", "5")
	}, h.DeleteManualArticle)
	return r
}

func TestDeleteManualArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		paramCategory string
		paramSlug     string
		svc           *mockManualArticleServiceForHandler
		wantStatus    int
	}{
		{
			name:          "deletes manual article successfully",
			paramCategory: "general",
			paramSlug:     "intro",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
					return &model.ManualArticle{ID: 1, Category: category, Slug: slug}, nil
				},
				deleteFn: func(_ context.Context, category model.ManualCategory, slug string) error {
					assert.Equal(t, model.ManualCategory("general"), category)
					assert.Equal(t, "intro", slug)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when article to delete not found",
			paramCategory: "general",
			paramSlug:     "missing",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, _ model.ManualCategory, _ string) (*model.ManualArticle, error) {
					return nil, apperrors.WrapNotFound("manual_article", "missing")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 when delete fails",
			paramCategory: "general",
			paramSlug:     "intro",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
					return &model.ManualArticle{ID: 1, Category: category, Slug: slug}, nil
				},
				deleteFn: func(_ context.Context, _ model.ManualCategory, _ string) error {
					return fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteManualArticleRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/manual/articles/"+tt.paramCategory+"/"+tt.paramSlug, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithManualArticleSvc(&mockManualArticleServiceForHandler{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{
			{Key: "category", Value: "general"},
			{Key: "slug", Value: "intro"},
		}
		h.DeleteManualArticle(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 when slug is empty", func(t *testing.T) {
		h := newHandlerWithManualArticleSvc(&mockManualArticleServiceForHandler{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{
			{Key: "category", Value: "general"},
			{Key: "slug", Value: ""},
		}
		setClinicID(c)
		h.DeleteManualArticle(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ---- ListManualArticleVersions ----

func TestListManualArticleVersions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		paramCategory string
		paramSlug     string
		svc           *mockManualArticleServiceForHandler
		wantStatus    int
		wantBody      string
	}{
		{
			name:          "returns list of manual article versions",
			paramCategory: "general",
			paramSlug:     "intro",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
					return &model.ManualArticle{ID: 7, Category: category, Slug: slug}, nil
				},
				findVersionsByArticleFn: func(_ context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
					assert.Equal(t, uint64(7), articleID)
					return []model.ManualArticleVersion{{ID: 100, ArticleID: articleID, Title: "旧版"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"旧版"`,
		},
		{
			name:          "returns 400 when slug is empty",
			paramCategory: "general",
			paramSlug:     "",
			svc:           &mockManualArticleServiceForHandler{},
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "returns 404 when article not found",
			paramCategory: "general",
			paramSlug:     "missing",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, _ model.ManualCategory, _ string) (*model.ManualArticle, error) {
					return nil, apperrors.WrapNotFound("manual_article", "missing")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 when version lookup fails",
			paramCategory: "general",
			paramSlug:     "intro",
			svc: &mockManualArticleServiceForHandler{
				findByCategoryAndSlugFn: func(_ context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
					return &model.ManualArticle{ID: 7, Category: category, Slug: slug}, nil
				},
				findVersionsByArticleFn: func(_ context.Context, _ uint64) ([]model.ManualArticleVersion, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithManualArticleSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{
				{Key: "category", Value: tt.paramCategory},
				{Key: "slug", Value: tt.paramSlug},
			}

			h.ListManualArticleVersions(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
