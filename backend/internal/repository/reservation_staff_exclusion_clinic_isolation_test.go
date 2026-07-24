package repository

// reservation_staff_exclusion_clinic_isolation_test.go — クロステナント write 横断監査 回帰テスト
//
// テスト対象: ReservationStaffRepository.UpdateExcludedReservationTypes の clinic_id 境界
// 保護する不変条件: clinic A のスタッフの除外コースに clinic B の予約区分IDを
//   書き込めない（staff_reservation_exclusions は自前 clinic_id を持たず、
//   親 reservation_types 経由でのみテナント判定されるため、service/repo 層で
//   型IDの所有権を検証しなければクロステナント write になる）。
//
// 型ID所有権検証（clinic_id = ? AND id IN ?）を削除すると必ず失敗するよう設計されている。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupExclusionIsolationTestDB は除外コース clinic_id 隔離テスト用の DB を整備する。
func setupExclusionIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.Staff{}, &model.StaffClinicAssignment{},
		&model.ReservationType{}, &model.StaffReservationExclusion{},
	))
	// 親 reservation_types を CASCADE で初期化（exclusions も連鎖クリア）。
	db.Exec("TRUNCATE TABLE staff_reservation_exclusions, staff_clinic_assignments, reservation_types, staffs CASCADE")
	return db
}

// TestReservationStaffRepository_UpdateExcludedReservationTypes_ClinicIsolation は
// clinic A のスタッフに clinic B の予約区分を除外コースとして書き込めないことを検証する。
// 型IDの clinic_id 検証を削除すると「別クリニックの区分IDは拒否される」が失敗する。
func TestReservationStaffRepository_UpdateExcludedReservationTypes_ClinicIsolation(t *testing.T) {
	db := setupExclusionIsolationTestDB(t)
	repo := NewReservationStaffRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	// clinic A に現在所属するスタッフ。repository は置換と所属解除を直列化するため、
	// staff identity と active assignment の双方を lock する。
	staffA := makeDoctorAssignedToClinic(t, db, clinicA, "除外テストA用スタッフ")

	typeA := makeReservationType(t, db, clinicA)
	typeB := makeReservationType(t, db, clinicB)

	countExclusions := func(staffID uint64) int64 {
		var n int64
		require.NoError(t, db.Model(&model.StaffReservationExclusion{}).
			Where("staff_id = ?", staffID).Count(&n).Error)
		return n
	}

	t.Run("別クリニックの区分IDは拒否され、行が永続化されない（clinic_id 隔離）", func(t *testing.T) {
		err := repo.UpdateExcludedReservationTypes(ctx, clinicA, staffA.ID, []uint64{typeB.ID})
		require.Error(t, err, "clinic A のスタッフに clinic B の予約区分を除外設定できてはならない")
		assert.Zero(t, countExclusions(staffA.ID), "拒否時に staff_reservation_exclusions 行を残してはならない")
	})

	t.Run("同一クリニックの区分IDは許可され、行が永続化される", func(t *testing.T) {
		err := repo.UpdateExcludedReservationTypes(ctx, clinicA, staffA.ID, []uint64{typeA.ID})
		require.NoError(t, err)
		assert.Equal(t, int64(1), countExclusions(staffA.ID), "同一クリニックの除外設定は1件保存されるべき")
	})

	t.Run("一部が別クリニックの区分なら全体を拒否する（部分書き込み防止）", func(t *testing.T) {
		// staffA は前ケースで typeA の除外1件を持つ。混在入力は拒否され、既存も変化しない。
		err := repo.UpdateExcludedReservationTypes(ctx, clinicA, staffA.ID, []uint64{typeA.ID, typeB.ID})
		require.Error(t, err, "clinic B の区分が混在する場合は全体を拒否すべき")
		assert.Equal(t, int64(1), countExclusions(staffA.ID), "拒否時に既存の除外設定を破壊してはならない")
	})
}

// TestReservationStaffRepository_UpdateExcludedReservationTypes_DeleteScopedToClinic は
// BE-refactor.md H-2 の回帰テスト。多施設所属スタッフ（staff_clinic_assignments で
// clinic A/B 双方に所属）が clinic B で正当に持つ除外設定（staff_reservation_exclusions）が、
// clinic A での保存操作（DELETE + INSERT の全置換）によって無警告で削除されないことを検証する。
// staff_reservation_exclusions は自前 clinic_id を持たないため、DELETE を
// staff_id のみでスコープすると（修正前の実装）、clinic を跨いで他クリニック分の
// 除外設定まで消えてしまう。DELETE を reservation_types の clinic_id サブクエリでスコープすると
// このテストは PASS する。
func TestReservationStaffRepository_UpdateExcludedReservationTypes_DeleteScopedToClinic(t *testing.T) {
	db := setupExclusionIsolationTestDB(t)
	repo := NewReservationStaffRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	// 多施設所属スタッフ: 主所属は clinic A、加えて clinic B にも所属する。
	staff := makeDoctorAssignedToClinic(t, db, clinicA, "多施設所属スタッフ")
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinicB,
	}).Error)

	typeA := makeReservationType(t, db, clinicA)
	typeB := makeReservationType(t, db, clinicB)

	// clinic B での過去の正当な保存操作を模して、直接 clinic B の予約区分への
	// 除外設定を紐づけておく（repo.UpdateExcludedReservationTypes(clinicB, ...) を
	// 経由しても同じ状態になる）。
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationExclusion{
		StaffID: staff.ID, ReservationTypeID: typeB.ID,
	}).Error)

	exclusionExists := func(reservationTypeID uint64) bool {
		var n int64
		require.NoError(t, db.Model(&model.StaffReservationExclusion{}).
			Where("staff_id = ? AND reservation_type_id = ?", staff.ID, reservationTypeID).Count(&n).Error)
		return n == 1
	}

	// clinic A での保存操作（clinic A の区分へ差し替え）。
	err := repo.UpdateExcludedReservationTypes(ctx, clinicA, staff.ID, []uint64{typeA.ID})
	require.NoError(t, err)

	assert.True(t, exclusionExists(typeB.ID), "clinic B の除外設定は clinic A の保存操作で削除されてはならない")
	assert.True(t, exclusionExists(typeA.ID), "clinic A の新規除外設定は保存されるべき")
}
