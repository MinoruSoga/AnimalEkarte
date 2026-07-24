package pet

import "github.com/gin-gonic/gin"

func allowAllPermission(_, _ string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}
