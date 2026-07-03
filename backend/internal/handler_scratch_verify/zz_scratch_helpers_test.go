package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

func setStaffID(c *gin.Context) {
	c.Set("user_id", "1")
}

func setNonSystemAdmin(c *gin.Context) {
	c.Set("is_system_admin", false)
	c.Set("user_id", "1")
	c.Set("clinic_id", "1")
}

type mockEffectivePermissionService struct {
	getEffectivePermissionsFn func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
}

func (m *mockEffectivePermissionService) GetEffectivePermissions(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
	if m.getEffectivePermissionsFn != nil {
		return m.getEffectivePermissionsFn(ctx, staffID, clinicID)
	}
	return nil, nil
}
