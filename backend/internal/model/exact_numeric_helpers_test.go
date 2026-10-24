package model

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exactFromToken はテスト用に10進 token から ExactNumeric を作る（不正な literal はテストの誤り）。
func exactFromToken(t *testing.T, token string) ExactNumeric {
	t.Helper()
	value, err := ParseExactNumeric(token)
	require.NoError(t, err)
	return value
}

func TestParseExactNumeric_AcceptsStrictDecimalTokens(t *testing.T) {
	cases := []struct {
		token   string
		wantInt string
		wantExp int32
	}{
		{token: "0", wantInt: "0", wantExp: 0},
		{token: "10", wantInt: "10", wantExp: 0},
		{token: "10.50", wantInt: "1050", wantExp: -2},
		{token: "0.08", wantInt: "8", wantExp: -2},
		{token: "-3.25", wantInt: "-325", wantExp: -2},
		{token: "-0", wantInt: "0", wantExp: 0},
	}
	for _, tc := range cases {
		t.Run(tc.token, func(t *testing.T) {
			got, err := ParseExactNumeric(tc.token)
			require.NoError(t, err)
			assert.True(t, got.Valid)
			assert.Equal(t, tc.wantInt, got.Int.String())
			assert.Equal(t, tc.wantExp, got.Exp)
		})
	}
}

func TestParseExactNumeric_RejectsNonStrictTokens(t *testing.T) {
	tooLong := strings.Repeat("1", exactNumericMaxTokenLen+1)
	tokens := []string{
		"", "+1", "1e3", "1E3", ".5", "5.", "01", "00.5", "NaN", "Infinity",
		"-", "--1", " 1", "1 ", "1_000", "1,000", "1.2.3", "０",
		tooLong,
	}
	for _, token := range tokens {
		_, err := ParseExactNumeric(token)
		assert.ErrorIs(t, err, ErrExactNumericInvalid, "token=%q", token)
	}
}

func TestParseExactNumeric_AcceptsTokenAtLengthLimit(t *testing.T) {
	atLimit := "1" + strings.Repeat("0", exactNumericMaxTokenLen-1)
	_, err := ParseExactNumeric(atLimit)
	assert.NoError(t, err)
}

func TestExactNumericFromFloat64_UsesShortestDecimal(t *testing.T) {
	cases := []struct {
		value float64
		want  string
	}{
		{value: 0.1, want: "0.1"},
		{value: 12.5, want: "12.5"},
		{value: 5, want: "5"},
		{value: -0.05, want: "-0.05"},
		{value: math.Copysign(0, -1), want: "0"},
	}
	for _, tc := range cases {
		got, err := ExactNumericFromFloat64(tc.value)
		require.NoError(t, err)
		text, err := got.CanonicalString()
		require.NoError(t, err)
		assert.Equal(t, tc.want, text)
	}
}

func TestExactNumericFromFloat64_RejectsNonFiniteAndOversize(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), 1e300} {
		_, err := ExactNumericFromFloat64(value)
		assert.ErrorIs(t, err, ErrExactNumericInvalid)
	}
}

func TestExactNumeric_EqualComparesNumericValue(t *testing.T) {
	cases := []struct {
		name  string
		left  string
		right string
		want  bool
	}{
		{name: "scaleだけ違う整数", left: "10", right: "10.00", want: true},
		{name: "scaleだけ違う小数", left: "0.1", right: "0.10", want: true},
		{name: "epsilon内の差も別値", left: "0.1", right: "0.10001", want: false},
		{name: "符号違い", left: "1", right: "-1", want: false},
		{name: "正負のゼロ", left: "0", right: "-0", want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			left := exactFromToken(t, tc.left)
			right := exactFromToken(t, tc.right)
			assert.Equal(t, tc.want, left.Equal(right))
			assert.Equal(t, tc.want, right.Equal(left))
		})
	}
}

func TestExactNumeric_EqualHandlesNullAndNonFinite(t *testing.T) {
	null := ExactNumeric{}
	nan := ExactNumeric{Numeric: pgtype.Numeric{NaN: true, Valid: true}}
	inf := ExactNumeric{Numeric: pgtype.Numeric{InfinityModifier: pgtype.Infinity, Valid: true}}
	zero := exactFromToken(t, "0")
	assert.True(t, null.Equal(ExactNumeric{}))
	assert.False(t, null.Equal(zero))
	assert.False(t, zero.Equal(null))
	assert.False(t, nan.Equal(nan))
	assert.False(t, inf.Equal(inf))
	assert.False(t, nan.Equal(zero))
}

func TestExactNumeric_ExponentBeyondLimitIsNotComparable(t *testing.T) {
	const exponentBeyondLimit = exactNumericMaxExp + 1
	huge := ExactNumeric{Numeric: pgtype.Numeric{Int: big.NewInt(1), Exp: exponentBeyondLimit, Valid: true}}
	assert.False(t, huge.Equal(huge))
	assert.False(t, huge.IsZero())
	_, err := huge.CanonicalString()
	assert.ErrorIs(t, err, ErrExactNumericInvalid)
	_, err = huge.ScaledInt64(0)
	assert.ErrorIs(t, err, ErrExactNumericInvalid)
}

