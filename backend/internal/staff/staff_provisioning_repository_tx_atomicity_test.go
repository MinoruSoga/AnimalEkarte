package staff_test

// staff_provisioning_repository_tx_atomicity_test.go — ambient-tx participation proofs
// for staffProvisioningRepository methods enrolled on dbOrTxParticipatingMethods
// (lintscan TestDBOrTxInventory_MatchesAllowlist).
//
// Coverage policy (not tautology):
//  1. Writers under WithTx must roll back when a later step fails (DBOrTx participation).
//  2. Ambient-required methods fail closed when TxFromContext is absent.
//  3. Readers under ambient tx observe uncommitted writes from the same tx, then roll back.
//
// Method × test mapping is documented in each Test* name and the table in the package comment of
// the lintscan allowlist entry for staff_provisioning_repository.go.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func requireInternalAppError(t *testing.T, err error, msgSubstr string) {
	t.Helper()
	require.Error(t, err)
	var app *apperrors.AppError
	require.True(t, errors.As(err, &app), "want *apperrors.AppError, got %T: %v", err, err)
	assert.Equal(t, "INTERNAL", app.Code)
	assert.Contains(t, app.Message, msgSubstr)
}

// Covers CreateAccount, CreateStaff, CreateAssignment, AssignPermissionGroups, WriteAudit,
// LockOccupationForShare, AcquireBatchLock (all called under ambient WithTx then rolled back).
func TestStaffProvisioningRepository_WritersParticipateInAmbientTxRollback(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	ctx := context.Background()
	forced := errors.New("forced post-write rollback")

	err := f.repo.WithTx(ctx, func(txCtx context.Context) error {
		if err := f.repo.AcquireBatchLock(txCtx, "tx-atomicity-batch"); err != nil {
			return err
		}
		if err := f.repo.LockOccupationForShare(txCtx, f.clinicA, f.occupationID); err != nil {
			return err
		}

		account := &model.Account{
			Email:        "provision-tx-writer@example.test",
			PasswordHash: "hash",
			IsActive:     true,
		}
		if err := f.repo.CreateAccount(txCtx, account); err != nil {
			return err
		}
		accountID := account.ID
		staff := &model.Staff{
			ClinicID:           f.clinicA,
			AccountID:          &accountID,
			Name:               "tx-writer",
			OccupationID:       &f.occupationID,
			IsActive:           true,
			ReservationVisible: true,
		}
		if err := f.repo.CreateStaff(txCtx, staff); err != nil {
			return err
		}
		if err := f.repo.CreateAssignment(txCtx, &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: f.clinicA,
		}); err != nil {
			return err
		}
		if err := f.repo.AssignPermissionGroups(txCtx, f.clinicA, staff.ID, []uint64{f.groupID}); err != nil {
			return err
		}
		clinicID := f.clinicA
		if err := f.repo.WriteAudit(txCtx, &model.AuditLog{
			ClinicID:  &clinicID,
			ActorType: model.AuditActorTypeSystem,
			Action:    model.AuditActionStaffProvisionCreate,
			Resource:  model.AuditResourceStaff,
			CreatedAt: time.Now(),
		}); err != nil {
			return err
		}
		return forced
	})
	require.ErrorIs(t, err, forced)

	var accounts int64
	require.NoError(t, f.db.Model(&model.Account{}).
		Where("email = ?", "provision-tx-writer@example.test").
		Count(&accounts).Error)
	assert.Zero(t, accounts, "CreateAccount must participate in ambient tx and roll back")

	var staffCount int64
	require.NoError(t, f.db.Model(&model.Staff{}).
		Where("name = ?", "tx-writer").
		Count(&staffCount).Error)
	assert.Zero(t, staffCount, "CreateStaff must participate in ambient tx and roll back")

	var assignCount int64
	require.NoError(t, f.db.Model(&model.StaffClinicAssignment{}).Count(&assignCount).Error)
	assert.Zero(t, assignCount, "CreateAssignment must participate in ambient tx and roll back")

	var groupLink int64
	require.NoError(t, f.db.Model(&model.StaffPermissionGroup{}).Count(&groupLink).Error)
	assert.Zero(t, groupLink, "AssignPermissionGroups must participate in ambient tx and roll back")

	var audits int64
	require.NoError(t, f.db.Model(&model.AuditLog{}).
		Where("action = ?", model.AuditActionStaffProvisionCreate).
		Count(&audits).Error)
	assert.Zero(t, audits, "WriteAudit must participate in ambient tx and roll back")
}

