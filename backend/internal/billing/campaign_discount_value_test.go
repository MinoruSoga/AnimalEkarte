package billing

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

func mustCampaignDiscountValue(t *testing.T, text string) campaignDiscountValueDTO {
	t.Helper()
	value, err := parseCampaignDiscountValue(text)
	require.NoError(t, err)
	return campaignDiscountValueDTO{set: true, value: value}
}

// exactNumericOf は *testing.T を持たないhelperから使う ExactNumeric 構築（テスト専用）。
func exactNumericOf(intValue int64, exp int32) model.ExactNumeric {
	return model.ExactNumeric{pgtype.Numeric{Int: big.NewInt(intValue), Exp: exp, Valid: true}}
}

func campaignExactValue(t *testing.T, text string) model.ExactNumeric {
	t.Helper()
	value, err := parseCampaignDiscountValue(text)
	require.NoError(t, err)
	return value
}

// equalExactNumeric は Exp 正規化なしの数値同値比較（float不使用）。
func equalExactNumeric(a, b model.ExactNumeric) bool {
	return exactNumericRat(a).Cmp(exactNumericRat(b)) == 0
}

func decodeCreateCampaignRequest(t *testing.T, body string) (*createCampaignRequest, error) {
	t.Helper()
	var req createCampaignRequest
	err := json.Unmarshal([]byte(body), &req)
	return &req, err
}

func decodeUpdateCampaignRequest(t *testing.T, body string) (*updateCampaignRequest, error) {
	t.Helper()
	var req updateCampaignRequest
	err := json.Unmarshal([]byte(body), &req)
	return &req, err
}

// --- wire 契約fixture: create（EMR-530 ②完了条件） ---

func TestCreateCampaignRequest_DiscountValueContract(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    string // 期待値を10進文字列で表す。errWant 時は不使用
		errWant bool
	}{
		{name: "exact decimal string", body: `{"discount_value":"10.5"}`, want: "10.5"},
		{name: "integer string", body: `{"discount_value":"500"}`, want: "500"},
		{name: "explicit zero is kept", body: `{"discount_value":"0"}`, want: "0"},
		{name: "omitted falls back to zero", body: `{}`, want: "0"},
		{name: "null falls back to zero", body: `{"discount_value":null}`, want: "0"},
		{name: "fraction beyond scale is rejected without rounding", body: `{"discount_value":"10.555"}`, errWant: true},
		{name: "11 integer digits is rejected", body: `{"discount_value":"12345678901"}`, errWant: true},
		{name: "negative is rejected", body: `{"discount_value":"-5"}`, errWant: true},
		{name: "exponent is rejected", body: `{"discount_value":"1e2"}`, errWant: true},
		{name: "trailing dot is rejected", body: `{"discount_value":"10."}`, errWant: true},
		{name: "leading zero is rejected", body: `{"discount_value":"010"}`, errWant: true},
		{name: "json number is rejected as legacy float", body: `{"discount_value":10}`, errWant: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := decodeCreateCampaignRequest(t, tc.body)
			if tc.errWant {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			input, convErr := req.toServiceInput()
			require.NoError(t, convErr)
			assert.True(t, equalExactNumeric(campaignExactValue(t, tc.want), input.DiscountValue))
		})
	}
}

// --- wire 契約fixture: update（省略保持・明示0） ---

func TestUpdateCampaignRequest_DiscountValueContract(t *testing.T) {
	t.Run("omitted keeps current value", func(t *testing.T) {
		req, err := decodeUpdateCampaignRequest(t, `{"name":"update"}`)
		require.NoError(t, err)
		input, convErr := req.toServiceInput()
		require.NoError(t, convErr)
		assert.Nil(t, input.DiscountValue)
	})
	t.Run("null keeps current value", func(t *testing.T) {
		req, err := decodeUpdateCampaignRequest(t, `{"discount_value":null}`)
		require.NoError(t, err)
		input, convErr := req.toServiceInput()
		require.NoError(t, convErr)
		assert.Nil(t, input.DiscountValue)
	})
	t.Run("explicit zero updates to zero", func(t *testing.T) {
		req, err := decodeUpdateCampaignRequest(t, `{"discount_value":"0"}`)
		require.NoError(t, err)
		input, convErr := req.toServiceInput()
		require.NoError(t, convErr)
		require.NotNil(t, input.DiscountValue)
		assert.True(t, equalExactNumeric(campaignExactValue(t, "0"), *input.DiscountValue))
	})
	t.Run("json number is rejected as legacy float", func(t *testing.T) {
		_, err := decodeUpdateCampaignRequest(t, `{"discount_value":5}`)
		assert.Error(t, err)
	})
}

// --- 業務範囲fixture: validateCampaignDiscount（exact 比較・float不使用） ---

func TestValidateCampaignDiscount_ExactBoundaries(t *testing.T) {
	ok := []struct {
		dt   model.CampaignDiscountType
		text string
	}{
		{model.CampaignDiscountTypeRate, "0"},
		{model.CampaignDiscountTypeRate, "100"},
		{model.CampaignDiscountTypeRate, "99.99"},
		{model.CampaignDiscountTypeAmount, "0"},
		{model.CampaignDiscountTypeAmount, "12345678.99"},
	}
	for _, tc := range ok {
		assert.NoError(t, validateCampaignDiscount(tc.dt, campaignExactValue(t, tc.text)), "%s %s", tc.dt, tc.text)
	}
	ng := []struct {
		dt   model.CampaignDiscountType
		text string
	}{
		{model.CampaignDiscountTypeRate, "100.01"},
		{model.CampaignDiscountTypeAmount, "-0.01"},
	}
	for _, tc := range ng {
		assert.Error(t, validateCampaignDiscount(tc.dt, campaignExactValue(t, tc.text)), "%s %s", tc.dt, tc.text)
	}
}

// --- 応答fixture: 保存値の十進文字列（numeric(12,2) の保存表記をそのまま返す） ---

func TestExactNumericString_Response(t *testing.T) {
	assert.Equal(t, "10.50", exactNumericString(model.ExactNumeric{pgtype.Numeric{Int: big.NewInt(1050), Exp: -2, Valid: true}}))
	assert.Equal(t, "0", exactNumericString(model.ExactNumeric{pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}}))
	assert.Equal(t, "500.00", exactNumericString(model.ExactNumeric{pgtype.Numeric{Int: big.NewInt(50000), Exp: -2, Valid: true}}))
}
