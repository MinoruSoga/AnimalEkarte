package medicalrecord

// EMR-323 GREEN evidence on real synthetic PostgreSQL (independent
// transactions, no mocks for the lock behaviour under test):
//
//  1. Delete-first: the delete transaction holding the parent FOR UPDATE lock
//     forces a concurrent parent header UPDATE to wait until commit — the same
//     X-11 serialization medical_record_finalize_lock_concurrency_test.go proves
//     for treatment writes.
//  2. (EMR-495 P1: the "finalize-first rejects Delete" direction was retired
//     with medical_records.status; its test was removed.)
//  3. A concurrent relink committed between the pre-transaction snapshot and
//     the child row lock is caught by the ORIGINAL-snapshot revalidation and
//     rejects with Conflict — the stale deletion is never performed.
//  4. A successful delete writes exactly one real audit_logs row through the
//     shared audit kernel (test-only adapter mirroring the composition-root
//     medicalRecordAuditTxBridge) inside the same transaction.
//  5. Opposite-direction relinks over the same {lower, higher} parent set run
//     as two genuinely concurrent transactions: the second queues on the
//     contested first parent, then both complete without deadlock —
//     deterministic ascending lock order under real competition.
//  6. Clinic-boundary fail-closed evidence: a foreign child row, a foreign
//     parent link, and a corrupt foreign pet reference all reject before any
//     lock, delete, audit, or tag side effect, leaving the row intact.

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// vaccinationTestAuditTxBridge mirrors cmd/api's medicalRecordAuditTxBridge:
// it adapts the shared audit kernel's LogEntryTx to the package-local
// AuditTxLogger interface so the real audit path is exercised end to end.
type vaccinationTestAuditTxBridge struct {
	logger audit.TxLogger
}

func (b vaccinationTestAuditTxBridge) LogEntryTx(ctx context.Context, entry *AuditEntry) error {
	if b.logger == nil {
		return errors.New("vaccination test audit logger is required")
	}
	return b.logger.LogEntryTx(ctx, &audit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		Metadata:   entry.Metadata,
	})
}

// gatingVaccinationParentLocker wraps the real medical-record repository and
// pauses while still holding the parent FOR UPDATE lock inside Delete's write
// transaction, so a concurrent parent write provably blocks on the same row lock.
type gatingVaccinationParentLocker struct {
	inner    medicalRecordLocker
	locked   chan struct{}
	proceed  chan struct{}
	attempts chan uint64
	once     sync.Once
}

