package support

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 support package (1 route)
//
// Proves GET /api/v1/support/bug-reports returns only the selected clinic's
// rows through the real repository + service + permission middleware + HTTP
// handler path. Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// realDBSupportRequirePermission mirrors the composition-root wiring:
// auth.RequirePermission semantics for the selected clinic, evaluated through
// the per-request ClinicPermissionChecker.
func realDBSupportRequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !httpapi.RequireSelectedClinicGrant(c, resource, action) {
			c.Abort()
			return
		}
		c.Next()
	}
}

func TestRealDB_SupportBugReportsSelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.SupportBugReport{}))
	testdb.Truncate(t, db, "support_bug_reports")

	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3support realDB")
	ctx := context.Background()

	repo := NewRepository(db)
	svc := NewService(repo)

	const titleA = "D3support-realdb-report-A"
	const titleB = "D3support-realdb-report-B"
	_, err := svc.Create(ctx, fx.ClinicA, fx.StaffID, CreateBugReportInput{Title: titleA})
	require.NoError(t, err)
	_, err = svc.Create(ctx, fx.ClinicB, fx.StaffID, CreateBugReportInput{Title: titleB})
	require.NoError(t, err)

	handler := NewHandler(svc, nil, nil, realDBSupportRequirePermission, nil, nil)
	res := string(model.ResourceHospitalSettings)
	const listPath = "/api/v1/support/bug-reports"

	t.Run("list_grantA_only_403", func(t *testing.T) {
		configure := testdb.ConfigureSelectedClinicBGrant(fx, res, fx.ClinicA)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, listPath, configure)
		realDBSupportRequirePermission(res, "view")(c)
		require.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("list_grantB_returns_only_B_rows", func(t *testing.T) {
		configure := testdb.ConfigureSelectedClinicBGrant(fx, res, fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, listPath, configure)
		realDBSupportRequirePermission(res, "view")(c)
		require.False(t, c.IsAborted())
		handler.ListBugReports(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), titleB)
		assert.NotContains(t, w.Body.String(), titleA)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.ClinicA})
	})
}
