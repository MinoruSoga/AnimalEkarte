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
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
		openDayHolidayFinder(), // ConfirmTentative は create 経路と同じく休診日を無条件検証する
		nil,                    // EMR-308: nextVisitTagSync（既存テストは hook 未注入）
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
		// 2回目は既存 intent の upsert — handler が edit 権限を解決して立てる
		// CanUpdateExisting を service 層テストで再現する。
		got, err = svc.SetNextVisitIntent(ctx, clinicID, pet.ID, SetNextVisitIntentInput{IntendedDate: edge, CanUpdateExisting: true})
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
		updated, err := svc.SetNextVisitIntent(ctx, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: later, CanUpdateExisting: true})
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

	t.Run("rejects confirming a deceased pet's tentative", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "SVCCFMDEC")
		tentative, err := svc.SetNextVisitIntent(ctx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		// 登録後に死亡したケース: status ラベルと deceased_at の両方を立てる
		// （BE の死亡契約は OR 判定）。
		deceasedAt := time.Now().UTC()
		require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pet.ID).Updates(map[string]any{
			"status":      model.PetStatusDeceased,
			"deceased_at": deceasedAt,
		}).Error)

		got, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: newStart,
			EndTime:   newEnd,
		})
		require.Error(t, err)
		assert.Nil(t, got)
		// ガード自体を固定する — 容量/競合等の後段エラーで誤って合格しないよう
		// 死亡拒否メッセージを明示する（ガードを外すと capacity/conflict で失敗し
		// 「Error になる」だけでは検出できない）。
		assert.Contains(t, err.Error(), reservationDeceasedPetMessage)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID),
			"死亡ペットの tentative は実予約化されず残る")
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

// ---------------------------------------------------------------------------
// Codex PR #606 Session A: ConfirmTentative service 層の回帰テスト（mock 駆動）
//   #4  intent advisory lock → 行ロックの順序強制
//   #15 無効化された予約区分の拒否
//   #6  非 general placeholder 区分は明示指定なしに確定させない
//   #16 譲渡後の現飼い主追従（新旧 owner のタグ再同期）
//   #3  確定時の死亡ペット拒否（行ロック保持中の検証）
// ---------------------------------------------------------------------------

// buildConfirmTentativeSvc は ConfirmTentative の service テスト用の組み立て口。
// tentative 行・区分・ロック呼出を差し替えられるようにする。
func buildConfirmTentativeSvc(
	repo *mockReservationRepository,
	typeFinder reservationTypeFinder,
	syncer NextVisitTagSyncer,
) ReservationService {
	return NewReservationServiceWithClinicHolidays(
		repo,
		typeFinder,
		&mockTransactor{},
		nil,
		&mockUnavailableTimeRepository{},
		nil,
		nil,
		&mockClinicHolidayFinder{},
		syncer,
	)
}

func confirmTentativeFixture(petID, ownerID, typeID uint64) *model.Reservation {
	sentinelStart := time.Date(2030, 6, 15, 0, 0, 0, 0, config.JST)
	return &model.Reservation{
		ID:                42,
		ClinicID:          1,
		PetID:             &petID,
		OwnerID:           &ownerID,
		Status:            model.ReservationStatusTentative,
		StartTime:         sentinelStart,
		EndTime:           sentinelStart.Add(nextVisitIntentDurationMinutes * time.Minute),
		ReservationTypeID: typeID,
		Source:            model.ReservationSourceManual,
	}
}

func alivePetFinder(petID, ownerID uint64) func(ctx context.Context, clinicID, pid uint64) (*model.Pet, error) {
	return func(_ context.Context, _, _ uint64) (*model.Pet, error) {
		return &model.Pet{ID: petID, OwnerID: ownerID}, nil
	}
}

