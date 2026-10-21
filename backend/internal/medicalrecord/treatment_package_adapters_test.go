package medicalrecord

// EMR-521 adapter tests (NOT_RUN — execution deferred).

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestContextViewAuthorizer_MissingCheckerIsDenied(t *testing.T) {
	var authz contextViewAuthorizer
	assert.False(t, authz.Granted(context.Background(), 7, model.ResourceMasterMedical, "view"))
}

func TestTreatmentPackageViewCheckContext_BindsMiddlewareChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, _ string) bool {
		return clinicID == 7 && resource == string(model.ResourceCheckups)
	})

	ctx, ok := treatmentPackageViewCheckContext(c)
	require.True(t, ok)

	var authz contextViewAuthorizer
	assert.True(t, authz.Granted(ctx, 7, model.ResourceCheckups, "view"))
	assert.False(t, authz.Granted(ctx, 7, model.ResourceMasterMerchandise, "view"))
	assert.False(t, authz.Granted(ctx, 9, model.ResourceCheckups, "view"))
}

func TestTreatmentPackageViewCheckContext_SystemAdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("is_system_admin", true)

	ctx, ok := treatmentPackageViewCheckContext(c)
	require.True(t, ok)

	var authz contextViewAuthorizer
	assert.True(t, authz.Granted(ctx, 7, model.ResourceMasterMedical, "view"))
}

func TestTreatmentPackageViewCheckContext_NoCheckerNoAdminFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	_, ok := treatmentPackageViewCheckContext(c)
	assert.False(t, ok) // fail-closed
}