func TestExactNumeric_IsZero(t *testing.T) {
	assert.True(t, exactFromToken(t, "0").IsZero())
	assert.True(t, exactFromToken(t, "0.00").IsZero())
	assert.False(t, exactFromToken(t, "0.01").IsZero())
	assert.False(t, ExactNumeric{}.IsZero())
	assert.False(t, ExactNumeric{Numeric: pgtype.Numeric{NaN: true, Valid: true}}.IsZero())
	assert.True(t, ExactNumeric{Numeric: pgtype.Numeric{Valid: true}}.IsZero())
}

func TestExactNumeric_Cmp(t *testing.T) {
	got, ok := exactFromToken(t, "0.5").Cmp(exactFromToken(t, "0.25"))
	assert.True(t, ok)
	assert.Equal(t, 1, got)
	_, ok = exactFromToken(t, "1").Cmp(ExactNumeric{})
	assert.False(t, ok)
}

func TestExactNumeric_CanonicalString(t *testing.T) {
	cases := []struct {
		name  string
		value ExactNumeric
		want  string
	}{
		{name: "末尾の0を除く", value: exactFromToken(t, "10.50"), want: "10.5"},
		{name: "整数部の0は保つ", value: exactFromToken(t, "100"), want: "100"},
		{name: "ゼロは0", value: exactFromToken(t, "0.00"), want: "0"},
		{name: "負の小数", value: exactFromToken(t, "-0.050"), want: "-0.05"},
		{name: "先頭ゼロ付き小数", value: exactFromToken(t, "0.08"), want: "0.08"},
		{name: "正の指数", value: ExactNumeric{Numeric: pgtype.Numeric{Int: big.NewInt(12), Exp: 3, Valid: true}}, want: "12000"},
		{name: "Intがnilの有効値", value: ExactNumeric{Numeric: pgtype.Numeric{Valid: true}}, want: "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text, err := tc.value.CanonicalString()
			require.NoError(t, err)
			assert.Equal(t, tc.want, text)
		})
	}
}

func TestExactNumeric_CanonicalStringRejectsNullAndNonFinite(t *testing.T) {
	for _, value := range []ExactNumeric{{}, {Numeric: pgtype.Numeric{NaN: true, Valid: true}}} {
		_, err := value.CanonicalString()
		assert.ErrorIs(t, err, ErrExactNumericInvalid)
	}
}

func TestExactNumeric_PercentInt64(t *testing.T) {
	cases := []struct {
		token   string
		want    int64
		wantErr error
	}{
		{token: "0.10", want: 10},
		{token: "0.08", want: 8},
		{token: "0", want: 0},
		{token: "1", want: 100},
		{token: "0.105", wantErr: ErrExactNumericNotInteger},
	}
	for _, tc := range cases {
		t.Run(tc.token, func(t *testing.T) {
			got, err := exactFromToken(t, tc.token).PercentInt64()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestExactNumeric_Int64(t *testing.T) {
	got, err := exactFromToken(t, "1200.00").Int64()
	require.NoError(t, err)
	assert.Equal(t, int64(1200), got)

	_, err = exactFromToken(t, "1.5").Int64()
	assert.ErrorIs(t, err, ErrExactNumericNotInteger)

	_, err = exactFromToken(t, "9223372036854775808").Int64()
	assert.ErrorIs(t, err, ErrExactNumericOutOfRange)

	_, err = ExactNumeric{}.Int64()
	assert.ErrorIs(t, err, ErrExactNumericInvalid)

	_, err = exactFromToken(t, "1").ScaledInt64(-1)
	assert.ErrorIs(t, err, ErrExactNumericInvalid)
}

func TestExactNumericFromInt64_RoundTrips(t *testing.T) {
	for _, value := range []int64{0, 1, -1, 1234567, math.MaxInt64, math.MinInt64} {
		got, err := ExactNumericFromInt64(value).Int64()
		require.NoError(t, err)
		assert.Equal(t, value, got)
	}
}

func TestExactNumericFromScaledInt64(t *testing.T) {
	text, err := ExactNumericFromScaledInt64(1250, 2).CanonicalString()
	require.NoError(t, err)
	assert.Equal(t, "12.5", text)
	assert.True(t, ExactNumericFromScaledInt64(10, 2).Equal(exactFromToken(t, "0.1")))
}

func TestExactNumeric_Float64(t *testing.T) {
	assert.Equal(t, 0.1, exactFromToken(t, "0.10").Float64())
	assert.Equal(t, 0.0, ExactNumeric{}.Float64())
	assert.True(t, math.IsNaN(ExactNumeric{Numeric: pgtype.Numeric{NaN: true, Valid: true}}.Float64()))
}
