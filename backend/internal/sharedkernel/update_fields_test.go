package sharedkernel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetNullableUint64Field(t *testing.T) {
	t.Run("clearAssoc true: writes explicit NULL regardless of id", func(t *testing.T) {
		fields := map[string]any{}
		id := uint64(5)
		SetNullableUint64Field(fields, "group_id", true, &id)
		assert.Contains(t, fields, "group_id")
		assert.Nil(t, fields["group_id"])
	})

	t.Run("clearAssoc true and id nil: still writes explicit NULL", func(t *testing.T) {
		fields := map[string]any{}
		SetNullableUint64Field(fields, "group_id", true, nil)
		assert.Contains(t, fields, "group_id")
		assert.Nil(t, fields["group_id"])
	})

	t.Run("clearAssoc false and id set: writes the value", func(t *testing.T) {
		fields := map[string]any{}
		id := uint64(7)
		SetNullableUint64Field(fields, "group_id", false, &id)
		assert.Equal(t, uint64(7), fields["group_id"])
	})

	t.Run("clearAssoc false and id nil: leaves the column untouched (omitted)", func(t *testing.T) {
		fields := map[string]any{}
		SetNullableUint64Field(fields, "group_id", false, nil)
		assert.NotContains(t, fields, "group_id")
	})

	t.Run("does not disturb other keys already present in the map", func(t *testing.T) {
		fields := map[string]any{"name": "existing"}
		id := uint64(1)
		SetNullableUint64Field(fields, "parent_id", false, &id)
		assert.Equal(t, "existing", fields["name"])
		assert.Equal(t, uint64(1), fields["parent_id"])
	})
}
