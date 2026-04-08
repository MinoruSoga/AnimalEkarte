package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// センチネルエラー
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
)

// AppError はアプリケーション固有のエラー
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// エラーラッピングヘルパー
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func WrapNotFound(resource, id string) error {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: "not found",
		Err:     fmt.Errorf("%s(id=%s): %w", resource, id, ErrNotFound),
	}
}

func WrapInvalidInput(message string) error {
	return &AppError{
		Code:    "INVALID_INPUT",
		Message: message,
		Err:     ErrInvalidInput,
	}
}

// エラー判定ヘルパー
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// WrapAlreadyExists は重複リソースエラーを生成する
func WrapAlreadyExists(resource, identifier string) error {
	return &AppError{
		Code:    "ALREADY_EXISTS",
		Message: fmt.Sprintf("%s '%s' already exists", resource, identifier),
		Err:     ErrAlreadyExists,
	}
}

// IsAlreadyExists はErrAlreadyExistsかどうかを判定する
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// WrapConflict は依存データによる操作不可エラーを生成する（409 Conflict）
// 削除時に FK 参照先が存在する場合など。
func WrapConflict(message string) error {
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Err:     ErrAlreadyExists, // ErrAlreadyExists → 409 Conflict にマッピング済み
	}
}

// IsConflict は 409 Conflict 系エラーかどうかを判定する
func IsConflict(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// WrapForbidden はアクセス拒否エラーを生成する
func WrapForbidden(message string) error {
	return &AppError{
		Code:    "FORBIDDEN",
		Message: message,
		Err:     ErrForbidden,
	}
}

// WrapUnauthorized は認証エラーを生成する
func WrapUnauthorized(message string) error {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Err:     ErrUnauthorized,
	}
}

// WrapInternalServerError は内部サーバーエラーを生成する
func WrapInternalServerError(message string) error {
	return &AppError{
		Code:    "INTERNAL",
		Message: message,
		Err:     errors.New("internal server error"),
	}
}

// FromGORM は GORM のエラーを AppError に変換する
func FromGORM(err error, resource, id string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return WrapNotFound(resource, id)
	}
	// BUG-138: pgx ドライバのエンコードエラー（pgconn.PgError ではない）をキャッチ。
	// int32 範囲超過などで "unable to encode" が発生した場合。
	errMsg := err.Error()
	if strings.Contains(errMsg, "unable to encode") ||
		strings.Contains(errMsg, "greater than maximum value") ||
		strings.Contains(errMsg, "less than minimum value") {
		return WrapInvalidInput("数値が範囲外です")
	}
	// BUG-138: PostgreSQL エラーコードに基づくハンドリング
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation
			return WrapInvalidInput("参照先が存在しません")
		case "23505": // unique_violation
			return WrapAlreadyExists(resource, "")
		case "22003": // numeric_value_out_of_range
			return WrapInvalidInput("数値が範囲外です")
		case "22P02": // invalid_text_representation (e.g. invalid integer)
			return WrapInvalidInput("入力値の形式が正しくありません")
		}
	}
	// BUG-129: リソース名は内部ログ用。ユーザーには汎化メッセージを返す
	return Wrap(err, "database error")
}
