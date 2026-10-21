package manualarticle

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
)

// PermissionMiddleware builds the gin.HandlerFunc that gates a route on (resource, action).
// This is manualarticle's consumer-side view of the permission-checking middleware — BE9-2A
// classifies permission_middleware.go as target:auth, and manualarticle (topologically before
// auth in ADR-006's permitted dependency graph: "httpapi → clinic → inventory →
// manualarticle → owner → pet → staff → auth → ...") must not depend on the not-yet-migrated
// auth domain package. The composition root (cmd/api/main.go) supplies the concrete
// implementation (today, the transitional *handler.Handler.RequirePermission method value;
// once BE9-2C/2D migrates auth, an *auth.Handler method value instead) — manualarticle never
// imports internal/handler or internal/auth.
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// EditDeps は既存記事 edit の C06 接続部品。Runner が nil のとき edit 系
// endpoint は 500 を返す（fingerprinter keyring 未設定は fail-closed）。
type EditDeps struct {
	Runner  *operation.Runner
	Store   EditTxStore
	Grant   EditGrantChecker
	TxAudit EditTxAudit
}

// Handler serves the manualarticle HTTP boundary. Unlike the legacy *handler.Handler, it takes
// only the narrow dependencies it actually needs (service, audit, permission middleware,
// edit operation deps) — it never holds *service.Services / *repository.Repositories /
// *handler.Handler.
type Handler struct {
	service           ManualArticleService
	audit             AuditLogger
	requirePermission PermissionMiddleware
	edit              EditDeps
}

// NewHandler initializes a Handler.
func NewHandler(service ManualArticleService, audit AuditLogger, requirePermission PermissionMiddleware, edit EditDeps) *Handler {
	return &Handler{service: service, audit: audit, requirePermission: requirePermission, edit: edit}
}

// ListManualArticles は公開確定版のマニュアル記事一覧を返す（認証済みユーザー全員）
//
// GET /api/v1/manual/articles
// 応答形式は既存 ManualArticleResponse を維持し、内容は published_version_id の版だけ。
// 未公開・hidden は含めない。
func (h *Handler) ListManualArticles(c *gin.Context) {
	articles, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toManualArticleListResponse(articles))
}

// GetManualArticle は単一マニュアル記事の公開版を返す
//
// GET /api/v1/manual/articles/:category/:slug
// 未公開・hidden・不存在は 404。
func (h *Handler) GetManualArticle(c *gin.Context) {
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}
	article, err := h.service.FindByCategoryAndSlug(c.Request.Context(), category, slug)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toManualArticleResponse(article))
}

// GetManualArticleEditView は既存記事の編集元と CAS 版を返す
//
// GET /api/v1/manual/articles/:category/:slug/edit
// Requires: ResourceManualEdit, edit permission
// current は現在の編集内容版（NULL なら null）。出所未解決は legacy_unknown の
// まま示し、架空の digest/revision で補完しない。
func (h *Handler) GetManualArticleEditView(c *gin.Context) {
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}
	view, err := h.service.FindEditView(c.Request.Context(), category, slug)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toManualArticleEditViewResponse(view))
}

// UpsertManualArticle は既存マニュアル記事の編集保存を行う
//
// PUT /api/v1/manual/articles/:category/:slug
// Requires: ResourceManualEdit, edit permission
// C06: claim → fingerprint 一致 → participant Apply（同一 owner tx で
// 内容版・current・履歴・監査・receipt）。応答は確定済み receipt のみ。
func (h *Handler) UpsertManualArticle(c *gin.Context) {
	ctx := c.Request.Context()
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		httpapi.RespondError(c, apperrors.WrapForbidden("staff identity required for manual article edit"))
		return
	}
	if h.edit.Runner == nil || h.edit.Store == nil || h.edit.Grant == nil || h.edit.TxAudit == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("manual article edit is not configured"))
		return
	}
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("request body unreadable"))
		return
	}
	systemAdmin, _ := httpapi.PeekIsSystemAdmin(c)
	// 単一 factory: 一度の検証で canonical Content/Preconditions と
	// immutable command の participant を組み立てる（別 payload 禁止）。
	exec, err := newManualEditExecution(
		raw, category, slug,
		h.edit.Store, h.edit.Grant, h.edit.TxAudit,
		clinicID, staffID, systemAdmin,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	// 経路 selector が指す記事を解決する。不存在は 404（新規 upsert はしない）。
	article, err := h.service.FindArticleByCategoryAndSlug(ctx, category, slug)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	// body.article_id と経路記事の一致を先に検査する。違えば 409 target_mismatch。
	if exec.ArticleID != article.ID {
		c.JSON(http.StatusConflict, ManualEditConflictBody{Error: "manual edit target mismatch", Code: "target_mismatch"})
		return
	}
	outcome, err := h.edit.Runner.Execute(ctx, exec.Request, exec.Participant)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	h.respondEditOutcome(c, outcome, exec.Participant)
}

