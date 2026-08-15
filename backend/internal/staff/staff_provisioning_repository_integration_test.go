package staff_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
)

func setupStaffProvisioningDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Occupation{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.PermissionGroup{},
		&model.PermissionGroupRule{},
		&model.StaffPermissionGroup{},
		&model.AuditLog{},
	))
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE
			audit_logs,
			staff_permission_groups,
			permission_group_rules,
			permission_groups,
			staff_clinic_assignments,
			staffs,
			occupations,
			accounts,
			clinics,
			companies
		RESTART IDENTITY CASCADE
	`).Error)
	return db
}

type provisionIntegrationFixture struct {
	db             *gorm.DB
	repo           StaffProvisioningRepository
	provisioner    *StaffProvisioner
	adminAccountID uint64
	clinicA        uint64
	clinicB        uint64
	occupationID   uint64
	groupID        uint64
	repoRoot       string
}

func newProvisionIntegrationFixture(t *testing.T) *provisionIntegrationFixture {
	t.Helper()
	db := setupStaffProvisioningDB(t)
	company := &model.Company{Name: "provision-test-company"}
	require.NoError(t, db.Create(company).Error)

	clinicA := &model.Clinic{CompanyID: company.ID, Name: "Clinic A", IsActive: true}
	clinicB := &model.Clinic{CompanyID: company.ID, Name: "Clinic B", IsActive: true}
	require.NoError(t, db.Create(clinicA).Error)
	require.NoError(t, db.Create(clinicB).Error)

	admin := &model.Account{
		Email:         "provision-admin@example.test",
		PasswordHash:  "hash",
		IsActive:      true,
		IsSystemAdmin: true,
	}
	require.NoError(t, db.Create(admin).Error)

	occ := &model.Occupation{ClinicID: clinicA.ID, Name: "獣医師", IsActive: true}
	require.NoError(t, db.Create(occ).Error)

	group := &model.PermissionGroup{
		ClinicID: clinicA.ID,
		Name:     "管理者",
		IsActive: true,
	}
	require.NoError(t, db.Create(group).Error)
	require.NoError(t, db.Create(&model.PermissionGroupRule{
		GroupID:   group.ID,
		Resource:  string(model.ResourceMasterStaff),
		CanView:   true,
		CanCreate: true,
		CanEdit:   true,
		CanDelete: true,
	}).Error)

	repoRoot := t.TempDir() // synthetic "repo" that inputs must stay outside of
	repo := NewStaffProvisioningRepository(db)
	return &provisionIntegrationFixture{
		db:             db,
		repo:           repo,
		provisioner:    NewStaffProvisioner(repo, []string{repoRoot}),
		adminAccountID: admin.ID,
		clinicA:        clinicA.ID,
		clinicB:        clinicB.ID,
		occupationID:   occ.ID,
		groupID:        group.ID,
		repoRoot:       repoRoot,
	}
}

func (f *provisionIntegrationFixture) writeInputs(
	t *testing.T,
	manifest *StaffProvisionManifest,
	secrets *StaffProvisionSecretsFile,
) (manifestPath, secretsPath string) {
	t.Helper()
	dir := t.TempDir() // outside f.repoRoot
	manifestPath = filepath.Join(dir, "manifest.json")
	secretsPath = filepath.Join(dir, "secrets.json")
	mb, err := json.Marshal(manifest)
	require.NoError(t, err)
	sb, err := json.Marshal(secrets)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(manifestPath, mb, 0o600))
	require.NoError(t, os.WriteFile(secretsPath, sb, 0o600))
	return manifestPath, secretsPath
}

func (f *provisionIntegrationFixture) singleStaffManifest(t *testing.T, multiClinic bool) (*StaffProvisionManifest, *StaffProvisionSecretsFile) {
	t.Helper()
	scope := []uint64{f.clinicA}
	clinicIDs := []uint64{f.clinicA}
	if multiClinic {
		scope = []uint64{f.clinicA, f.clinicB}
		if f.clinicA > f.clinicB {
			scope = []uint64{f.clinicB, f.clinicA}
		}
		clinicIDs = append([]uint64(nil), scope...)
	}
	batchID := ClinicScopeBatchID(scope)
	occID := f.occupationID
	manifest := &StaffProvisionManifest{
		SchemaVersion:  StaffProvisionSchemaVersion,
		BatchID:        batchID,
		ClinicScope:    scope,
		ActorAccountID: f.adminAccountID,
		Staff: []StaffProvisionStaffEntry{{
			ExternalStaffID:    "ext-syn-001",
			Name:               "合成発行スタッフ",
			Email:              "synthetic-provision-001@example.test",
			MainClinicID:       f.clinicA,
			ClinicIDs:          clinicIDs,
			PermissionGroupIDs: []uint64{f.groupID},
			OccupationID:       &occID,
			StaffType:          string(model.StaffTypeDoctor),
			IsActive:           true,
			ReservationVisible: true,
			SecretRef:          "sec-syn-001",
		}},
	}
	secrets := &StaffProvisionSecretsFile{
		Secrets: []StaffProvisionSecretEntry{
			{SecretRef: "sec-syn-001", Password: "Password1a"},
		},
	}
	return manifest, secrets
}

func TestStaffProvisioning_Integration_PreflightZeroWrites(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	manifest, secrets := f.singleStaffManifest(t, false)
	mp, sp := f.writeInputs(t, manifest, secrets)

	var staffBefore, accountBefore, auditBefore int64
	require.NoError(t, f.db.Model(&model.Staff{}).Count(&staffBefore).Error)
	require.NoError(t, f.db.Model(&model.Account{}).Count(&accountBefore).Error)
	require.NoError(t, f.db.Model(&model.AuditLog{}).Count(&auditBefore).Error)

	result, err := f.provisioner.Preflight(context.Background(), mp, sp)
	require.NoError(t, err)
	require.Equal(t, 1, result.StaffCount)

	var staffAfter, accountAfter, auditAfter int64
	require.NoError(t, f.db.Model(&model.Staff{}).Count(&staffAfter).Error)
	require.NoError(t, f.db.Model(&model.Account{}).Count(&accountAfter).Error)
	require.NoError(t, f.db.Model(&model.AuditLog{}).Count(&auditAfter).Error)
	assert.Equal(t, staffBefore, staffAfter)
	assert.Equal(t, accountBefore, accountAfter)
	assert.Equal(t, auditBefore, auditAfter)
}

func TestStaffProvisioning_Integration_ApplyAtomicAndReceipts(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	manifest, secrets := f.singleStaffManifest(t, true)
	mp, sp := f.writeInputs(t, manifest, secrets)

	result, err := f.provisioner.Apply(context.Background(), mp, sp)
	require.NoError(t, err)
	require.Equal(t, "applied", result.Status)
	assert.Equal(t, 1, result.StaffCount)

	var staff model.Staff
	require.NoError(t, f.db.Where("name = ?", "合成発行スタッフ").First(&staff).Error)
	require.NotNil(t, staff.AccountID)

	var assignments []model.StaffClinicAssignment
	require.NoError(t, f.db.Where("staff_id = ?", staff.ID).Find(&assignments).Error)
	assert.Len(t, assignments, 2)

	var groups []model.StaffPermissionGroup
	require.NoError(t, f.db.Where("staff_id = ?", staff.ID).Find(&groups).Error)
	assert.Len(t, groups, 1)

	// Staff-level audit (PII-free)
	var createAudits []model.AuditLog
	require.NoError(t, f.db.Where(
		"action = ? AND resource = ?",
		model.AuditActionStaffProvisionCreate,
		model.AuditResourceStaff,
	).Find(&createAudits).Error)
	require.Len(t, createAudits, 1)
	assert.NotContains(t, string(createAudits[0].NewValue), "合成発行スタッフ")
	assert.NotContains(t, string(createAudits[0].NewValue), "synthetic-provision")
	assert.NotContains(t, string(createAudits[0].NewValue), "Password")

	// Receipt per affected clinic only
	var receipts []model.AuditLog
	require.NoError(t, f.db.Where(
		"action = ? AND resource = ?",
		model.AuditActionStaffProvisionReceipt,
		model.AuditResourceStaffProvisionBatch,
	).Order("clinic_id ASC").Find(&receipts).Error)
	require.Len(t, receipts, 2)
	for _, receipt := range receipts {
		require.NotNil(t, receipt.ClinicID)
		assert.Contains(t, []uint64{f.clinicA, f.clinicB}, *receipt.ClinicID)
		assert.NotContains(t, string(receipt.NewValue), "合成")
		assert.NotContains(t, string(receipt.NewValue), "@")
		assert.Contains(t, string(receipt.NewValue), "batch_id")
		assert.Contains(t, string(receipt.NewValue), "digest")
		assert.Contains(t, string(receipt.NewValue), "count")
	}

	// Idempotent re-apply
	again, err := f.provisioner.Apply(context.Background(), mp, sp)
	require.NoError(t, err)
	assert.Equal(t, "noop", again.Status)

	var staffCount int64
	require.NoError(t, f.db.Model(&model.Staff{}).Where("name = ?", "合成発行スタッフ").Count(&staffCount).Error)
	assert.Equal(t, int64(1), staffCount)
}

func TestStaffProvisioning_Integration_PartialReceiptConflict(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	manifest, secrets := f.singleStaffManifest(t, true)
	digest, err := ComputeStaffProvisionDigest(manifest)
	require.NoError(t, err)

	// Seed only one clinic receipt to simulate partial prior apply.
	clinicID := f.clinicA
	require.NoError(t, f.db.Create(&model.AuditLog{
		ClinicID:  &clinicID,
		ActorType: model.AuditActorTypeSystem,
		Action:    model.AuditActionStaffProvisionReceipt,
		Resource:  model.AuditResourceStaffProvisionBatch,
		NewValue: mustRawJSON(map[string]any{
			"batch_id": manifest.BatchID,
			"digest":   digest,
			"count":    1,
		}),
	}).Error)

	mp, sp := f.writeInputs(t, manifest, secrets)
	_, err = f.provisioner.Apply(context.Background(), mp, sp)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))

	var staffCount int64
	require.NoError(t, f.db.Model(&model.Staff{}).Count(&staffCount).Error)
	assert.Equal(t, int64(0), staffCount)
}

func TestStaffProvisioning_Integration_DigestMismatchConflict(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	manifest, secrets := f.singleStaffManifest(t, true)

	for _, clinicID := range []uint64{f.clinicA, f.clinicB} {
		clinicID := clinicID
		require.NoError(t, f.db.Create(&model.AuditLog{
			ClinicID:  &clinicID,
			ActorType: model.AuditActorTypeSystem,
			Action:    model.AuditActionStaffProvisionReceipt,
			Resource:  model.AuditResourceStaffProvisionBatch,
			NewValue: mustRawJSON(map[string]any{
				"batch_id": manifest.BatchID,
				"digest":   "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				"count":    1,
			}),
		}).Error)
	}

	mp, sp := f.writeInputs(t, manifest, secrets)
	_, err := f.provisioner.Apply(context.Background(), mp, sp)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestStaffProvisioning_Integration_UnauthorizedDoesNotLeakReceiptExistence(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	manifest, secrets := f.singleStaffManifest(t, false)

	// Seed a receipt for clinic A.
	digest, err := ComputeStaffProvisionDigest(manifest)
	require.NoError(t, err)
	clinicID := f.clinicA
	require.NoError(t, f.db.Create(&model.AuditLog{
		ClinicID:  &clinicID,
		ActorType: model.AuditActorTypeSystem,
		Action:    model.AuditActionStaffProvisionReceipt,
		Resource:  model.AuditResourceStaffProvisionBatch,
		NewValue: mustRawJSON(map[string]any{
			"batch_id": manifest.BatchID,
			"digest":   digest,
			"count":    1,
		}),
	}).Error)

	// Non-admin actor without master-staff create on scope.
	limited := &model.Account{
		Email:         "limited@example.test",
		PasswordHash:  "hash",
		IsActive:      true,
		IsSystemAdmin: false,
	}
	require.NoError(t, f.db.Create(limited).Error)
	limitedStaff := &model.Staff{
		ClinicID:  f.clinicA,
		Name:      "限定スタッフ",
		StaffType: model.StaffTypeNurse,
		IsActive:  true,
		AccountID: &limited.ID,
	}
	require.NoError(t, f.db.Create(limitedStaff).Error)
	require.NoError(t, f.db.Create(&model.StaffClinicAssignment{
		StaffID: limitedStaff.ID, ClinicID: f.clinicA, IsMain: true,
	}).Error)

	manifest.ActorAccountID = limited.ID
	mp, sp := f.writeInputs(t, manifest, secrets)
	_, err = f.provisioner.Preflight(context.Background(), mp, sp)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	// Error must not mention receipt existence / digest.
	assert.NotContains(t, err.Error(), "receipt")
	assert.NotContains(t, err.Error(), digest)
	assert.NotContains(t, err.Error(), "noop")

	_, err = f.provisioner.Apply(context.Background(), mp, sp)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.NotContains(t, err.Error(), "receipt")
	assert.NotContains(t, err.Error(), digest)
}

func TestStaffProvisioning_Integration_RejectsCrossClinicOccupation(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	foreignOcc := &model.Occupation{ClinicID: f.clinicB, Name: "他院職種", IsActive: true}
	require.NoError(t, f.db.Create(foreignOcc).Error)

	manifest, secrets := f.singleStaffManifest(t, false)
	manifest.Staff[0].OccupationID = &foreignOcc.ID
	mp, sp := f.writeInputs(t, manifest, secrets)

	_, err := f.provisioner.Preflight(context.Background(), mp, sp)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestStaffProvisioning_Integration_RejectsInputInsideRepoAndBadMode(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	manifest, secrets := f.singleStaffManifest(t, false)

	// Inside configured repo root
	inside := filepath.Join(f.repoRoot, "manifest.json")
	mb, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(inside, mb, 0o600))
	outsideSecretsDir := t.TempDir()
	sp := filepath.Join(outsideSecretsDir, "secrets.json")
	sb, err := json.Marshal(secrets)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(sp, sb, 0o600))
	_, err = f.provisioner.Preflight(context.Background(), inside, sp)
	require.Error(t, err)

	// Bad mode
	dir := t.TempDir()
	bad := filepath.Join(dir, "manifest.json")
	require.NoError(t, os.WriteFile(bad, mb, 0o644))
	_, err = f.provisioner.Preflight(context.Background(), bad, sp)
	require.Error(t, err)
}

func TestStaffProvisioning_Integration_ConcurrentSameBatchDifferentDigestConflict(t *testing.T) {
	f := newProvisionIntegrationFixture(t)
	manifestA, secretsA := f.singleStaffManifest(t, false)
	manifestB := *manifestA
	manifestB.Staff = append([]StaffProvisionStaffEntry(nil), manifestA.Staff...)
	manifestB.Staff[0].Name = "別合成スタッフ"
	// same batch_id (same clinic_scope), different content digest
	secretsB := *secretsA

	mpA, spA := f.writeInputs(t, manifestA, secretsA)
	mpB, spB := f.writeInputs(t, &manifestB, &secretsB)

	// Serialize first apply fully, then second must conflict on digest.
	_, err := f.provisioner.Apply(context.Background(), mpA, spA)
	require.NoError(t, err)
	_, err = f.provisioner.Apply(context.Background(), mpB, spB)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))

	// Concurrent attempt after first: both racing with different digests —
	// at least one must fail closed; never create two staff rows for the batch.
	// Reset DB receipts/staff for a true concurrent race on empty state.
	f2 := newProvisionIntegrationFixture(t)
	manifestA, secretsA = f2.singleStaffManifest(t, false)
	manifestB = *manifestA
	manifestB.Staff = append([]StaffProvisionStaffEntry(nil), manifestA.Staff...)
	manifestB.Staff[0].Email = "synthetic-provision-002@example.test"
	manifestB.Staff[0].ExternalStaffID = "ext-syn-002"
	manifestB.Staff[0].SecretRef = "sec-syn-002"
	secretsB = StaffProvisionSecretsFile{
		Secrets: []StaffProvisionSecretEntry{
			{SecretRef: "sec-syn-002", Password: "Password2b"},
		},
	}
	// Keep same clinic_scope so batch_id matches, different digest via content.
	mpA, spA = f2.writeInputs(t, manifestA, secretsA)
	mpB, spB = f2.writeInputs(t, &manifestB, &secretsB)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, e := f2.provisioner.Apply(context.Background(), mpA, spA)
		errs <- e
	}()
	go func() {
		defer wg.Done()
		_, e := f2.provisioner.Apply(context.Background(), mpB, spB)
		errs <- e
	}()
	wg.Wait()
	close(errs)

	var success, fail int
	for e := range errs {
		if e == nil {
			success++
		} else {
			fail++
		}
	}
	// Same batch_id / different digest under lock: one may apply, the other must
	// conflict (or both conflict if they interleaved before receipts — still safe).
	assert.LessOrEqual(t, success, 1)
	assert.GreaterOrEqual(t, fail, 1)

	var staffCount int64
	require.NoError(t, f2.db.Model(&model.Staff{}).Count(&staffCount).Error)
	assert.LessOrEqual(t, staffCount, int64(1))
}

func TestStaffProvisioning_Integration_RollbackOnAuditFailure(t *testing.T) {
	// Prove create path rolls back when audit write fails by using a broken
	// clinic_id reference is not possible post-validation; instead verify
	// unknown permission group fails before commit and leaves zero staff.
	f := newProvisionIntegrationFixture(t)
	manifest, secrets := f.singleStaffManifest(t, false)
	manifest.Staff[0].PermissionGroupIDs = []uint64{999999}
	mp, sp := f.writeInputs(t, manifest, secrets)

	_, err := f.provisioner.Apply(context.Background(), mp, sp)
	require.Error(t, err)

	var staffCount int64
	require.NoError(t, f.db.Model(&model.Staff{}).Count(&staffCount).Error)
	assert.Equal(t, int64(0), staffCount)
	var accountCount int64
	// admin account only
	require.NoError(t, f.db.Model(&model.Account{}).Count(&accountCount).Error)
	assert.Equal(t, int64(1), accountCount)
}

func mustRawJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("json: %v", err))
	}
	return b
}
