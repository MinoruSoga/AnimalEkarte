package medicalrecord

// lstep_eligibility_ambient_tx_test.go — EMR-330 適格性再検証 read の ambient tx 参加証明（DB-backed）
//
// delivery トリガーは候補抽出 → delivery-owner lock → claim/write 直前にイベント
// ソースを再読する。再読が DBOrTx で同一接続・同一スナップショットを使わなければ、
// 抽出後に死亡・削除されたペットを見逃す TOCTOU になる。ここでは ambient tx 内の
// 未コミット変更を read が観測することで参加を実証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestMedicalRecordRepository_OwnerHasAliveVisitBoundary_AmbientTransaction(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "visit boundary owner")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "visit boundary pet")
	target := time.Date(2026, 10, 1, 12, 0, 0, 0, time.Local)
	makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "VB-001", target)
	repo := NewMedicalRecordRepository(db)

	ambientTx := db.WithContext(ctx).Begin()
	require.NoError(t, ambientTx.Error)
	txCtx := persistence.WithTxValue(ctx, ambientTx)
	require.NoError(t, ambientTx.Exec(
		"UPDATE pets SET deceased_at = now(), status = ? WHERE id = ?",
		model.PetStatusDeceased, pet.ID,
	).Error)

	inTx, err := repo.ownerHasAliveVisitBoundary(txCtx, clinicA, owner.ID, target, false)
	require.NoError(t, err)
	assert.False(t, inTx, "ambient tx の未コミット死亡を観測し生存ペット由来カルテから除外する")

	outside, err := repo.ownerHasAliveVisitBoundary(ctx, clinicA, owner.ID, target, false)
	require.NoError(t, err)
	assert.True(t, outside, "コミット済み状態は tx 外の読み取りを変えない")

	require.NoError(t, ambientTx.Rollback().Error)
}

func TestVaccinationRepository_OwnerHasAliveVaccineDeadline_AmbientTransaction(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "vaccine deadline owner")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "vaccine deadline pet")
	vaccine := makeVaccineMaster(t, db, clinicA, "vaccine deadline vaccine")
	target := time.Date(2026, 10, 1, 12, 0, 0, 0, config.JST)
	rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)
	require.NoError(t, db.Exec("UPDATE vaccinations SET next_date = ? WHERE id = ?", target, rec.ID).Error)
	repo := NewVaccinationRepository(db)

	baseline, err := repo.OwnerHasAliveVaccineDeadline(ctx, clinicA, owner.ID, target)
	require.NoError(t, err)
	require.True(t, baseline)

	ambientTx := db.WithContext(ctx).Begin()
	require.NoError(t, ambientTx.Error)
	txCtx := persistence.WithTxValue(ctx, ambientTx)
	require.NoError(t, ambientTx.Exec(
		"UPDATE pets SET deceased_at = now(), status = ? WHERE id = ?",
		model.PetStatusDeceased, pet.ID,
	).Error)

	inTx, err := repo.OwnerHasAliveVaccineDeadline(txCtx, clinicA, owner.ID, target)
	require.NoError(t, err)
	assert.False(t, inTx, "ambient tx の未コミット死亡を観測し生存ペットの期限から除外する")

	outside, err := repo.OwnerHasAliveVaccineDeadline(ctx, clinicA, owner.ID, target)
	require.NoError(t, err)
	assert.True(t, outside, "コミット済み状態は tx 外の読み取りを変えない")

	require.NoError(t, ambientTx.Rollback().Error)
}
