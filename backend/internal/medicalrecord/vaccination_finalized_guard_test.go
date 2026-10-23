package medicalrecord

// EMR-323 tests: vaccination Delete/Create/Update hold the parent medical-record
// row lock inside the write transaction (ascending order for source+destination
// relink). EMR-495 P1: the old finalized-parent rejection (and its tests) was retired
// together with medical_records.status; the lock, ownership and relink protections remain.

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// vaccinationFinalizedGuardTxKey marks the ambient-transaction context so the
// mocks below can distinguish pre-transaction reads from in-transaction work.
type vaccinationFinalizedGuardTxKey struct{}

func vaccinationGuardMarkingTransactor() *mockTransactor {
	return &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		return fn(context.WithValue(ctx, vaccinationFinalizedGuardTxKey{}, true))
	}}
}

func vaccinationGuardInTx(ctx context.Context) bool {
	marked, _ := ctx.Value(vaccinationFinalizedGuardTxKey{}).(bool)
	return marked
}

// recordingVaccinationRecordLocker wraps vaccinationMedicalRecordLockerStub to
// capture parent lock order and whether each lock ran inside the ambient tx.
type recordingVaccinationRecordLocker struct {
	records map[uint64]*model.MedicalRecord
	calls   []uint64
	inTx    []bool
}

func (l *recordingVaccinationRecordLocker) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	l.calls = append(l.calls, id)
	l.inTx = append(l.inTx, vaccinationGuardInTx(ctx))
	stub := vaccinationMedicalRecordLockerStub{records: l.records}
	return stub.LockByIDForUpdate(ctx, clinicID, id)
}

// opsRecordingVaccinationRecordLocker appends "parent_lock" events into a
// shared op timeline so tests can assert parent-before-child ordering across
// the locker and repository mocks.
type opsRecordingVaccinationRecordLocker struct {
	records map[uint64]*model.MedicalRecord
	ops     *[]vaccinationGuardOp
}

func (l *opsRecordingVaccinationRecordLocker) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	*l.ops = append(*l.ops, vaccinationGuardOp{"parent_lock", vaccinationGuardInTx(ctx)})
	stub := vaccinationMedicalRecordLockerStub{records: l.records}
	return stub.LockByIDForUpdate(ctx, clinicID, id)
}

// vaccinationGuardOp records a repository/locker operation and whether it ran
// inside the ambient transaction.
type vaccinationGuardOp struct {
	name string
	inTx bool
}

func vaccinationGuardOpNames(ops []vaccinationGuardOp) []string {
	names := make([]string, 0, len(ops))
	for _, op := range ops {
		if op.name == "find" {
			continue
		}
		names = append(names, op.name)
	}
	return names
}

// distinctVaccinationLockOrder returns parent lock IDs in first-occurrence
// order so tests can assert the deterministic ascending lock order Lead
// requires for source+destination relink.
func distinctVaccinationLockOrder(calls []uint64) []uint64 {
	seen := map[uint64]bool{}
	out := make([]uint64, 0, len(calls))
	for _, id := range calls {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func draftGuardRecord(id, clinicID uint64) *model.MedicalRecord {
	return &model.MedicalRecord{ID: id, ClinicID: clinicID}
}

// ownerLoadedVaccination builds a vaccination whose Pet/Owner preloads satisfy
// safeVaccinationOwnerID for the given clinic/owner pair.
func ownerLoadedVaccination(id, clinicID, petID, ownerID uint64, medicalRecordID *uint64) *model.Vaccination {
	return &model.Vaccination{
		ID:              id,
		ClinicID:        clinicID,
		PetID:           &petID,
		MedicalRecordID: medicalRecordID,
		VaccineID:       1,
		Pet: &model.Pet{
			ID: petID, ClinicID: clinicID, OwnerID: ownerID,
			Owner: &model.Owner{ID: ownerID, ClinicID: clinicID},
		},
	}
}

// A linked parent that is missing/foreign under the caller clinic must fail
// closed, not silently delete the child row.
func TestVaccinationService_Delete_CrossClinicParentRejected(t *testing.T) {
	const clinicID, vaccinationID, recordID uint64 = 1, 10, 30
	deleteCalls := 0
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, MedicalRecordID: ptrUint64(recordID), VaccineID: 1}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalls++
			return nil
		},
	}
	// records map has no recordID: LockByIDForUpdate under the caller clinic fails.
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{}}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.Error(t, err, "cross-clinic or missing parent must fail closed")
	assert.Zero(t, deleteCalls)
}

