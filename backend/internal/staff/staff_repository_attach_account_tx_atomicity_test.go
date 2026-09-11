package staff_test

// staff_repository_attach_account_tx_atomicity_test.go — ambient-tx participation proofs
// for staffRepository.AttachAccountID enrolled on dbOrTxParticipatingMethods
// (lintscan TestDBOrTxInventory_MatchesAllowlist).
//
// Coverage policy (not tautology):
//  1. Writer under WithTx must roll back when a later step fails (DBOrTx participation).
//  2. Ambient-required method fails closed when TxFromContext is absent.

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	. "github.com/animal-ekarte/backend/internal/staff"
)

func TestStaffRepository_AttachAccountID_SourceContract(t *testing.T) {
	source, err := os.ReadFile("staff_repository.go")
	require.NoError(t, err)
	text := string(source)
	signature := "func (r *staffRepository) AttachAccountID("
	start := strings.Index(text, signature)
	require.NotEqual(t, -1, start, "missing method %q", signature)
	methodSource := text[start:]
	if next := strings.Index(methodSource[len(signature):], "\nfunc "); next >= 0 {
		methodSource = methodSource[:len(signature)+next]
	}

	assert.Contains(t, methodSource, "persistence.TxFromContext(ctx)")
	assert.Contains(t, methodSource, "persistence.DBOrTx(ctx, r.db)")
	assert.Contains(t, methodSource, "staffs.account_id IS NULL")
}

func TestStaffRepository_AttachAccountID_RequiresAmbientTransaction(t *testing.T) {
	repo := NewRepository(nil)
	err := repo.AttachAccountID(context.Background(), 1, 1, 1)

	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "unexpected error: %v", err)
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.Contains(t, appErr.Message, "staff account attach requires an active transaction")
}

func TestStaffRepository_AttachAccountID_ParticipatesInAmbientTxRollback(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	transactor := persistence.NewTransactor(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	staff := makeDoctor(t, db, clinicID, "attach-account-tx")
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	account := makeStaffAccount(t, db, "attach-account-tx@example.test")
	forced := errors.New("forced attach rollback")

	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		if attachErr := repo.AttachAccountID(txCtx, clinicID, staff.ID, account.ID); attachErr != nil {
			return attachErr
		}
		return forced
	})
	require.ErrorIs(t, err, forced)

	var stored model.Staff
	require.NoError(t, db.First(&stored, staff.ID).Error)
	assert.Nil(t, stored.AccountID, "AttachAccountID must participate in ambient tx and roll back")
}
