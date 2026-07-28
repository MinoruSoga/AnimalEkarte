package pet

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func allowAllPermission(_, _ string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

func setStaffID(c *gin.Context, staffID uint64) {
	c.Set("user_id", strconv.FormatUint(staffID, 10))
}
