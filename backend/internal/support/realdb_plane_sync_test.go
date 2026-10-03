package support

// realdb_plane_sync_test.go — plane_sync 同期対象抽出の実DB検証
//
// ListOpenWithPlaneTicket は status=open かつ plane_issue_id IS NOT NULL の
// 報告だけを全医院横断で返す（resolved / 未起票 / soft delete 済みは除外）。
// 返却行の clinic_id は後続の UpdateStatus を報告元医院スコープで行うための
// 必須フィールド。Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestRealDB_ListOpenWithPlaneTicket_FiltersAndBounds(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.SupportBugReport{}))
	testdb.Truncate(t, db, "support_bug_reports")

	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3support realDB plane sync")
	ctx := context.Background()
	repo := NewRepository(db)

	mkReport := func(clinicID uint64, title string) *model.SupportBugReport {
		r := &model.SupportBugReport{
			ClinicID:        clinicID,
			ReporterStaffID: fx.StaffID,
			Title:           title,
		}
		require.NoError(t, repo.Create(ctx, r))
		return r
	}

	noTicket := mkReport(fx.ClinicA, "open without ticket")          // 除外: 未起票
	doneA := mkReport(fx.ClinicA, "open with ticket (clinic A)")     // 対象
	doneB := mkReport(fx.ClinicB, "open with ticket (clinic B)")     // 対象: 他医院も横断
	resolvedTicketed := mkReport(fx.ClinicB, "resolved with ticket") // 除外: resolved
	deletedTicketed := mkReport(fx.ClinicA, "deleted with ticket")   // 除外: soft delete

	for _, ticketed := range []struct {
		clinicID uint64
		id       uint64
		issueID  string
	}{
		{fx.ClinicA, doneA.ID, "issue-A"},
		{fx.ClinicB, doneB.ID, "issue-B"},
		{fx.ClinicB, resolvedTicketed.ID, "issue-C"},
		{fx.ClinicA, deletedTicketed.ID, "issue-D"},
	} {
		claimed, err := repo.SetPlaneTicket(ctx, ticketed.clinicID, ticketed.id, ticketed.issueID, "")
		require.NoError(t, err)
		require.True(t, claimed)
	}
	require.NoError(t, repo.UpdateStatus(ctx, fx.ClinicB, resolvedTicketed.ID, model.SupportBugReportStatusResolved))
	require.NoError(t, repo.SoftDeleteBugReport(ctx, fx.ClinicA, deletedTicketed.ID))

	t.Run("returns only open reports with plane tickets across clinics", func(t *testing.T) {
		got, err := repo.ListOpenWithPlaneTicket(ctx, 0)
		require.NoError(t, err)

		ids := make([]uint64, 0, len(got))
		byID := make(map[uint64]model.SupportBugReport, len(got))
		for _, r := range got {
			ids = append(ids, r.ID)
			byID[r.ID] = r
		}
		require.Contains(t, ids, doneA.ID)
		require.Contains(t, ids, doneB.ID)
		assert.NotContains(t, ids, noTicket.ID)
		assert.NotContains(t, ids, resolvedTicketed.ID)
		assert.NotContains(t, ids, deletedTicketed.ID)

		// clinic_id は後続 UpdateStatus を報告元医院スコープで行うため必須。
		assert.Equal(t, fx.ClinicA, byID[doneA.ID].ClinicID)
		assert.Equal(t, fx.ClinicB, byID[doneB.ID].ClinicID)
	})

	t.Run("limit bounds the result in stable id order", func(t *testing.T) {
		got, err := repo.ListOpenWithPlaneTicket(ctx, 1)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, doneA.ID, got[0].ID)
	})
}
