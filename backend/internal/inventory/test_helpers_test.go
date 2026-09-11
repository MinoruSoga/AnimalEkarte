package inventory

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, _ uint64, _, _ string) bool {
		return true
	})
}

func allowInventoryTestPermission(_, _ string) gin.HandlerFunc {
	return func(*gin.Context) {}
}
