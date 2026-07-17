package repository

// pet_repository_tx_atomicity_test.go — ペット死亡/復活の status 更新 + 監査の tx 内原子性（DB-backed, BUG-407）
//
// lstepLifecycleService.HandlePetDeath/HandlePetRevival は petRepo.Update と一次監査ログ書込
// (AuditTxLogger.LogEntryTx) を同一 Transactor.WithTx に包み、監査書込が失敗したら status 更新も
// ロールバックする（fail-closed）。この原子性は mock では検証不可能（refund_tx_atomicity_test.go と
// 同じ教訓 — security M2）なので、実 DB で petRepo.Update が ambient tx に参加し、tx 内後続処理
// （= 監査書込の失敗を模倣）が失敗すると status 変更がロールバックされることを実証する。
//
// temp-revert RED: pet_repository.go の Update を dbOrTx(ctx, r.db) → r.db.WithContext(ctx) に
// 戻すと、Updates が独立 tx で即 commit され、ambient tx の rollback では巻き戻らない →
// RollsBackWhenAmbientTxFails が RED になる。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestPetRepository_Update_RollsBackWhenAmbientTxFails は
// ambient tx 内で pet の status を更新した直後に後続処理（監査書込を模倣）が失敗したとき、
// status 更新がロールバックされ alive のまま残ることを検証する（fail-closed）。
func TestPetRepository_Update_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "原子性テスト飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "原子性テストペット")

	repo := NewPetRepository(db)
	tx := NewTransactor(db)

	sentinel := errors.New("simulated post-update audit failure")
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		fields := map[string]any{
			"status": model.PetStatusDeceased,
		}
		if err := repo.Update(txCtx, clinicA, pet.ID, fields); err != nil {
			return err
		}
		return sentinel // fail-closed: Update 後に監査書込失敗を模倣 → tx 中断
	})
	require.Error(t, txErr, "ambient tx の後続失敗で WithTx はエラーを返す")
	require.ErrorIs(t, txErr, sentinel)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, reloaded.Status, "監査失敗で status 更新はロールバックされ alive のまま（fail-closed）")
}

// TestPetRepository_Update_CommitsWithinAmbientTx は
// ambient tx 内で pet の status 更新が成功し commit されると、status 変更が永続化されることを検証する。
func TestPetRepository_Update_CommitsWithinAmbientTx(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "コミットテスト飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "コミットテストペット")

	repo := NewPetRepository(db)
	tx := NewTransactor(db)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.Update(txCtx, clinicA, pet.ID, map[string]any{"status": model.PetStatusDeceased})
	}))

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusDeceased, reloaded.Status, "commit 後は status 更新が永続化される")
}
