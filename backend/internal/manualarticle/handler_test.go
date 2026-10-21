package manualarticle

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
	"github.com/animal-ekarte/backend/internal/testdb"
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

func TestPublicRead_NoEditPermissionRequired(t *testing.T) {
	// 公開 READ は認証のみ（FR-internal-manual-publication）。呼び出されれば
	// 403 になる permission middleware を差し込んでも GET list / GET
	// {category}/{slug} は到達する（= route に middleware が登録されない証明）。
	denyAll := func(resource, action string) gin.HandlerFunc {
		return func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) }
	}
	svc := &mockManualArticleServiceForHandler{
		findAllFn: func(context.Context) ([]PublishedArticle, error) {
			return []PublishedArticle{{ID: 1, Category: "screens", Slug: "a", Title: "T", Section: "S", BodyMarkdown: "B"}}, nil
		},
		findPublishedFn: func(context.Context, model.ManualCategory, string) (*PublishedArticle, error) {
			return &PublishedArticle{ID: 1, Category: "screens", Slug: "a", Title: "T", Section: "S", BodyMarkdown: "B"}, nil
		},
	}
	h := NewHandler(svc, &mockAuditLogger{}, denyAll, EditDeps{})
	w := doRequest(t, h, http.MethodGet, "/api/v1/manual/articles", "")
	assert.Equal(t, http.StatusOK, w.Code)
	w = doRequest(t, h, http.MethodGet, "/api/v1/manual/articles/screens/a", "")
	assert.Equal(t, http.StatusOK, w.Code)
	// 対比: 編集系 GET は同じ denyAll で 403（permission 登録が残る証明）。
	w = doRequest(t, h, http.MethodGet, "/api/v1/manual/articles/screens/a/edit", "")
	assert.Equal(t, http.StatusForbidden, w.Code)
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

func TestUpsertManualArticle_TargetMismatchWithoutGrant403(t *testing.T) {
	// 不一致でも現在 grant が失効していれば 403。未認可の caller に
	// target_mismatch の詳細を開示しない（採択契約 §3）。
	svc := &mockManualArticleServiceForHandler{
		findArticleFn: func(context.Context, model.ManualCategory, string) (*model.ManualArticle, error) {
			return &model.ManualArticle{ID: 1, Category: "screens", Slug: "a"}, nil
		},
	}
	deps := EditDeps{
		Runner:  &operation.Runner{},
		Store:   &stubEditTxStore{},
		Grant:   func(context.Context, uint64, uint64) (bool, error) { return false, nil },
		TxAudit: &mockEditTxAudit{},
	}
	w := doRequest(t, newTestHandler(svc, deps), http.MethodPut, "/api/v1/manual/articles/screens/a", validEditBody)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NotContains(t, w.Body.String(), "target_mismatch")
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

// ---- lookup HTTP 写像（real store + real Runner, NOT_RUN） ----
//
// fix-9a: 記録済み unknown だけが 200 unknown。in-progress / not-succeeded /
// scope-not-found / Lookup error は generic 500、認可失効は 403、
// 認可後の selector 不一致は 409 target_mismatch。
// authorized の結果 store 再読取失敗→500 は operation.TestLookup_ResultStoreReadFailureAuthorizedIsError
// と本 handler の `httpapi.RespondError(err)`（Lookup error → 500）で担保。
// 共有 testdb schema の表を mutate する注入はここには置かない。

func setupLookupHandlerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.OperationBinding{},
		&model.OperationResult{},
		&model.OperationAttemptOutcome{},
	))
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS operation_bindings_scope_key_test
		ON operation_bindings (clinic_id, owner_kind, actor_kind, actor_ref, intent, operation_key)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS manual_article_operation_targets (
		clinic_id bigint NOT NULL, binding_id bigint NOT NULL, article_id bigint NOT NULL,
		PRIMARY KEY (clinic_id, binding_id))`).Error)
	require.NoError(t, db.Exec(`TRUNCATE TABLE manual_article_operation_targets,
		operation_attempt_outcomes, operation_results, operation_bindings`).Error)
	testdb.SeedClinicsForFK(t, db, 1)
	return db
}

func lookupTestRunner(t *testing.T, db *gorm.DB) *operation.Runner {
	t.Helper()
	fp, err := operation.NewFingerprinter("k1", map[string][]byte{"k1": make([]byte, 32)})
	require.NoError(t, err)
	return operation.NewRunner(db, fp)
}

func seedLookupBinding(t *testing.T, db *gorm.DB, key uuid.UUID) model.OperationBinding {
	t.Helper()
	b := model.OperationBinding{
		ClinicID:              1,
		OwnerKind:             OwnerKind,
		ActorKind:             string(operation.ActorStaff),
		ActorRef:              "7",
		Intent:                IntentEdit,
		OperationKey:          key,
		Fingerprint:           make([]byte, 32),
		FingerprintKeyVersion: "k1",
		CreatedAt:             time.Now(),
	}
	require.NoError(t, db.Create(&b).Error)
	return b
}

func seedLookupAnchor(t *testing.T, db *gorm.DB, bindingID, articleID uint64) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO manual_article_operation_targets (clinic_id, binding_id, article_id) VALUES (1, ?, ?)`,
		bindingID, articleID).Error)
}

