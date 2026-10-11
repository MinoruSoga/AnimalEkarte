package trimming

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func waitForPIDAdvisoryWait(ctx context.Context, db *gorm.DB, pid int, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var waiting bool
		if err := db.WithContext(ctx).
			Raw("SELECT COUNT(*) > 0 FROM pg_locks WHERE pid = ? AND locktype = 'advisory' AND NOT granted", pid).
			Scan(&waiting).Error; err != nil {
			return false, err
		}
		if waiting {
			return true, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false, nil
}

func TestDeleteUnreferencedImageObject_WaitedSecondConnClaimRejectedAfterDelete(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	db := setupImageReferenceTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	clinicID := uint64(21)
	keyK := "trimming-images/21/window-claim.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	dir := t.TempDir()
	uploader := infra.NewLocalUploader(dir, "/uploads")
	_, err := uploader.Upload(ctx, keyK, strings.NewReader("img"), "image/jpeg")
	require.NoError(t, err)

	svc := newLifecycleService(t, db)

	db2, err := gorm.Open(postgres.Open(lifecycleTestDSN()), &gorm.Config{})
	require.NoError(t, err)
	repo2 := NewAppointmentTrimmingDetailRepository(db2)
	bindLifecycleRepoUploader(repo2, uploader)

	inWindow := make(chan struct{})
	var inWindowOnce sync.Once
	releaseWindow := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseWindow) }) }
	cleanupDone := make(chan struct {
		deleted bool
		err     error
	}, 1)
	cleanupFinished := make(chan struct{})
	writerPID := make(chan int, 1)
	writerDone := make(chan error, 1)
	writerFinished := make(chan struct{})
	writerStarted := false

	defer func() {
		release()
		cancel()
		select {
		case <-cleanupFinished:
		case <-time.After(15 * time.Second):
			t.Errorf("cleanup goroutine did not finish")
		}
		if writerStarted {
			select {
			case <-writerFinished:
			case <-time.After(15 * time.Second):
				t.Errorf("writer goroutine did not finish")
			}
		}
		if sqlDB2, err := db2.DB(); assert.NoError(t, err) {
			assert.NoError(t, sqlDB2.Close())
		}
	}()

	deleteFn := func(delCtx context.Context, key string) error {
		inWindowOnce.Do(func() { close(inWindow) })
		select {
		case <-releaseWindow:
		case <-delCtx.Done():
			return delCtx.Err()
		}
		return uploader.Delete(delCtx, key)
	}

	go func() {
		defer close(cleanupFinished)
		deleted, err := svc.DeleteUnreferencedImageObject(ctx, clinicID, keyK, deleteFn)
		cleanupDone <- struct {
			deleted bool
			err     error
		}{deleted: deleted, err: err}
	}()

	select {
	case <-inWindow:
	case <-ctx.Done():
		require.FailNow(t, "cleanup did not reach the delete window")
	}

	writerStarted = true
	go func() {
		defer close(writerFinished)
		err := db2.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var pid int
			if err := tx.Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
				return err
			}
			writerPID <- pid
			return repo2.Create(persistence.WithTxValue(ctx, tx), &model.AppointmentTrimmingDetail{
				ClinicID:      clinicID,
				AppointmentID: appt.ID,
				StyleImage:    keyK,
				BWUnit:        model.BodyWeightUnitKg,
			})
		})
		writerDone <- err
	}()

	var pid int
	select {
	case pid = <-writerPID:
	case <-ctx.Done():
		require.FailNow(t, "writer did not report its backend pid")
	}

	waiting, werr := waitForPIDAdvisoryWait(ctx, db, pid, 5*time.Second)
	require.NoError(t, werr, "pg_locks 観測クエリの失敗は待機なしと区別する")
	require.True(t, waiting,
		"writer pid %d did not appear waiting on the key advisory lock inside the window", pid)

	release()

	select {
	case result := <-cleanupDone:
		require.NoError(t, result.err)
		assert.True(t, result.deleted, "cleanup はオブジェクトを削除したはず")
	case <-ctx.Done():
		require.FailNow(t, "cleanup did not finish")
	}
	_, statErr := os.Stat(filepath.Join(dir, filepath.FromSlash(keyK)))
	require.True(t, os.IsNotExist(statErr), "オブジェクトは削除済み")

	select {
	case err := <-writerDone:
		require.Error(t, err, "削除済みオブジェクトキーへの claim は拒否されなければならない")
		assert.True(t, apperrors.IsConflict(err),
			"ウィンドウ後の writer は typed conflict で拒否されなければならない: %v", err)
	case <-ctx.Done():
		require.FailNow(t, "writer did not finish")
	}

	var refs int64
	require.NoError(t, db.Model(&model.AppointmentTrimmingDetail{}).
		Where("clinic_id = ? AND appointment_id = ? AND style_image = ?", clinicID, appt.ID, keyK).
		Count(&refs).Error)
	assert.Zero(t, refs, "dangling reference をコミットしてはいけない")
}
