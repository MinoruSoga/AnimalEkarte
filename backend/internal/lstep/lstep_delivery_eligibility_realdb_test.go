package lstep

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-330 R2: 最終適格性再読は delivery-owner lock が握る同一
// トランザクション接続（ambient persistence.DBOrTx）で実行されなければ
// ならない。MaxOpenConns=1 の実接続プールでは、ロック内再読が
// r.db.WithContext で第 2 接続を取りに行くと永久にブロックされる —
// この bounded context はその場合に期限切れとなり fail-closed で返る。
// GREEN 側ではロックトランザクション上で再読が走り送信まで到達する。
func TestDeliveryOwnerLock_RealDB_EligibilityReReadUsesLockTx(t *testing.T) {
	shared := setupDeliveryOwnerLockDB(t)
	pool1 := openDeliveryPoolDB(t, 1)
	const clinicID = uint64(1)
	enableDeliverySync(t, shared, clinicID)
	owner := makeCleanupOwner(t, shared, clinicID, "elig-pool1-owner", "U_elig1", false)

	// 生存ペット由来の初診カルテを持つ飼い主（committed fixture）。
	visitDate := time.Now().AddDate(0, 0, -3)
	species := &model.AnimalSpecies{Name: "elig-pool1 species"}
	require.NoError(t, shared.Create(species).Error)
	eventPet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         owner.ID,
		AnimalSpeciesID: species.ID,
		Name:            "elig-pool1 pet",
		Status:          model.PetStatusAlive,
	}
	require.NoError(t, shared.Create(eventPet).Error)
	require.NoError(t, shared.Create(&model.MedicalRecord{
		ClinicID: clinicID,
		OwnerID:  &owner.ID,
		PetID:    &eventPet.ID,
		Date:     visitDate,
	}).Error)

	repo := &lstepTagCacheRepository{db: pool1}
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(pool1, repo, recorder.client())
	svc.medRecordRepo = medicalrecord.NewMedicalRecordRepository(pool1)
	elig := svc.aliveFirstVisitEligibility(clinicID, visitDate)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	done := make(chan deliveryAttemptResult, 1)
	go func() {
		fired, err := svc.processSingleOwner(ctx, recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", time.Now(), nil, elig)
		done <- deliveryAttemptResult{fired: fired, err: err}
	}()

	select {
	case res := <-done:
		require.NoError(t, res.err)
		assert.True(t, res.fired)
		assert.Equal(t, 1, recorder.attemptsFor("U_elig1"))
	case <-time.After(30 * time.Second):
		t.Fatal("delivery deadlocked with MaxOpenConns=1 during eligibility re-read")
	}

	log := deliveryTriggerLogFor(t, shared, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusFired, log.Status)
}
