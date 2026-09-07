package medicalrecord

// treatment_plan_discount_toctou_test.go — SEC-CS-F10: discount TOCTOU under concurrent authorized edit.

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

func setupTreatmentPlanDiscountTOCTOUDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.TreatmentPlan{}))
	db.Exec("TRUNCATE TABLE treatment_plans CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func seedTreatmentPlanWithDiscount(t *testing.T, db *gorm.DB, clinicID uint64, discountAmount int64) (*model.MedicalRecord, *model.TreatmentPlan) {
	t.Helper()
	mr := makeTreatmentPlanMedicalRecord(t, db, clinicID, "TP-DISC-TOCTOU")
	// Override finalized seed to draft-friendly content; plan itself has no parent status gate.
	tp := &model.TreatmentPlan{
		ClinicID:         clinicID,
		MedicalRecordID:  &mr.ID,
		TreatmentContent: "seed plan",
		UnitPrice:        1000,
		Quantity:         1,
		DiscountAmount:   discountAmount,
		DiscountRate:     0,
		Subtotal:         1000 - discountAmount,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tp).Error)
	return mr, tp
}

type pauseAfterFirstTreatmentPlanLock struct {
	TreatmentPlanRepository
	reached chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *pauseAfterFirstTreatmentPlanLock) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
	locked, err := r.TreatmentPlanRepository.LockByIDForUpdate(ctx, clinicID, id)
	if err != nil {
		return nil, err
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
			return nil, ctx.Err()
		}
	}
	return locked, nil
}

func TestTreatmentPlanService_Update_DiscountTOCTOU_StaleZeroRejected(t *testing.T) {
	db := setupTreatmentPlanDiscountTOCTOUDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	const clinicID = uint64(1)

	mr, plan := seedTreatmentPlanWithDiscount(t, db, clinicID, 0)
	realRepo := NewTreatmentPlanRepository(db)
	pausingRepo := &pauseAfterFirstTreatmentPlanLock{
		TreatmentPlanRepository: realRepo,
		reached:                 make(chan struct{}),
		release:                 make(chan struct{}),
	}
	svc := NewTreatmentPlanService(pausingRepo, persistence.NewTransactor(db))
	mrID := mr.ID

	authDone := make(chan error, 1)
	go func() {
		amount := int64(10)
		_, err := svc.Update(ctx, clinicID, plan.ID, &mrID, nil, &UpdateTreatmentPlanInput{
			DiscountAmount:      &amount,
			DiscountEditAllowed: true,
		})
		authDone <- err
	}()

	select {
	case <-pausingRepo.reached:
	case <-ctx.Done():
		t.Fatal("authorized update did not reach treatment plan FOR UPDATE lock")
	}

	unprivDone := make(chan error, 1)
	go func() {
		staleZero := int64(0)
		_, err := svc.Update(ctx, clinicID, plan.ID, &mrID, nil, &UpdateTreatmentPlanInput{
			DiscountAmount:      &staleZero,
			DiscountEditAllowed: false,
		})
		unprivDone <- err
	}()

	select {
	case err := <-unprivDone:
		t.Fatalf("unprivileged update bypassed treatment plan FOR UPDATE lock: %v", err)
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

	got, err := realRepo.FindByID(context.Background(), clinicID, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10), got.DiscountAmount, "authorized discount must survive stale unprivileged overwrite")
}

func TestTreatmentPlanService_Update_DiscountTOCTOU_NonDiscountFieldStillOK(t *testing.T) {
	db := setupTreatmentPlanDiscountTOCTOUDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	mr, plan := seedTreatmentPlanWithDiscount(t, db, clinicID, 10)
	svc := NewTreatmentPlanService(NewTreatmentPlanRepository(db), persistence.NewTransactor(db))
	mrID := mr.ID
	memo := "non-discount edit without discount:edit"
	got, err := svc.Update(ctx, clinicID, plan.ID, &mrID, nil, &UpdateTreatmentPlanInput{
		Memo:                &memo,
		DiscountEditAllowed: false,
	})
	require.NoError(t, err)
	assert.Equal(t, memo, got.Memo)
	assert.Equal(t, int64(10), got.DiscountAmount)
}

func TestTreatmentPlanService_Update_DiscountTOCTOU_LockedDiffWithoutPermission(t *testing.T) {
	const clinicID, planID = uint64(1), uint64(20)
	mrID := uint64(10)
	zero := int64(0)
	repo := &mockTreatmentPlanRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.TreatmentPlan, error) {
			return &model.TreatmentPlan{
				ID: id, ClinicID: clinicID, MedicalRecordID: &mrID,
				UnitPrice: 100, Quantity: 1, DiscountAmount: 0,
			}, nil
		},
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.TreatmentPlan, error) {
			return &model.TreatmentPlan{
				ID: id, ClinicID: clinicID, MedicalRecordID: &mrID,
				UnitPrice: 100, Quantity: 1, DiscountAmount: 10,
			}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _, _ *uint64, _ UpdateTreatmentPlanInput) error {
			t.Fatal("Update must not run when discount recheck forbids the write")
			return nil
		},
	}
	svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})
	_, err := svc.Update(context.Background(), clinicID, planID, &mrID, nil, &UpdateTreatmentPlanInput{
		DiscountAmount:      &zero,
		DiscountEditAllowed: false,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "got: %v", err)
}

func TestTreatmentPlanRepository_LockByIDForUpdate_RequiresAmbientTransaction(t *testing.T) {
	db := setupTreatmentPlanDiscountTOCTOUDB(t)
	repo := NewTreatmentPlanRepository(db)
	_, err := repo.LockByIDForUpdate(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}
