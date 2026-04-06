package handler

import (
	"encoding/json"
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
	"github.com/jackc/pgx/v5/pgconn"

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
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case isPgError(err):
		// BUG-138: FromGORM を経由しなかった PostgreSQL エラーをここでキャッチ
		pgMsg := classifyPgError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": pgMsg})
	default:
		slog.ErrorContext(c.Request.Context(), "internal server error",
			slog.String("error", err.Error()),
			slog.String("path", c.FullPath()),
			slog.String("method", c.Request.Method))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// isPgError はエラーチェーンに pgconn.PgError が含まれるか判定する
func isPgError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr)
}

// classifyPgError は PostgreSQL エラーコードに基づいてユーザー向けメッセージを返す
func classifyPgError(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation
			return "参照先が存在しません"
		case "23505": // unique_violation
			return "既に登録されています"
		case "22003": // numeric_value_out_of_range
			return "数値が範囲外です"
		case "22P02": // invalid_text_representation
			return "入力値の形式が正しくありません"
		}
	}
	return "入力値が正しくありません"
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
	// BUG-129: Go 内部エラーメッセージをサニタイズし、構造体名・フィールド型の漏洩を防止
	var unmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalErr) {
		return fmt.Sprintf("%s: 正しい形式で入力してください", unmarshalErr.Field)
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "リクエストの形式が正しくありません"
	}
	return "入力値が正しくありません"
}

func formatValidationError(fe validator.FieldError) string {
	// BUG-135: フィールド名を除去し、汎化メッセージを返す（内部構造の漏洩防止）
	switch fe.Tag() {
	case "required":
		return "必須項目が入力されていません"
	case "min":
		return fmt.Sprintf("%s 以上で入力してください", fe.Param())
	case "max":
		return fmt.Sprintf("%s 以下で入力してください", fe.Param())
	case "oneof":
		return fmt.Sprintf("次のいずれかで指定してください: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return "入力値が正しくありません"
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

// extractStaffID はJWT認証済みコンテキストから user_id（=staff_id）を取得してパースする。
func extractStaffID(c *gin.Context) (uint64, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return 0, false
	}
	userIDStr, ok := val.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return 0, false
	}
	staffID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return 0, false
	}
	return staffID, true
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

// extractIsSystemAdmin はJWT認証済みコンテキストから is_system_admin を取得する。
// 取得失敗時は即座にHTTPエラーレスポンスを書いて (false, false) を返す。
// 戻り値: (isSystemAdmin bool, ok bool)
func extractIsSystemAdmin(c *gin.Context) (bool, bool) {
	val, exists := c.Get("is_system_admin")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return false, false
	}
	isAdmin, ok := val.(bool)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return false, false
	}
	return isAdmin, true
}

// RequirePermission は指定リソース・アクションの権限を持つユーザーのみ通過させる gin ミドルウェアを返す。
// system_admin は全権限バイパス。それ以外は permission_group_rules から判定する。
func (h *Handler) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.hasPermission(c, resource, action) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// parsePagination はページネーションパラメータを安全にパースする。
// page: 1以上の整数, limit: 1〜100の整数
func parsePagination(c *gin.Context) (page, limit int, err error) {
	pageStr := c.DefaultQuery("page", "1")
	// BUG-143: limit と per_page の両方をサポート（per_page は limit のエイリアス）
	limitStr := c.DefaultQuery("limit", "")
	if limitStr == "" {
		limitStr = c.DefaultQuery("per_page", "20")
	}

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
