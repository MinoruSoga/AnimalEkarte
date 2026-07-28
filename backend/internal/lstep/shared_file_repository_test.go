package lstep

// repository_test.go — Repository の統合テスト（内部カバレッジ向上）。
//
// 対象: Create / FindByID / FindAll / Delete / FindExpired
// 検証観点: 正常系、clinic_id 隔離、deleted_at IS NULL 除外、NotFound ラップ、期限切れ抽出。
//
// 実装上の注意（テストで明示的に固定する）:
//   model.SharedFile.DeletedAt は gorm.DeletedAt のため、GORM の自動ソフトデリート機構が
//   Delete() 発行時に発火する（UPDATE ... SET deleted_at = ... であり物理 DELETE ではない）。
//   別クリニックからの Delete は RowsAffected == 0 を検知して NotFound を返す
//   （appointment_admin_repository.go の SoftDelete と同型パターン）。

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

// setupSharedFileTestDB は shared_files テーブルを整備する（依存する他テーブルなし）。
func setupSharedFileTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.SharedFile{}))
	db.Exec("TRUNCATE TABLE shared_files CASCADE")
	return db
}

func makeSharedFile(t *testing.T, repo SharedFileRepository, clinicID uint64, fileName string) *model.SharedFile {
	t.Helper()
	f := &model.SharedFile{
		ClinicID:   clinicID,
		UploadedBy: 1,
		FileType:   model.SharedFileTypePDF,
		FileName:   fileName,
		FileKey:    "keys/" + fileName,
		FileSize:   1024,
		Purpose:    model.SharedFilePurposeOther,
	}
	require.NoError(t, repo.Create(context.Background(), f))
	require.NotZero(t, f.ID)
	return f
}

// TestSharedFileRepository_Create_FindByID は作成と単件取得（clinic_id 隔離・NotFound）を検証する。
func TestSharedFileRepository_Create_FindByID(t *testing.T) {
	db := setupSharedFileTestDB(t)
	repo := NewSharedFileRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	f := makeSharedFile(t, repo, clinicA, "result-a.pdf")

	t.Run("FindByID は自クリニックのレコードを返す", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, f.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "result-a.pdf", got.FileName)
	})

	t.Run("FindByID は別クリニックのレコードを返さない（NotFound）", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, f.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("FindByID は存在しないIDに対して NotFound を返す", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 9999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("FindByID は deleted_at 設定済みのレコードを除外する", func(t *testing.T) {
		softDeleted := makeSharedFile(t, repo, clinicA, "soft-deleted.pdf")
		now := time.Now()
		require.NoError(t, db.WithContext(ctx).Model(&model.SharedFile{}).Where("id = ?", softDeleted.ID).Update("deleted_at", &now).Error)

		_, err := repo.FindByID(ctx, clinicA, softDeleted.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.SharedFile
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", softDeleted.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "行自体は物理削除されず残る")
	})
}

// TestSharedFileRepository_FindAll は一覧取得（clinic_id 隔離・deleted_at 除外・降順ソート）を検証する。
func TestSharedFileRepository_FindAll(t *testing.T) {
	db := setupSharedFileTestDB(t)
	repo := NewSharedFileRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	older := makeSharedFile(t, repo, clinicA, "older.pdf")
	require.NoError(t, db.WithContext(ctx).Model(&model.SharedFile{}).Where("id = ?", older.ID).
		Update("created_at", time.Now().Add(-1*time.Hour)).Error)
	newer := makeSharedFile(t, repo, clinicA, "newer.pdf")

	// 別クリニックのファイルは混入しない
	makeSharedFile(t, repo, clinicB, "clinicB.pdf")

	// deleted_at 済みは除外される
	deleted := makeSharedFile(t, repo, clinicA, "deleted.pdf")
	now := time.Now()
	require.NoError(t, db.WithContext(ctx).Model(&model.SharedFile{}).Where("id = ?", deleted.ID).Update("deleted_at", &now).Error)

	t.Run("自クリニックのファイルのみ created_at 降順で返す", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 2, "clinicB・deleted 済みは含まれない")
		assert.Equal(t, newer.ID, got[0].ID, "新しい順（DESC）で先頭に来る")
		assert.Equal(t, older.ID, got[1].ID)
	})

	t.Run("別クリニックは自身のファイルのみ返す", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "clinicB.pdf", got[0].FileName)
	})
}

// TestSharedFileRepository_Delete は削除の正常系・clinic_id 隔離を検証する。
func TestSharedFileRepository_Delete(t *testing.T) {
	db := setupSharedFileTestDB(t)
	repo := NewSharedFileRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	f := makeSharedFile(t, repo, clinicA, "to-delete.pdf")

	t.Run("別クリニックからの削除は RowsAffected==0 を検知して NotFound を返す", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, f.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		// 実際にはまだ削除されていないことを確認する
		got, err := repo.FindByID(ctx, clinicA, f.ID)
		require.NoError(t, err)
		assert.Equal(t, f.ID, got.ID)
	})

	t.Run("正常系: 自クリニックからの削除でソフトデリートされる（行は残る）", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, f.ID))

		_, err := repo.FindByID(ctx, clinicA, f.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.SharedFile
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", f.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "Delete はソフトデリートであり行自体は残る")
	})
}

// TestSharedFileRepository_FindExpired は created_at ベースの期限切れ抽出（clinic 横断・deleted_at 除外）を検証する。
func TestSharedFileRepository_FindExpired(t *testing.T) {
	db := setupSharedFileTestDB(t)
	repo := NewSharedFileRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	old := makeSharedFile(t, repo, clinicA, "old.pdf")
	require.NoError(t, db.WithContext(ctx).Model(&model.SharedFile{}).Where("id = ?", old.ID).
		Update("created_at", time.Now().Add(-48*time.Hour)).Error)

	oldOtherClinic := makeSharedFile(t, repo, clinicB, "old-clinicb.pdf")
	require.NoError(t, db.WithContext(ctx).Model(&model.SharedFile{}).Where("id = ?", oldOtherClinic.ID).
		Update("created_at", time.Now().Add(-48*time.Hour)).Error)

	recent := makeSharedFile(t, repo, clinicA, "recent.pdf")
	_ = recent

	oldButDeleted := makeSharedFile(t, repo, clinicA, "old-deleted.pdf")
	now := time.Now()
	require.NoError(t, db.WithContext(ctx).Model(&model.SharedFile{}).Where("id = ?", oldButDeleted.ID).
		Updates(map[string]any{"created_at": time.Now().Add(-48 * time.Hour), "deleted_at": &now}).Error)

	threshold := time.Now().Add(-24 * time.Hour).Unix()

	t.Run("しきい値より古く未削除のファイルをクリニック横断で返す", func(t *testing.T) {
		got, err := repo.FindExpired(ctx, threshold)
		require.NoError(t, err)

		ids := make(map[uint64]bool, len(got))
		for _, f := range got {
			ids[f.ID] = true
		}
		assert.True(t, ids[old.ID], "しきい値より古いファイルが含まれる")
		assert.True(t, ids[oldOtherClinic.ID], "FindExpired はクリニックを横断して抽出する（バッチ用途）")
		assert.False(t, ids[recent.ID], "しきい値より新しいファイルは含まれない")
		assert.False(t, ids[oldButDeleted.ID], "deleted_at 設定済みは除外される")
	})
}
