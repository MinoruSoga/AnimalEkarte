package httpapi

import (
	"errors"
	"maps"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// RespondError はエラーを適切なHTTPステータスコードとメッセージにマッピングして返す。
// 内部エラー(5xx)は details を露出しない。
func RespondError(c *gin.Context, err error) {
	status, message, _ := ResolveErrorResponse(err)
	if err != nil && status >= http.StatusInternalServerError {
		_ = c.Error(err)
	}
	c.JSON(status, gin.H{"error": message})
}

// RespondErrorWithExtras は custom extra fields を含むエラーレスポンスを返す。
// `error` と `code` を必須フィールドとして生成し、extras をマージして JSON で返却する。
// ステータスコードはエラー種別から自動判定される。
func RespondErrorWithExtras(c *gin.Context, err error, extras map[string]any) {
	status, message, code := ResolveErrorResponse(err)
	if err != nil && status >= http.StatusInternalServerError {
		_ = c.Error(err)
	}

	response := gin.H{
		"error": message,
		"code":  code,
	}

	// extras をマージ（既存キーを上書き可能）
	maps.Copy(response, extras)

	c.JSON(status, response)
}

// RespondErrorPreferringConflictCode emits stable domain conflict `code` and
// safe `params` via RespondErrorWithExtras when present; otherwise RespondError.
// Additive only for domain name-conflict paths — other endpoints keep RespondError.
func RespondErrorPreferringConflictCode(c *gin.Context, err error) {
	if apperrors.RespondWithConflictCode(err) {
		RespondErrorWithExtras(c, err, apperrors.ConflictHTTPExtras(err))
		return
	}
	RespondError(c, err)
}

// ResolveErrorResponse はエラーから HTTP ステータスコード・メッセージ・エラーコードを決定する。
// RespondError / RespondErrorWithExtras 共通の唯一の分類ロジック。
//
// この関数は分類だけを行う。domain 固有の ReservationLimitError は
// internal/reservation/response_error.go で先に処理される。
// RespondError / RespondErrorWithExtras は、分類結果が 5xx の場合に Gin error を登録し、
// RequestLoggingMiddleware が request 境界で記録できるようにする。
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
			// ペット一覧全滅の原因特定が遅れた。
			//
			// ResolveErrorResponse 自体はログを出さない。RespondError /
			// RespondErrorWithExtras が 5xx を Gin error として登録し、
			// RequestLoggingMiddleware が request 境界で SQLSTATE を記録する。
			// domain service に同じ error のログが残る重複は、サーバー側欠陥の診断に
			// SQLSTATE が不可欠なため §8 の例外として許容する。応答本文には pg 詳細を出さない。
			status = http.StatusInternalServerError
			message = "internal server error"
		}
	default:
		// 未分類エラー（既知センチネルに属さない AppError・素の error・domain-specific custom
		// error 型）は詳細を露出しない 500 に落とす。reservation.ReservationLimitError は
		// internal/reservation/response_error.go で先に処理する。
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
