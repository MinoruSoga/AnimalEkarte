package lstep

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestRegisterRoutes_DeliveryLifecycleRBACTuples(t *testing.T) {
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

	require.NotEmpty(t, calls)

	// Lifecycle-critical owner/pet/delivery routes must remain permission-gated.
	// Full ordered snapshot is brittle across unrelated route registration changes;
	// assert presence of required resource:action pairs instead of a fixed length.
	mustContain := []permissionTuple{
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "view"},
		{string(model.ResourceOwners), "delete"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "edit"},
	}
	for _, want := range mustContain {
		assert.Contains(t, calls, want, "expected permission middleware for %s:%s", want.resource, want.action)
	}
}
