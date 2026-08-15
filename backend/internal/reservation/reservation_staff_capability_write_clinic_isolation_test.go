package reservation

// reservation_staff_capability_write_clinic_isolation_test.go — G11-4 テスト負債解消
//
// テスト対象: ReservationStaffRepository.UpdateReservationCapabilities の clinic_id 境界
// 保護する不変条件: clinic A のスタッフの対応可能予約区分に clinic B の予約区分IDを
//   書き込めない（staff_reservation_capabilities は自前 clinic_id を持たず、
//   親 reservation_types 経由でのみテナント判定されるため、repository 層で
//   型IDの所有権を検証しなければクロステナント write になる）。
//   兄弟の UpdateExcludedReservationTypes には既に
//   reservation_staff_exclusion_clinic_isolation_test.go の isolation test があるが、
//   同一パターンの UpdateReservationCapabilities 側は動作証明が欠落していた。
//
// 型ID所有権検証（clinic_id = ? AND id IN ?）を削除すると必ず失敗するよう設計されている。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupCapabilityIsolationTestDB は対応可能予約区分 clinic_id 隔離テスト用の DB を整備する。
func setupCapabilityIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.ReservationType{}, &model.StaffReservationCapability{},
	))
	// 親 reservation_types を CASCADE で初期化（capabilities も連鎖クリア）。
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE staff_reservation_capabilities, staff_clinic_assignments, reservation_types, staffs CASCADE",
	).Error)
	seedClinicsForFK(t, db, 1, 2)
	return db
}

// makeDoctorAssignedToClinic はスタッフを1件作成し、指定クリニックへの所属
// （staff_clinic_assignments）を紐づける。UpdateReservationCapabilities は
// 内部で FindByID を呼びスタッフの当該クリニック所属を要求するため、
// 単純な makeDoctor だけでは "record not found" になる。
func makeDoctorAssignedToClinic(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	staff := makeDoctor(t, db, clinicID, name)
	require.NoError(t, db.WithContext(context.Background()).Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinicID, IsMain: true,
	}).Error)
	return staff
}

// TestReservationStaffRepository_UpdateReservationCapabilities_ClinicIsolation は
// clinic A のスタッフに clinic B の予約区分を対応可能コースとして書き込めないことを検証する。
// 型IDの clinic_id 検証を削除すると「別クリニックの区分IDは拒否される」が失敗する。
func TestReservationStaffRepository_UpdateReservationCapabilities_ClinicIsolation(t *testing.T) {
	db := setupCapabilityIsolationTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db))
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	staffA := makeDoctorAssignedToClinic(t, db, clinicA, "対応区分テストA用スタッフ")

	typeA := makeReservationType(t, db, clinicA)
	typeB := makeReservationType(t, db, clinicB)

	countCapabilities := func(staffID uint64) int64 {
		var n int64
		require.NoError(t, db.Model(&model.StaffReservationCapability{}).
			Where("staff_id = ?", staffID).Count(&n).Error)
		return n
	}

	t.Run("別クリニックの区分IDは拒否され、行が永続化されない（clinic_id 隔離）", func(t *testing.T) {
		err := repo.UpdateReservationCapabilities(ctx, clinicA, staffA.ID, []uint64{typeB.ID})
		require.Error(t, err, "clinic A のスタッフに clinic B の予約区分を対応可能設定できてはならない")
		assert.Zero(t, countCapabilities(staffA.ID), "拒否時に staff_reservation_capabilities 行を残してはならない")
	})

	t.Run("同一クリニックの区分IDは許可され、行が永続化される", func(t *testing.T) {
		err := repo.UpdateReservationCapabilities(ctx, clinicA, staffA.ID, []uint64{typeA.ID})
		require.NoError(t, err)
		assert.Equal(t, int64(1), countCapabilities(staffA.ID), "同一クリニックの対応区分設定は1件保存されるべき")
	})

	t.Run("一部が別クリニックの区分なら全体を拒否する（DELETE前検証・部分書き込み防止）", func(t *testing.T) {
		// staffA は前ケースで typeA の対応区分1件を持つ。混在入力は拒否され、既存も変化しない。
		err := repo.UpdateReservationCapabilities(ctx, clinicA, staffA.ID, []uint64{typeA.ID, typeB.ID})
		require.Error(t, err, "clinic B の区分が混在する場合は全体を拒否すべき")
		assert.Equal(t, int64(1), countCapabilities(staffA.ID), "拒否時に既存の対応区分設定を破壊してはならない（count検証はDELETE実行前）")
	})
}

// TestReservationStaffRepository_SupportsReservationType はLIFF予約ゲートで使われる
// 対応可能判定の正/負ケースを固定する。
func TestReservationStaffRepository_SupportsReservationType(t *testing.T) {
	db := setupCapabilityIsolationTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "SupportsReservationTypeテスト用スタッフ")
	typeSupported := makeReservationType(t, db, clinicA)
	typeUnsupported := makeReservationType(t, db, clinicA)

	require.NoError(t, repo.UpdateReservationCapabilities(ctx, clinicA, staff.ID, []uint64{typeSupported.ID}))

	t.Run("対応可能な区分にはtrueを返す", func(t *testing.T) {
		ok, err := repo.SupportsReservationType(ctx, clinicA, staff.ID, typeSupported.ID)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("対応区分に含まれない区分にはfalseを返す", func(t *testing.T) {
		ok, err := repo.SupportsReservationType(ctx, clinicA, staff.ID, typeUnsupported.ID)
		require.NoError(t, err)
		assert.False(t, ok)
	})
}
