package clinicale2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func testModels() []any {
	return []any{
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.PermissionGroup{},
		&model.PermissionGroupRule{},
		&model.StaffPermissionGroup{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.ExaminationType{},
		&model.Examination{},
		&model.Vaccine{},
		&model.Vaccination{},
		&model.CheckupType{},
		&model.Checkup{},
		&model.Hospitalization{},
		&model.Cage{},
		&model.LabDevice{},
		&model.LabDeviceItemMaster{},
		&model.Estimate{},
		&model.EstimateItem{},
		&model.AuditLog{},
	}
}

func ensureAuditLogForeignKeys(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'audit_logs_clinic_id_fkey'
			) THEN
				ALTER TABLE audit_logs
					ADD CONSTRAINT audit_logs_clinic_id_fkey
					FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;
			END IF;
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'audit_logs_actor_id_fkey'
			) THEN
				ALTER TABLE audit_logs
					ADD CONSTRAINT audit_logs_actor_id_fkey
					FOREIGN KEY (actor_id) REFERENCES staffs(id) ON DELETE RESTRICT;
			END IF;
		END $$;
	`).Error)
}

func TestCreate_RejectsUnsafeRequest(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()

	t.Run("staging env", func(t *testing.T) {
		_, err := Create(ctx, db, Request{AppEnv: "staging", DBHost: "db", PasswordHash: "x"})
		require.Error(t, err)
		assert.ErrorContains(t, err, "APP_ENV")
	})

	t.Run("empty password hash", func(t *testing.T) {
		_, err := Create(ctx, db, Request{AppEnv: "test", DBHost: "db"})
		require.Error(t, err)
		assert.ErrorContains(t, err, "password hash")
	})

	t.Run("reserved clinic remains rejected", func(t *testing.T) {
		require.Error(t, RejectReservedClinicID(1))
	})
}

func TestCreateAndDelete_DisposableClinicGraph(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, testModels()...))
	ensureAuditLogForeignKeys(t, db)
	ctx := context.Background()

	got, err := Create(ctx, db, Request{AppEnv: "test", DBHost: "db", PasswordHash: "test-hash-not-for-login"})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NoError(t, RejectReservedClinicID(got.ClinicID))
	assert.GreaterOrEqual(t, got.ClinicID, clinicIDBase)
	assert.Equal(t, ownerSearchToken, got.OwnerSearch)
	assert.Contains(t, got.OwnerName, ownerSearchToken)
	assert.NotZero(t, got.PetID)
	assert.NotZero(t, got.OutsideFirstPagePetID)
	assert.NotEqual(t, got.PetID, got.OutsideFirstPagePetID)
	assert.Contains(t, got.OutsideFirstPagePetName, outsideFirstPagePrefix)
	assert.Equal(t, medicalRecordCount, got.MedicalRecordCount)
	assert.Contains(t, got.EstimateTitle, "e2e-est-")

	var clinic model.Clinic
	require.NoError(t, db.First(&clinic, got.ClinicID).Error)
	assert.True(t, clinic.IsActive)

	var account model.Account
	require.NoError(t, db.Where("email = ?", LoginEmail(got.ClinicID)).First(&account).Error)
	assert.True(t, account.IsSystemAdmin)
	assert.Equal(t, "test-hash-not-for-login", account.PasswordHash)

	var pets []model.Pet
	require.NoError(t, db.Where("clinic_id = ?", got.ClinicID).Order("id ASC").Find(&pets).Error)
	require.Len(t, pets, padPetCount+2)
	assert.Equal(t, got.PetID, pets[0].ID)
	assert.Equal(t, got.OutsideFirstPagePetID, pets[len(pets)-1].ID)

	var records []model.MedicalRecord
	require.NoError(t, db.Where("clinic_id = ?", got.ClinicID).Find(&records).Error)
	require.Len(t, records, medicalRecordCount)
	for _, record := range records {
		assert.Equal(t, model.MedicalRecordStatusFinalized, record.Status)
	}

	var examinations []model.Examination
	require.NoError(t, db.Where("clinic_id = ?", got.ClinicID).Find(&examinations).Error)
	require.Len(t, examinations, 2)
	var linkedExams, standaloneExams int
	for _, e := range examinations {
		if e.MedicalRecordID != nil {
			linkedExams++
			require.NotNil(t, e.PetID)
			assert.Equal(t, got.PetID, *e.PetID)
		} else {
			standaloneExams++
			require.NotNil(t, e.PetID)
			assert.Equal(t, got.OutsideFirstPagePetID, *e.PetID)
		}
	}
	assert.Equal(t, 1, linkedExams)
	assert.Equal(t, 1, standaloneExams)

	var vaccinations []model.Vaccination
	require.NoError(t, db.Where("clinic_id = ?", got.ClinicID).Find(&vaccinations).Error)
	require.Len(t, vaccinations, 2)
	var linkedVaccinations, standaloneVaccinations int
	for _, v := range vaccinations {
		if v.MedicalRecordID != nil {
			linkedVaccinations++
			require.NotNil(t, v.PetID)
			assert.Equal(t, got.PetID, *v.PetID)
		} else {
			standaloneVaccinations++
			require.NotNil(t, v.PetID)
			assert.Equal(t, got.OutsideFirstPagePetID, *v.PetID)
		}
	}
	assert.Equal(t, 1, linkedVaccinations)
	assert.Equal(t, 1, standaloneVaccinations)

	var hospitalizations []model.Hospitalization
	require.NoError(t, db.Where("clinic_id = ?", got.ClinicID).Find(&hospitalizations).Error)
	require.Len(t, hospitalizations, 1)
	assert.Equal(t, model.HospitalizationStatusAdmitted, hospitalizations[0].Status)

	encoded, err := EncodeResult(got)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "password")
	assert.NotContains(t, string(encoded), account.PasswordHash)
	var decoded Result
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, got.ClinicID, decoded.ClinicID)

	require.NoError(t, Delete(ctx, db, "test", "db", got.ClinicID))
	var leftover model.Clinic
	err = db.First(&leftover, got.ClinicID).Error
	require.Error(t, err)
	var leftoverPets int64
	require.NoError(t, db.Model(&model.Pet{}).Where("clinic_id = ?", got.ClinicID).Count(&leftoverPets).Error)
	assert.Zero(t, leftoverPets)
}

func TestDelete_RemovesOnlyFixtureStaffAuditLogs(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, testModels()...))
	ensureAuditLogForeignKeys(t, db)
	ctx := context.Background()

	fixture, err := Create(ctx, db, Request{AppEnv: "test", DBHost: "db", PasswordHash: "test-hash-not-for-login"})
	require.NoError(t, err)

	var fixtureStaff model.Staff
	require.NoError(t, db.Where("clinic_id = ?", fixture.ClinicID).First(&fixtureStaff).Error)
	fixtureAuditLog := &model.AuditLog{
		ClinicID:  &fixture.ClinicID,
		ActorID:   &fixtureStaff.ID,
		ActorType: model.AuditActorTypeStaff,
		Action:    model.AuditActionAuthLoginSuccess,
		Resource:  model.AuditResourceStaff,
	}
	require.NoError(t, db.Create(fixtureAuditLog).Error)
	foreignCompany := &model.Company{Name: "foreign-audit-company"}
	require.NoError(t, db.Create(foreignCompany).Error)
	foreignClinicID := fixture.ClinicID + 100000
	foreignClinic := &model.Clinic{
		ID:        foreignClinicID,
		CompanyID: foreignCompany.ID,
		Name:      "foreign-audit-clinic",
		IsActive:  true,
	}
	require.NoError(t, db.Create(foreignClinic).Error)
	foreignStaff := &model.Staff{
		ClinicID:  foreignClinicID,
		Name:      "foreign-audit-staff",
		IsActive:  true,
		StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(foreignStaff).Error)
	foreignAuditLog := &model.AuditLog{
		ClinicID:  &foreignClinicID,
		ActorID:   &foreignStaff.ID,
		ActorType: model.AuditActorTypeStaff,
		Action:    model.AuditActionAuthLoginSuccess,
		Resource:  model.AuditResourceStaff,
	}
	require.NoError(t, db.Create(foreignAuditLog).Error)

	require.NoError(t, Delete(ctx, db, "test", "db", fixture.ClinicID))

	var fixtureAuditCount int64
	require.NoError(t, db.Model(&model.AuditLog{}).Where("id = ?", fixtureAuditLog.ID).Count(&fixtureAuditCount).Error)
	assert.Zero(t, fixtureAuditCount)

	var remainingForeignAudit model.AuditLog
	require.NoError(t, db.First(&remainingForeignAudit, foreignAuditLog.ID).Error)
	assert.Equal(t, foreignClinicID, *remainingForeignAudit.ClinicID)
	assert.Equal(t, foreignStaff.ID, *remainingForeignAudit.ActorID)
}

func TestDelete_RejectsReservedAndMismatchedClinic(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Company{}, &model.Clinic{}))
	ctx := context.Background()

	err := Delete(ctx, db, "test", "db", 1)
	require.Error(t, err)
	assert.ErrorContains(t, err, "reserved")

	company := &model.Company{Name: "not-e2e-company"}
	require.NoError(t, db.Create(company).Error)
	foreign := &model.Clinic{
		ID:        clinicIDBase + uint64(time.Now().UnixNano()%7000) + 50,
		CompanyID: company.ID,
		Name:      "foreign-clinic",
		IsActive:  true,
	}
	require.NoError(t, db.Create(foreign).Error)
	err = Delete(ctx, db, "test", "db", foreign.ID)
	require.Error(t, err)
	assert.ErrorContains(t, err, "not a clinical e2e fixture")
	var still model.Clinic
	require.NoError(t, db.First(&still, foreign.ID).Error)
}

func TestCreateAndDelete_V04ViewOnlyAccount(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, testModels()...))
	ensureAuditLogForeignKeys(t, db)
	ctx := context.Background()

	const hash = "test-hash-not-for-login"
	got, err := Create(ctx, db, Request{AppEnv: "test", DBHost: "db", PasswordHash: hash})
	require.NoError(t, err)
	require.NotNil(t, got)
	clinicID := got.ClinicID

	assert.Equal(t, ViewOnlyLoginEmail(clinicID), got.V04.ViewOnlyEmail)
	assert.NotEqual(t, LoginEmail(clinicID), got.V04.ViewOnlyEmail)

	var viewAccount model.Account
	require.NoError(t, db.Where("email = ?", got.V04.ViewOnlyEmail).First(&viewAccount).Error)
	assert.False(t, viewAccount.IsSystemAdmin)
	assert.True(t, viewAccount.IsActive)
	assert.Equal(t, hash, viewAccount.PasswordHash)

	var viewStaff model.Staff
	require.NoError(t, db.Where("clinic_id = ? AND account_id = ?", clinicID, viewAccount.ID).First(&viewStaff).Error)
	assert.Equal(t, model.StaffTypeNurse, viewStaff.StaffType)
	assert.True(t, viewStaff.IsActive)

	var assignment model.StaffClinicAssignment
	require.NoError(t, db.Where("staff_id = ? AND clinic_id = ?", viewStaff.ID, clinicID).First(&assignment).Error)
	assert.True(t, assignment.IsMain)

	var links []model.StaffPermissionGroup
	require.NoError(t, db.Where("staff_id = ?", viewStaff.ID).Find(&links).Error)
	require.Len(t, links, 1)

	var group model.PermissionGroup
	require.NoError(t, db.Where("id = ? AND clinic_id = ?", links[0].GroupID, clinicID).First(&group).Error)
	assert.Equal(t, v04GroupName, group.Name)
	assert.True(t, group.IsActive)

	var rules []model.PermissionGroupRule
	require.NoError(t, db.Where("group_id = ?", group.ID).Order("id ASC").Find(&rules).Error)
	require.Len(t, rules, len(v04ViewOnlyResources))
	wantResources := make([]string, 0, len(v04ViewOnlyResources))
	for _, resource := range v04ViewOnlyResources {
		wantResources = append(wantResources, string(resource))
	}
	assert.Equal(t, wantResources, got.V04.ViewOnlyResources)
	excluded := map[string]bool{
		string(model.ResourceAccounting):              true,
		string(model.ResourceAccountingCancel):        true,
		string(model.ResourceAccountingPostCloseEdit): true,
		string(model.ResourceAccountingReports):       true,
		string(model.ResourceCashRegisterClose):       true,
		string(model.ResourceLstepCsvImport):          true,
		string(model.ResourceLstepAnalytics):          true,
		string(model.ResourceMasterPermission):        true,
		string(model.ResourceMasterStaff):             true,
	}
	ruleResources := make([]string, 0, len(rules))
	for _, rule := range rules {
		assert.True(t, rule.CanView, rule.Resource)
		assert.False(t, rule.CanCreate, rule.Resource)
		assert.False(t, rule.CanEdit, rule.Resource)
		assert.False(t, rule.CanDelete, rule.Resource)
		assert.True(t, model.IsValidResource(rule.Resource), rule.Resource)
		assert.False(t, excluded[rule.Resource], rule.Resource)
		ruleResources = append(ruleResources, rule.Resource)
	}
	assert.ElementsMatch(t, wantResources, ruleResources)

	assert.Equal(t, fmt.Sprintf("%scage-%d", v04RowPrefix, clinicID), got.V04.CageName)
	var cage model.Cage
	require.NoError(t, db.Where("clinic_id = ? AND name = ?", clinicID, got.V04.CageName).First(&cage).Error)
	assert.True(t, cage.IsActive)
	assert.Equal(t, model.CageTypeGeneral, cage.CageType)
	assert.Equal(t, model.CageSizeMedium, cage.CageSize)

	assert.Equal(t, fmt.Sprintf("%sdevice-%d", v04RowPrefix, clinicID), got.V04.LabDeviceName)
	var device model.LabDevice
	require.NoError(t, db.Where("clinic_id = ? AND name = ?", clinicID, got.V04.LabDeviceName).First(&device).Error)
	assert.True(t, device.IsActive)
	assert.Equal(t, string(model.LabImportSourceTypeFujiNX600), device.SourceType)
	assert.Nil(t, device.ExamTypeID)

	encoded, err := EncodeResult(got)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "password")
	assert.NotContains(t, string(encoded), "hash")
	assert.Equal(t, 1, strings.Count(string(encoded), `"clinicId"`))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	v04raw, ok := decoded["v04"].(map[string]any)
	require.True(t, ok, "encoded JSON must carry a v04 object")
	_, nestedClinicID := v04raw["clinicId"]
	assert.False(t, nestedClinicID, "v04 object must not contain a clinicId key")

	// teardown は V04 行だけでなく device item master（ensure 相当の孤立行）も回収する。
	require.NoError(t, db.Create(&model.LabDeviceItemMaster{
		ClinicID:       clinicID,
		SourceType:     string(model.LabImportSourceTypeFujiNX600),
		DeviceItemCode: "V04-ITEM",
		ValueShape:     model.LabDeviceValueShapeNumeric,
	}).Error)

	require.NoError(t, Delete(ctx, db, "test", "db", clinicID))

	var count int64
	require.NoError(t, db.Unscoped().Model(&model.Account{}).
		Where("email IN ?", []string{LoginEmail(clinicID), ViewOnlyLoginEmail(clinicID)}).
		Count(&count).Error)
	assert.Zero(t, count, "accounts")
	for _, table := range []struct {
		name  string
		model any
	}{
		{"staffs", &model.Staff{}},
		{"staff_clinic_assignments", &model.StaffClinicAssignment{}},
		{"permission_groups", &model.PermissionGroup{}},
		{"cages", &model.Cage{}},
		{"lab_devices", &model.LabDevice{}},
		{"lab_device_item_masters", &model.LabDeviceItemMaster{}},
	} {
		var leftover int64
		require.NoError(t, db.Unscoped().Model(table.model).Where("clinic_id = ?", clinicID).Count(&leftover).Error)
		assert.Zero(t, leftover, table.name)
	}
	var leftoverRules int64
	require.NoError(t, db.Unscoped().Model(&model.PermissionGroupRule{}).Where("group_id = ?", group.ID).Count(&leftoverRules).Error)
	assert.Zero(t, leftoverRules, "permission_group_rules")
	var leftoverLinks int64
	require.NoError(t, db.Model(&model.StaffPermissionGroup{}).Where("staff_id = ?", viewStaff.ID).Count(&leftoverLinks).Error)
	assert.Zero(t, leftoverLinks, "staff_permission_groups")
}