// respondEditOutcome は PUT の Outcome を HTTP 応答へ写像する。
//
//	succeeded/replayed        → 200 ManualEditReceipt（永続化 receipt のみ）
//	operation-key-conflict    → 409 {code: operation_key_reused}
//	version-conflict          → 409 {code: version_conflict}
//	validation（participant code）→ 409 {code: target_mismatch|source_unresolved}
//	not-authorized/edit-right-lost → 403
//	その他                    → 500 generic ErrorResponse
func (h *Handler) respondEditOutcome(c *gin.Context, outcome operation.Outcome, participant *EditParticipant) {
	switch outcome.Status {
	case operation.StatusSucceeded, operation.StatusReplayed:
		if outcome.Result == nil || outcome.Binding == nil {
			httpapi.RespondError(c, apperrors.WrapInternalServerError("manual edit result unavailable"))
			return
		}
		receipt, ok := receiptFromStoredResult(outcome.Binding.Key, outcome.Result)
		if !ok {
			httpapi.RespondError(c, apperrors.WrapInternalServerError("manual edit result unavailable"))
			return
		}
		c.JSON(http.StatusOK, receipt)
	case operation.StatusKeyConflict:
		c.JSON(http.StatusConflict, ManualEditConflictBody{Error: "operation key reused for different content", Code: "operation_key_reused"})
	case operation.StatusVersionConflict:
		c.JSON(http.StatusConflict, ManualEditConflictBody{Error: "manual article version conflict", Code: "version_conflict"})
	case operation.StatusValidation:
		// 拒否理由の区別は request-scope participant に保持される。
		// validation は target_mismatch または source_unresolved のみ発行する。
		code := participant.ConflictCode()
		if code != "target_mismatch" && code != "source_unresolved" {
			httpapi.RespondError(c, apperrors.WrapInternalServerError("manual article edit failed"))
			return
		}
		c.JSON(http.StatusConflict, ManualEditConflictBody{Error: "manual edit conflict: " + code, Code: code})
	case operation.StatusNotAuthorized, operation.StatusEditRightLost:
		httpapi.RespondError(c, apperrors.WrapForbidden("manual article edit not authorized"))
	default:
		httpapi.RespondError(c, apperrors.WrapInternalServerError("manual article edit failed"))
	}
}

// GetManualEditOperationResult は編集操作の結果照会を返す
//
// GET /api/v1/manual/articles/:category/:slug/operations/:operation_key
// Requires: ResourceManualEdit, edit permission
// 200 succeeded は元の確定 receipt、200 unknown は照合不能。
// anchor 認可後にだけ route selector を比較し、不一致は 409 target_mismatch。
func (h *Handler) GetManualEditOperationResult(c *gin.Context) {
	ctx := c.Request.Context()
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		httpapi.RespondError(c, apperrors.WrapForbidden("staff identity required"))
		return
	}
	if h.edit.Runner == nil || h.edit.Store == nil || h.edit.Grant == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("manual article edit is not configured"))
		return
	}
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}
	key, err := operation.ParseOperationKey(c.Param("operation_key"))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	scope := operation.Scope{
		ClinicID:  clinicID,
		OwnerKind: OwnerKind,
		ActorKind: operation.ActorStaff,
		ActorRef:  strconv.FormatUint(staffID, 10),
		Intent:    IntentEdit,
	}
	systemAdmin, _ := httpapi.PeekIsSystemAdmin(c)
	participant := NewEditParticipant(
		h.edit.Store, h.edit.Grant, h.edit.TxAudit,
		clinicID, staffID, systemAdmin, category, slug, nil,
	)
	outcome, err := h.edit.Runner.Lookup(ctx, scope, key, participant)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if outcome.Status == operation.StatusNotAuthorized {
		httpapi.RespondError(c, apperrors.WrapForbidden("manual edit operation not authorized"))
		return
	}
	// anchor 認可の後でだけ route selector と解決済み記事を比較する。
	if resolvedCategory, resolvedSlug, ok := participant.ResolvedSelectors(); ok {
		if resolvedCategory != category || resolvedSlug != slug {
			c.JSON(http.StatusConflict, ManualEditConflictBody{Error: "manual edit target mismatch", Code: "target_mismatch"})
			return
		}
	}
	switch outcome.Status {
	case operation.StatusSucceeded, operation.StatusReplayed:
		if outcome.Result == nil {
			httpapi.RespondError(c, apperrors.WrapInternalServerError("manual edit result unavailable"))
			return
		}
		receipt, ok := receiptFromStoredResult(key, outcome.Result)
		if !ok {
			httpapi.RespondError(c, apperrors.WrapInternalServerError("manual edit result unavailable"))
			return
		}
		c.JSON(http.StatusOK, ManualEditOperationResultResponse{
			OperationKey: key.String(),
			State:        "succeeded",
			Receipt:      receipt,
		})
	case operation.StatusResultUnavailable:
		// 記録済み unknown（照合不能）だけを 200 unknown で返す。
		// 入力は保持し、新 key で救済しない。
		c.JSON(http.StatusOK, ManualEditOperationResultResponse{
			OperationKey: key.String(),
			State:        "unknown",
			Receipt:      nil,
		})
	default:
		// in-progress / not-succeeded / scope-not-found 等の未選定結果は
		// 結果不明として返さず generic 500 ErrorResponse に写像する。
		httpapi.RespondError(c, apperrors.WrapInternalServerError("manual edit result unavailable"))
	}
}

