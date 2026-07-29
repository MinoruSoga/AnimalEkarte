package medicalrecord

// hospitalization_discharge_lock_concurrency_test.go — Discharge Q2-C の実DB並行性証明。
//
// 背景: DischargeWithBilling は tx 内 FindByID → Owner/Pet 再検証 → UpdateIfNotDischarged
// の順で動く。FindByID に行ロックが無いと、検証済みスナップショット取得後に別セッションが
// Owner/Pet 汚染や退院を割り込ませ、二重退院/汚染永続化の窓が開く。
// LockByIDForUpdate は FOR UPDATE で入院行をロックし、並行 UPDATE を直列化する。
//
// Discharge と同じ ctx-txKey 機構（withTx test helper = Transactor.WithTx 等価・prescription先例）で検証する（BE9-2D ⑤ Phase1 で
// hospitalizationService の repos.Transaction を WithTx 化したため、production 実機構へ追随。
// LockByIDForUpdate/UpdateIfNotDischarged は dbOrTx 化済み — 本テストが WithTx で green である
// こと自体が両メソッドの ambient tx 参加の実DB証明を兼ねる・④b X-11 テストと同型）。
//
// Callers: DischargeWithBilling（service）。API: HospitalizationRepository.LockByIDForUpdate。
// Schema: 変更なし（FOR UPDATE のみ）。
// User: 「GATE 1 承認。方針 FOR UPDATE（version 追加なし）… 既存 Discharge / clinic isolation 回帰を必ず残す」

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestHospitalizationRepository_LockByIDForUpdate(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "Q2-C Lock飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "Q2-C Lockポチ")
	hosp := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, func(h *model.Hospitalization) {
		h.Status = model.HospitalizationStatusAdmitted
	})

	t.Run("ambient tx 不在は fail-closed（MRB-02）", func(t *testing.T) {
		_, err := repo.LockByIDForUpdate(ctx, clinicA, hosp.ID)
		require.Error(t, err)
		assert.False(t, apperrors.IsNotFound(err), "tx 不在は NotFound ではなく internal であるべき: %v", err)
	})

	t.Run("同一医院からは行を取得し OwnerID/PetID スナップショットが空でない", func(t *testing.T) {
		require.NoError(t, withTx(ctx, db, func(txCtx context.Context) error {
			got, err := repo.LockByIDForUpdate(txCtx, clinicA, hosp.ID)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, hosp.ID, got.ID)
			assert.Equal(t, ownerA.ID, got.OwnerID)
			assert.Equal(t, petA.ID, got.PetID)
			assert.NotZero(t, got.OwnerID, "Q2-A 再検証用 OwnerID が空であってはならない")
			assert.NotZero(t, got.PetID, "Q2-A 再検証用 PetID が空であってはならない")
			assert.Equal(t, model.HospitalizationStatusAdmitted, got.Status)
			return nil
		}))
	})

	t.Run("他院からは NotFound（clinic_id 越境防止）", func(t *testing.T) {
		require.NoError(t, withTx(ctx, db, func(txCtx context.Context) error {
			_, err := repo.LockByIDForUpdate(txCtx, clinicB, hosp.ID)
			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
			return nil
		}))
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		require.NoError(t, withTx(ctx, db, func(txCtx context.Context) error {
			_, err := repo.LockByIDForUpdate(txCtx, clinicA, uint64(999999))
			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			return nil
		}))
	})
}