func (l *gatingVaccinationParentLocker) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	if l.attempts != nil {
		select {
		case l.attempts <- id:
		default:
		}
	}
	record, err := l.inner.LockByIDForUpdate(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	l.once.Do(func() { close(l.locked) })
	select {
	case <-l.proceed:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return record, nil
}

// relinkBeforeVaccinationChildLockRepository simulates a concurrent relink that
// commits between Delete's pre-transaction snapshot and its child row lock: the
// external UPDATE commits on a separate connection before the real scoped lock
// returns the moved row.
type relinkBeforeVaccinationChildLockRepository struct {
	VaccinationRepository
	db          *gorm.DB
	newRecordID uint64
	once        sync.Once
	relinkErr   error
}

func (r *relinkBeforeVaccinationChildLockRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	r.once.Do(func() {
		r.relinkErr = r.db.WithContext(context.Background()).
			Model(&model.Vaccination{}).
			Where("clinic_id = ? AND id = ?", clinicID, id).
			Update("medical_record_id", r.newRecordID).Error
	})
	if r.relinkErr != nil {
		return nil, r.relinkErr
	}
	return r.VaccinationRepository.LockByIDForUpdate(ctx, clinicID, id)
}

func newRealVaccinationServiceForConcurrency(db *gorm.DB, repo VaccinationRepository, locker medicalRecordLocker, auditTx AuditTxLogger) VaccinationService {
	return NewVaccinationService(
		repo,
		NewVaccineRepository(db),
		nil,
		reservation.NewReservationRepository(db),
		locker,
		vaccinationTestTransactor{db: db},
		auditTx,
	)
}

// Direction 1: Delete's transaction holds the parent FOR UPDATE lock while a
// concurrent parent header UPDATE must wait; once Delete commits, the UPDATE succeeds.
func TestVaccinationService_Delete_RealDBParentLockBlocksParentWriteUntilCommit(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicA)

	owner := makeTestOwner(t, db, clinicA, "delete-lock owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "delete-lock pet")
	vaccine := makeVaccineMaster(t, db, clinicA, "delete-lock vaccine")
	medRecRepo := NewMedicalRecordRepository(db)
	realRepo := NewVaccinationRepository(db)

	record := &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "VAC-DEL-LOCK", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(record).Error)
	vaccination := &model.Vaccination{
		ClinicID: clinicA, PetID: &pet.ID, MedicalRecordID: &record.ID,
		VaccineID: vaccine.ID, Date: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(vaccination).Error)

	gate := &gatingVaccinationParentLocker{
		inner:   medRecRepo,
		locked:  make(chan struct{}),
		proceed: make(chan struct{}),
	}
	svc := newRealVaccinationServiceForConcurrency(db, realRepo, gate, &mockAuditTxLogger{})

	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(gate.proceed) }) }
	defer release()

	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- svc.Delete(deleteCtx, clinicA, vaccination.ID)
	}()

	select {
	case <-gate.locked:
	case err := <-deleteDone:
		t.Fatalf("delete transaction failed before acquiring the parent lock: %v", err)
	case <-deleteCtx.Done():
		t.Fatal("delete transaction never acquired the parent lock")
	}

	finCtx, finCancel := context.WithTimeout(ctx, 30*time.Second)
	defer finCancel()
	parentWriteDone := make(chan error, 1)
	go func() {
		_, err := medRecRepo.Update(finCtx, clinicA, record.ID,
			medicalRecordParentWrite(), nil)
		parentWriteDone <- err
	}()

	// Real lock evidence: the parent header UPDATE cannot complete while the delete
	// transaction still holds the parent row lock.
	select {
	case err := <-parentWriteDone:
		t.Fatalf("parent write completed while the delete tx held the parent lock: %v", err)
	case <-time.After(300 * time.Millisecond):
	}

	release()
	select {
	case err := <-deleteDone:
		require.NoError(t, err, "delete must commit once released")
	case <-deleteCtx.Done():
		t.Fatal("delete transaction never completed after release")
	}
	select {
	case err := <-parentWriteDone:
		require.NoError(t, err, "parent write must succeed after the delete commit released the lock")
	case <-finCtx.Done():
		t.Fatal("parent write never completed")
	}

	_, findErr := realRepo.FindByID(ctx, clinicA, vaccination.ID)
	assert.True(t, apperrors.IsNotFound(findErr), "committed delete must remove the row")
	final, err := medRecRepo.FindByID(ctx, clinicA, record.ID)
	require.NoError(t, err)
	require.NotNil(t, final.VisitType)
	assert.Equal(t, model.VisitTypeRevisit, *final.VisitType)
}

// A relink committed by another transaction between the snapshot read and the
// child lock must be caught by ORIGINAL-snapshot revalidation: the stale
// deletion is rejected and the moved row survives.
func TestVaccinationService_Delete_RealDBCommittedRelinkBetweenSnapshotAndLockRejected(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicA)

	owner := makeTestOwner(t, db, clinicA, "relink owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "relink pet")
	vaccine := makeVaccineMaster(t, db, clinicA, "relink vaccine")
	medRecRepo := NewMedicalRecordRepository(db)
	realRepo := NewVaccinationRepository(db)

	source := &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "VAC-RELINK-S", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(source).Error)
	dest := &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "VAC-RELINK-D", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(dest).Error)
	vaccination := &model.Vaccination{
		ClinicID: clinicA, PetID: &pet.ID, MedicalRecordID: &source.ID,
		VaccineID: vaccine.ID, Date: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(vaccination).Error)

	repo := &relinkBeforeVaccinationChildLockRepository{
		VaccinationRepository: realRepo,
		db:                    db,
		newRecordID:           dest.ID,
	}
	svc := newRealVaccinationServiceForConcurrency(db, repo, medRecRepo, &mockAuditTxLogger{})

	err := svc.Delete(ctx, clinicA, vaccination.ID)

	require.Error(t, err, "a relink committed under the snapshot must reject the stale deletion")
	assert.True(t, apperrors.IsConflict(err), "concurrent relink must map to conflict: %v", err)
	still, findErr := realRepo.FindByID(ctx, clinicA, vaccination.ID)
	require.NoError(t, findErr, "rejected delete must leave the row untouched")
	require.NotNil(t, still)
	require.NotNil(t, still.MedicalRecordID)
	assert.Equal(t, dest.ID, *still.MedicalRecordID, "the committed relink must be preserved, not deleted")
}

