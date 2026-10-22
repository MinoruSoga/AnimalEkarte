package medicalrecord

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-447: 保存値（ExactNumeric）を float を経由せず canonical decimal 文字列で返す（末尾0は除く）。
func TestToMedicineResponse_ExactDecimalWire(t *testing.T) {
	sortOrder := 3
	m := &model.Medicine{
		ID:              1,
		Name:            "メロキシカム",
		DefaultQuantity: exactNumPtr("1.50"),
		TaxRate:         *exactNumPtr("0.10"),
		Strength:        exactNumPtr("0.0005"),
		SortOrder:       &sortOrder,
	}

	out, err := json.Marshal(toMedicineResponse(m))
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	assert.Equal(t, "1.5", got["default_quantity"])
	assert.Equal(t, "0.1", got["tax_rate"])
	assert.Equal(t, "0.0005", got["strength"])
	assert.Equal(t, float64(3), got["sort_order"])
}

// EMR-447: default_quantity の NULL は null、strength の NULL は既存どおり省略、明示 0 は "0"。
func TestToMedicineResponse_NullAndZero(t *testing.T) {
	m := &model.Medicine{ID: 2, Name: "未設定", TaxRate: *exactNumPtr("0")}

	out, err := json.Marshal(toMedicineResponse(m))
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	v, ok := got["default_quantity"]
	assert.True(t, ok, "default_quantity key is present")
	assert.Nil(t, v)
	assert.Equal(t, "0", got["tax_rate"])
	_, hasStrength := got["strength"]
	assert.False(t, hasStrength)
	assert.Equal(t, float64(0), got["sort_order"])
}