// TestHospitalizationRepository_LockByIDForUpdate_SerializesConcurrentDischargeUpdate は
// Lock 保持中に並行の UpdateIfNotDischarged がブロックされ、ロック解放後にのみ完了することを検証する。
func TestHospitalizationRepository_LockByIDForUpdate_SerializesConcurrentDischargeUpdate(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	hospRepo := NewHospitalizationRepository(db)
	ctx := context.Background()

	const clinicID = uint64(90201)
	owner := makeTestOwner(t, db, clinicID, "Q2-C 直列化飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "Q2-C 直列化ポチ")
	hosp := makeHospitalizationFixture(t, db, clinicID, owner.ID, pet.ID, func(h *model.Hospitalization) {
		h.Status = model.HospitalizationStatusAdmitted
	})

	lockAcquired := make(chan struct{})
	proceed := make(chan struct{})
	holderDone := make(chan struct{})
	var holderErr error
	var lockedOwnerID, lockedPetID uint64

	go func() {
		defer close(holderDone)
		holderErr = withTx(context.Background(), db, func(txCtx context.Context) error {
			locked, err := hospRepo.LockByIDForUpdate(txCtx, clinicID, hosp.ID)
			if err != nil {
				return err
			}
			lockedOwnerID = locked.OwnerID
			lockedPetID = locked.PetID
			close(lockAcquired)
			<-proceed
			_, err = hospRepo.UpdateIfNotDischarged(txCtx, clinicID, hosp.ID, map[string]any{
				"status":   model.HospitalizationStatusDischarged,
				"end_date": time.Now().UTC().Truncate(time.Second),
			})
			return err
		})
	}()

	<-lockAcquired
	assert.NotZero(t, lockedOwnerID, "Q2-A 再検証用 OwnerID が空であってはならない")
	assert.NotZero(t, lockedPetID, "Q2-A 再検証用 PetID が空であってはならない")
	assert.Equal(t, owner.ID, lockedOwnerID)
	assert.Equal(t, pet.ID, lockedPetID)

	updateStarted := make(chan struct{})
	updateDone := make(chan struct{})
	var updateErr error
	go func() {
		defer close(updateDone)
		close(updateStarted)
		_, updateErr = hospRepo.UpdateIfNotDischarged(context.Background(), clinicID, hosp.ID, map[string]any{
			"status":   model.HospitalizationStatusDischarged,
			"end_date": time.Now().UTC().Truncate(time.Second),
		})
	}()
	<-updateStarted

	select {
	case <-updateDone:
		t.Fatal("並行 UpdateIfNotDischarged が Lock 保持中に完了した（Q2-C 直列化が機能していない）")
	case <-time.After(300 * time.Millisecond):
	}

	close(proceed)
	<-holderDone
	<-updateDone

	require.NoError(t, holderErr, "ロック保持側の退院は成功するべき")
	require.Error(t, updateErr, "後続の退院は既に discharged のため失敗するべき")
	assert.True(t, apperrors.IsNotFound(updateErr), "後続は UpdateIfNotDischarged の NotFound: %v", updateErr)

	got, err := hospRepo.FindByID(ctx, clinicID, hosp.ID)
	require.NoError(t, err)
	assert.Equal(t, model.HospitalizationStatusDischarged, got.Status)
	assert.Equal(t, owner.ID, got.OwnerID, "OwnerID が汚染・消失していないこと")
	assert.Equal(t, pet.ID, got.PetID, "PetID が汚染・消失していないこと")
}

// TestHospitalizationRepository_LockByIDForUpdate_SerializesOwnerPetContamination は
// Lock 保持中に並行の Owner/Pet 更新がブロックされ、Discharge 完了後にのみ走ることを検証する。
func TestHospitalizationRepository_LockByIDForUpdate_SerializesOwnerPetContamination(t *testing.T) {
	db := setupHospitalizationRepoTestDB(t)
	hospRepo := NewHospitalizationRepository(db)

	const clinicID = uint64(90202)
	owner := makeTestOwner(t, db, clinicID, "Q2-C 汚染飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "Q2-C 汚染ポチ")
	foreignOwner := makeTestOwner(t, db, clinicID, "Q2-C 別飼主")
	foreignPet := makeSpeciesAndPet(t, db, clinicID, foreignOwner.ID, "Q2-C 別ポチ")
	hosp := makeHospitalizationFixture(t, db, clinicID, owner.ID, pet.ID, func(h *model.Hospitalization) {
		h.Status = model.HospitalizationStatusAdmitted
	})

	lockAcquired := make(chan struct{})
	proceed := make(chan struct{})
	holderDone := make(chan struct{})
	var holderErr error
	var lockedOwnerID, lockedPetID uint64

	go func() {
		defer close(holderDone)
		holderErr = withTx(context.Background(), db, func(txCtx context.Context) error {
			locked, err := hospRepo.LockByIDForUpdate(txCtx, clinicID, hosp.ID)
			if err != nil {
				return err
			}
			lockedOwnerID = locked.OwnerID
			lockedPetID = locked.PetID
			close(lockAcquired)
			<-proceed
			_, err = hospRepo.UpdateIfNotDischarged(txCtx, clinicID, hosp.ID, map[string]any{
				"status":   model.HospitalizationStatusDischarged,
				"end_date": time.Now().UTC().Truncate(time.Second),
			})
			return err
		})
	}()

	<-lockAcquired
	assert.Equal(t, owner.ID, lockedOwnerID)
	assert.Equal(t, pet.ID, lockedPetID)

	contamStarted := make(chan struct{})
	contamDone := make(chan struct{})
	var contamErr error
	go func() {
		defer close(contamDone)
		close(contamStarted)
		_, contamErr = hospRepo.Update(context.Background(), clinicID, hosp.ID, map[string]any{
			"owner_id": foreignOwner.ID,
			"pet_id":   foreignPet.ID,
		})
	}()
	<-contamStarted

	select {
	case <-contamDone:
		t.Fatal("並行 Owner/Pet 更新が Lock 保持中に完了した（汚染割り込み窓が開いている）")
	case <-time.After(300 * time.Millisecond):
	}

	close(proceed)
	<-holderDone
	<-contamDone

	require.NoError(t, holderErr)
	require.NoError(t, contamErr, "ロック解放後の Update 自体は成功しうる")

	got, err := hospRepo.FindByID(context.Background(), clinicID, hosp.ID)
	require.NoError(t, err)
	assert.Equal(t, model.HospitalizationStatusDischarged, got.Status)
}
