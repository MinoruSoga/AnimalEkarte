package reservation

// reservation_staff_capability_preload_clinic_isolation_test.go
// P0 lint follow-up（read 監査と同じ bar）: reservation_staff の StaffReservationCapability.ReservationType
// Preload の clinic_id 述語が実際にクロステナント混入を防ぐことを runtime で実証する。
//
// 対象修正サイト（8a51c2eb が clinic_id 述語を付与）:
//   - FindAllReservationCapabilities            : Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID)
//   - FindAllReservationCapabilitiesByStaffIDs  : 同上
// base クエリは clinic-scoped（WHERE clinic_id = ?）だが、能力行の reservation_type_id が別クリニックの
// 区分を指す汚染データ（#124/#125 型の write-FK 検証漏れ・過去データ）が存在すると、clinic_id 述語の
// 無い Preload は別クリニックの予約区分名を応答に混入させる（IDOR / read 漏洩）。
// 静的 lint（preload_clinic_scope_lint_test.go）は述語の存在を、本テストは隔離の動作を保証する。

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

func setupReservationStaffCapabilityPreloadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	// setupTestDB は staff_type / reservation_type_category ENUM を DROP CASCADE するため
	// staffs.staff_type・reservation_types.category 列が消える。AutoMigrate で再整備する。
	// companies・clinics（seedClinicsForFK の FK 親）と staff_clinic_assignments（下の TRUNCATE 対象）も
	// 明示作成する。fresh DB（CI）ではこれらを AutoMigrate しないと存在せず TRUNCATE / seed が失敗する。
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.ReservationType{},
		&model.StaffReservationCapability{}, &model.StaffClinicAssignment{},
	))
	// クリーン状態にする。staffs CASCADE は staffs を参照する全テーブル（reservations・
	// medical_records・audit_logs 等）も連鎖 TRUNCATE する点に注意。共有テスト DB は逐次実行で
	// 各テストが自前データを seed するため許容される。TRUNCATE 失敗は偽陽性を招くため即停止させる。
	require.NoError(t, db.Exec("TRUNCATE TABLE staff_reservation_capabilities, staff_clinic_assignments, reservation_types, staffs CASCADE").Error)
	return db
}

func makeStaffReservationCapability(t *testing.T, db *gorm.DB, clinicID, staffID, reservationTypeID uint64) *model.StaffReservationCapability {
	t.Helper()
	c := &model.StaffReservationCapability{
		ClinicID:          clinicID,
		StaffID:           staffID,
		ReservationTypeID: reservationTypeID,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

// TestReservationStaffRepository_Capabilities_CrossClinicReservationTypePreloadIsolation は
// StaffReservationCapability.ReservationType の clinic_id 述語が別クリニックのマスタ混入を防ぐことを
// 動作で証明する。legit（同一clinicは Preload される）と cross（別clinicは nil）の対で、
// 述語が「効いている」かつ「正規データを壊していない」ことを同時に示す（anti-vacuous）。
func TestReservationStaffRepository_Capabilities_CrossClinicReservationTypePreloadIsolation(t *testing.T) {
	db := setupReservationStaffCapabilityPreloadTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	seedClinicsForFK(t, db, clinicA, clinicB)
	staffA := makeDoctor(t, db, clinicA, "医院Aのスタッフ")

	typeA := makeReservationType(t, db, clinicA) // clinic A の正規区分
	typeB := makeReservationType(t, db, clinicB) // clinic B の区分（越境FK 植え付け用）

	// (legit) clinic A の能力行 → clinic A の区分を参照（述語が正規データを壊さない）
	makeStaffReservationCapability(t, db, clinicA, staffA.ID, typeA.ID)
	// (cross) clinic A の能力行に clinic B の区分IDを植え付け（#124/#125 型の汚染データ模擬）
	makeStaffReservationCapability(t, db, clinicA, staffA.ID, typeB.ID)

	assertPreloadIsolation := func(t *testing.T, items []model.StaffReservationCapability) {
		t.Helper()
		require.Len(t, items, 2, "clinic A の能力行は2件取得されるべき")
		byType := make(map[uint64]*model.StaffReservationCapability, len(items))
		for i := range items {
			byType[items[i].ReservationTypeID] = &items[i]
		}

		legit := byType[typeA.ID]
		require.NotNil(t, legit, "同一クリニックの能力行が取得されるべき")
		require.NotNil(t, legit.ReservationType, "同一クリニックの ReservationType は Preload されるべき")
		assert.Equal(t, typeA.ID, legit.ReservationType.ID)

		cross := byType[typeB.ID]
		require.NotNil(t, cross, "越境FK の能力行自体は base クエリで返るべき")
		assert.Nil(t, cross.ReservationType, "別クリニックの ReservationType マスタが混入してはならない")
	}

	t.Run("FindAllReservationCapabilities", func(t *testing.T) {
		items, err := repo.FindAllReservationCapabilities(ctx, clinicA, staffA.ID)
		require.NoError(t, err)
		assertPreloadIsolation(t, items)
	})

	t.Run("FindAllReservationCapabilitiesByStaffIDs", func(t *testing.T) {
		items, err := repo.FindAllReservationCapabilitiesByStaffIDs(ctx, clinicA, []uint64{staffA.ID})
		require.NoError(t, err)
		assertPreloadIsolation(t, items)
	})
}
