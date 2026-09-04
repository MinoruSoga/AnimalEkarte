package reservation

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// setClinicID は handler テスト用に clinic_id を gin.Context へ設定する
// （internal/medicalrecord の同名ヘルパーの最小限の複製）。
func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, _ uint64, _, _ string) bool {
		return true
	})
}

func setReservationsViewOnlyClinic(c *gin.Context, clinicID uint64) {
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, id uint64, resource, action string) bool {
		return id == clinicID && resource == string(model.ResourceReservations) && action == "view"
	})
}

// withJSTLocal は handler/pet_birthdate_consistency_test.go の同名ヘルパーの複製。
func withJSTLocal(t *testing.T) {
	t.Helper()
	orig := time.Local
	time.Local = config.JST
	t.Cleanup(func() { time.Local = orig })
}
