package lstep

// lstep_csv_import_repository_test.go — LstepCsvImportRepository 統合テスト。
//
// 保護する不変条件:
//   - Create は新規インポート履歴を作成し UUID を採番する。
//   - Update は clinic_id が一致する場合のみ反映される（不一致は静かに no-op）。
//   - FindByID は clinic_id スコープで分離され、該当なしは NotFound を返す。
//   - FindAllByClinicID は imported_at DESC NULLS LAST 順・limit・clinic_id 分離を守る。

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLstepCsvImportTestDB は lstep_csv_imports テーブルを用意する。
func setupLstepCsvImportTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepCsvImport{}))
	db.Exec("TRUNCATE TABLE lstep_csv_imports CASCADE")
	return db
}

func TestLstepCsvImportRepository_Create(t *testing.T) {
	db := setupLstepCsvImportTestDB(t)
	repo := NewLstepCsvImportRepository(db)
	ctx := context.Background()

	imp := &model.LstepCsvImport{
		ClinicID: 1,
		CsvType:  "friend_attribute",
		FileName: "friends.csv",
		RowCount: 10,
	}
	require.NoError(t, repo.Create(ctx, imp))
	assert.NotEqual(t, uuid.Nil, imp.ID, "UUID が採番されること")

	var stored model.LstepCsvImport
	require.NoError(t, db.Where("id = ?", imp.ID).First(&stored).Error)
	assert.Equal(t, "friends.csv", stored.FileName)
	assert.Equal(t, "pending", stored.Status, "デフォルトステータスは pending")
}

func TestLstepCsvImportRepository_Update(t *testing.T) {
	db := setupLstepCsvImportTestDB(t)
	repo := NewLstepCsvImportRepository(db)
	ctx := context.Background()

	t.Run("clinic_id が一致すれば更新が反映される", func(t *testing.T) {
		imp := &model.LstepCsvImport{ClinicID: 1, CsvType: "friend_attribute", FileName: "a.csv"}
		require.NoError(t, repo.Create(ctx, imp))

		now := time.Now().Truncate(time.Second)
		imp.Status = "completed"
		imp.SuccessCount = 8
		imp.ErrorCount = 2
		imp.ImportedAt = &now
		require.NoError(t, repo.Update(ctx, imp))

		var stored model.LstepCsvImport
		require.NoError(t, db.Where("id = ?", imp.ID).First(&stored).Error)
		assert.Equal(t, "completed", stored.Status)
		assert.Equal(t, 8, stored.SuccessCount)
		assert.Equal(t, 2, stored.ErrorCount)
	})

	t.Run("clinic_id が不一致だと静かに no-op となり行は変化しない", func(t *testing.T) {
		imp := &model.LstepCsvImport{ClinicID: 5, CsvType: "friend_attribute", FileName: "b.csv"}
		require.NoError(t, repo.Create(ctx, imp))

		mismatched := *imp
		mismatched.ClinicID = 999 // 実際の行と異なる clinic_id
		mismatched.Status = "failed"
		require.NoError(t, repo.Update(ctx, &mismatched), "不一致でもエラーにはならない")

		var stored model.LstepCsvImport
		require.NoError(t, db.Where("id = ?", imp.ID).First(&stored).Error)
		assert.Equal(t, "pending", stored.Status, "clinic_id 不一致の更新は反映されない")
	})
}

func TestLstepCsvImportRepository_FindByID(t *testing.T) {
	db := setupLstepCsvImportTestDB(t)
	repo := NewLstepCsvImportRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	imp := &model.LstepCsvImport{ClinicID: clinicA, CsvType: "friend_attribute", FileName: "found.csv"}
	require.NoError(t, repo.Create(ctx, imp))

	t.Run("存在するIDかつ正しいclinic_idで見つかる", func(t *testing.T) {
		found, err := repo.FindByID(ctx, clinicA, imp.ID)
		require.NoError(t, err)
		assert.Equal(t, "found.csv", found.FileName)
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, uuid.New())
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックからは見えない（clinic_id分離）", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, imp.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestLstepCsvImportRepository_FindAllByClinicID(t *testing.T) {
	db := setupLstepCsvImportTestDB(t)
	repo := NewLstepCsvImportRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(10), uint64(20)
	old := time.Now().Add(-48 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)

	oldImp := &model.LstepCsvImport{ClinicID: clinicA, CsvType: "friend_attribute", FileName: "old.csv", ImportedAt: &old}
	newImp := &model.LstepCsvImport{
		ClinicID: clinicA, CsvType: "friend_attribute", FileName: "new.csv", ImportedAt: &newer,
		ErrorLog: datatypes.JSON(`[{"row":2,"reason":"unknown_line_user_id"}]`),
	}
	pendingImp := &model.LstepCsvImport{ClinicID: clinicA, CsvType: "friend_attribute", FileName: "pending.csv"} // imported_at NULL
	otherClinicImp := &model.LstepCsvImport{ClinicID: clinicB, CsvType: "friend_attribute", FileName: "other.csv", ImportedAt: &newer}

	require.NoError(t, repo.Create(ctx, oldImp))
	require.NoError(t, repo.Create(ctx, newImp))
	require.NoError(t, repo.Create(ctx, pendingImp))
	require.NoError(t, repo.Create(ctx, otherClinicImp))

	t.Run("imported_at DESC NULLS LAST の順で clinic_id 分離される", func(t *testing.T) {
		results, err := repo.FindAllByClinicID(ctx, clinicA, 10)
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "new.csv", results[0].FileName, "最新の imported_at が先頭")
		assert.Equal(t, "old.csv", results[1].FileName)
		assert.Equal(t, "pending.csv", results[2].FileName, "imported_at NULL は最後")
		for _, result := range results {
			assert.Empty(t, result.ErrorLog, "履歴一覧では詳細エラーログを取得しない")
		}

		var stored model.LstepCsvImport
		require.NoError(t, db.Where("id = ?", newImp.ID).First(&stored).Error)
		assert.NotEmpty(t, stored.ErrorLog, "永続化済みのエラーログ自体は保持する")
	})

	t.Run("limit で件数を制限する", func(t *testing.T) {
		results, err := repo.FindAllByClinicID(ctx, clinicA, 1)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "new.csv", results[0].FileName)
	})

	t.Run("別クリニックの履歴は含まれない", func(t *testing.T) {
		results, err := repo.FindAllByClinicID(ctx, clinicB, 10)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "other.csv", results[0].FileName)
	})
}
