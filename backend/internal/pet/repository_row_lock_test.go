package pet

// repository_row_lock_test.go — Codex PR606 #26 再指摘: cage swap の死亡判定を
// 交換 transaction の行ロック下で行うための RowLockingRepository.LockByIDForUpdate 契約検証。
// - ambient tx 不在は fail-closed（autocommit では FOR UPDATE が文終了で解放され直列化にならない）
// - ambient tx 内では FOR UPDATE で行を取得する
// - clinic スコープ越境・未存在 id は NotFound
// hospitalizationRepository.LockByIDForUpdate（hospitalization_discharge_lock_concurrency_test.go）
// と同じ契約。実 Postgres テスト DB で検証する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestPetRepository_LockByIDForUpdate(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "行ロック飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "行ロックポチ")

	t.Run("ambient tx 不在は fail-closed", func(t *testing.T) {
		got, err := repo.LockByIDForUpdate(ctx, clinicA, petA.ID)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.False(t, apperrors.IsNotFound(err), "tx 不在は NotFound ではなく internal であるべき: %v", err)
	})

	t.Run("ambient tx 内では同一医院から行を取得できる", func(t *testing.T) {
		require.NoError(t, persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			got, err := repo.LockByIDForUpdate(txCtx, clinicA, petA.ID)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, petA.ID, got.ID)
			assert.Equal(t, ownerA.ID, got.OwnerID)
			return nil
		}))
	})

	t.Run("他院からは NotFound（clinic_id 越境防止）", func(t *testing.T) {
		require.NoError(t, persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			got, err := repo.LockByIDForUpdate(txCtx, clinicB, petA.ID)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.True(t, apperrors.IsNotFound(err), "他院 id は NotFound であるべき: %v", err)
			return nil
		}))
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		require.NoError(t, persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			_, err := repo.LockByIDForUpdate(txCtx, clinicA, uint64(999999))
			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			return nil
		}))
	})
}
