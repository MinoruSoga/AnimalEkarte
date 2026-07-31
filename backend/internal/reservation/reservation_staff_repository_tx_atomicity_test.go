package reservation

// reservation_staff_repository_tx_atomicity_test.go — BE-refactor.md X-8 の DB-backed 原子性証明
//
// 背景: reservationStaffService.Create は s.repo.Create + s.repo.UpdateExcludedReservationTypes を
// Transactor.WithTx で括り、「staff+assignment 作成」と「除外コース設定」を原子化しようとしていた。
// しかし reservation_staff_repository.go の Create / UpdateExcludedReservationTypes /
// UpdateReservationCapabilities / UpdateSortOrder はそれぞれ独自に r.db.WithContext(ctx).Transaction(...)
// を呼び、ambient tx（WithTx が ctx に埋め込む tx）に一切参加しない独立トランザクションを開始していた。
// このため Create が内部 tx で先にコミットされたあと UpdateExcludedReservationTypes が失敗しても、
// 外側の WithTx の「ロールバック」は何も巻き戻さず、除外コース未設定の孤児 staff が残っていた。
//
// reservationStaffService.Update も同様に、staff 本体の update と除外コース置換を WithTx で
// 括っていなかった（そもそも原子性を主張していなかった）ため、本ユニットで新たに WithTx で括った。
//
// temp-revert RED の手順:
//   - reservation_staff_repository.go の対象メソッドを dbOrTx(ctx, r.db) → r.db.WithContext(ctx) に戻す
//     （Transaction 呼び出しは dbOrTx(ctx, r.db).Transaction(...) → r.db.WithContext(ctx).Transaction(...)）
//   - reservation_staff_service.go の Update から s.transactor.WithTx ラップを外す
//   - 本ファイルのテストを実行 → Rollback 系が RED（書込が残ってしまう）
//   - 元に戻す → GREEN

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupReservationStaffTxAtomicityTestDB は reservation_staff 系 tx 原子性テスト用の DB を整備する。
// reservation_staff_exclusion_clinic_isolation_test.go の setupExclusionIsolationTestDB と同じ対象
// テーブル構成（staffs / staff_clinic_assignments / reservation_types / staff_reservation_exclusions）
// に加え、UpdateReservationCapabilities 用の staff_reservation_capabilities も整備する。
func setupReservationStaffTxAtomicityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Staff{}, &model.StaffClinicAssignment{},
		&model.ReservationType{}, &model.StaffReservationExclusion{},
		&model.StaffReservationCapability{},
	))
	db.Exec("TRUNCATE TABLE staff_reservation_exclusions, staff_reservation_capabilities, " +
		"staff_clinic_assignments, reservation_types, staffs CASCADE")
	return db
}

var errSentinelReservationStaffTx = errors.New("simulated post-write failure in ambient tx")

// ─── Create（X-8 の本体バグ: staff+assignment 作成と除外コース設定の原子性） ──────────

// TestReservationStaffRepository_Create_RollsBackWhenAmbientTxFails は
// reservationStaffService.Create が実際に踏む経路（repo.Create → repo.UpdateExcludedReservationTypes）を
// 同一 ambient tx 内で再現し、後段が失敗した場合に staff/assignment が残らないことを検証する。
//
// バグ時（r.db.WithContext(ctx).Transaction(...) が独立 tx を開く）は Create が独立 tx で
// 即コミットされるため、外側 WithTx の rollback を素通りして staff/assignment 行が残り FAIL する。
func TestReservationStaffRepository_Create_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db))
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := &model.Staff{
		ClinicID:  clinicA,
		Name:      "原子性テスト用スタッフ",
		StaffType: model.StaffTypeDoctor,
	}

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, staff, clinicA); err != nil {
			return err
		}
		// 存在しない reservation_type_id を渡し、Create 後段の UpdateExcludedReservationTypes を
		// 意図的に失敗させる（clinic_id 所有権検証で拒否される）。
		return repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{999999})
	})
	require.Error(t, txErr)

	var staffCount int64
	require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", staff.ID).Count(&staffCount).Error)
	assert.Zero(t, staffCount,
		"ambient tx 失敗時、Create による staff 作成はロールバックされる。"+
			"バグ時は独立 tx で即コミットされるため count=1 になり FAIL する（孤児 staff の実証）")

	var assignCount int64
	require.NoError(t, db.Model(&model.StaffClinicAssignment{}).Where("staff_id = ?", staff.ID).Count(&assignCount).Error)
	assert.Zero(t, assignCount, "ambient tx 失敗時、StaffClinicAssignment の作成もロールバックされる")
}

