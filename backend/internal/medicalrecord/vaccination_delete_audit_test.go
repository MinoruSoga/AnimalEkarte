package medicalrecord

// EMR-323 RED tests: vaccination Delete must write a fail-closed audit_logs
// entry via AuditTxLogger.LogEntryTx inside the SAME ambient transaction that
// removes the row (P323-AUDIT — slog alone does not satisfy the ticket). The
// audit dependency is the new required final argument of NewVaccinationService;
// the current unsafe Delete performs no audit at all, so these tests fail
// until GREEN. Tag resync may only run after a successful commit.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// A successful delete on a draft-linked vaccination must emit one audit entry
// inside the ambient transaction, carrying action/resource/clinic/resource-id,
// the typed old snapshot, and the authenticated staff actor propagated through
// withVaccinationAuditActor.
func TestVaccinationService_DeleteAudit_WritesEntryInsideTransaction(t *testing.T) {
	const (
		clinicID, vaccinationID, recordID uint64 = 1, 10, 30
		petID, ownerID, vaccineID         uint64 = 20, 40, 7
		doctorID, staffID                 uint64 = 5, 42
	)
	vaccDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	row := func(id uint64) *model.Vaccination {
		return &model.Vaccination{
			ID: id, ClinicID: clinicID, MedicalRecordID: ptrUint64(recordID),
			PetID: ptrUint64(petID), VaccineID: vaccineID,
			DoctorID: ptrUint64(doctorID), Date: vaccDate,
		}
	}

	var entry *AuditEntry
	var auditInTx bool
	audit := &mockAuditTxLogger{logEntryTxFn: func(ctx context.Context, e *AuditEntry) error {
		entry = e
		auditInTx = vaccinationGuardInTx(ctx)
		return nil
	}}
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) { return row(id), nil },
		lockByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) { return row(id), nil },
		deleteFn:   func(_ context.Context, _, _ uint64) error { return nil },
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: draftGuardRecord(recordID, clinicID),
	}}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationGuardMarkingTransactor(), audit)

	err := svc.Delete(withVaccinationAuditActor(context.Background(), staffID), clinicID, vaccinationID)

	require.NoError(t, err)
	require.NotNil(t, entry, "delete must write one audit entry via LogEntryTx")
	assert.True(t, auditInTx, "audit entry must be written inside the ambient transaction")
	assert.Equal(t, "delete", entry.Action)
	assert.Equal(t, "vaccination", entry.Resource)
	require.NotNil(t, entry.ClinicID)
	assert.Equal(t, clinicID, *entry.ClinicID)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, vaccinationID, *entry.ResourceID)
	require.NotNil(t, entry.ActorID, "authenticated staff id must become the audit actor")
	assert.Equal(t, staffID, *entry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	assert.Equal(t, vaccinationDeleteAuditSnapshot{
		ID:              vaccinationID,
		MedicalRecordID: ptrUint64(recordID),
		PetID:           ptrUint64(petID),
		VaccineID:       vaccineID,
		DoctorID:        ptrUint64(doctorID),
		Date:            vaccDate,
	}, entry.OldValue, "old_value must be the typed minimal snapshot of the deleted row")
}

// A direct internal call (no withVaccinationAuditActor on the ctx) must audit
// with a system actor — the actor must never be invented or taken from the
// request/body.
func TestVaccinationService_DeleteAudit_SystemActorWithoutAuthenticatedContext(t *testing.T) {
	const clinicID, vaccinationID uint64 = 1, 10

	var entry *AuditEntry
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, e *AuditEntry) error {
		entry = e
		return nil
	}}
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, VaccineID: 1}, nil
		},
		lockByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, VaccineID: 1}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, nil, vaccinationTestTransactor{}, audit)

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Nil(t, entry.ActorID, "no authenticated actor must not be fabricated")
	assert.Equal(t, model.AuditActorTypeSystem, entry.ActorType)
}

