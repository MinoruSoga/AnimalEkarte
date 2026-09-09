package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type grantOnlyClinicAPermissions struct{}

func (grantOnlyClinicAPermissions) GetEffectivePermissions(
	_ context.Context,
	_, clinicID uint64,
) ([]model.PermissionGroupRule, error) {
	return allResourceRulesForClinic(clinicID, 1), nil
}

type grantOnlyClinicBPermissions struct{}

func (grantOnlyClinicBPermissions) GetEffectivePermissions(
	_ context.Context,
	_, clinicID uint64,
) ([]model.PermissionGroupRule, error) {
	return allResourceRulesForClinic(clinicID, 2), nil
}

func allResourceRulesForClinic(clinicID, grantedClinicID uint64) []model.PermissionGroupRule {
	if clinicID != grantedClinicID {
		return nil
	}
	rules := make([]model.PermissionGroupRule, 0, len(model.AllResources))
	for _, resource := range model.AllResources {
		rules = append(rules, model.PermissionGroupRule{
			Resource:  string(resource),
			CanView:   true,
			CanCreate: true,
			CanEdit:   true,
			CanDelete: true,
		})
	}
	return rules
}

func attachMembershipABGrantASelectedB() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("clinic_id", "2")
		c.Set("clinic_ids", []uint64{1, 2})
		c.Set("is_system_admin", false)
		c.Set("user_id", "17")
		c.Next()
	}
}

var inventoryPathParam = regexp.MustCompile(`:[A-Za-z_][A-Za-z0-9_]*|\*[^/]*`)

func inventoryRequestPath(path string) string {
	return inventoryPathParam.ReplaceAllStringFunc(path, func(match string) string {
		lower := strings.ToLower(match)
		switch {
		case strings.Contains(lower, "date"):
			return "2026-01-01"
		case strings.Contains(lower, "category"):
			return "general"
		case strings.Contains(lower, "slug"):
			return "intro"
		case strings.HasPrefix(match, "*"):
			return "file.png"
		default:
			return "2"
		}
	})
}

func TestGETHEADClinicFixedSelectedClinicWithoutGrantIsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	permission := auth.NewHTTPHandler(auth.HTTPDependencies{
		EffectivePermissions: grantOnlyClinicAPermissions{},
	}, auth.CookieConfigForProduction(false))
	clinicFixed := 0
	for _, entry := range readGETHEADInventory(t) {
		if entry.Class != "clinic-fixed" {
			continue
		}
		clinicFixed++
		entry := entry
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			require.NotEqual(t, "none", entry.Resource, entry.Path)
			reached := false
			router := gin.New()
			router.Handle(
				entry.Method,
				entry.Path,
				attachMembershipABGrantASelectedB(),
				permission.RequirePermission(entry.Resource, entry.Action),
				func(c *gin.Context) {
					reached = true
					c.Status(http.StatusOK)
				},
			)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(entry.Method, inventoryRequestPath(entry.Path), http.NoBody)
			router.ServeHTTP(recorder, req)
			require.False(t, reached, "handler must not run for selected clinic B")
			require.Equal(t, http.StatusForbidden, recorder.Code)
		})
	}
	require.Greater(t, clinicFixed, 100)
}

func TestGETHEADClinicFixedSelectedClinicWithGrantAllowsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	permission := auth.NewHTTPHandler(auth.HTTPDependencies{
		EffectivePermissions: grantOnlyClinicBPermissions{},
	}, auth.CookieConfigForProduction(false))
	clinicFixed := 0
	for _, entry := range readGETHEADInventory(t) {
		if entry.Class != "clinic-fixed" {
			continue
		}
		clinicFixed++
		entry := entry
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			require.NotEqual(t, "none", entry.Resource, entry.Path)
			reached := false
			router := gin.New()
			router.Handle(
				entry.Method,
				entry.Path,
				attachMembershipABGrantASelectedB(),
				permission.RequirePermission(entry.Resource, entry.Action),
				func(c *gin.Context) {
					reached = true
					c.Status(http.StatusOK)
				},
			)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(entry.Method, inventoryRequestPath(entry.Path), http.NoBody)
			router.ServeHTTP(recorder, req)
			require.True(t, reached, "handler must run when selected clinic B holds the grant")
			require.Equal(t, http.StatusOK, recorder.Code)
		})
	}
	require.Greater(t, clinicFixed, 100)
}

func TestGETHEADCrossClinicAllowingDoesNotGrantSelectedClinicB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	permission := auth.NewHTTPHandler(auth.HTTPDependencies{
		EffectivePermissions: grantOnlyClinicAPermissions{},
	}, auth.CookieConfigForProduction(false))
	crossClinic := 0
	for _, entry := range readGETHEADInventory(t) {
		if entry.Class != "cross-clinic" {
			continue
		}
		crossClinic++
		entry := entry
		t.Run(entry.Method+" "+entry.Path, func(t *testing.T) {
			reached := false
			var checkerOK bool
			var selectedGranted bool
			router := gin.New()
			router.Handle(
				entry.Method,
				entry.Path,
				attachMembershipABGrantASelectedB(),
				permission.RequirePermissionAllowingAssignedClinicGrant(entry.Resource, entry.Action),
				func(c *gin.Context) {
					reached = true
					check, ok := httpapi.PeekClinicPermissionChecker(c)
					checkerOK = ok
					if ok {
						selectedGranted = check(c, 2, entry.Resource, entry.Action)
					}
					c.Status(http.StatusOK)
				},
			)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(entry.Method, inventoryRequestPath(entry.Path), http.NoBody)
			router.ServeHTTP(recorder, req)
			require.True(t, reached, "allowing middleware may pass because clinic A has the grant")
			require.Equal(t, http.StatusOK, recorder.Code)
			require.True(t, checkerOK)
			require.False(t, selectedGranted, "selected clinic B must remain ungranted")
		})
	}
	require.Equal(t, 21, crossClinic)
}

func TestGETHEADPublicLiffSelfAreNotSelectedClinicGrantSurfaces(t *testing.T) {
	for _, entry := range readGETHEADInventory(t) {
		switch entry.Class {
		case "public", "liff", "self":
			require.Equal(t, "none", entry.Resource, entry.Path)
			require.Equal(t, "none", entry.Action, entry.Path)
		case "shared-master", "clinic-fixed", "cross-clinic":
			require.NotEqual(t, "none", entry.Resource, entry.Path)
		default:
			t.Fatalf("unclassified GET/HEAD route %s %s", entry.Method, entry.Path)
		}
	}
}
