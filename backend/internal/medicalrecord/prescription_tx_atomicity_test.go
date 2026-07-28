package medicalrecord

// tx_atomicity_test.go — BE-refactor.md H-8e: prescriptionRepository.Delete
// の tx 内原子性（DB-backed）。
//
// Delete が caller の ambient transaction に参加し、削除後に同一 tx 内の後続処理（=
// finalize ロック確認や監査書込相当）が失敗したら削除がロールバックして対象 prescription が
// DB に残存することを実 DB で実証する。examination_repository_tx_atomicity_test.go（H-8d）と同型。
//
//   - この原子性は mock では検証不可能（#211 security M2 の教訓）。repository が dbOrTx で
//     ambient tx に join することが正本である。
//   - temp-revert RED: prescription_repository.go の Delete の dbOrTx(ctx, r.db) を
//     r.db.WithContext(ctx) に戻すと、削除が独立 tx で即 commit され、ambient tx の rollback
//     では巻き戻らない → 削除後の prescription が消えたままになり RollsBack ケースが RED になる。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ambient tx 内で削除が成功した直後に後続処理が失敗 → 削除がロールバックし対象 prescription が残存する。
// fail-closed の DB-backed 正本テスト。
func TestPrescriptionRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A（Delete原子性RB）")
	p := makePrescription(t, db, clinicA, ownerA.ID, nil, time.Now())

	repo := NewPrescriptionRepository(db)

	sentinel := errors.New("simulated post-delete failure")
	txErr := withTx(ctx, db, func(txCtx context.Context) error {
		if e := repo.Delete(txCtx, clinicA, p.ID); e != nil {
			return e
		}
		return sentinel // fail-closed: 削除後の後続失敗 → tx を中断
	})
	require.Error(t, txErr, "ambient tx 内の後続失敗で WithTx はエラーを返す")

	// 削除はロールバックされ、prescription が残存する。
	found, e := repo.FindByID(ctx, clinicA, p.ID)
	require.NoError(t, e, "後続失敗で削除はロールバックされ prescription が残存しなければならない（fail-closed）")
	assert.Equal(t, p.ID, found.ID)
}

// ambient tx 内で削除が成功し commit されると prescription が永続化され消える（原子コミット）。
func TestPrescriptionRepository_Delete_CommitsWithinAmbientTx(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A（Delete原子性CM）")
	p := makePrescription(t, db, clinicA, ownerA.ID, nil, time.Now())

	repo := NewPrescriptionRepository(db)

	txErr := withTx(ctx, db, func(txCtx context.Context) error {
		return repo.Delete(txCtx, clinicA, p.ID)
	})
	require.NoError(t, txErr)

	_, e := repo.FindByID(ctx, clinicA, p.ID)
	require.Error(t, e, "commit 後は prescription が削除されている")
	assert.True(t, apperrors.IsNotFound(e))
}
