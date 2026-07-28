package auth

// permission_group_staff_clinic_isolation_test.go — クロステナント write 横断監査 回帰テスト
//
// テスト対象: PermissionGroupRepository.UpdateStaffGroups の clinic_id 境界
// 保護する不変条件: clinic A のスタッフに clinic B の権限グループを紐付けられない
//   （staff_permission_groups は自前 clinic_id を持たず、親 permission_groups 経由でのみ
//   テナント判定されるため、group_id の所有権を検証しなければクロステナント write になる）。
//
// group_id 所有権検証（clinic_id = ? AND id IN ?）を削除すると必ず失敗するよう設計されている。

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

// setupPermissionGroupStaffIsolationTestDB は権限グループ紐付け clinic_id 隔離テスト用の DB を整備する。
func setupPermissionGroupStaffIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.PermissionGroup{}, &model.StaffPermissionGroup{},
	))
	ensureStaffPermissionGroupsCreatedAt(t, db)
	db.Exec("TRUNCATE TABLE staff_permission_groups, staff_clinic_assignments, permission_groups, staffs CASCADE")
	return db
}

// ensureStaffPermissionGroupsCreatedAt self-heals a stale ekarte_db_test where
// staff_permission_groups was created as a bare many2many join (staff_id, group_id only).
// model.StaffPermissionGroup.CreatedAt and 001_init.sql both require created_at; GORM
// AutoMigrate does not always widen an existing join table that PermissionGroup.Staffs
// many2many already provisioned without the extra column.
func ensureStaffPermissionGroupsCreatedAt(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`
		ALTER TABLE staff_permission_groups
		ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now()
	`).Error)
}

