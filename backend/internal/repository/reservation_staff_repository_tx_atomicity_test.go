package repository

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
)

// setupReservationStaffTxAtomicityTestDB は reservation_staff 系 tx 原子性テスト用の DB を整備する。
// reservation_staff_exclusion_clinic_isolation_test.go の setupExclusionIsolationTestDB と同じ対象
// テーブル構成（staffs / staff_clinic_assignments / reservation_types / staff_reservation_exclusions）
// に加え、UpdateReservationCapabilities 用の staff_reservation_capabilities も整備する。
func setupReservationStaffTxAtomicityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, 
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
	repo := NewReservationStaffRepository(db)
	tx := NewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := &model.Staff{Name: "原子性テスト用スタッフ", StaffType: model.StaffTypeDoctor}

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
	repo := NewReservationStaffRepository(db)
	tx := NewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := &model.Staff{Name: "原子性テスト用スタッフ2", StaffType: model.StaffTypeDoctor}
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

	var exclusionCount int64
	require.NoError(t, db.Model(&model.StaffReservationExclusion{}).Where("staff_id = ?", staff.ID).Count(&exclusionCount).Error)
	assert.EqualValues(t, 1, exclusionCount, "commit 後は除外コースが永続化される")
}

// ─── UpdateExcludedReservationTypes 単体（DELETE→INSERT の原子性） ─────────────────

func TestReservationStaffRepository_UpdateExcludedReservationTypes_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db)
	tx := NewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctor(t, db, clinicA, "除外原子性テスト用スタッフ")
	typeA := makeReservationType(t, db, clinicA)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{typeA.ID}); err != nil {
			return err
		}
		return errSentinelReservationStaffTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelReservationStaffTx)

	var count int64
	require.NoError(t, db.Model(&model.StaffReservationExclusion{}).Where("staff_id = ?", staff.ID).Count(&count).Error)
	assert.Zero(t, count,
		"ambient tx 失敗時、UpdateExcludedReservationTypes の DELETE→INSERT はロールバックされる。"+
			"バグ時は独立 tx で即コミットされるため count=1 になり FAIL する（部分コミットの実証）")
}

func TestReservationStaffRepository_UpdateExcludedReservationTypes_CommitsWithinAmbientTx(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db)
	tx := NewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctor(t, db, clinicA, "除外原子性テスト用スタッフ2")
	typeA := makeReservationType(t, db, clinicA)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.UpdateExcludedReservationTypes(txCtx, clinicA, staff.ID, []uint64{typeA.ID})
	}))

	var count int64
	require.NoError(t, db.Model(&model.StaffReservationExclusion{}).Where("staff_id = ?", staff.ID).Count(&count).Error)
	assert.EqualValues(t, 1, count, "commit 後は除外コースが永続化される")
}

// ─── Update + UpdateExcludedReservationTypes 合成（reservationStaffService.Update の judgment call） ──
//
// reservationStaffService.Update は本ユニットで新たに WithTx で括られた
// （staff 本体更新 + 除外コース置換の原子性）。mock ベースの service テストは transactor が
// passthrough のため実際の DB ロールバックを検証できない。ここでは Update の経路が実際に踏む
// repo.Update + repo.UpdateExcludedReservationTypes を同一 ambient tx で直接検証する。

func TestReservationStaffRepository_UpdateThenUpdateExcludedReservationTypes_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupReservationStaffTxAtomicityTestDB(t)
	repo := NewReservationStaffRepository(db)
	tx := NewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctor(t, db, clinicA, "更新前の名前")

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
	repo := NewReservationStaffRepository(db)
	tx := NewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	staff := makeDoctor(t, db, clinicA, "更新前の名前2")
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
	assert.EqualValues(t, 1, exclusionCount, "commit 後は除外コースが永続化される")
}
