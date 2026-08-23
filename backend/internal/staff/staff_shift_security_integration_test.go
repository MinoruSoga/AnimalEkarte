package staff

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type blockingStaffClinicLookup struct {
	clinic.ClinicRepository
	locked  chan struct{}
	release chan struct{}
}

func (r *blockingStaffClinicLookup) LockActiveByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	clinicRecord, err := r.ClinicRepository.LockActiveByID(ctx, id)
	if err != nil {
		return nil, err
	}
	close(r.locked)
	select {
	case <-r.release:
		return clinicRecord, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type observedStaffClinicLookup struct {
	clinic.ClinicRepository
	started chan struct{}
}

func (r *observedStaffClinicLookup) LockActiveByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	close(r.started)
	return r.ClinicRepository.LockActiveByID(ctx, id)
}

type blockingClinicDeleteRepository struct {
	clinic.ClinicRepository
	deleted chan struct{}
	release chan struct{}
}

func (r *blockingClinicDeleteRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.ClinicRepository.Delete(ctx, id); err != nil {
		return err
	}
	close(r.deleted)
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type blockingShiftEntryCreateRepository struct {
	ShiftEntryRepository
	created chan struct{}
	release chan struct{}
}

func (r *blockingShiftEntryCreateRepository) Create(ctx context.Context, entry *model.ShiftEntry) error {
	if err := r.ShiftEntryRepository.Create(ctx, entry); err != nil {
		return err
	}
	close(r.created)
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type observedShiftStaffLocker struct {
	StaffRepository
	started chan struct{}
}

func (r *observedShiftStaffLocker) LockActiveByIDForShare(
	ctx context.Context,
	staffID uint64,
) (*model.Staff, error) {
	close(r.started)
	return r.StaffRepository.LockActiveByIDForShare(ctx, staffID)
}

type observedStaffDeleteLocker struct {
	StaffRepository
	started chan struct{}
}

func (r *observedStaffDeleteLocker) LockActiveByIDForUpdateInClinic(
	ctx context.Context,
	clinicID, staffID uint64,
) (*model.Staff, error) {
	close(r.started)
	return r.StaffRepository.LockActiveByIDForUpdateInClinic(ctx, clinicID, staffID)
}

type observedStaffAssignmentLocker struct {
	StaffRepository
	started chan struct{}
}

func (r *observedStaffAssignmentLocker) LockActiveByIDForUpdate(
	ctx context.Context,
	staffID uint64,
) (*model.Staff, error) {
	close(r.started)
	return r.StaffRepository.LockActiveByIDForUpdate(ctx, staffID)
}

type blockingStaffDeleteRepository struct {
	StaffRepository
	deleted chan struct{}
	release chan struct{}
}

func (r *blockingStaffDeleteRepository) Delete(ctx context.Context, clinicID, staffID uint64) error {
	if err := r.StaffRepository.Delete(ctx, clinicID, staffID); err != nil {
		return err
	}
	close(r.deleted)
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func setupStaffShiftSecurityIntegrationTest(
	t *testing.T,
) (*gorm.DB, *model.Clinic, *model.Staff, StaffClinicAssignmentRepository) {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ShiftEntry{},
		&model.ShiftEntryBreak{},
		&model.PermissionGroup{},
		&model.Reservation{},
		&model.ExaminationType{},
		&model.ExamTypeField{},
		&model.Examination{},
		&model.Hospitalization{},
		&model.Vaccination{},
		&model.CheckupType{},
		&model.Checkup{},
		&model.ClinicIntegration{},
		&model.LstepSettings{},
		&model.MedicalRecordAddendum{},
		&model.CashRegisterClose{},
		&model.VitalRecord{},
	))
	testdb.EnsureClinicSettingsTable(t, db)
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE shift_entry_breaks, shift_entries, staff_clinic_assignments, staffs, clinics, companies CASCADE",
	).Error)
	require.NoError(t, db.Exec("DROP INDEX IF EXISTS ux_test_staff_shift_security_assignment").Error)
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX ux_test_staff_shift_security_assignment
		ON staff_clinic_assignments (staff_id, clinic_id)
	`).Error)

	company := &model.Company{Name: "SEC-STAFF integration company"}
	require.NoError(t, db.Create(company).Error)
	clinicRecord := &model.Clinic{CompanyID: company.ID, Name: "SEC-STAFF integration clinic", IsActive: true}
	require.NoError(t, db.Create(clinicRecord).Error)
	staff := &model.Staff{
		ClinicID:  clinicRecord.ID,
		Name:      "SEC-STAFF integration staff",
		IsActive:  true,
		StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)
	assignmentRepo := NewStaffClinicAssignmentRepository(db)
	require.NoError(t, assignmentRepo.Create(context.Background(), &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicRecord.ID,
		IsMain:   true,
	}))
	return db, clinicRecord, staff, assignmentRepo
}

func TestShiftEntryService_Create_RejectsSoftDeletedStaffWithActiveAssignmentDatabase(t *testing.T) {
	db, clinicRecord, staff, assignmentRepo := setupStaffShiftSecurityIntegrationTest(t)
	require.NoError(t, db.Delete(&model.Staff{}, staff.ID).Error)

	staffRepo := NewStaffRepository(db)
	shiftRepo := NewShiftEntryRepository(db)
	svc := NewShiftEntryService(
		shiftRepo,
		staffRepo,
		assignmentRepo,
		persistence.NewTransactor(db),
	)

	entry, err := svc.Create(context.Background(), clinicRecord.ID, &CreateShiftEntryInput{
		StaffID:   staff.ID,
		Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		ShiftType: string(model.ShiftTypeOff),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)
	assert.Nil(t, entry)
	var count int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).Where("staff_id = ?", staff.ID).Count(&count).Error)
	assert.Zero(t, count)
}

func TestStaffDeleteAndShiftCreate_ShiftWinnerPreservesStaffAndShiftDatabase(t *testing.T) {
	db, clinicRecord, staff, assignmentRepo := setupStaffShiftSecurityIntegrationTest(t)
	staffRepo := NewStaffRepository(db)
	baseShiftRepo := NewShiftEntryRepository(db)
	blockingShiftRepo := &blockingShiftEntryCreateRepository{
		ShiftEntryRepository: baseShiftRepo,
		created:              make(chan struct{}),
		release:              make(chan struct{}),
	}
	observedDeleteRepo := &observedStaffDeleteLocker{
		StaffRepository: staffRepo,
		started:         make(chan struct{}),
	}
	transactor := persistence.NewTransactor(db)
	shiftSvc := NewShiftEntryService(blockingShiftRepo, staffRepo, assignmentRepo, transactor)
	staffSvc := NewStaffService(
		observedDeleteRepo,
		nil,
		assignmentRepo,
		&stubReservationForStaff{},
		baseShiftRepo,
		nil,
		nil,
		nil,
		nil,
		transactor,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shiftDone := make(chan error, 1)
	go func() {
		_, err := shiftSvc.Create(ctx, clinicRecord.ID, &CreateShiftEntryInput{
			StaffID:   staff.ID,
			Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
			ShiftType: string(model.ShiftTypeOff),
		})
		shiftDone <- err
	}()
	select {
	case <-blockingShiftRepo.created:
	case <-ctx.Done():
		t.Fatal("shift create did not reach the blocking write")
	}

	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- staffSvc.Delete(ctx, clinicRecord.ID, staff.ID, false)
	}()
	select {
	case <-observedDeleteRepo.started:
	case <-ctx.Done():
		t.Fatal("staff delete did not attempt its update lock")
	}
	select {
	case err := <-deleteDone:
		t.Fatalf("staff delete escaped the shift transaction lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(blockingShiftRepo.release)

	require.NoError(t, <-shiftDone)
	deleteErr := <-deleteDone
	require.Error(t, deleteErr)
	assert.True(t, apperrors.IsConflict(deleteErr), "unexpected delete error: %v", deleteErr)

	var shiftCount int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).Where("staff_id = ?", staff.ID).Count(&shiftCount).Error)
	assert.Equal(t, int64(1), shiftCount)
	_, findErr := staffRepo.FindByID(context.Background(), staff.ID)
	require.NoError(t, findErr, "successful shift must keep its active staff")
}

func TestStaffDeleteAndShiftCreate_DeleteWinnerPreventsOrphanShiftDatabase(t *testing.T) {
	db, clinicRecord, staff, assignmentRepo := setupStaffShiftSecurityIntegrationTest(t)
	baseStaffRepo := NewStaffRepository(db)
	blockingDeleteRepo := &blockingStaffDeleteRepository{
		StaffRepository: baseStaffRepo,
		deleted:         make(chan struct{}),
		release:         make(chan struct{}),
	}
	observedShiftLocker := &observedShiftStaffLocker{
		StaffRepository: baseStaffRepo,
		started:         make(chan struct{}),
	}
	shiftRepo := NewShiftEntryRepository(db)
	transactor := persistence.NewTransactor(db)
	staffSvc := NewStaffService(
		blockingDeleteRepo,
		nil,
		assignmentRepo,
		&stubReservationForStaff{},
		shiftRepo,
		nil,
		nil,
		nil,
		nil,
		transactor,
	)
	shiftSvc := NewShiftEntryService(shiftRepo, observedShiftLocker, assignmentRepo, transactor)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- staffSvc.Delete(ctx, clinicRecord.ID, staff.ID, false)
	}()
	select {
	case <-blockingDeleteRepo.deleted:
	case <-ctx.Done():
		t.Fatal("staff delete did not reach its blocking write")
	}

	shiftDone := make(chan error, 1)
	go func() {
		_, err := shiftSvc.Create(ctx, clinicRecord.ID, &CreateShiftEntryInput{
			StaffID:   staff.ID,
			Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
			ShiftType: string(model.ShiftTypeOff),
		})
		shiftDone <- err
	}()
	select {
	case <-observedShiftLocker.started:
	case <-ctx.Done():
		t.Fatal("shift create did not attempt its staff share lock")
	}
	select {
	case err := <-shiftDone:
		t.Fatalf("shift create escaped the staff delete transaction lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(blockingDeleteRepo.release)

	require.NoError(t, <-deleteDone)
	shiftErr := <-shiftDone
	require.Error(t, shiftErr)
	assert.True(t, apperrors.IsNotFound(shiftErr), "unexpected shift error: %v", shiftErr)

	var shiftCount int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).Where("staff_id = ?", staff.ID).Count(&shiftCount).Error)
	assert.Zero(t, shiftCount, "deleted staff must not retain an orphan shift")
	_, findErr := baseStaffRepo.FindByID(context.Background(), staff.ID)
	require.Error(t, findErr)
	assert.True(t, apperrors.IsNotFound(findErr))
}

func TestStaffSetAssignmentsAndShiftCreate_ShiftWinnerPreservesAssignmentDatabase(t *testing.T) {
	db, sourceClinic, staff, assignmentRepo := setupStaffShiftSecurityIntegrationTest(t)
	targetClinic := &model.Clinic{
		CompanyID: sourceClinic.CompanyID,
		Name:      "SEC-STAFF shift winner target",
		IsActive:  true,
	}
	require.NoError(t, db.Create(targetClinic).Error)

	baseStaffRepo := NewStaffRepository(db)
	observedAssignmentLocker := &observedStaffAssignmentLocker{
		StaffRepository: baseStaffRepo,
		started:         make(chan struct{}),
	}
	baseShiftRepo := NewShiftEntryRepository(db)
	blockingShiftRepo := &blockingShiftEntryCreateRepository{
		ShiftEntryRepository: baseShiftRepo,
		created:              make(chan struct{}),
		release:              make(chan struct{}),
	}
	transactor := persistence.NewTransactor(db)
	shiftSvc := NewShiftEntryService(blockingShiftRepo, baseStaffRepo, assignmentRepo, transactor)
	staffSvc := NewStaffService(
		observedAssignmentLocker,
		nil,
		assignmentRepo,
		&stubReservationForStaff{},
		baseShiftRepo,
		nil,
		nil,
		nil,
		clinic.NewClinicRepository(db),
		transactor,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shiftDone := make(chan error, 1)
	go func() {
		_, err := shiftSvc.Create(ctx, sourceClinic.ID, &CreateShiftEntryInput{
			StaffID:   staff.ID,
			Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
			ShiftType: string(model.ShiftTypeOff),
		})
		shiftDone <- err
	}()
	select {
	case <-blockingShiftRepo.created:
	case <-ctx.Done():
		t.Fatal("shift create did not reach the blocking write")
	}

	setDone := make(chan error, 1)
	go func() {
		setDone <- staffSvc.SetClinicAssignments(ctx, &SetClinicAssignmentsInput{
			StaffID:             staff.ID,
			ClinicIDs:           []uint64{targetClinic.ID},
			AuthorizedClinicIDs: []uint64{sourceClinic.ID, targetClinic.ID},
		})
	}()
	select {
	case <-observedAssignmentLocker.started:
	case <-ctx.Done():
		t.Fatal("assignment replacement did not attempt its staff update lock")
	}
	select {
	case err := <-setDone:
		t.Fatalf("assignment replacement escaped the shift transaction lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(blockingShiftRepo.release)

	require.NoError(t, <-shiftDone)
	setErr := <-setDone
	require.Error(t, setErr)
	assert.True(t, apperrors.IsConflict(setErr), "unexpected assignment replacement error: %v", setErr)

	assignments, findAssignmentsErr := assignmentRepo.FindByStaffID(context.Background(), staff.ID)
	require.NoError(t, findAssignmentsErr)
	require.Len(t, assignments, 1)
	assert.Equal(t, sourceClinic.ID, assignments[0].ClinicID)
	unchangedStaff, findStaffErr := baseStaffRepo.FindByID(context.Background(), staff.ID)
	require.NoError(t, findStaffErr)
	assert.Equal(t, sourceClinic.ID, unchangedStaff.ClinicID)
	var shiftCount int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).
		Where("clinic_id = ? AND staff_id = ?", sourceClinic.ID, staff.ID).
		Count(&shiftCount).Error)
	assert.Equal(t, int64(1), shiftCount)
}

func TestStaffSetAssignmentsAndShiftCreate_AssignmentWinnerPreventsOrphanShiftDatabase(t *testing.T) {
	db, sourceClinic, staff, assignmentRepo := setupStaffShiftSecurityIntegrationTest(t)
	targetClinic := &model.Clinic{
		CompanyID: sourceClinic.CompanyID,
		Name:      "SEC-STAFF assignment winner shift target",
		IsActive:  true,
	}
	require.NoError(t, db.Create(targetClinic).Error)

	baseClinicRepo := clinic.NewClinicRepository(db)
	blockingClinicLookup := &blockingStaffClinicLookup{
		ClinicRepository: baseClinicRepo,
		locked:           make(chan struct{}),
		release:          make(chan struct{}),
	}
	baseStaffRepo := NewStaffRepository(db)
	observedShiftLocker := &observedShiftStaffLocker{
		StaffRepository: baseStaffRepo,
		started:         make(chan struct{}),
	}
	shiftRepo := NewShiftEntryRepository(db)
	transactor := persistence.NewTransactor(db)
	staffSvc := NewStaffService(
		baseStaffRepo,
		nil,
		assignmentRepo,
		&stubReservationForStaff{},
		shiftRepo,
		nil,
		nil,
		nil,
		blockingClinicLookup,
		transactor,
	)
	shiftSvc := NewShiftEntryService(shiftRepo, observedShiftLocker, assignmentRepo, transactor)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	setDone := make(chan error, 1)
	go func() {
		setDone <- staffSvc.SetClinicAssignments(ctx, &SetClinicAssignmentsInput{
			StaffID:             staff.ID,
			ClinicIDs:           []uint64{targetClinic.ID},
			AuthorizedClinicIDs: []uint64{sourceClinic.ID, targetClinic.ID},
		})
	}()
	select {
	case <-blockingClinicLookup.locked:
	case <-ctx.Done():
		t.Fatal("assignment replacement did not acquire the target clinic share lock")
	}

	shiftDone := make(chan error, 1)
	go func() {
		_, err := shiftSvc.Create(ctx, sourceClinic.ID, &CreateShiftEntryInput{
			StaffID:   staff.ID,
			Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
			ShiftType: string(model.ShiftTypeOff),
		})
		shiftDone <- err
	}()
	select {
	case <-observedShiftLocker.started:
	case <-ctx.Done():
		t.Fatal("shift create did not attempt its staff share lock")
	}
	select {
	case err := <-shiftDone:
		t.Fatalf("shift create escaped the assignment replacement transaction lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(blockingClinicLookup.release)

	require.NoError(t, <-setDone)
	shiftErr := <-shiftDone
	require.Error(t, shiftErr)
	assert.True(t, apperrors.IsNotFound(shiftErr), "unexpected shift error: %v", shiftErr)

	assignments, findAssignmentsErr := assignmentRepo.FindByStaffID(context.Background(), staff.ID)
	require.NoError(t, findAssignmentsErr)
	require.Len(t, assignments, 1)
	assert.Equal(t, targetClinic.ID, assignments[0].ClinicID)
	updatedStaff, findStaffErr := baseStaffRepo.FindByID(context.Background(), staff.ID)
	require.NoError(t, findStaffErr)
	assert.Equal(t, targetClinic.ID, updatedStaff.ClinicID)
	var shiftCount int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).
		Where("clinic_id = ? AND staff_id = ?", sourceClinic.ID, staff.ID).
		Count(&shiftCount).Error)
	assert.Zero(t, shiftCount)
}

func TestStaffSetAssignmentsAndClinicDelete_AssignmentWinnerPreservesClinicAndAssignmentDatabase(t *testing.T) {
	db, sourceClinic, staff, assignmentRepo := setupStaffShiftSecurityIntegrationTest(t)
	targetClinic := &model.Clinic{
		CompanyID: sourceClinic.CompanyID,
		Name:      "SEC-STAFF assignment winner target",
		IsActive:  true,
	}
	require.NoError(t, db.Create(targetClinic).Error)
	baseClinicRepo := clinic.NewClinicRepository(db)
	lockingClinicLookup := &blockingStaffClinicLookup{
		ClinicRepository: baseClinicRepo,
		locked:           make(chan struct{}),
		release:          make(chan struct{}),
	}
	transactor := persistence.NewTransactor(db)
	staffSvc := NewStaffService(
		NewStaffRepository(db),
		nil,
		assignmentRepo,
		&stubReservationForStaff{},
		NewShiftEntryRepository(db),
		nil,
		nil,
		nil,
		lockingClinicLookup,
		transactor,
	)
	clinicSvc := clinic.NewClinicService(
		baseClinicRepo,
		authdomain.NewPermissionGroupRepository(db),
		transactor,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	setDone := make(chan error, 1)
	go func() {
		setDone <- staffSvc.SetClinicAssignments(ctx, &SetClinicAssignmentsInput{
			StaffID:             staff.ID,
			ClinicIDs:           []uint64{targetClinic.ID},
			AuthorizedClinicIDs: []uint64{sourceClinic.ID, targetClinic.ID},
		})
	}()
	select {
	case <-lockingClinicLookup.locked:
	case <-ctx.Done():
		t.Fatal("assignment replacement did not acquire the target clinic SHARE lock")
	}

	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- clinicSvc.DeleteClinic(ctx, targetClinic.ID)
	}()
	select {
	case err := <-deleteDone:
		t.Fatalf("clinic delete escaped the target clinic SHARE lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(lockingClinicLookup.release)

	require.NoError(t, <-setDone)
	deleteErr := <-deleteDone
	require.Error(t, deleteErr)
	assert.True(t, apperrors.IsConflict(deleteErr), "unexpected clinic delete error: %v", deleteErr)
	_, findClinicErr := baseClinicRepo.FindByID(context.Background(), targetClinic.ID)
	require.NoError(t, findClinicErr)
	assignments, findAssignmentsErr := assignmentRepo.FindByStaffID(context.Background(), staff.ID)
	require.NoError(t, findAssignmentsErr)
	require.Len(t, assignments, 1)
	assert.Equal(t, targetClinic.ID, assignments[0].ClinicID)
	updatedStaff, findStaffErr := NewStaffRepository(db).FindByID(context.Background(), staff.ID)
	require.NoError(t, findStaffErr)
	assert.Equal(t, targetClinic.ID, updatedStaff.ClinicID)
}

func TestStaffSetAssignmentsAndClinicDelete_DeleteWinnerLeavesAssignmentsUnchangedDatabase(t *testing.T) {
	db, sourceClinic, staff, assignmentRepo := setupStaffShiftSecurityIntegrationTest(t)
	targetClinic := &model.Clinic{
		CompanyID: sourceClinic.CompanyID,
		Name:      "SEC-STAFF clinic delete winner target",
		IsActive:  true,
	}
	require.NoError(t, db.Create(targetClinic).Error)
	baseClinicRepo := clinic.NewClinicRepository(db)
	blockingDeleteRepo := &blockingClinicDeleteRepository{
		ClinicRepository: baseClinicRepo,
		deleted:          make(chan struct{}),
		release:          make(chan struct{}),
	}
	transactor := persistence.NewTransactor(db)
	clinicSvc := clinic.NewClinicService(
		blockingDeleteRepo,
		authdomain.NewPermissionGroupRepository(db),
		transactor,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- clinicSvc.DeleteClinic(ctx, targetClinic.ID)
	}()
	select {
	case <-blockingDeleteRepo.deleted:
	case <-ctx.Done():
		t.Fatal("clinic delete did not acquire its delete lock")
	}

	observedLookup := &observedStaffClinicLookup{
		ClinicRepository: baseClinicRepo,
		started:          make(chan struct{}),
	}
	staffSvc := NewStaffService(
		NewStaffRepository(db),
		nil,
		assignmentRepo,
		&stubReservationForStaff{},
		NewShiftEntryRepository(db),
		nil,
		nil,
		nil,
		observedLookup,
		transactor,
	)
	setDone := make(chan error, 1)
	go func() {
		setDone <- staffSvc.SetClinicAssignments(ctx, &SetClinicAssignmentsInput{
			StaffID:             staff.ID,
			ClinicIDs:           []uint64{targetClinic.ID},
			AuthorizedClinicIDs: []uint64{sourceClinic.ID, targetClinic.ID},
		})
	}()
	select {
	case <-observedLookup.started:
	case <-ctx.Done():
		t.Fatal("assignment replacement did not attempt the target clinic lock")
	}
	select {
	case err := <-setDone:
		t.Fatalf("assignment replacement escaped the clinic delete lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(blockingDeleteRepo.release)

	require.NoError(t, <-deleteDone)
	setErr := <-setDone
	require.Error(t, setErr)
	assert.True(t, apperrors.IsNotFound(setErr), "unexpected assignment error: %v", setErr)
	_, findClinicErr := baseClinicRepo.FindByID(context.Background(), targetClinic.ID)
	require.Error(t, findClinicErr)
	assert.True(t, apperrors.IsNotFound(findClinicErr))
	assignments, findAssignmentsErr := assignmentRepo.FindByStaffID(context.Background(), staff.ID)
	require.NoError(t, findAssignmentsErr)
	require.Len(t, assignments, 1)
	assert.Equal(t, sourceClinic.ID, assignments[0].ClinicID)
	unchangedStaff, findStaffErr := NewStaffRepository(db).FindByID(context.Background(), staff.ID)
	require.NoError(t, findStaffErr)
	assert.Equal(t, sourceClinic.ID, unchangedStaff.ClinicID)
}
