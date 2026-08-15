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

	// L①/L② register 15 tuples first. L③a tag-core is 17 tuples after SOLO-09:
	// tag-config POST/DELETE use requireSystemAdmin (not hospital-settings RBAC).
	require.Len(t, calls, 61)
	tagCoreCalls := calls[15:32]
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
		{string(model.ResourceHospitalSettings), "view"},
		{string(model.ResourceHospitalSettings), "view"},
	}
	assert.Equal(t, expected, tagCoreCalls)
}
