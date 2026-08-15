package staff

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type staffCredentialAuditTxMarker struct{}

type noopStaffCredentialAuditTxLogger struct{}

func (noopStaffCredentialAuditTxLogger) LogEntryTx(
	context.Context,
	CredentialAuditEntry,
) error {
	return nil
}

type staffCredentialAuditTxCapture struct {
	entry CredentialAuditEntry
	err   error
}

func (a *staffCredentialAuditTxCapture) LogEntryTx(
	ctx context.Context,
	entry CredentialAuditEntry,
) error {
	if ctx.Value(staffCredentialAuditTxMarker{}) != true {
		return errors.New("staff credential audit did not join ambient transaction")
	}
	a.entry = entry
	return a.err
}

func TestStaffService_UpdatePassword_AuditFailureRollsBackCredential(t *testing.T) {
	const (
		accountID = uint64(41)
		staffID   = uint64(29)
		clinicID  = uint64(23)
		actorID   = uint64(17)
	)
	password := "newPassw0rd"
	committedHash := "old-password-hash"
	stagedHash := ""
	tokensDeleted := false
	rolledBack := false
	auditErr := errors.New(
		"staff credential audit unavailable; redaction fixture contains sensitive-alpha, sensitive-beta, sensitive-gamma@example.test",
	)
	audit := &staffCredentialAuditTxCapture{err: auditErr}
	transactor := passwordUpdateTransactorFunc(func(
		ctx context.Context,
		fn func(context.Context) error,
	) error {
		txCtx := context.WithValue(ctx, staffCredentialAuditTxMarker{}, true)
		if txErr := fn(txCtx); txErr != nil {
			rolledBack = true
			stagedHash = ""
			tokensDeleted = false
			return txErr
		}
		committedHash = stagedHash
		return nil
	})
	repo := &coreMockStaffRepository{
		lockInClinicFn: func(
			ctx context.Context,
			gotClinicID, gotStaffID uint64,
		) (*model.Staff, error) {
			assert.Equal(t, true, ctx.Value(staffCredentialAuditTxMarker{}))
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, staffID, gotStaffID)
			return &model.Staff{
				ID:        staffID,
				ClinicID:  clinicID,
				AccountID: ptrUint64(accountID),
			}, nil
		},
		findByIDInClinicFn: func(
			context.Context,
			uint64,
			uint64,
		) (*model.Staff, error) {
			return &model.Staff{
				ID:        staffID,
				ClinicID:  clinicID,
				AccountID: ptrUint64(accountID),
			}, nil
		},
	}
	accountRepo := &coreMockAccountRepository{
		updatePasswordHashFn: func(
			ctx context.Context,
			id uint64,
			newHash string,
			_ time.Time,
		) error {
			assert.Equal(t, true, ctx.Value(staffCredentialAuditTxMarker{}))
			assert.Equal(t, accountID, id)
			stagedHash = newHash
			return nil
		},
		deletePasswordResetTokensFn: func(
			ctx context.Context,
			id uint64,
		) error {
			assert.Equal(t, true, ctx.Value(staffCredentialAuditTxMarker{}))
			assert.Equal(t, accountID, id)
			tokensDeleted = true
			return nil
		},
	}
	assignments := &coreMockStaffClinicAssignmentRepository{
		lockActiveFn: func(
			context.Context,
			uint64,
		) ([]model.StaffClinicAssignment, error) {
			return []model.StaffClinicAssignment{
				{StaffID: staffID, ClinicID: clinicID, IsMain: true},
			}, nil
		},
	}
	service := NewStaffServiceWithCredentialAudit(
		repo,
		accountRepo,
		assignments,
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		nil,
		nil,
		nil,
		nil,
		transactor,
		audit,
	)
	var logs bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	result, updateErr := service.Update(
		context.Background(),
		clinicID,
		staffID,
		&UpdateStaffInput{
			Password:            &password,
			AuthorizedClinicIDs: []uint64{clinicID},
			CredentialAudit: &CredentialMutationAudit{
				ClinicID:      clinicID,
				ActorStaffID:  actorID,
				TargetStaffID: staffID,
				IPAddress:     "192.0.2.29",
				UserAgent:     "staff-credential-audit-test",
			},
		},
	)

	require.Error(t, updateErr)
	assert.ErrorIs(t, updateErr, auditErr)
	assert.Nil(t, result)
	assert.True(t, rolledBack)
	assert.False(t, tokensDeleted)
	assert.Empty(t, stagedHash)
	assert.Equal(t, "old-password-hash", committedHash)
	require.NotNil(t, audit.entry.ResourceID)
	assert.Equal(t, accountID, *audit.entry.ResourceID)
	assert.Equal(t, map[string]any{"staff_id": staffID}, audit.entry.NewValue)
	for _, secret := range []string{
		"sensitive-alpha",
		"sensitive-beta",
		"sensitive-gamma@example.test",
		"newPassw0rd",
	} {
		assert.NotContains(t, logs.String(), secret)
	}
}

func TestStaffService_UpdatePasswordFailsClosedWithoutAuditBeforeTransaction(
	t *testing.T,
) {
	password := "newPassw0rd"
	tx := &coreFakeTransactor{}
	service := NewStaffService(
		&coreMockStaffRepository{},
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		nil,
		nil,
		nil,
		nil,
		tx,
	)

	result, updateErr := service.Update(
		context.Background(),
		23,
		29,
		&UpdateStaffInput{
			Password:            &password,
			AuthorizedClinicIDs: []uint64{23},
			CredentialAudit:     testStaffCredentialAudit(23, 29),
		},
	)

	require.Error(t, updateErr)
	var appErr *apperrors.AppError
	require.ErrorAs(t, updateErr, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.Nil(t, result)
	assert.Zero(t, tx.calls)
}

func ptrUint64(value uint64) *uint64 {
	return &value
}
