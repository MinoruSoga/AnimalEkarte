package medicalrecord

// treatment_discount_toctou_test.go — SEC-CS-F09: discount TOCTOU under concurrent authorized edit.
//
// Race: handler early-check uses pre-TX GetByID (discount=0). Concurrent authorized write sets
// discount=10. Unprivileged PATCH still carrying discount=0 (stale equal to old snapshot) must
// fail after FOR UPDATE recheck; locked value stays 10. Non-discount field updates remain OK.

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

func setupTreatmentDiscountTOCTOUDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Treatment{},
	))
	db.Exec("TRUNCATE TABLE treatments CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func seedDraftTreatmentWithDiscount(t *testing.T, db *gorm.DB, clinicID uint64, discountAmount int64) (*model.MedicalRecord, *model.Treatment) {
	t.Helper()
	owner := makeTestOwner(t, db, clinicID, "discount-toctou-owner")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "discount-toctou-pet")
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		PetID:    &pet.ID,
		RecordNo: "TR-DISC-TOCTOU",
		Date:     time.Now(),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	tr := &model.Treatment{
		MedicalRecordID: mr.ID,
		ItemType:        model.TreatmentItemTypeConsultation,
		Status:          model.TreatmentStatusPending,
		UnitPrice:       1000,
		Quantity:        1,
		DiscountAmount:  discountAmount,
		DiscountRate:    0,
		Content:         "seed",
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tr).Error)
	return mr, tr
}

// draftMRForTreatment seeds no staff tables; service only needs draft status for the parent gate.
// Treatment FOR UPDATE concurrency is still exercised against the real treatment row.
func draftMRForTreatment(mrID uint64) *mockMedicalRecordRepository {
	return &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: id, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
}

// pauseAfterFirstTreatmentLock holds the first LockByIDForUpdate until release is closed,
// so a concurrent unprivileged update can queue behind the authorized write.
type pauseAfterFirstTreatmentLock struct {
	TreatmentRepository
	reached chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *pauseAfterFirstTreatmentLock) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
	locked, err := r.TreatmentRepository.LockByIDForUpdate(ctx, clinicID, id)
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

func TestTreatmentService_Update_DiscountTOCTOU_StaleZeroRejected(t *testing.T) {
	db := setupTreatmentDiscountTOCTOUDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	const clinicID = uint64(1)

	mr, tr := seedDraftTreatmentWithDiscount(t, db, clinicID, 0)
	realRepo := NewTreatmentRepository(db)
	pausingRepo := &pauseAfterFirstTreatmentLock{
		TreatmentRepository: realRepo,
		reached:             make(chan struct{}),
		release:             make(chan struct{}),
	}
	svc := NewTreatmentServiceWithAudit(
		pausingRepo,
		draftMRForTreatment(mr.ID),
		okMedicineRepo(),
		okProcedureRepo(),
		okConsultationRepo(),
		&mockInventoryRepository{},
		benignVitalRepo(),
		&mockMedicineDoseParamRepository{},
		persistence.NewTransactor(db),
		nil,
	)

	// Authorized concurrent write: set discount_amount=10 while holding the row lock.
	authDone := make(chan error, 1)
	go func() {
		amount := int64(10)
		_, err := svc.Update(ctx, clinicID, mr.ID, tr.ID, &UpdateTreatmentInput{
			DiscountAmount:      &amount,
			DiscountEditAllowed: true,
		})
		authDone <- err
	}()

	select {
	case <-pausingRepo.reached:
	case <-ctx.Done():
		t.Fatal("authorized update did not reach treatment FOR UPDATE lock")
	}

	// Unprivileged stale PATCH starts while authorized still holds FOR UPDATE.
	// Pre-TX FindByID still sees discount=0 (stale equality); without lock recheck it would
	// overwrite the authorized 10 back to 0 after the lock is released.
	unprivDone := make(chan error, 1)
	go func() {
		staleZero := int64(0)
		_, err := svc.Update(ctx, clinicID, mr.ID, tr.ID, &UpdateTreatmentInput{
			DiscountAmount:      &staleZero,
			DiscountEditAllowed: false,
		})
		unprivDone <- err
	}()

	// Unprivileged must block on the treatment row lock (not finish early).
	select {
	case err := <-unprivDone:
		t.Fatalf("unprivileged update bypassed treatment FOR UPDATE lock: %v", err)
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

	got, err := realRepo.FindByID(context.Background(), clinicID, tr.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10), got.DiscountAmount, "authorized discount must survive stale unprivileged overwrite")
}

func TestTreatmentService_Update_DiscountTOCTOU_NonDiscountFieldStillOK(t *testing.T) {
	db := setupTreatmentDiscountTOCTOUDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	mr, tr := seedDraftTreatmentWithDiscount(t, db, clinicID, 10)
	// Bump content so seed is distinct.
	require.NoError(t, db.WithContext(ctx).Model(tr).Update("content", "before").Error)

	svc := NewTreatmentServiceWithAudit(
		NewTreatmentRepository(db),
		draftMRForTreatment(mr.ID),
		okMedicineRepo(),
		okProcedureRepo(),
		okConsultationRepo(),
		&mockInventoryRepository{},
		benignVitalRepo(),
		&mockMedicineDoseParamRepository{},
		persistence.NewTransactor(db),
		nil,
	)

	memo := "non-discount edit without discount:edit"
	got, err := svc.Update(ctx, clinicID, mr.ID, tr.ID, &UpdateTreatmentInput{
		Memo:                &memo,
		DiscountEditAllowed: false,
	})
	require.NoError(t, err)
	assert.Equal(t, memo, got.Memo)
	assert.Equal(t, int64(10), got.DiscountAmount, "non-discount update must not touch discount")
}

func TestTreatmentService_Update_DiscountTOCTOU_LockedDiffWithoutPermission(t *testing.T) {
	// Unit-style: pre-TX FindByID sees 0, locked row sees 10, request carries 0 without permission.
	const clinicID, mrID, trID = uint64(1), uint64(10), uint64(20)
	zero := int64(0)
	repo := &mockTreatmentRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Treatment, error) {
			return &model.Treatment{
				ID: id, MedicalRecordID: mrID, DiscountAmount: 0, Quantity: 1, UnitPrice: 100,
				ItemType: model.TreatmentItemTypeConsultation, Status: model.TreatmentStatusPending,
			}, nil
		},
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Treatment, error) {
			return &model.Treatment{
				ID: id, MedicalRecordID: mrID, DiscountAmount: 10, Quantity: 1, UnitPrice: 100,
				ItemType: model.TreatmentItemTypeConsultation, Status: model.TreatmentStatusPending,
			}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			t.Fatal("Update must not run when discount recheck forbids the write")
			return nil
		},
	}
	svc := newTreatmentSvc(repo, draftMedicalRecordRepo(), &mockInventoryRepository{}, nil)
	_, err := svc.Update(context.Background(), clinicID, mrID, trID, &UpdateTreatmentInput{
		DiscountAmount:      &zero,
		DiscountEditAllowed: false,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden), "got: %v", err)
}

func TestTreatmentRepository_LockByIDForUpdate_RequiresAmbientTransaction(t *testing.T) {
	db := setupTreatmentDiscountTOCTOUDB(t)
	repo := NewTreatmentRepository(db)
	_, err := repo.LockByIDForUpdate(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}
