package commonbasic

// hydrate_001_test.go — 実 001_init.sql 上の reader 検証。
//
// AutoMigrate の double（hydrate_test.go）は deferred FK・publish trigger・RLS を持たない。本ファイルは
// 001 だけを適用した使い捨て DB に、本番と同じ制約下で subject・binding・revision を作り（operation の
// 成功 receipt を同一 tx に置き、commit 時に publish trigger が現行ポインタを動かす）、reader が本番 DDL の
// 列名・JOIN 条件で動くことを確認する。「別 subject の revision を指す」状態は本番 DDL が commit で
// 拒否するため、ここでは作れず hydrate_test.go が担う。

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// real001Seed は seedReal001 が作った行の ID。
type real001Seed struct {
	clinicA         uint64
	clinicB         uint64
	boundOwnerID    uint64
	unboundOwnerID  uint64
	otherOwnerID    uint64
	boundPetID      uint64
	unboundPetID    uint64
	speciesID       uint64
	ownerSubjectID  uuid.UUID
	ownerRevisionID uint64
}

// seedReal001 は 2 医院分の fixture を 1 tx で作って commit する。
// 医院 A: 束縛済み owner / pet と未束縛の owner / pet。医院 B: 束縛済み owner。
func seedReal001(t *testing.T, db *gorm.DB) real001Seed {
	t.Helper()
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer func() { _ = tx.Rollback() }()

	insertID := func(query string, args ...any) uint64 {
		t.Helper()
		var id int64
		require.NoError(t, tx.Raw(query, args...).Scan(&id).Error)
		return uint64(id)
	}
	companyID := insertID(`INSERT INTO companies (name) VALUES ('hydrate-001') RETURNING id`)
	newClinic := func(name string) (clinicID, staffID, operationID uint64) {
		clinicID = insertID(`INSERT INTO clinics (company_id, name) VALUES (?, ?) RETURNING id`, companyID, name)
		staffID = insertID(`INSERT INTO staffs (clinic_id, name) VALUES (?, ?) RETURNING id`, clinicID, name)
		operationID = insertID(`INSERT INTO operation_bindings (clinic_id, owner_kind, actor_kind, actor_ref, intent, operation_key, fingerprint, fingerprint_key_version) VALUES (?, 'owner', 'staff', 'hydrate-001', 'hydrate-001', gen_random_uuid(), decode(repeat('00', 32), 'hex'), 'v1') RETURNING id`, clinicID)
		// operation_success_required() / common_basic_publish() が要求する success receipt を同じ tx に置く。
		require.NoError(t, tx.Exec(`INSERT INTO operation_results (clinic_id, binding_id, result_refs, committed_versions, effect_summary) VALUES (?, ?, '{}', '{}', '{}')`, clinicID, operationID).Error)
		return clinicID, staffID, operationID
	}
	linkOwner := func(clinicID, staffID, operationID, ownerID uint64, name string) (uuid.UUID, uint64) {
		subjectID := uuid.New()
		require.NoError(t, tx.Exec(`INSERT INTO common_owner_subjects (id) VALUES (?)`, subjectID).Error)
		require.NoError(t, tx.Exec(`INSERT INTO common_owner_clinic_bindings (clinic_id, common_owner_id, local_owner_id, confirmed_by, confirmed_at, operation_id) VALUES (?, ?, ?, ?, now(), ?)`, clinicID, subjectID, ownerID, staffID, operationID).Error)
		revisionID := insertID(`INSERT INTO common_owner_basic_revisions (common_owner_id, revision, source_clinic_id, source_local_owner_id, operation_id, name, provenance, recorded_by) VALUES (?, 1, ?, ?, ?, ?, '{}'::jsonb, ?) RETURNING id`, subjectID, clinicID, ownerID, operationID, name, staffID)
		return subjectID, revisionID
	}
	linkPet := func(clinicID, staffID, operationID, petID, speciesID uint64, name string) {
		subjectID := uuid.New()
		require.NoError(t, tx.Exec(`INSERT INTO common_pet_subjects (id) VALUES (?)`, subjectID).Error)
		require.NoError(t, tx.Exec(`INSERT INTO common_pet_clinic_bindings (clinic_id, common_pet_id, local_pet_id, confirmed_by, confirmed_at, operation_id) VALUES (?, ?, ?, ?, now(), ?)`, clinicID, subjectID, petID, staffID, operationID).Error)
		insertID(`INSERT INTO common_pet_basic_revisions (common_pet_id, revision, source_clinic_id, source_local_pet_id, operation_id, name, animal_species_id, provenance, recorded_by) VALUES (?, 1, ?, ?, ?, ?, ?, '{}'::jsonb, ?) RETURNING id`, subjectID, clinicID, petID, operationID, name, speciesID, staffID)
	}

	var seed real001Seed
	var staffA, operationA, staffB, operationB uint64
	seed.clinicA, staffA, operationA = newClinic("hydrate-001-a")
	seed.clinicB, staffB, operationB = newClinic("hydrate-001-b")
	seed.speciesID = insertID(`INSERT INTO animal_species (name) VALUES ('hydrate-001') RETURNING id`)

	seed.boundOwnerID = insertID(`INSERT INTO owners (clinic_id) VALUES (?) RETURNING id`, seed.clinicA)
	seed.unboundOwnerID = insertID(`INSERT INTO owners (clinic_id) VALUES (?) RETURNING id`, seed.clinicA)
	seed.otherOwnerID = insertID(`INSERT INTO owners (clinic_id) VALUES (?) RETURNING id`, seed.clinicB)
	seed.boundPetID = insertID(`INSERT INTO pets (clinic_id, owner_id) VALUES (?, ?) RETURNING id`, seed.clinicA, seed.boundOwnerID)
	seed.unboundPetID = insertID(`INSERT INTO pets (clinic_id, owner_id) VALUES (?, ?) RETURNING id`, seed.clinicA, seed.unboundOwnerID)

	seed.ownerSubjectID, seed.ownerRevisionID = linkOwner(seed.clinicA, staffA, operationA, seed.boundOwnerID, "real-001 owner")
	linkOwner(seed.clinicB, staffB, operationB, seed.otherOwnerID, "real-001 other clinic owner")
	linkPet(seed.clinicA, staffA, operationA, seed.boundPetID, seed.speciesID, "real-001 pet")

	require.NoError(t, tx.Commit().Error)
	return seed
}

