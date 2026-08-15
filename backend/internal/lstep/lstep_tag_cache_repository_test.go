package lstep

// lstep_tag_cache_repository_test.go — LstepTagCacheRepository 統合テスト。
//
// 保護する不変条件:
//   - UpsertTag は (clinic_id, owner_id, tag_name) で UPSERT する（重複行を作らない、reason 空文字は NULL 保存）。
//   - DeleteTag / DeleteAllByOwner / FindByOwner / FindByOwners / CountByTag / FindOwnerIDsByTag は clinic_id で分離される。
//   - FindByOwners はタグを持たない owner_id をキーとして含めず、ownerIDs 空引数は空mapを即返す。
//   - TagSummary はタグ名・カテゴリ別の飼い主数集計を返し、totalOwnersWithLstep はタグ保持飼い主の重複排除数。
//   - FindOwnersByTag は owners.clinic_id + owners.deleted_at IS NULL を満たす飼い主のみ対象とし、
//     nameQuery 部分一致・ページネーション・タグ一覧付与を行う。
//   - BulkReplaceOwnerTags は既存タグを全削除してから指定タグを一括挿入する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLstepTagCacheTestDB は lstep_tag_cache テーブルを用意する。
// owners テーブルは setupTestDB (ltv_repository_test.go 定義) が既に AutoMigrate + TRUNCATE 済み。
// 本番マイグレーション（001_init.sql）では UNIQUE(clinic_id, owner_id, tag_name) が定義されているが、
// GORM AutoMigrate はモデルに uniqueIndex タグが無いため複合 UNIQUE を再現しない。
// UpsertTag の ON CONFLICT を意味のある形で検証するため、明示的に追加する。
func setupLstepTagCacheTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepTagCache{}))
	db.Exec("TRUNCATE TABLE lstep_tag_cache CASCADE")
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_lstep_tag_cache_conflict
		ON lstep_tag_cache (clinic_id, owner_id, tag_name)`)
	return db
}

func TestLstepTagCacheRepository_UpsertTag(t *testing.T) {
	db := setupLstepTagCacheTestDB(t)
	repo := NewLstepTagCacheRepository(db)
	ctx := context.Background()

	t.Run("新規タグを作成する", func(t *testing.T) {
		require.NoError(t, repo.UpsertTag(ctx, 1, 100, "dormant_365d", "auto", "365日休眠"))

		var stored model.LstepTagCache
		require.NoError(t, db.Where("clinic_id = ? AND owner_id = ? AND tag_name = ?", 1, 100, "dormant_365d").First(&stored).Error)
		assert.Equal(t, "auto", stored.Category)
		require.NotNil(t, stored.Reason)
		assert.Equal(t, "365日休眠", *stored.Reason)
	})

	t.Run("reason 空文字は NULL として保存される", func(t *testing.T) {
		require.NoError(t, repo.UpsertTag(ctx, 1, 101, "manual_tag", "manual", ""))

		var stored model.LstepTagCache
		require.NoError(t, db.Where("clinic_id = ? AND owner_id = ? AND tag_name = ?", 1, 101, "manual_tag").First(&stored).Error)
		assert.Nil(t, stored.Reason)
	})

	t.Run("同一 (clinic_id, owner_id, tag_name) は UPSERT で重複行を作らず category/reason を更新する", func(t *testing.T) {
		require.NoError(t, repo.UpsertTag(ctx, 1, 102, "checkup_followup", "auto", "old-reason"))
		require.NoError(t, repo.UpsertTag(ctx, 1, 102, "checkup_followup", "manual", "new-reason"))

		var count int64
		require.NoError(t, db.Model(&model.LstepTagCache{}).
			Where("clinic_id = ? AND owner_id = ? AND tag_name = ?", 1, 102, "checkup_followup").
			Count(&count).Error)
		assert.Equal(t, int64(1), count)

		var stored model.LstepTagCache
		require.NoError(t, db.Where("clinic_id = ? AND owner_id = ? AND tag_name = ?", 1, 102, "checkup_followup").First(&stored).Error)
		assert.Equal(t, "manual", stored.Category)
		require.NotNil(t, stored.Reason)
		assert.Equal(t, "new-reason", *stored.Reason)
	})
}

func TestLstepTagCacheRepository_DeleteTag(t *testing.T) {
	db := setupLstepTagCacheTestDB(t)
	repo := NewLstepTagCacheRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.UpsertTag(ctx, 1, 200, "tag-a", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, 200, "tag-b", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 2, 200, "tag-a", "auto", ""))

	require.NoError(t, repo.DeleteTag(ctx, 1, 200, "tag-a"))

	t.Run("指定タグのみ削除される", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Model(&model.LstepTagCache{}).Where("clinic_id = ? AND owner_id = ? AND tag_name = ?", 1, 200, "tag-a").Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("同一飼い主の別タグは残る", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Model(&model.LstepTagCache{}).Where("clinic_id = ? AND owner_id = ? AND tag_name = ?", 1, 200, "tag-b").Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックの同名タグは削除されない（clinic_id分離）", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Model(&model.LstepTagCache{}).Where("clinic_id = ? AND owner_id = ? AND tag_name = ?", 2, 200, "tag-a").Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})
}

func TestLstepTagCacheRepository_DeleteAllByOwner(t *testing.T) {
	db := setupLstepTagCacheTestDB(t)
	repo := NewLstepTagCacheRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.UpsertTag(ctx, 1, 300, "tag-a", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, 300, "tag-b", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, 301, "tag-a", "auto", ""))

	require.NoError(t, repo.DeleteAllByOwner(ctx, 1, 300))

	t.Run("対象飼い主の全タグが削除される", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Model(&model.LstepTagCache{}).Where("clinic_id = ? AND owner_id = ?", 1, 300).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("別飼い主のタグは残る", func(t *testing.T) {
		var count int64
		require.NoError(t, db.Model(&model.LstepTagCache{}).Where("clinic_id = ? AND owner_id = ?", 1, 301).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})
}

func TestLstepTagCacheRepository_FindByOwner(t *testing.T) {
	db := setupLstepTagCacheTestDB(t)
	repo := NewLstepTagCacheRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.UpsertTag(ctx, 1, 400, "tag-a", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, 400, "tag-b", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 2, 400, "tag-c", "auto", ""))

	t.Run("対象飼い主のタグ一覧を返す", func(t *testing.T) {
		records, err := repo.FindByOwner(ctx, 1, 400)
		require.NoError(t, err)
		require.Len(t, records, 2)
	})

	t.Run("別クリニックの同一owner_idは含まれない（clinic_id分離）", func(t *testing.T) {
		records, err := repo.FindByOwner(ctx, 2, 400)
		require.NoError(t, err)
		require.Len(t, records, 1)
		assert.Equal(t, "tag-c", records[0].TagName)
	})

	t.Run("該当なしは空スライスを返す", func(t *testing.T) {
		records, err := repo.FindByOwner(ctx, 1, 999)
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestLstepTagCacheRepository_FindByOwners(t *testing.T) {
	db := setupLstepTagCacheTestDB(t)
	repo := NewLstepTagCacheRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.UpsertTag(ctx, 1, 400, "tag-a", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, 400, "tag-b", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, 401, "tag-c", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 2, 400, "tag-d", "auto", ""))

	t.Run("複数owner_idのタグをowner_id別にまとめて返す", func(t *testing.T) {
		result, err := repo.FindByOwners(ctx, 1, []uint64{400, 401, 999})
		require.NoError(t, err)
		require.Len(t, result[400], 2)
		require.Len(t, result[401], 1)
		assert.Equal(t, "tag-c", result[401][0].TagName)
		_, hasMissing := result[999]
		assert.False(t, hasMissing, "タグなしowner_idはキーとして存在しない")
	})

	t.Run("別クリニックのタグは含まれない（clinic_id分離）", func(t *testing.T) {
		result, err := repo.FindByOwners(ctx, 2, []uint64{400})
		require.NoError(t, err)
		require.Len(t, result[400], 1)
		assert.Equal(t, "tag-d", result[400][0].TagName)
	})

	t.Run("ownerIDsが空の場合は空mapを即返す", func(t *testing.T) {
		result, err := repo.FindByOwners(ctx, 1, []uint64{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestLstepTagCacheRepository_TagSummary(t *testing.T) {
	db := setupLstepTagCacheTestDB(t)
	repo := NewLstepTagCacheRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.UpsertTag(ctx, 1, 600, "dormant_365d", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, 601, "dormant_365d", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, 601, "manual_note", "manual", ""))
	require.NoError(t, repo.UpsertTag(ctx, 2, 602, "dormant_365d", "auto", "")) // 別クリニック

	rows, total, err := repo.TagSummary(ctx, 1)
	require.NoError(t, err)

	t.Run("タグ名・カテゴリ別の飼い主数を集計する", func(t *testing.T) {
		require.Len(t, rows, 2)
		byTag := make(map[string]TagSummaryRow, len(rows))
		for _, r := range rows {
			byTag[r.TagName] = r
		}
		require.Contains(t, byTag, "dormant_365d")
		assert.Equal(t, int64(2), byTag["dormant_365d"].OwnerCount)
		require.Contains(t, byTag, "manual_note")
		assert.Equal(t, int64(1), byTag["manual_note"].OwnerCount)
	})

	t.Run("totalOwnersWithLstep はタグ保持飼い主のユニーク数", func(t *testing.T) {
		assert.Equal(t, int64(2), total, "owner 600, 601 の2名（601は2タグ持つが重複排除）")
	})

	t.Run("別クリニックは集計に含まれない", func(t *testing.T) {
		otherRows, otherTotal, err := repo.TagSummary(ctx, 2)
		require.NoError(t, err)
		require.Len(t, otherRows, 1)
		assert.Equal(t, int64(1), otherTotal)
	})
}

func TestLstepTagCacheRepository_FindOwnersByTag(t *testing.T) {
	db := setupLstepTagCacheTestDB(t)
	repo := NewLstepTagCacheRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	tanaka := testdb.MakeTestOwner(t, db, clinicA, "田中太郎")
	yamada := testdb.MakeTestOwner(t, db, clinicA, "山田花子")
	deletedOwner := testdb.MakeTestOwner(t, db, clinicA, "削除済み飼い主")
	otherClinicOwner := testdb.MakeTestOwner(t, db, clinicB, "別クリニック飼い主")

	require.NoError(t, repo.UpsertTag(ctx, clinicA, tanaka.ID, "dormant_365d", "auto", "reason-tanaka"))
	require.NoError(t, repo.UpsertTag(ctx, clinicA, tanaka.ID, "manual_note", "manual", ""))
	require.NoError(t, repo.UpsertTag(ctx, clinicA, yamada.ID, "dormant_365d", "auto", "reason-yamada"))
	require.NoError(t, repo.UpsertTag(ctx, clinicA, deletedOwner.ID, "dormant_365d", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, clinicB, otherClinicOwner.ID, "dormant_365d", "auto", ""))

	// ソフトデリート
	require.NoError(t, db.Delete(&model.Owner{}, deletedOwner.ID).Error)

	t.Run("該当タグを持つ飼い主をタグ一覧付きで返す（ソフトデリート・他クリニック除外）", func(t *testing.T) {
		results, total, err := repo.FindOwnersByTag(ctx, clinicA, "dormant_365d", "", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, results, 2)

		byID := make(map[uint64]TagOwnerRow, len(results))
		for _, r := range results {
			byID[r.OwnerID] = r
		}
		require.Contains(t, byID, tanaka.ID)
		assert.ElementsMatch(t, []string{"dormant_365d", "manual_note"}, byID[tanaka.ID].Tags)
		require.NotNil(t, byID[tanaka.ID].Reason)
		assert.Equal(t, "reason-tanaka", *byID[tanaka.ID].Reason)
		assert.NotContains(t, byID, deletedOwner.ID, "ソフトデリート済み飼い主は除外される")
	})

	t.Run("nameQuery で部分一致フィルタする", func(t *testing.T) {
		results, total, err := repo.FindOwnersByTag(ctx, clinicA, "dormant_365d", "田中", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, results, 1)
		assert.Equal(t, tanaka.ID, results[0].OwnerID)
	})

	t.Run("limit/offset でページネーションする", func(t *testing.T) {
		page1, total, err := repo.FindOwnersByTag(ctx, clinicA, "dormant_365d", "", 0, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, page1, 1)

		page2, _, err := repo.FindOwnersByTag(ctx, clinicA, "dormant_365d", "", 1, 1)
		require.NoError(t, err)
		require.Len(t, page2, 1)
		assert.NotEqual(t, page1[0].OwnerID, page2[0].OwnerID, "ページ間で重複しない")
	})

	t.Run("該当なしは空スライスを返す", func(t *testing.T) {
		results, total, err := repo.FindOwnersByTag(ctx, clinicA, "nonexistent-tag", "", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, results)
	})
}

func TestLstepTagCacheRepository_FindOwnerIDsByTag(t *testing.T) {
	db := setupLstepTagCacheTestDB(t)
	repo := NewLstepTagCacheRepository(db)
	ctx := context.Background()

	ownerA1 := &model.Owner{ClinicID: 1, Name: "休眠A1"}
	ownerA2 := &model.Owner{ClinicID: 1, Name: "休眠A2"}
	ownerB := &model.Owner{ClinicID: 2, Name: "休眠B"}
	require.NoError(t, db.WithContext(ctx).Create([]*model.Owner{ownerA1, ownerA2, ownerB}).Error)
	require.NoError(t, repo.UpsertTag(ctx, 1, ownerA1.ID, "dormant_365d", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 1, ownerA2.ID, "dormant_365d", "auto", ""))
	require.NoError(t, repo.UpsertTag(ctx, 2, ownerB.ID, "dormant_365d", "auto", "")) // 別クリニック
	require.NoError(t, repo.UpsertTag(ctx, 1, ownerB.ID, "dormant_365d", "auto", "")) // 不整合行

	t.Run("該当タグを持つ飼い主IDを重複なく返す", func(t *testing.T) {
		ids, err := repo.FindOwnerIDsByTag(ctx, 1, "dormant_365d")
		require.NoError(t, err)
		assert.ElementsMatch(t, []uint64{ownerA1.ID, ownerA2.ID}, ids)
	})

	t.Run("別クリニックのIDは含まれない", func(t *testing.T) {
		ids, err := repo.FindOwnerIDsByTag(ctx, 2, "dormant_365d")
		require.NoError(t, err)
		assert.ElementsMatch(t, []uint64{ownerB.ID}, ids)
	})

	t.Run("該当なしは空スライスを返す", func(t *testing.T) {
		ids, err := repo.FindOwnerIDsByTag(ctx, 1, "nonexistent-tag")
		require.NoError(t, err)
		assert.Empty(t, ids)
	})
}
