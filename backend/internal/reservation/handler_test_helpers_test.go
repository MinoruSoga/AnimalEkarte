package reservation

import (
	"github.com/gin-gonic/gin"
)

// setClinicID は handler テスト用に clinic_id を gin.Context へ設定する
// （internal/medicalrecord の同名ヘルパーの最小限の複製）。
func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}
