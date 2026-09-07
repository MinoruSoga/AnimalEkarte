package reservation

// BE9-2C R②: 実装は internal/reservation へ移動済み。本テストは意図的に repository 残置 —
// production ctor（facade）が staff domain の staffs write 書き込み者を注入する配線そのものを
// 実 DB で検証するため、両 domain を構築できる本 package が唯一の置き場（count_clinic_scope 先例）。

// reservation_staff_repository_test.go — #212 カバレッジ向上（ローカル実測 0% のメソッド群）
//
// 対象: FindAll / UpdateSortOrder /
//       FindAllExcludedReservationTypes / FindAllExcludedReservationTypesByStaffIDs
//
// 優先順位は異常系・clinic_id 隔離・FK 検証（#212 の明示要件）。正常系の重複でカバレッジを稼がない。
// ヘルパー命名衝突を避けるため、本ファイル専有のトップレベル識別子は
// "ForReservationStaffRepoTest" サフィックスまたは "ReservationStaffRepoTest" を含む一意名を用いる。

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

// setupReservationStaffRepoTestDB は本ファイル専用の DB セットアップ。
// staff_reservation_exclusions / staff_reservation_capabilities / staff_clinic_assignments /
// reservation_types / staffs を CASCADE TRUNCATE する（reservation_types・staffs CASCADE で
// appointments も連鎖クリアされる。既存 setupCapabilityIsolationTestDB / setupExclusionIsolationTestDB
// と同一対象テーブル構成 + Reservation を追加）。
func setupReservationStaffRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{},
		&model.Staff{}, &model.StaffClinicAssignment{},
		&model.ReservationType{}, &model.Reservation{},
		&model.StaffReservationExclusion{}, &model.StaffReservationCapability{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE staff_reservation_exclusions, staff_reservation_capabilities, "+
			"staff_clinic_assignments, reservation_types, staffs CASCADE").Error)
	seedClinicsForFK(t, db, 1, 2, 3)
	return db
}

// ─── FindAll ────────────────────────────────────────────────────────────────

