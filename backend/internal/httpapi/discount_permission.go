// discount_permission.go — 割引フィールド権限チェック。
// medicalrecord / billing が PermissionChecker を注入して呼ぶ。
package httpapi

import (
	"math"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionChecker は (resource, action) の RBAC 判定 closure（composition root が注入）。
type PermissionChecker func(c *gin.Context, resource, action string) bool

const discountFloatEpsilon = 0.0001

// FloatEquals は割引率比較用の epsilon 比較。
func FloatEquals(a, b float64) bool {
	return math.Abs(a-b) < discountFloatEpsilon
}

// RequireDiscountEditFloat は discount_rate 等 float フィールドの編集権限を新旧値が異なる場合にのみ要求する。
func RequireDiscountEditFloat(c *gin.Context, hasPermission PermissionChecker, newVal *float64, oldVal float64) error {
	if newVal == nil {
		return nil
	}
	if FloatEquals(*newVal, oldVal) {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "edit") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
}

// RequireDiscountEditInt は discount_amount 等 int64 フィールドの編集権限を新旧値が異なる場合にのみ要求する。
func RequireDiscountEditInt(c *gin.Context, hasPermission PermissionChecker, newVal *int64, oldVal int64) error {
	if newVal == nil {
		return nil
	}
	if *newVal == oldVal {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "edit") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
}

// RequireDiscountCreateFloat は新規作成時にゼロ以外の discount_rate を指定する場合に create 権限を要求する。
func RequireDiscountCreateFloat(c *gin.Context, hasPermission PermissionChecker, val float64) error {
	if FloatEquals(val, 0) {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "create") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの作成権限がありません")
}

// RequireDiscountCreateInt は新規作成時にゼロ以外の discount_amount を指定する場合に create 権限を要求する。
func RequireDiscountCreateInt(c *gin.Context, hasPermission PermissionChecker, val int64) error {
	if val == 0 {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "create") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの作成権限がありません")
}

// RequireDiscountEditExact は discount_rate 等の exact 数値フィールドの編集権限を、新旧値が数値として異なる場合にのみ要求する。
// 比較は model.ExactNumeric.Equal の厳密比較で epsilon を使わない（10 と 10.00 は同値、0.1 と 0.10001 は別値）。
// 比較できない値（NaN など）は変更とみなして権限を要求する（fail-closed）。未移行の float64 caller は RequireDiscountEditFloat を使う。
func RequireDiscountEditExact(c *gin.Context, hasPermission PermissionChecker, newVal *model.ExactNumeric, oldVal model.ExactNumeric) error {
	if newVal == nil {
		return nil
	}
	if newVal.Equal(oldVal) {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "edit") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
}

// RequireDiscountCreateExact は新規作成時にゼロ以外の exact 数値を指定する場合に create 権限を要求する。
// 未指定は呼出し側が 0 として渡す。NULL・NaN など 0 と確認できない値は権限を要求する（fail-closed）。
// 未移行の float64 caller は RequireDiscountCreateFloat を使う。
func RequireDiscountCreateExact(c *gin.Context, hasPermission PermissionChecker, val model.ExactNumeric) error {
	if val.IsZero() {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "create") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの作成権限がありません")
}
