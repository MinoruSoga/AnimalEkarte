package auth

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
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupPermissionGroupRepositoryTestDB は本ファイルの全テストで共有する DB を整備する。
func setupPermissionGroupRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Staff{}, &model.StaffClinicAssignment{}, &model.PermissionGroup{},
		&model.PermissionGroupRule{}, &model.StaffPermissionGroup{},
	))
	ensureStaffPermissionGroupsCreatedAt(t, db)
	testdb.Truncate(t, db, "staff_permission_groups", "staff_clinic_assignments", "permission_group_rules", "permission_groups", "staffs")
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

func TestPermissionGroupRepository_DeleteSoftDeletedByClinicID_ClinicIsolation(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	activeA := makePermissionGroup(t, db, clinicA, "cleanup active clinic A")
	deletedA := makePermissionGroup(t, db, clinicA, "cleanup deleted clinic A")
	deletedB := makePermissionGroup(t, db, clinicB, "cleanup deleted clinic B")
	require.NoError(t, db.WithContext(ctx).Delete(deletedA).Error)
	require.NoError(t, db.WithContext(ctx).Delete(deletedB).Error)

	require.NoError(t, repo.DeleteSoftDeletedByClinicID(ctx, clinicA))

	countPhysicalRows := func(id uint64) int64 {
		var count int64
		require.NoError(t, db.Unscoped().
			Model(&model.PermissionGroup{}).
			Where("id = ?", id).
			Count(&count).Error)
		return count
	}
	assert.Equal(t, int64(1), countPhysicalRows(activeA.ID), "active group in target clinic must remain")
	assert.Zero(t, countPhysicalRows(deletedA.ID), "only soft-deleted groups in target clinic are hard-deleted")
	assert.Equal(t, int64(1), countPhysicalRows(deletedB.ID), "soft-deleted group in another clinic must remain")

	t.Run("zero matching rows is successful", func(t *testing.T) {
		require.NoError(t, repo.DeleteSoftDeletedByClinicID(ctx, clinicA))
	})

	t.Run("database error preserves its cause", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		err := repo.DeleteSoftDeletedByClinicID(cancelledCtx, clinicA)

		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestPermissionGroupRepository_UpdateRules(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("全削除→再挿入で既存ルールが新ルールに完全置換される", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "UpdateRules対象グループ")
		makeEffPermRule(t, db, g.ID, "medical_record", true, true, true, true)
		makeEffPermRule(t, db, g.ID, "billing", true, false, false, false)

		newRules := []model.PermissionGroupRule{
			{Resource: "owner", CanView: true, CanCreate: false, CanEdit: false, CanDelete: false},
		}
		err := repo.UpdateRules(ctx, clinicA, g.ID, newRules)
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, err)
		require.Len(t, got.Rules, 1, "旧ルールは全て削除され新ルールのみ残るべき")
		assert.Equal(t, "owner", got.Rules[0].Resource)
		assert.True(t, got.Rules[0].CanView)
	})

	// BUG-024: uncheck (true→false) must survive replace-all insert, including when
	// GORM default tags would otherwise omit zero-value bool columns.
	t.Run("trueからfalseへのフラグ反転が永続化される", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "UpdateRules false persist")
		makeEffPermRule(t, db, g.ID, "reception", true, true, true, true)

		atomicRepo := repo.(PermissionGroupRulesAtomicWriter)
		updated, err := atomicRepo.UpdateWithRules(
			ctx,
			clinicA,
			g.ID,
			map[string]any{"name": "UpdateRules false persist"},
			[]model.PermissionGroupRule{{
				Resource:  "reception",
				CanView:   false,
				CanCreate: true,
				CanEdit:   false,
				CanDelete: false,
			}},
		)
		require.NoError(t, err)
		require.Len(t, updated.Rules, 1)
		assert.False(t, updated.Rules[0].CanView)
		assert.True(t, updated.Rules[0].CanCreate)
		assert.False(t, updated.Rules[0].CanEdit)
		assert.False(t, updated.Rules[0].CanDelete)

		got, findErr := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, findErr)
		require.Len(t, got.Rules, 1)
		assert.Equal(t, "reception", got.Rules[0].Resource)
		assert.False(t, got.Rules[0].CanView, "can_view false must round-trip after UpdateWithRules")
		assert.True(t, got.Rules[0].CanCreate)
		assert.False(t, got.Rules[0].CanEdit)
		assert.False(t, got.Rules[0].CanDelete)
	})

	t.Run("空スライスを渡すと全ルールが削除される", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "UpdateRules空スライス対象")
		makeEffPermRule(t, db, g.ID, "medical_record", true, true, true, true)

		err := repo.UpdateRules(ctx, clinicA, g.ID, []model.PermissionGroupRule{})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, err)
		assert.Empty(t, got.Rules)
	})

	t.Run("別クリニックの対象グループはNotFoundで既存ルールを変更しない", func(t *testing.T) {
		g := makePermissionGroup(t, db, clinicA, "UpdateRules越境拒否対象")
		existing := makeEffPermRule(t, db, g.ID, "medical_record", true, false, false, false)

		err := repo.UpdateRules(ctx, clinicB, g.ID, []model.PermissionGroupRule{
			{Resource: "billing", CanView: true},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別院対象と不存在は同じ NotFound 境界にする: %v", err)

		got, findErr := repo.FindByID(ctx, clinicA, g.ID)
		require.NoError(t, findErr)
		require.Len(t, got.Rules, 1)
		assert.Equal(t, existing.ID, got.Rules[0].ID)
		assert.Equal(t, "medical_record", got.Rules[0].Resource)
	})
}

func TestPermissionGroupRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	group := &model.PermissionGroup{
		ClinicID: clinicID,
		Name:     "create inactive",
		Color:    "#112233",
		IsActive: false,
	}
	require.NoError(t, repo.Create(ctx, group))
	require.NotZero(t, group.ID)
	assert.False(t, group.IsActive, "in-memory struct must keep false after Create")

	got, err := repo.FindByID(ctx, clinicID, group.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive, "DB read-back must keep explicit false")

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Select("is_active").
		Where("id = ?", group.ID).
		Scan(&rawActive).Error)
	assert.False(t, rawActive, "raw is_active column must be false")
}

