package repository

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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupPermissionGroupStaffIsolationTestDB は権限グループ紐付け clinic_id 隔離テスト用の DB を整備する。
func setupPermissionGroupStaffIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.Staff{}, &model.PermissionGroup{}, &model.StaffPermissionGroup{},
	))
	db.Exec("TRUNCATE TABLE staff_permission_groups, permission_groups, staffs CASCADE")
	return db
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

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	// clinic A のスタッフ。スタッフ所有権は呼び出し側（handler verifyStaffClinicMembership）で検証済みのため、
	// ここでは group_id（permission_group）のテナント越境のみを対象に検証する。
	staffA := makeDoctor(t, db, clinicA, "権限紐付けテストA用スタッフ")
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
