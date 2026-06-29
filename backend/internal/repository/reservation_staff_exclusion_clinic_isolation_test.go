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
	require.NoError(t, db.AutoMigrate(
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

	// clinic A のスタッフ。スタッフ所有権は呼び出し側（service/handler）で検証済みのため、
	// ここでは型ID（reservation_type）のテナント越境のみを対象に検証する。
	staffA := makeDoctor(t, db, clinicA, "除外テストA用スタッフ")

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