// Covers AcquireBatchLock, LockOccupationForShare, WriteAudit fail-closed without ambient tx.
func TestStaffProvisioningRepository_AmbientRequiredMethodsFailClosedWithoutTx(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	ctx := context.Background()

	requireInternalAppError(t, f.repo.AcquireBatchLock(ctx, "no-ambient"),
		"staff provision lock requires an ambient transaction")
	requireInternalAppError(t, f.repo.LockOccupationForShare(ctx, f.clinicA, f.occupationID),
		"occupation lock requires an ambient transaction")
	clinicID := f.clinicA
	requireInternalAppError(t, f.repo.WriteAudit(ctx, &model.AuditLog{
		ClinicID:  &clinicID,
		ActorType: model.AuditActorTypeSystem,
		Action:    model.AuditActionStaffProvisionCreate,
		Resource:  model.AuditResourceStaff,
		CreatedAt: time.Now(),
	}), "staff provision audit requires an ambient transaction")
}

// Covers FindAccountByID, FindStaffByAccountID, EmailExists, ClinicExists,
// OccupationBelongsToClinic, PermissionGroupsBelongToClinic, StaffAssignedToClinic,
// HasMasterStaffCreate, FindReceiptsInScope — each observes ambient *uncommitted*
// state written inside the same WithTx (would RED if DBOrTx fell back to base db).
func TestStaffProvisioningRepository_ReadersParticipateInAmbientTx(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	ctx := context.Background()
	forced := errors.New("forced reader probe rollback")

	// Company is committed so ambient clinic create has a valid FK parent.
	var companyID uint64
	require.NoError(t, f.db.Model(&model.Clinic{}).Select("company_id").
		Where("id = ?", f.clinicA).Scan(&companyID).Error)

	err := f.repo.WithTx(ctx, func(txCtx context.Context) error {
		db := persistence.DBOrTx(txCtx, f.db)

		// --- Uncommitted clinic → ClinicExists ---
		probeClinic := &model.Clinic{CompanyID: companyID, Name: "ambient-clinic", IsActive: true}
		if err := db.Create(probeClinic).Error; err != nil {
			return err
		}
		ok, err := f.repo.ClinicExists(txCtx, probeClinic.ID)
		if err != nil || !ok {
			return errors.New("ClinicExists did not see uncommitted ambient clinic")
		}

		// --- Uncommitted occupation → OccupationBelongsToClinic ---
		probeOcc := &model.Occupation{ClinicID: f.clinicA, Name: "ambient-occ", IsActive: true}
		if err := db.Create(probeOcc).Error; err != nil {
			return err
		}
		ok, err = f.repo.OccupationBelongsToClinic(txCtx, f.clinicA, probeOcc.ID)
		if err != nil || !ok {
			return errors.New("OccupationBelongsToClinic did not see uncommitted ambient occupation")
		}

		// --- Uncommitted permission group → PermissionGroupsBelongToClinic ---
		probeGroup := &model.PermissionGroup{ClinicID: f.clinicA, Name: "ambient-group", IsActive: true}
		if err := db.Create(probeGroup).Error; err != nil {
			return err
		}
		if err := f.repo.PermissionGroupsBelongToClinic(txCtx, f.clinicA, []uint64{probeGroup.ID}); err != nil {
			return errors.New("PermissionGroupsBelongToClinic did not see uncommitted ambient group: " + err.Error())
		}

		// --- Uncommitted account/staff/assignment/rules → Find*/StaffAssigned/HasMaster ---
		probeAccount := &model.Account{
			Email: "provision-ambient-read@example.test", PasswordHash: "h", IsActive: true,
		}
		if err := f.repo.CreateAccount(txCtx, probeAccount); err != nil {
			return err
		}
		got, err := f.repo.FindAccountByID(txCtx, probeAccount.ID)
		if err != nil {
			return err
		}
		if got.Email != probeAccount.Email {
			return errors.New("FindAccountByID did not see uncommitted ambient write")
		}
		exists, err := f.repo.EmailExists(txCtx, probeAccount.Email)
		if err != nil || !exists {
			return errors.New("EmailExists did not see uncommitted ambient write")
		}
		accountID := probeAccount.ID
		probeStaff := &model.Staff{
			ClinicID: f.clinicA, AccountID: &accountID, Name: "ambient-staff",
			OccupationID: &probeOcc.ID, IsActive: true, ReservationVisible: true,
		}
		if err := f.repo.CreateStaff(txCtx, probeStaff); err != nil {
			return err
		}
		if _, err := f.repo.FindStaffByAccountID(txCtx, probeAccount.ID); err != nil {
			return errors.New("FindStaffByAccountID did not see uncommitted ambient staff: " + err.Error())
		}
		if err := f.repo.CreateAssignment(txCtx, &model.StaffClinicAssignment{
			StaffID: probeStaff.ID, ClinicID: f.clinicA,
		}); err != nil {
			return err
		}
		ok, err = f.repo.StaffAssignedToClinic(txCtx, probeStaff.ID, f.clinicA)
		if err != nil || !ok {
			return errors.New("StaffAssignedToClinic did not see uncommitted ambient assignment")
		}
		// master-staff create rule on the ambient group, then link staff → HasMasterStaffCreate
		if err := db.Create(&model.PermissionGroupRule{
			GroupID:   probeGroup.ID,
			Resource:  string(model.ResourceMasterStaff),
			CanView:   true,
			CanCreate: true,
			CanEdit:   true,
			CanDelete: true,
		}).Error; err != nil {
			return err
		}
		if err := f.repo.AssignPermissionGroups(txCtx, f.clinicA, probeStaff.ID, []uint64{probeGroup.ID}); err != nil {
			return err
		}
		ok, err = f.repo.HasMasterStaffCreate(txCtx, probeStaff.ID, f.clinicA)
		if err != nil || !ok {
			return errors.New("HasMasterStaffCreate did not see uncommitted ambient permission graph")
		}

		// --- Uncommitted receipt audit → FindReceiptsInScope ---
		clinicID := f.clinicA
		payload, _ := json.Marshal(map[string]any{
			"batch_id": "reader-batch", "digest": "d", "count": 1,
		})
		if err := f.repo.WriteAudit(txCtx, &model.AuditLog{
			ClinicID:  &clinicID,
			ActorType: model.AuditActorTypeSystem,
			Action:    model.AuditActionStaffProvisionReceipt,
			Resource:  model.AuditResourceStaffProvisionBatch,
			NewValue:  payload,
			CreatedAt: time.Now(),
		}); err != nil {
			return err
		}
		receipts, err := f.repo.FindReceiptsInScope(txCtx, []uint64{f.clinicA}, "reader-batch")
		if err != nil {
			return err
		}
		if len(receipts) != 1 {
			return errors.New("FindReceiptsInScope did not see ambient receipt audit")
		}

		return forced
	})
	require.ErrorIs(t, err, forced)

	var probeAccounts int64
	require.NoError(t, f.db.Model(&model.Account{}).
		Where("email = ?", "provision-ambient-read@example.test").
		Count(&probeAccounts).Error)
	assert.Zero(t, probeAccounts, "ambient CreateAccount probe must roll back")

	var probeClinics int64
	require.NoError(t, f.db.Model(&model.Clinic{}).
		Where("name = ?", "ambient-clinic").
		Count(&probeClinics).Error)
	assert.Zero(t, probeClinics, "ambient clinic probe must roll back")

	var receiptAudits int64
	require.NoError(t, f.db.Model(&model.AuditLog{}).
		Where("action = ?", model.AuditActionStaffProvisionReceipt).
		Count(&receiptAudits).Error)
	assert.Zero(t, receiptAudits, "ambient WriteAudit probe must roll back")
}

// Explicit ambient handle identity for CreateAccount: if DBOrTx were replaced by r.db,
// the row would survive the forced rollback (RED-documented contract).
func TestStaffProvisioningRepository_CreateAccountAmbientHandleIsRequiredForRollback(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	ctx := context.Background()
	forced := errors.New("rollback create-account handle probe")

	err := f.repo.WithTx(ctx, func(txCtx context.Context) error {
		// Sanity: ambient tx must be present for the proof to mean anything.
		if persistence.TxFromContext(txCtx) == nil {
			return errors.New("WithTx did not install ambient tx")
		}
		account := &model.Account{
			Email: "provision-handle-probe@example.test", PasswordHash: "h", IsActive: true,
		}
		if err := f.repo.CreateAccount(txCtx, account); err != nil {
			return err
		}
		return forced
	})
	require.ErrorIs(t, err, forced)

	var n int64
	require.NoError(t, f.db.Model(&model.Account{}).
		Where("email = ?", "provision-handle-probe@example.test").
		Count(&n).Error)
	assert.Zero(t, n)
}
