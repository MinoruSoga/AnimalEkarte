package httpapi

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// DefaultMaxPaginationLimit is the default upper bound for ParsePagination's limit query
// parameter (HTTP pagination upper bound).
const DefaultMaxPaginationLimit = 100

// ParsePagination parses page/limit (or the per_page alias) query parameters using
// DefaultMaxPaginationLimit as the upper bound.
func ParsePagination(c *gin.Context) (page, limit int, err error) {
	return ParsePaginationWithMax(c, DefaultMaxPaginationLimit)
}

// ParsePaginationWithMax parses pagination query parameters with a caller-supplied upper
// bound for limit. page must be >= 1; limit must be in [1, maxLimit]. limit falls back to
// the per_page alias when limit is not supplied (BUG-143).
func ParsePaginationWithMax(c *gin.Context, maxLimit int) (page, limit int, err error) {
	pageStr := c.DefaultQuery("page", "1")
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

// ParseOptionalUint64Query parses a query parameter into an optional uint64. An absent/empty
// parameter returns (nil, true). A malformed parameter writes an HTTP 400 response and
// returns (nil, false) — callers must return immediately when ok is false.
func ParseOptionalUint64Query(c *gin.Context, key string) (*uint64, bool) {
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

// ParseIDParam parses a URL path parameter into a uint64 ID (must be >= 1). A malformed or
// missing parameter writes an HTTP 400 response and returns (0, false) — callers must return
// immediately when ok is false.
func ParseIDParam(c *gin.Context, key string) (uint64, bool) {
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

// ParseUUIDParam parses a URL path parameter into a uuid.UUID. A malformed parameter writes
// an HTTP 400 response and returns (uuid.Nil, false) — callers must return immediately when
// ok is false.
func ParseUUIDParam(c *gin.Context, key string) (uuid.UUID, bool) {
	s := c.Param(key)
	id, err := uuid.Parse(s)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("パラメータの形式が不正です: %s は有効な UUID ではありません", key)))
		return uuid.Nil, false
	}
	return id, true
}
