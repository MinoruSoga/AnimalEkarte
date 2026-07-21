package reservation

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
)

// setClinicID は handler テスト用に clinic_id を gin.Context へ設定する
// （internal/medicalrecord の同名ヘルパーの最小限の複製）。
func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

// withJSTLocal は handler/pet_birthdate_consistency_test.go の同名ヘルパーの複製。
func withJSTLocal(t *testing.T) {
	t.Helper()
	orig := time.Local
	time.Local = config.JST
	t.Cleanup(func() { time.Local = orig })
}
