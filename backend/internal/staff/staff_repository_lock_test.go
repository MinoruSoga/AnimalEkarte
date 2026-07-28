package staff_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	. "github.com/animal-ekarte/backend/internal/staff"
)

func TestStaffRepository_ActiveRowLocks_SourceContract(t *testing.T) {
	source, err := os.ReadFile("staff_repository.go")
	require.NoError(t, err)
	text := string(source)

	for _, signature := range []string{
		"func (r *staffRepository) LockActiveByIDForUpdate(",
		"func (r *staffRepository) LockActiveByIDForUpdateInClinic(",
		"func (r *staffRepository) LockActiveByIDForShare(",
	} {
		start := strings.Index(text, signature)
		require.NotEqual(t, -1, start, "missing method %q", signature)
		methodSource := text[start:]
		if next := strings.Index(methodSource[len(signature):], "\nfunc "); next >= 0 {
			methodSource = methodSource[:len(signature)+next]
		}

		assert.Contains(t, methodSource, "persistence.TxFromContext(ctx)")
		assert.Contains(t, methodSource, "persistence.DBOrTx(ctx, r.db)")
		assert.Contains(t, methodSource, "deleted_at IS NULL")
		assert.Contains(t, methodSource, "apperrors.FromGORM")
	}

	updateStart := strings.Index(text, "func (r *staffRepository) LockActiveByIDForUpdate(")
	require.NotEqual(t, -1, updateStart)
	updateSource := text[updateStart:]
	assert.Contains(t, updateSource, `Strength: "UPDATE"`)

	shareStart := strings.Index(text, "func (r *staffRepository) LockActiveByIDForShare(")
	require.NotEqual(t, -1, shareStart)
	shareSource := text[shareStart:]
	assert.Contains(t, shareSource, `Strength: "SHARE"`)

	scopedStart := strings.Index(text, "func (r *staffRepository) LockActiveByIDForUpdateInClinic(")
	require.NotEqual(t, -1, scopedStart)
	scopedSource := text[scopedStart:shareStart]
	assert.Contains(t, scopedSource, "staff_clinic_assignments.clinic_id = ?")
	assert.Contains(t, scopedSource, "staff_clinic_assignments.deleted_at IS NULL")
}

func TestStaffRepository_ActiveRowLocks_RequireAmbientTransaction(t *testing.T) {
	repo := NewStaffRepository(nil)

	for _, lock := range []struct {
		name string
		fn   func(context.Context, uint64) error
	}{
		{
			name: "update",
			fn: func(ctx context.Context, id uint64) error {
				_, err := repo.LockActiveByIDForUpdate(ctx, id)
				return err
			},
		},
		{
			name: "share",
			fn: func(ctx context.Context, id uint64) error {
				_, err := repo.LockActiveByIDForShare(ctx, id)
				return err
			},
		},
		{
			name: "clinic-scoped update",
			fn: func(ctx context.Context, id uint64) error {
				_, err := repo.LockActiveByIDForUpdateInClinic(ctx, 9, id)
				return err
			},
		},
	} {
		t.Run(lock.name, func(t *testing.T) {
			err := lock.fn(context.Background(), 1)

			require.Error(t, err)
			var appErr *apperrors.AppError
			require.True(t, errors.As(err, &appErr), "unexpected error: %v", err)
			assert.Equal(t, "INTERNAL", appErr.Code)
		})
	}
}

