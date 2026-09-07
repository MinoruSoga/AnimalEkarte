package medicalrecord

// repository_test.go — Repository の統合テスト（Phase 4 カバレッジ向上）。
//
// 対象: FindAll / FindByID / Create / Update / Delete / Reorder /
//       CountUsageByCheckupTypeID / CountChildrenByParentID
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ。
//
// makeSpeciesAndPet/makeHistoryMedicalRecord は medicalrecord テストパッケージ共有の
// ローカルコピー（diagnosis_name_repository_test.go 定義）をそのまま使う。makeCheckupRec は
// checkup_field 系テストも参照する共有ヘルパーとして本ファイルに 1 箇所だけ定義する。
// makeTestOwner はフラット側と同様 testdb.MakeTestOwner に委譲する。

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

// setupCheckupTypeTestDB は checkup_types / checkups 周りを整備する。
// CountUsageByCheckupTypeID は checkups を JOIN 相当で参照するため owner/pet/medical_record も揃える。
func setupCheckupTypeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.CheckupType{}, &model.Checkup{}))
	db.Exec("TRUNCATE TABLE checkups CASCADE")
	db.Exec("TRUNCATE TABLE checkup_types CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeCheckupRec(t *testing.T, db *gorm.DB, clinicID, mrID, petID, checkupTypeID uint64) uint64 {
	t.Helper()
	pid := petID
	c := &model.Checkup{
		ClinicID:        clinicID,
		MedicalRecordID: mrID,
		PetID:           &pid,
		CheckupTypeID:   checkupTypeID,
		Date:            time.Now(),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c.ID
}

func TestCheckupTypeRepository_FindAll_ClinicIsolationAndSortOrder(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ctB := &model.CheckupType{ClinicID: clinicB, Name: "医院Bの健診", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(ctB).Error)
	ctSecond := &model.CheckupType{ClinicID: clinicA, Name: "Bタイプ", SortOrder: 2}
	require.NoError(t, db.WithContext(ctx).Create(ctSecond).Error)
	ctFirst := &model.CheckupType{ClinicID: clinicA, Name: "Aタイプ", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(ctFirst).Error)

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "clinic A の健診種別のみ返る")
	assert.Equal(t, ctFirst.ID, got[0].ID, "sort_order 昇順で先頭に来る")
	assert.Equal(t, ctSecond.ID, got[1].ID)
	for _, ct := range got {
		assert.NotEqual(t, ctB.ID, ct.ID, "別クリニックの健診種別が混入してはならない")
	}
}

func TestCheckupTypeRepository_FindByID(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ct := &model.CheckupType{ClinicID: clinicA, Name: "血液検査パック"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, "血液検査パック", got.Name)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, ct.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCheckupTypeRepository_Create(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ct := &model.CheckupType{ClinicID: clinicA, Name: "新規健診種別"}
	require.NoError(t, repo.Create(ctx, ct))
	assert.NotZero(t, ct.ID)

	got, err := repo.FindByID(ctx, clinicA, ct.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規健診種別", got.Name)
}

func TestCheckupTypeRepository_Update(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ct := &model.CheckupType{ClinicID: clinicA, Name: "旧名称"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		name := "新名称"
		got, err := repo.Update(ctx, clinicA, ct.ID, UpdateCheckupTypeInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "新名称", got.Name)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		name := "乗っ取り"
		_, err := repo.Update(ctx, clinicB, ct.ID, UpdateCheckupTypeInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		name := "x"
		_, err := repo.Update(ctx, clinicA, 999999, UpdateCheckupTypeInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCheckupTypeRepository_Delete(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ct := &model.CheckupType{ClinicID: clinicA, Name: "削除対象"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, ct.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, ct.ID))

		// FindAll / FindByID からは除外される。
		_, err := repo.FindByID(ctx, clinicA, ct.ID)
		assert.True(t, apperrors.IsNotFound(err))

		all, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, x := range all {
			assert.NotEqual(t, ct.ID, x.ID)
		}

		// 物理行はまだ存在する（ソフトデリート）。
		var raw model.CheckupType
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", ct.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき")
	})
}

func TestCheckupTypeRepository_Reorder(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c1 := &model.CheckupType{ClinicID: clinicA, Name: "C1"}
	require.NoError(t, db.WithContext(ctx).Create(c1).Error)
	c2 := &model.CheckupType{ClinicID: clinicA, Name: "C2"}
	require.NoError(t, db.WithContext(ctx).Create(c2).Error)
	c3 := &model.CheckupType{ClinicID: clinicA, Name: "C3"}
	require.NoError(t, db.WithContext(ctx).Create(c3).Error)

	t.Run("指定順に sort_order が振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{c3.ID, c1.ID, c2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, c3.ID, got[0].ID)
		assert.Equal(t, c1.ID, got[1].ID)
		assert.Equal(t, c2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含むと失敗する", func(t *testing.T) {
		other := &model.CheckupType{ClinicID: clinicB, Name: "他院"}
		require.NoError(t, db.WithContext(ctx).Create(other).Error)
		err := repo.Reorder(ctx, clinicA, []uint64{c1.ID, other.ID})
		require.Error(t, err)
	})
}

func TestCheckupTypeRepository_CountUsageByCheckupTypeID(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := testdb.MakeTestOwner(t, db, clinicA, "健診種別使用飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "使用ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CTUSE", time.Now().UTC())
	ct := &model.CheckupType{ClinicID: clinicA, Name: "使用中の健診種別"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)

	t.Run("使用実績が無ければ 0", func(t *testing.T) {
		count, err := repo.CountUsageByCheckupTypeID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)

	t.Run("生存する checkup が使用していれば 1", func(t *testing.T) {
		count, err := repo.CountUsageByCheckupTypeID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは 0（クロステナント越境なし）", func(t *testing.T) {
		count, err := repo.CountUsageByCheckupTypeID(ctx, clinicB, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリートされた checkup は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Where("id = ?", checkupID).Delete(&model.Checkup{}).Error)
		count, err := repo.CountUsageByCheckupTypeID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestCheckupTypeRepository_CountChildrenByParentID(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	parent := &model.CheckupType{ClinicID: clinicA, Name: "親パッケージ"}
	require.NoError(t, db.WithContext(ctx).Create(parent).Error)

	t.Run("子が無ければ 0", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	child1 := &model.CheckupType{ClinicID: clinicA, Name: "子1", ParentID: &parent.ID}
	require.NoError(t, db.WithContext(ctx).Create(child1).Error)
	child2 := &model.CheckupType{ClinicID: clinicA, Name: "子2", ParentID: &parent.ID}
	require.NoError(t, db.WithContext(ctx).Create(child2).Error)

	t.Run("子が2件あれば 2", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("別クリニックIDでは 0", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicB, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリートされた子は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(child1).Error)
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}
