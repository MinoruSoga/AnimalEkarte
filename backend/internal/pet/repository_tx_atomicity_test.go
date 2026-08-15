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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const (
	deathAlreadyRecordedMessage = "死亡記録は既に登録されています"
	revivalNotRecordedMessage   = "死亡記録が登録されていないため解除できません"
)

func beginAmbientPetTransaction(t *testing.T, db *gorm.DB) (*gorm.DB, context.Context) {
	t.Helper()
	tx := db.Begin()
	require.NoError(t, tx.Error)
	return tx, persistence.WithTxValue(context.Background(), tx)
}

func runConcurrentPetLifecycleMutations(t *testing.T, mutations ...func() error) []error {
	t.Helper()

	type mutationResult struct {
		index int
		err   error
	}
	start := make(chan struct{})
	ready := make(chan struct{}, len(mutations))
	results := make(chan mutationResult, len(mutations))
	for index, mutation := range mutations {
		go func(index int, mutate func() error) {
			ready <- struct{}{}
			<-start
			results <- mutationResult{index: index, err: mutate()}
		}(index, mutation)
	}
	for range mutations {
		<-ready
	}
	close(start)

	errs := make([]error, len(mutations))
	for range mutations {
		result := <-results
		errs[result.index] = result.err
	}
	return errs
}

func requireLifecycleConflict(t *testing.T, err error, conflictMessage string) {
	t.Helper()

	require.True(t, apperrors.IsConflict(err), "loser must return conflict: %v", err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "conflict must retain AppError: %v", err)
	assert.Equal(t, conflictMessage, appErr.Message)
}

func requireOneSuccessOneConflict(t *testing.T, errs []error, conflictMessage string) int {
	t.Helper()

	successCount := 0
	conflictCount := 0
	winner := -1
	for i, err := range errs {
		if err == nil {
			successCount++
			winner = i
			continue
		}

		requireLifecycleConflict(t, err, conflictMessage)
		conflictCount++
	}
	require.Equal(t, 1, successCount)
	require.Equal(t, 1, conflictCount)
	return winner
}

func TestPetRepository_RecordDeath_ConcurrentRequestsPreserveWinner(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "死亡CAS飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "死亡CASペット")
	repo := NewRepository(db)
	requests := []struct {
		deceasedAt time.Time
		reason     string
	}{
		{
			deceasedAt: time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC),
			reason:     "first concurrent reason",
		},
		{
			deceasedAt: time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC),
			reason:     "second concurrent reason",
		},
	}

	errs := runConcurrentPetLifecycleMutations(
		t,
		func() error {
			return repo.RecordDeath(ctx, clinicA, pet.ID, requests[0].deceasedAt, requests[0].reason)
		},
		func() error {
			return repo.RecordDeath(ctx, clinicA, pet.ID, requests[1].deceasedAt, requests[1].reason)
		},
	)
	winner := requireOneSuccessOneConflict(t, errs, deathAlreadyRecordedMessage)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusDeceased, reloaded.Status)
	require.NotNil(t, reloaded.DeceasedAt)
	assert.True(t, requests[winner].deceasedAt.Equal(*reloaded.DeceasedAt))
	require.NotNil(t, reloaded.DeceasedReason)
	assert.Equal(t, requests[winner].reason, *reloaded.DeceasedReason)
}

func TestPetRepository_ClearDeath_ConcurrentRevivalRequests(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "復活CAS飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "復活CASペット")
	repo := NewRepository(db)
	require.NoError(t, repo.RecordDeath(
		ctx,
		clinicA,
		pet.ID,
		time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC),
		"revival setup reason",
	))

	errs := runConcurrentPetLifecycleMutations(
		t,
		func() error { return repo.ClearDeath(ctx, clinicA, pet.ID) },
		func() error { return repo.ClearDeath(ctx, clinicA, pet.ID) },
	)
	requireOneSuccessOneConflict(t, errs, revivalNotRecordedMessage)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, reloaded.Status)
	assert.Nil(t, reloaded.DeceasedAt)
	assert.Nil(t, reloaded.DeceasedReason)
}

