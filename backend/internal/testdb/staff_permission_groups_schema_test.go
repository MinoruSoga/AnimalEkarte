package testdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedTestSchema_StaffPermissionGroupsHasCreatedAt(t *testing.T) {
	db := SetupTestDB(t)

	var exists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = 'staff_permission_groups'
			  AND column_name = 'created_at'
		)
	`).Scan(&exists).Error
	require.NoError(t, err)
	assert.True(t, exists, "staff_permission_groups.created_at must exist after setupSharedTestSchema")
}
