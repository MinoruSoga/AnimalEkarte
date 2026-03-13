package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		// 内部エラーの詳細は絶対に露出しない
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
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
