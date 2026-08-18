package apperrors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// センチネルエラー
var (
	ErrNotFound        = errors.New("resource not found")
	ErrAlreadyExists   = errors.New("resource already exists")
	ErrConflict        = errors.New("resource conflict") // FK依存・削除不可など 409 Conflict 専用
	ErrInvalidInput    = errors.New("invalid input")
	ErrPayloadTooLarge = errors.New("payload too large")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrNotImplemented  = errors.New("not implemented")
)

// Stable domain conflict codes (HTTP 409). FE maps these to localized messages.
const (
	// CodePermissionGroupNameConflict is returned when a clinic-scoped permission
	// group name collides with an existing group (constraint uk_permission_groups).
	CodePermissionGroupNameConflict = "permission_group_name_conflict"
	// CodeAnimalSpeciesNameConflict is returned when an active animal species name
	// collides (constraint idx_animal_species_name).
	CodeAnimalSpeciesNameConflict = "animal_species_name_conflict"
	// CodeShiftTemplateNameConflict is returned when a clinic-scoped shift template
	// name collides (constraint uk_shift_templates_clinic_name). BUG-026.
	CodeShiftTemplateNameConflict = "shift_template_name_conflict"
	// CodeLstepAutoManagedPrefixConflict is returned when an auto-managed L-step
	// prefix collides (constraint lstep_auto_managed_prefixes_prefix_key). BUG-026.
	CodeLstepAutoManagedPrefixConflict = "lstep_auto_managed_prefix_conflict"
	// CodeCageNameConflict is returned when a clinic-scoped cage name collides
	// (constraint idx_cages_clinic_name). BUG-022.
	CodeCageNameConflict = "cage_name_conflict"
	// CodeOccupationNameConflict is returned when a clinic-scoped occupation name
	// collides (constraint idx_occupations_clinic_name). BUG-022.
	CodeOccupationNameConflict = "occupation_name_conflict"
)

// Measured PostgreSQL unique constraint names used for fail-closed mapping.
// Values match backend/migrations/001_init.sql / live DB (not guessed).
const (
	// ConstraintPermissionGroupName is UNIQUE (clinic_id, name) WHERE deleted_at IS NULL.
	ConstraintPermissionGroupName = "uk_permission_groups"
	// ConstraintPermissionGroupRules is UNIQUE (group_id, resource) — must NOT map to name conflict.
	ConstraintPermissionGroupRules = "uk_permission_group_rules"
	// ConstraintAnimalSpeciesName is UNIQUE (name) WHERE is_active = true.
	ConstraintAnimalSpeciesName = "idx_animal_species_name"
	// ConstraintShiftTemplateName is UNIQUE (clinic_id, name) WHERE deleted_at IS NULL.
	ConstraintShiftTemplateName = "uk_shift_templates_clinic_name"
	// ConstraintLstepAutoManagedPrefix is UNIQUE (prefix) on lstep_auto_managed_prefixes.
	ConstraintLstepAutoManagedPrefix = "lstep_auto_managed_prefixes_prefix_key"
	// ConstraintCageName is UNIQUE (clinic_id, name) WHERE deleted_at IS NULL.
	ConstraintCageName = "idx_cages_clinic_name"
	// ConstraintOccupationName is UNIQUE (clinic_id, name) WHERE deleted_at IS NULL.
	ConstraintOccupationName = "idx_occupations_clinic_name"
)