// A successful delete must persist exactly one audit_logs row through the real
// audit kernel, carrying the authenticated staff actor, clinic, resource
// identity, and the typed old-value snapshot — in the same transaction as the
// deletion.
func TestVaccinationService_Delete_RealDBAuditEntryPersisted(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AuditLog{}))
	ctx := context.Background()
	const clinicA = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicA)

	owner := makeTestOwner(t, db, clinicA, "audit owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "audit pet")
	vaccine := makeVaccineMaster(t, db, clinicA, "audit vaccine")
	staff := makeDoctor(t, db, clinicA, "audit actor")
	require.NoError(t, db.WithContext(ctx).
		Create(&model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA, IsMain: true}).Error)
	medRecRepo := NewMedicalRecordRepository(db)
	realRepo := NewVaccinationRepository(db)

	record := &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "VAC-AUDIT-OK", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(record).Error)
	vaccination := &model.Vaccination{
		ClinicID: clinicA, PetID: &pet.ID, MedicalRecordID: &record.ID,
		VaccineID: vaccine.ID, DoctorID: &staff.ID, Date: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(vaccination).Error)

	auditSvc := audit.NewService(audit.NewRepository(db))
	svc := newRealVaccinationServiceForConcurrency(
		db, realRepo, medRecRepo, vaccinationTestAuditTxBridge{logger: auditSvc},
	)
	var previousAuditID uint64
	require.NoError(t, db.WithContext(ctx).Model(&model.AuditLog{}).
		Select("COALESCE(MAX(id), 0)").Scan(&previousAuditID).Error)
	storedBeforeDelete, err := realRepo.FindByID(ctx, clinicA, vaccination.ID)
	require.NoError(t, err)

	err = svc.Delete(withVaccinationAuditActor(ctx, staff.ID), clinicA, vaccination.ID)
	require.NoError(t, err)

	_, findErr := realRepo.FindByID(ctx, clinicA, vaccination.ID)
	assert.True(t, apperrors.IsNotFound(findErr), "the delete itself must have committed")

	var logs []model.AuditLog
	require.NoError(t, db.WithContext(ctx).
		Where("id > ? AND resource = ? AND resource_id = ?", previousAuditID, "vaccination", vaccination.ID).
		Find(&logs).Error)
	require.Len(t, logs, 1, "exactly one audit entry must be persisted for the delete")

	entry := logs[0]
	require.NotNil(t, entry.ClinicID)
	assert.Equal(t, clinicA, *entry.ClinicID)
	require.NotNil(t, entry.ActorID, "authenticated staff must be the audit actor")
	assert.Equal(t, staff.ID, *entry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	assert.Equal(t, "delete", entry.Action)
	assert.Equal(t, "vaccination", entry.Resource)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, vaccination.ID, *entry.ResourceID)

	var snapshot vaccinationDeleteAuditSnapshot
	require.NoError(t, json.Unmarshal(entry.OldValue, &snapshot), "old_value must be the typed snapshot JSON")
	assert.Equal(t, vaccination.ID, snapshot.ID)
	assert.Equal(t, vaccine.ID, snapshot.VaccineID)
	assert.Equal(t, &record.ID, snapshot.MedicalRecordID)
	assert.Equal(t, &pet.ID, snapshot.PetID)
	assert.Equal(t, &staff.ID, snapshot.DoctorID)
	assert.Equal(t, storedBeforeDelete.Date, snapshot.Date)
}

