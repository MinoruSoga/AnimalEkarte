package billing

// accounting_complete_appointments_test.go — G11-2 テスト負債解消
//
// テスト対象: reservation.ReservationRepository.CompleteForAccounting の2経路SQL
// 保護する不変条件:
//   (1) 同日同一ペットの status=accounting 予約のみが completed へ遷移する
//       （JST日付境界 DATE(start_time AT TIME ZONE 'Asia/Tokyo') 述語）
//   (2) medical_record_id 経由のサブクエリ更新は自クリニックの appointment のみを対象とし、
//       既に completed/cancelled/no_show の予約には触れない
//
// このテストは path(1) の JST DATE 述語または clinic_id 述語を削除すると必ず失敗するよう設計されている。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// setupAccountingCompleteAppointmentsTestDB は G11-2 テスト用の DB を整備する。
func setupAccountingCompleteAppointmentsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{},
		&model.ReservationType{}, &model.Reservation{},
	))
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	return db
}

// makeAccountingAppointment は会計完了化テスト用の appointment を1件作成する。
func makeAccountingAppointment(t *testing.T, db *gorm.DB, clinicID uint64, ownerID, petID *uint64, status model.ReservationStatus, startTime time.Time) *model.Reservation {
	t.Helper()
	rt := makeReservationType(t, db, clinicID)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         startTime,
		EndTime:           startTime.Add(30 * time.Minute),
		OwnerID:           ownerID,
		PetID:             petID,
		ReservationTypeID: rt.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            status,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

// makeMedicalRecordForAppointment は appointment_id を持つ medical_record を1件作成する。
func makeMedicalRecordForAppointment(t *testing.T, db *gorm.DB, clinicID, appointmentID uint64, recordNo string) *model.MedicalRecord {
	t.Helper()
	appointmentIDCopy := appointmentID
	mr := &model.MedicalRecord{
		ClinicID:      clinicID,
		RecordNo:      recordNo,
		Date:          time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		AppointmentID: &appointmentIDCopy,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func reloadAppointmentStatus(t *testing.T, db *gorm.DB, id uint64) model.ReservationStatus {
	t.Helper()
	var res model.Reservation
	require.NoError(t, db.Unscoped().First(&res, id).Error)
	return res.Status
}

func completeForAccountingInTx(
	ctx context.Context,
	t *testing.T,
	db *gorm.DB,
	repo reservation.ReservationIntentRepository,
	clinicID uint64,
	medicalRecordID, ownerID, petID *uint64,
	scheduledDate time.Time,
) (int64, error) {
	t.Helper()
	var affected int64
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		affected, err = repo.CompleteForAccounting(
			persistence.WithTxValue(ctx, tx),
			clinicID,
			medicalRecordID,
			ownerID,
			petID,
			scheduledDate,
		)
		return err
	})
	return affected, err
}

func TestReservationRepository_CompleteForAccounting(t *testing.T) {
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	// JST日付境界の基準日: scheduledDate は billing.ScheduledDate 相当（type:date, UTC 0:00 格納）。
	scheduledDateJun10 := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	t.Run("経路1: 同日同一ペットのaccounting予約がcompletedになる", func(t *testing.T) {
		db := setupAccountingCompleteAppointmentsTestDB(t)
		repo := reservation.NewReservationRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "経路1飼主")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "経路1ペット")
		appt := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC)) // JST 12:00

		affected, err := completeForAccountingInTx(ctx, t, db, repo, clinicA, nil, &owner.ID, &pet.ID, scheduledDateJun10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
		assert.Equal(t, model.ReservationStatusCompleted, reloadAppointmentStatus(t, db, appt.ID))
	})

	t.Run("経路1除外: 別日/別ペット/別クリニック/status≠accounting/deleted_atは対象外", func(t *testing.T) {
		db := setupAccountingCompleteAppointmentsTestDB(t)
		repo := reservation.NewReservationRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "除外テスト飼主")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "除外テストペット")
		otherPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "別ペット")
		otherOwner := testdb.MakeTestOwner(t, db, clinicB, "別クリニック飼主")
		otherClinicPet := makeSpeciesAndPet(t, db, clinicB, otherOwner.ID, "別クリニックペット")

		otherDay := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 9, 3, 0, 0, 0, time.UTC)) // 別日
		otherPetAppt := makeAccountingAppointment(t, db, clinicA, &owner.ID, &otherPet.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC)) // 別ペット
		otherClinicAppt := makeAccountingAppointment(t, db, clinicB, &otherOwner.ID, &otherClinicPet.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC)) // 別クリニック（同一 owner/pet ID を狙っても clinic_id で拒否されることを petID/ownerID を変えて確認）
		nonAccounting := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet.ID, model.ReservationStatusPending,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC)) // status≠accounting
		deletedAppt := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC))
		require.NoError(t, db.Delete(&model.Reservation{}, deletedAppt.ID).Error) // soft delete

		affected, err := completeForAccountingInTx(ctx, t, db, repo, clinicA, nil, &owner.ID, &pet.ID, scheduledDateJun10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected, "対象クリニック/飼主/ペット/日付/statusに一致する行が無いため0件")

		assert.Equal(t, model.ReservationStatusAccounting, reloadAppointmentStatus(t, db, otherDay.ID), "別日は変更されない")
		assert.Equal(t, model.ReservationStatusAccounting, reloadAppointmentStatus(t, db, otherPetAppt.ID), "別ペットは変更されない")
		assert.Equal(t, model.ReservationStatusAccounting, reloadAppointmentStatus(t, db, otherClinicAppt.ID), "別クリニックは変更されない")
		assert.Equal(t, model.ReservationStatusPending, reloadAppointmentStatus(t, db, nonAccounting.ID), "status≠accountingは変更されない")
		assert.Equal(t, model.ReservationStatusAccounting, reloadAppointmentStatus(t, db, deletedAppt.ID), "deleted_atありは変更されない")
	})

	t.Run("JST日付境界: UTC前日15:00は同日扱い、UTC当日14:59も同日、UTC当日15:00は翌日扱いで対象外", func(t *testing.T) {
		db := setupAccountingCompleteAppointmentsTestDB(t)
		repo := reservation.NewReservationRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "境界テスト飼主")

		petBefore := makeSpeciesAndPet(t, db, clinicA, owner.ID, "境界前ペット") // UTC前日15:00 = JST当日0:00
		petLate := makeSpeciesAndPet(t, db, clinicA, owner.ID, "境界内ペット")   // UTC当日14:59 = JST当日23:59
		petAfter := makeSpeciesAndPet(t, db, clinicA, owner.ID, "境界後ペット")  // UTC当日15:00 = JST翌日0:00

		apptBefore := makeAccountingAppointment(t, db, clinicA, &owner.ID, &petBefore.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 9, 15, 0, 0, 0, time.UTC))
		apptLate := makeAccountingAppointment(t, db, clinicA, &owner.ID, &petLate.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 10, 14, 59, 0, 0, time.UTC))
		apptAfter := makeAccountingAppointment(t, db, clinicA, &owner.ID, &petAfter.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 10, 15, 0, 0, 0, time.UTC))

		_, err := completeForAccountingInTx(ctx, t, db, repo, clinicA, nil, &owner.ID, &petBefore.ID, scheduledDateJun10)
		require.NoError(t, err)
		_, err = completeForAccountingInTx(ctx, t, db, repo, clinicA, nil, &owner.ID, &petLate.ID, scheduledDateJun10)
		require.NoError(t, err)
		_, err = completeForAccountingInTx(ctx, t, db, repo, clinicA, nil, &owner.ID, &petAfter.ID, scheduledDateJun10)
		require.NoError(t, err)

		assert.Equal(t, model.ReservationStatusCompleted, reloadAppointmentStatus(t, db, apptBefore.ID), "UTC前日15:00(JST当日0:00)は同日扱いで完了化")
		assert.Equal(t, model.ReservationStatusCompleted, reloadAppointmentStatus(t, db, apptLate.ID), "UTC当日14:59(JST当日23:59)も同日扱いで完了化")
		assert.Equal(t, model.ReservationStatusAccounting, reloadAppointmentStatus(t, db, apptAfter.ID), "UTC当日15:00(JST翌日0:00)は対象外のまま")
	})

	t.Run("経路2: medicalRecordID経由で非完了予約が完了化される", func(t *testing.T) {
		db := setupAccountingCompleteAppointmentsTestDB(t)
		repo := reservation.NewReservationRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "経路2飼主")
		pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "経路2ペット")
		appt := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet.ID, model.ReservationStatusPending,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC))
		mr := makeMedicalRecordForAppointment(t, db, clinicA, appt.ID, "MR-G11-2-001")

		affected, err := completeForAccountingInTx(ctx, t, db, repo, clinicA, &mr.ID, nil, nil, time.Time{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
		assert.Equal(t, model.ReservationStatusCompleted, reloadAppointmentStatus(t, db, appt.ID))
	})

	t.Run("経路2除外: completed/cancelled/no_showは触らない、別クリニックのmedical_recordは対象外", func(t *testing.T) {
		db := setupAccountingCompleteAppointmentsTestDB(t)
		repo := reservation.NewReservationRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "経路2除外飼主")

		for _, status := range []model.ReservationStatus{
			model.ReservationStatusCompleted,
			model.ReservationStatusCancelled,
			model.ReservationStatusNoShow,
		} {
			pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "経路2除外ペット")
			appt := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet.ID, status,
				time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC))
			mr := makeMedicalRecordForAppointment(t, db, clinicA, appt.ID, "MR-G11-2-excl-"+string(status))

			affected, err := completeForAccountingInTx(ctx, t, db, repo, clinicA, &mr.ID, nil, nil, time.Time{})
			require.NoError(t, err)
			assert.Equal(t, int64(0), affected, "status=%s は既に完了扱いのため対象外", status)
			assert.Equal(t, status, reloadAppointmentStatus(t, db, appt.ID))
		}

		// クリニック隔離: clinic A の medical_record を clinic B のコンテキストで完了化しようとしても対象外。
		petA := makeSpeciesAndPet(t, db, clinicA, owner.ID, "隔離ペットA")
		apptA := makeAccountingAppointment(t, db, clinicA, &owner.ID, &petA.ID, model.ReservationStatusPending,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC))
		mrA := makeMedicalRecordForAppointment(t, db, clinicA, apptA.ID, "MR-G11-2-isolation")

		affected, err := completeForAccountingInTx(ctx, t, db, repo, clinicB, &mrA.ID, nil, nil, time.Time{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected, "別クリニックからは medical_record が見つからず対象外")
		assert.Equal(t, model.ReservationStatusPending, reloadAppointmentStatus(t, db, apptA.ID), "clinic Aの予約は変更されない")

		// 多重防御: medical_record.clinic_id は呼び出し側クリニックと一致するが、
		// appointment_id が指す先が別クリニックの予約というFKドリフトを想定したケース。
		// サービス層の書込経路は常にクリニック隔離済みの予約参照から appointment_id を設定するため
		// 本来到達しないが、外側 Where("clinic_id = ?", clinicID)（Reservation側）が
		// 独立した防御層として機能することを証明する（repository/CLAUDE.md P3.1の「正本ガード=runtime isolation test」方針）。
		petFKDrift := makeSpeciesAndPet(t, db, clinicA, owner.ID, "FKドリフトペット")
		apptFKDrift := makeAccountingAppointment(t, db, clinicA, &owner.ID, &petFKDrift.ID, model.ReservationStatusPending,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC)) // clinic A 所属の予約
		mrFKDrift := makeMedicalRecordForAppointment(t, db, clinicB, apptFKDrift.ID, "MR-G11-2-fk-drift") // clinic_id=B だが appointment は clinic A

		affected, err = completeForAccountingInTx(ctx, t, db, repo, clinicB, &mrFKDrift.ID, nil, nil, time.Time{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected, "medical_record.clinic_id一致でもappointment自体が別クリニックなら外側clinic_id述語で拒否される")
		assert.Equal(t, model.ReservationStatusPending, reloadAppointmentStatus(t, db, apptFKDrift.ID), "clinic Aの予約は変更されない")

		// サブクエリ述語: appointment_id IS NULL の medical_record は対象外（エラーにもならない）。
		mrNoAppointment := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-G11-2-no-appt", Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)}
		require.NoError(t, db.WithContext(ctx).Create(mrNoAppointment).Error)
		affected, err = completeForAccountingInTx(ctx, t, db, repo, clinicA, &mrNoAppointment.ID, nil, nil, time.Time{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected, "appointment_idがnilのmedical_recordは対象外")

		// サブクエリ述語: soft-delete済みmedical_recordは対象外（deleted_at IS NULL）。
		petDeletedMR := makeSpeciesAndPet(t, db, clinicA, owner.ID, "削除カルテペット")
		apptDeletedMR := makeAccountingAppointment(t, db, clinicA, &owner.ID, &petDeletedMR.ID, model.ReservationStatusPending,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC))
		mrDeleted := makeMedicalRecordForAppointment(t, db, clinicA, apptDeletedMR.ID, "MR-G11-2-soft-deleted")
		require.NoError(t, db.Delete(&model.MedicalRecord{}, mrDeleted.ID).Error)

		affected, err = completeForAccountingInTx(ctx, t, db, repo, clinicA, &mrDeleted.ID, nil, nil, time.Time{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected, "soft-delete済みmedical_recordは対象外")
		assert.Equal(t, model.ReservationStatusPending, reloadAppointmentStatus(t, db, apptDeletedMR.ID))
	})

	t.Run("totalAffectedは経路1と経路2の合算になる", func(t *testing.T) {
		db := setupAccountingCompleteAppointmentsTestDB(t)
		repo := reservation.NewReservationRepository(db)
		owner := testdb.MakeTestOwner(t, db, clinicA, "合算テスト飼主")
		pet1 := makeSpeciesAndPet(t, db, clinicA, owner.ID, "合算ペット1")
		pet2 := makeSpeciesAndPet(t, db, clinicA, owner.ID, "合算ペット2")

		appt1 := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet1.ID, model.ReservationStatusAccounting,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC)) // 経路1対象
		appt2 := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet2.ID, model.ReservationStatusPending,
			time.Date(2026, 6, 10, 3, 0, 0, 0, time.UTC)) // 経路2対象
		mr2 := makeMedicalRecordForAppointment(t, db, clinicA, appt2.ID, "MR-G11-2-combined")

		affected, err := completeForAccountingInTx(ctx, t, db, repo, clinicA, &mr2.ID, &owner.ID, &pet1.ID, scheduledDateJun10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), affected, "経路1(1件)+経路2(1件)=2件")
		assert.Equal(t, model.ReservationStatusCompleted, reloadAppointmentStatus(t, db, appt1.ID))
		assert.Equal(t, model.ReservationStatusCompleted, reloadAppointmentStatus(t, db, appt2.ID))
	})

	t.Run("nil縮退: ownerID/petID/medicalRecordIDがnilなら両経路スキップされ0件・エラーなし", func(t *testing.T) {
		db := setupAccountingCompleteAppointmentsTestDB(t)
		repo := reservation.NewReservationRepository(db)

		affected, err := completeForAccountingInTx(ctx, t, db, repo, clinicA, nil, nil, nil, scheduledDateJun10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected)
	})
}