func TestReservationStaffRepository_Create_CommitsWithinAmbientTx(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db))
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := &model.Staff{
		ClinicID:  clinicA,
		Name:      "原子性テスト用スタッフ2",
		StaffType: model.StaffTypeDoctor,
	}
	typeA := makeReservationType(t, db, clinicA)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, staff, clinicA); err != nil {
			return err
		}
		return repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{typeA.ID})
	}))

	var staffCount int64
	require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", staff.ID).Count(&staffCount).Error)
	assert.EqualValues(t, 1, staffCount, "commit 後は staff が永続化される")

	// Stage B: exclusion facade writes capabilities (exclude typeA with universe={typeA}
	// ⇒ capable empty). Dual-write to exclusions must stay zero.
	var exclusionCount int64
	require.NoError(t, db.Model(&model.StaffReservationExclusion{}).Where("staff_id = ?", staff.ID).Count(&exclusionCount).Error)
	assert.Zero(t, exclusionCount, "Stage B: exclusions table must not receive production writes")
	// empty excluded list would write all capable; here excluded={typeA}=universe → capable empty.
	// Re-run with exclude-none to prove capability persistence under ambient tx:
	typeB := makeReservationType(t, db, clinicA)
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		// exclude typeA only → capable={typeB}
		return repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{typeA.ID})
	}))
	var capCount int64
	require.NoError(t, db.Model(&model.StaffReservationCapability{}).
		Where("clinic_id = ? AND staff_id = ?", clinicA, staff.ID).Count(&capCount).Error)
	assert.EqualValues(t, 1, capCount, "commit 後は capabilities が永続化される")
	_ = typeB
}

// ─── UpdateExcludedReservationTypes 単体（capability facade 原子性） ─────────────────

func TestReservationStaffRepository_UpdateExcludedReservationTypes_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db))
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "除外原子性テスト用スタッフ")
	typeA := makeReservationType(t, db, clinicA)
	_ = makeReservationType(t, db, clinicA) // so exclude typeA leaves ≥1 capable on success path

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{typeA.ID}); err != nil {
			return err
		}
		return errSentinelReservationStaffTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelReservationStaffTx)

	var exclCount int64
	require.NoError(t, db.Model(&model.StaffReservationExclusion{}).Where("staff_id = ?", staff.ID).Count(&exclCount).Error)
	assert.Zero(t, exclCount)
	var capCount int64
	require.NoError(t, db.Model(&model.StaffReservationCapability{}).
		Where("clinic_id = ? AND staff_id = ?", clinicA, staff.ID).Count(&capCount).Error)
	assert.Zero(t, capCount,
		"ambient tx 失敗時、UpdateExcludedReservationTypes の capability 置換はロールバックされる")
}

func TestReservationStaffRepository_UpdateExcludedReservationTypes_CommitsWithinAmbientTx(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db))
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "除外原子性テスト用スタッフ2")
	typeA := makeReservationType(t, db, clinicA)
	typeB := makeReservationType(t, db, clinicA)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		// exclude typeA → capable={typeB}
		return repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{typeA.ID})
	}))

	var exclCount int64
	require.NoError(t, db.Model(&model.StaffReservationExclusion{}).Where("staff_id = ?", staff.ID).Count(&exclCount).Error)
	assert.Zero(t, exclCount, "Stage B: no exclusion dual-write")
	var capCount int64
	require.NoError(t, db.Model(&model.StaffReservationCapability{}).
		Where("clinic_id = ? AND staff_id = ?", clinicA, staff.ID).Count(&capCount).Error)
	assert.EqualValues(t, 1, capCount, "commit 後は capabilities が永続化される")
	caps, err := repo.FindAllReservationCapabilities(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	require.Len(t, caps, 1)
	assert.Equal(t, typeB.ID, caps[0].ReservationTypeID)
}

// ─── Update + UpdateExcludedReservationTypes 合成（reservationStaffService.Update の judgment call） ──
//
// reservationStaffService.Update は本ユニットで新たに WithTx で括られた
// （staff 本体更新 + 除外コース置換の原子性）。mock ベースの service テストは transactor が
// passthrough のため実際の DB ロールバックを検証できない。ここでは Update の経路が実際に踏む
// repo.Update + repo.UpdateExcludedReservationTypes を同一 ambient tx で直接検証する。

