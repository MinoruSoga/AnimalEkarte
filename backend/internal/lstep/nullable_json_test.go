package lstep

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-504 contract A: nullable jsonb columns (*json.RawMessage) keep NULL / empty / valid JSON apart.
func TestNullableJSONPtr(t *testing.T) {
	assert.Nil(t, nullableJSONPtr(nil), "no bytes -> NULL")
	assert.Nil(t, nullableJSONPtr([]byte{}), "zero-length bytes would be invalid jsonb -> NULL")

	empty := nullableJSONPtr([]byte(`[]`))
	require.NotNil(t, empty, "an empty JSON array is a value, not NULL")
	assert.JSONEq(t, `[]`, string(*empty))

	value := nullableJSONPtr([]byte(`["a","b"]`))
	require.NotNil(t, value)
	assert.JSONEq(t, `["a","b"]`, string(*value))
}

func TestNullableJSONBytes(t *testing.T) {
	assert.Nil(t, nullableJSONBytes(nil))
	raw := json.RawMessage(`["a"]`)
	assert.JSONEq(t, `["a"]`, string(nullableJSONBytes(&raw)))
}

func TestToLstepFriendAttributeResponse_TagsAndScenariosNullEmptyAndValue(t *testing.T) {
	now := time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)
	empty := json.RawMessage(`[]`)
	tags := json.RawMessage(`["tagA","tagB"]`)
	scenarios := json.RawMessage(`["s1"]`)

	t.Run("NULL tags and scenarios are omitted", func(t *testing.T) {
		raw, err := json.Marshal(toLstepFriendAttributeResponse(&model.LstepFriendAttributeSnapshot{LineUserID: "U1", SnapshotTakenAt: now}))
		require.NoError(t, err)
		assert.NotContains(t, string(raw), `"tags"`)
		assert.NotContains(t, string(raw), `"scenarios"`)
	})
	t.Run("empty arrays are returned as empty arrays", func(t *testing.T) {
		raw, err := json.Marshal(toLstepFriendAttributeResponse(&model.LstepFriendAttributeSnapshot{
			LineUserID: "U1", SnapshotTakenAt: now, Tags: &empty, Scenarios: &empty,
		}))
		require.NoError(t, err)
		assert.Contains(t, string(raw), `"tags":[]`)
		assert.Contains(t, string(raw), `"scenarios":[]`)
	})
	t.Run("stored values are returned", func(t *testing.T) {
		raw, err := json.Marshal(toLstepFriendAttributeResponse(&model.LstepFriendAttributeSnapshot{
			LineUserID: "U1", SnapshotTakenAt: now, Tags: &tags, Scenarios: &scenarios,
		}))
		require.NoError(t, err)
		assert.Contains(t, string(raw), `"tags":["tagA","tagB"]`)
		assert.Contains(t, string(raw), `"scenarios":["s1"]`)
	})
}
