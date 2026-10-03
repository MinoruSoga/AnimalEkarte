package support

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// screenshotSignedURLTTL は一覧レスポンスの署名付き URL 有効期限
const screenshotSignedURLTTL = 15 * time.Minute

// fileUploader is the consumer-side view of infra.FileUploader.
type fileUploader interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
	GetSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// Handler serves the support HTTP boundary.
type Handler struct {
	service       Service
	uploader      fileUploader
	audit         AuditLogger
	chat          ChatCompleter
	chatRateLimit gin.HandlerFunc
}

// NewHandler initializes a Handler. uploader may be nil (screenshot upload then fails closed).
// audit may be nil (audit logging is then skipped, best-effort).
// chat may be nil (help chat is then disabled and frontend falls back to manual search).
// chatRateLimit may be nil (chat POST is then unbounded — production wiring always supplies one).
func NewHandler(service Service, uploader fileUploader, audit AuditLogger, chat ChatCompleter, chatRateLimit gin.HandlerFunc) *Handler {
	return &Handler{service: service, uploader: uploader, audit: audit, chat: chat, chatRateLimit: chatRateLimit}
}

// RegisterRoutes はサポート関連ルートを登録する。
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	s := rg.Group("/support")

	// ヘルプチャットは認証済みスタッフ全員が使える（権限ゲートなし）。
	// POST は LLM 呼び出しコストがかかるためレート制限を挟む。
	// 履歴の参照・リセットは常に JWT の clinic_id×staff_id でスコープされる。
	s.GET("/chat/status", h.ChatStatus)
	s.GET("/chat/history", h.ChatHistory)
	s.DELETE("/chat/history", h.ClearChatHistory)
	// 共有一覧はバグ報告と同じく全医院公開（質問傾向の横断分析用途 — product 決定）。
	s.GET("/chat/exchanges", h.ListChatExchanges)
	chatHandlers := []gin.HandlerFunc{h.Chat}
	if h.chatRateLimit != nil {
		chatHandlers = append([]gin.HandlerFunc{h.chatRateLimit}, chatHandlers...)
	}
	s.POST("/chat", chatHandlers...)

	g := s.Group("/bug-reports")

	// バグ報告は認証済みスタッフ全員・全医院に公開する共有ボード
	//（権限ゲート・医院絞りなし — 製品フィードバック基盤としての意図的な製品判断）。
	// 一覧・作成は全医院公開。status 更新・Plane 起票・削除は報告元医院スコープ
	//（他医院の報告は共有ボードで閲覧のみ — 2026-10 セキュリティレビュー変更）。
	g.POST("", h.CreateBugReport)
	g.GET("", h.ListBugReports)
	g.PATCH("/:id/status", h.UpdateBugReportStatus)
	g.POST("/:id/plane-ticket", h.CreatePlaneTicket)
	g.DELETE("/:id", h.DeleteBugReport)
}

