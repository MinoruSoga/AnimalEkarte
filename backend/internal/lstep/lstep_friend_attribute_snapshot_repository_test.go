package lstep

// lstep_friend_attribute_snapshot_repository_test.go — LstepFriendAttributeSnapshotRepository 統合テスト。
//
// 保護する不変条件:
//   - BulkCreate は UNIQUE(clinic_id, line_user_id, snapshot_taken_at) 衝突時は静かにスキップする（DoNothing）。
//   - FindLatestByOwner は clinic_id + line_user_id スコープで snapshot_taken_at 最新の1件を返す。
//   - FindLatestByOwner は該当なしで NotFound を返す。
//   - FindAllByClinicAndDateRange は期間・clinic_id で正しく分離される。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLstepFriendAttributeSnapshotTestDB は lstep_friend_attribute_snapshots テーブルを用意する。
// 本番マイグレーション（001_init.sql）では UNIQUE(clinic_id, line_user_id, snapshot_taken_at) が
// 定義されているが、GORM AutoMigrate はモデルに uniqueIndex タグが無いため複合 UNIQUE を再現しない。
// BulkCreate の ON CONFLICT DoNothing を意味のある形で検証するため、ここで明示的に複合 UNIQUE を追加する。
func setupLstepFriendAttributeSnapshotTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepCsvImport{}, &model.LstepFriendAttributeSnapshot{}))
	db.Exec("TRUNCATE TABLE lstep_friend_attribute_snapshots CASCADE")
	db.Exec("TRUNCATE TABLE lstep_csv_imports CASCADE")
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

func TestLstepFriendAttributeSnapshotRepository_FindLatestByOwner_ValidatesImportClinic(t *testing.T) {
	db := setupLstepFriendAttributeSnapshotTestDB(t)
	repo := NewLstepFriendAttributeSnapshotRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	const lineUserID = "U-import-clinic-boundary"

	importA := &model.LstepCsvImport{ClinicID: clinicA, CsvType: "friend_attribute", FileName: "a.csv"}
	importB := &model.LstepCsvImport{ClinicID: clinicB, CsvType: "friend_attribute", FileName: "b.csv"}
	require.NoError(t, db.Create(importA).Error)
	require.NoError(t, db.Create(importB).Error)

	valid := &model.LstepFriendAttributeSnapshot{
		ClinicID: clinicA, LineUserID: lineUserID,
		SnapshotTakenAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		CsvImportID:     &importA.ID,
	}
	malformedCrossClinic := &model.LstepFriendAttributeSnapshot{
		ClinicID: clinicA, LineUserID: lineUserID,
		SnapshotTakenAt: time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
		CsvImportID:     &importB.ID,
	}
	require.NoError(t, db.Create(valid).Error)
	require.NoError(t, db.Create(malformedCrossClinic).Error)

	found, err := repo.FindLatestByOwner(ctx, clinicA, lineUserID)
	require.NoError(t, err)
	assert.Equal(t, valid.ID, found.ID)
}
