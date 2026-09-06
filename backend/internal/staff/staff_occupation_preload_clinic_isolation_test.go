package staff_test

// staff_occupation_preload_clinic_isolation_test.go
// P0 lint follow-up（read 監査と同じ bar）: staff の Staff.Occupation Preload の clinic_id 述語が
// 実際にクロステナント混入を防ぐことを runtime で実証する。
//
// 対象修正サイト（8a51c2eb が clinic_id 述語を付与）:
//   - staffRepository.FindAll : Preload("Occupation", "clinic_id = ? AND deleted_at IS NULL", clinicID)
// base クエリは staff_clinic_assignments で clinic-scoped だが、スタッフの occupation_id が別クリニックの
// 職種マスタを指す汚染データが存在すると、clinic_id 述語の無い Preload は別クリニックの職種名を
// 応答に混入させる（IDOR / read 漏洩）。静的 lint は述語の存在を、本テストは隔離の動作を保証する。
//
// 注: Doctor preload（多医院所属の reservation）は staff_preload_clinic_isolation_test.go が別途カバー。
// 本テストは occupations マスタ（clinic-scoped・単一クリニック述語）の隔離を対象とする。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
)

func setupStaffOccupationPreloadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	// setupTestDB は per-call で staff_type ENUM を DROP しない。
	// ensureAutoMigrated で companies・clinics（seedClinicsForFK の FK 親）・occupations・staffs・
	// assignments を整備する。fresh DB（CI）では FK 親テーブルも明示作成しないと存在しない。
	require.NoError(t, ensureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Occupation{}, &model.Staff{}, &model.StaffClinicAssignment{},
	))
	// クリーン状態にする。staffs / occupations CASCADE は両者を参照する全テーブル（reservations・
	// medical_records・audit_logs・reservation_type_occupations 等）も連鎖 TRUNCATE する点に注意。
	// 共有テスト DB は逐次実行で各テストが自前データを seed するため許容。失敗は即停止させる。
	require.NoError(t, db.Exec("TRUNCATE TABLE staff_clinic_assignments, staffs, occupations CASCADE").Error)
	return db
}

// TestStaffRepository_FindAll_CrossClinicOccupationPreloadIsolation は
// Staff.Occupation の clinic_id 述語が別クリニックの職種マスタ混入を防ぐことを動作で証明する。
// legit（同一clinicは Preload される）と cross（別clinicは nil）の対で、述語が「効いている」かつ
// 「正規データを壊していない」ことを同時に示す（anti-vacuous）。
func TestStaffRepository_FindAll_CrossClinicOccupationPreloadIsolation(t *testing.T) {
	db := setupStaffOccupationPreloadTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	seedClinicsForFK(t, db, clinicA, clinicB)
	occA := makeOccupation(t, db, clinicA, "医院Aの職種")
	occB := makeOccupation(t, db, clinicB, "医院Bの職種")

	// (legit) clinic A のスタッフ → clinic A の職種を参照（述語が正規データを壊さない）
	staffLegit := makeStaffWithOccupation(t, db, clinicA, occA.ID, "正規職種スタッフ")
	makeStaffClinicAssignment(t, db, staffLegit.ID, clinicA)
	// (cross) clinic A のスタッフに clinic B の職種IDを植え付け（汚染データ模擬）
	staffCross := makeStaffWithOccupation(t, db, clinicA, occB.ID, "越境職種スタッフ")
	makeStaffClinicAssignment(t, db, staffCross.ID, clinicA)

	staffs, total, err := repo.FindAll(ctx, clinicA, 1, 100)
	require.NoError(t, err)
	require.Equal(t, int64(2), total, "clinic A 配属スタッフは2名取得されるべき")

	byID := make(map[uint64]*model.Staff, len(staffs))
	for i := range staffs {
		byID[staffs[i].ID] = &staffs[i]
	}

	legit := byID[staffLegit.ID]
	require.NotNil(t, legit, "同一クリニックのスタッフが取得されるべき")
	require.NotNil(t, legit.Occupation, "同一クリニックの Occupation は Preload されるべき")
	assert.Equal(t, occA.ID, legit.Occupation.ID)

	cross := byID[staffCross.ID]
	require.NotNil(t, cross, "越境FK のスタッフ自体は base クエリで返るべき")
	assert.Nil(t, cross.Occupation, "別クリニックの Occupation マスタが混入してはならない")
}