func TestReservationStaffRepository_FindAll(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
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

	t.Run("ソフトデリート済み所属だけのスタッフは含まれない", func(t *testing.T) {
		staff := makeDoctorAssignedToClinic(t, db, clinicA, "FindAll所属解除テスト用スタッフ")
		require.NoError(t, db.WithContext(ctx).
			Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicA).
			Delete(&model.StaffClinicAssignment{}).Error)

		staffs, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, item := range staffs {
			assert.NotEqual(t, staff.ID, item.ID, "soft-delete 済み所属は一覧権限を与えてはならない")
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

// ─── UpdateSortOrder ──────────────────────────────────────────────────────

func TestReservationStaffRepository_UpdateSortOrder(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
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
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	// Stage B: exclusion GET is derived from capabilities (universe \ capable).
	// Seed only rt1/rt2 for this staff's universe; capable empty ⇒ both excluded.
	staff := makeDoctorAssignedToClinic(t, db, clinicA, "除外一覧単数テスト用スタッフ")
	rt1 := makeReservationType(t, db, clinicA)
	rt2 := makeReservationType(t, db, clinicA)
	otherStaff := makeDoctorAssignedToClinic(t, db, clinicA, "除外一覧単数対象外スタッフ")
	// otherStaff is capable of both → derived excluded empty for otherStaff
	require.NoError(t, repo.UpdateReservationCapabilities(ctx, clinicA, otherStaff.ID, []uint64{rt1.ID, rt2.ID}))

	t.Run("指定staffIDの除外設定のみを返す", func(t *testing.T) {
		items, err := repo.FindAllExcludedReservationTypes(ctx, clinicA, staff.ID)
		require.NoError(t, err)
		require.Len(t, items, 2, "capable 空なら universe 全件が derived excluded")

		gotTypeIDs := map[uint64]bool{}
		for _, it := range items {
			gotTypeIDs[it.ReservationTypeID] = true
			require.NotNil(t, it.ReservationType, "ReservationType が付与されるべき")
		}
		assert.True(t, gotTypeIDs[rt1.ID])
		assert.True(t, gotTypeIDs[rt2.ID])
	})

	t.Run("全対応可能なstaffは derived excluded が空", func(t *testing.T) {
		items, err := repo.FindAllExcludedReservationTypes(ctx, clinicA, otherStaff.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("ソフトデリート済みReservationTypeはuniverseに入らない", func(t *testing.T) {
		rtDeleted := makeReservationType(t, db, clinicA)
		s := makeDoctorAssignedToClinic(t, db, clinicA, "除外一覧削除区分テスト用スタッフ")
		// capable of all non-deleted types → empty excluded after delete of rtDeleted
		require.NoError(t, repo.UpdateReservationCapabilities(ctx, clinicA, s.ID, []uint64{rt1.ID, rt2.ID, rtDeleted.ID}))
		require.NoError(t, db.WithContext(ctx).Delete(&model.ReservationType{}, rtDeleted.ID).Error)

		items, err := repo.FindAllExcludedReservationTypes(ctx, clinicA, s.ID)
		require.NoError(t, err)
		// still capable of rt1/rt2; universe no longer includes deleted → empty excluded
		assert.Empty(t, items, "削除済み予約区分は derived excluded に出さない")
	})
}

func TestReservationStaffRepository_FindAllExcludedReservationTypes_DoesNotLeakSharedStaffData(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	staff := makeDoctor(t, db, clinicA, "除外一覧多施設スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)
	makeStaffClinicAssignment(t, db, staff.ID, clinicB)
	typeA := makeReservationType(t, db, clinicA)
	typeB := makeReservationType(t, db, clinicB)
	// capable of nothing in either clinic → clinic A derived excluded is only typeA
	items, err := repo.FindAllExcludedReservationTypes(ctx, clinicA, staff.ID)

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, typeA.ID, items[0].ReservationTypeID)
	require.NotNil(t, items[0].ReservationType)
	assert.Equal(t, clinicA, items[0].ReservationType.ClinicID)
	assert.NotEqual(t, typeB.ID, items[0].ReservationTypeID)
	_ = typeB
}

func TestReservationStaffRepository_FindAllExcludedReservationTypesByStaffIDs(t *testing.T) {
	db := setupReservationStaffRepoTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staff1 := makeDoctorAssignedToClinic(t, db, clinicA, "除外一覧複数テスト用スタッフ1")
	staff2 := makeDoctorAssignedToClinic(t, db, clinicA, "除外一覧複数テスト用スタッフ2")
	staff3 := makeDoctorAssignedToClinic(t, db, clinicA, "除外一覧複数対象外スタッフ")
	rt1 := makeReservationType(t, db, clinicA)
	rt2 := makeReservationType(t, db, clinicA)

	// staff1 capable of rt2 only → excluded {rt1}
	// staff2 capable of rt1 only → excluded {rt2}
	// staff3 capable of both → excluded empty
	require.NoError(t, repo.UpdateReservationCapabilities(ctx, clinicA, staff1.ID, []uint64{rt2.ID}))
	require.NoError(t, repo.UpdateReservationCapabilities(ctx, clinicA, staff2.ID, []uint64{rt1.ID}))
	require.NoError(t, repo.UpdateReservationCapabilities(ctx, clinicA, staff3.ID, []uint64{rt1.ID, rt2.ID}))

	t.Run("指定staffIDsの除外設定のみを一括取得する（N+1回避）", func(t *testing.T) {
		items, err := repo.FindAllExcludedReservationTypesByStaffIDs(ctx, clinicA, []uint64{staff1.ID, staff2.ID})
		require.NoError(t, err)
		require.Len(t, items, 2, "staff1・staff2の derived excluded のみ（staff3は対象外）")

		gotStaffIDs := map[uint64]bool{}
		for _, it := range items {
			gotStaffIDs[it.StaffID] = true
		}
		assert.True(t, gotStaffIDs[staff1.ID])
		assert.True(t, gotStaffIDs[staff2.ID])
		assert.False(t, gotStaffIDs[staff3.ID])
	})

	t.Run("空のstaffIDsはnilを返しエラーなし", func(t *testing.T) {
		items, err := repo.FindAllExcludedReservationTypesByStaffIDs(ctx, clinicA, []uint64{})
		require.NoError(t, err)
		assert.Nil(t, items)
	})

	t.Run("全対応可能なstaffIDsは空", func(t *testing.T) {
		items, err := repo.FindAllExcludedReservationTypesByStaffIDs(ctx, clinicA, []uint64{staff3.ID})
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}
