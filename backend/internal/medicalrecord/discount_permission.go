package medicalrecord

import (
	"math"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BUG-372: 割引フィールド専用の権限チェック（treatment / treatment-plan 共有・BE9-2D ⑥で
// package-level 化。原本 internal/handler/discount_permission.go は accounting 等の残存
// ハンドラが使い続ける）。設計方針 (AC-10): 未指定 nil / ゼロ値&既存ゼロ / 値不変は権限不要、
// それ以外は discount:create / discount:edit を要求。

const discountFloatEpsilon = 0.0001

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < discountFloatEpsilon
}

func requireDiscountEditFloat(c *gin.Context, hasPermission PermissionChecker, newVal *float64, oldVal float64) error {
	if newVal == nil {
		return nil
	}
	if floatEquals(*newVal, oldVal) {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "edit") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
}

func requireDiscountEditInt(c *gin.Context, hasPermission PermissionChecker, newVal *int64, oldVal int64) error {
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

func requireDiscountCreateFloat(c *gin.Context, hasPermission PermissionChecker, val float64) error {
	if floatEquals(val, 0) {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "create") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの作成権限がありません")
}

func requireDiscountCreateInt(c *gin.Context, hasPermission PermissionChecker, val int64) error {
	if val == 0 {
		return nil
	}
	if hasPermission(c, string(model.ResourceDiscount), "create") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの作成権限がありません")
}
