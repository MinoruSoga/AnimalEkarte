package billing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// bindJSONBody binds payload onto dest with gin ShouldBindJSON (owner/pet freetext pattern).
func bindJSONBody(t *testing.T, payload any, dest any) error {
	t.Helper()
	gin.SetMode(gin.TestMode)
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c.ShouldBindJSON(dest)
}

func allowAllClinicPermission(c *gin.Context) {
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, _ uint64, _, _ string) bool {
		return true
	})
}

func setResourcePermissionOnlyClinic(c *gin.Context, clinicID uint64, resource, action string) {
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, id uint64, res, act string) bool {
		return id == clinicID && res == resource && act == action
	})
}

func setAccountingPermissionOnlyClinic(c *gin.Context, clinicID uint64, action string) {
	setResourcePermissionOnlyClinic(c, clinicID, string(model.ResourceAccounting), action)
}

// setClinicID は handler テスト用に clinic_id を gin.Context へ設定する
// （internal/medicalrecord・internal/reservation の同名ヘルパーの最小限の複製）。
func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
	allowAllClinicPermission(c)
}

// setStaffID / setNonSystemAdmin — internal/handler/clinic_handler_test.go の同名ヘルパーの複製。
func setStaffID(c *gin.Context) {
	c.Set("user_id", "1")
}

func setNonSystemAdmin(c *gin.Context) {
	c.Set("is_system_admin", false)
	c.Set("user_id", "1")
	c.Set("clinic_id", "1")
	allowAllClinicPermission(c)
}