func seedLookupArticle(t *testing.T, db *gorm.DB, slug string) {
	t.Helper()
	require.NoError(t, db.Create(&model.ManualArticle{
		ID: 42, Category: model.ManualCategoryScreens, Slug: slug, Version: 1,
	}).Error)
}

func lookupHandler(t *testing.T, db *gorm.DB, grant EditGrantChecker) *Handler {
	t.Helper()
	return newTestHandler(&mockManualArticleServiceForHandler{}, EditDeps{
		Runner:  lookupTestRunner(t, db),
		Store:   New(db),
		Grant:   grant,
		TxAudit: &mockEditTxAudit{},
	})
}

func TestGetManualEditOperationResult_RecordedUnknown200(t *testing.T) {
	db := setupLookupHandlerDB(t)
	key := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	b := seedLookupBinding(t, db, key)
	seedLookupAnchor(t, db, b.ID, 42)
	seedLookupArticle(t, db, "intro")
	require.NoError(t, db.Create(&model.OperationAttemptOutcome{
		ClinicID: 1, BindingID: b.ID, AttemptKey: uuid.New(),
		Outcome: string(operation.OutcomeUnknown), Details: json.RawMessage(`{"status":"result-unavailable"}`),
		RecordedAt: time.Now(),
	}).Error)

	h := lookupHandler(t, db, func(context.Context, uint64, uint64) (bool, error) { return true, nil })
	w := doRequest(t, h, http.MethodGet,
		"/api/v1/manual/articles/screens/intro/operations/"+key.String(), "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"state":"unknown"`)
}

func TestGetManualEditOperationResult_ScopeNotFound500(t *testing.T) {
	db := setupLookupHandlerDB(t)
	key := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	h := lookupHandler(t, db, func(context.Context, uint64, uint64) (bool, error) { return true, nil })
	w := doRequest(t, h, http.MethodGet,
		"/api/v1/manual/articles/screens/intro/operations/"+key.String(), "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetManualEditOperationResult_NotSucceeded500(t *testing.T) {
	db := setupLookupHandlerDB(t)
	key := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	b := seedLookupBinding(t, db, key)
	seedLookupAnchor(t, db, b.ID, 42)
	seedLookupArticle(t, db, "intro")
	h := lookupHandler(t, db, func(context.Context, uint64, uint64) (bool, error) { return true, nil })
	w := doRequest(t, h, http.MethodGet,
		"/api/v1/manual/articles/screens/intro/operations/"+key.String(), "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetManualEditOperationResult_InProgress500(t *testing.T) {
	db := setupLookupHandlerDB(t)
	key := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	b := seedLookupBinding(t, db, key)
	seedLookupAnchor(t, db, b.ID, 42)
	seedLookupArticle(t, db, "intro")

	holder := db.Begin()
	require.NoError(t, holder.Error)
	t.Cleanup(func() { holder.Rollback() })
	require.NoError(t, holder.Exec(
		`SELECT * FROM operation_bindings WHERE id = ? FOR UPDATE`, b.ID).Error)

	h := lookupHandler(t, db, func(context.Context, uint64, uint64) (bool, error) { return true, nil })
	w := doRequest(t, h, http.MethodGet,
		"/api/v1/manual/articles/screens/intro/operations/"+key.String(), "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetManualEditOperationResult_NotAuthorized403(t *testing.T) {
	db := setupLookupHandlerDB(t)
	key := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	b := seedLookupBinding(t, db, key)
	seedLookupAnchor(t, db, b.ID, 42)
	seedLookupArticle(t, db, "intro")
	h := lookupHandler(t, db, func(context.Context, uint64, uint64) (bool, error) { return false, nil })
	w := doRequest(t, h, http.MethodGet,
		"/api/v1/manual/articles/screens/intro/operations/"+key.String(), "")
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetManualEditOperationResult_SelectorMismatch409(t *testing.T) {
	db := setupLookupHandlerDB(t)
	key := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	b := seedLookupBinding(t, db, key)
	// anchor が指す記事の slug は "other" — 経路 "intro" と不一致。
	seedLookupAnchor(t, db, b.ID, 42)
	seedLookupArticle(t, db, "other")
	h := lookupHandler(t, db, func(context.Context, uint64, uint64) (bool, error) { return true, nil })
	w := doRequest(t, h, http.MethodGet,
		"/api/v1/manual/articles/screens/intro/operations/"+key.String(), "")
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"target_mismatch"`)
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
