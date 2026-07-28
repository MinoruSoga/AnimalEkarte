package lstep

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestRegisterRoutes_CheckupSyncRBACTuples(t *testing.T) {
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
		&LstepSettingsHandler{},
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

	// L①/L② register 15 tuples, L③a registers 23, and L③b inserts the
	// two checkup-sync tuples before L④ and the final two line-customer tuples.
	require.Len(t, calls, 67)
	assert.Equal(t, []permissionTuple{
		{string(model.ResourceOwners), "view"},
		{string(model.ResourceOwners), "edit"},
	}, calls[38:40])
}