func TestReservationStaffRepository_UpdateThenUpdateExcludedReservationTypes_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db))
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "更新前の名前")

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Update(txCtx, clinicA, staff.ID, map[string]any{"name": "更新後の名前"}); err != nil {
			return err
		}
		// 存在しない reservation_type_id で除外コース置換を失敗させる。
		return repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{999999})
	})
	require.Error(t, txErr)

	var reloaded model.Staff
	require.NoError(t, db.WithContext(ctx).First(&reloaded, staff.ID).Error)
	assert.Equal(t, "更新前の名前", reloaded.Name,
		"ambient tx 失敗時、Update による name 変更はロールバックされる。"+
			"バグ時は独立 tx で即コミットされるため name が更新後の値になり FAIL する（非原子な部分更新の実証）")
}

func TestReservationStaffRepository_UpdateThenUpdateExcludedReservationTypes_CommitsWithinAmbientTx(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db))
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "更新前の名前2")
	typeA := makeReservationType(t, db, clinicA)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Update(txCtx, clinicA, staff.ID, map[string]any{"name": "更新後の名前2"}); err != nil {
			return err
		}
		return repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{typeA.ID})
	}))

	var reloaded model.Staff
	require.NoError(t, db.WithContext(ctx).First(&reloaded, staff.ID).Error)
	assert.Equal(t, "更新後の名前2", reloaded.Name, "commit 後は name 変更が永続化される")

	var exclusionCount int64
	require.NoError(t, db.Model(&model.StaffReservationExclusion{}).Where("staff_id = ?", staff.ID).Count(&exclusionCount).Error)
	assert.Zero(t, exclusionCount, "Stage B: exclusion facade must not dual-write exclusions")
	// exclude typeA with universe={typeA} → capable empty is a successful affinity write
	excluded, err := repo.FindAllExcludedReservationTypes(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	require.Len(t, excluded, 1)
	assert.Equal(t, typeA.ID, excluded[0].ReservationTypeID)
}

// ─── Stage B leaf reads (DBOrTx) + optional exclusion facade composition ─────────────
//
// FindAllExcluded* are pure facades (universe \ capable) with no DBOrTx in their bodies.
// Ambient-tx participation lives in the leaf helpers below. Each leaf must observe
// uncommitted writes in the same WithTx; rollback must leave zero durable rows.
// A facade composition probe exercises FindAllExcluded* as a multi-leaf call path.

