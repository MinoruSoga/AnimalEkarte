package commonbasic

// hydrate_test.go — Hydrate* / ReadOwnerNames / Join* の契約テスト（AutoMigrate の double 上）。
//
// 検証する契約:
//   - binding → subject → 現行 revision を (clinic_id, ローカル ID) の組で解決し、現行 revision を載せる
//   - binding・subject・現行ポインタの欠落、別 subject の revision、他院の組は unresolved（422）になり、
//     1 件でも失敗すれば全体が失敗して引数の owner / pet は更新されない
//   - soft-delete 済みでも binding が残っていれば解決する
//   - pet は revision の animal_species_id から AnimalSpecies を載せ、マスタに無ければ unresolved
//
// 注: テスト DB は GORM AutoMigrate の double（SetupTestDB）。001_init.sql の deferred FK
// （fk_common_*_current_basic）・publish trigger・RLS が無いため、「別 subject の revision を指す」
// 「現行ポインタが無い」のように本番 DDL が commit 時に拒否する不整合状態を組み立てられる
// （本番 DDL ではその不整合は存在できず、reader のこの検査は defense in depth）。
// 本番 DDL での正常系と SQL の妥当性は hydrate_001_test.go が担う。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	testClinicA = uint64(1)
	testClinicB = uint64(2)
)

func setupDoubleDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{},
		&model.CommonOwnerSubject{}, &model.CommonOwnerClinicBinding{}, &model.CommonOwnerBasicRevision{},
		&model.CommonPetSubject{}, &model.CommonPetClinicBinding{}, &model.CommonPetBasicRevision{},
	))
	testdb.Truncate(t, db,
		"common_owner_basic_revisions", "common_owner_clinic_bindings", "common_owner_subjects",
		"common_pet_basic_revisions", "common_pet_clinic_bindings", "common_pet_subjects",
		"pets", "animal_species", "owners",
	)
	return db
}

func createOwner(t *testing.T, db *gorm.DB, clinicID uint64) *model.Owner {
	t.Helper()
	return testdb.MakeTestOwner(t, db, clinicID, "owner")
}

func createPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, tag string) *model.Pet {
	t.Helper()
	pet := &model.Pet{ClinicID: clinicID, OwnerID: ownerID, PetNumber: tag}
	require.NoError(t, db.Create(pet).Error)
	return pet
}

func createSpecies(t *testing.T, db *gorm.DB, name string) *model.AnimalSpecies {
	t.Helper()
	species := &model.AnimalSpecies{Name: name}
	require.NoError(t, db.Create(species).Error)
	return species
}

// bindOwner は owner の subject と binding だけを作る（revision は作らない）。
func bindOwner(t *testing.T, db *gorm.DB, clinicID uint64, owner *model.Owner) uuid.UUID {
	t.Helper()
	subjectID := uuid.New()
	require.NoError(t, db.Create(&model.CommonOwnerSubject{ID: subjectID}).Error)
	require.NoError(t, db.Create(&model.CommonOwnerClinicBinding{ClinicID: clinicID, CommonOwnerID: subjectID, LocalOwnerID: owner.ID, ConfirmedBy: 1, ConfirmedAt: time.Now(), OperationID: 1}).Error)
	return subjectID
}

// addOwnerRevision は subject に revision 行を追加する（現行ポインタは動かさない）。
func addOwnerRevision(t *testing.T, db *gorm.DB, subjectID uuid.UUID, clinicID, ownerID uint64, revision int64, name string) *model.CommonOwnerBasicRevision {
	t.Helper()
	row := &model.CommonOwnerBasicRevision{}
	row.CommonOwnerID = subjectID
	row.Revision = revision
	row.SourceClinicID = clinicID
	row.SourceLocalOwnerID = ownerID
	row.OperationID = 1
	row.Name = name
	row.Provenance = json.RawMessage(`{}`)
	row.RecordedBy = 1
	row.RecordedAt = time.Now()
	require.NoError(t, db.Create(row).Error)
	return row
}

// publishOwnerRevision は subject の現行ポインタを revisionID へ向ける（本番では publish trigger が commit 時に行う）。
func publishOwnerRevision(t *testing.T, db *gorm.DB, subjectID uuid.UUID, revisionID uint64) {
	t.Helper()
	require.NoError(t, db.Model(&model.CommonOwnerSubject{}).
		Where("id = ?", subjectID).
		Update("current_basic_revision_id", revisionID).Error)
}

