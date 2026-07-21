package lstep

import (
	"github.com/gin-gonic/gin"
)

// setClinicID は handler テスト用に clinic_id を gin.Context へ設定する
// （medicalrecord/reservation/billing の同名ヘルパーの最小限の複製）。
func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

func strPtr(s string) *string { return &s }

func intPtr(v int) *int { return &v }

func ptrInt64(v int64) *int64 { return &v }
