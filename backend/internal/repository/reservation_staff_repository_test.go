package repository

// reservation_staff_repository_test.go — #212 カバレッジ向上（ローカル実測 0% のメソッド群）
//
// 対象: FindAll / Delete / CountUsageByStaffID / UpdateSortOrder /
//       FindAllExcludedReservationTypes / FindAllExcludedReservationTypesByStaffIDs
//
// 優先順位は異常系・clinic_id 隔離・FK 検証（#212 の明示要件）。正常系の重複でカバレッジを稼がない。
// ヘルパー命名衝突を避けるため、本ファイル専有のトップレベル識別子は
// "ForReservationStaffRepoTest" サフィックスまたは "ReservationStaffRepoTest" を含む一意名を用いる。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupReservationStaffRepoTestDB は本ファイル専用の DB セットアップ。
// staff_reservation_exclusions / staff_reservation_capabilities / staff_clinic_assignments /
// reservation_types / staffs を CASCADE TRUNCATE する（reservation_types・staffs CASCADE で
// appointments も連鎖クリアされる。既存 setupCapabilityIsolationTestDB / setupExclusionIsolationTestDB
// と同一対象テーブル構成 + Reservation を追加）。
func setupReservationStaffRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, 
		&model.Staff{}, &model.StaffClinicAssignment{},
		&model.ReservationType{}, &model.Reservation{},
		&model.StaffReservationExclusion{}, &model.StaffReservationCapability{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE staff_reservation_exclusions, staff_reservation_capabilities, "+
			"staff_clinic_assignments, reservation_types, staffs CASCADE").Error)
	return db
}

// makeReservationWithDoctorAtOffset は makeReservationWithDoctor と異なり start_time を
// offset で明示的にずらして作成する。appointments には
// uk_appointment_staff_time (clinic_id, doctor_id, start_time) の部分 UNIQUE INDEX
// (WHERE deleted_at IS NULL AND status != 'cancelled') があるため、同一クリニック・同一
// staff の予約を複数件作るテスト（CountUsageByStaffID）では start_time を必ず分散させる。
func makeReservationWithDoctorAtOffset(t *testing.T, db *gorm.DB, clinicID, reservationTypeID, doctorID uint64, offset time.Duration) *model.Reservation {
	t.Helper()
	start := time.Now().UTC().Truncate(time.Minute).Add(offset)
	did := doctorID
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           start.Add(15 * time.Minute),
		ReservationTypeID: reservationTypeID,
		DoctorID:          &did,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

// ─── FindAll ────────────────────────────────────────────────────────────────

func TestReservationStaffRepository_FindAll(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("clinic_id 隔離: 別クリニック所属のスタッフは含まれない", func(t *testing.T) {
		staffA := makeDoctorAssignedToClinic(t, db, clinicA, "FindAllテストA用スタッフ")
		staffB := makeDoctorAssignedToClinic(t, db, clinicB, "FindAllテストB用スタッフ")

		staffs, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)

		ids := make(map[uint64]bool, len(staffs))
		for _, s := range staffs {
			ids[s.ID] = true
		}
		assert.True(t, ids[staffA.ID], "clinic A のスタッフは含まれるべき")
		assert.False(t, ids[staffB.ID], "clinic B のスタッフは含まれてはならない")
	})

	t.Run("ソフトデリート済みスタッフは含まれない", func(t *testing.T) {
		staff := makeDoctorAssignedToClinic(t, db, clinicA, "FindAllソフトデリートテスト用スタッフ")
		require.NoError(t, db.WithContext(ctx).Delete(&model.Staff{}, staff.ID).Error)

		staffs, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, s := range staffs {
			assert.NotEqual(t, staff.ID, s.ID, "ソフトデリート済みスタッフは一覧に含まれてはならない")
		}
	})

	t.Run("多施設所属スタッフはどちらのクリニックからも取得できる", func(t *testing.T) {
		staff := makeDoctor(t, db, clinicA, "FindAll多施設所属テスト用スタッフ")
		makeStaffClinicAssignment(t, db, staff.ID, clinicA)
		makeStaffClinicAssignment(t, db, staff.ID, clinicB)

		staffsA, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		staffsB, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)

		containsID := func(list []model.Staff, id uint64) bool {
			for _, s := range list {
				if s.ID == id {
					return true
				}
			}
			return false
		}
		assert.True(t, containsID(staffsA, staff.ID), "clinic A からも取得できるべき")
		assert.True(t, containsID(staffsB, staff.ID), "clinic B からも取得できるべき")
	})
}