// CreateBugReport はバグ報告を作成する。
//
// POST /api/v1/support/bug-reports
// multipart/form-data: title, detail, page_url, route_path, user_agent, viewport,
// app_version + 任意 file "screenshot"（PNG/JPEG/WebP）
func (h *Handler) CreateBugReport(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	if c.Request.ContentLength > bugReportMaxRequestSize {
		httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("bug report request exceeds size limit"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, bugReportMaxRequestSize)
	if err := c.Request.ParseMultipartForm(bugReportMultipartMemory); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("bug report request exceeds size limit"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid multipart form"))
		return
	}

	req, err := parseCreateBugReportRequest(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	var screenshotKey *string
	file, fileHeader, err := c.Request.FormFile("screenshot")
	switch {
	case err == nil:
		defer file.Close() //nolint:errcheck // multipart ファイルのクローズ失敗は復旧不可のため無視
		key, ok := h.uploadScreenshot(c, clinicID, file, fileHeader)
		if !ok {
			return
		}
		screenshotKey = &key
	case errors.Is(err, http.ErrMissingFile):
		// スクショなしも許容
	default:
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("bug report request exceeds size limit"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid screenshot file"))
		return
	}

	report, err := h.service.Create(c.Request.Context(), clinicID, staffID, CreateBugReportInput{
		Title:         req.Title,
		Detail:        req.Detail,
		PageURL:       req.PageURL,
		RoutePath:     req.RoutePath,
		UserAgent:     req.UserAgent,
		Viewport:      req.Viewport,
		AppVersion:    req.AppVersion,
		ScreenshotKey: screenshotKey,
	})
	if err != nil {
		if screenshotKey != nil {
			if delErr := h.uploader.Delete(context.WithoutCancel(c.Request.Context()), *screenshotKey); delErr != nil {
				slog.WarnContext(c.Request.Context(), "failed to delete uploaded screenshot on create error (best-effort)", "error", delErr, "key", *screenshotKey)
			}
		}
		httpapi.RespondError(c, err)
		return
	}

	h.logAudit(c, "support_bug_report.create", report)
	resp := toBugReportResponse(report, "", "")
	h.signScreenshotURL(c.Request.Context(), report, &resp)
	c.JSON(http.StatusCreated, resp)
}

// ListBugReports は全医院のバグ報告一覧を新しい順で返す。
// 認証済みスタッフ全員が利用可能（権限ゲート・医院絞りなし — 意図的な製品判断）。
//
// GET /api/v1/support/bug-reports
func (h *Handler) ListBugReports(c *gin.Context) {
	reports, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	data := make([]BugReportResponse, len(reports))
	for i := range reports {
		data[i] = toBugReportResponse(&reports[i].SupportBugReport, reports[i].ReporterName, reports[i].ClinicName)
		h.signScreenshotURL(c.Request.Context(), &reports[i].SupportBugReport, &data[i])
	}
	c.JSON(http.StatusOK, BugReportListResponse{Data: data})
}

// UpdateBugReportStatus は報告の対応状況を更新する。
// 操作は報告元医院スコープ — 他医院の報告は共有ボードで閲覧のみ（404）。
//
// PATCH /api/v1/support/bug-reports/:id/status
func (h *Handler) UpdateBugReportStatus(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateBugReportStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid request body"))
		return
	}
	report, err := h.service.UpdateStatus(c.Request.Context(), clinicID, id, model.SupportBugReportStatus(req.Status))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	h.logAudit(c, "support_bug_report.update_status", report)
	c.JSON(http.StatusOK, toBugReportResponse(report, "", ""))
}

// CreatePlaneTicket は報告を Plane ワークアイテムとして起票する。
// 自動起票に失敗した報告の手動再送経路でもある。起票済みなら現行状態を返す（冪等）。
// 操作は報告元医院スコープ — 他医院の報告は共有ボードで閲覧のみ（404）。
//
// POST /api/v1/support/bug-reports/:id/plane-ticket
func (h *Handler) CreatePlaneTicket(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	report, err := h.service.EnsurePlaneTicket(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	h.logAudit(c, "support_bug_report.plane_ticket", report)
	c.JSON(http.StatusOK, toBugReportResponse(report, "", ""))
}

// DeleteBugReport は報告を論理削除し、添付スクショのオブジェクトも削除する。
// 操作は報告元医院スコープ — 他医院の報告は共有ボードで閲覧のみ（404）。
//
// DELETE /api/v1/support/bug-reports/:id
func (h *Handler) DeleteBugReport(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	report, err := h.service.Get(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	// スクショのオブジェクト削除は論理削除より先に行う — ストレージ失敗時に
	// 行だけ消えると retry 経路を失い孤立オブジェクトが残るため、失敗は 502
	// で報告行を残す（呼び出し側が再試行できる）。
	if report.ScreenshotKey != nil && h.uploader != nil {
		if err := h.uploader.Delete(c.Request.Context(), *report.ScreenshotKey); err != nil {
			slog.WarnContext(c.Request.Context(), "failed to delete bug report screenshot", "error", err, "report_id", report.ID)
			httpapi.RespondError(c, apperrors.WrapBadGateway("screenshot storage delete failed; report was not deleted"))
			return
		}
	}
	report, err = h.service.Delete(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	h.logAudit(c, "support_bug_report.delete", report)
	c.Status(http.StatusNoContent)
}

// uploadScreenshot は検証済みスクショを FileUploader へ保存しオブジェクト key を返す。
// 失敗時はエラーレスポンスを書き false を返す。
func (h *Handler) uploadScreenshot(c *gin.Context, clinicID uint64, file multipart.File, fileHeader *multipart.FileHeader) (string, bool) {
	if fileHeader.Size > bugReportMaxRequestSize {
		httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("screenshot exceeds size limit"))
		return "", false
	}
	if h.uploader == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "screenshot upload unavailable"})
		return "", false
	}
	meta, err := validateScreenshot(file)
	if err != nil {
		httpapi.RespondError(c, err)
		return "", false
	}
	key := meta.uploadKey(clinicID, time.Now())
	if _, err := h.uploader.Upload(c.Request.Context(), key, file, meta.mimeType); err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to upload screenshot"))
		return "", false
	}
	return key, true
}

// signScreenshotURL は署名付き URL をレスポンスに設定する（失敗時は空のまま）
func (h *Handler) signScreenshotURL(ctx context.Context, report *model.SupportBugReport, resp *BugReportResponse) {
	if report.ScreenshotKey == nil || h.uploader == nil {
		return
	}
	url, err := h.uploader.GetSignedURL(ctx, *report.ScreenshotKey, screenshotSignedURLTTL)
	if err != nil {
		slog.WarnContext(ctx, "failed to sign bug report screenshot url", "error", err, "report_id", report.ID)
		return
	}
	resp.ScreenshotURL = url
}

func (h *Handler) logAudit(c *gin.Context, action string, report *model.SupportBugReport) {
	if h.audit == nil {
		return
	}
	if err := h.audit.LogEntry(c.Request.Context(), AuditEntry{
		ClinicID: &report.ClinicID,
		// actor は操作者（削除・起票では報告者と異なり得る）。取得失敗時は nil のまま残す。
		ActorID:    httpapi.OptionalStaffID(c),
		ActorType:  "staff",
		Action:     action,
		Resource:   "support_bug_report",
		ResourceID: &report.ID,
		NewValue:   fmt.Sprintf("title=%q status=%s", report.Title, report.Status),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}); err != nil {
		slog.WarnContext(c.Request.Context(), "failed to write audit log for support bug report", "error", err, "report_id", report.ID)
	}
}
