package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Runtime proof for lintscan dbOrTxParticipatingMethods enrollment of
// gormEnteredByActorGuard.AssertEnteredByActor / isActiveSystemAdminStaff:
// the SHARE lock on staffs must remain held until the ambient create tx commits.
func TestEnteredByActorGuard_Assert_HoldsStaffShareLockUntilAmbientTransactionCommits(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyEnteredBySingleColumnMigration(t, db)
	_, clinicB, actor, _, _ := seedCrossClinicEnteredByActor(t, db)

	guard := newGormEnteredByActorGuard(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ambientTx := db.WithContext(ctx).Begin()
	require.NoError(t, ambientTx.Error)
	ambientCommitted := false
	defer func() {
		if !ambientCommitted {
			_ = ambientTx.Rollback().Error
		}
	}()

	require.NoError(t, guard.AssertEnteredByActor(
		persistence.WithTxValue(ctx, ambientTx),
		clinicB,
		actor.ID,
		false,
	))

	competingTx := db.WithContext(ctx).Begin()
	require.NoError(t, competingTx.Error)
	require.NoError(t, competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	err := competingTx.Exec(
		"UPDATE staffs SET name = ? WHERE id = ?",
		"blocked before ambient commit",
		actor.ID,
	).Error
	require.ErrorContains(
		t,
		err,
		"lock timeout",
		"exclusive update to staffs must block while AssertEnteredByActor holds SHARE",
	)
	require.NoError(t, competingTx.Rollback().Error)

	require.NoError(t, ambientTx.Commit().Error)
	ambientCommitted = true

	afterCommitTx := db.WithContext(ctx).Begin()
	require.NoError(t, afterCommitTx.Error)
	defer afterCommitTx.Rollback()
	require.NoError(t, afterCommitTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	result := afterCommitTx.Exec(
		"UPDATE staffs SET name = ? WHERE id = ?",
		"succeeds after ambient commit",
		actor.ID,
	)
	require.NoError(t, result.Error)
	require.Equal(t, int64(1), result.RowsAffected)
	require.NoError(t, afterCommitTx.Commit().Error)
}

func TestEnteredByActorGuard_Assert_RequiresAmbientTransaction(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyEnteredBySingleColumnMigration(t, db)
	_, clinicB, actor, _, _ := seedCrossClinicEnteredByActor(t, db)

	guard := newGormEnteredByActorGuard(db)
	err := guard.AssertEnteredByActor(context.Background(), clinicB, actor.ID, false)
	require.Error(t, err)
	require.Contains(
		t,
		err.Error(),
		"requires an active transaction",
		"want fail-closed without ambient tx, got %v",
		err,
	)
}

func TestEnteredByActorGuard_SystemAdminPath_JoinsAmbientTransaction(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyEnteredBySingleColumnMigration(t, db)
	clinicB := uint64(2)
	account := &model.Account{
		Email:         "sysadmin-entered-by-guard@example.test",
		PasswordHash:  "x",
		IsActive:      true,
		IsSystemAdmin: true,
	}
	require.NoError(t, db.Create(account).Error)
	admin := &model.Staff{
		ClinicID:  1,
		AccountID: &account.ID,
		Name:      "sysadmin記録者",
		StaffType: model.StaffTypeDoctor,
		IsActive:  true,
	}
	require.NoError(t, db.Create(admin).Error)

	guard := newGormEnteredByActorGuard(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ambientTx := db.WithContext(ctx).Begin()
	require.NoError(t, ambientTx.Error)
	defer ambientTx.Rollback()

	require.NoError(t, guard.AssertEnteredByActor(
		persistence.WithTxValue(ctx, ambientTx),
		clinicB,
		admin.ID,
		true,
	))
}
