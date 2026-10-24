package medicalrecord

// exact_numeric_bridge.go — EMR-505: float64 の request DTO・既存の float64 計算式と
// model.ExactNumeric 列の境界変換。
//
// 保存値（NUMERIC 列）は ExactNumeric のまま扱い、float64 を要する箇所との境界だけで変換する。
//   - float64 -> ExactNumeric は最短10進表現経由（2進誤差を取り込まない）。NaN・±Inf は 400。
//   - ExactNumeric -> float64 は計算入力の境界のみ。NULL・NaN・±Inf・float64 に収まらない値は
//     error とし、0 や既定値へ補完しない（nil は「不明」）。
//   - 等値・大小の判定は Float64() ではなく ExactNumeric.Equal / Cmp の厳密比較を使う。

import (
	"math"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// exactNumericFromFloat は float64 の request 値を ExactNumeric へ変換する。NaN・±Inf は 400（field 名付き）。
func exactNumericFromFloat(v float64, field string) (model.ExactNumeric, error) {
	n, err := model.ExactNumericFromFloat64(v)
	if err != nil {
		return model.ExactNumeric{}, apperrors.WrapInvalidInput(field + " は有限な数値で指定してください")
	}
	return n, nil
}

// exactNumericPtrFromFloatPtr は任意項目の変換。nil（未指定・上限なし）は nil のまま返し、0 へ補完しない。
func exactNumericPtrFromFloatPtr(v *float64, field string) (*model.ExactNumeric, error) {
	if v == nil {
		return nil, nil
	}
	n, err := exactNumericFromFloat(*v, field)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// exactNumericToFloat64 は保存値を float64 の計算入力へ変換する境界。
// Float64 は NULL を 0、変換不能を NaN で返すため、先に有限性を確認する。
// NULL・NaN・±Inf・float64 に収まらない値は error とし、黙って別の値にしない。
func exactNumericToFloat64(n model.ExactNumeric, field string) (float64, error) {
	if !n.IsFinite() {
		return 0, apperrors.WrapInternalServerError(field + " の保存値が有限な数値ではありません")
	}
	f := n.Float64()
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, apperrors.WrapInternalServerError(field + " の保存値を数値として扱えません")
	}
	return f, nil
}

// exactNumericPtrToFloat64Ptr は任意項目の変換。nil（上限なし等）は nil のまま返し、値がある場合のみ検証する。
func exactNumericPtrToFloat64Ptr(n *model.ExactNumeric, field string) (*float64, error) {
	if n == nil {
		return nil, nil
	}
	f, err := exactNumericToFloat64(*n, field)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// exactNumericRequiredFloat64 は必須項目の nullable な保存値を float64 へ変換する。
// nil（不明）は 0 や既定値へ補完せず 400 とする。
func exactNumericRequiredFloat64(n *model.ExactNumeric, field string) (float64, error) {
	if n == nil {
		return 0, apperrors.WrapInvalidInput(field + " が未設定です。値を指定してください")
	}
	return exactNumericToFloat64(*n, field)
}

// exactNumericOrNull は nullable な保存値を値渡しの ExactNumeric にする。
// nil は NULL（Valid=false）のまま渡し、0 へ補完しない。
func exactNumericOrNull(n *model.ExactNumeric) model.ExactNumeric {
	if n == nil {
		return model.ExactNumeric{}
	}
	return *n
}

// exactNumericEqualsFloat は保存値と float64 の request 値が数値として等しいかを厳密に返す（epsilon なし）。
// 保存値が NULL（nil）・非有限なら等しいとみなさない（fail-closed）。request 値が NaN・±Inf なら 400。
func exactNumericEqualsFloat(stored *model.ExactNumeric, requested float64, field string) (bool, error) {
	exact, err := exactNumericFromFloat(requested, field)
	if err != nil {
		return false, err
	}
	return exact.Equal(exactNumericOrNull(stored)), nil
}

// exactNumericIsPositive は有限で 0 より大きいことを返す。NULL・NaN・±Inf は比較不能として false。
func exactNumericIsPositive(n model.ExactNumeric) bool {
	cmp, ok := n.Cmp(model.ExactNumericFromInt64(0))
	return ok && cmp > 0
}
