package handler

import (
	"errors"
	"maps"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// RespondError はエラーを適切なHTTPステータスコードとメッセージにマッピングして返す。
// 内部エラー(5xx)は details を露出しない。
func RespondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		var appErr *apperrors.AppError
		msg := "resource not found"
		if errors.As(err, &appErr) {
			msg = appErr.Message
		}
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
	case errors.Is(err, apperrors.ErrInvalidInput):
		var appErr *apperrors.AppError
		msg := "invalid input"
		if errors.As(err, &appErr) {
			msg = appErr.Message
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	case errors.Is(err, apperrors.ErrConflict):
		var appErr *apperrors.AppError
		msg := "resource conflict"
		if errors.As(err, &appErr) {
			msg = appErr.Message
		}
		c.JSON(http.StatusConflict, gin.H{"error": msg})
	case errors.Is(err, apperrors.ErrAlreadyExists):
		var appErr *apperrors.AppError
		msg := "resource already exists"
		if errors.As(err, &appErr) {
			msg = appErr.Message
		}
		c.JSON(http.StatusConflict, gin.H{"error": msg})
	case errors.Is(err, apperrors.ErrUnauthorized):
		var appErr *apperrors.AppError
		msg := "unauthorized"
		if errors.As(err, &appErr) {
			msg = appErr.Message
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
	case errors.Is(err, apperrors.ErrForbidden):
		var appErr *apperrors.AppError
		msg := "forbidden"
		if errors.As(err, &appErr) && appErr.Message != "" {
			msg = appErr.Message
		}
		c.JSON(http.StatusForbidden, gin.H{"error": msg})
	case errors.Is(err, apperrors.ErrNotImplemented):
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
	case errors.Is(err, apperrors.ErrBadGateway):
		var appErr *apperrors.AppError
		msg := "bad gateway"
		if errors.As(err, &appErr) && appErr.Message != "" {
			msg = appErr.Message
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": msg})
	case isPgError(err):
		// BUG-138: FromGORM を経由しなかった PostgreSQL エラーをここでキャッチ
		pgMsg := classifyPgError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": pgMsg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
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
// RespondError と同じ分類ロジックをベースにする。
func resolveErrorResponse(err error) (status int, message, code string) {
	// AppError からの抽出（Code / Message）
	var appErr *apperrors.AppError
	hasApp := errors.As(err, &appErr)

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		status = http.StatusNotFound
		message = "resource not found"
		if hasApp {
			message = appErr.Message
			code = appErr.Code
		}
	case errors.Is(err, apperrors.ErrInvalidInput):
		status = http.StatusBadRequest
		message = "invalid input"
		if hasApp {
			message = appErr.Message
			code = appErr.Code
		}
	case errors.Is(err, apperrors.ErrConflict):
		status = http.StatusConflict
		message = "resource conflict"
		if hasApp {
			message = appErr.Message
			code = appErr.Code
		}
	case errors.Is(err, apperrors.ErrAlreadyExists):
		status = http.StatusConflict
		message = "resource already exists"
		if hasApp {
			message = appErr.Message
			code = appErr.Code
		}
	case errors.Is(err, apperrors.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
		if hasApp {
			message = appErr.Message
			code = appErr.Code
		}
	case errors.Is(err, apperrors.ErrForbidden):
		status = http.StatusForbidden
		message = "forbidden"
		if hasApp && appErr.Message != "" {
			message = appErr.Message
			code = appErr.Code
		}
	case errors.Is(err, apperrors.ErrNotImplemented):
		status = http.StatusNotImplemented
		message = "not implemented"
		if hasApp {
			code = appErr.Code
		}
	case isPgError(err):
		status = http.StatusBadRequest
		message = classifyPgError(err)
	default:
		// カスタムエラー型（ReservationLimitError 等）のフォールバック:
		// 公開フィールド Code / Message を持つ構造体を reflect で抽出する。
		status = http.StatusConflict
		message = err.Error()
		if c, m, ok := extractCodeMessage(err); ok {
			code = c
			if m != "" {
				message = m
			}
		}
	}
	return status, message, code
}

// extractCodeMessage は err が持つ Code / Message 公開フィールドを reflect で抽出する。
// ReservationLimitError 等、apperrors 外のカスタムエラー型をサポートする。
func extractCodeMessage(err error) (code, message string, ok bool) {
	v := reflect.ValueOf(err)
	if !v.IsValid() {
		return "", "", false
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", "", false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", "", false
	}
	codeField := v.FieldByName("Code")
	msgField := v.FieldByName("Message")
	if codeField.IsValid() && codeField.Kind() == reflect.String {
		code = codeField.String()
		ok = true
	}
	if msgField.IsValid() && msgField.Kind() == reflect.String {
		message = msgField.String()
	}
	return code, message, ok
}
