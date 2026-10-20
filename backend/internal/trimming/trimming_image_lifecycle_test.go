package trimming

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type imageTestTxKey struct{}
type imageTestTx struct {
	unlock []func()
	commit []func()
}

// This fake models the required ambient-transaction lock lifetime and staged
// commit. It proves service interleavings without a database or external storage.
type imageLifecycleFake struct {
	mockTrimmingDetailRepository
	mu        sync.Mutex
	locks     map[ImageKey]chan struct{}
	deleted   map[ImageKey]bool
	refs      map[ImageKey]map[uint64]bool
	attempts  chan ImageKey
	probe     func()
	commitErr error
	markErr   error
}

func newImageLifecycleFake() *imageLifecycleFake {
	return &imageLifecycleFake{
		locks: make(map[ImageKey]chan struct{}), deleted: make(map[ImageKey]bool),
		refs: make(map[ImageKey]map[uint64]bool), attempts: make(chan ImageKey, 10),
	}
}

func (f *imageLifecycleFake) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx := &imageTestTx{}
	defer func() {
		for i := len(tx.unlock) - 1; i >= 0; i-- {
			tx.unlock[i]()
		}
	}()
	if err := fn(context.WithValue(ctx, imageTestTxKey{}, tx)); err != nil {
		return err
	}
	if f.commitErr != nil {
		return f.commitErr
	}
	for _, commit := range tx.commit {
		commit()
	}
	return nil
}

func (f *imageLifecycleFake) LockImageKeys(ctx context.Context, keys []ImageKey) error {
	tx, ok := ctx.Value(imageTestTxKey{}).(*imageTestTx)
	if !ok {
		return errors.New("missing fake transaction")
	}
	for _, key := range keys {
		f.mu.Lock()
		lock := f.locks[key]
		if lock == nil {
			lock = make(chan struct{}, 1)
			lock <- struct{}{}
			f.locks[key] = lock
		}
		f.mu.Unlock()
		f.attempts <- key
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-lock:
			tx.unlock = append(tx.unlock, func() { lock <- struct{}{} })
		}
	}
	return nil
}

func (f *imageLifecycleFake) ImageKeyDeleted(_ context.Context, key ImageKey) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.deleted[key], nil
}

func (f *imageLifecycleFake) MarkImageKeyDeleted(ctx context.Context, key ImageKey, _ uint64) error {
	if f.markErr != nil {
		return f.markErr
	}
	tx := ctx.Value(imageTestTxKey{}).(*imageTestTx)
	tx.commit = append(tx.commit, func() { f.mu.Lock(); defer f.mu.Unlock(); f.deleted[key] = true })
	return nil
}

func (f *imageLifecycleFake) ImageKeyInUse(_ context.Context, clinicID uint64, key string, exceptID uint64) (bool, error) {
	if f.probe != nil {
		f.probe()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for id := range f.refs[ImageKey{clinicID: clinicID, value: key}] {
		if id != exceptID {
			return true, nil
		}
	}
	return false, nil
}

func (f *imageLifecycleFake) stageReference(ctx context.Context, key ImageKey, id uint64) {
	tx := ctx.Value(imageTestTxKey{}).(*imageTestTx)
	tx.commit = append(tx.commit, func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.refs[key] == nil {
			f.refs[key] = make(map[uint64]bool)
		}
		f.refs[key][id] = true
	})
}

func lifecycleService(f *imageLifecycleFake) *trimmingService {
	return &trimmingService{trimmingDetail: f, transactor: f}
}

