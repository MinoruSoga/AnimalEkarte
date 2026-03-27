package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
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
	case errors.Is(err, apperrors.ErrAlreadyExists):
		var appErr *apperrors.AppError
		msg := "resource already exists"
		if errors.As(err, &appErr) {
			msg = appErr.Message
		}
		c.JSON(http.StatusConflict, gin.H{"error": msg})
	case errors.Is(err, apperrors.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		slog.ErrorContext(c.Request.Context(), "internal server error",
			slog.String("error", err.Error()),
			slog.String("path", c.FullPath()),
			slog.String("method", c.Request.Method))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// parseBindError は validator.ValidationErrors を人間可読メッセージに変換する。
// ValidationErrors でない場合はそのまま err.Error() を返す。
func parseBindError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msgs := make([]string, 0, len(ve))
		for _, fe := range ve {
			msgs = append(msgs, formatValidationError(fe))
		}
		return strings.Join(msgs, "; ")
	}
	return err.Error()
}

func formatValidationError(fe validator.FieldError) string {
	field := camelToSnake(fe.Field())
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return fmt.Sprintf("%s is invalid (%s)", field, fe.Tag())
	}
}

// camelToSnake は CamelCase を snake_case に変換する。
// "OwnerName" → "owner_name", "IsDangerous" → "is_dangerous"
func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// parseDateQuery はクエリパラメータから YYYY-MM-DD 形式の日付を安全にパースする。
// 空文字列の場合は nil を返す。不正な形式の場合はエラーを返す。
func parseDateQuery(c *gin.Context, key string) (*string, error) {
	s := c.Query(key)
	if s == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s must be YYYY-MM-DD format", key))
	}
	return &s, nil
}

// extractClinicID はJWT認証済みコンテキストから clinic_id を取得してパースする。
// 取得・パース失敗時は即座にHTTPエラーレスポンスを書いて false を返す。
// 呼び出し元はfalse時に即return すること。
func extractClinicID(c *gin.Context) (uint64, bool) {
	val, exists := c.Get("clinic_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing clinic context"})
		return 0, false
	}
	clinicIDStr, ok := val.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid clinic context"})
		return 0, false
	}
	clinicID, err := strconv.ParseUint(clinicIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid clinic context"})
		return 0, false
	}
	return clinicID, true
}

// extractUserType はJWT認証済みコンテキストから user_type を取得する。
// 取得失敗時は即座にHTTPエラーレスポンスを書いて false を返す。
func extractUserType(c *gin.Context) (model.UserType, bool) {
	val, exists := c.Get("user_type")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return "", false
	}
	ut, ok := val.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return "", false
	}
	return model.UserType(ut), true
}

// parsePagination はページネーションパラメータを安全にパースする。
// page: 1以上の整数, limit: 1〜100の整数
func parsePagination(c *gin.Context) (page, limit int, err error) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err = strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, 0, apperrors.WrapInvalidInput("page must be a positive integer")
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, apperrors.WrapInvalidInput("limit must be between 1 and 100")
	}

	return page, limit, nil
}
