package reservation

// repository_test.go — ReservationTypeGroupRepository の CRUD・使用数カウント・並び替えを実 DB で検証する。
// Group 名の Preload clinic_id 隔離はフラット repository package の
// reservation_type_preload_clinic_isolation_test.go が別途カバーしているため、本ファイルは
// CRUD/Count/Reorder/NotFound/clinic_id 分離に焦点を当てる。
// makeReservationTypeGroup/makeReservationTypeLinked はフラット package の同名ヘルパーの複製
// （BE8-4 batch1: import cycle を避けるための最小限の重複、移動時の型リネームはしない方針の対象外）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupTestDB は ReservationTypeGroupRepository の CRUD テスト用に DB を整備する。setupTestDB が reservation_type_category
// ENUM を DROP CASCADE するため reservation_types.category 列が消える。AutoMigrate で列・テーブルを
// 再整備してから TRUNCATE でクリーンな状態にする。
func setupGroupRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationTypeGroup{}, &model.ReservationType{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	db.Exec("TRUNCATE TABLE reservation_type_groups CASCADE")
	return db
}

func makeReservationTypeGroup(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.ReservationTypeGroup {
	t.Helper()
	g := &model.ReservationTypeGroup{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(g).Error)
	return g
}

func makeReservationTypeLinked(t *testing.T, db *gorm.DB, clinicID uint64, name string, groupID, parentID *uint64) *model.ReservationType {
	t.Helper()
	rt := &model.ReservationType{
		ClinicID: clinicID,
		Name:     name,
		Category: model.ReservationTypeCategoryGeneral,
		GroupID:  groupID,
		ParentID: parentID,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rt).Error)
	return rt
}

func TestReservationTypeGroupRepository_FindAll(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	repo := NewReservationTypeGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	groupA := makeReservationTypeGroup(t, db, clinicA, "医院Aグループ")
	_ = makeReservationTypeGroup(t, db, clinicB, "医院Bグループ")

	t.Run("同一クリニックのみ取得できる", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, groupA.ID, got[0].ID)
	})

	t.Run("別クリニックIDでは見えない", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.NotEqual(t, groupA.ID, got[0].ID)
	})

	t.Run("ソフトデリート済みは除外される", func(t *testing.T) {
		deleted := makeReservationTypeGroup(t, db, clinicA, "削除予定グループ")
		require.NoError(t, repo.Delete(ctx, clinicA, deleted.ID))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, g := range got {
			assert.NotEqual(t, deleted.ID, g.ID, "ソフトデリート済みのグループが含まれてはならない")
		}
	})
}

func TestReservationTypeGroupRepository_FindByID(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	repo := NewReservationTypeGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	groupA := makeReservationTypeGroup(t, db, clinicA, "単体取得グループ")

	t.Run("同一クリニックIDで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, groupA.ID)
		require.NoError(t, err)
		assert.Equal(t, groupA.ID, got.ID)
	})

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, groupA.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDは NotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestReservationTypeGroupRepository_CountUsageByReservationTypeGroupID(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	groupRepo := NewReservationTypeGroupRepository(db)
	typeRepo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	groupA := makeReservationTypeGroup(t, db, clinicA, "使用状況確認用グループ")

	t.Run("未使用は0件", func(t *testing.T) {
		count, err := groupRepo.CountUsageByReservationTypeGroupID(ctx, clinicA, groupA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	_ = makeReservationTypeLinked(t, db, clinicA, "紐付け区分", &groupA.ID, nil)
	deletedLinked := makeReservationTypeLinked(t, db, clinicA, "削除済み紐付け区分", &groupA.ID, nil)
	require.NoError(t, typeRepo.Delete(ctx, clinicA, deletedLinked.ID))

	t.Run("有効な区分のみカウントされる（ソフトデリート除外）", func(t *testing.T) {
		count, err := groupRepo.CountUsageByReservationTypeGroupID(ctx, clinicA, groupA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは0件（clinic_id 隔離）", func(t *testing.T) {
		count, err := groupRepo.CountUsageByReservationTypeGroupID(ctx, clinicB, groupA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestReservationTypeGroupRepository_Create(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	repo := NewReservationTypeGroupRepository(db)
	ctx := context.Background()

	g := &model.ReservationTypeGroup{ClinicID: 1, Name: "新規作成グループ"}
	require.NoError(t, repo.Create(ctx, g))
	assert.NotZero(t, g.ID)

	got, err := repo.FindByID(ctx, 1, g.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規作成グループ", got.Name)
}

// BUG-455-S6: gorm default:true omits zero bools from INSERT.
func TestReservationTypeGroupRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	repo := NewReservationTypeGroupRepository(db)
	ctx := context.Background()

	g := &model.ReservationTypeGroup{ClinicID: 1, Name: "inactive group", IsActive: false}
	require.NoError(t, repo.Create(ctx, g))
	assert.False(t, g.IsActive)

	got, err := repo.FindByID(ctx, 1, g.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ReservationTypeGroup{}).
		Select("is_active").
		Where("id = ?", g.ID).
		Scan(&rawActive).Error)
	assert.False(t, rawActive, "raw is_active must be false")
}

func TestReservationTypeGroupRepository_Update(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	repo := NewReservationTypeGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	g := makeReservationTypeGroup(t, db, clinicA, "更新前グループ")

	t.Run("正しいクリニックIDで更新できる", func(t *testing.T) {
		updated, err := repo.Update(ctx, clinicA, g.ID, map[string]any{"name": "更新後グループ"})
		require.NoError(t, err)
		assert.Equal(t, "更新後グループ", updated.Name)
	})

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		updated, err := repo.Update(ctx, clinicB, g.ID, map[string]any{"name": "不正更新"})
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestReservationTypeGroupRepository_Delete(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	repo := NewReservationTypeGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	g := makeReservationTypeGroup(t, db, clinicA, "削除対象グループ")

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, g.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("正しいクリニックIDで削除できる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, g.ID))
		_, err := repo.FindByID(ctx, clinicA, g.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("既に削除済みの再削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, g.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestReservationTypeGroupRepository_Reorder(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	repo := NewReservationTypeGroupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	g1 := makeReservationTypeGroup(t, db, clinicA, "並び1")
	g2 := makeReservationTypeGroup(t, db, clinicA, "並び2")
	g3 := makeReservationTypeGroup(t, db, clinicA, "並び3")

	t.Run("指定順に sort_order が振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{g3.ID, g1.ID, g2.ID}))

		got3, err := repo.FindByID(ctx, clinicA, g3.ID)
		require.NoError(t, err)
		got1, err := repo.FindByID(ctx, clinicA, g1.ID)
		require.NoError(t, err)
		got2, err := repo.FindByID(ctx, clinicA, g2.ID)
		require.NoError(t, err)

		assert.Equal(t, 1, got3.SortOrder)
		assert.Equal(t, 2, got1.SortOrder)
		assert.Equal(t, 3, got2.SortOrder)
	})

	t.Run("他クリニックの id を含むと失敗する", func(t *testing.T) {
		otherClinicGroup := makeReservationTypeGroup(t, db, uint64(2), "別医院グループ")
		err := repo.Reorder(ctx, clinicA, []uint64{g1.ID, otherClinicGroup.ID})
		assert.Error(t, err, "clinic A のスコープに存在しない id を含む Reorder は失敗するべき")
	})
}