// linkOwner は subject + binding + revision 1 を作り、現行ポインタを revision 1 に向ける。
func linkOwner(t *testing.T, db *gorm.DB, clinicID uint64, owner *model.Owner, name string) (uuid.UUID, *model.CommonOwnerBasicRevision) {
	t.Helper()
	subjectID := bindOwner(t, db, clinicID, owner)
	revision := addOwnerRevision(t, db, subjectID, clinicID, owner.ID, 1, name)
	publishOwnerRevision(t, db, subjectID, revision.ID)
	return subjectID, revision
}

// bindPet は pet の subject と binding だけを作る（revision は作らない）。
func bindPet(t *testing.T, db *gorm.DB, clinicID uint64, pet *model.Pet) uuid.UUID {
	t.Helper()
	subjectID := uuid.New()
	require.NoError(t, db.Create(&model.CommonPetSubject{ID: subjectID}).Error)
	require.NoError(t, db.Create(&model.CommonPetClinicBinding{ClinicID: clinicID, CommonPetID: subjectID, LocalPetID: pet.ID, ConfirmedBy: 1, ConfirmedAt: time.Now(), OperationID: 1}).Error)
	return subjectID
}

// addPetRevision は subject に revision 行を追加する（現行ポインタは動かさない）。
func addPetRevision(t *testing.T, db *gorm.DB, subjectID uuid.UUID, clinicID, petID, speciesID uint64, revision int64, name string) *model.CommonPetBasicRevision {
	t.Helper()
	status := model.PetStatusAlive
	row := &model.CommonPetBasicRevision{}
	row.CommonPetID = subjectID
	row.Revision = revision
	row.SourceClinicID = clinicID
	row.SourceLocalPetID = petID
	row.OperationID = 1
	row.Name = name
	row.AnimalSpeciesID = speciesID
	row.Status = &status
	row.NeuteredStatus = model.PetNeuteredStatusUnknown
	row.Provenance = json.RawMessage(`{}`)
	row.RecordedBy = 1
	row.RecordedAt = time.Now()
	require.NoError(t, db.Create(row).Error)
	return row
}

// publishPetRevision は subject の現行ポインタを revisionID へ向ける。
func publishPetRevision(t *testing.T, db *gorm.DB, subjectID uuid.UUID, revisionID uint64) {
	t.Helper()
	require.NoError(t, db.Model(&model.CommonPetSubject{}).
		Where("id = ?", subjectID).
		Update("current_basic_revision_id", revisionID).Error)
}

// linkPet は subject + binding + revision 1 を作り、現行ポインタを revision 1 に向ける。
func linkPet(t *testing.T, db *gorm.DB, clinicID uint64, pet *model.Pet, speciesID uint64, name string) (uuid.UUID, *model.CommonPetBasicRevision) {
	t.Helper()
	subjectID := bindPet(t, db, clinicID, pet)
	revision := addPetRevision(t, db, subjectID, clinicID, pet.ID, speciesID, 1, name)
	publishPetRevision(t, db, subjectID, revision.ID)
	return subjectID, revision
}

func TestHydrateOwners_ResolvesCurrentRevision(t *testing.T) {
	db := setupDoubleDB(t)
	owner := createOwner(t, db, testClinicA)
	subjectID, first := linkOwner(t, db, testClinicA, owner, "旧 氏名")
	second := addOwnerRevision(t, db, subjectID, testClinicA, owner.ID, 2, "現行 氏名")
	publishOwnerRevision(t, db, subjectID, second.ID)

	err := HydrateOwners(context.Background(), db, []uint64{testClinicA}, []*model.Owner{owner})

	require.NoError(t, err)
	require.NotNil(t, owner.Basic)
	assert.Equal(t, "現行 氏名", owner.Basic.Name)
	assert.Equal(t, second.ID, owner.Basic.ID)
	assert.NotEqual(t, first.ID, owner.Basic.ID)
	assert.Nil(t, owner.Basic.NameKana, "nullable revision columns stay nil, never blank")
	require.NotNil(t, owner.BasicSubject)
	assert.Equal(t, subjectID, owner.BasicSubject.CommonID)
	assert.Equal(t, second.ID, owner.BasicSubject.CurrentRevisionID)
	assert.EqualValues(t, 1, owner.BasicSubject.Version)
}

