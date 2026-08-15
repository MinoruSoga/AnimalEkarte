package lstep

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterWebhookRoutes_AppliesInjectedRateLimitBeforeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) { c.Next() }
	}
	h := NewHandler(
		NewLstepSettingsHandler(nil, noopPermission),
		NewLineSendHandler(nil, noopPermission),
		NewLineLinkHandler(nil, noopPermission),
		NewLineCustomerHandler(nil, noopPermission),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		noopPermission,
		noopPermissionAny,
	)
	r := gin.New()
	h.RegisterWebhookRoutes(r, func(c *gin.Context) {
		c.AbortWithStatus(http.StatusTooManyRequests)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/line/webhook", http.NoBody)
	request.Header.Set("X-Line-Signature", "signed")
	r.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
}
