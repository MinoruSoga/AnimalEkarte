package auth

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

func TestAccountRepository_AmbientTransactionReadYourWritesAndRollback(t *testing.T) {
	db := setupAccountTestDB(t)
	repo := NewAccountRepository(db)
	ctx := context.Background()
	forcedErr := errors.New("force account transaction rollback")

	account := &model.Account{
		Email:        "ambient-account@example.test",
		PasswordHash: "initial-hash",
		IsActive:     true,
	}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		require.NoError(t, repo.Create(txCtx, account))

		byID, findErr := repo.FindByID(txCtx, account.ID)
		require.NoError(t, findErr)
		assert.Equal(t, account.Email, byID.Email)

		byEmail, findErr := repo.FindByEmail(txCtx, account.Email)
		require.NoError(t, findErr)
		assert.Equal(t, account.ID, byEmail.ID)

		require.NoError(t, repo.UpdatePasswordHash(
			txCtx,
			account.ID,
			"updated-hash",
			time.Now(),
		))
		updated, findErr := repo.FindByID(txCtx, account.ID)
		require.NoError(t, findErr)
		assert.Equal(t, "updated-hash", updated.PasswordHash)
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	got, err := repo.FindByID(ctx, account.ID)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNotFound(err), "ambient transaction rollback must remove the account")
}

func TestPasswordResetTokenRepository_AmbientTransactionReadYourWritesAndRollback(t *testing.T) {
	db := setupPasswordResetTokenTestDB(t)
	repo := NewPasswordResetTokenRepository(db)
	ctx := context.Background()

	t.Run("create and find participate in rollback", func(t *testing.T) {
		forcedErr := errors.New("force password reset create rollback")
		token := &model.PasswordResetToken{
			AccountID: 1,
			TokenHash: "ambient-create-hash",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			require.NoError(t, repo.Create(txCtx, token))
			got, findErr := repo.FindByTokenHash(txCtx, token.TokenHash)
			require.NoError(t, findErr)
			assert.Equal(t, token.ID, got.ID)
			return forcedErr
		})
		require.ErrorIs(t, err, forcedErr)

		got, findErr := repo.FindByTokenHash(ctx, token.TokenHash)
		require.Error(t, findErr)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(findErr))
	})

	t.Run("delete methods participate in rollback", func(t *testing.T) {
		const targetAccountID = uint64(10)
		byID := makePasswordResetToken(t, repo, targetAccountID, "ambient-delete-id", time.Now().Add(time.Hour))
		byAccount := makePasswordResetToken(t, repo, targetAccountID, "ambient-delete-account", time.Now().Add(time.Hour))
		other := makePasswordResetToken(t, repo, 11, "ambient-delete-other", time.Now().Add(time.Hour))
		forcedErr := errors.New("force password reset delete rollback")

		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			require.NoError(t, repo.DeleteByID(txCtx, byID.ID))
			require.NoError(t, repo.DeleteByAccountID(txCtx, targetAccountID))

			for _, hash := range []string{byID.TokenHash, byAccount.TokenHash} {
				got, findErr := repo.FindByTokenHash(txCtx, hash)
				require.Error(t, findErr)
				assert.Nil(t, got)
				assert.True(t, apperrors.IsNotFound(findErr))
			}
			got, findErr := repo.FindByTokenHash(txCtx, other.TokenHash)
			require.NoError(t, findErr)
			assert.Equal(t, other.ID, got.ID)
			return forcedErr
		})
		require.ErrorIs(t, err, forcedErr)

		for _, token := range []*model.PasswordResetToken{byID, byAccount, other} {
			got, findErr := repo.FindByTokenHash(ctx, token.TokenHash)
			require.NoError(t, findErr)
			assert.Equal(t, token.ID, got.ID)
		}
	})
}