func TestReservationService_ConfirmTentative_AcquiresIntentLockBeforeRowLock(t *testing.T) {
	petID, ownerID, typeID := uint64(22), uint64(10), uint64(7)
	tentative := confirmTentativeFixture(petID, ownerID, typeID)

	var order []string
	repo := &mockReservationRepository{
		acquireBookingLockFn: func(_ context.Context, _ uint64) error {
			order = append(order, "booking_lock")
			return nil
		},
		findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) {
			order = append(order, "read")
			return &petID, nil
		},
		acquireNextVisitIntentLockFn: func(_ context.Context, _, _ uint64) error {
			order = append(order, "intent_lock")
			return nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			order = append(order, "row_lock")
			return tentative, nil
		},
		findPetByIDInClinicFn: alivePetFinder(petID, ownerID),
		confirmTentativeFn: func(_ context.Context, _, _ uint64, _ ConfirmTentativeUpdate) (*model.Reservation, error) {
			order = append(order, "confirm")
			return tentative, nil
		},
	}
	typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: true}, nil
	}}
	svc := buildConfirmTentativeSvc(repo, typeFinder, nil)

	got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
		StartTime: time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST),
		EndTime:   time.Date(2030, 6, 15, 10, 30, 0, 0, config.JST),
	})
	require.NoError(t, err)
	require.NotNil(t, got)

	intentIdx, rowIdx := -1, -1
	for i, step := range order {
		switch step {
		case "intent_lock":
			if intentIdx == -1 {
				intentIdx = i
			}
		case "row_lock":
			if rowIdx == -1 {
				rowIdx = i
			}
		}
	}
	require.NotEqual(t, -1, intentIdx, "(clinic,pet) intent lock が取得されていない")
	require.NotEqual(t, -1, rowIdx, "行ロックが取得されていない")
	assert.Less(t, intentIdx, rowIdx,
		"Codex #4: lock 順序は intent advisory lock → 行ロック（逆順だと set/clear との AB-BA デッドロックになる）: %v", order)
}

func TestReservationService_ConfirmTentative_RejectsInactiveReservationType(t *testing.T) {
	petID, ownerID, typeID := uint64(22), uint64(10), uint64(7)
	tentative := confirmTentativeFixture(petID, ownerID, typeID)

	repo := &mockReservationRepository{
		findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) { return &petID, nil },
		lockAndFindByIDFn:        func(_ context.Context, _, _ uint64) (*model.Reservation, error) { return tentative, nil },
		findPetByIDInClinicFn:    alivePetFinder(petID, ownerID),
		confirmTentativeFn: func(_ context.Context, _, _ uint64, _ ConfirmTentativeUpdate) (*model.Reservation, error) {
			t.Fatal("inactive 区分の confirm は repo 書込へ到達してはならない")
			return nil, nil
		},
	}
	typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: false}, nil
	}}
	svc := buildConfirmTentativeSvc(repo, typeFinder, nil)

	got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
		StartTime: time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST),
		EndTime:   time.Date(2030, 6, 15, 10, 30, 0, 0, config.JST),
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "Codex #15: 無効化区分の確定は InvalidInput, 実際: %v", err)
	assert.Nil(t, got)
}

func TestReservationService_ConfirmTentative_PlaceholderTypeRequiresExplicitOverride(t *testing.T) {
	petID, ownerID := uint64(22), uint64(10)
	start := time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)

	newSvc := func(tentativeType *model.ReservationType, onConfirm func(ConfirmTentativeUpdate)) ReservationService {
		tentative := confirmTentativeFixture(petID, ownerID, tentativeType.ID)
		repo := &mockReservationRepository{
			findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) { return &petID, nil },
			lockAndFindByIDFn:        func(_ context.Context, _, _ uint64) (*model.Reservation, error) { return tentative, nil },
			findPetByIDInClinicFn:    alivePetFinder(petID, ownerID),
			confirmTentativeFn: func(_ context.Context, _, _ uint64, input ConfirmTentativeUpdate) (*model.Reservation, error) {
				if onConfirm != nil {
					onConfirm(input)
				}
				return tentative, nil
			},
		}
		typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
			if id == tentativeType.ID {
				return tentativeType, nil
			}
			return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: true}, nil
		}}
		return buildConfirmTentativeSvc(repo, typeFinder, nil)
	}

	t.Run("rejects confirm when the stored placeholder is not general and no type is given", func(t *testing.T) {
		trimmingPlaceholder := &model.ReservationType{
			ID: 99, ClinicID: 1, Name: "migration placeholder",
			Category: model.ReservationTypeCategoryTrimming, IsActive: true,
		}
		svc := newSvc(trimmingPlaceholder, func(ConfirmTentativeUpdate) {
			t.Fatal("placeholder 維持の confirm は repo 書込へ到達してはならない")
		})

		got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
			StartTime: start,
			EndTime:   end,
			// ReservationTypeID 未指定 = stored placeholder 維持
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "Codex #6: 非 general placeholder の黙認昇格は InvalidInput, 実際: %v", err)
		assert.Nil(t, got)
	})

	t.Run("confirms when an explicit general type overrides the placeholder", func(t *testing.T) {
		trimmingPlaceholder := &model.ReservationType{
			ID: 99, ClinicID: 1, Name: "migration placeholder",
			Category: model.ReservationTypeCategoryTrimming, IsActive: true,
		}
		var captured ConfirmTentativeUpdate
		svc := newSvc(trimmingPlaceholder, func(input ConfirmTentativeUpdate) {
			captured = input
		})

		explicitTypeID := uint64(7)
		got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
			StartTime:         start,
			EndTime:           end,
			ReservationTypeID: &explicitTypeID,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, explicitTypeID, captured.ReservationTypeID,
			"明示された区分が repo 更新へ渡る")
	})
}

