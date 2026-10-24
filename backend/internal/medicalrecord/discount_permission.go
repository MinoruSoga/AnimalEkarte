package medicalrecord

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// BUG-372: 割引フィールド専用の権限チェック（treatment / treatment-plan 共有）。
// 実装の正本は httpapi。ここは既存呼び出し面互換の delegate。

// discountFloatEpsilon は httpapi 側 unexported 定数と同値（test の境界値検証用に残置）。
const discountFloatEpsilon = 0.0001

func requireDiscountEditFloat(c *gin.Context, hasPermission PermissionChecker, newVal *float64, oldVal float64) error {
	return httpapi.RequireDiscountEditFloat(c, httpapi.PermissionChecker(hasPermission), newVal, oldVal)
}

// requireDiscountEditRate は float64 の request 値と保存済みの割引率（exact）を厳密比較し、
// 数値として異なる場合にのみ discount:edit を要求する（epsilon なし）。
// 保存値が NULL・非有限の場合は変更とみなして権限を要求する（fail-closed）。
func requireDiscountEditRate(c *gin.Context, hasPermission PermissionChecker, newVal *float64, oldVal *model.ExactNumeric) error {
	if newVal == nil {
		return nil
	}
	requested, err := exactNumericFromFloat(*newVal, "discount_rate")
	if err != nil {
		return err
	}
	return httpapi.RequireDiscountEditExact(c, httpapi.PermissionChecker(hasPermission), &requested, exactNumericOrNull(oldVal))
}

func requireDiscountEditInt(c *gin.Context, hasPermission PermissionChecker, newVal *int64, oldVal int64) error {
	return httpapi.RequireDiscountEditInt(c, httpapi.PermissionChecker(hasPermission), newVal, oldVal)
}

// requireDiscountEditStoredInt は保存値が NULL 許容の discount_amount 用。
// 保存値が NULL（未記録）の場合は新旧一致と確認できないため変更とみなし、discount:edit を要求する（fail-closed）。
func requireDiscountEditStoredInt(c *gin.Context, hasPermission PermissionChecker, newVal, oldVal *int64) error {
	if newVal == nil {
		return nil
	}
	if oldVal != nil {
		return requireDiscountEditInt(c, hasPermission, newVal, *oldVal)
	}
	if hasPermission(c, string(model.ResourceDiscount), "edit") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
}

func requireDiscountCreateFloat(c *gin.Context, hasPermission PermissionChecker, val float64) error {
	return httpapi.RequireDiscountCreateFloat(c, httpapi.PermissionChecker(hasPermission), val)
}

func requireDiscountCreateInt(c *gin.Context, hasPermission PermissionChecker, val int64) error {
	return httpapi.RequireDiscountCreateInt(c, httpapi.PermissionChecker(hasPermission), val)
}
