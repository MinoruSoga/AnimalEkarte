package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DEC-40: category_breakdown JSONB の UnclassifiedOtherCount 後方互換。
func TestCategoryBreakdownSchema_UnclassifiedOtherCountBackwardCompat(t *testing.T) {
	t.Run("旧 snapshot（フィールド欠落）は nil", func(t *testing.T) {
		raw := `{"categories":{"other":{"cash":500}},"tax_breakdown":{"standard":{"taxable_amount":0,"tax_amount":0},"reduced":{"taxable_amount":0,"tax_amount":0}}}`
		var schema CategoryBreakdownSchema
		require.NoError(t, json.Unmarshal([]byte(raw), &schema))
		assert.Nil(t, schema.UnclassifiedOtherCount)
	})

	t.Run("新 snapshot は 0 件でもポインタで記録する", func(t *testing.T) {
		zero := int64(0)
		schema := CategoryBreakdownSchema{
			Categories:             map[string]map[string]int64{},
			UnclassifiedOtherCount: &zero,
		}
		b, err := json.Marshal(schema)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"unclassified_other_count":0`)

		var decoded CategoryBreakdownSchema
		require.NoError(t, json.Unmarshal(b, &decoded))
		require.NotNil(t, decoded.UnclassifiedOtherCount)
		assert.Equal(t, int64(0), *decoded.UnclassifiedOtherCount)
	})

	t.Run("新 snapshot の非ゼロ件数を復元する", func(t *testing.T) {
		raw := `{"categories":{},"tax_breakdown":{"standard":{"taxable_amount":0,"tax_amount":0},"reduced":{"taxable_amount":0,"tax_amount":0}},"unclassified_other_count":4}`
		var schema CategoryBreakdownSchema
		require.NoError(t, json.Unmarshal([]byte(raw), &schema))
		require.NotNil(t, schema.UnclassifiedOtherCount)
		assert.Equal(t, int64(4), *schema.UnclassifiedOtherCount)
	})
}
