package reservation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupNextVisitIntentTestDB provisions the shared test schema plus the partial
// unique index from migration 026 (testdb is a GORM double — migration indexes
// are not replayed, so the gate creates it idempotently).
func setupNextVisitIntentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupReservationRepoTestDB(t)
	// staff_clinic_assignments.clinic_id → clinics.id の FK を満たすため、
	// 会社/医院テーブルも揃える（医院行は fixture 側で OnConflict 作成する）。
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Company{}, &model.Clinic{}))
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_appointments_one_active_tentative
		ON appointments (clinic_id, pet_id)
		WHERE status = 'tentative' AND deleted_at IS NULL`).Error)
	return db
}

func setNextVisitIntentInTx(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationNextVisitIntentRepository,
	clinicID, petID uint64,
	input SetNextVisitIntentInput,
) (*model.Reservation, error) {
	// この helper は repo の upsert 機構を対象にするため edit 権限相当を常に許可する。
	// CanUpdateExisting=false の拒否自体は専用テストで固定する。
	input.CanUpdateExisting = true
	var got *model.Reservation
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		got, _, err = repo.SetNextVisitIntent(txCtx, clinicID, petID, input)
		return err
	})
	return got, err
}

func clearNextVisitIntentInTx(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationNextVisitIntentRepository,
	clinicID, petID uint64,
) (bool, error) {
	var found bool
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		found, err = repo.ClearNextVisitIntent(txCtx, clinicID, petID)
		return err
	})
	return found, err
}

func confirmTentativeInTx(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationNextVisitIntentRepository,
	clinicID, id uint64,
	input ConfirmTentativeUpdate,
) (*model.Reservation, error) {
	var got *model.Reservation
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		got, err = repo.ConfirmTentative(txCtx, clinicID, id, input)
		return err
	})
	return got, err
}

// makeNextVisitIntentFixture builds owner+pet+general type for one clinic.
// 医院行は staff_clinic_assignments→clinics FK のため fixture 呼出毎に idempotent に用意する
// （他テストの TRUNCATE 後でも自己修復する）。
func makeNextVisitIntentFixture(t *testing.T, db *gorm.DB, clinicID uint64, label string) (*model.Owner, *model.Pet, *model.ReservationType) {
	t.Helper()
	makeReservationClinicCorrelationClinic(t, db, clinicID, "次回予定"+label+"医院")
	owner := testdb.MakeTestOwner(t, db, clinicID, "次回予定"+label+"飼主")
	pet := makePetForReservationIntentTest(t, db, clinicID, owner.ID, "次回予定"+label+"ペット")
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "次回予定" + label + "一般区分",
		Category: model.ReservationTypeCategoryGeneral,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	return owner, pet, reservationType
}

func makeMedicalRecordForNextVisitIntentTest(t *testing.T, db *gorm.DB, clinicID, ownerID, petID uint64, label string) *model.MedicalRecord {
	t.Helper()
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: fmt.Sprintf("MR-INTENT-%s", label),
		Date:     time.Now().UTC(),
		OwnerID:  &ownerID,
		PetID:    &petID,
	}
	require.NoError(t, db.Create(record).Error)
	return record
}

// futureIntentDateJST は JST 暦日で daysAhead 日後の 00:00 を返す。
// repo 層が intent lock 内で日付窓（過去・2年上限）を再検証する（Codex #19）
// ため、固定の遠未来日は使わず実行日基準で生成する。
func futureIntentDateJST(daysAhead int) time.Time {
	nowJST := time.Now().In(config.JST)
	return time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, config.JST).AddDate(0, 0, daysAhead)
}

func countTentativeForPet(t *testing.T, db *gorm.DB, clinicID, petID uint64) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("clinic_id = ? AND pet_id = ? AND status = ? AND deleted_at IS NULL",
			clinicID, petID, model.ReservationStatusTentative).
		Count(&count).Error)
	return count
}

func TestReservationRepository_SetNextVisitIntent(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA, petA, typeA := makeNextVisitIntentFixture(t, db, clinicA, "A")
	intendedDate := futureIntentDateJST(30)

	t.Run("requires an ambient transaction", func(t *testing.T) {
		got, _, err := repo.SetNextVisitIntent(ctx, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, petA.ID))
	})

	t.Run("creates a tentative appointment with JST sentinel and provenance", func(t *testing.T) {
		record := makeMedicalRecordForNextVisitIntentTest(t, db, clinicA, ownerA.ID, petA.ID, "SRC")
		got, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{
			IntendedDate:          intendedDate,
			SourceMedicalRecordID: &record.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, model.ReservationStatusTentative, got.Status)
		assert.Equal(t, clinicA, got.ClinicID)
		require.NotNil(t, got.PetID)
		assert.Equal(t, petA.ID, *got.PetID)
		require.NotNil(t, got.OwnerID)
		assert.Equal(t, ownerA.ID, *got.OwnerID)
		assert.Equal(t, typeA.ID, got.ReservationTypeID)
		assert.Nil(t, got.DoctorID)
		require.NotNil(t, got.ReservationRoute)
		assert.Equal(t, "record_intent", *got.ReservationRoute)
		require.NotNil(t, got.IntendedDate)
		assert.Equal(t, intendedDate.Format(time.DateOnly), got.IntendedDate.In(config.JST).Format(time.DateOnly))
		require.NotNil(t, got.SourceMedicalRecordID)
		assert.Equal(t, record.ID, *got.SourceMedicalRecordID)

		// Sentinel: JST 00:00–00:15 of the intended date.
		startJST := got.StartTime.In(config.JST)
		endJST := got.EndTime.In(config.JST)
		assert.Equal(t, intendedDate.Format("2006-01-02 15:04"), startJST.Format("2006-01-02 15:04"))
		assert.Equal(t, intendedDate.Add(15*time.Minute).Format("2006-01-02 15:04"), endJST.Format("2006-01-02 15:04"))
	})

	t.Run("updates the existing tentative instead of creating a second row", func(t *testing.T) {
		existing, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		newDate := futureIntentDateJST(45)
		updated, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: newDate})
		require.NoError(t, err)
		assert.Equal(t, existing.ID, updated.ID)
		require.NotNil(t, updated.IntendedDate)
		assert.Equal(t, newDate.Format(time.DateOnly), updated.IntendedDate.In(config.JST).Format(time.DateOnly))
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicA, petA.ID))
	})

	t.Run("rejects overwrite without update permission (CanUpdateExisting=false)", func(t *testing.T) {
		// Codex 指摘: PUT intent は upsert — 既存行の更新は reservations:edit 相当を
		// 要求する（create のみのロールが他人の intent を上書きできないよう fail-closed）。
		pet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "権限拒否ペット")
		created, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		var got *model.Reservation
		newDate := intendedDate.AddDate(0, 0, 7)
		err = testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			var err error
			got, _, err = repo.SetNextVisitIntent(txCtx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: newDate})
			return err
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrForbidden), "既存 intent 上書きは edit 権限なしで 403")
		assert.Nil(t, got)
		var reloaded model.Reservation
		require.NoError(t, db.First(&reloaded, created.ID).Error)
		require.NotNil(t, reloaded.IntendedDate)
		assert.Equal(t, intendedDate.Format("2006-01-02"),
			reloaded.IntendedDate.In(config.JST).Format(time.DateOnly),
			"拒否された上書きは intended_date を変更しない")
	})

	t.Run("refreshes owner_id on upsert after pet transfer and reports previous owner", func(t *testing.T) {
		// Codex 指摘: ペット譲渡後の upsert で owner_id を追従させ、旧 owner を
		// post-commit タグ再同期へ返す。
		pet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "譲渡ペット")
		_, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		newOwner := testdb.MakeTestOwner(t, db, clinicA, "譲渡先飼主")
		require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pet.ID).Update("owner_id", newOwner.ID).Error)

		var updated *model.Reservation
		var prevOwner *uint64
		err = testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			var err error
			updated, prevOwner, err = repo.SetNextVisitIntent(txCtx, clinicA, pet.ID, SetNextVisitIntentInput{
				IntendedDate:      intendedDate.AddDate(0, 0, 3),
				CanUpdateExisting: true,
			})
			return err
		})
		require.NoError(t, err)
		require.NotNil(t, updated.OwnerID)
		assert.Equal(t, newOwner.ID, *updated.OwnerID, "upsert は現 owner を書き込む")
		require.NotNil(t, prevOwner)
		assert.Equal(t, ownerA.ID, *prevOwner, "旧 owner を返してタグ再同期できる")
	})

	t.Run("rejects a pet owned by another clinic", func(t *testing.T) {
		_, _, _ = makeNextVisitIntentFixture(t, db, clinicB, "B")
		got, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, 999999, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
	})

	t.Run("rejects a source medical record owned by another pet", func(t *testing.T) {
		otherPet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "次回予定別ペット")
		record := makeMedicalRecordForNextVisitIntentTest(t, db, clinicA, ownerA.ID, otherPet.ID, "WRONGPET")
		got, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{
			IntendedDate:          intendedDate,
			SourceMedicalRecordID: &record.ID,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
	})

	t.Run("rejects a source medical record owned by another clinic", func(t *testing.T) {
		ownerB, petB, _ := makeNextVisitIntentFixture(t, db, clinicB, "B2")
		recordB := makeMedicalRecordForNextVisitIntentTest(t, db, clinicB, ownerB.ID, petB.ID, "FOREIGN")
		got, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{
			IntendedDate:          intendedDate,
			SourceMedicalRecordID: &recordB.ID,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
	})

	t.Run("rolls back with the caller transaction", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回予定rollback飼主")
		pet := makePetForReservationIntentTest(t, db, clinicA, owner.ID, "次回予定rollbackペット")
		sentinel := errors.New("rollback intent write")
		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if _, _, err := repo.SetNextVisitIntent(txCtx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate}); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, pet.ID))
	})

	t.Run("concurrent sets produce exactly one tentative row", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回予定競合飼主")
		pet := makePetForReservationIntentTest(t, db, clinicA, owner.ID, "次回予定競合ペット")
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for i := range 2 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				date := intendedDate.AddDate(0, 0, i)
				_, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: date})
				errs <- err
			}(i)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			require.NoError(t, err)
		}
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicA, pet.ID))
	})

	t.Run("partial unique index rejects a second raw tentative row", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回予定一意飼主")
		pet := makePetForReservationIntentTest(t, db, clinicA, owner.ID, "次回予定一意ペット")
		_, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		// Bypass the service/repository lock path with a raw INSERT to prove the
		// database-level guard is live under this schema.
		dup := &model.Reservation{
			ClinicID:          clinicA,
			ReservationTypeID: typeA.ID,
			StartTime:         intendedDate,
			EndTime:           intendedDate.Add(15 * time.Minute),
			OwnerID:           &owner.ID,
			PetID:             &pet.ID,
			Status:            model.ReservationStatusTentative,
			Source:            model.ReservationSourceManual,
			CustomerFields:    []byte("{}"),
		}
		createErr := db.Create(dup).Error
		require.Error(t, createErr)
		assert.True(t, persistence.IsUniqueConstraintErr(createErr))
	})

	t.Run("revalidates the date after an actor FOR SHARE wait crosses a JST boundary", func(t *testing.T) {
		// Codex #19 rereview: CreatedBy 付きの新規 intent は Create 内部で staff/
		// assignment の FOR SHARE を取る。最終日付検証より後の actor ロック待ちで
		// JST 暦日を跨ぐと前日付の intended_date を INSERT し得た。actor 検証を
		// 最終検証より前へ移した経路を2接続 barrier で固定する。
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "BARRIER")
		actor := makeDoctor(t, db, clinicA, "次回予定actor待機獣医")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: actor.ID, ClinicID: clinicA, IsMain: true}).Error)

		// 検証は注入時計の JST 暦日を見る。fake day D で開始し、actor FOR SHARE の
		// 待機中に D+1 へ進める（実時計ではなく注入時計だけを跨がせる）。
		var clockMu sync.Mutex
		fakeNow := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC) // = 21:00 JST on day D
		repoImpl := &reservationRepository{db: db}
		repoImpl.nowFn = func() time.Time {
			clockMu.Lock()
			defer clockMu.Unlock()
			return fakeNow
		}
		intendedDate := time.Date(2026, 6, 15, 0, 0, 0, 0, config.JST) // 注入時計の「今日」

		// blocker 接続が staffs 行を FOR UPDATE で保持し、actor 検証の FOR SHARE を待機させる。
		blocker := db.Begin()
		require.NoError(t, blocker.Error)
		t.Cleanup(func() { _ = blocker.Rollback() })
		require.NoError(t, blocker.Exec("SELECT id FROM staffs WHERE id = ? FOR UPDATE", actor.ID).Error)

		writeDone := make(chan error, 1)
		go func() {
			writeDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				_, _, err := repoImpl.SetNextVisitIntent(txCtx, clinicA, pet.ID, SetNextVisitIntentInput{
					IntendedDate:      intendedDate,
					CreatedBy:         &actor.ID,
					CanUpdateExisting: true,
				})
				return err
			})
		}()

		// 対象 goroutine が staffs FOR SHARE で Lock 待機するまでポーリングする。
		deadline := time.After(15 * time.Second)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for locked := false; !locked; {
			select {
			case err := <-writeDone:
				t.Fatalf("set intent finished before reaching the actor lock wait: %v", err)
			case <-deadline:
				t.Fatal("set intent did not reach the staff FOR SHARE wait")
			case <-ticker.C:
				var waiters int64
				require.NoError(t, db.Raw(`
					SELECT count(*) FROM pg_stat_activity
					WHERE datname = current_database()
					  AND wait_event_type = 'Lock'
					  AND query LIKE '%"staffs"%FOR SHARE%'
				`).Scan(&waiters).Error)
				locked = waiters > 0
			}
		}

		// ロック待機中に JST 暦日を跨がせてから blocker を解放する。
		clockMu.Lock()
		fakeNow = fakeNow.AddDate(0, 0, 1)
		clockMu.Unlock()
		require.NoError(t, blocker.Commit().Error)

		err := <-writeDone
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err),
			"actor ロック待ちで日跨ぎした intent は InvalidInput、実際: %v", err)
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, pet.ID),
			"日跨ぎ後の INSERT は行われない")
	})
}

func TestReservationRepository_ClearNextVisitIntent(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	intendedDate := futureIntentDateJST(30)

	t.Run("requires an ambient transaction", func(t *testing.T) {
		_, petA, _ := makeNextVisitIntentFixture(t, db, clinicA, "CLR")
		found, err := repo.ClearNextVisitIntent(ctx, clinicA, petA.ID)
		require.Error(t, err)
		assert.False(t, found)
	})

	t.Run("soft-deletes the active tentative and is idempotent", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "CLR1")
		created, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		found, err := clearNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, pet.ID))

		var row model.Reservation
		require.NoError(t, db.Unscoped().First(&row, created.ID).Error)
		assert.NotNil(t, row.DeletedAt)
		// 終了はキャンセル経路と同じ終端 semantics: sentinel status のまま消すのではなく
		// status=cancelled を記録してから論理削除する（provenance は行に残る）。
		assert.Equal(t, model.ReservationStatusCancelled, row.Status)
		require.NotNil(t, row.IntendedDate)
		assert.Equal(t, intendedDate.Format("2006-01-02"), row.IntendedDate.Format("2006-01-02"))

		found, err = clearNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID)
		require.NoError(t, err)
		assert.False(t, found, "second clear must be an idempotent no-op")
	})

	t.Run("does not touch another clinic's tentative", func(t *testing.T) {
		_, petA, _ := makeNextVisitIntentFixture(t, db, clinicA, "CLR2A")
		_, petB, _ := makeNextVisitIntentFixture(t, db, clinicB, "CLR2B")
		_, err := setNextVisitIntentInTx(ctx, db, repo, clinicB, petB.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		// Clearing via clinicA + petB's id: the find predicate is clinic-scoped, so nothing deletes.
		found, err := clearNextVisitIntentInTx(ctx, db, repo, clinicA, petB.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicB, petB.ID))

		// Sanity: clinicA pet unaffected on a real clear.
		_, err = setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)
		found, err = clearNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID)
		require.NoError(t, err)
		assert.True(t, found)
	})

	t.Run("leaves non-tentative appointments untouched", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "CLR3")
		confirmed := makeReservationForReservationRepoTest(
			t, db, clinicA, makeReservationType(t, db, clinicA).ID,
			time.Now().UTC().Add(24*time.Hour),
			model.ReservationStatusConfirmed, model.ReservationSourceManual, nil, nil,
		)
		found, err := clearNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Equal(t, model.ReservationStatusConfirmed, reloadReservationIntentStatus(t, db, confirmed.ID))
	})
}

func TestReservationRepository_UpdateTentativeStatus(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	intendedDate := futureIntentDateJST(30)

	t.Run("writes on a tentative row and reloads", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "UTS1")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		now := time.Now().UTC()
		updated, err := repo.updateTentativeStatus(ctx, clinicA, tentative.ID, map[string]any{
			"status":        model.ReservationStatusCheckedIn,
			"checked_in_at": now,
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, model.ReservationStatusCheckedIn, updated.Status)
	})

	t.Run("rejects a row already promoted by a concurrent confirm", func(t *testing.T) {
		// TOCTOU 再現: FindByID で tentative を見た後、ConfirmTentative が先に
		// confirmed へ commit したケース。status=tentative 述語で RowsAffected=0 →
		// conflict になり、checked_in/cancelled での上書きを防ぐ。
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "UTS2")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		promoted := tentative
		promoted.Status = model.ReservationStatusConfirmed
		require.NoError(t, db.Model(&model.Reservation{}).
			Where("id = ?", tentative.ID).
			Update("status", model.ReservationStatusConfirmed).Error)

		got, err := repo.updateTentativeStatus(ctx, clinicA, tentative.ID, map[string]any{
			"status": model.ReservationStatusCancelled,
		}, nil)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsConflict(err), "昇格済み行への tentative CAS は 409")
		assert.Equal(t, model.ReservationStatusConfirmed, reloadReservationIntentStatus(t, db, tentative.ID),
			"confirmed 昇格を cancelled で上書きしない")
	})

	t.Run("does not write across clinic boundary", func(t *testing.T) {
		_, petA, _ := makeNextVisitIntentFixture(t, db, clinicA, "UTS3A")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		got, err := repo.updateTentativeStatus(ctx, uint64(2), tentative.ID, map[string]any{
			"status": model.ReservationStatusCheckedIn,
		}, nil)
		require.Error(t, err)
		assert.Nil(t, got)
		// lock 用の tx 内 FindByID が先に走るため跨医院行は 404（ConfirmTentative と同契約）。
		assert.True(t, apperrors.IsNotFound(err))
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("invokes the post-lock validator on the reloaded row and merges its fields", func(t *testing.T) {
		// Codex #3/#19: validator は intent lock 取得後・CAS 前に再読み込みした行へ
		// 適用される（ロック待ち中にコミットされた mutation を反映する）。
		// validator が返す追加フィールドは CAS の fields へマージされて書き込まれる。
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "UTS4")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		var sawLocked *model.Reservation
		checkinStamp := time.Now()
		updated, err := repo.updateTentativeStatus(ctx, clinicA, tentative.ID, map[string]any{
			"status": model.ReservationStatusCheckedIn,
		}, func(_ context.Context, locked *model.Reservation) (map[string]any, error) {
			sawLocked = locked
			return map[string]any{
				"checked_in_at": checkinStamp,
				"start_time":    checkinStamp,
			}, nil
		})
		require.NoError(t, err)
		require.NotNil(t, sawLocked, "validator が post-lock で再読み込みした行を受け取る")
		assert.Equal(t, tentative.ID, sawLocked.ID)
		assert.Equal(t, model.ReservationStatusCheckedIn, updated.Status)
		require.NotNil(t, updated.CheckedInAt)
		assert.True(t, updated.CheckedInAt.Equal(checkinStamp) || updated.CheckedInAt.After(checkinStamp.Add(-time.Second)),
			"validator 由来の checked_in_at が CAS へマージされる")
		assert.True(t, updated.StartTime.Equal(checkinStamp) || updated.StartTime.After(checkinStamp.Add(-time.Second)))
	})

	t.Run("validator error aborts the CAS write and leaves the row tentative", func(t *testing.T) {
		// Codex #14: 未来日拒否等の validator エラーは tx ごとロールバックし、
		// tentative 行を消費しない。
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "UTS5")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		wantErr := apperrors.WrapInvalidInput("未来日の次回予定は受付できません")
		got, err := repo.updateTentativeStatus(ctx, clinicA, tentative.ID, map[string]any{
			"status": model.ReservationStatusCheckedIn,
		}, func(_ context.Context, _ *model.Reservation) (map[string]any, error) {
			return nil, wantErr
		})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID),
			"validator 拒否時は行が tentative のまま残る")
	})
}

func TestReservationRepository_ConfirmTentative(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	intendedDate := futureIntentDateJST(30)

	_, petA, typeA := makeNextVisitIntentFixture(t, db, clinicA, "CFM")

	// 確定日時は intended_date と同じ JST 暦日の日中に置く。
	newStart := intendedDate.Add(10 * time.Hour)
	newEnd := newStart.Add(30 * time.Minute)

	t.Run("requires an ambient transaction", func(t *testing.T) {
		got, err := repo.ConfirmTentative(ctx, clinicA, 1, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("transitions a tentative appointment to a real booking", func(t *testing.T) {
		record := makeMedicalRecordForNextVisitIntentTest(t, db, clinicA, petA.ID, petA.ID, "CFMSRC")
		_ = record
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		doctor := makeDoctor(t, db, clinicA, "次回予定確定獣医")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicA, IsMain: true}).Error)

		confirmed, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			DoctorID:          &doctor.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.NoError(t, err)
		require.NotNil(t, confirmed)
		assert.Equal(t, model.ReservationStatusConfirmed, confirmed.Status)
		assert.True(t, confirmed.StartTime.Equal(newStart))
		assert.True(t, confirmed.EndTime.Equal(newEnd))
		require.NotNil(t, confirmed.DoctorID)
		assert.Equal(t, doctor.ID, *confirmed.DoctorID)
		// provenance columns survive the transition.
		require.NotNil(t, confirmed.IntendedDate)
		assert.Equal(t, intendedDate.Format(time.DateOnly), confirmed.IntendedDate.In(config.JST).Format(time.DateOnly))
		require.NotNil(t, confirmed.ReservationRoute)
		assert.Equal(t, "record_intent", *confirmed.ReservationRoute)
	})

	t.Run("rejects a non-tentative appointment", func(t *testing.T) {
		normal := makeReservationForReservationRepoTest(
			t, db, clinicA, typeA.ID, time.Now().UTC().Add(48*time.Hour),
			model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
		)
		got, err := confirmTentativeInTx(ctx, db, repo, clinicA, normal.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, normal.ID))
	})

	t.Run("rejects a tentative owned by another clinic", func(t *testing.T) {
		_, petB, typeB := makeNextVisitIntentFixture(t, db, clinicB, "CFMB")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicB, petB.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		got, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeB.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("rejects a doctor assigned only to another clinic", func(t *testing.T) {
		owner2 := testdb.MakeTestOwner(t, db, clinicA, "次回予定確定飼主2")
		pet2 := makePetForReservationIntentTest(t, db, clinicA, owner2.ID, "次回予定確定ペット2")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet2.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		doctorB := makeDoctor(t, db, clinicB, "次回予定確定獣医B")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctorB.ID, ClinicID: clinicB, IsMain: true}).Error)

		got, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			DoctorID:          &doctorB.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("rejects a target status outside pending/confirmed", func(t *testing.T) {
		owner3 := testdb.MakeTestOwner(t, db, clinicA, "次回予定確定飼主3")
		pet3 := makePetForReservationIntentTest(t, db, clinicA, owner3.ID, "次回予定確定ペット3")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet3.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		got, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusTentative,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
	})

	t.Run("rolls back with the caller transaction", func(t *testing.T) {
		owner4 := testdb.MakeTestOwner(t, db, clinicA, "次回予定確定飼主4")
		pet4 := makePetForReservationIntentTest(t, db, clinicA, owner4.ID, "次回予定確定ペット4")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet4.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		sentinel := errors.New("audit write failed")
		err = testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := repo.ConfirmTentative(txCtx, clinicA, tentative.ID, ConfirmTentativeUpdate{
				StartTime:         newStart,
				EndTime:           newEnd,
				ReservationTypeID: typeA.ID,
				Status:            model.ReservationStatusConfirmed,
			}); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("follows the pet's current owner after transfer", func(t *testing.T) {
		// Codex #16: tentative 作成後に pet が譲渡されると行の owner_id は旧飼い主の
		// まま残る。confirm は intent lock + 行ロック保持中に pet から現飼い主を解決し
		// owner_id を書き換える（確定済み予約が旧飼い主宛に紐付き続けるのを防ぐ）。
		oldOwner := testdb.MakeTestOwner(t, db, clinicA, "次回予定譲渡元飼主")
		newOwner := testdb.MakeTestOwner(t, db, clinicA, "次回予定譲渡先飼主")
		pet := makePetForReservationIntentTest(t, db, clinicA, oldOwner.ID, "次回予定譲渡ペット")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)
		require.NotNil(t, tentative.OwnerID)
		assert.Equal(t, oldOwner.ID, *tentative.OwnerID, "intent 作成時点では旧飼い主")

		// tentative 作成後の譲渡を再現する。
		require.NoError(t, db.Model(&model.Pet{}).
			Where("clinic_id = ? AND id = ?", clinicA, pet.ID).
			Update("owner_id", newOwner.ID).Error)

		confirmed, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.NoError(t, err)
		require.NotNil(t, confirmed)
		require.NotNil(t, confirmed.OwnerID)
		assert.Equal(t, newOwner.ID, *confirmed.OwnerID,
			"Codex #16: 確定後の owner_id はロック保持中に解決した現飼い主")
	})
}
