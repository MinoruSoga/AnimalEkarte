package repository

// permission_group_repository_test.go
// permission_group_repository.go の実DB結合テスト（#212: internal/repository カバレッジ向上）。
// FindAll / FindByID / Update / Delete / UpdateRules / CountUsageByGroupID /
// FindAllGroupIDsByStaffID / Reorder / UpdateStaffGroups(WrapInvalidInput 追加検証) を対象とする。
//
// makePermissionGroup は permission_group_staff_clinic_isolation_test.go、
// makeEffPermGroup / makeEffPermRule は permission_group_effective_permissions_test.go、
// makeDoctor は isolation_test_helpers_test.go を再利用する（重複定義禁止のため新規追加しない）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupPermissionGroupRepositoryTestDB は本ファイルの全テストで共有する DB を整備する。
func setupPermissionGroupRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.Staff{}, &model.PermissionGroup{}, &model.PermissionGroupRule{}, &model.StaffPermissionGroup{},
	))
	ensureStaffPermissionGroupsCreatedAt(t, db)
	db.Exec("TRUNCATE TABLE staff_permission_groups, permission_group_rules, permission_groups, staffs CASCADE")
	return db
}

func TestPermissionGroupRepository_FindByID(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("正常系(Rules Preload込み)", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "FindByID正常系グループ")
		makeEffPermRule(t, db, g.ID, "medical_record", true, false, false, false)

		got, err := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, err)
		assert.Equal(t, g.ID, got.ID)
		require.Len(t, got.Rules, 1, "Rules が Preload されるべき")
		assert.Equal(t, "medical_record", got.Rules[0].Resource)
	})

	t.Run("別クリニックはNotFound", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "別クリニック検証用グループ")
		_, err := repo.FindByID(ctx, clinicB, g.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPermissionGroupRepository_FindAll(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("clinic_idでスコープされ他院データが混入しない", func(t *testing.T) {
		makePermissionGroup(t, db, clinicA, "clinic Aグループ1")
		makePermissionGroup(t, db, clinicA, "clinic Aグループ2")
		makePermissionGroup(t, db, clinicB, "clinic Bグループ")

		groups, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, groups, 2, "clinic A のグループのみ返るべき")
		for _, g := range groups {
			assert.Equal(t, clinicA, g.ClinicID, "他院データが混入してはならない")
		}
	})

	t.Run("該当0件は空スライスを返す", func(t *testing.T) {
		groups, err := repo.FindAll(ctx, 999999)
		require.NoError(t, err)
		assert.Empty(t, groups)
	})
}

func TestPermissionGroupRepository_Update(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックからの更新は成功する", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "Update正常系グループ")
		got, err := repo.Update(ctx, clinicA, g.ID, map[string]any{"name": "更新後の名前"})
		require.NoError(t, err)
		assert.Equal(t, "更新後の名前", got.Name)
	})

	t.Run("別クリニックからの更新はNotFound", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "別クリニック更新拒否対象")
		_, err := repo.Update(ctx, clinicB, g.ID, map[string]any{"name": "改ざん後の名前"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		// 実際に更新されていないことを確認する
		got, err := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, err)
		assert.Equal(t, "別クリニック更新拒否対象", got.Name, "他クリニックからの更新でデータが書き換わってはならない")
	})
}

func TestPermissionGroupRepository_Delete(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックからの削除は成功する", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "Delete正常系グループ")
		err := repo.Delete(ctx, clinicA, g.ID)
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, clinicA, g.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックからの削除はNotFoundで対象は削除されない", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "別クリニック削除拒否対象")
		err := repo.Delete(ctx, clinicB, g.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, err, "他クリニックからの削除要求でレコードが消えてはならない")
		assert.Equal(t, g.ID, got.ID)
	})
}

func TestPermissionGroupRepository_UpdateRules(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t.Run("全削除→再挿入で既存ルールが新ルールに完全置換される", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "UpdateRules対象グループ")
		makeEffPermRule(t, db, g.ID, "medical_record", true, true, true, true)
		makeEffPermRule(t, db, g.ID, "billing", true, false, false, false)

		newRules := []model.PermissionGroupRule{
			{Resource: "owner", CanView: true, CanCreate: false, CanEdit: false, CanDelete: false},
		}
		err := repo.UpdateRules(ctx, g.ID, newRules)
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, err)
		require.Len(t, got.Rules, 1, "旧ルールは全て削除され新ルールのみ残るべき")
		assert.Equal(t, "owner", got.Rules[0].Resource)
		assert.True(t, got.Rules[0].CanView)
	})

	t.Run("空スライスを渡すと全ルールが削除される", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "UpdateRules空スライス対象")
		makeEffPermRule(t, db, g.ID, "medical_record", true, true, true, true)

		err := repo.UpdateRules(ctx, g.ID, []model.PermissionGroupRule{})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, err)
		assert.Empty(t, got.Rules)
	})
}

