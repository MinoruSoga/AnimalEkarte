package trimming

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// Run only against an explicitly approved disposable test database. -short
// skips setup entirely. Table DDL comes from the real migration; full migration
// and RLS application remain a separate fresh-schema gate.
func setupImageLifecyclePostgres(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupImageReferenceTestDB(t)
	ddl, err := os.ReadFile(filepath.Join("..", "..", "migrations", "032_trimming_image_deletions.sql"))
	require.NoError(t, err)
	tableDDL, _, found := strings.Cut(string(ddl), "SELECT app_private.apply_rls_policy")
	require.True(t, found)
	tableDDL = strings.Replace(tableDDL, "CREATE TABLE trimming_image_deletions", "CREATE TABLE IF NOT EXISTS trimming_image_deletions", 1)
	require.NoError(t, db.Exec(tableDDL).Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE trimming_image_deletions").Error)
	return db
}

func TestImageLifecycle_PostgresRollbackLocksAndScope(t *testing.T) {
	db := setupImageLifecyclePostgres(t)
	repo := NewAppointmentTrimmingDetailRepository(db).(ImageLifecycleRepository)
	transactor := newTestTransactor(db)
	key := ImageKey{clinicID: 1, value: "trimming-images/1/style.jpg"}
	rollback := errors.New("roll back reservation")
	require.ErrorIs(t, transactor.WithTx(context.Background(), func(ctx context.Context) error {
		require.NoError(t, repo.LockImageKeys(ctx, []ImageKey{key}))
		require.NoError(t, repo.MarkImageKeyDeleted(ctx, key, 42))
		return rollback
	}), rollback)
	require.NoError(t, transactor.WithTx(context.Background(), func(ctx context.Context) error {
		deleted, err := repo.ImageKeyDeleted(ctx, key)
		require.NoError(t, err)
		assert.False(t, deleted)
		return nil
	}))

	locked, release := make(chan struct{}), make(chan struct{})
	committed := make(chan error, 1)
	workerDone := make(chan struct{})
	workerCtx, cancelWorker := context.WithTimeout(context.Background(), 10*time.Second)
	var releaseOnce sync.Once
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
		cancelWorker()
		select {
		case <-workerDone:
		case <-time.After(10 * time.Second):
			t.Error("image lifecycle worker did not stop within 10s")
		}
	})
	go func() {
		defer close(workerDone)
		committed <- transactor.WithTx(workerCtx, func(ctx context.Context) error {
			if err := repo.LockImageKeys(ctx, []ImageKey{key}); err != nil {
				return err
			}
			close(locked)
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
			return repo.MarkImageKeyDeleted(ctx, key, 42)
		})
	}()
	select {
	case <-locked:
	case err := <-committed:
		t.Fatalf("image lifecycle worker exited before acquiring the lock: %v", err)
	case <-workerCtx.Done():
		t.Fatalf("image lifecycle worker did not acquire the lock before deadline: %v", workerCtx.Err())
	}
	// A competing transaction cannot pass the same key lock before commit.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		return repo.LockImageKeys(txCtx, []ImageKey{key})
	})
	cancel()
	releaseOnce.Do(func() { close(release) })
	require.Error(t, err)
	require.NoError(t, <-committed)
	require.NoError(t, transactor.WithTx(context.Background(), func(ctx context.Context) error {
		require.NoError(t, repo.LockImageKeys(ctx, []ImageKey{key}))
		deleted, err := repo.ImageKeyDeleted(ctx, key)
		require.NoError(t, err)
		assert.True(t, deleted)
		foreignKey := ImageKey{clinicID: 2, value: "trimming-images/2/style.jpg"}
		deleted, err = repo.ImageKeyDeleted(ctx, foreignKey)
		require.NoError(t, err)
		assert.False(t, deleted)
		// A retry cannot replace the first actor/timestamp reservation.
		return repo.MarkImageKeyDeleted(ctx, key, 43)
	}))
	var actorID uint64
	require.NoError(t, db.Raw("SELECT actor_staff_id FROM trimming_image_deletions WHERE clinic_id = ? AND object_key = ?", 1, key.value).Scan(&actorID).Error)
	assert.Equal(t, uint64(42), actorID)
}

func TestImageLifecycle_PostgresRestoredAndSharedReferencesRetained(t *testing.T) {
	db := setupImageLifecyclePostgres(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	svc := &trimmingService{trimmingDetail: repo, transactor: newTestTransactor(db)}
	const key = "trimming-images/1/shared.jpg"
	selfID := insertTrimmingDetailWithImages(t, db, 1, key, "")
	reserved, err := svc.ReserveImageDeletion(context.Background(), 1, key, selfID, 42)
	require.NoError(t, err)
	assert.False(t, reserved, "an active appointment cannot be excluded from references")
	otherID := insertTrimmingDetailWithImages(t, db, 1, "", key)
	reserved, err = svc.ReserveImageDeletion(context.Background(), 1, key, 0, 42)
	require.NoError(t, err)
	assert.False(t, reserved, "restored current reference participates in PATCH cleanup")
	require.NoError(t, db.Delete(&model.Reservation{ID: selfID}).Error)
	reserved, err = svc.ReserveImageDeletion(context.Background(), 1, key, selfID, 42)
	require.NoError(t, err)
	assert.False(t, reserved, "shared live reference survives DELETE cleanup")
	require.NoError(t, db.Model(&model.AppointmentTrimmingDetail{}).Where("clinic_id = ? AND appointment_id = ?", 1, otherID).Update("completed_image", "").Error)
	reserved, err = svc.ReserveImageDeletion(context.Background(), 1, key, selfID, 42)
	require.NoError(t, err)
	assert.True(t, reserved)
	err = newTestTransactor(db).WithTx(context.Background(), func(ctx context.Context) error {
		return svc.lockImageReferencesInTx(ctx, 1, key)
	})
	require.Error(t, err, "reference must be rejected after deletion reservation commits")
}