func TestTokenBlacklistRepository_AmbientTransactionReadYourWritesAndRollback(t *testing.T) {
	db := setupTokenBlacklistTestDB(t)
	repo := NewTokenBlacklistRepository(db)
	ctx := context.Background()

	t.Run("create and exists participate in rollback", func(t *testing.T) {
		forcedErr := errors.New("force blacklist create rollback")
		entry := &model.TokenBlacklist{
			JTI:       "ambient-blacklist-create",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			require.NoError(t, repo.Create(txCtx, entry))
			exists, existsErr := repo.ExistsByJTI(txCtx, entry.JTI)
			require.NoError(t, existsErr)
			assert.True(t, exists)
			return forcedErr
		})
		require.ErrorIs(t, err, forcedErr)

		exists, existsErr := repo.ExistsByJTI(ctx, entry.JTI)
		require.NoError(t, existsErr)
		assert.False(t, exists)
	})

	t.Run("delete expired participates in rollback", func(t *testing.T) {
		expired := &model.TokenBlacklist{
			JTI:       "ambient-blacklist-expired",
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		require.NoError(t, repo.Create(ctx, expired))
		forcedErr := errors.New("force blacklist delete rollback")

		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			require.NoError(t, repo.DeleteExpired(txCtx))

			var count int64
			require.NoError(t, tx.WithContext(txCtx).
				Model(&model.TokenBlacklist{}).
				Where("jti = ?", expired.JTI).
				Count(&count).Error)
			assert.Zero(t, count)
			return forcedErr
		})
		require.ErrorIs(t, err, forcedErr)

		var count int64
		require.NoError(t, db.WithContext(ctx).
			Model(&model.TokenBlacklist{}).
			Where("jti = ?", expired.JTI).
			Count(&count).Error)
		assert.Equal(t, int64(1), count, "rollback must restore the expired blacklist entry")
	})
}

func TestPermissionGroupRepository_AmbientTransactionReadYourWritesAndRollback(t *testing.T) {
	ctx := context.Background()

	t.Run("rule replacement participates in ambient rollback", func(t *testing.T) {
		db := setupPermissionGroupRepositoryTestDB(t)
		repo := NewPermissionGroupRepository(db)
		const clinicID = uint64(1)
		group := makePermissionGroup(t, db, clinicID, "ambient permission rule group")
		original := makeEffPermRule(t, db, group.ID, "medical_record", true, false, false, false)
		forcedErr := errors.New("force permission rule rollback")

		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			require.NoError(t, repo.UpdateRules(txCtx, clinicID, group.ID, []model.PermissionGroupRule{
				{Resource: "billing", CanView: true},
			}))

			got, findErr := repo.FindByID(txCtx, clinicID, group.ID)
			require.NoError(t, findErr)
			require.Len(t, got.Rules, 1)
			assert.Equal(t, "billing", got.Rules[0].Resource)
			return forcedErr
		})
		require.ErrorIs(t, err, forcedErr)

		got, findErr := repo.FindByID(ctx, clinicID, group.ID)
		require.NoError(t, findErr)
		require.Len(t, got.Rules, 1)
		assert.Equal(t, original.ID, got.Rules[0].ID)
		assert.Equal(t, "medical_record", got.Rules[0].Resource)
	})

	t.Run("staff group replacement participates in ambient rollback", func(t *testing.T) {
		db := setupPermissionGroupStaffIsolationTestDB(t)
		repo := NewPermissionGroupRepository(db)
		clinic := makePermissionGroupTestClinic(t, db, "ambient staff permission clinic")
		staff := makeDoctorAssignedToClinic(t, db, clinic.ID, "ambient staff permission staff")
		group := makePermissionGroup(t, db, clinic.ID, "ambient staff permission group")
		forcedErr := errors.New("force staff permission rollback")

		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			require.NoError(t, repo.UpdateStaffGroups(txCtx, clinic.ID, staff.ID, []uint64{group.ID}))

			groupIDs, findErr := repo.FindAllGroupIDsByStaffID(txCtx, clinic.ID, staff.ID)
			require.NoError(t, findErr)
			assert.Equal(t, []uint64{group.ID}, groupIDs)
			return forcedErr
		})
		require.ErrorIs(t, err, forcedErr)

		groupIDs, findErr := repo.FindAllGroupIDsByStaffID(ctx, clinic.ID, staff.ID)
		require.NoError(t, findErr)
		assert.Empty(t, groupIDs)
	})
}

