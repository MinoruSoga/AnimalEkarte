package lstep

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// lastHandlerSegment は internal/medicalrecord の同名ヘルパーと同一実装。
func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}

// TestRegisterRoutes_Snapshot は lstep 側の BE9-2C route-snapshot 回帰チェック
// （medicalrecord/reservation/billing の先例を踏襲）。L① で
// internal/handler/testdata/route_snapshot.golden から 8 route（lstep-settings 4 +
// clinics エイリアス 4）を drop し、本 package の RegisterRoutes が登録する。L② で LINE
// 送信・紐付け・顧客管理の 9 route（protected）と Webhook 1 route（engine 直下・認証なし）を追加。
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(
		NewLstepSettingsHandler(nil, noopPermission),
		NewLineSendHandler(nil, noopPermission),
		NewLineLinkHandler(nil, func(_ *gin.Context, _ *model.Owner) {}, noopPermission),
		NewLineCustomerHandler(nil, noopPermission),
		noopPermission,
	)

	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	// Webhook（engine 直下・JWT 認証なし）も同一 snapshot で保護する
	h.RegisterWebhookRoutes(r)

	lines := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s %s", route.Method, route.Path, lastHandlerSegment(route.Handler)))
	}
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"

	want := "" +
		"DELETE /api/v1/clinics/:clinic_id/lstep-settings DeleteLstepSettings\n" +
		"DELETE /api/v1/lstep-settings DeleteLstepSettings\n" +
		"GET /api/v1/clinics/:clinic_id/line-customers ListLineCustomers\n" +
		"GET /api/v1/clinics/:clinic_id/lstep-settings GetLstepSettings\n" +
		"GET /api/v1/clinics/:clinic_id/owners/:id/line/send-logs GetLineSendLogs\n" +
		"GET /api/v1/lstep-settings GetLstepSettings\n" +
		"GET /api/v1/owners/:id/line/send-logs GetLineSendLogs\n" +
		"GET /api/v1/owners/:id/lstep/send-history GetLineSendLogs\n" +
		"PATCH /api/v1/clinics/:clinic_id/line-customers/:customerId/link-owner LinkOwnerToLineCustomer\n" +
		"PATCH /api/v1/clinics/:clinic_id/lstep-settings UpdateLstepSettings\n" +
		"PATCH /api/v1/lstep-settings UpdateLstepSettings\n" +
		"POST /api/line/webhook ReceiveLineWebhook\n" +
		"POST /api/v1/clinics/:clinic_id/lstep-settings/test-connection TestLstepConnection\n" +
		"POST /api/v1/clinics/:clinic_id/owners/:id/line/send SendLineMessage\n" +
		"POST /api/v1/lstep-settings/test-connection TestLstepConnection\n" +
		"POST /api/v1/owners/:id/line/link-token GenerateLineLinkToken\n" +
		"POST /api/v1/owners/:id/line/send SendLineMessage\n" +
		"POST /api/v1/owners/:id/lstep/send SendLineMessage\n"

	assert.Equal(t, want, got)
}