func TestHydrateOwners_UnboundOwnerFailsClosed(t *testing.T) {
	db := setupDoubleDB(t)
	owner := createOwner(t, db, testClinicA)

	err := HydrateOwners(context.Background(), db, []uint64{testClinicA}, []*model.Owner{owner})

	requireUnresolved(t, err)
	assert.Nil(t, owner.Basic)
	assert.Nil(t, owner.BasicSubject)
}

func TestHydrateOwners_RevisionOfAnotherSubjectFailsClosed(t *testing.T) {
	db := setupDoubleDB(t)
	owner := createOwner(t, db, testClinicA)
	other := createOwner(t, db, testClinicA)
	subjectID, _ := linkOwner(t, db, testClinicA, owner, "本人")
	_, otherRevision := linkOwner(t, db, testClinicA, other, "他人")
	// 現行ポインタが別 subject の revision を指す不整合状態。
	publishOwnerRevision(t, db, subjectID, otherRevision.ID)

	err := HydrateOwners(context.Background(), db, []uint64{testClinicA}, []*model.Owner{owner})

	requireUnresolved(t, err)
	assert.Nil(t, owner.Basic, "another subject's data must never be loaded")
}

func TestHydrateOwners_SubjectWithoutCurrentRevisionFailsClosed(t *testing.T) {
	db := setupDoubleDB(t)
	owner := createOwner(t, db, testClinicA)
	bindOwner(t, db, testClinicA, owner)

	err := HydrateOwners(context.Background(), db, []uint64{testClinicA}, []*model.Owner{owner})

	requireUnresolved(t, err)
	assert.Nil(t, owner.Basic)
}

func TestHydrateOwners_OtherClinicIsNotResolved(t *testing.T) {
	db := setupDoubleDB(t)
	ownerB := createOwner(t, db, testClinicB)
	linkOwner(t, db, testClinicB, ownerB, "B 医院の飼い主")
	ctx := context.Background()

	// 認可範囲外の医院の owner は binding があっても解決しない。
	requireUnresolved(t, HydrateOwners(ctx, db, []uint64{testClinicA}, []*model.Owner{ownerB}))
	// 医院 ID を詐称しても (clinic, ローカル ID) の組の binding が無いので解決しない。
	forged := &model.Owner{ID: ownerB.ID, ClinicID: testClinicA}
	requireUnresolved(t, HydrateOwners(ctx, db, []uint64{testClinicA}, []*model.Owner{forged}))
	assert.Nil(t, forged.Basic)
	// 正しい医院の範囲では解決する。
	require.NoError(t, HydrateOwners(ctx, db, []uint64{testClinicB}, []*model.Owner{ownerB}))
	assert.Equal(t, "B 医院の飼い主", ownerB.Basic.Name)
}

func TestHydrateOwners_OneUnresolvedFailsTheWholeCall(t *testing.T) {
	db := setupDoubleDB(t)
	bound := createOwner(t, db, testClinicA)
	unbound := createOwner(t, db, testClinicA)
	linkOwner(t, db, testClinicA, bound, "束縛済み")

	err := HydrateOwners(context.Background(), db, []uint64{testClinicA}, []*model.Owner{bound, unbound})

	requireUnresolved(t, err)
	assert.Nil(t, bound.Basic, "no partial update when any requested owner is unresolved")
	assert.Nil(t, unbound.Basic)
}

func TestHydrateOwners_SoftDeletedOwnerStillResolves(t *testing.T) {
	db := setupDoubleDB(t)
	owner := createOwner(t, db, testClinicA)
	linkOwner(t, db, testClinicA, owner, "削除済み飼い主")
	require.NoError(t, db.Delete(owner).Error)
	var deleted model.Owner
	require.NoError(t, db.Unscoped().First(&deleted, owner.ID).Error)
	require.True(t, deleted.DeletedAt.Valid)

	err := HydrateOwners(context.Background(), db, []uint64{testClinicA}, []*model.Owner{&deleted})

	require.NoError(t, err)
	assert.Equal(t, "削除済み飼い主", deleted.Basic.Name)
}