func TestPermissionGroupRepository_Create_IsActiveTruePersists(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := &model.PermissionGroup{
		ClinicID: clinicID,
		Name:     "create active true",
		Color:    "#112233",
		IsActive: true,
	}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)

	gotActive, err := repo.FindByID(ctx, clinicID, active.ID)
	require.NoError(t, err)
	assert.True(t, gotActive.IsActive)

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Select("is_active").
		Where("id = ?", active.ID).
		Scan(&rawActive).Error)
	assert.True(t, rawActive)
}

func TestPermissionGroupRepository_CreateWithRules_RollsBackParentWhenRulesInsertFails(
	t *testing.T,
) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	atomicRepo := repo.(PermissionGroupRulesAtomicWriter)
	ctx := context.Background()
	const clinicID = uint64(1)

	group := &model.PermissionGroup{
		ClinicID: clinicID,
		Name:     "atomic create rollback",
		Color:    "#112233",
		IsActive: true,
	}
	invalidRules := []model.PermissionGroupRule{
		{Resource: strings.Repeat("x", 51), CanView: true},
	}

	_, err := atomicRepo.CreateWithRules(ctx, group, invalidRules)

	require.Error(t, err)
	var groupCount int64
	require.NoError(t, db.Model(&model.PermissionGroup{}).
		Where("clinic_id = ? AND name = ?", clinicID, group.Name).
		Count(&groupCount).Error)
	assert.Zero(t, groupCount, "failed child insert must roll back parent creation")
}

func TestPermissionGroupRepository_UpdateWithRules_RollsBackParentAndPriorRules(
	t *testing.T,
) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	atomicRepo := repo.(PermissionGroupRulesAtomicWriter)
	ctx := context.Background()
	const clinicID = uint64(1)

	group := makePermissionGroup(t, db, clinicID, "atomic update original")
	originalRule := makeEffPermRule(
		t,
		db,
		group.ID,
		"medical_record",
		true,
		false,
		false,
		false,
	)
	invalidRules := []model.PermissionGroupRule{
		{Resource: strings.Repeat("x", 51), CanView: true},
	}

	_, err := atomicRepo.UpdateWithRules(
		ctx,
		clinicID,
		group.ID,
		map[string]any{"name": "atomic update changed"},
		invalidRules,
	)

	require.Error(t, err)
	got, findErr := repo.FindByID(ctx, clinicID, group.ID)
	require.NoError(t, findErr)
	assert.Equal(t, "atomic update original", got.Name)
	require.Len(t, got.Rules, 1)
	assert.Equal(t, originalRule.ID, got.Rules[0].ID)
	assert.Equal(t, "medical_record", got.Rules[0].Resource)
}