func TestImageLifecycle_DeletionBeforeSaveRejectsReference(t *testing.T) {
	f := newImageLifecycleFake()
	s := lifecycleService(f)
	key := ImageKey{clinicID: 1, value: "trimming-images/1/style.jpg"}
	checked, proceed := make(chan struct{}), make(chan struct{})
	f.probe = func() { close(checked); <-proceed }
	deleted := make(chan error, 1)
	go func() {
		reserved, err := s.ReserveImageDeletion(context.Background(), 1, key.value, 0, 42)
		if err == nil && !reserved {
			err = errors.New("deletion was not reserved")
		}
		deleted <- err
	}()
	<-checked // deletion holds the key lock after its no-reference query
	<-f.attempts
	written := make(chan error, 1)
	go func() {
		written <- f.WithTx(context.Background(), func(ctx context.Context) error {
			if err := s.lockImageReferencesInTx(ctx, 1, key.value); err != nil {
				return err
			}
			f.stageReference(ctx, key, 7)
			return nil
		})
	}()
	<-f.attempts // writer tries the same lock before deletion commits
	close(proceed)
	require.NoError(t, <-deleted)
	require.ErrorIs(t, <-written, apperrors.ErrConflict)
	assert.True(t, f.deleted[key], "storage deletion can now run outside the DB transaction")
	assert.Empty(t, f.refs[key], "no new reference can commit after the tombstone")
}

func TestImageLifecycle_SaveBeforeDeletionRetainsSharedObject(t *testing.T) {
	f := newImageLifecycleFake()
	s := lifecycleService(f)
	key := ImageKey{clinicID: 1, value: "trimming-images/1/style.jpg"}
	locked, proceed := make(chan struct{}), make(chan struct{})
	written := make(chan error, 1)
	go func() {
		written <- f.WithTx(context.Background(), func(ctx context.Context) error {
			if err := s.lockImageReferencesInTx(ctx, 1, key.value); err != nil {
				return err
			}
			f.stageReference(ctx, key, 8)
			close(locked)
			<-proceed
			return nil
		})
	}()
	<-locked
	<-f.attempts
	deleted := make(chan error, 1)
	go func() {
		reserved, err := s.ReserveImageDeletion(context.Background(), 1, key.value, 7, 42)
		if err == nil && reserved {
			err = errors.New("shared key deletion was incorrectly reserved")
		}
		deleted <- err
	}()
	<-f.attempts // cleanup cannot check references until the writer commits
	close(proceed)
	require.NoError(t, <-written)
	require.NoError(t, <-deleted)
	assert.False(t, f.deleted[key])
	assert.True(t, f.refs[key][8])
}

func TestImageLifecycle_RollbackAndRetry(t *testing.T) {
	for _, failure := range []string{"mark", "commit"} {
		t.Run(failure, func(t *testing.T) {
			f := newImageLifecycleFake()
			s := lifecycleService(f)
			key := ImageKey{clinicID: 1, value: "trimming-images/1/style.jpg"}
			sentinel := errors.New("database failure")
			if failure == "mark" {
				f.markErr = sentinel
			} else {
				f.commitErr = sentinel
			}
			reserved, err := s.ReserveImageDeletion(context.Background(), 1, key.value, 0, 42)
			require.ErrorIs(t, err, sentinel)
			assert.False(t, reserved, "the handler must not delete storage on DB failure")
			assert.False(t, f.deleted[key])
			f.markErr, f.commitErr = nil, nil
			reserved, err = s.ReserveImageDeletion(context.Background(), 1, key.value, 0, 42)
			require.NoError(t, err)
			require.True(t, reserved)
			// Storage failure cannot undo the committed marker. A retry is safe.
			reserved, err = s.ReserveImageDeletion(context.Background(), 1, key.value, 0, 42)
			require.NoError(t, err)
			assert.True(t, reserved)
			err = f.WithTx(context.Background(), func(ctx context.Context) error {
				return s.lockImageReferencesInTx(ctx, 1, key.value)
			})
			require.ErrorIs(t, err, apperrors.ErrConflict)
		})
	}
}

func TestImageLifecycle_DeletionRetryPreservesRetiredDetailReference(t *testing.T) {
	f := newImageLifecycleFake()
	s := lifecycleService(f)
	key := ImageKey{clinicID: 1, value: "trimming-images/1/style.jpg"}
	f.refs[key] = map[uint64]bool{7: true}
	reserved, err := s.ReserveImageDeletion(t.Context(), 1, key.value, 7, 42)
	require.NoError(t, err)
	assert.False(t, reserved)
	assert.False(t, f.deleted[key])
}

