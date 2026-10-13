package support

// realdb_create_plane_test.go — Create 後の Plane 自動起票を実 DB で検証
//
// syncPlaneTicket（EMR-340 で復帰した保存後 best-effort 起票）が、実 repository
// 経由で「保存された新規報告のみ」を対象に起票・claim・失敗記録を行うことを
// 検証する。TicketCreator は fn-field mock で Plane/Discord への実通信は発生しない。
// 起票前から存在する報告への backfill がないこと、他医院スコープの操作が
// NotFound/claim 失敗で行を変えないことも併せて固定する。
// Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestRealDB_CreateSyncsPlaneTicket(t *testing.T) {
	issue := &PlaneIssue{ID: "realdb-issue-1", URL: "https://app.plane.so/ws/browse/EMR-1/"}

	tests := []struct {
		name        string
		creator     *mockTicketCreator
		wantIssueID string // 非空 = claim 成功で永続化されるべき issue ID
		wantSyncErr string // 非空 = plane_sync_error に記録されるべき sanitize 文言
	}{
		{
			name:        "auto create persists claimed ticket only on the new report",
			creator:     &mockTicketCreator{result: issue},
			wantIssueID: issue.ID,
		},
		{
			name:        "upstream failure persists sanitized sync error only on the new report",
			creator:     &mockTicketCreator{err: &planeUpstreamError{status: 500}},
			wantSyncErr: "plane api error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testdb.SetupTestDB(t)
			require.NoError(t, testdb.EnsureAutoMigrated(db, &model.SupportBugReport{}))
			testdb.Truncate(t, db, "support_bug_reports")
			fx := testdb.SeedDualClinicGrantFixture(t, db, "EMR340 realDB create plane")
			ctx := context.Background()
			repo := NewRepository(db)

			// clinic A に自動起票導入前の既存報告を直接 seed（backfill 対象外となる行）。
			legacy := &model.SupportBugReport{
				ClinicID:        fx.ClinicA,
				ReporterStaffID: fx.StaffID,
				Title:           "legacy-unissued",
			}
			require.NoError(t, repo.Create(ctx, legacy))

			var gotReport *model.SupportBugReport
			tt.creator.fn = func(_ context.Context, r *model.SupportBugReport) { gotReport = r }
			svc := NewService(repo, tt.creator, nil)

			reportB, err := svc.Create(ctx, fx.ClinicB, fx.StaffID, CreateBugReportInput{Title: "auto-create-b"})
			require.NoError(t, err)
			require.NotNil(t, reportB)

			// creator は保存済み report をちょうど 1 回受け取る。
			assert.Equal(t, 1, tt.creator.calls)
			require.NotNil(t, gotReport)
			assert.Equal(t, reportB.ID, gotReport.ID)
			assert.Equal(t, fx.ClinicB, gotReport.ClinicID)
			assert.Equal(t, fx.StaffID, gotReport.ReporterStaffID)

			// DB 永続化の確認: B の報告だけが起票結果を持つ。
			assertPersistedB := func() *model.SupportBugReport {
				persisted, err := repo.FindByIDForClinic(ctx, fx.ClinicB, reportB.ID)
				require.NoError(t, err)
				if tt.wantIssueID != "" {
					require.NotNil(t, persisted.PlaneIssueID)
					assert.Equal(t, tt.wantIssueID, *persisted.PlaneIssueID)
					require.NotNil(t, persisted.PlaneIssueURL)
					assert.Equal(t, issue.URL, *persisted.PlaneIssueURL)
					assert.Nil(t, persisted.PlaneSyncError)
				} else {
					assert.Nil(t, persisted.PlaneIssueID)
					assert.Nil(t, persisted.PlaneIssueURL)
					require.NotNil(t, persisted.PlaneSyncError)
					assert.Equal(t, tt.wantSyncErr, *persisted.PlaneSyncError)
				}
				return persisted
			}
			assertPersistedB()

			// 既存（clinic A）報告は backfill されない — 未起票・sync error なしのまま。
			persistedA, err := repo.FindByIDForClinic(ctx, fx.ClinicA, legacy.ID)
			require.NoError(t, err)
			assert.Nil(t, persistedA.PlaneIssueID)
			assert.Nil(t, persistedA.PlaneIssueURL)
			assert.Nil(t, persistedA.PlaneSyncError)

			// 他医院スコープの手動再送は NotFound で上流を呼ばない。
			_, err = svc.EnsurePlaneTicket(ctx, fx.ClinicA, reportB.ID)
			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assert.Equal(t, 1, tt.creator.calls)

			// 他医院スコープでの claim・失敗記録も行を変えない。
			claimed, err := repo.SetPlaneTicket(ctx, fx.ClinicA, reportB.ID, "foreign-issue", "u")
			require.NoError(t, err)
			assert.False(t, claimed)
			written, err := repo.SetPlaneSyncError(ctx, fx.ClinicA, reportB.ID, "foreign-error")
			assert.False(t, written)
			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assertPersistedB() // B の行は変化なし

			if tt.wantSyncErr != "" {
				// 自動起票失敗の報告は自医院の手動再送で起票でき、旧失敗理由が消える。
				retryCreator := &mockTicketCreator{
					result: &PlaneIssue{ID: "realdb-issue-2", URL: "https://app.plane.so/ws/browse/EMR-2/"},
				}
				retrySvc := NewService(repo, retryCreator, nil)

				_, err = retrySvc.EnsurePlaneTicket(ctx, fx.ClinicB, reportB.ID)
				require.NoError(t, err)
				assert.Equal(t, 1, retryCreator.calls)

				retried, err := repo.FindByIDForClinic(ctx, fx.ClinicB, reportB.ID)
				require.NoError(t, err)
				require.NotNil(t, retried.PlaneIssueID)
				assert.Equal(t, "realdb-issue-2", *retried.PlaneIssueID)
				require.NotNil(t, retried.PlaneIssueURL)
				assert.Nil(t, retried.PlaneSyncError)

				// 起票済みの再送は冪等 — creator を呼ばない。
				_, err = retrySvc.EnsurePlaneTicket(ctx, fx.ClinicB, reportB.ID)
				require.NoError(t, err)
				assert.Equal(t, 1, retryCreator.calls)
			}
		})
	}
}
