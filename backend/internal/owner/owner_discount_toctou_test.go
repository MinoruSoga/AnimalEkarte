package owner

// owner_discount_toctou_test.go — SEC-CS-F15: discount_rate TOCTOU under concurrent authorized edit.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupOwnerDiscountTOCTOUDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.SetupTestDB(t)
}

// pauseAfterFirstOwnerApply holds the first UpdateAndFindApplying after lock until release,
// so a concurrent unprivileged update queues on FOR UPDATE.
type pauseAfterFirstOwnerApply struct {
	inner   *ownerRepository
	reached chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *pauseAfterFirstOwnerApply) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	return r.inner.FindAll(ctx, clinicIDs, page, limit, search)
}
func (r *pauseAfterFirstOwnerApply) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return r.inner.FindByID(ctx, clinicID, id)
}
func (r *pauseAfterFirstOwnerApply) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
	return r.inner.FindByIDForClinics(ctx, clinicIDs, id)
}
func (r *pauseAfterFirstOwnerApply) FindByEmail(ctx context.Context, clinicID uint64, email string) (*model.Owner, error) {
	return r.inner.FindByEmail(ctx, clinicID, email)
}
func (r *pauseAfterFirstOwnerApply) FindByPhone(ctx context.Context, clinicID uint64, phone string) (*model.Owner, error) {
	return r.inner.FindByPhone(ctx, clinicID, phone)
}
func (r *pauseAfterFirstOwnerApply) FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error) {
	return r.inner.FindByLineUserID(ctx, clinicID, lineUserID)
}
func (r *pauseAfterFirstOwnerApply) CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error {
	return r.inner.CreateWithPets(ctx, owner, pets)
}
func (r *pauseAfterFirstOwnerApply) UpdateAndFind(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Owner, error) {
	return r.inner.UpdateAndFind(ctx, clinicID, id, fields)
}
func (r *pauseAfterFirstOwnerApply) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return r.inner.LockByIDForUpdate(ctx, clinicID, id)
}
func (r *pauseAfterFirstOwnerApply) UpdateLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error {
	return r.inner.UpdateLineUserID(ctx, clinicID, id, lineUserID)
}
func (r *pauseAfterFirstOwnerApply) Delete(ctx context.Context, clinicID, id uint64) error {
	return r.inner.Delete(ctx, clinicID, id)
}
func (r *pauseAfterFirstOwnerApply) CountPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	return r.inner.CountPetsByOwnerID(ctx, clinicID, ownerID)
}

func (r *pauseAfterFirstOwnerApply) UpdateAndFindApplying(
	ctx context.Context,
	clinicID, id uint64,
	apply OwnerUpdateApplier,
) (*model.Owner, error) {
	if apply == nil {
		return nil, apperrors.WrapInternalServerError("owner update applier is required")
	}
	var loaded *model.Owner
	err := persistence.DBOrTx(ctx, r.inner.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		locked, err := r.inner.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return err
		}

		shouldPause := false
		r.once.Do(func() {
			shouldPause = true
			close(r.reached)
		})
		if shouldPause {
			select {
			case <-r.release:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		fields, err := apply(locked)
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			return apperrors.WrapInvalidInput("at least one field must be provided")
		}
		if err := persistence.UpdateScopedByID(txCtx, tx, &model.Owner{}, "owner", clinicID, id, fields); err != nil {
			return err
		}
		loaded, err = r.inner.findOwnerByIDWithDB(txCtx, tx, []uint64{clinicID}, id)
		return err
	})
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, apperrors.Wrap(err, "failed to update and reload owner")
	}
	return loaded, nil
}

func TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected(t *testing.T) {
	db := setupOwnerDiscountTOCTOUDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	const clinicID = uint64(1)

	seed := testdb.MakeTestOwner(t, db, clinicID, "discount-toctou-owner")
	require.NoError(t, db.WithContext(ctx).Model(seed).Update("discount_rate", 0).Error)

	inner := NewRepository(db, nil).(*ownerRepository)
	pausingRepo := &pauseAfterFirstOwnerApply{
		inner:   inner,
		reached: make(chan struct{}),
		release: make(chan struct{}),
	}
	svc := NewService(pausingRepo, nil, nil, nil)

	authDone := make(chan error, 1)
	go func() {
		rate := 10.0
		_, err := svc.Update(ctx, clinicID, seed.ID, &UpdateOwnerInput{
			DiscountRate:        &rate,
			DiscountEditAllowed: true,
		})
		authDone <- err
	}()

	select {
	case <-pausingRepo.reached:
	case <-ctx.Done():
		t.Fatal("authorized update did not reach owner FOR UPDATE lock")
	}

	unprivDone := make(chan error, 1)
	go func() {
		staleZero := 0.0
		_, err := svc.Update(ctx, clinicID, seed.ID, &UpdateOwnerInput{
			DiscountRate:        &staleZero,
			DiscountEditAllowed: false,
		})
		unprivDone <- err
	}()

	select {
	case err := <-unprivDone:
		t.Fatalf("unprivileged update bypassed owner FOR UPDATE lock: %v", err)
	case <-time.After(250 * time.Millisecond):
	}

	close(pausingRepo.release)

	select {
	case err := <-authDone:
		require.NoError(t, err, "authorized discount edit must succeed")
	case <-ctx.Done():
		t.Fatal("authorized update did not complete")
	}

	select {
	case err := <-unprivDone:
		require.Error(t, err, "stale unprivileged discount=0 must fail after lock recheck")
		assert.True(t, errors.Is(err, apperrors.ErrForbidden), "want Forbidden, got: %v", err)
	case <-ctx.Done():
		t.Fatal("unprivileged update did not complete")
	}

	got, err := inner.FindByID(context.Background(), clinicID, seed.ID)
	require.NoError(t, err)
	assert.InDelta(t, 10.0, got.DiscountRate, 0.0001, "authorized discount must survive stale unprivileged overwrite")
}

func TestOwnerService_Update_DiscountTOCTOU_NonDiscountFieldStillOK(t *testing.T) {
	db := setupOwnerDiscountTOCTOUDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	seed := testdb.MakeTestOwner(t, db, clinicID, "before")
	require.NoError(t, db.WithContext(ctx).Model(seed).Update("discount_rate", 10).Error)

	svc := NewService(NewRepository(db, nil), nil, nil, nil)
	name := "after non-discount edit"
	got, err := svc.Update(ctx, clinicID, seed.ID, &UpdateOwnerInput{
		OwnerName:           &name,
		DiscountEditAllowed: false,
	})
	require.NoError(t, err)
	assert.Equal(t, name, got.Name)
	assert.InDelta(t, 10.0, got.DiscountRate, 0.0001)
}

func TestOwnerService_Update_DiscountTOCTOU_LockedDiffWithoutPermission(t *testing.T) {
	const clinicID, ownerID = uint64(1), uint64(10)
	zero := 0.0
	repo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			// Locked snapshot has discount_rate=10; request carries stale 0 without permission.
			return &model.Owner{ID: id, ClinicID: clinicID, DiscountRate: 10}, nil
		},
		updateAndFindFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Owner, error) {
			t.Fatal("UpdateAndFind must not run when discount recheck forbids the write")
			return nil, nil
		},
	}
	svc := NewService(repo, nil, nil, nil)
	_, err := svc.Update(context.Background(), clinicID, ownerID, &UpdateOwnerInput{
		DiscountRate:        &zero,
		DiscountEditAllowed: false,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "got: %v", err)
}

func TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction(t *testing.T) {
	db := setupOwnerDiscountTOCTOUDB(t)
	repo := NewRepository(db, nil)
	_, err := repo.LockByIDForUpdate(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}
