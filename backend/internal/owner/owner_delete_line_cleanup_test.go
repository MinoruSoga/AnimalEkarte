package owner_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lstepdomain "github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupOwnerDeleteLineCleanupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupOwnerDeletePetOwnerGuardTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS lstep_next_visit_sync_retries (
		clinic_id BIGINT NOT NULL,
		owner_id BIGINT NOT NULL,
		revision BIGSERIAL NOT NULL,
		cleanup_line_user_id VARCHAR(64),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (clinic_id, owner_id)
	)`).Error)
	require.NoError(t, db.Exec(`ALTER TABLE lstep_next_visit_sync_retries
		ADD COLUMN IF NOT EXISTS cleanup_line_user_id VARCHAR(64)`).Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE lstep_next_visit_sync_retries").Error)
	testdb.SeedClinicsForFK(t, db, 1, 2)
	require.NoError(t, db.Exec(`DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_owners_clinic_id_id') THEN
		ALTER TABLE owners ADD CONSTRAINT uq_owners_clinic_id_id UNIQUE (clinic_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'lstep_next_visit_sync_retries_clinic_id_fkey') THEN
		ALTER TABLE lstep_next_visit_sync_retries
			ADD CONSTRAINT lstep_next_visit_sync_retries_clinic_id_fkey
			FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_lstep_next_visit_sync_retries_owner_clinic') THEN
		ALTER TABLE lstep_next_visit_sync_retries
			ADD CONSTRAINT fk_lstep_next_visit_sync_retries_owner_clinic
			FOREIGN KEY (clinic_id, owner_id) REFERENCES owners(clinic_id, id) ON DELETE RESTRICT;
	END IF;
END $$`).Error)
	return db
}

func makeLineLinkedOwner(t *testing.T, db *gorm.DB, clinicID uint64, name, lineUserID string) *model.Owner {
	t.Helper()
	owner := testdb.MakeTestOwner(t, db, clinicID, name)
	require.NoError(t, db.Exec("UPDATE owners SET line_user_id = ? WHERE id = ?", lineUserID, owner.ID).Error)
	owner.LineUserID = &lineUserID
	return owner
}

func countCleanupIntentRows(t *testing.T, db *gorm.DB, clinicID, ownerID uint64) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM lstep_next_visit_sync_retries
		WHERE clinic_id = ? AND owner_id = ? AND cleanup_line_user_id IS NOT NULL`, clinicID, ownerID).Scan(&count).Error)
	return count
}

type ownerDeleteCleanupRecorder struct {
	enqueueNextVisitCleanupFn func(context.Context, uint64, uint64, string) error
	syncNextVisitTagFn        func(context.Context, uint64, uint64) error
	enqueuedRecipient         string
	enqueuedInTx              bool
	syncCalls                 int
	syncCtxCanceled           bool
	syncHadAmbientTx          bool
	syncDeadline              time.Time
	syncHasDeadline           bool
}

func (r *ownerDeleteCleanupRecorder) SyncOwnerAnimalClassificationTags(context.Context, uint64, uint64) error {
	return nil
}
func (r *ownerDeleteCleanupRecorder) SyncExclusionTags(context.Context, uint64, uint64) error {
	return nil
}
func (r *ownerDeleteCleanupRecorder) EnqueueNextVisitCleanup(ctx context.Context, clinicID, ownerID uint64, lineUserID string) error {
	r.enqueuedRecipient = lineUserID
	r.enqueuedInTx = persistence.TxFromContext(ctx) != nil
	if r.enqueueNextVisitCleanupFn != nil {
		return r.enqueueNextVisitCleanupFn(ctx, clinicID, ownerID, lineUserID)
	}
	return nil
}
func (r *ownerDeleteCleanupRecorder) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	r.syncCalls++
	r.syncCtxCanceled = ctx.Err() != nil
	r.syncHadAmbientTx = persistence.TxFromContext(ctx) != nil
	r.syncDeadline, r.syncHasDeadline = ctx.Deadline()
	if r.syncNextVisitTagFn != nil {
		return r.syncNextVisitTagFn(ctx, clinicID, ownerID)
	}
	return nil
}

type plainTagNotifier struct{}

func (plainTagNotifier) SyncOwnerAnimalClassificationTags(context.Context, uint64, uint64) error {
	return nil
}
func (plainTagNotifier) SyncExclusionTags(context.Context, uint64, uint64) error { return nil }

func realEnqueueNextVisitCleanup(t *testing.T, db *gorm.DB) func(context.Context, uint64, uint64, string) error {
	t.Helper()
	queue, ok := lstepdomain.NewLstepTagCacheRepository(db).(interface {
		EnqueueNextVisitCleanup(context.Context, uint64, uint64, string) error
	})
	require.True(t, ok, "the production tag cache repository must expose the cleanup queue")
	return queue.EnqueueNextVisitCleanup
}