func TestReservationService_ConfirmTentative_RejectsDeceasedPetPostLock(t *testing.T) {
	petID, ownerID, typeID := uint64(22), uint64(10), uint64(7)
	tentative := confirmTentativeFixture(petID, ownerID, typeID)

	repo := &mockReservationRepository{
		findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) { return &petID, nil },
		lockAndFindByIDFn:        func(_ context.Context, _, _ uint64) (*model.Reservation, error) { return tentative, nil },
		// Codex #3: 死亡は行ロック保持中（intent lock 後）の検証で拒否される。
		readPetLifeStateFn: func(_ context.Context, _, _ uint64) (sharedkernel.PetLifeState, error) {
			return sharedkernel.PetLifeStateDeceased, nil
		},
		confirmTentativeFn: func(_ context.Context, _, _ uint64, _ ConfirmTentativeUpdate) (*model.Reservation, error) {
			t.Fatal("死亡ペットの confirm は repo 書込へ到達してはならない")
			return nil, nil
		},
	}
	typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: true}, nil
	}}
	svc := buildConfirmTentativeSvc(repo, typeFinder, nil)

	got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
		StartTime: time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST),
		EndTime:   time.Date(2030, 6, 15, 10, 30, 0, 0, config.JST),
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "死亡ペットの確定は InvalidInput, 実際: %v", err)
	assert.Nil(t, got)
}

func TestReservationService_ConfirmTentative_SyncsPreviousAndCurrentOwnerAfterTransfer(t *testing.T) {
	petID, prevOwnerID, newOwnerID, typeID := uint64(22), uint64(10), uint64(20), uint64(7)
	tentative := confirmTentativeFixture(petID, prevOwnerID, typeID)

	repo := &mockReservationRepository{
		findPetIDForIntentLockFn: func(_ context.Context, _, _ uint64) (*uint64, error) { return &petID, nil },
		lockAndFindByIDFn:        func(_ context.Context, _, _ uint64) (*model.Reservation, error) { return tentative, nil },
		findPetByIDInClinicFn:    alivePetFinder(petID, newOwnerID),
		confirmTentativeFn: func(_ context.Context, _, _ uint64, _ ConfirmTentativeUpdate) (*model.Reservation, error) {
			// repo 側（ロック保持中）が owner_id を現飼い主へ追従した結果を再現する。
			confirmed := *tentative
			confirmed.Status = model.ReservationStatusConfirmed
			confirmed.OwnerID = &newOwnerID
			return &confirmed, nil
		},
	}
	typeFinder := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: 1, Category: model.ReservationTypeCategoryGeneral, IsActive: true}, nil
	}}
	syncer := &recordingNextVisitTagSyncer{}
	svc := buildConfirmTentativeSvc(repo, typeFinder, syncer)

	got, err := svc.ConfirmTentative(context.Background(), 1, 42, &ConfirmTentativeInput{
		StartTime: time.Date(2030, 6, 15, 10, 0, 0, 0, config.JST),
		EndTime:   time.Date(2030, 6, 15, 10, 30, 0, 0, config.JST),
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.OwnerID)
	assert.Equal(t, newOwnerID, *got.OwnerID)

	synced := map[uint64]bool{}
	for _, ownerID := range syncer.calls {
		synced[ownerID] = true
	}
	assert.True(t, synced[newOwnerID], "Codex #16: 現飼い主の next_visit タグが再同期される")
	assert.True(t, synced[prevOwnerID], "Codex #16: 譲渡元（旧飼い主）の stale タグも除去される")
}
