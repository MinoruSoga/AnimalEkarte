package inventory

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- テストヘルパー（EMR-530 exact decimal 契約fixture） ---

func testTaxRateOf(intValue int64, exp int32) model.ExactNumeric {
	return testExactOf(intValue, exp)
}

// testExactOf は inventory テスト共通の ExactNumeric 構築（campaign DiscountValue 等も含む）。
func testExactOf(intValue int64, exp int32) model.ExactNumeric {
	return model.ExactNumeric{pgtype.Numeric{Int: big.NewInt(intValue), Exp: exp, Valid: true}}
}

func mustTaxRate(t *testing.T, text string) model.ExactNumeric {
	t.Helper()
	value, err := parseMerchandiseTaxRate(text)
	require.NoError(t, err)
	return value
}

// equalTaxRate は Exp 正規化なしの数値同値比較（float不使用）。
func equalTaxRate(a, b model.ExactNumeric) bool {
	return exactTaxRateRat(a).Cmp(exactTaxRateRat(b)) == 0
}

// --- wire 契約fixture: tax_rate の exact decimal 文字列（EMR-530 ②完了条件） ---

func TestParseMerchandiseTaxRate_Contract(t *testing.T) {
	ok := []struct {
		text string
		want string
	}{
		{"0.10", "0.10"},
		{"0.1", "0.1"},
		{"0", "0"},
		{"1", "1"},
		{"0.08", "0.08"},
		// 税率列に固定 scale は無いため精度をそのまま保持する
		{"0.08333", "0.08333"},
	}
	for _, tc := range ok {
		got, err := parseMerchandiseTaxRate(tc.text)
		require.NoError(t, err, tc.text)
		assert.True(t, equalTaxRate(mustTaxRate(t, tc.want), got), "%s -> %v", tc.text, got)
	}
	ng := []string{
		"-0.1",  // 符号
		"1e-1",  // 指数
		"NaN",   // 非数
		"Infinity",
		"10.",   // 小数部空
		"01",    // 先頭0
		"",      // 空
		"0.1.2", // 二重ドット
	}
	for _, text := range ng {
		_, err := parseMerchandiseTaxRate(text)
		assert.Error(t, err, text)
	}
}

func TestCreateMerchandiseItemRequest_TaxRateContract(t *testing.T) {
	t.Run("exact decimal string is accepted", func(t *testing.T) {
		var req createMerchandiseItemRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"n","category":"food","tax_type":"excluded","tax_rate":"0.10"}`), &req))
		input := req.toServiceInput()
		require.NotNil(t, input.TaxRate)
		assert.True(t, equalTaxRate(mustTaxRate(t, "0.10"), *input.TaxRate))
	})
	t.Run("json number is rejected as legacy float", func(t *testing.T) {
		var req createMerchandiseItemRequest
		assert.Error(t, json.Unmarshal([]byte(`{"name":"n","category":"food","tax_type":"excluded","tax_rate":0.1}`), &req))
	})
	t.Run("omitted falls back to default via nil", func(t *testing.T) {
		var req createMerchandiseItemRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"n","category":"food","tax_type":"excluded"}`), &req))
		assert.Nil(t, req.toServiceInput().TaxRate)
	})
	t.Run("null falls back to default via nil", func(t *testing.T) {
		var req createMerchandiseItemRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"n","category":"food","tax_type":"excluded","tax_rate":null}`), &req))
		assert.Nil(t, req.toServiceInput().TaxRate)
	})
	t.Run("explicit zero is kept as zero", func(t *testing.T) {
		var req createMerchandiseItemRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"n","category":"food","tax_type":"excluded","tax_rate":"0"}`), &req))
		input := req.toServiceInput()
		require.NotNil(t, input.TaxRate)
		assert.True(t, equalTaxRate(mustTaxRate(t, "0"), *input.TaxRate))
	})
}

func TestUpdateMerchandiseItemRequest_TaxRateContract(t *testing.T) {
	t.Run("omitted keeps current value", func(t *testing.T) {
		var req updateMerchandiseItemRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"n"}`), &req))
		assert.Nil(t, req.toServiceInput().TaxRate)
	})
	t.Run("null keeps current value", func(t *testing.T) {
		var req updateMerchandiseItemRequest
		require.NoError(t, json.Unmarshal([]byte(`{"tax_rate":null}`), &req))
		assert.Nil(t, req.toServiceInput().TaxRate)
	})
	t.Run("explicit zero updates to zero", func(t *testing.T) {
		var req updateMerchandiseItemRequest
		require.NoError(t, json.Unmarshal([]byte(`{"tax_rate":"0"}`), &req))
		input := req.toServiceInput()
		require.NotNil(t, input.TaxRate)
		assert.True(t, equalTaxRate(mustTaxRate(t, "0"), *input.TaxRate))
	})
	t.Run("json number is rejected as legacy float", func(t *testing.T) {
		var req updateMerchandiseItemRequest
		assert.Error(t, json.Unmarshal([]byte(`{"tax_rate":0.08}`), &req))
	})
}

// --- 業務範囲fixture: 0..1 の exact 比較 ---

func TestValidateMerchandiseTaxRate_ExactBoundaries(t *testing.T) {
	assert.NoError(t, validateMerchandiseTaxRate(mustTaxRate(t, "0")))
	assert.NoError(t, validateMerchandiseTaxRate(mustTaxRate(t, "0.10")))
	assert.NoError(t, validateMerchandiseTaxRate(mustTaxRate(t, "1")))
	assert.NoError(t, validateMerchandiseTaxRate(mustTaxRate(t, "0.08333")))
	assert.Error(t, validateMerchandiseTaxRate(mustTaxRate(t, "1.0001")))
}

// --- 応答fixture: 保存値の十進文字列 ---

func TestMerchandiseTaxRateString_Response(t *testing.T) {
	assert.Equal(t, "0.10", merchandiseTaxRateString(testTaxRateOf(10, -2)))
	assert.Equal(t, "0", merchandiseTaxRateString(testTaxRateOf(0, 0)))
	assert.Equal(t, "0.08", merchandiseTaxRateString(testTaxRateOf(8, -2)))
}