type failingDeleteRepository struct {
	ownerdomain.Repository
	err error
}

func (r failingDeleteRepository) Delete(context.Context, uint64, uint64) error { return r.err }

type completionFailureTransactor struct {
	db  *gorm.DB
	err error
}

func (tr completionFailureTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx := tr.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := fn(persistence.WithTxValue(ctx, tx)); err != nil {
		_ = tx.Rollback().Error
		return err
	}
	if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
		return rollbackErr
	}
	return tr.err
}

type cancelAfterCommitTransactor struct {
	inner  ownerdomain.OwnerDeleteTransactor
	cancel context.CancelFunc
}

func (tr cancelAfterCommitTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	err := tr.inner.WithTx(ctx, fn)
	if err == nil {
		tr.cancel()
	}
	return err
}

func TestOwnerService_Delete_LineLinkedOwnerPersistsCleanupIntent(t *testing.T) {
	db := setupOwnerDeleteLineCleanupTestDB(t)
	repo := newOwnerRepository(db)
	petOwnerRepo := petdomain.NewPetOwnerRepository(db)
	const clinicID = uint64(1)
	target := makeLineLinkedOwner(t, db, clinicID, "linked delete target", "Ucleanup-target")

	recorder := &ownerDeleteCleanupRecorder{
		enqueueNextVisitCleanupFn: realEnqueueNextVisitCleanup(t, db),
	}
	svc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
		repo, nil, recorder, nil, repo, petOwnerRepo, persistence.NewTransactor(db))

	require.NoError(t, svc.Delete(context.Background(), clinicID, target.ID))
	assert.Equal(t, "Ucleanup-target", recorder.enqueuedRecipient)
	assert.True(t, recorder.enqueuedInTx, "cleanup intent must enqueue inside the deletion transaction")
	assert.Equal(t, int64(1), countCleanupIntentRows(t, db, clinicID, target.ID),
		"the immutable recipient snapshot must be durable after the delete commit")
	assert.Equal(t, 1, recorder.syncCalls, "bounded post-commit cleanup runs once")
	assert.False(t, recorder.syncCtxCanceled, "post-commit cleanup uses a live detached context")
	assert.False(t, recorder.syncHadAmbientTx, "post-commit cleanup must not borrow the committed transaction")
	require.True(t, recorder.syncHasDeadline, "post-commit cleanup context must carry a bounded deadline")
	assert.LessOrEqual(t, time.Until(recorder.syncDeadline), 30*time.Second,
		"the owner boundary supplies at most a 30 second cleanup window")

	var deleted model.Owner
	require.NoError(t, db.Unscoped().Where("id = ?", target.ID).First(&deleted).Error)
	assert.True(t, deleted.DeletedAt.Valid)
}

func TestOwnerService_Delete_CleanupDependencyFailuresRollBack(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(t *testing.T, db *gorm.DB, repo ownerdomain.Repository, counter ownerdomain.PetOwnerCounter, target *model.Owner) error
	}{
		{"enqueue failure rolls back deletion", func(t *testing.T, db *gorm.DB, repo ownerdomain.Repository, counter ownerdomain.PetOwnerCounter, target *model.Owner) error {
			failure := errors.New("cleanup enqueue failed")
			recorder := &ownerDeleteCleanupRecorder{
				enqueueNextVisitCleanupFn: func(context.Context, uint64, uint64, string) error { return failure },
			}
			svc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
				repo, nil, recorder, nil, repo, counter, persistence.NewTransactor(db))
			err := svc.Delete(context.Background(), 1, target.ID)
			require.ErrorIs(t, err, failure)
			return err
		}},
		{"missing cleanup dependency fails closed", func(t *testing.T, db *gorm.DB, repo ownerdomain.Repository, counter ownerdomain.PetOwnerCounter, target *model.Owner) error {
			svc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
				repo, nil, plainTagNotifier{}, nil, repo, counter, persistence.NewTransactor(db))
			err := svc.Delete(context.Background(), 1, target.ID)
			require.Error(t, err)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, "INTERNAL", appErr.Code)
			return err
		}},
		{"owner delete failure rolls back the intent", func(t *testing.T, db *gorm.DB, repo ownerdomain.Repository, counter ownerdomain.PetOwnerCounter, target *model.Owner) error {
			deleteErr := errors.New("owner delete failed")
			failingRepo := failingDeleteRepository{Repository: repo, err: deleteErr}
			recorder := &ownerDeleteCleanupRecorder{
				enqueueNextVisitCleanupFn: realEnqueueNextVisitCleanup(t, db),
			}
			svc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
				failingRepo, nil, recorder, nil, repo, counter, persistence.NewTransactor(db))
			err := svc.Delete(context.Background(), 1, target.ID)
			require.ErrorIs(t, err, deleteErr)
			return err
		}},
		{"transactor completion failure rolls back the intent", func(t *testing.T, db *gorm.DB, repo ownerdomain.Repository, counter ownerdomain.PetOwnerCounter, target *model.Owner) error {
			commitErr := errors.New("owner delete completion failed")
			recorder := &ownerDeleteCleanupRecorder{
				enqueueNextVisitCleanupFn: realEnqueueNextVisitCleanup(t, db),
			}
			svc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
				repo, nil, recorder, nil, repo, counter, completionFailureTransactor{db: db, err: commitErr})
			err := svc.Delete(context.Background(), 1, target.ID)
			require.ErrorIs(t, err, commitErr)
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupOwnerDeleteLineCleanupTestDB(t)
			repo := newOwnerRepository(db)
			petOwnerRepo := petdomain.NewPetOwnerRepository(db)
			target := makeLineLinkedOwner(t, db, 1, tc.name, "Urollback")

			tc.run(t, db, repo, petOwnerRepo, target)

			assert.Zero(t, countCleanupIntentRows(t, db, 1, target.ID),
				"a failed deletion must never leave a committed cleanup intent")
			var reloaded model.Owner
			require.NoError(t, db.Unscoped().Where("id = ?", target.ID).First(&reloaded).Error)
			assert.False(t, reloaded.DeletedAt.Valid, "the owner deletion itself must roll back")
		})
	}
}

