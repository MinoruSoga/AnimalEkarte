package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestPermissionPolicyLock_RequiresTransactionAndClinic(t *testing.T) {
	repo := NewPermissionGroupRepository(nil)
	var appErr *apperrors.AppError
	require.ErrorAs(t, repo.LockPermissionPolicy(context.Background(), 1), &appErr)
	require.Equal(t, "INTERNAL", appErr.Code)
	require.Error(t, repo.LockPermissionPolicy(context.Background(), 0))
}

func TestPermissionPolicyLock_ClinicKeyAndFailureBeforeGroupRead(t *testing.T) {
	// The callback captures SQL without a database connection; lock concurrency
	// itself is covered by TestPermissionPolicyDB_ConcurrentGroupDeactivation.
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=unused dbname=unused"}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	lockFailure := errors.New("lock unavailable")
	var keys []string
	require.NoError(t, db.Callback().Raw().Replace("gorm:raw", func(tx *gorm.DB) {
		require.Equal(t, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", tx.Statement.SQL.String())
		require.Len(t, tx.Statement.Vars, 1)
		key, ok := tx.Statement.Vars[0].(string)
		require.True(t, ok)
		keys = append(keys, key)
		tx.AddError(lockFailure)
	}))
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("reject-read", func(*gorm.DB) {
		t.Fatal("group read must not precede a successful policy lock")
	}))
	repo := NewPermissionGroupRepository(db)
	ctx := persistence.WithTxValue(context.Background(), db)
	for _, clinicID := range []uint64{1, 2} {
		group, lockErr := repo.LockByIDForUpdate(ctx, clinicID, 3)
		require.Nil(t, group)
		require.ErrorIs(t, lockErr, lockFailure)
	}
	require.Equal(t, []string{"permission-policy:clinic:1", "permission-policy:clinic:2"}, keys)
}
