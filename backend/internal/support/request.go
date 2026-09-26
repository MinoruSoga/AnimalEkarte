package support

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/google/uuid"
)

const (
	// bugReportMaxRequestSize は multipart リクエスト全体の上限（スクショ込み）
	bugReportMaxRequestSize = 8 << 20 // 8 MiB
	// bugReportMultipartMemory は ParseMultipartForm のメモリ上限
	bugReportMultipartMemory = 4 << 20 // 4 MiB

	bugReportTitleMaxLength = 200
	bugReportTextMaxLength  = 4000
	bugReportMetaMaxLength  = 500
)

// allowedScreenshotMIME は受け付けるスクショ形式 → 拡張子
var allowedScreenshotMIME = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

// createBugReportRequest は POST /support/bug-reports のフォームフィールド
type createBugReportRequest struct {
	Title      string
	Detail     string
	PageURL    string
	RoutePath  string
	UserAgent  string
	Viewport   string
	AppVersion string
}

// parseCreateBugReportRequest は multipart フォームのフィールドを読み取り検証する。
// 先に ParseMultipartForm が呼ばれている前提。
func parseCreateBugReportRequest(c *gin.Context) (createBugReportRequest, error) {
	req := createBugReportRequest{
		Title:      strings.TrimSpace(c.PostForm("title")),
		Detail:     strings.TrimSpace(c.PostForm("detail")),
		PageURL:    strings.TrimSpace(c.PostForm("page_url")),
		RoutePath:  strings.TrimSpace(c.PostForm("route_path")),
		UserAgent:  strings.TrimSpace(c.PostForm("user_agent")),
		Viewport:   strings.TrimSpace(c.PostForm("viewport")),
		AppVersion: strings.TrimSpace(c.PostForm("app_version")),
	}
	if req.Title == "" {
		return req, apperrors.WrapInvalidInput("title is required")
	}
	if len(req.Title) > bugReportTitleMaxLength {
		return req, apperrors.WrapInvalidInput("title is too long")
	}
	if len(req.Detail) > bugReportTextMaxLength {
		return req, apperrors.WrapInvalidInput("detail is too long")
	}
	for _, v := range []string{req.PageURL, req.RoutePath, req.UserAgent, req.Viewport, req.AppVersion} {
		if len(v) > bugReportMetaMaxLength {
			return req, apperrors.WrapInvalidInput("metadata field is too long")
		}
	}
	return req, nil
}

// updateBugReportStatusRequest は PATCH /support/bug-reports/:id/status のボディ
type updateBugReportStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// screenshotUpload は検証済みスクショのアップロードメタデータ
type screenshotUpload struct {
	mimeType  string
	extension string
}

// validateScreenshot は Content-Type スニッフィングで PNG/JPEG/WebP のみ許可する。
// file は検証後も先頭から読めるよう Seek 位置を戻す。
func validateScreenshot(file multipart.File) (*screenshotUpload, error) {
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && n == 0 {
		return nil, apperrors.WrapInvalidInput("failed to read screenshot")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, apperrors.WrapInvalidInput("failed to read screenshot")
	}
	mimeType := http.DetectContentType(buf[:n])
	ext, ok := allowedScreenshotMIME[mimeType]
	if !ok {
		return nil, apperrors.WrapInvalidInput("screenshot must be PNG, JPEG or WebP")
	}
	return &screenshotUpload{mimeType: mimeType, extension: ext}, nil
}

// uploadKey は FileUploader に渡すオブジェクト key を生成する。
// clinic_id プレフィックスでストレージ上もテナント分離する。
func (u *screenshotUpload) uploadKey(clinicID uint64, now time.Time) string {
	return fmt.Sprintf(
		"support-bug-reports/clinic-%d/%s-%s%s",
		clinicID,
		now.UTC().Format("20060102-150405"),
		uuid.NewString(),
		u.extension,
	)
}
