package lstep

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestRegisterRoutes_L5RBACTuples(t *testing.T) {
	type permissionTuple struct {
		resource string
		action   string
	}

	var calls []permissionTuple
	recordPermission := func(resource, action string) gin.HandlerFunc {
		calls = append(calls, permissionTuple{resource: resource, action: action})
		return func(*gin.Context) {}
	}

	h := NewHandler(
		&SettingsHandler{},
		&LineSendHandler{},
		&LineLinkHandler{},
		&LineCustomerHandler{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		recordPermission,
		noopPermissionAny,
	)
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))

	// Total requirePermission calls is 61 after SOLO-09 removed 6 tag-config
	// hospital-settings create/delete tuples (mutate is system_admin only).
	require.Len(t, calls, 61)
	assert.Equal(t, []permissionTuple{
		{string(model.ResourceOwners), "view"},
		{string(model.ResourceLstepCsvImport), "edit"},
		{string(model.ResourceLstepCsvImport), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
	}, calls[50:56])
}
