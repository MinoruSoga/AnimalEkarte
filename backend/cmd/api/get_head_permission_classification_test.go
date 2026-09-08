package main

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
)

type getAuthClass string

const (
	getAuthPublic       getAuthClass = "public"
	getAuthInternal     getAuthClass = "internal"
	getAuthClinicFixed  getAuthClass = "clinic-fixed"
	getAuthCrossClinic  getAuthClass = "cross-clinic"
	getAuthSharedMaster getAuthClass = "shared-master"
)

func TestGETHEADRoutesAreClassifiedForSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Setenv("STORAGE_TYPE", "")
	t.Setenv("SCHEDULER_INTERNAL_TOKEN", "test-scheduler-internal-token-32b!!")

	composition := newRuntimeComposition(runtimeCompositionDependencies{
		Config: &config.Config{
			JWTSecret: "test-secret-for-route-registration",
		},
	})
	router := gin.New()
	require.NoError(t, composition.registerRoutes(ctx, router, nil, false))

	unclassified := make([]string, 0)
	for _, route := range router.Routes() {
		if route.Method != http.MethodGet && route.Method != http.MethodHead {
			continue
		}
		if _, ok := classifyGETHEADRoute(route.Method, route.Path); !ok {
			unclassified = append(unclassified, route.Method+" "+route.Path)
		}
	}
	require.Empty(t, unclassified, "unclassified GET/HEAD routes: %s", strings.Join(unclassified, ", "))
}

func classifyGETHEADRoute(method, path string) (getAuthClass, bool) {
	switch {
	case path == "/health",
		path == "/api/v1/health",
		strings.HasPrefix(path, "/uploads/"),
		path == "/api/v1/login",
		strings.HasPrefix(path, "/api/v1/auth/"),
		strings.HasPrefix(path, "/api/liff/"),
		path == "/api/line/webhook":
		return getAuthPublic, true
	case strings.HasPrefix(path, "/_internal/"):
		return getAuthInternal, true
	case strings.HasPrefix(path, "/api/v1/owners"),
		strings.HasPrefix(path, "/api/v1/pets"),
		strings.HasPrefix(path, "/api/v1/identity-links"),
		strings.HasPrefix(path, "/api/v1/reservations"),
		strings.HasPrefix(path, "/api/v1/accountings"),
		strings.HasPrefix(path, "/api/v1/billing"),
		strings.HasPrefix(path, "/api/v1/estimates"),
		strings.HasPrefix(path, "/api/v1/medical-records"),
		strings.HasPrefix(path, "/api/v1/hospitalization"),
		strings.HasPrefix(path, "/api/v1/examinations"),
		strings.HasPrefix(path, "/api/v1/vaccinations"),
		strings.HasPrefix(path, "/api/v1/checkups"),
		strings.HasPrefix(path, "/api/v1/lab-"),
		strings.HasPrefix(path, "/api/v1/manual-articles"),
		strings.HasPrefix(path, "/api/v1/manual_articles"),
		strings.HasPrefix(path, "/api/v1/manual/articles"),
		path == "/api/v1/me":
		return getAuthCrossClinic, true
	case strings.HasPrefix(path, "/api/v1/masters/staffs"),
		strings.HasPrefix(path, "/api/v1/masters/occupations"),
		strings.HasPrefix(path, "/api/v1/masters/permission-groups"),
		strings.HasPrefix(path, "/api/v1/shifts"),
		strings.HasPrefix(path, "/api/v1/shift-templates"),
		strings.HasPrefix(path, "/api/v1/inventory"),
		strings.HasPrefix(path, "/api/v1/merchandise"),
		strings.HasPrefix(path, "/api/v1/trimming"),
		strings.HasPrefix(path, "/api/v1/lstep"),
		strings.HasPrefix(path, "/api/v1/clinics"),
		strings.HasPrefix(path, "/api/v1/companies"),
		path == "/api/v1/company",
		strings.HasPrefix(path, "/api/v1/payment-methods"),
		strings.HasPrefix(path, "/api/v1/closing-settings"):
		return getAuthClinicFixed, true
	case strings.HasPrefix(path, "/api/v1/masters/"):
		return getAuthSharedMaster, true
	case strings.HasPrefix(path, "/api/v1/shared-files"),
		strings.HasPrefix(path, "/api/v1/files"),
		strings.HasPrefix(path, "/api/v1/clinic-holidays"),
		strings.HasPrefix(path, "/api/v1/cash-register"),
		strings.HasPrefix(path, "/api/v1/discounts"),
		strings.HasPrefix(path, "/api/v1/reports"),
		strings.HasPrefix(path, "/api/v1/owner-reports"),
		strings.HasPrefix(path, "/api/v1/users"):
		return getAuthClinicFixed, true
	default:
		return "", false
	}
}