func TestReal001_HydrateOwnersAndPets(t *testing.T) {
	db := testdb.SetupIsolated001TestDB(t)
	seed := seedReal001(t, db)
	ctx := context.Background()

	t.Run("bound owner and pet resolve their current revision", func(t *testing.T) {
		owner := &model.Owner{ID: seed.boundOwnerID, ClinicID: seed.clinicA}
		pet := &model.Pet{ID: seed.boundPetID, ClinicID: seed.clinicA, OwnerID: seed.boundOwnerID}

		require.NoError(t, HydrateOwners(ctx, db, []uint64{seed.clinicA}, []*model.Owner{owner}))
		require.NoError(t, HydratePets(ctx, db, []uint64{seed.clinicA}, []*model.Pet{pet}))

		assert.Equal(t, "real-001 owner", owner.Basic.Name)
		assert.Equal(t, seed.ownerSubjectID, owner.BasicSubject.CommonID)
		assert.Equal(t, seed.ownerRevisionID, owner.BasicSubject.CurrentRevisionID)
		assert.Positive(t, owner.BasicSubject.Version)
		assert.Equal(t, "real-001 pet", pet.Basic.Name)
		require.NotNil(t, pet.AnimalSpecies)
		assert.Equal(t, seed.speciesID, pet.AnimalSpecies.ID)
	})

	t.Run("unbound owner and pet fail closed", func(t *testing.T) {
		owner := &model.Owner{ID: seed.unboundOwnerID, ClinicID: seed.clinicA}
		pet := &model.Pet{ID: seed.unboundPetID, ClinicID: seed.clinicA, OwnerID: seed.unboundOwnerID}

		requireUnresolved(t, HydrateOwners(ctx, db, []uint64{seed.clinicA}, []*model.Owner{owner}))
		requireUnresolved(t, HydratePets(ctx, db, []uint64{seed.clinicA}, []*model.Pet{pet}))
		assert.Nil(t, owner.Basic)
		assert.Nil(t, pet.Basic)
	})

	t.Run("another clinic's owner is resolved only inside its own clinic scope", func(t *testing.T) {
		owner := &model.Owner{ID: seed.otherOwnerID, ClinicID: seed.clinicB}

		requireUnresolved(t, HydrateOwners(ctx, db, []uint64{seed.clinicA}, []*model.Owner{owner}))
		require.NoError(t, HydrateOwners(ctx, db, []uint64{seed.clinicB}, []*model.Owner{owner}))
		assert.Equal(t, "real-001 other clinic owner", owner.Basic.Name)
	})

	t.Run("join helpers produce valid SQL against 001", func(t *testing.T) {
		type joinedRow struct {
			ID   uint64  `gorm:"column:id"`
			Name *string `gorm:"column:name"`
		}
		scan := func(table string, join func(*gorm.DB, string, string, bool) *gorm.DB, left bool) []joinedRow {
			t.Helper()
			var rows []joinedRow
			tx := db.Table(table).Select(table + ".id AS id, rev.name AS name")
			require.NoError(t, join(tx, table, "rev", left).Order(table+".id").Scan(&rows).Error)
			return rows
		}

		assert.Len(t, scan("owners", JoinCurrentOwnerBasic, false), 2)
		assert.Len(t, scan("owners", JoinCurrentOwnerBasic, true), 3)
		inner := scan("pets", JoinCurrentPetBasic, false)
		require.Len(t, inner, 1)
		assert.Equal(t, seed.boundPetID, inner[0].ID)
		outer := scan("pets", JoinCurrentPetBasic, true)
		require.Len(t, outer, 2)
		assert.Nil(t, outer[1].Name, "an unbound pet stays visible with NULL revision columns")
	})

	// owners を更新するため最後に実行する。
	t.Run("soft-deleted owner still resolves its name", func(t *testing.T) {
		require.NoError(t, db.Exec(`UPDATE owners SET deleted_at = now() WHERE id = ?`, seed.boundOwnerID).Error)

		names, err := ReadOwnerNames(ctx, db, seed.clinicA, []uint64{seed.boundOwnerID})

		require.NoError(t, err)
		assert.Equal(t, "real-001 owner", names[seed.boundOwnerID])
	})
}
