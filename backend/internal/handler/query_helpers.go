package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

const defaultMaxPaginationLimit = 100

// parsePagination はページネーションパラメータを安全にパースする。
// page: 1以上の整数, limit: 1〜100の整数
func parsePagination(c *gin.Context) (page, limit int, err error) {
	return parsePaginationWithMax(c, defaultMaxPaginationLimit)
}

// parsePaginationWithMax はページネーション上限だけを呼び出し元で調整できる共通パーサ。
func parsePaginationWithMax(c *gin.Context, maxLimit int) (page, limit int, err error) {
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
	if err != nil || limit < 1 || limit > maxLimit {
		return 0, 0, apperrors.WrapInvalidInput(fmt.Sprintf("limit は1〜%dの範囲で指定してください", maxLimit))
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
