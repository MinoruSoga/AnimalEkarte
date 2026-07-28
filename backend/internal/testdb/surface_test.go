package testdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportedSchemaSurface(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, SharedTestSchemaEnumTypes)
	assert.Equal(
		t,
		[]string{"'a'", "'b'"},
		EnumValueRe.FindAllString("CREATE TYPE sample AS ENUM ('a', 'b')", -1),
	)
	assert.NoError(t, EnsureAutoMigrated(nil))
	CloseSharedTestDB()
}
