package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListManualArticles は全マニュアル記事を返す（認証済みユーザー全員）
//
// GET /api/v1/manual/articles
func (h *Handler) ListManualArticles(c *gin.Context) {
	articles, err := h.svc.ManualArticle.FindAll(c.Request.Context())
	if err != nil {
		RespondError(c, err)
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
		RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}
	article, err := h.svc.ManualArticle.FindByCategoryAndSlug(c.Request.Context(), category, slug)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toManualArticleResponse(article))
}

// UpsertManualArticle はマニュアルの作成・更新を行う
//
// PUT /api/v1/manual/articles/:category/:slug
// Requires: ResourceManualEdit, edit permission
func (h *Handler) UpsertManualArticle(c *gin.Context) {
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}

	var req UpsertManualArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	var editorStaffID *uint64
	if id, ok := extractStaffID(c); ok {
		editorStaffID = &id
	}

	saved, err := h.svc.ManualArticle.Upsert(c.Request.Context(), req.toServiceInput(category, slug), editorStaffID)
	if err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: マニュアル編集（ベストエフォート、失敗はログ記録）
	if staffID, ok := extractStaffID(c); ok {
		if err := h.svc.Audit.LogEntry(c.Request.Context(), &service.AuditLogInput{
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
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}

	// 削除前に対象を取得しておく（監査ログ用）
	target, findErr := h.svc.ManualArticle.FindByCategoryAndSlug(c.Request.Context(), category, slug)
	if findErr != nil {
		RespondError(c, findErr)
		return
	}

	if err := h.svc.ManualArticle.Delete(c.Request.Context(), category, slug); err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: マニュアル削除（ベストエフォート、失敗はログ記録）
	if staffID, ok := extractStaffID(c); ok {
		if err := h.svc.Audit.LogEntry(c.Request.Context(), &service.AuditLogInput{
			ActorID:    &staffID,
			ActorType:  "staff",
			Action:     model.AuditActionManualArticleDelete,
			Resource:   "manual_article",
			ResourceID: &target.ID,
			OldValue:   target,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.Header.Get("User-Agent"),
		}); err != nil {
			slog.WarnContext(c.Request.Context(), "failed to write audit log for manual article delete",
				"error", err, "resource_id", target.ID, "actor_id", staffID)
		}
	}

	c.Status(http.StatusNoContent)
}

// registerManualArticleRoutes はマニュアル記事ルートを登録する。
//
// マニュアルは医院共通のため clinic_id を持たない。
// GET は認証済みユーザー全員、編集系は ResourceManualEdit 権限が必要。
func (h *Handler) registerManualArticleRoutes(rg *gin.RouterGroup) {
	manual := rg.Group("/manual/articles")
	// 閲覧系: 認証済みユーザー全員（マニュアル閲覧は全スタッフ向け）
	// SEC-602: P5 RequirePermission(view) 付与
	manual.GET("", h.RequirePermission(string(model.ResourceManualEdit), "view"), h.ListManualArticles)
	manual.GET("/:category/:slug", h.RequirePermission(string(model.ResourceManualEdit), "view"), h.GetManualArticle)
	manual.GET("/:category/:slug/versions", h.RequirePermission(string(model.ResourceManualEdit), "view"), h.ListManualArticleVersions)

	// 編集系: ResourceManualEdit 権限必須
	manual.PUT("/:category/:slug", h.RequirePermission(string(model.ResourceManualEdit), "edit"), h.UpsertManualArticle)
	manual.DELETE("/:category/:slug", h.RequirePermission(string(model.ResourceManualEdit), "delete"), h.DeleteManualArticle)
}

// ListManualArticleVersions は指定記事の編集履歴を返す
//
// GET /api/v1/manual/articles/:category/:slug/versions
func (h *Handler) ListManualArticleVersions(c *gin.Context) {
	category := model.ManualCategory(c.Param("category"))
	slug := c.Param("slug")
	if slug == "" {
		RespondError(c, apperrors.WrapInvalidInput("slug is required"))
		return
	}
	// まず article を特定して ID を取得
	article, err := h.svc.ManualArticle.FindByCategoryAndSlug(c.Request.Context(), category, slug)
	if err != nil {
		RespondError(c, err)
		return
	}
	versions, err := h.svc.ManualArticle.FindVersionsByArticleID(c.Request.Context(), article.ID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toManualArticleVersionListResponse(versions))
}
