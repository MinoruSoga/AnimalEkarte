package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// bindPermissionGroupCreateRequest binds JSON through the auth package boundary
// (bindAuthJSON) so the fixture matches the path AUTH-B will reuse.
func bindPermissionGroupCreateRequest(t *testing.T, body string) (PermissionGroupCreateRequest, error) {
	t.Helper()

	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/permission-groups",
		bytes.NewBufferString(body),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	var request PermissionGroupCreateRequest
	err := bindAuthJSON(c, &request)
	return request, err
}

func TestPermissionGroupCreateRequest_IsActiveOmitted(t *testing.T) {
	t.Parallel()

	request, err := bindPermissionGroupCreateRequest(t, `{
		"name": "Reception",
		"color": "#112233"
	}`)
	require.NoError(t, err)
	require.Nil(t, request.IsActive, "omitted is_active must remain nil (presence absent)")
	require.Equal(t, "Reception", request.Name)
	require.Equal(t, "#112233", request.Color)
}

func TestPermissionGroupCreateRequest_IsActiveFalse(t *testing.T) {
	t.Parallel()

	request, err := bindPermissionGroupCreateRequest(t, `{
		"name": "Reception",
		"color": "#112233",
		"is_active": false
	}`)
	require.NoError(t, err)
	require.NotNil(t, request.IsActive, "explicit false must be present")
	require.False(t, *request.IsActive)
}

func TestPermissionGroupCreateRequest_IsActiveTrue(t *testing.T) {
	t.Parallel()

	request, err := bindPermissionGroupCreateRequest(t, `{
		"name": "Reception",
		"color": "#112233",
		"is_active": true
	}`)
	require.NoError(t, err)
	require.NotNil(t, request.IsActive, "explicit true must be present")
	require.True(t, *request.IsActive)
}
