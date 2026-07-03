package handler

import (
	"math"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BUG-372: 割引フィールド専用の権限チェックヘルパー群。
// 飼主 / 治療 / 入院 / 見積 / 会計 の各ハンドラで割引フィールド変更時に呼び出す。
//
// 設計方針（AC-10 準拠）:
//   - 未指定（*ptr == nil）: 権限不要
//   - ゼロ値 & 既存値もゼロ値: 権限不要（再送耐性）
//   - 既存値と一致: 権限不要（値未変化）
//   - 上記以外: "discount:edit" または "discount:create" 権限を要求
//
// 権限なしの場合は apperrors.ErrForbidden を返す。呼び出し元は RespondError に委ねる。

const discountFloatEpsilon = 0.0001

// floatEquals は浮動小数の実用上の等価判定。
func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < discountFloatEpsilon
}

// requireDiscountEditFloat は discount_rate 等 float64 フィールドの編集権限を
// 新旧値が実質的に異なる場合にのみ要求する。
func (h *Handler) requireDiscountEditFloat(c *gin.Context, newVal *float64, oldVal float64) error {
	if newVal == nil {
		return nil
	}
	if floatEquals(*newVal, oldVal) {
		return nil
	}
	if h.hasPermission(c, string(model.ResourceDiscount), "edit") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
}

// requireDiscountEditInt は discount_amount 等 int64 フィールドの編集権限を
// 新旧値が異なる場合にのみ要求する。
func (h *Handler) requireDiscountEditInt(c *gin.Context, newVal *int64, oldVal int64) error {
	if newVal == nil {
		return nil
	}
	if *newVal == oldVal {
		return nil
	}
	if h.hasPermission(c, string(model.ResourceDiscount), "edit") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
}

// requireDiscountCreateFloat は新規作成時にゼロ以外の discount_rate を指定する場合に
// "discount:create" 権限を要求する。ゼロ値は権限不要。
func (h *Handler) requireDiscountCreateFloat(c *gin.Context, val float64) error {
	if floatEquals(val, 0) {
		return nil
	}
	if h.hasPermission(c, string(model.ResourceDiscount), "create") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの作成権限がありません")
}

// requireDiscountCreateInt は新規作成時にゼロ以外の discount_amount を指定する場合に
// "discount:create" 権限を要求する。ゼロ値は権限不要。
func (h *Handler) requireDiscountCreateInt(c *gin.Context, val int64) error {
	if val == 0 {
		return nil
	}
	if h.hasPermission(c, string(model.ResourceDiscount), "create") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの作成権限がありません")
}
