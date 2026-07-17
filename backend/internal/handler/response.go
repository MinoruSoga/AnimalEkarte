package handler

import (
	"errors"
	"maps"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// RespondError はエラーを適切なHTTPステータスコードとメッセージにマッピングして返す。
// 内部エラー(5xx)は details を露出しない。
func RespondError(c *gin.Context, err error) {
	status, message, _ := resolveErrorResponse(err)
	c.JSON(status, gin.H{"error": message})
}

// RespondErrorWithExtras は custom extra fields を含むエラーレスポンスを返す。
// `error` と `code` を必須フィールドとして生成し、extras をマージして JSON で返却する。
// ステータスコードはエラー種別から自動判定される。
func RespondErrorWithExtras(c *gin.Context, err error, extras map[string]any) {
	status, message, code := resolveErrorResponse(err)

	response := gin.H{
		"error": message,
		"code":  code,
	}

	// extras をマージ（既存キーを上書き可能）
	maps.Copy(response, extras)

	c.JSON(status, response)
}

// resolveErrorResponse はエラーから HTTP ステータスコード・メッセージ・エラーコードを決定する。
// RespondError / RespondErrorWithExtras 共通の唯一の分類ロジック。
func resolveErrorResponse(err error) (status int, message, code string) {
	// AppError からの抽出（Code / Message）
	var appErr *apperrors.AppError
	hasApp := errors.As(err, &appErr)

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		status = http.StatusNotFound
		message, code = appMessageAndCode(hasApp, appErr, "resource not found")
	case errors.Is(err, apperrors.ErrInvalidInput):
		status = http.StatusBadRequest
		message, code = appMessageAndCode(hasApp, appErr, "invalid input")
	case errors.Is(err, apperrors.ErrConflict):
		status = http.StatusConflict
		message, code = appMessageAndCode(hasApp, appErr, "resource conflict")
	case errors.Is(err, apperrors.ErrAlreadyExists):
		status = http.StatusConflict
		message, code = appMessageAndCode(hasApp, appErr, "resource already exists")
	case errors.Is(err, apperrors.ErrUnauthorized):
		status = http.StatusUnauthorized
		message, code = appMessageAndCode(hasApp, appErr, "unauthorized")
	case errors.Is(err, apperrors.ErrForbidden):
		status = http.StatusForbidden
		message, code = appMessageAndCode(hasApp, appErr, "forbidden")
	case errors.Is(err, apperrors.ErrNotImplemented):
		// NotImplemented は統一前から常に固定メッセージ（appErr.Message を無視）だった。
		// appMessageAndCode の Message!="" 上書きを適用すると、既存の
		// WrapNotImplemented(customMsg) 呼出元（reservation_handler.go 等）で
		// レスポンス本文が意図せず変わるため、この 1 ケースのみ現状維持する。
		status = http.StatusNotImplemented
		message = "not implemented"
		if hasApp {
			code = appErr.Code
		}
	case errors.Is(err, apperrors.ErrBadGateway):
		status = http.StatusBadGateway
		message, code = appMessageAndCode(hasApp, appErr, "bad gateway")
	case isPgError(err):
		status = http.StatusBadRequest
		message = classifyPgError(err)
	case !hasApp:
		// カスタムエラー型のフォールバック: 対象は service.ReservationLimitError のみ
		// （唯一の利用元は liff_handler.go の RespondErrorWithExtras）。
		var rle *service.ReservationLimitError
		if errors.As(err, &rle) {
			status = http.StatusConflict
			code = rle.Code
			message = rle.Message
			if message == "" {
				message = err.Error()
			}
			break
		}
		status = http.StatusInternalServerError
		message = "internal server error"
	default:
		// 未分類エラー（既知センチネルに属さない AppError・素の error）は
		// 詳細を露出しない 500 に落とす。
		status = http.StatusInternalServerError
		message = "internal server error"
	}
	return status, message, code
}

// appMessageAndCode は AppError から (message, code) を導出する。
// Message が空の場合は defaultMessage にフォールバックする（空文字列露出防止）。
func appMessageAndCode(hasApp bool, appErr *apperrors.AppError, defaultMessage string) (message, code string) {
	message = defaultMessage
	if hasApp {
		code = appErr.Code
		if appErr.Message != "" {
			message = appErr.Message
		}
	}
	return message, code
}