// After the child row is locked, a medical_record_id that no longer matches the
// validated snapshot indicates a concurrent move; the delete must be rejected.
func TestVaccinationService_Delete_ConcurrentParentRelinkRejected(t *testing.T) {
	const clinicID, vaccinationID, recordID uint64 = 1, 10, 30
	deleteCalls := 0
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, MedicalRecordID: ptrUint64(recordID), VaccineID: 1}, nil
		},
		lockByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, MedicalRecordID: ptrUint64(99), VaccineID: 1}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalls++
			return nil
		},
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: draftGuardRecord(recordID, clinicID),
		99:       draftGuardRecord(99, clinicID),
	}}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.Error(t, err, "a parent changed under us must be rejected")
	assert.True(t, apperrors.IsConflict(err), "concurrent parent change must map to conflict: %v", err)
	assert.Zero(t, deleteCalls)
}

// The parent medical record must be locked FOR UPDATE before the child row,
// and every mutating step must run inside the write transaction.
func TestVaccinationService_Delete_LocksParentBeforeChildInTransaction(t *testing.T) {
	const clinicID, vaccinationID, recordID uint64 = 1, 10, 30
	ops := make([]vaccinationGuardOp, 0, 4)
	row := &model.Vaccination{ID: vaccinationID, ClinicID: clinicID, MedicalRecordID: ptrUint64(recordID), VaccineID: 1}
	repo := &mockVaccinationRepository{
		findByIDFn: func(ctx context.Context, _, _ uint64) (*model.Vaccination, error) {
			ops = append(ops, vaccinationGuardOp{"find", vaccinationGuardInTx(ctx)})
			return row, nil
		},
		lockByIDFn: func(ctx context.Context, _, _ uint64) (*model.Vaccination, error) {
			ops = append(ops, vaccinationGuardOp{"child_lock", vaccinationGuardInTx(ctx)})
			return row, nil
		},
		deleteFn: func(ctx context.Context, _, _ uint64) error {
			ops = append(ops, vaccinationGuardOp{"delete", vaccinationGuardInTx(ctx)})
			return nil
		},
	}
	locker := &opsRecordingVaccinationRecordLocker{
		records: map[uint64]*model.MedicalRecord{recordID: draftGuardRecord(recordID, clinicID)},
		ops:     &ops,
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationGuardMarkingTransactor(), &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), clinicID, vaccinationID)
	require.NoError(t, err)

	names := vaccinationGuardOpNames(ops)
	require.Equal(t, []string{"parent_lock", "child_lock", "delete"}, names,
		"delete must lock the parent record before the child row")
	for _, op := range ops {
		if op.name == "find" {
			continue
		}
		assert.True(t, op.inTx, "%s must run inside the write transaction", op.name)
	}
}