func TestHydratePets_ResolvesRevisionAndAnimalSpecies(t *testing.T) {
	db := setupDoubleDB(t)
	species := createSpecies(t, db, "犬")
	owner := createOwner(t, db, testClinicA)
	pet := createPet(t, db, testClinicA, owner.ID, "P-1")
	subjectID, revision := linkPet(t, db, testClinicA, pet, species.ID, "ポチ")

	err := HydratePets(context.Background(), db, []uint64{testClinicA}, []*model.Pet{pet})

	require.NoError(t, err)
	require.NotNil(t, pet.Basic)
	assert.Equal(t, "ポチ", pet.Basic.Name)
	assert.Equal(t, revision.ID, pet.Basic.ID)
	assert.Nil(t, pet.Basic.Breed, "nullable revision columns stay nil, never blank")
	require.NotNil(t, pet.BasicSubject)
	assert.Equal(t, subjectID, pet.BasicSubject.CommonID)
	require.NotNil(t, pet.AnimalSpecies)
	assert.Equal(t, species.ID, pet.AnimalSpecies.ID)
	assert.Equal(t, "犬", pet.AnimalSpecies.Name)
}

func TestHydratePets_MissingAnimalSpeciesFailsClosed(t *testing.T) {
	db := setupDoubleDB(t)
	owner := createOwner(t, db, testClinicA)
	pet := createPet(t, db, testClinicA, owner.ID, "P-2")
	const missingSpeciesID = uint64(999999)
	linkPet(t, db, testClinicA, pet, missingSpeciesID, "ポチ")

	err := HydratePets(context.Background(), db, []uint64{testClinicA}, []*model.Pet{pet})

	requireUnresolved(t, err)
	assert.Nil(t, pet.Basic)
	assert.Nil(t, pet.AnimalSpecies)
}

func TestHydratePets_UnboundPetFailsClosed(t *testing.T) {
	db := setupDoubleDB(t)
	owner := createOwner(t, db, testClinicA)
	pet := createPet(t, db, testClinicA, owner.ID, "P-3")

	err := HydratePets(context.Background(), db, []uint64{testClinicA}, []*model.Pet{pet})

	requireUnresolved(t, err)
	assert.Nil(t, pet.Basic)
}

func TestReadOwnerNames_ReturnsNamesAndFailsClosedOnUnbound(t *testing.T) {
	db := setupDoubleDB(t)
	first := createOwner(t, db, testClinicA)
	second := createOwner(t, db, testClinicA)
	unbound := createOwner(t, db, testClinicA)
	linkOwner(t, db, testClinicA, first, "一郎")
	linkOwner(t, db, testClinicA, second, "二郎")
	ctx := context.Background()

	names, err := ReadOwnerNames(ctx, db, testClinicA, []uint64{first.ID, second.ID, first.ID})
	require.NoError(t, err)
	assert.Equal(t, map[uint64]string{first.ID: "一郎", second.ID: "二郎"}, names)

	names, err = ReadOwnerNames(ctx, db, testClinicA, []uint64{first.ID, unbound.ID})
	requireUnresolved(t, err)
	assert.Nil(t, names)

	names, err = ReadOwnerNames(ctx, db, testClinicA, nil)
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestJoinCurrentBasic_InnerExcludesUnboundAndLeftKeepsIt(t *testing.T) {
	db := setupDoubleDB(t)
	bound := createOwner(t, db, testClinicA)
	unbound := createOwner(t, db, testClinicA)
	linkOwner(t, db, testClinicA, bound, "束縛済み")

	type joinedRow struct {
		ID   uint64  `gorm:"column:id"`
		Name *string `gorm:"column:name"`
	}
	query := func(left bool) []joinedRow {
		t.Helper()
		var rows []joinedRow
		tx := db.Table("owners").Select("owners.id AS id, rev.name AS name")
		require.NoError(t, JoinCurrentOwnerBasic(tx, "owners", "rev", left).Order("owners.id").Scan(&rows).Error)
		return rows
	}

	inner := query(false)
	require.Len(t, inner, 1)
	assert.Equal(t, bound.ID, inner[0].ID)
	require.NotNil(t, inner[0].Name)
	assert.Equal(t, "束縛済み", *inner[0].Name)

	outer := query(true)
	require.Len(t, outer, 2)
	assert.Equal(t, unbound.ID, outer[1].ID)
	assert.Nil(t, outer[1].Name, "an unbound owner stays visible with NULL revision columns")
}