func TestImageLifecycle_ScopeSortedLocksAndMissingCapability(t *testing.T) {
	f := newImageLifecycleFake()
	s := lifecycleService(f)
	for _, tc := range []struct {
		name     string
		clinicID uint64
		key      string
	}{
		{"foreign", 1, "trimming-images/2/a.jpg"},
		{"traversal", 1, "trimming-images/1/../a.jpg"},
		{"URL", 1, "https://cdn.example/trimming-images/1/a.jpg"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reserved, err := s.ReserveImageDeletion(context.Background(), tc.clinicID, tc.key, 0, 42)
			require.ErrorIs(t, err, apperrors.ErrInvalidInput)
			assert.False(t, reserved)
		})
	}
	require.NoError(t, f.WithTx(context.Background(), func(ctx context.Context) error {
		return s.lockImageReferencesInTx(ctx, 1, "trimming-images/1/z.jpg", "trimming-images/1/a.jpg", "trimming-images/1/a.jpg")
	}))
	assert.Equal(t, "trimming-images/1/a.jpg", (<-f.attempts).value)
	assert.Equal(t, "trimming-images/1/z.jpg", (<-f.attempts).value)
	assert.Empty(t, f.attempts, "duplicate references must not relock a key")
	missing := &trimmingService{trimmingDetail: &mockTrimmingDetailRepository{}, transactor: &mockTransactor{}}
	reserved, err := missing.ReserveImageDeletion(context.Background(), 1, "trimming-images/1/a.jpg", 0, 42)
	require.Error(t, err)
	assert.False(t, reserved)
	err = missing.lockImageReferencesInTx(context.Background(), 1, "trimming-images/1/a.jpg")
	require.Error(t, err)
}

func TestImageLifecycle_RepositoryRequiresAmbientTransaction(t *testing.T) {
	repo := &appointmentTrimmingDetailRepository{}
	key := ImageKey{clinicID: 1, value: "trimming-images/1/a.jpg"}
	require.ErrorContains(t, repo.LockImageKeys(context.Background(), []ImageKey{key}), "ambient transaction")
	_, err := repo.ImageKeyDeleted(context.Background(), key)
	require.Error(t, err)
	require.ErrorContains(t, repo.MarkImageKeyDeleted(context.Background(), key, 42), "ambient transaction")
}

func TestDeleteWithImages_UsesLockedSnapshotAndDiscardsOnRollback(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint("rollback=", fail), func(t *testing.T) {
			locked := false
			reservationRepo := &mockTrimmingReservationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
					t.Fatal("delete must not read a pre-lock snapshot")
					return nil, nil
				},
				lockAndFindByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
					locked = true
					return newTrimmingAuditAppointment(7, 1), nil
				},
			}
			detail := &mockTrimmingDetailRepository{findByAppointmentIDFn: func(context.Context, uint64, uint64) (*model.AppointmentTrimmingDetail, error) {
				require.True(t, locked)
				return &model.AppointmentTrimmingDetail{StyleImage: "trimming-images/1/new.jpg"}, nil
			}}
			audit := &trimmingAuditRecorder{}
			if fail {
				audit.err = errors.New("audit failed")
			}
			svc := newTrimmingAuditTestService(reservationRepo, detail, &mockTransactor{}, audit)
			images, err := svc.DeleteWithImages(context.Background(), 1, 7, ptrUint64(42))
			if fail {
				require.Error(t, err)
				assert.Nil(t, images)
			} else {
				require.NoError(t, err)
				require.NotNil(t, images)
				assert.Equal(t, "trimming-images/1/new.jpg", images.Style)
			}
		})
	}
}

// These lifecycle fakes represent positively proven unsubmitted work copies.
func (f *imageLifecycleFake) CourseImageDiscardEligible(context.Context, ImageKey, uint64) (bool, error) {
	return true, nil
}
func (f *imageLifecycleFake) ProtectCourseImageReference(context.Context, ImageKey) error { return nil }
