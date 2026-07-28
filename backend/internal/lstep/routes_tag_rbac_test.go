package lstep

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestRegisterRoutes_TagCoreRBACTuples(t *testing.T) {
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

	// L①/L② register 15 tuples first. L③a's 23 tuples remain at the same
	// positions; L③b/L④ append their tuples before the final line-customer pair.
	require.Len(t, calls, 67)
	tagCoreCalls := calls[15:38]
	expected := []permissionTuple{
		{string(model.ResourceOwners), "view"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "delete"},
		{string(model.ResourceOwners), "view"},
		{string(model.ResourceOwners), "edit"},
		{string(model.ResourceOwners), "delete"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceLstepAnalytics), "view"},
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "edit"},
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "edit"},
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "create"},
		{string(model.ResourceHospitalSettings), "delete"},
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "create"},
		{string(model.ResourceHospitalSettings), "delete"},
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "create"},
		{string(model.ResourceHospitalSettings), "delete"},
	}
	assert.Equal(t, expected, tagCoreCalls)
}
