package pet

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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

func beginAmbientPetTransaction(t *testing.T, db *gorm.DB) (*gorm.DB, context.Context) {
	t.Helper()
	tx := db.Begin()
	require.NoError(t, tx.Error)
	return tx, repohelpers.WithTxValue(context.Background(), tx)
}

func TestPetRepository_RecordDeath_RollsBackWithAmbientTransaction(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "死亡ロールバック飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "死亡ロールバックペット")
	repo := NewRepository(db)
	deceasedAt := time.Date(2026, 7, 24, 10, 30, 0, 0, time.UTC)

	tx, txCtx := beginAmbientPetTransaction(t, db)
	require.NoError(t, repo.RecordDeath(txCtx, clinicA, pet.ID, deceasedAt, "test reason"))
	require.NoError(t, tx.Rollback().Error)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, reloaded.Status)
	assert.Nil(t, reloaded.DeceasedAt)
	assert.Nil(t, reloaded.DeceasedReason)
}

func TestPetRepository_ClearDeath_RollsBackWithAmbientTransaction(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "復活ロールバック飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "復活ロールバックペット")
	repo := NewRepository(db)
	deceasedAt := time.Date(2026, 7, 24, 10, 30, 0, 0, time.UTC)
	const reason = "test reason"
	require.NoError(t, repo.RecordDeath(ctx, clinicA, pet.ID, deceasedAt, reason))

	tx, txCtx := beginAmbientPetTransaction(t, db)
	require.NoError(t, repo.ClearDeath(txCtx, clinicA, pet.ID))
	require.NoError(t, tx.Rollback().Error)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusDeceased, reloaded.Status)
	require.NotNil(t, reloaded.DeceasedAt)
	assert.True(t, deceasedAt.Equal(*reloaded.DeceasedAt))
	require.NotNil(t, reloaded.DeceasedReason)
	assert.Equal(t, reason, *reloaded.DeceasedReason)
}