// ─── Delete ─────────────────────────────────────────────────────────────────

func TestReservationStaffRepository_Delete(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックからは削除成功しソフトデリートされる", func(t *testing.T) {
		staff := makeDoctorAssignedToClinic(t, db, clinicA, "Deleteテスト用スタッフ")

		err := repo.Delete(ctx, clinicA, staff.ID)
		require.NoError(t, err)

		var count int64
		require.NoError(t, db.Unscoped().Model(&model.Staff{}).Where("id = ? AND deleted_at IS NOT NULL", staff.ID).Count(&count).Error)
		assert.EqualValues(t, 1, count, "対象行は物理削除でなくソフトデリートされるべき")
	})

	// Fixed: Delete は EXISTS(...) 相関サブクエリで clinic_id 隔離を検証する
	// （reservation_staff_repository.go 参照）。旧 Joins()+Delete() 実装は
	// GORM のソフトデリート UPDATE で JOIN 条件が無視されるバグがあったが、
	// 現行の EXISTS サブクエリはただの WHERE 述語であり、同様の欠落は発生しない。
	t.Run("別クリニックからの削除はNotFoundで行が残る", func(t *testing.T) {
		staff := makeDoctorAssignedToClinic(t, db, clinicA, "Delete別クリニックテスト用スタッフ")

		err := repo.Delete(ctx, clinicB, staff.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var count int64
		require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", staff.ID).Count(&count).Error)
		assert.EqualValues(t, 1, count, "別クリニックからの削除要求で行が消えてはならない")
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// ─── CountUsageByStaffID ──────────────────────────────────────────────────

func TestReservationStaffRepository_CountUsageByStaffID(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	staffA := makeDoctor(t, db, clinicA, "CountUsageテストA用スタッフ")
	otherStaff := makeDoctor(t, db, clinicA, "CountUsageテスト対象外スタッフ")
	rtA := makeReservationType(t, db, clinicA)
	rtB := makeReservationType(t, db, clinicB)

	// staffA 担当の予約を clinic A に2件作成（uk_appointment_staff_time 部分UNIQUE回避のため
	// start_time を分散させる）
	makeReservationWithDoctorAtOffset(t, db, clinicA, rtA.ID, staffA.ID, 0)
	res2 := makeReservationWithDoctorAtOffset(t, db, clinicA, rtA.ID, staffA.ID, time.Hour)

	// 別スタッフ担当の予約（カウントされないはず）
	makeReservationWithDoctorAtOffset(t, db, clinicA, rtA.ID, otherStaff.ID, 2*time.Hour)

	// staffA 担当だが clinic B の予約（clinic_id 隔離でカウントされないはず）
	makeReservationWithDoctorAtOffset(t, db, clinicB, rtB.ID, staffA.ID, 3*time.Hour)

	t.Run("同一クリニック内の担当予約数のみをカウントする", func(t *testing.T) {
		count, err := repo.CountUsageByStaffID(ctx, clinicA, staffA.ID)
		require.NoError(t, err)
		assert.EqualValues(t, 2, count, "clinic A 内の staffA 担当予約2件のみカウントされるべき")
	})

	t.Run("別クリニックの担当予約は対象クリニックのカウントに含まれない", func(t *testing.T) {
		count, err := repo.CountUsageByStaffID(ctx, clinicB, staffA.ID)
		require.NoError(t, err)
		assert.EqualValues(t, 1, count, "clinic B 側から見た staffA 担当予約は1件のみ")
	})

	t.Run("ソフトデリート済み予約はカウントされない", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(&model.Reservation{}, res2.ID).Error)

		count, err := repo.CountUsageByStaffID(ctx, clinicA, staffA.ID)
		require.NoError(t, err)
		assert.EqualValues(t, 1, count, "ソフトデリート済み予約を除いた1件になるべき")
	})

	t.Run("担当予約が無いスタッフは0件", func(t *testing.T) {
		unused := makeDoctor(t, db, clinicA, "CountUsage未使用スタッフ")
		count, err := repo.CountUsageByStaffID(ctx, clinicA, unused.ID)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
}

// ─── UpdateSortOrder ──────────────────────────────────────────────────────

func TestReservationStaffRepository_UpdateSortOrder(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	sortOrderOf := func(id uint64) int {
		var s model.Staff
		require.NoError(t, db.WithContext(ctx).First(&s, id).Error)
		return s.SortOrder
	}

	t.Run("up: 直前のスタッフとsort_orderを入れ替える", func(t *testing.T) {
		low := makeDoctorAssignedToClinic(t, db, clinicA, "SortOrder下位スタッフ")
		mid := makeDoctorAssignedToClinic(t, db, clinicA, "SortOrder中位スタッフ")
		high := makeDoctorAssignedToClinic(t, db, clinicA, "SortOrder上位スタッフ")
		require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", low.ID).Update("sort_order", 10).Error)
		require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", mid.ID).Update("sort_order", 20).Error)
		require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", high.ID).Update("sort_order", 30).Error)

		err := repo.UpdateSortOrder(ctx, clinicA, mid.ID, "up")
		require.NoError(t, err)

		assert.Equal(t, 10, sortOrderOf(mid.ID), "mid は low の元の値(10)を引き継ぐべき")
		assert.Equal(t, 20, sortOrderOf(low.ID), "low は mid の元の値(20)を引き継ぐべき")
		assert.Equal(t, 30, sortOrderOf(high.ID), "high は変化しないべき")
	})

	t.Run("down: 直後のスタッフとsort_orderを入れ替える", func(t *testing.T) {
		low := makeDoctorAssignedToClinic(t, db, clinicA, "SortOrder down下位スタッフ")
		high := makeDoctorAssignedToClinic(t, db, clinicA, "SortOrder down上位スタッフ")
		require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", low.ID).Update("sort_order", 100).Error)
		require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", high.ID).Update("sort_order", 200).Error)

		err := repo.UpdateSortOrder(ctx, clinicA, low.ID, "down")
		require.NoError(t, err)

		assert.Equal(t, 200, sortOrderOf(low.ID))
		assert.Equal(t, 100, sortOrderOf(high.ID))
	})

	t.Run("隣接するスタッフが存在しない場合は変更なし・エラーなし", func(t *testing.T) {
		// clinicA は本 Test 関数内の前段サブテスト(up/down)で作成済みの staff が残存し
		// sort_order 10〜200 の範囲に複数存在する（サブテスト間で TRUNCATE されない共有 DB 接続のため）。
		// 「隣接なし」を検証するには、それらと競合しない専用クリニックを使う。
		const clinicSortOrderIsolated = uint64(3)
		only := makeDoctorAssignedToClinic(t, db, clinicSortOrderIsolated, "SortOrder単独スタッフ")
		require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", only.ID).Update("sort_order", 500).Error)

		err := repo.UpdateSortOrder(ctx, clinicSortOrderIsolated, only.ID, "up")
		require.NoError(t, err)
		assert.Equal(t, 500, sortOrderOf(only.ID), "隣接なしの場合 sort_order は変化しないべき")
	})

	t.Run("別クリニックのIDを指定するとNotFound", func(t *testing.T) {
		staffB := makeDoctorAssignedToClinic(t, db, clinicB, "SortOrder別クリニックスタッフ")

		err := repo.UpdateSortOrder(ctx, clinicA, staffB.ID, "up")
		require.Error(t, err)
	})

	t.Run("存在しないIDを指定するとエラー", func(t *testing.T) {
		err := repo.UpdateSortOrder(ctx, clinicA, 999999, "up")
		require.Error(t, err)
	})
}

// ─── FindAllExcludedReservationTypes / FindAllExcludedReservationTypesByStaffIDs ──

func TestReservationStaffRepository_FindAllExcludedReservationTypes(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctor(t, db, clinicA, "除外一覧単数テスト用スタッフ")
	rt1 := makeReservationType(t, db, clinicA)
	rt2 := makeReservationType(t, db, clinicA)
	otherStaff := makeDoctor(t, db, clinicA, "除外一覧単数対象外スタッフ")

	require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationExclusion{
		StaffID: staff.ID, ReservationTypeID: rt1.ID,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationExclusion{
		StaffID: staff.ID, ReservationTypeID: rt2.ID,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationExclusion{
		StaffID: otherStaff.ID, ReservationTypeID: rt1.ID,
	}).Error)

	t.Run("指定staffIDの除外設定のみを返す", func(t *testing.T) {
		items, err := repo.FindAllExcludedReservationTypes(ctx, staff.ID)
		require.NoError(t, err)
		require.Len(t, items, 2, "staff の除外設定2件のみ返るべき（他staffの1件は含まれない）")

		gotTypeIDs := map[uint64]bool{}
		for _, it := range items {
			gotTypeIDs[it.ReservationTypeID] = true
			require.NotNil(t, it.ReservationType, "ReservationType が Preload されるべき")
		}
		assert.True(t, gotTypeIDs[rt1.ID])
		assert.True(t, gotTypeIDs[rt2.ID])
	})

	t.Run("除外設定が無いstaffIDは空", func(t *testing.T) {
		unrelated := makeDoctor(t, db, clinicA, "除外一覧単数無関係スタッフ")
		items, err := repo.FindAllExcludedReservationTypes(ctx, unrelated.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("ソフトデリート済みReservationTypeはPreloadされない", func(t *testing.T) {
		rtDeleted := makeReservationType(t, db, clinicA)
		s := makeDoctor(t, db, clinicA, "除外一覧削除区分テスト用スタッフ")
		require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationExclusion{
			StaffID: s.ID, ReservationTypeID: rtDeleted.ID,
		}).Error)
		require.NoError(t, db.WithContext(ctx).Delete(&model.ReservationType{}, rtDeleted.ID).Error)

		items, err := repo.FindAllExcludedReservationTypes(ctx, s.ID)
		require.NoError(t, err)
		require.Len(t, items, 1, "除外行自体はソフトデリートされた区分でも残るべき")
		assert.Nil(t, items[0].ReservationType, "deleted_at IS NULL 条件により削除済み区分はPreloadされないべき")
	})
}

func TestReservationStaffRepository_FindAllExcludedReservationTypesByStaffIDs(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff1 := makeDoctor(t, db, clinicA, "除外一覧複数テスト用スタッフ1")
	staff2 := makeDoctor(t, db, clinicA, "除外一覧複数テスト用スタッフ2")
	staff3 := makeDoctor(t, db, clinicA, "除外一覧複数対象外スタッフ")
	rt1 := makeReservationType(t, db, clinicA)
	rt2 := makeReservationType(t, db, clinicA)

	require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationExclusion{
		StaffID: staff1.ID, ReservationTypeID: rt1.ID,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationExclusion{
		StaffID: staff2.ID, ReservationTypeID: rt2.ID,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationExclusion{
		StaffID: staff3.ID, ReservationTypeID: rt1.ID,
	}).Error)

	t.Run("指定staffIDsの除外設定のみを一括取得する（N+1回避）", func(t *testing.T) {
		items, err := repo.FindAllExcludedReservationTypesByStaffIDs(ctx, []uint64{staff1.ID, staff2.ID})
		require.NoError(t, err)
		require.Len(t, items, 2, "staff1・staff2の除外設定のみ返るべき（staff3は含まれない）")

		gotStaffIDs := map[uint64]bool{}
		for _, it := range items {
			gotStaffIDs[it.StaffID] = true
		}
		assert.True(t, gotStaffIDs[staff1.ID])
		assert.True(t, gotStaffIDs[staff2.ID])
		assert.False(t, gotStaffIDs[staff3.ID])
	})

	t.Run("空のstaffIDsはnilを返しエラーなし", func(t *testing.T) {
		items, err := repo.FindAllExcludedReservationTypesByStaffIDs(ctx, []uint64{})
		require.NoError(t, err)
		assert.Nil(t, items)
	})

	t.Run("該当する除外設定がないstaffIDsは空", func(t *testing.T) {
		unrelated := makeDoctor(t, db, clinicA, "除外一覧複数無関係スタッフ")
		items, err := repo.FindAllExcludedReservationTypesByStaffIDs(ctx, []uint64{unrelated.ID})
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}
