package medicalrecord

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// BUG-372: 割引フィールド専用の権限チェック（treatment / treatment-plan 共有）。
// 実装の正本は httpapi。ここは既存呼び出し面互換の delegate。

// discountFloatEpsilon は httpapi 側 unexported 定数と同値（test の境界値検証用に残置）。
const discountFloatEpsilon = 0.0001

func requireDiscountEditFloat(c *gin.Context, hasPermission PermissionChecker, newVal *float64, oldVal float64) error {
	return httpapi.RequireDiscountEditFloat(c, httpapi.PermissionChecker(hasPermission), newVal, oldVal)
}

func requireDiscountEditInt(c *gin.Context, hasPermission PermissionChecker, newVal *int64, oldVal int64) error {
	return httpapi.RequireDiscountEditInt(c, httpapi.PermissionChecker(hasPermission), newVal, oldVal)
}

func requireDiscountCreateFloat(c *gin.Context, hasPermission PermissionChecker, val float64) error {
	return httpapi.RequireDiscountCreateFloat(c, httpapi.PermissionChecker(hasPermission), val)
}

func requireDiscountCreateInt(c *gin.Context, hasPermission PermissionChecker, val int64) error {
	return httpapi.RequireDiscountCreateInt(c, httpapi.PermissionChecker(hasPermission), val)
}
