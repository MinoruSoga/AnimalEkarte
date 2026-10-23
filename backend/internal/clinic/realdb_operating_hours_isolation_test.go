package clinic

// realdb_operating_hours_isolation_test.go — D3 evidence for the three common operating-hours reads
// (GET /clinics/:clinic_id/operating-hours, GET /clinic-holidays/common,
// GET /closing-settings/holidays/common).
//
// Real repository + service + HTTP handler on a disposable 001+008 database, membership {A,B}, selected clinic B,
// hospital-settings:view held on A only, shifts:view held on B only. Skipped under -short (testdb).

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	realDBOperatingHoursReasonA = "OH-realdb-closed-A"
	realDBOperatingHoursReasonB = "OH-realdb-closed-B"
)

func TestRealDB_OperatingHours_SelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupOperatingHoursRepositoryDB(t)
	clinicA := seedOperatingHoursClinic(t, db, "oh-isolation-a")
	clinicB := seedOperatingHoursClinic(t, db, "oh-isolation-b")
	insertOperatingHoursRow(t, db, clinicA, `[{"date":"2026-09-10","reason":"`+realDBOperatingHoursReasonA+`"}]`, true)
	insertOperatingHoursRow(t, db, clinicB, `[{"date":"2026-09-18","reason":"`+realDBOperatingHoursReasonB+`"}]`, false)

	passThrough := func(string, string) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }
	newRouter := func(shiftsGrantOn uint64) *gin.Engine {
		handler := NewOperatingHoursHandler(
			NewOperatingHoursReadService(NewOperatingHoursRepository(db)), passThrough, passThrough)
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("clinic_id", strconv.FormatUint(clinicB, 10))
			c.Set("clinic_ids", []uint64{clinicA, clinicB})
			c.Set("is_system_admin", false)
			c.Set("user_id", "17")
			httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
				switch {
				case resource == string(model.ResourceHospitalSettings) && action == "view":
					return clinicID == clinicA
				case resource == string(model.ResourceShifts) && action == "view":
					return clinicID == shiftsGrantOn
				default:
					return false
				}
			})
			c.Next()
		})
		handler.RegisterOperatingHoursRoutes(router.Group("/api/v1"))
		return router
	}
	get := func(router *gin.Engine, target string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, http.NoBody))
		return w
	}
	idText := func(id uint64) string { return strconv.FormatUint(id, 10) }

	t.Run("path clinic A with the grant on A returns only A's stored settings", func(t *testing.T) {
		w := get(newRouter(clinicB), "/api/v1/clinics/"+idText(clinicA)+"/operating-hours")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"clinic_id":"`+idText(clinicA)+`"`)
		assert.Contains(t, w.Body.String(), realDBOperatingHoursReasonA)
		assert.NotContains(t, w.Body.String(), realDBOperatingHoursReasonB)
	})

	t.Run("path clinic B without the grant on B is forbidden and returns none of B's data", func(t *testing.T) {
		w := get(newRouter(clinicB), "/api/v1/clinics/"+idText(clinicB)+"/operating-hours")
		require.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBOperatingHoursReasonB)
		assert.NotContains(t, w.Body.String(), "weekly_hours")
	})

	for _, path := range []string{"/api/v1/clinic-holidays/common", "/api/v1/closing-settings/holidays/common"} {
		t.Run(path+" returns the selected clinic's month only", func(t *testing.T) {
			w := get(newRouter(clinicB), path+"?year_month=2026-09")
			require.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"clinic_id":"`+idText(clinicB)+`"`)
			assert.Contains(t, w.Body.String(), realDBOperatingHoursReasonB)
			assert.NotContains(t, w.Body.String(), realDBOperatingHoursReasonA)
		})
		t.Run(path+" without shifts:view on the selected clinic returns no data", func(t *testing.T) {
			w := get(newRouter(clinicA), path+"?year_month=2026-09")
			require.Equal(t, http.StatusForbidden, w.Code)
			assert.NotContains(t, w.Body.String(), realDBOperatingHoursReasonA)
			assert.NotContains(t, w.Body.String(), realDBOperatingHoursReasonB)
		})
	}
}