// Delete must fail closed when required dependencies are absent and must reject
// a nil row returned without error.
func TestVaccinationService_Delete_FailsClosedOnMissingDependencies(t *testing.T) {
	const clinicID, vaccinationID, recordID uint64 = 1, 10, 30
	linkedRow := &model.Vaccination{ID: vaccinationID, ClinicID: clinicID, MedicalRecordID: ptrUint64(recordID), VaccineID: 1}
	draftLocker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: draftGuardRecord(recordID, clinicID),
	}}

	t.Run("nil_transactor", func(t *testing.T) {
		deleteCalls := 0
		repo := &mockVaccinationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) { return linkedRow, nil },
			deleteFn:   func(_ context.Context, _, _ uint64) error { deleteCalls++; return nil },
		}
		svc := NewVaccinationService(repo, okVaccineRepo(), nil,
			&permissiveVaccinationRelationVerifier{}, draftLocker, nil, &mockAuditTxLogger{})

		err := svc.Delete(context.Background(), clinicID, vaccinationID)

		require.Error(t, err, "delete without a transactor must fail closed")
		assert.Zero(t, deleteCalls)
	})

	t.Run("nil_parent_locker_with_linked_record", func(t *testing.T) {
		deleteCalls := 0
		repo := &mockVaccinationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) { return linkedRow, nil },
			deleteFn:   func(_ context.Context, _, _ uint64) error { deleteCalls++; return nil },
		}
		svc := NewVaccinationService(repo, okVaccineRepo(), nil,
			&permissiveVaccinationRelationVerifier{}, nil, vaccinationTestTransactor{}, &mockAuditTxLogger{})

		err := svc.Delete(context.Background(), clinicID, vaccinationID)

		require.Error(t, err, "delete of a linked row without a parent locker must fail closed")
		assert.Zero(t, deleteCalls)
	})

	t.Run("nil_row_from_find", func(t *testing.T) {
		deleteCalls := 0
		repo := &mockVaccinationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) { return nil, nil },
			deleteFn:   func(_ context.Context, _, _ uint64) error { deleteCalls++; return nil },
		}
		svc := NewVaccinationService(repo, okVaccineRepo(), nil,
			&permissiveVaccinationRelationVerifier{}, draftLocker, vaccinationTestTransactor{}, &mockAuditTxLogger{})

		err := svc.Delete(context.Background(), clinicID, vaccinationID)

		require.Error(t, err, "a nil row without error must fail closed")
		assert.Zero(t, deleteCalls)
	})
}

// Deleting a vaccination linked to a DRAFT record stays allowed.
func TestVaccinationService_Delete_AllowsDraftLinkedRecord(t *testing.T) {
	const clinicID, vaccinationID, recordID, petID, ownerID uint64 = 1, 10, 30, 20, 40
	deleteCalls := 0
	resyncCalls := 0
	row := ownerLoadedVaccination(vaccinationID, clinicID, petID, ownerID, ptrUint64(recordID))
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) { return row, nil },
		deleteFn:   func(_ context.Context, _, _ uint64) error { deleteCalls++; return nil },
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: draftGuardRecord(recordID, clinicID),
	}}
	tagSync := &mockLstepTagSyncService{
		resyncOwnerVaccineTagsFn: func(_ context.Context, _, syncedOwnerID uint64) error {
			resyncCalls++
			assert.Equal(t, ownerID, syncedOwnerID)
			return nil
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), tagSync,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.NoError(t, err)
	assert.Equal(t, 1, deleteCalls)
	assert.Equal(t, 1, resyncCalls)
}

// An unlinked vaccination must delete without consulting the parent locker.
func TestVaccinationService_Delete_AllowsUnlinkedRecord(t *testing.T) {
	const clinicID, vaccinationID, petID, ownerID uint64 = 1, 10, 20, 40
	deleteCalls := 0
	row := ownerLoadedVaccination(vaccinationID, clinicID, petID, ownerID, nil)
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) { return row, nil },
		deleteFn:   func(_ context.Context, _, _ uint64) error { deleteCalls++; return nil },
	}
	locker := &recordingVaccinationRecordLocker{records: map[uint64]*model.MedicalRecord{}}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.NoError(t, err)
	assert.Equal(t, 1, deleteCalls)
	assert.Empty(t, locker.calls, "unlinked delete must not lock any parent record")
}

// Tag resync must derive the owner from the CURRENT record read inside the
// transaction, not from the stale pre-transaction snapshot. The pre-tx FindByID
// returns a row with no owner preloads; only the in-transaction view carries
// owner 40.
func TestVaccinationService_Delete_ResyncsTagsFromCurrentLockedRecord(t *testing.T) {
	const clinicID, vaccinationID, recordID, petID, ownerID uint64 = 1, 10, 30, 20, 40
	resynced := uint64(0)
	repo := &mockVaccinationRepository{
		findByIDFn: func(ctx context.Context, _, id uint64) (*model.Vaccination, error) {
			if vaccinationGuardInTx(ctx) {
				return ownerLoadedVaccination(id, clinicID, petID, ownerID, ptrUint64(recordID)), nil
			}
			return &model.Vaccination{ID: id, ClinicID: clinicID, PetID: ptrUint64(petID),
				MedicalRecordID: ptrUint64(recordID), VaccineID: 1}, nil
		},
		lockByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, PetID: ptrUint64(petID),
				MedicalRecordID: ptrUint64(recordID), VaccineID: 1}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: draftGuardRecord(recordID, clinicID),
	}}
	tagSync := &mockLstepTagSyncService{
		resyncOwnerVaccineTagsFn: func(_ context.Context, _, syncedOwnerID uint64) error {
			resynced = syncedOwnerID
			return nil
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), tagSync,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationGuardMarkingTransactor(), &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.NoError(t, err)
	assert.Equal(t, ownerID, resynced, "resync must use the current in-transaction owner, not the stale snapshot")
}

