package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

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

// extractClinicIDs はJWT認証済みコンテキストから所属医院IDリスト (clinic_ids) を取得する。
// #84: 登録時の医院指定でユーザー入力 clinic_id の所属検証に使用する。
// 取得・型変換失敗時は即座にHTTPエラーレスポンスを書いて (nil, false) を返す。
// 呼び出し元は false 時に即 return すること。
func extractClinicIDs(c *gin.Context) ([]uint64, bool) {
	val, exists := c.Get("clinic_ids")
	if !exists {
		RespondError(c, apperrors.WrapUnauthorized("missing clinic context"))
		return nil, false
	}
	ids, ok := val.([]uint64)
	if !ok {
		RespondError(c, apperrors.WrapInvalidInput("invalid clinic context"))
		return nil, false
	}
	return ids, true
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