// DeleteManualArticle はマニュアルオーバーライドを削除する（MD ファイル版に戻す）
//
// DELETE /api/v1/manual/articles/:category/:slug
// Requires: ResourceManualEdit, delete permission
func (h *Handler) DeleteManualArticle(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}

	// TRM-02: destructive delete requires an authenticated staff actor and fail-closed audit.
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	// 削除前に対象を取得しておく（監査ログ用）
	target, findErr := h.service.FindArticleByCategoryAndSlug(c.Request.Context(), category, slug)
	if findErr != nil {
		httpapi.RespondError(c, findErr)
		return
	}

	if err := h.service.Delete(c.Request.Context(), category, slug); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// Fail-closed: do not report 204 when recovery snapshot audit cannot be written.
	if err := h.audit.LogEntry(c.Request.Context(), AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &staffID,
		ActorType:  "staff",
		Action:     model.AuditActionManualArticleDelete,
		Resource:   "manual_article",
		ResourceID: &target.ID,
		OldValue:   target,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.Header.Get("User-Agent"),
	}); err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to write audit log for manual article delete",
			"error", err, "resource_id", target.ID, "actor_id", staffID)
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to write audit log for manual article delete"))
		return
	}

	c.Status(http.StatusNoContent)
}

// RegisterRoutes はマニュアル記事ルートを登録する。
//
// マニュアルは医院共通のため clinic_id を持たない。
// GET は認証済みユーザー全員、編集系は ResourceManualEdit 権限が必要。
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	manual := rg.Group("/manual/articles")
	// 閲覧系: 認証済みユーザー全員（認証のみ・権限不要）。公開 READ は編集権限と
	// 分離する（FR-internal-manual-publication / docs/design/domains/internal-manual/technical-spec.md:31）。
	// これらの handler は published_version_id が示す公開版だけを返し、
	// hidden・未公開内容は返さないため未公開情報の開示はない。
	manual.GET("", h.ListManualArticles)
	manual.GET("/:category/:slug", h.GetManualArticle)
	manual.GET("/:category/:slug/versions", h.requirePermission(string(model.ResourceManualEdit), "edit"), h.ListManualArticleVersions)

	// 編集系: ResourceManualEdit 権限必須
	manual.GET("/:category/:slug/edit", h.requirePermission(string(model.ResourceManualEdit), "edit"), h.GetManualArticleEditView)
	manual.GET("/:category/:slug/operations/:operation_key", h.requirePermission(string(model.ResourceManualEdit), "edit"), h.GetManualEditOperationResult)
	manual.PUT("/:category/:slug", h.requirePermission(string(model.ResourceManualEdit), "edit"), h.UpsertManualArticle)
	manual.DELETE("/:category/:slug", h.requirePermission(string(model.ResourceManualEdit), "delete"), h.DeleteManualArticle)
}

// ListManualArticleVersions は指定記事の編集履歴を返す
//
// GET /api/v1/manual/articles/:category/:slug/versions
// Requires: ResourceManualEdit, edit permission（非公開内容を含むため edit 権限）。
func (h *Handler) ListManualArticleVersions(c *gin.Context) {
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}
	// まず article を特定して ID を取得
	article, err := h.service.FindArticleByCategoryAndSlug(c.Request.Context(), category, slug)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	versions, err := h.service.FindVersionsByArticleID(c.Request.Context(), article.ID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toManualArticleVersionListResponse(versions))
}
