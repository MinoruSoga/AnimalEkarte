package lstep

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
// internal/handler/testdata/route_snapshot.golden から移した route を保護する。L① は settings、
// L② は LINE 送信・紐付け・顧客管理、L③a は tag-sync core の 23 route を追加した。
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(
		NewSettingsHandler(nil, noopPermission),
		NewLineSendHandler(nil, noopPermission),
		NewLineLinkHandler(nil, noopPermission),
		NewLineCustomerHandler(nil, noopPermission),
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
		noopPermission,
		noopPermissionAny,
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

	want := `DELETE /api/v1/clinics/:clinic_id/lstep-settings DeleteLstepSettings
DELETE /api/v1/clinics/:clinic_id/owners/:id/lstep/tags/:tag_name DeleteOwnerLstepTag
DELETE /api/v1/clinics/:clinic_id/pets/:id/death DeletePetDeath
DELETE /api/v1/lstep-settings DeleteLstepSettings
DELETE /api/v1/lstep-tag-config/auto-managed-prefixes/:id DeleteAutoManagedPrefix
DELETE /api/v1/lstep-tag-config/condition-tag-mappings/:id DeleteConditionTagMapping
DELETE /api/v1/lstep-tag-config/send-purpose-tag-prefixes/:id DeleteSendPurposeTagPrefix
DELETE /api/v1/owners/:id/line DeleteOwnerLine
DELETE /api/v1/owners/:id/lstep/tags/:tag_name DeleteOwnerLstepTag
DELETE /api/v1/pets/:id/death DeletePetDeath
DELETE /api/v1/shared-files/:id DeleteSharedFile
GET /api/v1/clinics/:clinic_id/line-customers ListLineCustomers
GET /api/v1/clinics/:clinic_id/lstep-settings GetLstepSettings
GET /api/v1/clinics/:clinic_id/lstep-tag-code-mappings ListTagCodeMappings
GET /api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats GetLstepMonthlyDeliveryStats
GET /api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion GetLstepVisitConversionSummary
GET /api/v1/clinics/:clinic_id/lstep/checkup-sync/preview GetCheckupSyncPreview
GET /api/v1/clinics/:clinic_id/lstep/csv-imports ListLstepCsvImports
GET /api/v1/clinics/:clinic_id/lstep/delivery-monitor/logs GetLstepDeliveryTriggerLogs
GET /api/v1/clinics/:clinic_id/lstep/delivery-monitor/summary GetLstepDeliveryTriggerSummary
GET /api/v1/clinics/:clinic_id/lstep/owners SearchLstepOwnersByTag
GET /api/v1/clinics/:clinic_id/lstep/tag-summary GetLstepTagSummary
GET /api/v1/clinics/:clinic_id/lstep/trigger-priorities GetLstepTriggerPriorities
GET /api/v1/clinics/:clinic_id/owners/:id/line/send-logs GetLineSendLogs
GET /api/v1/clinics/:clinic_id/owners/:id/lstep/friend-attributes GetLstepOwnerFriendAttributes
GET /api/v1/clinics/:clinic_id/owners/:id/lstep/tags GetOwnerLstepTags
GET /api/v1/clinics/:clinic_id/owners/aggregations ListOwnerAggregation
GET /api/v1/lstep-settings GetLstepSettings
GET /api/v1/lstep-tag-code-mappings ListTagCodeMappings
GET /api/v1/lstep-tag-config/auto-managed-prefixes ListAutoManagedPrefixes
GET /api/v1/lstep-tag-config/condition-tag-mappings ListConditionTagMappings
GET /api/v1/lstep-tag-config/send-purpose-tag-prefixes ListSendPurposeTagPrefixes
GET /api/v1/lstep/delivery-monitor/logs GetLstepDeliveryTriggerLogs
GET /api/v1/lstep/delivery-monitor/summary GetLstepDeliveryTriggerSummary
GET /api/v1/lstep/owners SearchLstepOwnersByTag
GET /api/v1/lstep/tag-summary GetLstepTagSummary
GET /api/v1/lstep/trigger-priorities GetLstepTriggerPriorities
GET /api/v1/owners/:id/line/send-logs GetLineSendLogs
GET /api/v1/owners/:id/lstep/send-history GetLineSendLogs
GET /api/v1/owners/:id/lstep/tags GetOwnerLstepTags
GET /api/v1/shared-files ListSharedFiles
GET /api/v1/shared-files/:id/signed-url GetSharedFileSignedURL
PATCH /api/v1/clinics/:clinic_id/line-customers/:customerId/link-owner LinkOwnerToLineCustomer
PATCH /api/v1/clinics/:clinic_id/lstep-settings UpdateLstepSettings
PATCH /api/v1/clinics/:clinic_id/lstep/trigger-priorities UpdateLstepTriggerPriorities
PATCH /api/v1/clinics/:clinic_id/pets/:id/death UpdatePetDeath
PATCH /api/v1/lstep-settings UpdateLstepSettings
PATCH /api/v1/lstep/trigger-priorities UpdateLstepTriggerPriorities
PATCH /api/v1/owners/:id/lstep/opt-out PatchOwnerLstepOptOut
PATCH /api/v1/pets/:id/death UpdatePetDeath
POST /api/line/webhook ReceiveLineWebhook
POST /api/v1/clinics/:clinic_id/lstep-settings/test-connection TestLstepConnection
POST /api/v1/clinics/:clinic_id/lstep/checkup-sync CreateCheckupSync
POST /api/v1/clinics/:clinic_id/lstep/csv-imports/friend-attributes ImportLstepFriendAttributesCsv
POST /api/v1/clinics/:clinic_id/owners/:id/line/send SendLineMessage
POST /api/v1/clinics/:clinic_id/owners/:id/lstep-opt-out UpdateOwnerLstepOptOut
POST /api/v1/clinics/:clinic_id/owners/:id/lstep/tags AddOwnerLstepTag
POST /api/v1/lstep-settings/test-connection TestLstepConnection
POST /api/v1/lstep-tag-config/auto-managed-prefixes CreateAutoManagedPrefix
POST /api/v1/lstep-tag-config/condition-tag-mappings CreateConditionTagMapping
POST /api/v1/lstep-tag-config/send-purpose-tag-prefixes CreateSendPurposeTagPrefix
POST /api/v1/owners/:id/line/link-token GenerateLineLinkToken
POST /api/v1/owners/:id/line/send SendLineMessage
POST /api/v1/owners/:id/lstep-opt-out UpdateOwnerLstepOptOut
POST /api/v1/owners/:id/lstep/send SendLineMessage
POST /api/v1/owners/:id/lstep/tags AddOwnerLstepTag
POST /api/v1/shared-files UploadSharedFile
PUT /api/v1/clinics/:clinic_id/lstep-tag-code-mappings/:tag_name ReplaceTagCodeMappingsForTag
PUT /api/v1/lstep-tag-code-mappings/:tag_name ReplaceTagCodeMappingsForTag
`

	assert.Equal(t, want, got)
}
