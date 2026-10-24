package model

import (
	"errors"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// exactNumericMaxExp は比較・正規化で許す指数の絶対値の上限。
// DB の NUMERIC の実値ではなく、巨大な指数による過大なメモリ確保を防ぐ防御値で、
// 超過した値は比較・変換できない値として扱う。
const exactNumericMaxExp = 4096

// exactNumericMaxTokenLen は ParseExactNumeric が受理する token の最大バイト長（入力長の防御値）。
const exactNumericMaxTokenLen = 128

// ErrExactNumericInvalid は NULL・NaN・Infinity・不正な10進 token・上限超過のいずれかを示す。
var ErrExactNumericInvalid = errors.New("exact numeric: 有限な10進数ではありません")

// ErrExactNumericNotInteger は指定の桁で整数に表せず、丸めが必要になることを示す。
var ErrExactNumericNotInteger = errors.New("exact numeric: 指定の桁では整数になりません")

// ErrExactNumericOutOfRange は int64 に収まらないことを示す。
var ErrExactNumericOutOfRange = errors.New("exact numeric: int64 の範囲外です")

// ExactNumericFromInt64 は整数を Exp=0 の ExactNumeric へ無損失で変換する。
func ExactNumericFromInt64(v int64) ExactNumeric {
	return ExactNumericFromScaledInt64(v, 0)
}

// ExactNumericFromScaledInt64 は v を 10^scale で割った値（v * 10^-scale）を表す ExactNumeric を返す。
// scale は 0 以上を想定する。例: (10, 2) は 0.10、(1250, 2) は 12.50。丸めは行わない。
func ExactNumericFromScaledInt64(v int64, scale int32) ExactNumeric {
	return ExactNumeric{Numeric: pgtype.Numeric{Int: big.NewInt(v), Exp: -scale, Valid: true}}
}

// ParseExactNumeric は wire の10進 token を厳密に解釈する。
// 受理するのは「省略可能な先頭 '-'、先頭0なしの整数部、'.' に続く1桁以上の小数部（省略可）」だけで、
// '+'・指数表記・空白・NaN・Infinity・桁区切りは ErrExactNumericInvalid とする。
// 小数部の桁数（scale）は丸めず保持する（"10.50" は Int=1050, Exp=-2）。値の範囲検査は呼出し側が行う。
func ParseExactNumeric(token string) (ExactNumeric, error) {
	if token == "" || len(token) > exactNumericMaxTokenLen {
		return ExactNumeric{}, ErrExactNumericInvalid
	}
	body, negative := strings.CutPrefix(token, "-")
	intPart, fracPart, hasFrac := strings.Cut(body, ".")
	if !exactNumericIntegerPart(intPart) || (hasFrac && !exactNumericDigitsOnly(fracPart)) {
		return ExactNumeric{}, ErrExactNumericInvalid
	}
	unscaled, ok := new(big.Int).SetString(intPart+fracPart, 10)
	if !ok {
		return ExactNumeric{}, ErrExactNumericInvalid
	}
	if negative {
		unscaled.Neg(unscaled)
	}
	return ExactNumeric{Numeric: pgtype.Numeric{Int: unscaled, Exp: -int32(len(fracPart)), Valid: true}}, nil
}

// exactNumericDigitsOnly は s が1文字以上の ASCII 数字だけで構成されることを返す。
func exactNumericDigitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// exactNumericIntegerPart は整数部（"0" または先頭0なしの正整数）を検査する。
func exactNumericIntegerPart(s string) bool {
	return exactNumericDigitsOnly(s) && (s == "0" || s[0] != '0')
}

// ExactNumericFromFloat64 は未移行の float64 request DTO を exact へ橋渡しする暫定の変換。
// float64 の最短 round-trip 10進表現（0.1 は "0.1"）を ParseExactNumeric で解釈するため、
// 2進誤差は取り込まない（medicalrecord の hospitalizationPlanExactTaxRate と同じ方針）。
// NaN・±Inf と、token 長の上限を超える値は ErrExactNumericInvalid。新規の wire 受理には使わない。
func ExactNumericFromFloat64(v float64) (ExactNumeric, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return ExactNumeric{}, ErrExactNumericInvalid
	}
	return ParseExactNumeric(strconv.FormatFloat(v, 'f', -1, 64))
}

// IsFinite は NULL・NaN・±Infinity のいずれでもない（数値として比較できる）ことを返す。
func (n ExactNumeric) IsFinite() bool {
	return n.Valid && !n.NaN && n.InfinityModifier == pgtype.Finite
}

// exactNumericPow10 は 10^exp を返す。
func exactNumericPow10(exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
}

// rat は有限値を無損失の big.Rat にする。NULL・非有限・指数が上限超の値は ok=false。
func (n ExactNumeric) rat() (*big.Rat, bool) {
	if !n.IsFinite() || n.Exp > exactNumericMaxExp || n.Exp < -exactNumericMaxExp {
		return nil, false
	}
	unscaled := new(big.Int)
	if n.Int != nil {
		unscaled.Set(n.Int)
	}
	if n.Exp >= 0 {
		unscaled.Mul(unscaled, exactNumericPow10(int64(n.Exp)))
		return new(big.Rat).SetInt(unscaled), true
	}
	return new(big.Rat).SetFrac(unscaled, exactNumericPow10(-int64(n.Exp))), true
}