// Creating a vaccination linked to a draft record stays allowed.
func TestVaccinationService_Create_AllowsDraftLinkedRecord(t *testing.T) {
	const clinicID, recordID uint64 = 1, 30
	createCalls := 0
	repo := &mockVaccinationRepository{
		createFn: func(_ context.Context, vaccination *model.Vaccination) error {
			createCalls++
			vaccination.ID = 10
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, MedicalRecordID: ptrUint64(recordID), VaccineID: 1}, nil
		},
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: draftGuardRecord(recordID, clinicID),
	}}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, &mockAuditTxLogger{})

	got, err := svc.Create(context.Background(), clinicID, &CreateVaccinationInput{
		MedicalRecordID: ptrUint64(recordID), VaccineID: 1, Date: time.Now(),
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, createCalls)
}

// Comparing the post-input "effective" relations lets a concurrent move hide
// behind the caller's own override. The locked row's ORIGINAL relations must be
// compared against the snapshot instead.
func TestVaccinationService_Update_HiddenConcurrentRelinkRejected(t *testing.T) {
	const clinicID, vaccinationID, sourceID, destID uint64 = 1, 10, 30, 40
	updateCalls := 0
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, MedicalRecordID: ptrUint64(sourceID), VaccineID: 1}, nil
		},
		lockByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			// Concurrent writer moved the row to record 99 between snapshot and lock.
			return &model.Vaccination{ID: id, ClinicID: clinicID, MedicalRecordID: ptrUint64(99), VaccineID: 1}, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
			updateCalls++
			return &model.Vaccination{ID: id}, nil
		},
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		sourceID: draftGuardRecord(sourceID, clinicID),
		destID:   draftGuardRecord(destID, clinicID),
	}}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, &mockAuditTxLogger{})

	got, err := svc.Update(context.Background(), clinicID, vaccinationID,
		&UpdateVaccinationInput{MedicalRecordID: ptrUint64(destID)})

	require.Error(t, err, "a concurrent relink must not hide behind the caller's own override")
	assert.True(t, apperrors.IsConflict(err), "concurrent move must map to conflict: %v", err)
	assert.Nil(t, got)
	assert.Zero(t, updateCalls)
}

// When a relink touches two parent records, all parent locks must be taken in a
// deterministic ascending-ID order regardless of which is source/destination.
func TestVaccinationService_Update_RelinkParentsLockedAscending(t *testing.T) {
	const clinicID, vaccinationID uint64 = 1, 10
	remarks := "relink"

	run := func(t *testing.T, sourceID, destID uint64) []uint64 {
		locker := &recordingVaccinationRecordLocker{records: map[uint64]*model.MedicalRecord{
			sourceID: draftGuardRecord(sourceID, clinicID),
			destID:   draftGuardRecord(destID, clinicID),
		}}
		repo := &mockVaccinationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
				return &model.Vaccination{ID: id, ClinicID: clinicID, MedicalRecordID: ptrUint64(sourceID), VaccineID: 1}, nil
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
				return &model.Vaccination{ID: id}, nil
			},
		}
		svc := NewVaccinationService(repo, okVaccineRepo(), nil,
			&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, &mockAuditTxLogger{})

		_, err := svc.Update(context.Background(), clinicID, vaccinationID,
			&UpdateVaccinationInput{MedicalRecordID: ptrUint64(destID), Remarks: &remarks})
		require.NoError(t, err)
		return distinctVaccinationLockOrder(locker.calls)
	}

	t.Run("source_below_destination", func(t *testing.T) {
		assert.Equal(t, []uint64{30, 50}, run(t, 30, 50))
	})
	t.Run("source_above_destination", func(t *testing.T) {
		assert.Equal(t, []uint64{30, 50}, run(t, 50, 30))
	})
}
