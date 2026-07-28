package inventory

import "github.com/gin-gonic/gin"

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

func allowInventoryTestPermission(_, _ string) gin.HandlerFunc {
	return func(*gin.Context) {}
}
