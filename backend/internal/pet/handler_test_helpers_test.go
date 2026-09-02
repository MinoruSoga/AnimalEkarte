package pet

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func allowAllPermission(_, _ string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, _ uint64, _, _ string) bool {
		return true
	})
}

func setOwnersViewOnlyClinic(c *gin.Context, clinicID uint64) {
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, id uint64, resource, action string) bool {
		return id == clinicID && resource == string(model.ResourceOwners) && action == "view"
	})
}

func setStaffID(c *gin.Context, staffID uint64) {
	c.Set("user_id", strconv.FormatUint(staffID, 10))
}