// Cmp は数値として比較し、-1・0・1 を返す。どちらかが比較不能（NULL・非有限）なら ok=false。
func (n ExactNumeric) Cmp(other ExactNumeric) (int, bool) {
	left, leftOK := n.rat()
	right, rightOK := other.rat()
	if !leftOK || !rightOK {
		return 0, false
	}
	return left.Cmp(right), true
}

// Equal は数値としての等値を返す（10 と 10.00 は等しい。文字列や scale の一致は見ない）。
// NULL 同士は等しく、NULL と値は異なる。NaN・±Infinity・指数が上限超の値は常に不等とし、
// 権限判定などで「変更なし」と誤認しない側（fail-closed）に倒す。
func (n ExactNumeric) Equal(other ExactNumeric) bool {
	if !n.Valid && !other.Valid {
		return true
	}
	cmp, ok := n.Cmp(other)
	return ok && cmp == 0
}

// IsZero は有限で 0 と等しいことを返す。NULL・非有限は 0 とみなさない。
func (n ExactNumeric) IsZero() bool {
	value, ok := n.rat()
	return ok && value.Sign() == 0
}

// ScaledInt64 は値に 10^scale を掛けた結果を int64 で返す。丸めは行わない。
// 掛けた結果が整数にならなければ ErrExactNumericNotInteger、int64 に収まらなければ
// ErrExactNumericOutOfRange、NULL・非有限・scale が負または上限超なら ErrExactNumericInvalid。
// 例: 0.10 を scale=2 で 10、12.50 を scale=2 で 1250。
func (n ExactNumeric) ScaledInt64(scale int32) (int64, error) {
	if scale < 0 || scale > exactNumericMaxExp {
		return 0, ErrExactNumericInvalid
	}
	value, ok := n.rat()
	if !ok {
		return 0, ErrExactNumericInvalid
	}
	scaled := new(big.Rat).Mul(value, new(big.Rat).SetInt(exactNumericPow10(int64(scale))))
	if !scaled.IsInt() {
		return 0, ErrExactNumericNotInteger
	}
	integer := scaled.Num()
	if !integer.IsInt64() {
		return 0, ErrExactNumericOutOfRange
	}
	return integer.Int64(), nil
}

// Int64 は値が int64 の整数に正確に等しいときだけその値を返す（ScaledInt64(0)）。
func (n ExactNumeric) Int64() (int64, error) {
	return n.ScaledInt64(0)
}

// PercentInt64 は割合（0.10）を整数パーセント（10）へ厳密に変換する（ScaledInt64(2)）。
// 0.105 のように丸めが必要な値は ErrExactNumericNotInteger を返し、丸め規則は呼出し側が決める。
func (n ExactNumeric) PercentInt64() (int64, error) {
	return n.ScaledInt64(2)
}

// CanonicalString は末尾の0を除いた正規の10進文字列を返す（10.50 は "10.5"、100 は "100"、0.00 は "0"）。
// 指数表記は使わず、丸めもしない。wire へ10進文字列を出す将来の adapter 用で、現行の JSON number
// 出力（ExactNumeric の MarshalJSON）は変更しない。NULL・非有限・指数が上限超は ErrExactNumericInvalid。
func (n ExactNumeric) CanonicalString() (string, error) {
	if !n.IsFinite() || n.Exp > exactNumericMaxExp || n.Exp < -exactNumericMaxExp {
		return "", ErrExactNumericInvalid
	}
	if n.Int == nil || n.Int.Sign() == 0 {
		return "0", nil
	}
	digits := new(big.Int).Abs(n.Int).String()
	trimmed := strings.TrimRight(digits, "0")
	exp := int(n.Exp) + len(digits) - len(trimmed)
	if exp > exactNumericMaxExp {
		return "", ErrExactNumericInvalid
	}
	var sb strings.Builder
	if n.Int.Sign() < 0 {
		sb.WriteByte('-')
	}
	switch {
	case exp >= 0:
		sb.WriteString(trimmed)
		sb.WriteString(strings.Repeat("0", exp))
	case len(trimmed) > -exp:
		point := len(trimmed) + exp
		sb.WriteString(trimmed[:point])
		sb.WriteByte('.')
		sb.WriteString(trimmed[point:])
	default:
		sb.WriteString("0.")
		zeros := -exp - len(trimmed)
		sb.WriteString(strings.Repeat("0", zeros))
		sb.WriteString(trimmed)
	}
	return sb.String(), nil
}

// Float64 は float64 の JSON number を出力する未移行 response 向けの表示用変換。
// 2進誤差を伴うため計算・比較・保存には使わない（Equal・Cmp・ScaledInt64 を使う）。
// NULL は 0、変換できない値は NaN・±Inf を返し、JSON 出力時に失敗として表面化する。
func (n ExactNumeric) Float64() float64 {
	value, err := n.Float64Value()
	if err != nil {
		return math.NaN()
	}
	return value.Float64
}
