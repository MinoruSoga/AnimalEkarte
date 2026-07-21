package medicalrecord

import (

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// BUG-372: 割引フィールド専用の権限チェック（treatment / treatment-plan 共有・BE9-2D ⑥で
// package-level 化。原本 internal/handler/discount_permission.go は accounting 等の残存
// ハンドラが使い続ける）。設計方針 (AC-10): 未指定 nil / ゼロ値&既存ゼロ / 値不変は権限不要、
// それ以外は discount:create / discount:edit を要求。

// BE9-2C B②: 実装は httpapi へ昇格。以下は既存呼び出し面互換の delegate。

// discountFloatEpsilon は httpapi 側 unexported 定数と同値（test の境界値検証用に残置）。
const discountFloatEpsilon = 0.0001

func floatEquals(a, b float64) bool { return httpapi.FloatEquals(a, b) }

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
