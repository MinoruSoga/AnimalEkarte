package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// stubReservationStaffWriteGuard is a programmable ReservationStaffWriteGuard for
// service-level intent tests. Real clinic/capability enforcement is covered at
// repository level (reservation_staff_repository tests).
type stubReservationStaffWriteGuard struct {
	findErr    error
	supports   bool
	supportErr error
}

func (g stubReservationStaffWriteGuard) FindByID(_ context.Context, _, id uint64) (*model.Staff, error) {
	if g.findErr != nil {
		return nil, g.findErr
	}
	return &model.Staff{ID: id, StaffType: model.StaffTypeDoctor}, nil
}

func (g stubReservationStaffWriteGuard) SupportsReservationType(_ context.Context, _, _, _ uint64) (bool, error) {
	return g.supports, g.supportErr
}

func newNextVisitIntentService(t *testing.T, db *gorm.DB, guard ReservationStaffWriteGuard) *reservationService {
	t.Helper()
	return newReservationService(
		NewReservationRepository(db),
		NewReservationTypeRepository(db),
		testNewTransactor(db),
		guard,
		nil,
		nil,
		nil,
		nil,
		nil, // EMR-308: nextVisitTagSync（既存テストは hook 未注入）
	)
}

func TestReservationService_SetNextVisitIntent_DateValidation(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	svc := newNextVisitIntentService(t, db, nil)
	ctx := context.Background()
	const clinicID, petID = uint64(1), uint64(1)

	t.Run("rejects a zero date", func(t *testing.T) {
		got, err := svc.SetNextVisitIntent(ctx, clinicID, petID, SetNextVisitIntentInput{})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
	})

	t.Run("rejects a past date", func(t *testing.T) {
		yesterday := time.Now().In(config.JST).AddDate(0, 0, -1)
		got, err := svc.SetNextVisitIntent(ctx, clinicID, petID, SetNextVisitIntentInput{IntendedDate: yesterday})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
	})

	t.Run("rejects a date beyond the forward window", func(t *testing.T) {
		tooFar := time.Now().In(config.JST).AddDate(nextVisitIntentMaxYearsAhead, 0, 1)
		got, err := svc.SetNextVisitIntent(ctx, clinicID, petID, SetNextVisitIntentInput{IntendedDate: tooFar})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
	})

	t.Run("accepts today and the window edge", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicID, "SVCEDGE")
		today := time.Now().In(config.JST)
		got, err := svc.SetNextVisitIntent(ctx, clinicID, pet.ID, SetNextVisitIntentInput{IntendedDate: today})
		require.NoError(t, err)
		require.NotNil(t, got)
		edge := today.AddDate(nextVisitIntentMaxYearsAhead, 0, 0)
		got, err = svc.SetNextVisitIntent(ctx, clinicID, pet.ID, SetNextVisitIntentInput{IntendedDate: edge})
		require.NoError(t, err)
		assert.Equal(t, model.ReservationStatusTentative, got.Status)
	})
}

func TestReservationService_SetNextVisitIntent_Flow(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	svc := newNextVisitIntentService(t, db, nil)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	_, petA, _ := makeNextVisitIntentFixture(t, db, clinicA, "SVCFLOW")
	intendedDate := time.Now().In(config.JST).AddDate(0, 0, 14)

	t.Run("creates and updates a tentative through the service", func(t *testing.T) {
		got, err := svc.SetNextVisitIntent(ctx, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, got.Status)
		require.NotNil(t, got.ReservationRoute)
		assert.Equal(t, "record_intent", *got.ReservationRoute)

		later := intendedDate.AddDate(0, 1, 0)
		updated, err := svc.SetNextVisitIntent(ctx, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: later})
		require.NoError(t, err)
		assert.Equal(t, got.ID, updated.ID)
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicA, petA.ID))
	})

	t.Run("rejects a pet owned by another clinic", func(t *testing.T) {
		_, petB, _ := makeNextVisitIntentFixture(t, db, clinicB, "SVCFLOWB")
		got, err := svc.SetNextVisitIntent(ctx, clinicA, petB.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
	})

	t.Run("rejects a deceased pet", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回予定死去飼主")
		pet := makePetForReservationIntentTest(t, db, clinicA, owner.ID, "次回予定死去ペット")
		deceasedAt := time.Now().UTC()
		require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pet.ID).Updates(map[string]any{
			"status":      model.PetStatusDeceased,
			"deceased_at": deceasedAt,
		}).Error)
		got, err := svc.SetNextVisitIntent(ctx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestReservationService_ClearNextVisitIntent(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	svc := newNextVisitIntentService(t, db, nil)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	intendedDate := time.Now().In(config.JST).AddDate(0, 0, 7)

	t.Run("clears the active tentative and tolerates a missing one", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "SVCCLR")
		_, err := svc.SetNextVisitIntent(ctx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		require.NoError(t, svc.ClearNextVisitIntent(ctx, clinicA, pet.ID))
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, pet.ID))
		// idempotent second call
		require.NoError(t, svc.ClearNextVisitIntent(ctx, clinicA, pet.ID))
	})

	t.Run("is clinic-scoped for pet lookup", func(t *testing.T) {
		_, petB, _ := makeNextVisitIntentFixture(t, db, clinicB, "SVCCLRB")
		_, err := svc.SetNextVisitIntent(ctx, clinicB, petB.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)
		require.Error(t, svc.ClearNextVisitIntent(ctx, clinicA, petB.ID))
		assert.True(t, apperrors.IsNotFound(svc.ClearNextVisitIntent(ctx, clinicA, petB.ID)))
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicB, petB.ID))
	})
}

