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

	// L①〜L③b register 40 tuples before L④. L④ preserves the four owner
	// lifecycle routes, four pet-death routes, four delivery-monitor routes, and
	// four trigger-priority routes.
	require.Len(t, calls, 67)
	assert.Equal(t, []permissionTuple{
		{string(model.ResourceOwners), "delete"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
	}, calls[40:52])
	assert.Equal(t, []permissionTuple{
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "edit"},
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "edit"},
	}, calls[52:56])
}