func makePermissionGroup(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.PermissionGroup {
	t.Helper()
	g := &model.PermissionGroup{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(g).Error)
	return g
}

// TestPermissionGroupRepository_UpdateStaffGroups_ClinicIsolation は
// clinic A のスタッフに clinic B の権限グループを紐付けられないことを検証する。
// group_id の clinic_id 検証を削除すると「別クリニックのグループIDは拒否される」が失敗する。
func TestPermissionGroupRepository_UpdateStaffGroups_ClinicIsolation(t *testing.T) {
	db := setupPermissionGroupStaffIsolationTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()

	clinicA := makePermissionGroupTestClinic(t, db, "権限紐付け clinic A").ID
	clinicB := makePermissionGroupTestClinic(t, db, "権限紐付け clinic B").ID

	// clinic A のみに所属するスタッフ。UpdateStaffGroups 自身が所属を再検証するため、
	// staff_clinic_assignments に正規の所属行を作成する。
	staffA := makeDoctorAssignedToClinic(t, db, clinicA, "権限紐付けテストA用スタッフ")
	groupA := makePermissionGroup(t, db, clinicA, "clinic A グループ")
	groupB := makePermissionGroup(t, db, clinicB, "clinic B グループ")

	countAssignments := func(staffID uint64) int64 {
		var n int64
		require.NoError(t, db.Model(&model.StaffPermissionGroup{}).
			Where("staff_id = ?", staffID).Count(&n).Error)
		return n
	}

	t.Run("別クリニックのグループIDは拒否され、行が永続化されない（clinic_id 隔離）", func(t *testing.T) {
		err := repo.UpdateStaffGroups(ctx, clinicA, staffA.ID, []uint64{groupB.ID})
		require.Error(t, err, "clinic A のスタッフに clinic B の権限グループを紐付けできてはならない")
		assert.Zero(t, countAssignments(staffA.ID), "拒否時に staff_permission_groups 行を残してはならない")
	})

	t.Run("スタッフが対象クリニックに未所属なら同院グループでも拒否される", func(t *testing.T) {
		err := repo.UpdateStaffGroups(ctx, clinicB, staffA.ID, []uint64{groupB.ID})
		require.Error(t, err, "clinic B に未所属のスタッフへ clinic B 権限を保存できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "所属外と不存在は同じ NotFound 境界にする: %v", err)
		assert.Zero(t, countAssignments(staffA.ID), "拒否時に staff_permission_groups 行を残してはならない")
	})

	t.Run("同一クリニックのグループIDは許可され、行が永続化される", func(t *testing.T) {
		err := repo.UpdateStaffGroups(ctx, clinicA, staffA.ID, []uint64{groupA.ID})
		require.NoError(t, err)
		assert.Equal(t, int64(1), countAssignments(staffA.ID), "同一クリニックの紐付けは1件保存されるべき")
	})

	t.Run("一部が別クリニックのグループなら全体を拒否する（部分書き込み防止）", func(t *testing.T) {
		err := repo.UpdateStaffGroups(ctx, clinicA, staffA.ID, []uint64{groupA.ID, groupB.ID})
		require.Error(t, err, "clinic B のグループが混在する場合は全体を拒否すべき")
		assert.Equal(t, int64(1), countAssignments(staffA.ID), "拒否時に既存の紐付けを破壊してはならない")
	})
}

// TestPermissionGroupRepository_UpdateStaffGroups_DeleteScopedToClinic は
// BE-refactor.md H-1 の回帰テスト。多施設所属スタッフ（staff_clinic_assignments で
// clinic A/B 双方に所属）が clinic B で正当に紐付けた権限グループが、clinic A での
// 保存操作（DELETE + INSERT の全置換）によって無警告で削除されないことを検証する。
// staff_permission_groups は自前 clinic_id を持たないため、DELETE を
// staff_id のみでスコープすると（修正前の実装）、clinic を跨いで他クリニック分の
// 紐付けまで消えてしまう。DELETE を group 側の clinic_id サブクエリでスコープすると
// このテストは PASS する。
func TestPermissionGroupRepository_UpdateStaffGroups_DeleteScopedToClinic(t *testing.T) {
	db := setupPermissionGroupStaffIsolationTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()

	clinicA := makePermissionGroupTestClinic(t, db, "権限差し替え clinic A").ID
	clinicB := makePermissionGroupTestClinic(t, db, "権限差し替え clinic B").ID

	// 多施設所属スタッフ: 主所属は clinic A、加えて clinic B にも所属する。
	staff := makeDoctorAssignedToClinic(t, db, clinicA, "多施設所属スタッフ")
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinicB,
	}).Error)

	groupA := makePermissionGroup(t, db, clinicA, "clinic A グループ")
	groupB := makePermissionGroup(t, db, clinicB, "clinic B グループ")

	// clinic B での過去の正当な保存操作を模して、直接 clinic B のグループへ紐付けておく
	// （repo.UpdateStaffGroups(clinicB, ...) を経由しても同じ状態になる）。
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffPermissionGroup{
		StaffID: staff.ID, GroupID: groupB.ID,
	}).Error)

	linkExists := func(groupID uint64) bool {
		var n int64
		require.NoError(t, db.Model(&model.StaffPermissionGroup{}).
			Where("staff_id = ? AND group_id = ?", staff.ID, groupID).Count(&n).Error)
		return n == 1
	}

	// clinic A での保存操作（clinic A のグループへ差し替え）。
	err := repo.UpdateStaffGroups(ctx, clinicA, staff.ID, []uint64{groupA.ID})
	require.NoError(t, err)

	assert.True(t, linkExists(groupB.ID), "clinic B の紐付けは clinic A の保存操作で削除されてはならない")
	assert.True(t, linkExists(groupA.ID), "clinic A の新規紐付けは保存されるべき")
}

