package handler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateOwnerRequest_DMPreferenceNullableJSON(t *testing.T) {
	t.Run("omitted means no update", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{}`), &req))

		input := req.toServiceInput()

		assert.Nil(t, input.DMPreference)
	})

	t.Run("null clears dm_preference", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{"dm_preference":null}`), &req))

		input := req.toServiceInput()

		require.NotNil(t, input.DMPreference)
		assert.Nil(t, *input.DMPreference)
	})

	t.Run("false is preserved", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{"dm_preference":false}`), &req))

		input := req.toServiceInput()

		require.NotNil(t, input.DMPreference)
		require.NotNil(t, *input.DMPreference)
		assert.False(t, **input.DMPreference)
	})
}