func TestOwnerService_Delete_PostCommitSyncFailureKeepsDeletion(t *testing.T) {
	db := setupOwnerDeleteLineCleanupTestDB(t)
	repo := newOwnerRepository(db)
	petOwnerRepo := petdomain.NewPetOwnerRepository(db)
	const clinicID = uint64(1)
	target := makeLineLinkedOwner(t, db, clinicID, "post commit failure target", "Upostcommit")
	syncErr := errors.New("post commit cleanup failed")
	recorder := &ownerDeleteCleanupRecorder{
		enqueueNextVisitCleanupFn: realEnqueueNextVisitCleanup(t, db),
		syncNextVisitTagFn:        func(context.Context, uint64, uint64) error { return syncErr },
	}
	svc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
		repo, nil, recorder, nil, repo, petOwnerRepo, persistence.NewTransactor(db))

	require.NoError(t, svc.Delete(context.Background(), clinicID, target.ID),
		"post-commit cleanup failure must not reverse the committed deletion")
	assert.Equal(t, 1, recorder.syncCalls)
	assert.Equal(t, int64(1), countCleanupIntentRows(t, db, clinicID, target.ID),
		"the intent survives for the daily retry to converge")
	var deleted model.Owner
	require.NoError(t, db.Unscoped().Where("id = ?", target.ID).First(&deleted).Error)
	assert.True(t, deleted.DeletedAt.Valid)
}

func TestOwnerService_Delete_CancelledRequestStillRunsBoundedCleanup(t *testing.T) {
	db := setupOwnerDeleteLineCleanupTestDB(t)
	repo := newOwnerRepository(db)
	petOwnerRepo := petdomain.NewPetOwnerRepository(db)
	const clinicID = uint64(1)
	target := makeLineLinkedOwner(t, db, clinicID, "cancelled request target", "Ucancelled")
	recorder := &ownerDeleteCleanupRecorder{
		enqueueNextVisitCleanupFn: realEnqueueNextVisitCleanup(t, db),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	svc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
		repo, nil, recorder, nil, repo, petOwnerRepo,
		cancelAfterCommitTransactor{inner: persistence.NewTransactor(db), cancel: cancel})

	require.NoError(t, svc.Delete(ctx, clinicID, target.ID))
	require.ErrorIs(t, ctx.Err(), context.Canceled, "the caller context was cancelled after commit")
	assert.Equal(t, 1, recorder.syncCalls, "cleanup still attempts once after a cancelled request")
	assert.False(t, recorder.syncCtxCanceled, "post-commit cleanup is detached from request cancellation")
	require.True(t, recorder.syncHasDeadline, "post-commit cleanup context must carry a bounded deadline")
	assert.LessOrEqual(t, time.Until(recorder.syncDeadline), 30*time.Second)
	assert.False(t, recorder.syncHadAmbientTx)
}

func TestOwnerService_Delete_UnlinkedOwnerDoesNotRequireCleanupDependency(t *testing.T) {
	db := setupOwnerDeleteLineCleanupTestDB(t)
	repo := newOwnerRepository(db)
	petOwnerRepo := petdomain.NewPetOwnerRepository(db)
	const clinicID = uint64(1)
	target := testdb.MakeTestOwner(t, db, clinicID, "unlinked delete target")

	svc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
		repo, nil, plainTagNotifier{}, nil, repo, petOwnerRepo, persistence.NewTransactor(db))
	require.NoError(t, svc.Delete(context.Background(), clinicID, target.ID))
	assert.Zero(t, countCleanupIntentRows(t, db, clinicID, target.ID))
}
