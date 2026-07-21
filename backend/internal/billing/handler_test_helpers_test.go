package billing

import (
	"github.com/gin-gonic/gin"
)

// setClinicID は handler テスト用に clinic_id を gin.Context へ設定する
// （internal/medicalrecord・internal/reservation の同名ヘルパーの最小限の複製）。
func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

// setStaffID / setNonSystemAdmin — internal/handler/clinic_handler_test.go の同名ヘルパーの複製。
func setStaffID(c *gin.Context) {
	c.Set("user_id", "1")
}

func setNonSystemAdmin(c *gin.Context) {
	c.Set("is_system_admin", false)
	c.Set("user_id", "1")
	c.Set("clinic_id", "1")
}