// A nil AuditTxLogger dependency must fail closed before any write — a missing
// audit sink must never silently skip auditability.
func TestVaccinationService_DeleteAudit_MissingLoggerRefusesDelete(t *testing.T) {
	const clinicID, vaccinationID uint64 = 1, 10
	deleteCalls := 0
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, VaccineID: 1}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { deleteCalls++; return nil },
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{}, nil, vaccinationTestTransactor{}, nil)

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.Error(t, err, "missing audit dependency must fail closed")
	assert.Zero(t, deleteCalls, "missing audit must not reach the repository delete")
}

// A LogEntryTx failure inside the transaction must surface as a Delete error
// and skip the post-commit tag resync. (Mock-level rollback is invisible; the
// real-DB rollback case is covered by DeleteAudit_RealDBAuditFailure below.)
func TestVaccinationService_DeleteAudit_LoggerFailureAbortsDeleteAndSkipsTags(t *testing.T) {
	const clinicID, vaccinationID, recordID, petID, ownerID uint64 = 1, 10, 30, 20, 40
	auditErr := errors.New("audit write failed")
	resyncCalls := 0

	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return ownerLoadedVaccination(vaccinationID, clinicID, petID, ownerID, ptrUint64(recordID)), nil
		},
		lockByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return ownerLoadedVaccination(vaccinationID, clinicID, petID, ownerID, ptrUint64(recordID)), nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: draftGuardRecord(recordID, clinicID),
	}}
	tagSync := &mockLstepTagSyncService{resyncOwnerVaccineTagsFn: func(context.Context, uint64, uint64) error {
		resyncCalls++
		return nil
	}}
	audit := &mockAuditTxLogger{logEntryTxFn: func(context.Context, *AuditEntry) error { return auditErr }}
	svc := NewVaccinationService(repo, okVaccineRepo(), tagSync,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, audit)

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.Error(t, err, "audit failure must abort the delete")
	assert.ErrorIs(t, err, auditErr, "audit failure must propagate wrapped")
	assert.Zero(t, resyncCalls, "tag resync must not run when the transaction failed")
}

// A commit-time failure (fn succeeded, WithTx returns error) must likewise skip
// the post-commit tag resync.
func TestVaccinationService_DeleteAudit_CommitFailureSkipsTagResync(t *testing.T) {
	const clinicID, vaccinationID, recordID, petID, ownerID uint64 = 1, 10, 30, 20, 40
	commitErr := errors.New("commit failed")
	resyncCalls := 0

	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return ownerLoadedVaccination(vaccinationID, clinicID, petID, ownerID, ptrUint64(recordID)), nil
		},
		lockByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return ownerLoadedVaccination(vaccinationID, clinicID, petID, ownerID, ptrUint64(recordID)), nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: draftGuardRecord(recordID, clinicID),
	}}
	tagSync := &mockLstepTagSyncService{resyncOwnerVaccineTagsFn: func(context.Context, uint64, uint64) error {
		resyncCalls++
		return nil
	}}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		if err := fn(ctx); err != nil {
			return err
		}
		return commitErr
	}}
	svc := NewVaccinationService(repo, okVaccineRepo(), tagSync,
		&permissiveVaccinationRelationVerifier{}, locker, tx, &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.Error(t, err)
	assert.ErrorIs(t, err, commitErr)
	assert.Zero(t, resyncCalls, "tag resync may only run after a successful commit")
}

// Real synthetic PostgreSQL: a failing audit logger inside the transaction must
// roll back the vaccination delete so the row survives.
func TestVaccinationService_DeleteAudit_RealDBAuditFailureRollsBackDelete(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicA)

	owner := makeTestOwner(t, db, clinicA, "audit rollback owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "audit rollback pet")
	vaccine := makeVaccineMaster(t, db, clinicA, "audit rollback vaccine")
	realRepo := NewVaccinationRepository(db)
	record := &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "VAC-AUDIT-RB", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID, Status: model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.WithContext(ctx).Create(record).Error)
	vaccination := &model.Vaccination{
		ClinicID: clinicA, PetID: &pet.ID, MedicalRecordID: &record.ID,
		VaccineID: vaccine.ID, Date: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(vaccination).Error)

	auditCalls := 0
	audit := &mockAuditTxLogger{logEntryTxFn: func(context.Context, *AuditEntry) error {
		auditCalls++
		return errors.New("audit write failed")
	}}
	svc := NewVaccinationService(
		realRepo,
		NewVaccineRepository(db),
		nil,
		reservation.NewReservationRepository(db),
		NewMedicalRecordRepository(db),
		vaccinationTestTransactor{db: db},
		audit,
	)

	err := svc.Delete(ctx, clinicA, vaccination.ID)

	require.Error(t, err, "audit failure must abort the delete")
	assert.Equal(t, 1, auditCalls, "audit must be attempted inside the delete transaction")
	still, findErr := realRepo.FindByID(ctx, clinicA, vaccination.ID)
	require.NoError(t, findErr, "audit failure must roll back the vaccination delete")
	require.NotNil(t, still)
}

