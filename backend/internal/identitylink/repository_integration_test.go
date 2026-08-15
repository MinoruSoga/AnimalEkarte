package identitylink

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupIdentityLinkTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Treatment{},
		&model.OwnerIdentityGroup{},
		&model.OwnerIdentityGroupMember{},
		&model.PetIdentityGroup{},
		&model.PetIdentityGroupMember{},
		&model.AuditLog{},
	))
	// Clean identity tables each test.
	_ = db.Exec("TRUNCATE TABLE pet_identity_group_members, pet_identity_groups, owner_identity_group_members, owner_identity_groups, treatments, medical_records, pets, owners, animal_species, audit_logs RESTART IDENTITY CASCADE").Error
	return db
}

func seedClinicOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
	t.Helper()
	return testdb.MakeTestOwner(t, db, clinicID, name)
}

func seedPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, name string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "dog-" + name}
	require.NoError(t, db.Create(species).Error)
	p := &model.Pet{ClinicID: clinicID, OwnerID: ownerID, AnimalSpeciesID: species.ID, Name: name}
	require.NoError(t, db.Create(p).Error)
	return p
}

func TestRepository_CreateOwnerGroupAndLastMemberUnlink(t *testing.T) {
	db := setupIdentityLinkTestDB(t)
	repo := NewRepository(db)
	tx := persistence.NewTransactor(db)
	auditRepo := audit.NewRepository(db)
	auditSvc := audit.NewService(auditRepo)
	svc := NewService(repo, tx, auditSvc)

	o1 := seedClinicOwner(t, db, 1, "OwnerA")
	o2 := seedClinicOwner(t, db, 2, "OwnerB")
	actor := testActor(1, 2)

	group, members, err := svc.CreateOwnerGroup(context.Background(), actor, []OwnerMemberRef{
		{ClinicID: 1, OwnerID: o1.ID},
		{ClinicID: 2, OwnerID: o2.ID},
	})
	require.NoError(t, err)
	require.NotZero(t, group.ID)
	assert.Len(t, members, 2)

	// concurrent-ish idempotent retry
	group2, members2, err := svc.CreateOwnerGroup(context.Background(), actor, []OwnerMemberRef{
		{ClinicID: 1, OwnerID: o1.ID},
		{ClinicID: 2, OwnerID: o2.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, group.ID, group2.ID)
	assert.Len(t, members2, 2)

	// Unlink first member
	err = svc.UnlinkOwnerMember(context.Background(), actor, group.ID, OwnerMemberRef{ClinicID: 1, OwnerID: o1.ID})
	require.NoError(t, err)

	// Last member unlink soft-deletes group
	err = svc.UnlinkOwnerMember(context.Background(), actor, group.ID, OwnerMemberRef{ClinicID: 2, OwnerID: o2.ID})
	require.NoError(t, err)

	var activeGroups int64
	require.NoError(t, db.Model(&model.OwnerIdentityGroup{}).Where("deleted_at IS NULL").Count(&activeGroups).Error)
	assert.Equal(t, int64(0), activeGroups)

	// No revive: re-link creates a new group
	group3, _, err := svc.CreateOwnerGroup(context.Background(), actor, []OwnerMemberRef{
		{ClinicID: 1, OwnerID: o1.ID},
		{ClinicID: 2, OwnerID: o2.ID},
	})
	require.NoError(t, err)
	assert.NotEqual(t, group.ID, group3.ID)

	// audit rows present
	var auditCount int64
	require.NoError(t, db.Model(&model.AuditLog{}).
		Where("resource = ?", model.AuditResourceIdentityLink).
		Count(&auditCount).Error)
	assert.GreaterOrEqual(t, auditCount, int64(3))
}

func TestRepository_MixedHiddenReject_NoPartialWrite(t *testing.T) {
	db := setupIdentityLinkTestDB(t)
	repo := NewRepository(db)
	tx := persistence.NewTransactor(db)
	svc := NewService(repo, tx, audit.NewService(audit.NewRepository(db)))

	o1 := seedClinicOwner(t, db, 1, "OwnerA")
	actor := testActor(1, 2)

	_, _, err := svc.CreateOwnerGroup(context.Background(), actor, []OwnerMemberRef{
		{ClinicID: 1, OwnerID: o1.ID},
		{ClinicID: 2, OwnerID: 99999}, // hidden / nonexistent
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var groups int64
	require.NoError(t, db.Model(&model.OwnerIdentityGroup{}).Count(&groups).Error)
	assert.Equal(t, int64(0), groups, "no partial write")

	var members int64
	require.NoError(t, db.Model(&model.OwnerIdentityGroupMember{}).Count(&members).Error)
	assert.Equal(t, int64(0), members)

	var audits int64
	require.NoError(t, db.Model(&model.AuditLog{}).Count(&audits).Error)
	assert.Equal(t, int64(0), audits)
}

func TestRepository_CrossClinicActorReject(t *testing.T) {
	db := setupIdentityLinkTestDB(t)
	repo := NewRepository(db)
	tx := persistence.NewTransactor(db)
	svc := NewService(repo, tx, audit.NewService(audit.NewRepository(db)))

	o1 := seedClinicOwner(t, db, 1, "OwnerA")
	o2 := seedClinicOwner(t, db, 2, "OwnerB")

	// Actor only clinic 1 — must reject entire create.
	_, _, err := svc.CreateOwnerGroup(context.Background(), testActor(1), []OwnerMemberRef{
		{ClinicID: 1, OwnerID: o1.ID},
		{ClinicID: 2, OwnerID: o2.ID},
	})
	require.Error(t, err)
	assert.True(t, errorsIsForbidden(err))

	var groups int64
	require.NoError(t, db.Model(&model.OwnerIdentityGroup{}).Count(&groups).Error)
	assert.Equal(t, int64(0), groups)
}

func TestRepository_LinkedTreatmentHistory_PairScope(t *testing.T) {
	db := setupIdentityLinkTestDB(t)
	repo := NewRepository(db)
	tx := persistence.NewTransactor(db)
	svc := NewService(repo, tx, audit.NewService(audit.NewRepository(db)))

	o1 := seedClinicOwner(t, db, 1, "OwnerA")
	o2 := seedClinicOwner(t, db, 2, "OwnerB")
	p1 := seedPet(t, db, 1, o1.ID, "Pochi")
	p2 := seedPet(t, db, 2, o2.ID, "Tama")
	// Unrelated pet in clinic 2 with same numeric id risk — create extra pet
	o3 := seedClinicOwner(t, db, 2, "Other")
	pOther := seedPet(t, db, 2, o3.ID, "OtherPet")

	actor := testActor(1, 2)
	ownerGroup, _, err := svc.CreateOwnerGroup(context.Background(), actor, []OwnerMemberRef{
		{ClinicID: 1, OwnerID: o1.ID},
		{ClinicID: 2, OwnerID: o2.ID},
	})
	require.NoError(t, err)

	_, _, err = svc.CreatePetGroup(context.Background(), actor, ownerGroup.ID, []PetMemberRef{
		{ClinicID: 1, PetID: p1.ID},
		{ClinicID: 2, PetID: p2.ID},
	})
	require.NoError(t, err)

	// History for linked pets + unrelated
	makeMR := func(clinicID, petID uint64, no string) *model.MedicalRecord {
		pet := petID
		mr := &model.MedicalRecord{
			ClinicID: clinicID,
			RecordNo: no,
			Date:     time.Now(),
			PetID:    &pet,
			Status:   model.MedicalRecordStatusFinalized,
		}
		require.NoError(t, db.Create(mr).Error)
		tr := &model.Treatment{MedicalRecordID: mr.ID, Content: "tx-" + no, ItemType: model.TreatmentItemTypeOther}
		require.NoError(t, db.Create(tr).Error)
		return mr
	}
	makeMR(1, p1.ID, "A1")
	makeMR(2, p2.ID, "B1")
	makeMR(2, pOther.ID, "OTHER")

	items, total, err := svc.ListLinkedTreatmentHistory(context.Background(), actor, 1, p1.ID, true, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	for _, it := range items {
		assert.NotEqual(t, pOther.ID, it.PetID, "unrelated pet must not appear via bare IN expansion")
	}
}

func errorsIsForbidden(err error) bool {
	return errors.Is(err, apperrors.ErrForbidden)
}
