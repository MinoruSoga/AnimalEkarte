// Package httpapi holds generic Gin HTTP boundary helpers shared by every domain package
// (BE9-2A target:httpapi, ADR-006). It has no dependency on any business domain — every
// domain package may depend on httpapi, but httpapi must never depend on a domain package
// (it is the topologically-first node in ADR-006's permitted dependency graph:
// "httpapi → clinic → inventory → manualarticle → ...").
//
// Extraction history: BE9-2B's manualarticle pilot extracted the first 6 target:httpapi
// files it actually needed (context_helpers.go, response.go, response_pg.go, bind_errors.go,
// time_response.go, slice_helpers.go). BE9-2C added query_helpers.go, pagination_response.go,
// and date.go. internal/handler is retired; new HTTP helpers belong in this package when
// more than one domain needs them. Do not grow httpapi speculatively (YAGNI).
package httpapi

import (
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// maxClinicIDsParam は clinic_ids クエリパラメータの上限数 (staff_request の clinic_ids binding max=50 と整合)
const maxClinicIDsParam = 50

// ExtractContextUint64 はJWTコンテキストから string 型の値を取得し uint64 にパースする共通ヘルパー。
// missingMsg: 存在しない場合の 401 メッセージ / invalidMsg: 型変換・パース失敗時の 400 メッセージ
func ExtractContextUint64(c *gin.Context, key, missingMsg, invalidMsg string) (uint64, bool) {
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

// ExtractStaffID はJWT認証済みコンテキストから user_id（=staff_id）を取得してパースする。
func ExtractStaffID(c *gin.Context) (uint64, bool) {
	return ExtractContextUint64(c, "user_id", "missing user context", "invalid user context")
}

// OptionalStaffID はJWT認証済みコンテキストから user_id を取得する。
// 存在しない場合はエラーを書かずに nil を返す。監査ログ等のオプショナルな actor 取得に使用する。
func OptionalStaffID(c *gin.Context) *uint64 {
	id, ok := PeekStaffID(c)
	if !ok {
		return nil
	}
	return &id
}

// peekContextUint64 はコンテキスト値を読むだけで HTTP レスポンスを書かない。
// 権限判定など「後続が1回だけ RespondError する」経路で Extract* の二重書き込みを防ぐ (AUS-09 / X-10)。
func peekContextUint64(c *gin.Context, key string) (uint64, bool) {
	val, exists := c.Get(key)
	if !exists {
		return 0, false
	}
	s, ok := val.(string)
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// PeekStaffID は user_id を読むだけでレスポンスを書かない。
func PeekStaffID(c *gin.Context) (uint64, bool) {
	return peekContextUint64(c, "user_id")
}

// PeekClinicID は clinic_id を読むだけでレスポンスを書かない。
func PeekClinicID(c *gin.Context) (uint64, bool) {
	return peekContextUint64(c, "clinic_id")
}

// PeekIsSystemAdmin は is_system_admin を読むだけでレスポンスを書かない。
func PeekIsSystemAdmin(c *gin.Context) (isSystemAdmin, ok bool) {
	val, exists := c.Get("is_system_admin")
	if !exists {
		return false, false
	}
	isAdmin, ok := val.(bool)
	if !ok {
		return false, false
	}
	return isAdmin, true
}

// ExtractClinicID はJWT認証済みコンテキストから clinic_id を取得してパースする。
// 取得・パース失敗時は即座にHTTPエラーレスポンスを書いて false を返す。
// 呼び出し元はfalse時に即return すること。
func ExtractClinicID(c *gin.Context) (uint64, bool) {
	return ExtractContextUint64(c, "clinic_id", "missing clinic context", "invalid clinic context")
}

// ExtractClinicIDs はJWT認証済みコンテキストから所属医院IDリスト (clinic_ids) を取得する。
// #84: 登録時の医院指定でユーザー入力 clinic_id の所属検証に使用する。
// 取得・型変換失敗時は即座にHTTPエラーレスポンスを書いて (nil, false) を返す。
// 呼び出し元は false 時に即 return すること。
func ExtractClinicIDs(c *gin.Context) ([]uint64, bool) {
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

// ParseClinicIDsParam parses a comma-separated clinic_ids query value into a sorted,
// de-duplicated []uint64.
func ParseClinicIDsParam(raw string) ([]uint64, error) {
	parts := strings.Split(raw, ",")
	if len(parts) > maxClinicIDsParam {
		return nil, apperrors.WrapInvalidInput("too many clinic_ids")
	}
	requested := make([]uint64, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err != nil || id == 0 {
			return nil, apperrors.WrapInvalidInput("invalid clinic_ids")
		}
		requested = append(requested, id)
	}

	slices.Sort(requested)
	return slices.Compact(requested), nil
}

// AuthorizeClinicIDs verifies that every requested clinic ID is present in the
// request-time trusted clinic_ids authority. The middleware resolves that set from
// current active assignments for regular staff and from all active clinics for a
// system administrator, so neither caller type may bypass this subset check.
func AuthorizeClinicIDs(c *gin.Context, requested []uint64) bool {
	allowed, ok := ExtractClinicIDs(c)
	if !ok {
		return false
	}
	for _, id := range requested {
		if !slices.Contains(allowed, id) {
			RespondError(c, apperrors.WrapForbidden("not assigned to this clinic"))
			return false
		}
	}
	return true
}

// ResolveListClinicIDs は一覧系 API の拠点横断スコープ (#86) を解決する。
// クエリパラメータ clinic_ids（カンマ区切り）が:
//   - 無い場合: JWT/X-Clinic-ID 由来の現在の医院のみ（従来挙動・後方互換）
//   - ある場合: 所属医院 (clinic_ids context) の部分集合であることを検証して採用。
//     1件でも現在のactive clinic権限外なら403。system_adminのcontextには全active clinicが入る。
//
// 戻り値は必ず非空。検証失敗時は即座にHTTPエラーレスポンスを書いて (nil, false) を返す。
// 呼び出し元は false 時に即 return すること。
func ResolveListClinicIDs(c *gin.Context) ([]uint64, bool) {
	clinicID, ok := ExtractClinicID(c)
	if !ok {
		return nil, false
	}
	raw := c.Query("clinic_ids")
	if raw == "" {
		return []uint64{clinicID}, true
	}

	requested, err := ParseClinicIDsParam(raw)
	if err != nil {
		RespondError(c, err)
		return nil, false
	}
	if !AuthorizeClinicIDs(c, requested) {
		return nil, false
	}
	return requested, true
}

// ResolveAllClinicIDs はユーザーの全所属医院IDを返す（詳細画面の拠点横断閲覧用 #86）。
// system_admin は mainClinicID 単体を返す（全医院スキャン防止）。
// 戻り値は必ず1件以上。失敗時は HTTPエラーを書いて (nil, false) を返す。
//
// ExtractIsSystemAdmin/ExtractClinicIDs はエラー時に 401 を書くため、ここでは
// コンテキスト値を直接読み取り、存在しない場合は mainClinicID のみを返すフォールバック動作をとる。
func ResolveAllClinicIDs(c *gin.Context) ([]uint64, bool) {
	clinicID, ok := ExtractClinicID(c)
	if !ok {
		return nil, false
	}
	if val, exists := c.Get("is_system_admin"); exists {
		if isAdmin, _ := val.(bool); isAdmin {
			return []uint64{clinicID}, true
		}
	}
	if val, exists := c.Get("clinic_ids"); exists {
		if ids, ok2 := val.([]uint64); ok2 && len(ids) > 0 {
			return ids, true
		}
	}
	return []uint64{clinicID}, true
}

// ExtractIsSystemAdmin はJWT認証済みコンテキストから is_system_admin を取得する。
// 取得失敗時は即座にHTTPエラーレスポンスを書いて (false, false) を返す。
// 戻り値: (isSystemAdmin bool, ok bool)
func ExtractIsSystemAdmin(c *gin.Context) (isSystemAdmin, ok bool) {
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
