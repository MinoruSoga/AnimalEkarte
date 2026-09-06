package reservation

// reservation_staff_exclusion_clinic_isolation_test.go — Stage B: exclusion write facade
// maps to capabilities and must not dual-persist independent exclusion rows.
//
// 保護する不変条件:
//   - clinic A のスタッフに clinic B の予約区分IDを affinity write できない
//   - exclusion PUT は staff_reservation_exclusions に書かず capabilities のみ置換する
//   - clinic A の facade write は clinic B の capabilities を消さない

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

// setupExclusionIsolationTestDB は除外 facade clinic_id 隔離テスト用の DB を整備する。
func setupExclusionIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.ReservationType{}, &model.StaffReservationExclusion{}, &model.StaffReservationCapability{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE staff_reservation_capabilities, staff_reservation_exclusions, staff_clinic_assignments, reservation_types, staffs CASCADE",
	).Error)
	seedClinicsForFK(t, db, 1, 2)
	return db
}

// TestReservationStaffRepository_UpdateExcludedReservationTypes_ClinicIsolation は
// clinic A のスタッフに clinic B の予約区分を affinity 設定できないことを検証する。
func TestReservationStaffRepository_UpdateExcludedReservationTypes_ClinicIsolation(t *testing.T) {
	db := setupExclusionIsolationTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	staffA := makeDoctorAssignedToClinic(t, db, clinicA, "除外テストA用スタッフ")

	typeA := makeReservationType(t, db, clinicA)
	typeB := makeReservationType(t, db, clinicB)

	countExclusions := func(staffID uint64) int64 {
		var n int64
		require.NoError(t, db.Model(&model.StaffReservationExclusion{}).
			Where("staff_id = ?", staffID).Count(&n).Error)
		return n
	}
	countCapabilities := func(clinicID, staffID uint64) int64 {
		var n int64
		require.NoError(t, db.Model(&model.StaffReservationCapability{}).
			Where("clinic_id = ? AND staff_id = ?", clinicID, staffID).Count(&n).Error)
		return n
	}

	t.Run("別クリニックの区分IDは拒否され、行が永続化されない（clinic_id 隔離）", func(t *testing.T) {
		err := repo.UpdateExcludedReservationTypes(ctx, clinicA, staffA.ID, []uint64{typeB.ID})
		require.Error(t, err, "clinic A のスタッフに clinic B の予約区分を除外設定できてはならない")
		assert.Zero(t, countExclusions(staffA.ID), "拒否時に staff_reservation_exclusions 行を残してはならない")
		assert.Zero(t, countCapabilities(clinicA, staffA.ID), "拒否時に capabilities 行を残してはならない")
	})

	t.Run("同一クリニックの除外PUTはcapabilities置換のみ（exclusion dual-write禁止）", func(t *testing.T) {
		// universe = {typeA}; excluded = {typeA} → capable = empty
		err := repo.UpdateExcludedReservationTypes(ctx, clinicA, staffA.ID, []uint64{typeA.ID})
		require.NoError(t, err)
		assert.Zero(t, countExclusions(staffA.ID), "Stage B: exclusions テーブルへ書いてはならない")
		assert.Zero(t, countCapabilities(clinicA, staffA.ID), "全除外なら capable は空")

		// universe に typeA2 を追加し、typeA のみ除外 → capable = {typeA2}
		typeA2 := makeReservationType(t, db, clinicA)
		err = repo.UpdateExcludedReservationTypes(ctx, clinicA, staffA.ID, []uint64{typeA.ID})
		require.NoError(t, err)
		assert.Zero(t, countExclusions(staffA.ID))
		assert.Equal(t, int64(1), countCapabilities(clinicA, staffA.ID))

		caps, err := repo.FindAllReservationCapabilities(ctx, clinicA, staffA.ID)
		require.NoError(t, err)
		require.Len(t, caps, 1)
		assert.Equal(t, typeA2.ID, caps[0].ReservationTypeID)
	})

	t.Run("一部が別クリニックの区分なら全体を拒否する（部分書き込み防止）", func(t *testing.T) {
		before := countCapabilities(clinicA, staffA.ID)
		err := repo.UpdateExcludedReservationTypes(ctx, clinicA, staffA.ID, []uint64{typeA.ID, typeB.ID})
		require.Error(t, err, "clinic B の区分が混在する場合は全体を拒否すべき")
		assert.Equal(t, before, countCapabilities(clinicA, staffA.ID), "拒否時に既存 capabilities を破壊してはならない")
		assert.Zero(t, countExclusions(staffA.ID))
	})
}

// TestReservationStaffRepository_UpdateExcludedReservationTypes_DeleteScopedToClinic は
// Stage B で clinic A の capability facade write が clinic B の capabilities を消さないことを検証する。
func TestReservationStaffRepository_UpdateExcludedReservationTypes_DeleteScopedToClinic(t *testing.T) {
	db := setupExclusionIsolationTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "多施設所属スタッフ")
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinicB,
	}).Error)

	typeA := makeReservationType(t, db, clinicA)
	typeB := makeReservationType(t, db, clinicB)

	// clinic B 側の対応可能設定（capabilities SoT）
	require.NoError(t, repo.UpdateReservationCapabilities(ctx, clinicB, staff.ID, []uint64{typeB.ID}))

	// clinic A: exclude typeA → capable empty for clinic A universe
	err := repo.UpdateExcludedReservationTypes(ctx, clinicA, staff.ID, []uint64{typeA.ID})
	require.NoError(t, err)

	// clinic B capabilities remain
	capsB, err := repo.FindAllReservationCapabilities(ctx, clinicB, staff.ID)
	require.NoError(t, err)
	require.Len(t, capsB, 1)
	assert.Equal(t, typeB.ID, capsB[0].ReservationTypeID)

	// no dual-write to exclusions
	var exclN int64
	require.NoError(t, db.Model(&model.StaffReservationExclusion{}).
		Where("staff_id = ?", staff.ID).Count(&exclN).Error)
	assert.Zero(t, exclN)
}
