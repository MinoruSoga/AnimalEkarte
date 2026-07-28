package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceCheckupFieldResultsRequest_ToServiceInput(t *testing.T) {
	t.Run("nil results propagates as nil", func(t *testing.T) {
		req := replaceCheckupFieldResultsRequest{Results: nil}

		got := req.toServiceInput()

		assert.Nil(t, got)
	})

	t.Run("empty slice propagates as non-nil empty slice", func(t *testing.T) {
		req := replaceCheckupFieldResultsRequest{Results: []upsertCheckupFieldResultRequest{}}

		got := req.toServiceInput()

		require.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("maps all fields for each item", func(t *testing.T) {
		fieldID := uint64(10)
		valueNumber := 12.5
		valueBool := true
		req := replaceCheckupFieldResultsRequest{
			Results: []upsertCheckupFieldResultRequest{
				{
					CheckupTypeFieldID: &fieldID,
					ValueNumber:        &valueNumber,
					ValueText:          "text-value",
					ValueBool:          &valueBool,
					ValueList:          []string{"a", "b"},
				},
			},
		}

		got := req.toServiceInput()

		require.Len(t, got, 1)
		item := got[0]
		require.NotNil(t, item.CheckupTypeFieldID)
		assert.Equal(t, fieldID, *item.CheckupTypeFieldID)
		require.NotNil(t, item.ValueNumber)
		assert.InDelta(t, valueNumber, *item.ValueNumber, 0.0001)
		assert.Equal(t, "text-value", item.ValueText)
		require.NotNil(t, item.ValueBool)
		assert.True(t, *item.ValueBool)
		assert.Equal(t, []string{"a", "b"}, item.ValueList)
	})
}
