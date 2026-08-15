package manualarticle

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
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

// Handler serves the manualarticle HTTP boundary. Unlike the legacy *handler.Handler, it takes
// only the narrow dependencies it actually needs (service, audit, permission middleware) —
// it never holds *service.Services / *repository.Repositories / *handler.Handler.
type Handler struct {
	service           ManualArticleService
	audit             AuditLogger
	requirePermission PermissionMiddleware
}

// NewHandler initializes a Handler.
func NewHandler(service ManualArticleService, audit AuditLogger, requirePermission PermissionMiddleware) *Handler {
	return &Handler{service: service, audit: audit, requirePermission: requirePermission}
}

// ListManualArticles は全マニュアル記事を返す（認証済みユーザー全員）
//
// GET /api/v1/manual/articles
func (h *Handler) ListManualArticles(c *gin.Context) {
	articles, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toManualArticleListResponse(articles))
}

// GetManualArticle は単一マニュアル記事を返す
//
// GET /api/v1/manual/articles/:category/:slug
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

// UpsertManualArticle はマニュアルの作成・更新を行う
//
// PUT /api/v1/manual/articles/:category/:slug
// Requires: ResourceManualEdit, edit permission
func (h *Handler) UpsertManualArticle(c *gin.Context) {
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

	var req UpsertManualArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	var editorStaffID *uint64
	if id, ok := httpapi.ExtractStaffID(c); ok {
		editorStaffID = &id
	}

	saved, err := h.service.Upsert(c.Request.Context(), req.toServiceInput(category, slug), editorStaffID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// 監査ログ: マニュアル編集（ベストエフォート、失敗はログ記録）
	if staffID, ok := httpapi.ExtractStaffID(c); ok {
		if err := h.audit.LogEntry(c.Request.Context(), AuditEntry{
			ClinicID:   &clinicID,
			ActorID:    &staffID,
			ActorType:  "staff",
			Action:     model.AuditActionManualArticleUpsert,
			Resource:   "manual_article",
			ResourceID: &saved.ID,
			NewValue:   saved,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.Header.Get("User-Agent"),
		}); err != nil {
			slog.WarnContext(c.Request.Context(), "failed to write audit log for manual article upsert",
				"error", err, "resource_id", saved.ID, "actor_id", staffID)
		}
	}

	c.JSON(http.StatusOK, toManualArticleResponse(saved))
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
	target, findErr := h.service.FindByCategoryAndSlug(c.Request.Context(), category, slug)
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
	// 閲覧系: 認証済みユーザー全員（マニュアル閲覧は全スタッフ向け）
	// SEC-602: P5 RequirePermission(view) 付与
	manual.GET("", h.requirePermission(string(model.ResourceManualEdit), "view"), h.ListManualArticles)
	manual.GET("/:category/:slug", h.requirePermission(string(model.ResourceManualEdit), "view"), h.GetManualArticle)
	manual.GET("/:category/:slug/versions", h.requirePermission(string(model.ResourceManualEdit), "view"), h.ListManualArticleVersions)

	// 編集系: ResourceManualEdit 権限必須
	manual.PUT("/:category/:slug", h.requirePermission(string(model.ResourceManualEdit), "edit"), h.UpsertManualArticle)
	manual.DELETE("/:category/:slug", h.requirePermission(string(model.ResourceManualEdit), "delete"), h.DeleteManualArticle)
}

// ListManualArticleVersions は指定記事の編集履歴を返す
//
// GET /api/v1/manual/articles/:category/:slug/versions
func (h *Handler) ListManualArticleVersions(c *gin.Context) {
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}
	// まず article を特定して ID を取得
	article, err := h.service.FindByCategoryAndSlug(c.Request.Context(), category, slug)
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