// A delete rejected by the finalized-parent guard must produce no audit entry,
// no repository delete, and no tag resync — rejection means zero side effects.
func TestVaccinationService_DeleteAudit_FinalizedRejectionWritesNothing(t *testing.T) {
	const clinicID, vaccinationID, recordID, petID, ownerID uint64 = 1, 10, 30, 20, 40
	auditCalls, deleteCalls, resyncCalls := 0, 0, 0

	audit := &mockAuditTxLogger{logEntryTxFn: func(context.Context, *AuditEntry) error {
		auditCalls++
		return nil
	}}
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return ownerLoadedVaccination(vaccinationID, clinicID, petID, ownerID, ptrUint64(recordID)), nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { deleteCalls++; return nil },
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: finalizedGuardRecord(recordID, clinicID),
	}}
	tagSync := &mockLstepTagSyncService{resyncOwnerVaccineTagsFn: func(context.Context, uint64, uint64) error {
		resyncCalls++
		return nil
	}}
	svc := NewVaccinationService(repo, okVaccineRepo(), tagSync,
		&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, audit)

	err := svc.Delete(context.Background(), clinicID, vaccinationID)

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "finalized parent must map to conflict: %v", err)
	assert.Zero(t, auditCalls, "a rejected delete must write no audit entry")
	assert.Zero(t, deleteCalls)
	assert.Zero(t, resyncCalls)
}

// A nil or foreign locked child row must fail closed before the delete and the
// audit write — the locked row is the authoritative copy and an identity/clinic
// mismatch means the snapshot is stale.
func TestVaccinationService_DeleteAudit_NilOrForeignLockedChildRejected(t *testing.T) {
	const clinicID, vaccinationID, recordID uint64 = 1, 10, 30

	tests := []struct {
		name     string
		lockByID func(_ context.Context, _, id uint64) (*model.Vaccination, error)
	}{
		{
			name: "nil_locked_row",
			lockByID: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
				return nil, nil
			},
		},
		{
			name: "foreign_clinic_locked_row",
			lockByID: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
				return &model.Vaccination{
					ID: id, ClinicID: 2, MedicalRecordID: ptrUint64(recordID), VaccineID: 1,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditCalls, deleteCalls := 0, 0
			audit := &mockAuditTxLogger{logEntryTxFn: func(context.Context, *AuditEntry) error {
				auditCalls++
				return nil
			}}
			repo := &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
					return &model.Vaccination{
						ID: id, ClinicID: clinicID, MedicalRecordID: ptrUint64(recordID), VaccineID: 1,
					}, nil
				},
				lockByIDFn: tt.lockByID,
				deleteFn:   func(_ context.Context, _, _ uint64) error { deleteCalls++; return nil },
			}
			locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
				recordID: draftGuardRecord(recordID, clinicID),
			}}
			svc := NewVaccinationService(repo, okVaccineRepo(), nil,
				&permissiveVaccinationRelationVerifier{}, locker, vaccinationTestTransactor{}, audit)

			err := svc.Delete(context.Background(), clinicID, vaccinationID)

			require.Error(t, err, "nil or foreign locked child must fail closed")
			assert.Zero(t, deleteCalls)
			assert.Zero(t, auditCalls, "a rejected delete must write no audit entry")
		})
	}
}