func TestPetRepository_Update_LegacyDeathMapConcurrentRequestsPreserveWinner(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "死亡legacy CAS飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "死亡legacy CASペット")
	repo := NewRepository(db)
	requests := []struct {
		deceasedAt time.Time
		reason     string
		fields     map[string]any
	}{
		{
			deceasedAt: time.Date(2026, 7, 27, 11, 0, 0, 0, time.UTC),
			reason:     "first legacy reason",
			fields: map[string]any{
				"deceased_at":     time.Date(2026, 7, 27, 11, 0, 0, 0, time.UTC),
				"deceased_reason": "first legacy reason",
				"status":          model.PetStatusDeceased,
			},
		},
		{
			deceasedAt: time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
			reason:     "second legacy reason",
			fields: map[string]any{
				"deceased_at":     time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
				"deceased_reason": "second legacy reason",
				"status":          model.PetStatusDeceased,
			},
		},
	}

	errs := runConcurrentPetLifecycleMutations(
		t,
		func() error { return repo.Update(ctx, clinicA, pet.ID, requests[0].fields) },
		func() error { return repo.Update(ctx, clinicA, pet.ID, requests[1].fields) },
	)
	winner := requireOneSuccessOneConflict(t, errs, deathAlreadyRecordedMessage)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusDeceased, reloaded.Status)
	require.NotNil(t, reloaded.DeceasedAt)
	assert.True(t, requests[winner].deceasedAt.Equal(*reloaded.DeceasedAt))
	require.NotNil(t, reloaded.DeceasedReason)
	assert.Equal(t, requests[winner].reason, *reloaded.DeceasedReason)
}

func TestPetRepository_Update_LegacyRevivalMapConcurrentRequests(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "復活legacy CAS飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "復活legacy CASペット")
	repo := NewRepository(db)
	require.NoError(t, repo.RecordDeath(
		ctx,
		clinicA,
		pet.ID,
		time.Date(2026, 7, 27, 11, 0, 0, 0, time.UTC),
		"legacy revival setup",
	))
	requests := []map[string]any{
		{"deceased_at": nil, "deceased_reason": nil, "status": model.PetStatusAlive},
		{"deceased_at": nil, "deceased_reason": nil, "status": model.PetStatusAlive},
	}

	errs := runConcurrentPetLifecycleMutations(
		t,
		func() error { return repo.Update(ctx, clinicA, pet.ID, requests[0]) },
		func() error { return repo.Update(ctx, clinicA, pet.ID, requests[1]) },
	)
	requireOneSuccessOneConflict(t, errs, revivalNotRecordedMessage)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, reloaded.Status)
	assert.Nil(t, reloaded.DeceasedAt)
	assert.Nil(t, reloaded.DeceasedReason)
}

func TestPetRepository_DeathAndRevival_CrossClinicDoesNotMutate(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "死亡CAS越境飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "死亡CAS越境ペット")
	repo := NewRepository(db)

	err := repo.RecordDeath(
		ctx,
		clinicB,
		pet.ID,
		time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC),
		"cross-clinic reason",
	)
	requireLifecycleConflict(t, err, deathAlreadyRecordedMessage)

	reloaded, findErr := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, findErr)
	assert.Equal(t, model.PetStatusAlive, reloaded.Status)
	assert.Nil(t, reloaded.DeceasedAt)
	assert.Nil(t, reloaded.DeceasedReason)

	deceasedAt := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	const reason = "same-clinic death reason"
	require.NoError(t, repo.RecordDeath(ctx, clinicA, pet.ID, deceasedAt, reason))

	err = repo.ClearDeath(ctx, clinicB, pet.ID)
	requireLifecycleConflict(t, err, revivalNotRecordedMessage)

	reloaded, findErr = repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, findErr)
	assert.Equal(t, model.PetStatusDeceased, reloaded.Status)
	require.NotNil(t, reloaded.DeceasedAt)
	assert.True(t, deceasedAt.Equal(*reloaded.DeceasedAt))
	require.NotNil(t, reloaded.DeceasedReason)
	assert.Equal(t, reason, *reloaded.DeceasedReason)
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
