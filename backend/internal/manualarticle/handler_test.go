package manualarticle

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
)

// ---- mock ManualArticleService ----

type mockManualArticleServiceForHandler struct {
	findAllFn            func(ctx context.Context) ([]PublishedArticle, error)
	findPublishedFn      func(ctx context.Context, category model.ManualCategory, slug string) (*PublishedArticle, error)
	findArticleFn        func(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	findEditViewFn       func(ctx context.Context, category model.ManualCategory, slug string) (*EditView, error)
	deleteFn             func(ctx context.Context, category model.ManualCategory, slug string) error
	findVersionsByIDFn   func(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
}

func (m *mockManualArticleServiceForHandler) FindAll(ctx context.Context) ([]PublishedArticle, error) {
	return m.findAllFn(ctx)
}

func (m *mockManualArticleServiceForHandler) FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*PublishedArticle, error) {
	return m.findPublishedFn(ctx, category, slug)
}

func (m *mockManualArticleServiceForHandler) FindArticleByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
	return m.findArticleFn(ctx, category, slug)
}

func (m *mockManualArticleServiceForHandler) FindEditView(ctx context.Context, category model.ManualCategory, slug string) (*EditView, error) {
	return m.findEditViewFn(ctx, category, slug)
}

func (m *mockManualArticleServiceForHandler) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	return m.deleteFn(ctx, category, slug)
}

func (m *mockManualArticleServiceForHandler) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	return m.findVersionsByIDFn(ctx, articleID)
}

// ---- mock AuditLogger ----

type mockAuditLogger struct {
	lastLogEntry *AuditEntry
	logEntryErr  error
}

func (m *mockAuditLogger) LogEntry(_ context.Context, entry AuditEntry) error {
	m.lastLogEntry = &entry
	return m.logEntryErr
}

// ---- mock EditTxAudit ----

type mockEditTxAudit struct {
	calls int
	err   error
}

func (m *mockEditTxAudit) LogEntryTx(_ context.Context, _ AuditEntry) error {
	m.calls++
	return m.err
}

// ---- helpers ----

func noopPermissionMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// httpapi は clinic_id/user_id を文字列として読む（JWT クレーム由来）。
		c.Set("clinic_id", "1")
		c.Set("user_id", "7")
		c.Set("is_system_admin", false)
		c.Next()
	}
}

func newTestHandler(svc ManualArticleService, deps EditDeps) *Handler {
	return NewHandler(svc, &mockAuditLogger{}, noopPermissionMiddleware, deps)
}

