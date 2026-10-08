package pet

// repository_lifecycle_or_contract_test.go — 死亡判定 OR 契約のライフサイクル write 前置条件（EMR-328）
//
// 正準契約: deceased = (status='deceased' OR deceased_at IS NOT NULL)。
// status='alive' かつ deceased_at≠NULL の半死亡行（データ訂正・移行・手作業で発生しうる）に対し、
// RecordDeath / ClearDeath / 旧ライフサイクル update が OR 契約と整合して動くことを実 DB で検証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func makeHalfDeadSpecies(t *testing.T, db *gorm.DB) uint64 {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	return species.ID
}

// RecordDeath: 半死亡行は既に死亡として扱い、conflict + deceased_at を上書きしない。
func TestPetRepository_RecordDeath_RejectsHalfDeadRow(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "半死亡飼主")
	speciesID := makeHalfDeadSpecies(t, db)
	pet := makeHalfDeadPet(t, db, clinicA, owner.ID, speciesID, "半死亡ペット")
	originalDeceasedAt := *pet.DeceasedAt

	repo := NewRepository(db)
	err := repo.RecordDeath(ctx, clinicA, pet.ID, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), "new reason")
	requireLifecycleConflict(t, err, deathAlreadyRecordedMessage)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	require.NotNil(t, reloaded.DeceasedAt)
	assert.True(t, originalDeceasedAt.Equal(*reloaded.DeceasedAt), "existing deceased_at must not be overwritten")
	assert.Equal(t, model.PetStatusAlive, reloaded.Status)
}

// RecordDeath: 旧形状（status=deceased + deceased_at NULL）も既存通り conflict（EMR-290 挙動維持）。
func TestPetRepository_RecordDeath_RejectsStatusOnlyDeceasedRow(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "旧死亡飼主")
	speciesID := makeHalfDeadSpecies(t, db)
	pet := makeStatusDeceasedPet(t, db, clinicA, owner.ID, speciesID, "旧死亡ペット", nil)

	repo := NewRepository(db)
	err := repo.RecordDeath(ctx, clinicA, pet.ID, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), "new reason")
	requireLifecycleConflict(t, err, deathAlreadyRecordedMessage)
}

// ClearDeath: 半死亡行を復活可能にする（status 単独では対象外だったため従来は復活不能だった）。
func TestPetRepository_ClearDeath_RevivesHalfDeadRow(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "復活飼主")
	speciesID := makeHalfDeadSpecies(t, db)
	pet := makeHalfDeadPet(t, db, clinicA, owner.ID, speciesID, "復活ペット")

	repo := NewRepository(db)
	require.NoError(t, repo.ClearDeath(ctx, clinicA, pet.ID))

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, reloaded.Status)
	assert.Nil(t, reloaded.DeceasedAt)
	assert.Nil(t, reloaded.DeceasedReason)
}

// ClearDeath: 旧形状（status=deceased + deceased_at NULL）の復活は従来通り成功する。
func TestPetRepository_ClearDeath_StatusOnlyDeceasedStillWorks(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "旧復活飼主")
	speciesID := makeHalfDeadSpecies(t, db)
	pet := makeStatusDeceasedPet(t, db, clinicA, owner.ID, speciesID, "旧復活ペット", nil)

	repo := NewRepository(db)
	require.NoError(t, repo.ClearDeath(ctx, clinicA, pet.ID))

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, reloaded.Status)
	assert.Nil(t, reloaded.DeceasedAt)
}

// 旧ライフサイクル update（deceased_at/deceased_reason/status の3キー同時更新）:
// UpdatePetInput は status を持たないため r.update を直接呼ぶ（BUG-415 で公開経路は閉塞済み）。
// 死亡方向は status='alive' AND deceased_at IS NULL のみ通し、半死亡行への重複死亡を拒否する。
func TestPetRepository_LegacyUpdate_DeathDirectionRejectsHalfDeadRow(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "旧経路死亡飼主")
	speciesID := makeHalfDeadSpecies(t, db)
	pet := makeHalfDeadPet(t, db, clinicA, owner.ID, speciesID, "旧経路死亡ペット")
	originalDeceasedAt := *pet.DeceasedAt

	repo := &repository{db: db}
	newDeceasedAt := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	reason := "legacy form"
	err := repo.update(ctx, clinicA, pet.ID, map[string]any{
		"deceased_at":     newDeceasedAt,
		"deceased_reason": reason,
		"status":          model.PetStatusDeceased,
	})
	requireLifecycleConflict(t, err, deathAlreadyRecordedMessage)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	require.NotNil(t, reloaded.DeceasedAt)
	assert.True(t, originalDeceasedAt.Equal(*reloaded.DeceasedAt))
}

// 旧ライフサイクル update: 生存行への死亡記録は従来通り成功する。
func TestPetRepository_LegacyUpdate_DeathDirectionAllowsAliveRow(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "旧経路生存飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "旧経路生存ペット")

	repo := &repository{db: db}
	newDeceasedAt := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	reason := "legacy form"
	err := repo.update(ctx, clinicA, pet.ID, map[string]any{
		"deceased_at":     newDeceasedAt,
		"deceased_reason": reason,
		"status":          model.PetStatusDeceased,
	})
	require.NoError(t, err)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusDeceased, reloaded.Status)
	require.NotNil(t, reloaded.DeceasedAt)
}

// 旧ライフサイクル update: 取り消し方向は status='deceased' OR deceased_at≠NULL を通す。
// 半死亡行（status=alive + deceased_at≠NULL）を復活させる。
func TestPetRepository_LegacyUpdate_RevivalDirectionAllowsHalfDeadRow(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "旧経路復活飼主")
	speciesID := makeHalfDeadSpecies(t, db)
	pet := makeHalfDeadPet(t, db, clinicA, owner.ID, speciesID, "旧経路復活ペット")

	repo := &repository{db: db}
	err := repo.update(ctx, clinicA, pet.ID, map[string]any{
		"deceased_at":     nil,
		"deceased_reason": nil,
		"status":          model.PetStatusAlive,
	})
	require.NoError(t, err)

	reloaded, err := repo.FindByID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, reloaded.Status)
	assert.Nil(t, reloaded.DeceasedAt)
}

// 旧ライフサイクル update: 完全生存行への取り消しは従来通り conflict。
func TestPetRepository_LegacyUpdate_RevivalDirectionRejectsAliveRow(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "旧経路生存飼主2")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "旧経路生存ペット2")

	repo := &repository{db: db}
	err := repo.update(ctx, clinicA, pet.ID, map[string]any{
		"deceased_at":     nil,
		"deceased_reason": nil,
		"status":          model.PetStatusAlive,
	})
	requireLifecycleConflict(t, err, revivalNotRecordedMessage)
}