func TestPermissionGroupRepository_CreateWithRules_RollsBackWithAmbientTransaction(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db).(*permissionGroupRepository)
	ctx := context.Background()
	const clinicID = uint64(1)
	group := &model.PermissionGroup{ClinicID: clinicID, Name: "ambient create-with-rules"}
	rules := []model.PermissionGroupRule{{Resource: "medical_record", CanView: true}}
	forcedErr := errors.New("force create-with-rules rollback")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		created, createErr := repo.CreateWithRules(txCtx, group, rules)
		require.NoError(t, createErr)
		require.Len(t, created.Rules, 1)
		assert.Equal(t, "medical_record", created.Rules[0].Resource)
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	var groupCount, ruleCount int64
	require.NoError(t, db.Model(&model.PermissionGroup{}).Where("id = ?", group.ID).Count(&groupCount).Error)
	require.NoError(t, db.Model(&model.PermissionGroupRule{}).Where("group_id = ?", group.ID).Count(&ruleCount).Error)
	assert.Zero(t, groupCount)
	assert.Zero(t, ruleCount)
}

func TestPermissionGroupRepository_UpdateWithRules_RollsBackWithAmbientTransaction(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db).(*permissionGroupRepository)
	ctx := context.Background()
	const clinicID = uint64(1)
	group := makePermissionGroup(t, db, clinicID, "ambient update original")
	originalRule := makeEffPermRule(t, db, group.ID, "medical_record", true, false, false, false)
	forcedErr := errors.New("force update-with-rules rollback")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		updated, updateErr := repo.UpdateWithRules(
			txCtx,
			clinicID,
			group.ID,
			map[string]any{"name": "ambient update replacement"},
			[]model.PermissionGroupRule{{Resource: "billing", CanView: true}},
		)
		require.NoError(t, updateErr)
		assert.Equal(t, "ambient update replacement", updated.Name)
		require.Len(t, updated.Rules, 1)
		assert.Equal(t, "billing", updated.Rules[0].Resource)
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	got, findErr := repo.FindByID(ctx, clinicID, group.ID)
	require.NoError(t, findErr)
	assert.Equal(t, "ambient update original", got.Name)
	require.Len(t, got.Rules, 1)
	assert.Equal(t, originalRule.ID, got.Rules[0].ID)
	assert.Equal(t, "medical_record", got.Rules[0].Resource)
}

func TestPermissionGroupRepository_replaceRules_RollsBackWithAmbientTransaction(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := &permissionGroupRepository{db: db}
	ctx := context.Background()
	const clinicID = uint64(1)
	group := makePermissionGroup(t, db, clinicID, "ambient private replace")
	originalRule := makeEffPermRule(t, db, group.ID, "medical_record", true, false, false, false)
	forcedErr := errors.New("force private replace-rules rollback")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		require.NoError(t, repo.replaceRules(txCtx, group.ID, []model.PermissionGroupRule{
			{Resource: "billing", CanView: true},
		}))
		got, findErr := repo.FindByID(txCtx, clinicID, group.ID)
		require.NoError(t, findErr)
		require.Len(t, got.Rules, 1)
		assert.Equal(t, "billing", got.Rules[0].Resource)
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	got, findErr := repo.FindByID(ctx, clinicID, group.ID)
	require.NoError(t, findErr)
	require.Len(t, got.Rules, 1)
	assert.Equal(t, originalRule.ID, got.Rules[0].ID)
	assert.Equal(t, "medical_record", got.Rules[0].Resource)
}