func TestPermissionGroupRepository_UpdateWithRules_RejectsOtherClinicWithoutChildChanges(
	t *testing.T,
) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	atomicRepo := repo.(PermissionGroupRulesAtomicWriter)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	group := makePermissionGroup(t, db, clinicA, "atomic cross-clinic original")
	originalRule := makeEffPermRule(
		t,
		db,
		group.ID,
		"medical_record",
		true,
		false,
		false,
		false,
	)

	_, err := atomicRepo.UpdateWithRules(
		ctx,
		clinicB,
		group.ID,
		map[string]any{"name": "atomic cross-clinic changed"},
		[]model.PermissionGroupRule{{Resource: "owners", CanView: true}},
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	got, findErr := repo.FindByID(ctx, clinicA, group.ID)
	require.NoError(t, findErr)
	assert.Equal(t, "atomic cross-clinic original", got.Name)
	require.Len(t, got.Rules, 1)
	assert.Equal(t, originalRule.ID, got.Rules[0].ID)
	assert.Equal(t, "medical_record", got.Rules[0].Resource)
}

// TestPermissionGroupRepository_UpdateRules_SerializesParentDeletion は、
// 親グループの削除とルール全置換を同じ親行ロックで直列化することを検証する。
func TestPermissionGroupRepository_UpdateRules_SerializesParentDeletion(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	group := makePermissionGroup(t, db, clinicID, "UpdateRules親削除競合")
	makeEffPermRule(t, db, group.ID, "medical_record", true, false, false, false)

	parentDeleteStarted := make(chan struct{})
	releaseParentDelete := make(chan struct{})
	parentDeleteDone := make(chan error, 1)
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() {
			close(releaseParentDelete)
		})
	}
	t.Cleanup(release)

	go func() {
		parentDeleteDone <- db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			result := tx.Model(&model.PermissionGroup{}).
				Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", group.ID, clinicID).
				Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP"))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return apperrors.WrapInternalServerError("permission group deletion did not update one row")
			}
			close(parentDeleteStarted)
			<-releaseParentDelete
			return nil
		})
	}()
	select {
	case <-parentDeleteStarted:
	case holderErr := <-parentDeleteDone:
		require.NoError(t, holderErr)
		require.FailNow(t, "parent deletion transaction ended before holding its row lock")
	}

	updateErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL lock_timeout = '100ms'").Error; err != nil {
			return err
		}
		txCtx := persistence.WithTxValue(ctx, tx)
		return repo.UpdateRules(txCtx, clinicID, group.ID, []model.PermissionGroupRule{
			{Resource: "billing", CanView: true},
		})
	})
	release()
	require.NoError(t, <-parentDeleteDone)

	require.Error(t, updateErr, "ルール置換は競合する親削除の行ロックを待機すべき")
	assert.True(
		t,
		strings.Contains(updateErr.Error(), "55P03") ||
			strings.Contains(updateErr.Error(), "lock timeout"),
		"expected PostgreSQL lock timeout, got: %v",
		updateErr,
	)

	retryErr := repo.UpdateRules(ctx, clinicID, group.ID, []model.PermissionGroupRule{
		{Resource: "billing", CanView: true},
	})
	require.Error(t, retryErr, "親削除確定後の再試行は拒否すべき")
	assert.True(t, apperrors.IsNotFound(retryErr), "unexpected retry error: %v", retryErr)
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

		ids, err := repo.FindAllGroupIDsByStaffID(ctx, clinicA, staff.ID)
		require.NoError(t, err)
		assert.ElementsMatch(t, []uint64{groupOne.ID, groupTwo.ID}, ids)
	})

	t.Run("所属グループがなければ空スライスを返す", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "未所属グループID一覧テスト用スタッフ")
		ids, err := repo.FindAllGroupIDsByStaffID(ctx, clinicA, staff.ID)
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

	t.Run("別クリニックのgroup_idを混ぜるとWrapInvalidInputになる", func(t *testing.T) {
		clinicA := makePermissionGroupTestClinic(t, db, "InvalidGroupID clinic A").ID
		clinicB := makePermissionGroupTestClinic(t, db, "InvalidGroupID clinic B").ID
		staff := makeDoctorAssignedToClinic(t, db, clinicA, "InvalidGroupID検証用スタッフ")
		groupB := makePermissionGroup(t, db, clinicB, "InvalidGroupID検証用clinicBグループ")

		err := repo.UpdateStaffGroups(ctx, clinicA, staff.ID, []uint64{groupB.ID})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "別クリニックのgroup_idはWrapInvalidInputで拒否されるべき")
	})
}