func TestStaffRepository_ActiveRowLocks_DatabaseContract(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	transactor := persistence.NewTransactor(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)
	staff := makeDoctor(t, db, clinicA, "staff active lock contract")
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)

	for _, lock := range []struct {
		name string
		fn   func(context.Context) (*model.Staff, error)
	}{
		{
			name: "update",
			fn: func(txCtx context.Context) (*model.Staff, error) {
				return repo.LockActiveByIDForUpdate(txCtx, staff.ID)
			},
		},
		{
			name: "share",
			fn: func(txCtx context.Context) (*model.Staff, error) {
				return repo.LockActiveByIDForShare(txCtx, staff.ID)
			},
		},
		{
			name: "clinic scoped update",
			fn: func(txCtx context.Context) (*model.Staff, error) {
				return repo.LockActiveByIDForUpdateInClinic(txCtx, clinicA, staff.ID)
			},
		},
	} {
		t.Run(lock.name+" returns active staff", func(t *testing.T) {
			var locked *model.Staff
			err := transactor.WithTx(ctx, func(txCtx context.Context) error {
				var lockErr error
				locked, lockErr = lock.fn(txCtx)
				return lockErr
			})

			require.NoError(t, err)
			require.NotNil(t, locked)
			assert.Equal(t, staff.ID, locked.ID)
		})
	}

	t.Run("clinic scoped update rejects another clinic", func(t *testing.T) {
		err := transactor.WithTx(ctx, func(txCtx context.Context) error {
			_, lockErr := repo.LockActiveByIDForUpdateInClinic(txCtx, clinicB, staff.ID)
			return lockErr
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	require.NoError(t, db.Delete(&model.Staff{}, staff.ID).Error)
	for _, lock := range []struct {
		name string
		fn   func(context.Context) (*model.Staff, error)
	}{
		{
			name: "update",
			fn: func(txCtx context.Context) (*model.Staff, error) {
				return repo.LockActiveByIDForUpdate(txCtx, staff.ID)
			},
		},
		{
			name: "share",
			fn: func(txCtx context.Context) (*model.Staff, error) {
				return repo.LockActiveByIDForShare(txCtx, staff.ID)
			},
		},
		{
			name: "clinic scoped update",
			fn: func(txCtx context.Context) (*model.Staff, error) {
				return repo.LockActiveByIDForUpdateInClinic(txCtx, clinicA, staff.ID)
			},
		},
	} {
		t.Run(lock.name+" rejects soft deleted staff", func(t *testing.T) {
			err := transactor.WithTx(ctx, func(txCtx context.Context) error {
				_, lockErr := lock.fn(txCtx)
				return lockErr
			})

			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
		})
	}
}

func TestStaffRepository_ActiveRowLocks_HoldUntilTransactionEndsDatabase(t *testing.T) {
	tests := []struct {
		name string
		lock func(StaffRepository, context.Context, uint64, uint64) (*model.Staff, error)
	}{
		{
			name: "update",
			lock: func(repo StaffRepository, ctx context.Context, _, staffID uint64) (*model.Staff, error) {
				return repo.LockActiveByIDForUpdate(ctx, staffID)
			},
		},
		{
			name: "share",
			lock: func(repo StaffRepository, ctx context.Context, _, staffID uint64) (*model.Staff, error) {
				return repo.LockActiveByIDForShare(ctx, staffID)
			},
		},
		{
			name: "clinic scoped update",
			lock: func(repo StaffRepository, ctx context.Context, clinicID, staffID uint64) (*model.Staff, error) {
				return repo.LockActiveByIDForUpdateInClinic(ctx, clinicID, staffID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupStaffRepositoryTestDB(t)
			repo := NewStaffRepository(db)
			transactor := persistence.NewTransactor(db)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			const clinicID = uint64(1)
			seedClinicsForFK(t, db, clinicID)
			staff := makeDoctor(t, db, clinicID, "staff lock lifetime "+tt.name)
			makeStaffClinicAssignment(t, db, staff.ID, clinicID)

			locked := make(chan struct{})
			release := make(chan struct{})
			holderDone := make(chan error, 1)
			go func() {
				holderDone <- transactor.WithTx(ctx, func(txCtx context.Context) error {
					if _, err := tt.lock(repo, txCtx, clinicID, staff.ID); err != nil {
						return err
					}
					close(locked)
					select {
					case <-release:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				})
			}()

			select {
			case <-locked:
			case <-ctx.Done():
				t.Fatal("staff lock was not acquired")
			}
			deleteDone := make(chan error, 1)
			go func() {
				deleteDone <- db.WithContext(ctx).Delete(&model.Staff{}, staff.ID).Error
			}()

			var deleteErr error
			completedBeforeRelease := false
			select {
			case deleteErr = <-deleteDone:
				completedBeforeRelease = true
			case <-time.After(100 * time.Millisecond):
			}
			close(release)
			require.NoError(t, <-holderDone)
			if !completedBeforeRelease {
				select {
				case deleteErr = <-deleteDone:
				case <-ctx.Done():
					t.Fatal("staff delete did not resume after lock transaction ended")
				}
			}
			require.NoError(t, deleteErr)
			assert.False(t, completedBeforeRelease, "staff delete must wait for the row lock")
		})
	}
}
