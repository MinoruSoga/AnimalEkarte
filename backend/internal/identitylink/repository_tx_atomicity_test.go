package identitylink

// repository_tx_atomicity_test.go — ambient-tx participation proofs for
// repository.conn (DBOrTx) and repository.requireAmbientTx (TxFromContext-only)
// enrolled on lintscan dbOrTxParticipatingMethods.

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func requireIdentitylinkInternal(t *testing.T, err error, msgSubstr string) {
	t.Helper()
	require.Error(t, err)
	var app *apperrors.AppError
	require.True(t, errors.As(err, &app), "want *apperrors.AppError, got %T: %v", err, err)
	assert.Equal(t, "INTERNAL", app.Code)
	assert.Contains(t, app.Message, msgSubstr)
}

// requireAmbientTx: CreateOwnerGroup without ambient tx must fail closed.
func TestRepository_RequireAmbientTx_FailClosedWithoutTx(t *testing.T) {
	db := setupIdentityLinkTestDB(t)
	repo := NewRepository(db)

	err := repo.CreateOwnerGroup(context.Background(), &model.OwnerIdentityGroup{
		CreatedClinicID: 1,
		Version:         1,
	})
	requireIdentitylinkInternal(t, err, "identitylink write requires ambient transaction")

	var n int64
	require.NoError(t, db.Model(&model.OwnerIdentityGroup{}).Count(&n).Error)
	assert.Zero(t, n)
}

// requireAmbientTx: CreateOwnerGroup + CreateOwnerMembers under ambient tx roll back
// when a later step fails (covers requireAmbientTx participation).
func TestRepository_RequireAmbientTx_WriteRollbackOnFailure(t *testing.T) {
	db := setupIdentityLinkTestDB(t)
	repo := NewRepository(db)
	tx := persistence.NewTransactor(db)
	forced := errors.New("forced identitylink write rollback")

	o1 := seedClinicOwner(t, db, 1, "TxOwnerA")
	o2 := seedClinicOwner(t, db, 2, "TxOwnerB")

	err := tx.WithTx(context.Background(), func(txCtx context.Context) error {
		group := &model.OwnerIdentityGroup{CreatedClinicID: 1, Version: 1}
		if err := repo.CreateOwnerGroup(txCtx, group); err != nil {
			return err
		}
		if err := repo.CreateOwnerMembers(txCtx, []model.OwnerIdentityGroupMember{
			{GroupCreatedClinicID: 1, GroupID: group.ID, ClinicID: 1, OwnerID: o1.ID},
			{GroupCreatedClinicID: 1, GroupID: group.ID, ClinicID: 2, OwnerID: o2.ID},
		}); err != nil {
			return err
		}
		return forced
	})
	require.ErrorIs(t, err, forced)

	var groups int64
	require.NoError(t, db.Model(&model.OwnerIdentityGroup{}).Count(&groups).Error)
	assert.Zero(t, groups, "CreateOwnerGroup must participate via requireAmbientTx and roll back")

	var members int64
	require.NoError(t, db.Model(&model.OwnerIdentityGroupMember{}).Count(&members).Error)
	assert.Zero(t, members, "CreateOwnerMembers must participate via requireAmbientTx and roll back")
}

// conn: FindActiveOwnerMembership under ambient tx must observe uncommitted membership
// written in the same ambient tx (DBOrTx participation), then roll back.
func TestRepository_Conn_ObservesUncommittedAmbientMembership(t *testing.T) {
	db := setupIdentityLinkTestDB(t)
	repo := NewRepository(db)
	tx := persistence.NewTransactor(db)
	forced := errors.New("forced conn probe rollback")

	o1 := seedClinicOwner(t, db, 1, "ConnOwnerA")
	o2 := seedClinicOwner(t, db, 2, "ConnOwnerB")

	err := tx.WithTx(context.Background(), func(txCtx context.Context) error {
		if persistence.TxFromContext(txCtx) == nil {
			return errors.New("transactor did not install ambient tx")
		}
		group := &model.OwnerIdentityGroup{CreatedClinicID: 1, Version: 1}
		if err := repo.CreateOwnerGroup(txCtx, group); err != nil {
			return err
		}
		if err := repo.CreateOwnerMembers(txCtx, []model.OwnerIdentityGroupMember{
			{GroupCreatedClinicID: 1, GroupID: group.ID, ClinicID: 1, OwnerID: o1.ID},
			{GroupCreatedClinicID: 1, GroupID: group.ID, ClinicID: 2, OwnerID: o2.ID},
		}); err != nil {
			return err
		}
		// FindActiveOwnerMembership uses conn → DBOrTx; must see uncommitted member.
		mem, err := repo.FindActiveOwnerMembership(txCtx, 1, o1.ID)
		if err != nil {
			return err
		}
		if mem == nil || mem.GroupID != group.ID {
			return errors.New("conn-backed FindActiveOwnerMembership did not see ambient membership")
		}
		return forced
	})
	require.ErrorIs(t, err, forced)

	// Outside ambient: membership must not exist (rolled back).
	mem, err := repo.FindActiveOwnerMembership(context.Background(), 1, o1.ID)
	require.NoError(t, err)
	assert.Nil(t, mem)

	var groups int64
	require.NoError(t, db.Model(&model.OwnerIdentityGroup{}).Count(&groups).Error)
	assert.Zero(t, groups)
}