// AppError はアプリケーション固有のエラー
type AppError struct {
	Code    string
	Message string
	// Params holds safe, client-facing structured fields (e.g. input name echo).
	// Never put constraint names, table names, SQL, or tenant-existence signals here.
	Params map[string]string
	Err    error
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

func WrapPayloadTooLarge(message string) error {
	return &AppError{
		Code:    "PAYLOAD_TOO_LARGE",
		Message: message,
		Err:     ErrPayloadTooLarge,
	}
}

func IsPayloadTooLarge(err error) bool {
	return errors.Is(err, ErrPayloadTooLarge)
}

// WrapAlreadyExists は重複リソースエラーを生成する
func WrapAlreadyExists(resource, identifier string) error {
	return &AppError{
		Code:    "ALREADY_EXISTS",
		Message: fmt.Sprintf("%s '%s' already exists", resource, identifier),
		Err:     ErrAlreadyExists,
	}
}

// WrapAlreadyExistsMessage は完成済みのクライアント向けメッセージで ALREADY_EXISTS を返す。
// すでに日本語の完成文を持つ場合に使う（WrapAlreadyExists の英語テンプレートに埋め込まない）。
// BUG-024 / BUG-019: owner email/phone 重複など。
func WrapAlreadyExistsMessage(message string) error {
	if message == "" {
		message = "resource already exists"
	}
	return &AppError{
		Code:    "ALREADY_EXISTS",
		Message: message,
		Err:     ErrAlreadyExists,
	}
}

// IsAlreadyExists はErrAlreadyExistsかどうかを判定する
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// WrapNameConflict builds a 409 domain name-conflict with a stable code and
// optional safe name param (request input echo only).
func WrapNameConflict(code, name string) error {
	params := map[string]string{}
	if name != "" {
		params["name"] = name
	}
	return &AppError{
		Code:    code,
		Message: "resource already exists",
		Params:  params,
		Err:     ErrConflict,
	}
}

// IsNameConflict reports whether err is a domain name-conflict AppError for code.
func IsNameConflict(err error, code string) bool {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Code == code && errors.Is(err, ErrConflict)
}

// AsNameUniqueConflict returns a domain name-conflict error when err is a
// PostgreSQL unique_violation on the given constraint. Otherwise returns nil
// (fail-closed: do not elevate without a matching measured constraint name).
func AsNameUniqueConflict(err error, name, constraintName, code string) error {
	if err == nil || constraintName == "" || code == "" {
		return nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return nil
	}
	if pgErr.ConstraintName != constraintName {
		return nil
	}
	return WrapNameConflict(code, name)
}

// ConflictHTTPExtras returns additive JSON extras for RespondErrorWithExtras
// when the error carries safe Params. Returns nil when there are no params.
func ConflictHTTPExtras(err error) map[string]any {
	var appErr *AppError
	if !errors.As(err, &appErr) || len(appErr.Params) == 0 {
		return nil
	}
	// Copy to avoid exposing the live map to callers that mutate.
	params := make(map[string]string, len(appErr.Params))
	for k, v := range appErr.Params {
		params[k] = v
	}
	return map[string]any{"params": params}
}

// RespondWithConflictCode reports whether handlers should emit code via
// RespondErrorWithExtras (domain name conflicts always need the stable code).
func RespondWithConflictCode(err error) bool {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return false
	}
	switch appErr.Code {
	case CodePermissionGroupNameConflict,
		CodeAnimalSpeciesNameConflict,
		CodeShiftTemplateNameConflict,
		CodeLstepAutoManagedPrefixConflict,
		CodeCageNameConflict,
		CodeOccupationNameConflict:
		return true
	default:
		return len(appErr.Params) > 0
	}
}

// WrapConflict は依存データによる操作不可エラーを生成する（409 Conflict）
// 削除時に FK 参照先が存在する場合など。
func WrapConflict(message string) error {
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Err:     ErrConflict,
	}
}

// IsConflict は 409 Conflict 系エラーかどうかを判定する
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
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

// WrapNotImplemented は未実装エラーを生成する（501 Not Implemented）
func WrapNotImplemented(message string) error {
	return &AppError{
		Code:    "NOT_IMPLEMENTED",
		Message: message,
		Err:     ErrNotImplemented,
	}
}

// ErrBadGateway は上流サービス（LINE等）のエラーを表す（502 Bad Gateway）。
var ErrBadGateway = errors.New("bad gateway")

// WrapBadGateway は上流サービスエラーを AppError でラップする。
func WrapBadGateway(message string) error {
	return &AppError{
		Code:    "BAD_GATEWAY",
		Message: message,
		Err:     ErrBadGateway,
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
			// Preserve the original pg error so service-layer mappers can read
			// ConstraintName for fail-closed domain conflict elevation.
			return &AppError{
				Code:    "ALREADY_EXISTS",
				Message: fmt.Sprintf("%s '%s' already exists", resource, ""),
				Err:     fmt.Errorf("%w: %w", ErrAlreadyExists, err),
			}
		case "22003": // numeric_value_out_of_range
			return WrapInvalidInput("数値が範囲外です")
		case "22P02": // invalid_text_representation (e.g. invalid integer)
			return WrapInvalidInput("入力値の形式が正しくありません")
		}
	}
	// BUG-129: リソース名は内部ログ用。ユーザーには汎化メッセージを返す
	return Wrap(err, "database error")
}

// pgxEncodeRangeNeedles are the only message fragments FromGORM may use for
// encode/range classification (DEC-34 exception; no typed pgx encode error).
var pgxEncodeRangeNeedles = []string{
	"unable to encode",
	"greater than maximum value",
	"less than minimum value",
}

func isPgxEncodeRangeMessage(errMsg string) bool {
	for _, needle := range pgxEncodeRangeNeedles {
		if strings.Contains(errMsg, needle) {
			return true
		}
	}
	return false
}