// TestPermissionGroupRepository_FindAllGroupIDsByStaffID_ClinicIsolation は、
// 多施設所属スタッフの権限グループID取得が、認証済み clinic_id の境界を越えないことを検証する。
// staff_permission_groups 自体は clinic_id を持たないため、permission_groups への JOIN と
// clinic_id/deleted_at predicate がなければ clinic A の応答へ clinic B のIDが漏洩する。
func TestPermissionGroupRepository_FindAllGroupIDsByStaffID_ClinicIsolation(t *testing.T) {
	db := setupPermissionGroupStaffIsolationTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()

	clinicA := makePermissionGroupTestClinic(t, db, "権限取得 clinic A").ID
	clinicB := makePermissionGroupTestClinic(t, db, "権限取得 clinic B").ID

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "多施設所属権限取得スタッフ")
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinicB,
	}).Error)

	groupA := makePermissionGroup(t, db, clinicA, "clinic A 取得対象")
	groupB := makePermissionGroup(t, db, clinicB, "clinic B 取得対象")
	deletedGroupA := makePermissionGroup(t, db, clinicA, "clinic A 削除済み")
	require.NoError(t, db.WithContext(ctx).Create([]model.StaffPermissionGroup{
		{StaffID: staff.ID, GroupID: groupA.ID},
		{StaffID: staff.ID, GroupID: groupB.ID},
		{StaffID: staff.ID, GroupID: deletedGroupA.ID},
	}).Error)
	require.NoError(t, db.WithContext(ctx).Delete(deletedGroupA).Error)

	clinicAGroupIDs, err := repo.FindAllGroupIDsByStaffID(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, []uint64{groupA.ID}, clinicAGroupIDs)

	clinicBGroupIDs, err := repo.FindAllGroupIDsByStaffID(ctx, clinicB, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, []uint64{groupB.ID}, clinicBGroupIDs)
}

// TestPermissionGroupRepository_UpdateStaffGroups_SerializesAssignmentRemoval は、
// 所属解除と権限全置換の TOCTOU を実 PostgreSQL の行ロックで検証する。
// 所属解除が先に staff_clinic_assignments 行を更新中なら、権限置換は同じ active 行の
// FOR UPDATE 取得で待機し、古い所属スナップショットを根拠に権限を書き込んではならない。
func TestPermissionGroupRepository_UpdateStaffGroups_SerializesAssignmentRemoval(t *testing.T) {
	db := setupPermissionGroupStaffIsolationTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()

	clinic := makePermissionGroupTestClinic(t, db, "権限所属解除競合 clinic")
	staff := makeDoctorAssignedToClinic(t, db, clinic.ID, "権限所属解除競合スタッフ")
	group := makePermissionGroup(t, db, clinic.ID, "権限所属解除競合グループ")

	assignmentUpdateStarted := make(chan struct{})
	releaseAssignmentUpdate := make(chan struct{})
	assignmentUpdateDone := make(chan error, 1)
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() {
			close(releaseAssignmentUpdate)
		})
	}
	t.Cleanup(release)

	go func() {
		assignmentUpdateDone <- db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			result := tx.Model(&model.StaffClinicAssignment{}).
				Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staff.ID, clinic.ID).
				Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP"))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return apperrors.WrapInternalServerError("assignment removal did not update one row")
			}
			close(assignmentUpdateStarted)
			<-releaseAssignmentUpdate
			return nil
		})
	}()
	select {
	case <-assignmentUpdateStarted:
	case holderErr := <-assignmentUpdateDone:
		require.NoError(t, holderErr)
		require.FailNow(t, "assignment removal transaction ended before holding its row lock")
	}

	// lock_timeout makes the blocked/not-blocked outcome deterministic. Without
	// the assignment FOR UPDATE query, the old implementation reaches DELETE +
	// INSERT and commits successfully while the assignment removal is uncommitted.
	updateErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL lock_timeout = '100ms'").Error; err != nil {
			return err
		}
		txCtx := persistence.WithTxValue(ctx, tx)
		return repo.UpdateStaffGroups(txCtx, clinic.ID, staff.ID, []uint64{group.ID})
	})
	release()
	require.NoError(t, <-assignmentUpdateDone)

	require.Error(t, updateErr, "権限置換は競合する所属解除の行ロックを待機すべき")
	assert.True(
		t,
		strings.Contains(updateErr.Error(), "55P03") ||
			strings.Contains(updateErr.Error(), "lock timeout"),
		"expected PostgreSQL lock timeout, got: %v",
		updateErr,
	)

	retryErr := repo.UpdateStaffGroups(ctx, clinic.ID, staff.ID, []uint64{group.ID})
	require.Error(t, retryErr, "所属解除確定後の再試行は拒否すべき")
	assert.True(t, apperrors.IsNotFound(retryErr), "unexpected retry error: %v", retryErr)

	var links int64
	require.NoError(t, db.Model(&model.StaffPermissionGroup{}).
		Where("staff_id = ? AND group_id = ?", staff.ID, group.ID).
		Count(&links).Error)
	assert.Zero(t, links, "所属解除と競合した権限リンクを残してはならない")
}
