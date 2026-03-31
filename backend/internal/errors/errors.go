package errors

import (
	"errors"
	"fmt"

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
		Message: fmt.Sprintf("%s with id %s not found", resource, id),
		Err:     ErrNotFound,
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

// FromGORM は GORM のエラーを AppError に変換する
func FromGORM(err error, resource string, id string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return WrapNotFound(resource, id)
	}
	return Wrap(err, fmt.Sprintf("database error on %s", resource))
}