// TestReservationStaffRepository_LeafReads_SeeUncommittedAmbientWrites proves
// hasActiveClinicAssignment / filterStaffIDsWithActiveAssignment /
// listActiveReservationTypeUniverse / FindAllReservationCapabilities /
// FindAllReservationCapabilitiesByStaffIDs join ambient tx via DBOrTx.
func TestReservationStaffRepository_LeafReads_SeeUncommittedAmbientWrites(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db, staffpkg.NewStaffRepository(db)).(*reservationStaffRepository)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	// Staff exists outside ambient tx; assignment/type/capability are written only inside.
	staff := makeDoctor(t, db, clinicA, "leaf-read ambient staff")
	otherStaff := makeDoctor(t, db, clinicA, "leaf-read other staff")
	// otherStaff gets a committed assignment so filter can distinguish ambient-only assignment.
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID: otherStaff.ID, ClinicID: clinicA, IsMain: true,
	}).Error)

	// Pre-ambient: staff has no assignment → leaf returns false / empty.
	assigned, err := repo.hasActiveClinicAssignment(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	assert.False(t, assigned, "precondition: no committed assignment for staff")

	forced := errors.New("forced leaf-read ambient rollback")
	var uncommittedTypeID, uncommittedCapTypeID uint64

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		// Uncommitted assignment for staff.
		if err := persistence.DBOrTx(txCtx, db).Create(&model.StaffClinicAssignment{
			StaffID: staff.ID, ClinicID: clinicA, IsMain: true,
		}).Error; err != nil {
			return err
		}
		// Uncommitted active reservation type (universe).
		rt := &model.ReservationType{
			ClinicID: clinicA,
			Name:     "uncommitted universe type",
			Category: model.ReservationTypeCategoryGeneral,
			IsActive: true,
		}
		if err := persistence.DBOrTx(txCtx, db).Create(rt).Error; err != nil {
			return err
		}
		uncommittedTypeID = rt.ID
		// Second type so exclusion facade has non-empty excluded when only one is capable.
		rt2 := &model.ReservationType{
			ClinicID: clinicA,
			Name:     "uncommitted capable type",
			Category: model.ReservationTypeCategoryGeneral,
			IsActive: true,
		}
		if err := persistence.DBOrTx(txCtx, db).Create(rt2).Error; err != nil {
			return err
		}
		uncommittedCapTypeID = rt2.ID
		// Uncommitted capability (staff can handle rt2 only).
		if err := persistence.DBOrTx(txCtx, db).Create(&model.StaffReservationCapability{
			ClinicID:          clinicA,
			StaffID:           staff.ID,
			ReservationTypeID: rt2.ID,
		}).Error; err != nil {
			return err
		}

		// ── leaf: hasActiveClinicAssignment ──
		ok, err := repo.hasActiveClinicAssignment(txCtx, clinicA, staff.ID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("hasActiveClinicAssignment did not see uncommitted assignment")
		}

		// ── leaf: filterStaffIDsWithActiveAssignment ──
		filtered, err := repo.filterStaffIDsWithActiveAssignment(txCtx, clinicA, []uint64{staff.ID, otherStaff.ID})
		if err != nil {
			return err
		}
		if len(filtered) != 2 {
			return errors.New("filterStaffIDsWithActiveAssignment missed ambient and/or committed assignment")
		}

		// ── leaf: listActiveReservationTypeUniverse ──
		universe, err := repo.listActiveReservationTypeUniverse(txCtx, clinicA)
		if err != nil {
			return err
		}
		if len(universe) < 2 {
			return errors.New("listActiveReservationTypeUniverse did not see uncommitted types")
		}
		foundUniverse := map[uint64]bool{}
		for _, u := range universe {
			foundUniverse[u.ID] = true
		}
		if !foundUniverse[rt.ID] || !foundUniverse[rt2.ID] {
			return errors.New("listActiveReservationTypeUniverse missing uncommitted type ids")
		}

		// ── leaf: FindAllReservationCapabilities ──
		caps, err := repo.FindAllReservationCapabilities(txCtx, clinicA, staff.ID)
		if err != nil {
			return err
		}
		if len(caps) != 1 || caps[0].ReservationTypeID != rt2.ID {
			return errors.New("FindAllReservationCapabilities did not see uncommitted capability")
		}

		// ── leaf: FindAllReservationCapabilitiesByStaffIDs ──
		bulk, err := repo.FindAllReservationCapabilitiesByStaffIDs(txCtx, clinicA, []uint64{staff.ID})
		if err != nil {
			return err
		}
		if len(bulk) != 1 || bulk[0].ReservationTypeID != rt2.ID {
			return errors.New("FindAllReservationCapabilitiesByStaffIDs did not see uncommitted capability")
		}

		// ── optional composition probe: FindAllExcluded* (Stage B pure facade) ──
		// universe={rt,rt2}, capable={rt2} ⇒ excluded={rt}
		excluded, err := repo.FindAllExcludedReservationTypes(txCtx, clinicA, staff.ID)
		if err != nil {
			return err
		}
		if len(excluded) != 1 || excluded[0].ReservationTypeID != rt.ID {
			return errors.New("FindAllExcludedReservationTypes composition did not see ambient leaves")
		}
		excludedBulk, err := repo.FindAllExcludedReservationTypesByStaffIDs(txCtx, clinicA, []uint64{staff.ID})
		if err != nil {
			return err
		}
		if len(excludedBulk) != 1 || excludedBulk[0].ReservationTypeID != rt.ID {
			return errors.New("FindAllExcludedReservationTypesByStaffIDs composition did not see ambient leaves")
		}

		return forced
	})
	require.ErrorIs(t, txErr, forced)

	// Outside ambient: uncommitted assignment/type/capability must not exist.
	assigned, err = repo.hasActiveClinicAssignment(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	assert.False(t, assigned, "assignment must roll back with ambient tx")

	var typeCount int64
	require.NoError(t, db.Model(&model.ReservationType{}).
		Where("id IN ?", []uint64{uncommittedTypeID, uncommittedCapTypeID}).
		Count(&typeCount).Error)
	assert.Zero(t, typeCount, "uncommitted reservation types must roll back")

	var capCount int64
	require.NoError(t, db.Model(&model.StaffReservationCapability{}).
		Where("clinic_id = ? AND staff_id = ?", clinicA, staff.ID).
		Count(&capCount).Error)
	assert.Zero(t, capCount, "uncommitted capabilities must roll back")

	// otherStaff's committed assignment survives.
	filtered, err := repo.filterStaffIDsWithActiveAssignment(ctx, clinicA, []uint64{staff.ID, otherStaff.ID})
	require.NoError(t, err)
	assert.Equal(t, []uint64{otherStaff.ID}, filtered)
}