// Opposite-direction relinks exercise the deterministic ascending parent lock
// order on real rows under genuine competition: two independent transactions
// relink different vaccinations across the same {lower, higher} parent set in
// opposite directions. Both want the lower parent first; the first acquirer is
// held at the gate while the second provably queues on the same row, then both
// complete without deadlock. This is the concurrent version of the X-11
// serialization evidence — not sequential subtests.
func TestVaccinationService_Update_RealDBRelinkLocksBothParentsWithoutDeadlock(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicA)

	owner := makeTestOwner(t, db, clinicA, "relink-order owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "relink-order pet")
	vaccine := makeVaccineMaster(t, db, clinicA, "relink-order vaccine")
	medRecRepo := NewMedicalRecordRepository(db)
	realRepo := NewVaccinationRepository(db)

	// Creation order gives lower < higher IDs deterministically.
	lower := &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "VAC-RELINK-LO", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(lower).Error)
	higher := &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "VAC-RELINK-HI", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(higher).Error)
	require.Less(t, lower.ID, higher.ID)

	vacOnHigher := &model.Vaccination{
		ClinicID: clinicA, PetID: &pet.ID, MedicalRecordID: &higher.ID,
		VaccineID: vaccine.ID, Date: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(vacOnHigher).Error)
	vacOnLower := &model.Vaccination{
		ClinicID: clinicA, PetID: &pet.ID, MedicalRecordID: &lower.ID,
		VaccineID: vaccine.ID, Date: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(vacOnLower).Error)

	gate := &gatingVaccinationParentLocker{
		inner:    medRecRepo,
		locked:   make(chan struct{}),
		proceed:  make(chan struct{}),
		attempts: make(chan uint64, 8),
	}
	svc := newRealVaccinationServiceForConcurrency(db, realRepo, gate, &mockAuditTxLogger{})

	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(gate.proceed) }) }
	defer release()

	type updateResult struct {
		got *model.Vaccination
		err error
	}
	highToLow := make(chan updateResult, 1)
	lowToHigh := make(chan updateResult, 1)

	// Tx A moves a vaccination off the higher-ID parent onto the lower one;
	// its sorted lock set {lower, higher} acquires `lower` first and pauses
	// at the gate while holding that row lock.
	go func() {
		got, err := svc.Update(runCtx, clinicA, vacOnHigher.ID,
			&UpdateVaccinationInput{MedicalRecordID: &lower.ID})
		highToLow <- updateResult{got: got, err: err}
	}()

	select {
	case <-gate.locked:
	case res := <-highToLow:
		t.Fatalf("first relink finished without reaching the parent gate: %v", res.err)
	case <-runCtx.Done():
		t.Fatal("first relink transaction never acquired the parent lock")
	}

	// Tx B performs the opposite-direction relink of a different vaccination
	// over the same parent set; it must queue on `lower` while tx A holds it.
	go func() {
		got, err := svc.Update(runCtx, clinicA, vacOnLower.ID,
			&UpdateVaccinationInput{MedicalRecordID: &higher.ID})
		lowToHigh <- updateResult{got: got, err: err}
	}()

	// Wait for the second `lower` lock attempt — the first is tx A's own.
	// Reaching it proves tx B is genuinely contending for the same first
	// parent while tx A still holds the row (not sequential subtests).
	lowerAttempts := 0
	for lowerAttempts < 2 {
		select {
		case id := <-gate.attempts:
			if id == lower.ID {
				lowerAttempts++
			}
		case res := <-lowToHigh:
			t.Fatalf("opposite-direction relink finished while the gate held the parent lock: %v", res.err)
		case <-runCtx.Done():
			t.Fatal("second relink never reached the contended parent lock")
		}
	}

	release()

	var highRes, lowRes updateResult
	select {
	case highRes = <-highToLow:
	case <-runCtx.Done():
		t.Fatal("high→low relink never completed after release")
	}
	select {
	case lowRes = <-lowToHigh:
	case <-runCtx.Done():
		t.Fatal("low→high relink never completed after release")
	}
	require.NoError(t, highRes.err, "high→low relink must not deadlock")
	require.NoError(t, lowRes.err, "low→high relink must not deadlock")
	require.NotNil(t, highRes.got)
	require.NotNil(t, lowRes.got)
	require.NotNil(t, highRes.got.MedicalRecordID)
	require.NotNil(t, lowRes.got.MedicalRecordID)
	assert.Equal(t, lower.ID, *highRes.got.MedicalRecordID)
	assert.Equal(t, higher.ID, *lowRes.got.MedicalRecordID)

	movedHigh, err := realRepo.FindByID(ctx, clinicA, vacOnHigher.ID)
	require.NoError(t, err)
	require.NotNil(t, movedHigh.MedicalRecordID)
	assert.Equal(t, lower.ID, *movedHigh.MedicalRecordID)
	movedLow, err := realRepo.FindByID(ctx, clinicA, vacOnLower.ID)
	require.NoError(t, err)
	require.NotNil(t, movedLow.MedicalRecordID)
	assert.Equal(t, higher.ID, *movedLow.MedicalRecordID)
}

