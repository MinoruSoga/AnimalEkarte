package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

func TestRequirePermissionAny_SharedFilePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		rules      []model.PermissionGroupRule
		lookupErr  error
		wantStatus int
	}{
		{
			name:       "owners edit permits upload",
			rules:      []model.PermissionGroupRule{{Resource: string(model.ResourceOwners), CanEdit: true}},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "medical records create permits upload",
			rules:      []model.PermissionGroupRule{{Resource: string(model.ResourceMedicalRecords), CanCreate: true}},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "medical records edit permits upload",
			rules:      []model.PermissionGroupRule{{Resource: string(model.ResourceMedicalRecords), CanEdit: true}},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "unrelated permission is denied",
			rules:      []model.PermissionGroupRule{{Resource: string(model.ResourceOwners), CanView: true}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing permissions are denied",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "permission lookup failure is denied",
			lookupErr:  errors.New("lookup failed"),
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{svc: &service.Services{EffectivePermission: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(context.Context, uint64, uint64) ([]model.PermissionGroupRule, error) {
					return tt.rules, tt.lookupErr
				},
			}}}
			r := gin.New()
			r.GET("/shared-files", func(c *gin.Context) {
				c.Set("is_system_admin", false)
				c.Set("user_id", "7")
				c.Set("clinic_id", "1")
				c.Next()
			}, h.RequirePermissionAny(
				struct{ Resource, Action string }{Resource: string(model.ResourceOwners), Action: "edit"},
				struct{ Resource, Action string }{Resource: string(model.ResourceMedicalRecords), Action: "create"},
				struct{ Resource, Action string }{Resource: string(model.ResourceMedicalRecords), Action: "edit"},
			), func(c *gin.Context) { c.Status(http.StatusNoContent) })

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/shared-files", http.NoBody))

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