func doRequest(t *testing.T, h *Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---- tests ----

func TestListManualArticles_OK(t *testing.T) {
	svc := &mockManualArticleServiceForHandler{
		findAllFn: func(context.Context) ([]PublishedArticle, error) {
			return []PublishedArticle{{ID: 1, Category: "screens", Slug: "a", Title: "T", Section: "S", BodyMarkdown: "B"}}, nil
		},
	}
	w := doRequest(t, newTestHandler(svc, EditDeps{}), http.MethodGet, "/api/v1/manual/articles", "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"slug":"a"`)
}

func TestGetManualArticle_NotFound(t *testing.T) {
	svc := &mockManualArticleServiceForHandler{
		findPublishedFn: func(context.Context, model.ManualCategory, string) (*PublishedArticle, error) {
			return nil, apperrors.WrapNotFound("manual_article", "screens/x")
		},
	}
	w := doRequest(t, newTestHandler(svc, EditDeps{}), http.MethodGet, "/api/v1/manual/articles/screens/x", "")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetManualArticleEditView_OK(t *testing.T) {
	svc := &mockManualArticleServiceForHandler{
		findEditViewFn: func(context.Context, model.ManualCategory, string) (*EditView, error) {
			return &EditView{Article: model.ManualArticle{ID: 4, Category: "screens", Slug: "a", Version: 2}}, nil
		},
	}
	w := doRequest(t, newTestHandler(svc, EditDeps{}), http.MethodGet, "/api/v1/manual/articles/screens/a/edit", "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"version":2`)
	assert.Contains(t, w.Body.String(), `"current":null`)
}

func TestUpsertManualArticle_RunnerUnconfigured500(t *testing.T) {
	svc := &mockManualArticleServiceForHandler{}
	w := doRequest(t, newTestHandler(svc, EditDeps{}), http.MethodPut, "/api/v1/manual/articles/screens/a", validEditBody)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpsertManualArticle_InvalidBody400(t *testing.T) {
	svc := &mockManualArticleServiceForHandler{}
	deps := EditDeps{
		Runner:  &operation.Runner{},
		Store:   &stubEditTxStore{},
		Grant:   func(context.Context, uint64, uint64) (bool, error) { return true, nil },
		TxAudit: &mockEditTxAudit{},
	}
	w := doRequest(t, newTestHandler(svc, deps), http.MethodPut, "/api/v1/manual/articles/screens/a", `{"bad":1}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpsertManualArticle_RouteArticleMissing404(t *testing.T) {
	svc := &mockManualArticleServiceForHandler{
		findArticleFn: func(context.Context, model.ManualCategory, string) (*model.ManualArticle, error) {
			return nil, apperrors.WrapNotFound("manual_article", "screens/x")
		},
	}
	deps := EditDeps{
		Runner:  &operation.Runner{},
		Store:   &stubEditTxStore{},
		Grant:   func(context.Context, uint64, uint64) (bool, error) { return true, nil },
		TxAudit: &mockEditTxAudit{},
	}
	w := doRequest(t, newTestHandler(svc, deps), http.MethodPut, "/api/v1/manual/articles/screens/x", validEditBody)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpsertManualArticle_TargetMismatch409(t *testing.T) {
	// body.article_id=42 だが経路記事 ID=1 → 認可済み selector 不一致で
	// 業務効果なしの 409 target_mismatch。
	svc := &mockManualArticleServiceForHandler{
		findArticleFn: func(context.Context, model.ManualCategory, string) (*model.ManualArticle, error) {
			return &model.ManualArticle{ID: 1, Category: "screens", Slug: "a"}, nil
		},
	}
	deps := EditDeps{
		Runner:  &operation.Runner{},
		Store:   &stubEditTxStore{},
		Grant:   func(context.Context, uint64, uint64) (bool, error) { return true, nil },
		TxAudit: &mockEditTxAudit{},
	}
	w := doRequest(t, newTestHandler(svc, deps), http.MethodPut, "/api/v1/manual/articles/screens/a", validEditBody)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"target_mismatch"`)
}

func TestGetManualEditOperationResult_RunnerUnconfigured500(t *testing.T) {
	svc := &mockManualArticleServiceForHandler{}
	w := doRequest(t, newTestHandler(svc, EditDeps{}), http.MethodGet,
		"/api/v1/manual/articles/screens/a/operations/11111111-1111-4111-8111-111111111111", "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetManualEditOperationResult_BadKey400(t *testing.T) {
	svc := &mockManualArticleServiceForHandler{}
	deps := EditDeps{
		Runner: &operation.Runner{},
		Store:  &stubEditTxStore{},
		Grant:  func(context.Context, uint64, uint64) (bool, error) { return true, nil },
	}
	w := doRequest(t, newTestHandler(svc, deps), http.MethodGet,
		"/api/v1/manual/articles/screens/a/operations/not-a-uuid", "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// stubEditTxStore は compile 用の最小 EditTxStore 実装（NOT_RUN）。
type stubEditTxStore struct{}

func (s *stubEditTxStore) InsertOperationTarget(context.Context, uint64, uint64, uint64) error {
	return nil
}

func (s *stubEditTxStore) ResolveOperationTarget(context.Context, uint64, uint64) (uint64, bool, error) {
	return 0, false, nil
}

func (s *stubEditTxStore) FindArticleByID(context.Context, uint64) (*model.ManualArticle, error) {
	return nil, nil
}

func (s *stubEditTxStore) FindVersionByID(context.Context, uint64, uint64) (*model.ManualArticleVersion, error) {
	return nil, nil
}

func (s *stubEditTxStore) LockArticleForUpdate(context.Context, uint64) (*model.ManualArticle, error) {
	return nil, apperrors.WrapNotFound("manual_article", "")
}

func (s *stubEditTxStore) NextContentRevision(context.Context, uint64) (int64, error) {
	return 1, nil
}

func (s *stubEditTxStore) InsertVersion(context.Context, *model.ManualArticleVersion) error {
	return nil
}

func (s *stubEditTxStore) UpdateCurrentVersionCAS(context.Context, uint64, int64, uint64, uint64) (bool, error) {
	return false, nil
}

func (s *stubEditTxStore) InsertOperation(context.Context, *model.ManualArticleOperation) error {
	return nil
}

var _ EditTxStore = (*stubEditTxStore)(nil)
