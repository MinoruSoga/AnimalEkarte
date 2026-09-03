package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clinicPermissionAllowAll(_ *gin.Context, _ uint64, _, _ string) bool {
	return true
}

func clinicPermissionOnlyClinic(
	allowedClinic uint64,
	allowedResource, allowedAction string,
) ClinicPermissionChecker {
	return func(_ *gin.Context, clinicID uint64, resource, action string) bool {
		return clinicID == allowedClinic && resource == allowedResource && action == allowedAction
	}
}

func TestPeekClinicPermissionChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)

	_, ok := PeekClinicPermissionChecker(c)
	assert.False(t, ok)
	assert.False(t, c.Writer.Written())

	SetClinicPermissionChecker(c, nil)
	_, ok = PeekClinicPermissionChecker(c)
	assert.False(t, ok)

	SetClinicPermissionChecker(c, clinicPermissionAllowAll)
	check, ok := PeekClinicPermissionChecker(c)
	require.True(t, ok)
	assert.True(t, check(c, 9, "owners", "view"))
}

func TestAuthorizeClinicIDsForPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects destination without per-clinic grant", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
		c.Set("clinic_ids", []uint64{1, 2})
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(1, "owners", "create"))

		assert.False(t, AuthorizeClinicIDsForPermission(c, []uint64{2}, "owners", "create"))
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "forbidden")
	})

	t.Run("allows destination with membership and grant", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
		c.Set("clinic_ids", []uint64{1, 2})
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(2, "owners", "create"))

		assert.True(t, AuthorizeClinicIDsForPermission(c, []uint64{2}, "owners", "create"))
		assert.False(t, c.Writer.Written())
	})

	t.Run("fails closed without checker", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
		c.Set("is_system_admin", false)
		c.Set("clinic_ids", []uint64{1, 2})

		assert.False(t, AuthorizeClinicIDsForPermission(c, []uint64{1}, "owners", "create"))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("system admin bypasses per-clinic grant after membership", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
		c.Set("is_system_admin", true)
		c.Set("clinic_ids", []uint64{1, 2})

		assert.True(t, AuthorizeClinicIDsForPermission(c, []uint64{2}, "owners", "create"))
		assert.False(t, c.Writer.Written())
	})

	t.Run("still rejects unassigned clinic before permission check", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
		c.Set("clinic_ids", []uint64{1, 2})
		SetClinicPermissionChecker(c, clinicPermissionAllowAll)

		assert.False(t, AuthorizeClinicIDsForPermission(c, []uint64{99}, "owners", "create"))
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "not assigned to this clinic")
	})
}

func TestFilterClinicIDsForPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("keeps only clinics with the grant", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(1, "owners", "view"))

		got, ok := FilterClinicIDsForPermission(c, []uint64{1, 2, 1}, "owners", "view")
		require.True(t, ok)
		assert.Equal(t, []uint64{1}, got)
	})

	t.Run("rejects when no clinic remains", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(1, "owners", "view"))

		got, ok := FilterClinicIDsForPermission(c, []uint64{2}, "owners", "view")
		assert.False(t, ok)
		assert.Nil(t, got)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestResolveListClinicIDsForPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects clinic_ids that lack the grant", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?clinic_ids=2", http.NoBody)
		c.Set("clinic_id", "1")
		c.Set("clinic_ids", []uint64{1, 2})
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(1, "owners", "view"))

		got, ok := ResolveListClinicIDsForPermission(c, "owners", "view")
		assert.False(t, ok)
		assert.Nil(t, got)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("filters mixed clinic_ids to the granted subset", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?clinic_ids=1,2", http.NoBody)
		c.Set("clinic_id", "1")
		c.Set("clinic_ids", []uint64{1, 2})
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(1, "owners", "view"))

		got, ok := ResolveListClinicIDsForPermission(c, "owners", "view")
		require.True(t, ok)
		assert.Equal(t, []uint64{1}, got)
	})
}

func TestResolveAllClinicIDsForPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("filters assigned clinics by grant", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Set("clinic_id", "1")
		c.Set("is_system_admin", false)
		c.Set("clinic_ids", []uint64{1, 2})
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(1, "owners", "view"))

		got, ok := ResolveAllClinicIDsForPermission(c, "owners", "view")
		require.True(t, ok)
		assert.Equal(t, []uint64{1}, got)
	})

	t.Run("system admin stays on the selected clinic", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Set("clinic_id", "1")
		c.Set("is_system_admin", true)
		c.Set("clinic_ids", []uint64{1, 2, 3})
		SetClinicPermissionChecker(c, clinicPermissionAllowAll)

		got, ok := ResolveAllClinicIDsForPermission(c, "owners", "view")
		require.True(t, ok)
		assert.Equal(t, []uint64{1}, got)
	})
}

func TestRequireSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("fails closed when checker is absent", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Set("clinic_id", "23")

		assert.False(t, RequireSelectedClinicGrant(c, "shifts", "view"))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("rejects selected clinic without grant", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Set("is_system_admin", false)
		c.Set("clinic_id", "23")
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(99, "shifts", "view"))

		assert.False(t, RequireSelectedClinicGrant(c, "shifts", "view"))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("allows selected clinic with grant", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Set("is_system_admin", false)
		c.Set("clinic_id", "23")
		SetClinicPermissionChecker(c, clinicPermissionOnlyClinic(23, "shifts", "view"))

		assert.True(t, RequireSelectedClinicGrant(c, "shifts", "view"))
		assert.False(t, c.Writer.Written())
	})
}
