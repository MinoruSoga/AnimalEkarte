package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"reflect"
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
		slog.ErrorContext(c.Request.Context(), "internal server error",
			slog.String("error", err.Error()),
			slog.String("path", c.FullPath()),
			slog.String("method", c.Request.Method))
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
	// BUG-155: json タグに近い snake_case フィールド名を返す（API 仕様として公開情報）
	field := camelToSnake(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s は必須です", field)
	case "min":
		return fmt.Sprintf("%s は %s 以上で入力してください", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s は %s 以下で入力してください", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s は次のいずれかで指定してください: %s", field, strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return fmt.Sprintf("%s の値が正しくありません", field)
	}
}

// camelToSnake は CamelCase を snake_case に変換する。
// 連続した大文字（頭字語）は 1 単語として扱う。
// "OwnerName" → "owner_name"
// "IsDangerous" → "is_dangerous"
// "TypeID" → "type_id"     ← BUG-LINE-010: 以前は "type_i_d" になっていた
// "HTTPServer" → "http_server"
func camelToSnake(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			// 直前が小文字/数字 → 単語境界として `_` を挿入
			// 直前が大文字で次が小文字 → 頭字語の末尾扱いで `_` を挿入 ("HTTPServer" → "http_server")
			// それ以外（連続大文字の途中）は `_` を挿入しない
			insertUnderscore := !unicode.IsUpper(prev)
			if !insertUnderscore && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				insertUnderscore = true
			}
			if insertUnderscore {
				b.WriteByte('_')
			}
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
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s は YYYY-MM-DD 形式で入力してください", key))
	}
	return &s, nil
}

// extractContextUint64 はJWTコンテキストから string 型の値を取得し uint64 にパースする共通ヘルパー。
// missingMsg: 存在しない場合の 401 メッセージ / invalidMsg: 型変換・パース失敗時の 400 メッセージ
func extractContextUint64(c *gin.Context, key, missingMsg, invalidMsg string) (uint64, bool) {
	val, exists := c.Get(key)
	if !exists {
		RespondError(c, apperrors.WrapUnauthorized(missingMsg))
		return 0, false
	}
	s, ok := val.(string)
	if !ok {
		RespondError(c, apperrors.WrapInvalidInput(invalidMsg))
		return 0, false
	}
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(invalidMsg))
		return 0, false
	}
	return id, true
}

// extractStaffID はJWT認証済みコンテキストから user_id（=staff_id）を取得してパースする。
func extractStaffID(c *gin.Context) (uint64, bool) {
	return extractContextUint64(c, "user_id", "missing user context", "invalid user context")
}

// optionalStaffID はJWT認証済みコンテキストから user_id を取得する。
// 存在しない場合はエラーを書かずに nil を返す。監査ログ等のオプショナルな actor 取得に使用する。
func optionalStaffID(c *gin.Context) *uint64 {
	val, exists := c.Get("user_id")
	if !exists {
		return nil
	}
	s, ok := val.(string)
	if !ok {
		return nil
	}
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil
	}
	return &id
}

// extractClinicID はJWT認証済みコンテキストから clinic_id を取得してパースする。
// 取得・パース失敗時は即座にHTTPエラーレスポンスを書いて false を返す。
// 呼び出し元はfalse時に即return すること。
func extractClinicID(c *gin.Context) (uint64, bool) {
	return extractContextUint64(c, "clinic_id", "missing clinic context", "invalid clinic context")
}

// extractIsSystemAdmin はJWT認証済みコンテキストから is_system_admin を取得する。
// 取得失敗時は即座にHTTPエラーレスポンスを書いて (false, false) を返す。
// 戻り値: (isSystemAdmin bool, ok bool)
func extractIsSystemAdmin(c *gin.Context) (isSystemAdmin, ok bool) {
	val, exists := c.Get("is_system_admin")
	if !exists {
		RespondError(c, apperrors.WrapUnauthorized("missing user context"))
		return false, false
	}
	isAdmin, ok := val.(bool)
	if !ok {
		RespondError(c, apperrors.WrapInvalidInput("invalid user context"))
		return false, false
	}
	return isAdmin, true
}

// RequirePermission は指定リソース・アクションの権限を持つユーザーのみ通過させる gin ミドルウェアを返す。
// system_admin は全権限バイパス。それ以外は permission_group_rules から判定する。
func (h *Handler) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.hasPermission(c, resource, action) {
			RespondError(c, apperrors.WrapForbidden("forbidden"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePermissionAny は指定された複数の(リソース, アクション)ペアの内、いずれかの権限を持つユーザーのみ通過させる。
// system_admin は全権限バイパス。複数の権限オプション(OR)をサポート。
func (h *Handler) RequirePermissionAny(permissions ...struct{ Resource, Action string }) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, perm := range permissions {
			if h.hasPermission(c, perm.Resource, perm.Action) {
				c.Next()
				return
			}
		}
		RespondError(c, apperrors.WrapForbidden("forbidden"))
		c.Abort()
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
		return 0, 0, apperrors.WrapInvalidInput("page は1以上の整数で指定してください")
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, apperrors.WrapInvalidInput("limit は1〜100の範囲で指定してください")
	}

	return page, limit, nil
}

// parseOptionalUint64Query はクエリパラメータを optional な uint64 にパースする汎用ヘルパー。
// パラメータが空文字の場合は (nil, true) を返す。
// パース失敗時は即座に HTTP 400 レスポンスを書いて (nil, false) を返す。
// 呼び出し元は false 時に即 return すること。
func parseOptionalUint64Query(c *gin.Context, key string) (*uint64, bool) {
	s := c.Query(key)
	if s == "" {
		return nil, true
	}
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid %s", key)))
		return nil, false
	}
	return &id, true
}

// parseIDParam は URL path parameter を uint64 にパースする汎用ヘルパー。
// パース失敗時は即座に HTTP 400 レスポンスを書いて false を返す。
// 呼び出し元は false 時に即 return すること。
func parseIDParam(c *gin.Context, key string) (uint64, bool) {
	s := c.Param(key)
	if s == "" {
		RespondError(c, apperrors.WrapInvalidInput("パラメータが不足しています: "+key))
		return 0, false
	}
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("パラメータの形式が不正です: "+key))
		return 0, false
	}
	if id == 0 {
		RespondError(c, apperrors.WrapInvalidInput("IDは1以上を指定してください"))
		return 0, false
	}
	return id, true
}