// Clinic-boundary fail-closed evidence: rows the clinic-scoped repository
// cannot legitimately resolve — a foreign child row, a foreign parent link,
// or a corrupt foreign pet reference — must reject before any lock, delete,
// audit, or tag side effect. The scoped-repository boundary itself is the
// contract under test and stays unchanged.
func TestVaccinationService_Delete_RealDBForeignAndCorruptRowsRejected(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AuditLog{}))
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	ownerA := makeTestOwner(t, db, clinicA, "boundary owner A")
	petA := makeVaccinationRepoTestPet(t, db, clinicA, ownerA.ID, "boundary pet A")
	vaccineA := makeVaccineMaster(t, db, clinicA, "boundary vaccine A")
	ownerB := makeTestOwner(t, db, clinicB, "boundary owner B")
	petB := makeVaccinationRepoTestPet(t, db, clinicB, ownerB.ID, "boundary pet B")
	vaccineB := makeVaccineMaster(t, db, clinicB, "boundary vaccine B")

	recordA := &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "VAC-BND-A", Date: time.Now(),
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(recordA).Error)
	recordB := &model.MedicalRecord{
		ClinicID: clinicB, RecordNo: "VAC-BND-B", Date: time.Now(),
		OwnerID: &ownerB.ID, PetID: &petB.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(recordB).Error)

	auditSvc := audit.NewService(audit.NewRepository(db))
	synced := false
	svc := NewVaccinationService(
		NewVaccinationRepository(db),
		NewVaccineRepository(db),
		&mockLstepTagSyncService{resyncOwnerVaccineTagsFn: func(context.Context, uint64, uint64) error {
			synced = true
			return nil
		}},
		reservation.NewReservationRepository(db),
		NewMedicalRecordRepository(db),
		vaccinationTestTransactor{db: db},
		vaccinationTestAuditTxBridge{logger: auditSvc},
	)

	assertRejectedKeepsRow := func(t *testing.T, vaccination *model.Vaccination) {
		t.Helper()
		synced = false
		var auditCountBefore int64
		require.NoError(t, db.WithContext(ctx).Model(&model.AuditLog{}).
			Where("resource = ? AND resource_id = ?", "vaccination", vaccination.ID).
			Count(&auditCountBefore).Error)
		err := svc.Delete(ctx, clinicA, vaccination.ID)
		require.Error(t, err, "boundary-violating row must reject deletion")
		assert.True(t, apperrors.IsNotFound(err),
			"the scoped repository boundary rejects with NotFound: %v", err)

		// The row must physically remain — verified bypassing clinic scoping.
		var stored model.Vaccination
		require.NoError(t, db.WithContext(ctx).First(&stored, vaccination.ID).Error,
			"rejected delete must leave the row intact")

		var auditCount int64
		require.NoError(t, db.WithContext(ctx).Model(&model.AuditLog{}).
			Where("resource = ? AND resource_id = ?", "vaccination", vaccination.ID).
			Count(&auditCount).Error)
		assert.Equal(t, auditCountBefore, auditCount, "rejected delete must write no new audit entry")
		assert.False(t, synced, "rejected delete must not resync tags")
	}

	t.Run("foreign child row is invisible to the caller clinic", func(t *testing.T) {
		foreign := &model.Vaccination{
			ClinicID: clinicB, PetID: &petB.ID, VaccineID: vaccineB.ID, Date: time.Now(),
		}
		require.NoError(t, db.WithContext(ctx).Create(foreign).Error)
		assertRejectedKeepsRow(t, foreign)
	})

	t.Run("vaccination linked to a foreign medical record rejects", func(t *testing.T) {
		linked := &model.Vaccination{
			ClinicID: clinicA, PetID: &petA.ID, MedicalRecordID: &recordB.ID,
			VaccineID: vaccineA.ID, Date: time.Now(),
		}
		require.NoError(t, db.WithContext(ctx).Create(linked).Error)
		assertRejectedKeepsRow(t, linked)
	})

	t.Run("vaccination referencing a foreign pet rejects", func(t *testing.T) {
		corrupt := &model.Vaccination{
			ClinicID: clinicA, PetID: &petB.ID, VaccineID: vaccineA.ID, Date: time.Now(),
		}
		require.NoError(t, db.WithContext(ctx).Create(corrupt).Error)
		assertRejectedKeepsRow(t, corrupt)
	})
}
