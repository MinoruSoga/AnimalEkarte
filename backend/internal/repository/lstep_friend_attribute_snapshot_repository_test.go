package repository

// lstep_friend_attribute_snapshot_repository_test.go — LstepFriendAttributeSnapshotRepository 統合テスト。
//
// 保護する不変条件:
//   - BulkCreate は UNIQUE(clinic_id, line_user_id, snapshot_taken_at) 衝突時は静かにスキップする（DoNothing）。
//   - FindLatestByOwner は clinic_id + line_user_id スコープで snapshot_taken_at 最新の1件を返す。
//   - FindLatestByOwner は該当なしで NotFound を返す。
//   - ListByClinicAndDateRange は期間・clinic_id で正しく分離される。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupLstepFriendAttributeSnapshotTestDB は lstep_friend_attribute_snapshots テーブルを用意する。
// 本番マイグレーション（001_init.sql）では UNIQUE(clinic_id, line_user_id, snapshot_taken_at) が
// 定義されているが、GORM AutoMigrate はモデルに uniqueIndex タグが無いため複合 UNIQUE を再現しない。
// BulkCreate の ON CONFLICT DoNothing を意味のある形で検証するため、ここで明示的に複合 UNIQUE を追加する。
func setupLstepFriendAttributeSnapshotTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.LstepFriendAttributeSnapshot{}))
	db.Exec("TRUNCATE TABLE lstep_friend_attribute_snapshots CASCADE")
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_lstep_friend_snapshot_conflict
		ON lstep_friend_attribute_snapshots (clinic_id, line_user_id, snapshot_taken_at)`)
	return db
}

// makeFriendAttributeSnapshot はテスト用スナップショットを作成して返す。
func makeFriendAttributeSnapshot(t *testing.T, db *gorm.DB, clinicID uint64, lineUserID string, takenAt time.Time) *model.LstepFriendAttributeSnapshot {
	t.Helper()
	snapshot := &model.LstepFriendAttributeSnapshot{
		ClinicID:        clinicID,
		LineUserID:      lineUserID,
		SnapshotTakenAt: takenAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(snapshot).Error)
	return snapshot
}

func TestLstepFriendAttributeSnapshotRepository_Create(t *testing.T) {
	db := setupLstepFriendAttributeSnapshotTestDB(t)
	repo := NewLstepFriendAttributeSnapshotRepository(db)
	ctx := context.Background()

	takenAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	displayName := "テスト太郎"
	snapshot := &model.LstepFriendAttributeSnapshot{
		ClinicID:        1,
		LineUserID:      "U-line-001",
		DisplayName:     &displayName,
		SnapshotTakenAt: takenAt,
	}
	require.NoError(t, repo.Create(ctx, snapshot))
	assert.NotZero(t, snapshot.ID)

	var stored model.LstepFriendAttributeSnapshot
	require.NoError(t, db.First(&stored, snapshot.ID).Error)
	assert.Equal(t, "U-line-001", stored.LineUserID)
	require.NotNil(t, stored.DisplayName)
	assert.Equal(t, displayName, *stored.DisplayName)
}

func TestLstepFriendAttributeSnapshotRepository_BulkCreate(t *testing.T) {
	db := setupLstepFriendAttributeSnapshotTestDB(t)
	repo := NewLstepFriendAttributeSnapshotRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	takenAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)

	t.Run("空スライスは何もせず成功する", func(t *testing.T) {
		require.NoError(t, repo.BulkCreate(ctx, nil))
	})

	t.Run("複数件を一括作成できる", func(t *testing.T) {
		snapshots := []*model.LstepFriendAttributeSnapshot{
			{ClinicID: clinicA, LineUserID: "U-bulk-001", SnapshotTakenAt: takenAt},
			{ClinicID: clinicA, LineUserID: "U-bulk-002", SnapshotTakenAt: takenAt},
		}
		require.NoError(t, repo.BulkCreate(ctx, snapshots))

		var count int64
		require.NoError(t, db.Model(&model.LstepFriendAttributeSnapshot{}).
			Where("clinic_id = ? AND line_user_id IN ?", clinicA, []string{"U-bulk-001", "U-bulk-002"}).
			Count(&count).Error)
		assert.Equal(t, int64(2), count)
	})

	t.Run("UNIQUE 制約衝突時はスキップされ既存行は残る", func(t *testing.T) {
		displayName1 := "オリジナル"
		existing := makeFriendAttributeSnapshot(t, db, clinicA, "U-conflict-001", takenAt)
		require.NoError(t, db.Model(existing).Update("display_name", displayName1).Error)

		displayName2 := "衝突後の値（無視されるべき）"
		conflicting := []*model.LstepFriendAttributeSnapshot{
			{ClinicID: clinicA, LineUserID: "U-conflict-001", SnapshotTakenAt: takenAt, DisplayName: &displayName2},
		}
		require.NoError(t, repo.BulkCreate(ctx, conflicting), "DoNothing 衝突はエラーにならないべき")

		var count int64
		require.NoError(t, db.Model(&model.LstepFriendAttributeSnapshot{}).
			Where("clinic_id = ? AND line_user_id = ? AND snapshot_taken_at = ?", clinicA, "U-conflict-001", takenAt).
			Count(&count).Error)
		assert.Equal(t, int64(1), count, "衝突時は新規行が追加されず既存の1行のみであるべき")

		var stored model.LstepFriendAttributeSnapshot
		require.NoError(t, db.Where("id = ?", existing.ID).First(&stored).Error)
		require.NotNil(t, stored.DisplayName)
		assert.Equal(t, displayName1, *stored.DisplayName, "既存行は上書きされないべき（DoNothing）")
	})
}

func TestLstepFriendAttributeSnapshotRepository_FindLatestByOwner(t *testing.T) {
	db := setupLstepFriendAttributeSnapshotTestDB(t)
	repo := NewLstepFriendAttributeSnapshotRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	const lineUserID = "U-latest-001"
	older := makeFriendAttributeSnapshot(t, db, clinicA, lineUserID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	newer := makeFriendAttributeSnapshot(t, db, clinicA, lineUserID, time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
	_ = older
	// 別クリニックの同一 LINE ユーザーは対象外（clinic_id 分離）
	makeFriendAttributeSnapshot(t, db, clinicB, lineUserID, time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC))

	t.Run("clinic_id + line_user_id スコープで最新の1件を返す", func(t *testing.T) {
		found, err := repo.FindLatestByOwner(ctx, clinicA, lineUserID)
		require.NoError(t, err)
		assert.Equal(t, newer.ID, found.ID)
	})

	t.Run("別クリニックのみに存在する line_user_id は NotFound を返す（clinic_id 分離）", func(t *testing.T) {
		_, err := repo.FindLatestByOwner(ctx, clinicA, "U-not-in-clinic-a")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("未登録の line_user_id は NotFound を返す", func(t *testing.T) {
		_, err := repo.FindLatestByOwner(ctx, clinicA, "U-never-registered")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

func TestLstepFriendAttributeSnapshotRepository_ListByClinicAndDateRange(t *testing.T) {
	db := setupLstepFriendAttributeSnapshotTestDB(t)
	repo := NewLstepFriendAttributeSnapshotRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	inRange1 := makeFriendAttributeSnapshot(t, db, clinicA, "U-range-001", time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC))
	inRange2 := makeFriendAttributeSnapshot(t, db, clinicA, "U-range-002", time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC))
	makeFriendAttributeSnapshot(t, db, clinicA, "U-before-range", time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)) // 範囲外（前）
	makeFriendAttributeSnapshot(t, db, clinicA, "U-after-range", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))   // 範囲外（後、until は排他的）
	makeFriendAttributeSnapshot(t, db, clinicB, "U-other-clinic", time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)) // 別クリニック

	snapshots, err := repo.ListByClinicAndDateRange(ctx, clinicA, since, until)
	require.NoError(t, err)
	require.Len(t, snapshots, 2)
	assert.Equal(t, inRange1.ID, snapshots[0].ID, "snapshot_taken_at 昇順で返るべき")
	assert.Equal(t, inRange2.ID, snapshots[1].ID)
}
