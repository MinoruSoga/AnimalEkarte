package support

// realdb_bug_reports_global_list_test.go — バグ報告共有ボードの実DB検証
//
// Proves GET /api/v1/support/bug-reports returns rows from ALL clinics for any
// authenticated staff member through the real repository + service + HTTP
// handler path. Global read visibility is a deliberate product decision
// (2026-10): reports are product feedback, not clinic business data.
// Mutations (status update / plane ticket / delete) are the opposite — they are
// scoped to the report's own clinic_id since the 2026-10 security review, and a
// cross-clinic mutation target returns the same NotFound as a missing report.
// Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestRealDB_SupportBugReportsGlobalBoard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.SupportBugReport{}))
	testdb.Truncate(t, db, "support_bug_reports")

	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3support realDB global")
	ctx := context.Background()

	repo := NewRepository(db)
	svc := NewService(repo, nil)

	const titleA = "D3support-realdb-global-A"
	const titleB = "D3support-realdb-global-B"
	reportA, err := svc.Create(ctx, fx.ClinicA, fx.StaffID, CreateBugReportInput{Title: titleA})
	require.NoError(t, err)
	reportB, err := svc.Create(ctx, fx.ClinicB, fx.StaffID, CreateBugReportInput{Title: titleB})
	require.NoError(t, err)

	handler := NewHandler(svc, nil, nil, nil, nil)

	// 選択医院 A のスタッフコンテキスト。権限チェッカーは意図的に設定しない
	//（バグ報告は権限モデルの外側 = 全スタッフに開放）。
	configureClinicA := func(c *gin.Context) {
		c.Set("clinic_id", fmt.Sprintf("%d", fx.ClinicA))
		c.Set("user_id", fmt.Sprintf("%d", fx.StaffID))
	}
	idParam := func(c *gin.Context, id uint64) {
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", id)}}
	}

	t.Run("staff_in_clinic_A_lists_reports_from_all_clinics", func(t *testing.T) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/support/bug-reports", configureClinicA)
		handler.ListBugReports(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), titleA)
		assert.Contains(t, w.Body.String(), titleB)
		assert.Contains(t, w.Body.String(), `"clinic_name":"`+fx.ClinicAName+`"`)
		assert.Contains(t, w.Body.String(), `"clinic_name":"`+fx.ClinicBName+`"`)
	})

	t.Run("staff_in_clinic_A_updates_status_of_own_report", func(t *testing.T) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodPatch,
			fmt.Sprintf("/api/v1/support/bug-reports/%d/status", reportA.ID), configureClinicA)
		c.Request.Body = io.NopCloser(bytes.NewBufferString(`{"status":"resolved"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		idParam(c, reportA.ID)
		handler.UpdateBugReportStatus(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), `"status":"resolved"`)
	})

	t.Run("staff_in_clinic_A_cannot_update_status_of_clinic_B_report", func(t *testing.T) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodPatch,
			fmt.Sprintf("/api/v1/support/bug-reports/%d/status", reportB.ID), configureClinicA)
		c.Request.Body = io.NopCloser(bytes.NewBufferString(`{"status":"resolved"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		idParam(c, reportB.ID)
		handler.UpdateBugReportStatus(c)
		// 他医院の報告への操作は「操作対象として存在しない」= 未存在と同じ 404。
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())

		// 報告 B は変更されず open のまま残る（共有一覧にも引き続き出る）。
		var b model.SupportBugReport
		require.NoError(t, db.First(&b, reportB.ID).Error)
		assert.Equal(t, model.SupportBugReportStatusOpen, b.Status)
	})

	t.Run("staff_in_clinic_A_cannot_create_plane_ticket_for_clinic_B_report", func(t *testing.T) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodPost,
			fmt.Sprintf("/api/v1/support/bug-reports/%d/plane-ticket", reportB.ID), configureClinicA)
		idParam(c, reportB.ID)
		handler.CreatePlaneTicket(c)
		// スコープ判定は Plane 連携未設定チェック（501）より先に効いて 404。
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("staff_in_clinic_A_cannot_delete_clinic_B_report", func(t *testing.T) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodDelete,
			fmt.Sprintf("/api/v1/support/bug-reports/%d", reportB.ID), configureClinicA)
		idParam(c, reportB.ID)
		handler.DeleteBugReport(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())

		// 共有一覧からも消えていない（削除されていない）。
		c2, w2 := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/support/bug-reports", configureClinicA)
		handler.ListBugReports(c2)
		require.Equal(t, http.StatusOK, w2.Code, w2.Body.String())
		assert.Contains(t, w2.Body.String(), titleB)
	})

	t.Run("staff_in_clinic_A_deletes_own_report", func(t *testing.T) {
		c, _ := testdb.NewHTTPTestContext(t, http.MethodDelete,
			fmt.Sprintf("/api/v1/support/bug-reports/%d", reportA.ID), configureClinicA)
		idParam(c, reportA.ID)
		handler.DeleteBugReport(c)
		require.Equal(t, http.StatusNoContent, c.Writer.Status())

		// 論理削除後は共有一覧から消え、他医院の報告は残る
		c2, w2 := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/support/bug-reports", configureClinicA)
		handler.ListBugReports(c2)
		require.Equal(t, http.StatusOK, w2.Code, w2.Body.String())
		assert.NotContains(t, w2.Body.String(), titleA)
		assert.Contains(t, w2.Body.String(), titleB)
	})
}
