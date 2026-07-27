package httpapi

import (
	"errors"
	"log/slog"
	"maps"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// RespondError はエラーを適切なHTTPステータスコードとメッセージにマッピングして返す。
// 内部エラー(5xx)は details を露出しない。
func RespondError(c *gin.Context, err error) {
	status, message, _ := ResolveErrorResponse(err)
	c.JSON(status, gin.H{"error": message})
}

// RespondErrorWithExtras は custom extra fields を含むエラーレスポンスを返す。
// `error` と `code` を必須フィールドとして生成し、extras をマージして JSON で返却する。
// ステータスコードはエラー種別から自動判定される。
func RespondErrorWithExtras(c *gin.Context, err error, extras map[string]any) {
	status, message, code := ResolveErrorResponse(err)

	response := gin.H{
		"error": message,
		"code":  code,
	}

	// extras をマージ（既存キーを上書き可能）
	maps.Copy(response, extras)

	c.JSON(status, response)
}

// ResolveErrorResponse はエラーから HTTP ステータスコード・メッセージ・エラーコードを決定する。
// RespondError / RespondErrorWithExtras 共通の唯一の分類ロジック。
//
// BE9-2B note: this is the generic, domain-independent classification (apperrors sentinels +
// pg errors + a safe 500 fallback). internal/handler's RespondError/RespondErrorWithExtras
// facade wraps this with one additional domain-specific fallback
// (*reservation.ReservationLimitError（internal/reservation/response_error.go）, liff/reservation-only) BEFORE delegating here, so that
// httpapi itself never imports internal/service (httpapi must stay dependency-free per
// ADR-006's topological order) while today's 269 internal/handler call sites keep their exact
// existing behavior. See internal/handler/response.go for that residual branch.
func ResolveErrorResponse(err error) (status int, message, code string) {
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
	case errors.Is(err, apperrors.ErrPayloadTooLarge):
		status = http.StatusRequestEntityTooLarge
		message, code = appMessageAndCode(hasApp, appErr, "payload too large")
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
		if msg, known := classifyPgError(err); known {
			status = http.StatusBadRequest
			message = msg
		} else {
			// BUG-2026-07-27-01: 未知コードはクライアント入力起因と断定できない
			// （42703 undefined_column = model と稼働 DB スキーマの乖離＝サーバ側欠陥）。
			// 400「入力値が正しくありません」で返すとサーバ欠陥が利用者のせいに見え、
			// ペット一覧全滅の原因特定が遅れた。応答本文には pg 詳細を一切出さず、
			// SQLSTATE コードのみサーバ側ログへ残して診断可能性を確保する。
			if code, ok := pgErrorCode(err); ok {
				slog.Error("unclassified database error mapped to 500", "pg_code", code)
			}
			status = http.StatusInternalServerError
			message = "internal server error"
		}
	default:
		// 未分類エラー（既知センチネルに属さない AppError・素の error・domain-specific custom
		// error 型）は詳細を露出しない 500 に落とす。domain-specific な特別扱い（例:
		// service.ReservationLimitError）は internal/handler の facade 側で処理する。
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