func TestPermissionGroupRepository_CountUsageByGroupID(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t.Run("staff_permission_groups経由の使用数が正しくカウントされる", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "使用数カウント対象グループ")
		staff1 := makeDoctor(t, db, clinicA, "使用数カウントスタッフ1")
		staff2 := makeDoctor(t, db, clinicA, "使用数カウントスタッフ2")
		require.NoError(t, db.WithContext(ctx).Create(&model.StaffPermissionGroup{StaffID: staff1.ID, GroupID: g.ID}).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.StaffPermissionGroup{StaffID: staff2.ID, GroupID: g.ID}).Error)

		count, err := repo.CountUsageByGroupID(ctx, clinicA, g.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("使用しているスタッフがいない場合は0", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "未使用グループ")
		count, err := repo.CountUsageByGroupID(ctx, clinicA, g.ID)
		require.NoError(t, err)
		assert.Zero(t, count)
	})

	t.Run("別クリニックのgroup_idを指定すると0(clinic_id隔離)", func(t *testing.T) {
		const clinicB = uint64(2)
		g := makePermissionGroup(t, db, clinicA, "隔離検証用グループ")
		staff := makeDoctor(t, db, clinicA, "隔離検証用スタッフ")
		require.NoError(t, db.WithContext(ctx).Create(&model.StaffPermissionGroup{StaffID: staff.ID, GroupID: g.ID}).Error)

		count, err := repo.CountUsageByGroupID(ctx, clinicB, g.ID)
		require.NoError(t, err)
		assert.Zero(t, count, "別クリニックからの参照は使用数に混入してはならない")
	})
}

func TestPermissionGroupRepository_FindAllGroupIDsByStaffID(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t.Run("スタッフの所属グループID一覧が正しく返る", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "所属グループID一覧テスト用スタッフ")
		groupOne := makePermissionGroup(t, db, clinicA, "所属グループ1")
		groupTwo := makePermissionGroup(t, db, clinicA, "所属グループ2")
		require.NoError(t, db.WithContext(ctx).Create(&model.StaffPermissionGroup{StaffID: staff.ID, GroupID: groupOne.ID}).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.StaffPermissionGroup{StaffID: staff.ID, GroupID: groupTwo.ID}).Error)

		ids, err := repo.FindAllGroupIDsByStaffID(ctx, staff.ID)
		require.NoError(t, err)
		assert.ElementsMatch(t, []uint64{groupOne.ID, groupTwo.ID}, ids)
	})

	t.Run("所属グループがなければ空スライスを返す", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "未所属グループID一覧テスト用スタッフ")
		ids, err := repo.FindAllGroupIDsByStaffID(ctx, staff.ID)
		require.NoError(t, err)
		assert.Empty(t, ids)
	})
}

func TestPermissionGroupRepository_Reorder(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("sort_orderが指定順に更新される", func(t *testing.T) {
		g1 := makePermissionGroup(t, db, clinicA, "Reorder対象1")
		g2 := makePermissionGroup(t, db, clinicA, "Reorder対象2")
		g3 := makePermissionGroup(t, db, clinicA, "Reorder対象3")

		err := repo.Reorder(ctx, clinicA, []uint64{g3.ID, g1.ID, g2.ID})
		require.NoError(t, err)

		got1, err := repo.FindByID(ctx, clinicA, g1.ID)
		require.NoError(t, err)
		got2, err := repo.FindByID(ctx, clinicA, g2.ID)
		require.NoError(t, err)
		got3, err := repo.FindByID(ctx, clinicA, g3.ID)
		require.NoError(t, err)

		assert.Equal(t, 1, got3.SortOrder, "先頭に指定したg3は1になるべき")
		assert.Equal(t, 2, got1.SortOrder, "2番目に指定したg1は2になるべき")
		assert.Equal(t, 3, got2.SortOrder, "3番目に指定したg2は3になるべき")
	})

	t.Run("別クリニックのIDを混ぜるとエラーになる", func(t *testing.T) {
		gA := makePermissionGroup(t, db, clinicA, "Reorderエラー検証clinicA")
		gB := makePermissionGroup(t, db, clinicB, "Reorderエラー検証clinicB")

		err := repo.Reorder(ctx, clinicA, []uint64{gA.ID, gB.ID})
		require.Error(t, err, "対象クリニックに属さないIDが含まれる場合はエラーになるべき")
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

// TestPermissionGroupRepository_UpdateStaffGroups_InvalidGroupID は
// permission_group_staff_clinic_isolation_test.go のクロステナント検証を補完し、
// 別クリニックのgroup_idを混ぜた場合に返るエラーが apperrors.WrapInvalidInput 由来
// (IsInvalidInput=true) であることを明示的に検証する。
func TestPermissionGroupRepository_UpdateStaffGroups_InvalidGroupID(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("別クリニックのgroup_idを混ぜるとWrapInvalidInputになる", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "InvalidGroupID検証用スタッフ")
		groupB := makePermissionGroup(t, db, clinicB, "InvalidGroupID検証用clinicBグループ")

		err := repo.UpdateStaffGroups(ctx, clinicA, staff.ID, []uint64{groupB.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "別クリニックのgroup_idはWrapInvalidInputで拒否されるべき")
	})
}