func TestReservationService_ConfirmTentative(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	intendedDate := time.Now().In(config.JST).AddDate(0, 0, 30)

	newStart := time.Now().UTC().AddDate(0, 0, 30).Truncate(time.Minute)
	newEnd := newStart.Add(30 * time.Minute)

	t.Run("requires a valid time range", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		got, err := svc.ConfirmTentative(ctx, clinicA, 1, &ConfirmTentativeInput{
			StartTime: newEnd,
			EndTime:   newStart,
		})
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("confirms a tentative into a real booking with doctor", func(t *testing.T) {
		_, pet, typeA := makeNextVisitIntentFixture(t, db, clinicA, "SVCCFM")
		doctor := makeDoctor(t, db, clinicA, "確定テスト獣医")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicA, IsMain: true}).Error)

		svc := newNextVisitIntentService(t, db, stubReservationStaffWriteGuard{supports: true})
		tentative, err := svc.SetNextVisitIntent(ctx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		confirmed, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: &typeA.ID,
			DoctorID:          &doctor.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, confirmed)
		assert.Equal(t, model.ReservationStatusConfirmed, confirmed.Status)
		assert.True(t, confirmed.StartTime.Equal(newStart))
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, pet.ID))
	})

	t.Run("defaults to the tentative's own reservation type", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "SVCCFM2")
		doctor := makeDoctor(t, db, clinicA, "確定テスト獣医2")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicA, IsMain: true}).Error)

		svc := newNextVisitIntentService(t, db, stubReservationStaffWriteGuard{supports: true})
		tentative, err := svc.SetNextVisitIntent(ctx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		confirmed, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: newStart,
			EndTime:   newEnd,
			DoctorID:  &doctor.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, confirmed)
		assert.Equal(t, tentative.ReservationTypeID, confirmed.ReservationTypeID)
		assert.Equal(t, model.ReservationStatusConfirmed, confirmed.Status)
	})

	t.Run("supports pending as an explicit target status", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "SVCCFM3")
		svc := newNextVisitIntentService(t, db, nil)
		tentative, err := svc.SetNextVisitIntent(ctx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		pending := model.ReservationStatusPending
		// doctorless confirm requires on-duty capacity — add a shift entry.
		// 先行サブテストが newStart に確定済み予約を残すため、capacity 衝突しない
		// 同日 +2h の別枠を使う（shift_entries は日付粒度でカウントされる）。
		pendingStart := newStart.Add(2 * time.Hour)
		// CountOnDutyDoctors は DATE(? AT TIME ZONE 'Asia/Tokyo') で JST 暦日比較する。
		// shift_entries.date は UTC 暦日で格納されるため、実行時刻が JST 日跨ぎに近いと
		// pendingStart（UTC+2h）の JST 暦日と UTC 暦日がずれて flake する。既存 repo テストと
		// 同じく「pendingStart の JST 暦日を UTC 0 時で表した日付」を明示して格納する。
		jstPending := pendingStart.In(config.JST)
		shiftDate := time.Date(jstPending.Year(), jstPending.Month(), jstPending.Day(), 0, 0, 0, 0, time.UTC)
		doctor := makeDoctor(t, db, clinicA, "確定テスト獣医3")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicA, IsMain: true}).Error)
		makeShiftEntry(t, db, clinicA, doctor.ID, shiftDate)

		confirmed, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: pendingStart,
			EndTime:   pendingStart.Add(30 * time.Minute),
			Status:    &pending,
		})
		require.NoError(t, err)
		assert.Equal(t, model.ReservationStatusPending, confirmed.Status)
		assert.Nil(t, confirmed.DoctorID)
	})

	t.Run("rejects a non-tentative appointment", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		normal := makeReservationForReservationRepoTest(
			t, db, clinicA, makeReservationType(t, db, clinicA).ID,
			time.Now().UTC().Add(72*time.Hour),
			model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
		)
		got, err := svc.ConfirmTentative(ctx, clinicA, normal.ID, &ConfirmTentativeInput{
			StartTime: newStart,
			EndTime:   newEnd,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Nil(t, got)
	})

	t.Run("rejects a cross-clinic appointment", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		_, petB, _ := makeNextVisitIntentFixture(t, db, clinicB, "SVCCFMB")
		// create tentative under clinicB
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicB, petB.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)
		got, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: newStart,
			EndTime:   newEnd,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("rejects a reservation type owned by another clinic", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "SVCCFMT")
		_, _, typeB := makeNextVisitIntentFixture(t, db, clinicB, "SVCCFMTB")
		doctor := makeDoctor(t, db, clinicA, "確定テスト獣医4")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicA, IsMain: true}).Error)

		svc := newNextVisitIntentService(t, db, stubReservationStaffWriteGuard{supports: true})
		tentative, err := svc.SetNextVisitIntent(ctx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		got, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: &typeB.ID,
			DoctorID:          &doctor.ID,
		})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("rejects a slot conflict with an existing confirmed booking for the same doctor", func(t *testing.T) {
		_, pet, typeA := makeNextVisitIntentFixture(t, db, clinicA, "SVCCFM5")
		doctor := makeDoctor(t, db, clinicA, "確定テスト獣医5")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicA, IsMain: true}).Error)
		// conflicting confirmed booking at the same doctor/time
		makeReservationForReservationRepoTest(
			t, db, clinicA, typeA.ID, newStart,
			model.ReservationStatusConfirmed, model.ReservationSourceManual, nil, &doctor.ID,
		)

		svc := newNextVisitIntentService(t, db, stubReservationStaffWriteGuard{supports: true})
		tentative, err := svc.SetNextVisitIntent(ctx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		got, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: newStart,
			EndTime:   newEnd,
			DoctorID:  &doctor.ID,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})
}
