package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---------------------------------------------------------------------------
// EMR-308 lifecycle absorb: 実予約作成 / 当日受付での tentative 同一行昇格
// ---------------------------------------------------------------------------

func absorbTentativeOnBookingInTx(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationTentativeLifecycleRepository,
	booking *model.Reservation,
) (*model.Reservation, error) {
	var got *model.Reservation
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		got, err = repo.AbsorbTentativeOnBooking(txCtx, booking)
		return err
	})
	return got, err
}

func promoteTentativeOnCheckinInTx(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationTentativeLifecycleRepository,
	booking *model.Reservation,
) (*model.Reservation, error) {
	var got *model.Reservation
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		got, err = repo.PromoteTentativeOnCheckin(txCtx, booking)
		return err
	})
	return got, err
}

// countNonDeletedForPet は論理削除されていない予約行の総数（status 不問）を返す。
func countNonDeletedForPet(t *testing.T, db *gorm.DB, clinicID, petID uint64) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("clinic_id = ? AND pet_id = ? AND deleted_at IS NULL", clinicID, petID).
		Count(&count).Error)
	return count
}

func makeAbsorbBooking(ownerID, petID, reservationTypeID uint64, start, end time.Time) *model.Reservation {
	return &model.Reservation{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           end,
		OwnerID:           &ownerID,
		PetID:             &petID,
		ReservationTypeID: reservationTypeID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusConfirmed,
		Source:            model.ReservationSourceManual,
	}
}

func TestReservationRepository_AbsorbTentativeOnBooking(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA, petA, typeA := makeNextVisitIntentFixture(t, db, clinicA, "AB")
	_, petB, _ := makeNextVisitIntentFixture(t, db, clinicB, "AB-B")

	slotStart := time.Date(2030, 3, 11, 10, 0, 0, 0, config.JST)
	slotEnd := slotStart.Add(30 * time.Minute)

	t.Run("requires an ambient transaction", func(t *testing.T) {
		booking := makeAbsorbBooking(ownerA.ID, petA.ID, typeA.ID, slotStart, slotEnd)
		got, err := repo.AbsorbTentativeOnBooking(ctx, booking)
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("inserts a normal booking when no tentative exists", func(t *testing.T) {
		booking := makeAbsorbBooking(ownerA.ID, petA.ID, typeA.ID, slotStart, slotEnd)
		got, err := absorbTentativeOnBookingInTx(ctx, db, repo, booking)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.NotZero(t, got.ID)
		assert.Equal(t, model.ReservationStatusConfirmed, got.Status)
		assert.True(t, got.StartTime.Equal(slotStart))
		assert.Nil(t, got.IntendedDate)
		assert.Equal(t, int64(1), countNonDeletedForPet(t, db, clinicA, petA.ID))
	})

	t.Run("promotes the tentative row in place and preserves provenance", func(t *testing.T) {
		pet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "吸収ペット1")
		record := makeMedicalRecordForNextVisitIntentTest(t, db, clinicA, ownerA.ID, pet.ID, "ABSRC")
		intended := time.Date(2030, 4, 5, 0, 0, 0, 0, config.JST)
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{
			IntendedDate:          intended,
			SourceMedicalRecordID: &record.ID,
		})
		require.NoError(t, err)
		tentativeID := tentative.ID
		require.Equal(t, int64(1), countNonDeletedForPet(t, db, clinicA, pet.ID))

		otherType := &model.ReservationType{ClinicID: clinicA, Name: "吸収用別区分", Category: model.ReservationTypeCategoryGeneral, IsActive: true}
		require.NoError(t, db.Create(otherType).Error)
		doctor := makeDoctor(t, db, clinicA, "吸収テスト医")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicA, IsMain: true}).Error)

		booking := makeAbsorbBooking(ownerA.ID, pet.ID, otherType.ID, slotStart, slotEnd)
		booking.DoctorID = &doctor.ID
		booking.IsDesignated = true

		got, err := absorbTentativeOnBookingInTx(ctx, db, repo, booking)
		require.NoError(t, err)
		require.NotNil(t, got)

		// 同一行昇格: id 維持・二重 INSERT なし
		assert.Equal(t, tentativeID, got.ID)
		assert.Equal(t, int64(1), countNonDeletedForPet(t, db, clinicA, pet.ID))
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, pet.ID))

		// 実予約値で上書き
		assert.Equal(t, model.ReservationStatusConfirmed, got.Status)
		assert.Equal(t, otherType.ID, got.ReservationTypeID)
		require.NotNil(t, got.DoctorID)
		assert.Equal(t, doctor.ID, *got.DoctorID)
		assert.True(t, got.StartTime.Equal(slotStart))
		assert.True(t, got.EndTime.Equal(slotEnd))
		assert.True(t, got.IsDesignated)

		// provenance 維持
		require.NotNil(t, got.IntendedDate)
		assert.Equal(t, "2030-04-05", got.IntendedDate.In(config.JST).Format(time.DateOnly))
		require.NotNil(t, got.SourceMedicalRecordID)
		assert.Equal(t, record.ID, *got.SourceMedicalRecordID)
		require.NotNil(t, got.ReservationRoute)
		assert.Equal(t, nextVisitIntentRoute, *got.ReservationRoute)
	})

	t.Run("normalizes empty status to pending", func(t *testing.T) {
		pet2 := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "吸収ペット2")
		_, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet2.ID, SetNextVisitIntentInput{
			IntendedDate: time.Date(2030, 5, 1, 0, 0, 0, 0, config.JST),
		})
		require.NoError(t, err)

		booking := makeAbsorbBooking(ownerA.ID, pet2.ID, typeA.ID, slotStart, slotEnd)
		booking.Status = ""
		got, err := absorbTentativeOnBookingInTx(ctx, db, repo, booking)
		require.NoError(t, err)
		assert.Equal(t, model.ReservationStatusPending, got.Status)
	})

	t.Run("falls back to plain insert when pet is nil", func(t *testing.T) {
		booking := &model.Reservation{
			ClinicID:          clinicA,
			StartTime:         slotStart,
			EndTime:           slotEnd,
			OwnerID:           &ownerA.ID,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusPending,
			Source:            model.ReservationSourceManual,
		}
		got, err := absorbTentativeOnBookingInTx(ctx, db, repo, booking)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.NotZero(t, got.ID)
	})

	t.Run("does not touch tentative rows of other pets or clinics", func(t *testing.T) {
		pet3 := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "吸収ペット3")
		otherTentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet3.ID, SetNextVisitIntentInput{
			IntendedDate: time.Date(2030, 6, 1, 0, 0, 0, 0, config.JST),
		})
		require.NoError(t, err)
		tentativeB, err := setNextVisitIntentInTx(ctx, db, repo, clinicB, petB.ID, SetNextVisitIntentInput{
			IntendedDate: time.Date(2030, 6, 2, 0, 0, 0, 0, config.JST),
		})
		require.NoError(t, err)

		// 別ペットの予約は通常 INSERT になり、他ペット・他医院の tentative は不変。
		pet4 := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "吸収ペット4")
		booking := makeAbsorbBooking(ownerA.ID, pet4.ID, typeA.ID, slotStart, slotEnd)
		got, err := absorbTentativeOnBookingInTx(ctx, db, repo, booking)
		require.NoError(t, err)
		assert.NotEqual(t, otherTentative.ID, got.ID)
		assert.NotEqual(t, tentativeB.ID, got.ID)
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicA, pet3.ID))
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicB, petB.ID))
	})

	t.Run("does not revive a cancelled (soft-deleted) tentative-origin row", func(t *testing.T) {
		pet5 := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "吸収ペット5")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet5.ID, SetNextVisitIntentInput{
			IntendedDate: time.Date(2030, 7, 1, 0, 0, 0, 0, config.JST),
		})
		require.NoError(t, err)
		// cancel semantics: status=cancelled + soft delete（Update isCancel 経路と同じ形）
		require.NoError(t, db.Model(&model.Reservation{}).
			Where("id = ?", tentative.ID).
			Update("status", model.ReservationStatusCancelled).Error)
		require.NoError(t, db.Delete(&model.Reservation{}, tentative.ID).Error)

		booking := makeAbsorbBooking(ownerA.ID, pet5.ID, typeA.ID, slotStart, slotEnd)
		got, err := absorbTentativeOnBookingInTx(ctx, db, repo, booking)
		require.NoError(t, err)
		// 新規行が作られ、cancelled 行は復活しない
		assert.NotEqual(t, tentative.ID, got.ID)
		assert.Equal(t, int64(1), countNonDeletedForPet(t, db, clinicA, pet5.ID))
		var cancelled model.Reservation
		require.NoError(t, db.Unscoped().First(&cancelled, tentative.ID).Error)
		assert.Equal(t, model.ReservationStatusCancelled, cancelled.Status)
		assert.NotNil(t, cancelled.DeletedAt)
	})

	t.Run("rolls back the promotion when the ambient transaction fails", func(t *testing.T) {
		pet6 := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "吸収ペット6")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet6.ID, SetNextVisitIntentInput{
			IntendedDate: time.Date(2030, 8, 1, 0, 0, 0, 0, config.JST),
		})
		require.NoError(t, err)

		sentinel := errors.New("abort after absorb")
		txErr := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			booking := makeAbsorbBooking(ownerA.ID, pet6.ID, typeA.ID, slotStart, slotEnd)
			if _, err := repo.AbsorbTentativeOnBooking(txCtx, booking); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, txErr, sentinel)

		reloaded, err := repo.FindByID(ctx, clinicA, tentative.ID)
		require.NoError(t, err)
		require.NotNil(t, reloaded)
		assert.Equal(t, model.ReservationStatusTentative, reloaded.Status)
		assert.Equal(t, int64(1), countNonDeletedForPet(t, db, clinicA, pet6.ID))
	})
}

func TestReservationRepository_PromoteTentativeOnCheckin(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA, petA, typeA := makeNextVisitIntentFixture(t, db, clinicA, "CI")

	t.Run("requires an ambient transaction", func(t *testing.T) {
		got, err := repo.PromoteTentativeOnCheckin(ctx, &model.Reservation{ClinicID: clinicA, PetID: &petA.ID})
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("promotes tentative to checked_in keeping placeholder type and provenance", func(t *testing.T) {
		pet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "受付ペット1")
		record := makeMedicalRecordForNextVisitIntentTest(t, db, clinicA, ownerA.ID, pet.ID, "CISRC")
		intended := time.Date(2030, 9, 9, 0, 0, 0, 0, config.JST)
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{
			IntendedDate:          intended,
			SourceMedicalRecordID: &record.ID,
		})
		require.NoError(t, err)

		checkinStart := time.Date(2030, 9, 9, 11, 30, 0, 0, config.JST)
		checkinEnd := checkinStart.Add(15 * time.Minute)
		// 受付作成では別区分が渡ってきても tentative 行の general placeholder を維持する。
		otherType := &model.ReservationType{ClinicID: clinicA, Name: "受付別区分", Category: model.ReservationTypeCategoryGeneral, IsActive: true}
		require.NoError(t, db.Create(otherType).Error)

		booking := &model.Reservation{
			ClinicID:          clinicA,
			StartTime:         checkinStart,
			EndTime:           checkinEnd,
			OwnerID:           &ownerA.ID,
			PetID:             &pet.ID,
			ReservationTypeID: otherType.ID,
			Status:            model.ReservationStatusCheckedIn,
			Source:            model.ReservationSourceManual,
			VisitType:         model.VisitTypeRevisit,
		}
		got, err := promoteTentativeOnCheckinInTx(ctx, db, repo, booking)
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, tentative.ID, got.ID)
		assert.Equal(t, model.ReservationStatusCheckedIn, got.Status)
		assert.Equal(t, tentative.ReservationTypeID, got.ReservationTypeID, "当日受付昇格は placeholder 区分を維持する")
		assert.Equal(t, typeA.ID, got.ReservationTypeID)
		require.NotNil(t, got.CheckedInAt)
		assert.True(t, got.StartTime.Equal(checkinStart))
		assert.True(t, got.EndTime.Equal(checkinEnd))
		require.NotNil(t, got.IntendedDate)
		assert.Equal(t, "2030-09-09", got.IntendedDate.In(config.JST).Format(time.DateOnly))
		require.NotNil(t, got.SourceMedicalRecordID)
		assert.Equal(t, record.ID, *got.SourceMedicalRecordID)
		assert.Equal(t, int64(1), countNonDeletedForPet(t, db, clinicA, pet.ID))
	})

	t.Run("inserts a checked_in row when no tentative exists", func(t *testing.T) {
		pet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "受付ペット2")
		booking := &model.Reservation{
			ClinicID:          clinicA,
			StartTime:         time.Date(2030, 9, 10, 9, 0, 0, 0, config.JST),
			EndTime:           time.Date(2030, 9, 10, 9, 15, 0, 0, config.JST),
			OwnerID:           &ownerA.ID,
			PetID:             &pet.ID,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusCheckedIn,
			Source:            model.ReservationSourceManual,
			VisitType:         model.VisitTypeRevisit,
		}
		got, err := promoteTentativeOnCheckinInTx(ctx, db, repo, booking)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, model.ReservationStatusCheckedIn, got.Status)
		require.NotNil(t, got.CheckedInAt)
	})
}

// ---------------------------------------------------------------------------
// サービス配線: 各作成経路が Create ではなく absorb/promote 入口を通ること
// ---------------------------------------------------------------------------

func TestReservationService_Create_UsesTentativeAbsorbEntrypoint(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	ownerID, petID := uint64(11), uint64(22)
	absorbCalled, createCalled := false, false

	repo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) { return 1, nil },
		countConflictsFn:     func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) { return 0, nil },
		findPetOwnerInClinicFn: func(_ context.Context, _ uint64, _ uint64) (uint64, error) {
			return ownerID, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
		absorbTentativeOnBookingFn: func(_ context.Context, booking *model.Reservation) (*model.Reservation, error) {
			absorbCalled = true
			assert.Equal(t, petID, *booking.PetID)
			booking.ID = 555
			return booking, nil
		},
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, typeRepo, &mockTransactor{}, nil, nil, nil, &mockLineReservationSettingFinder{}, &mockClinicHolidayFinder{})

	got, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		OwnerID:           &ownerID,
		PetID:             &petID,
		ReservationTypeID: 9,
		Status:            model.ReservationStatusConfirmed,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, absorbCalled, "Create は tentative absorb 入口を通る")
	assert.False(t, createCalled, "Create は repo.Create を直接呼ばない")
	assert.Equal(t, uint64(555), got.ID)
}

func TestReservationService_Create_CheckedInDelegatesToPromoteEntrypoint(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	promoteCalled, absorbCalled := false, false

	repo := &mockReservationRepository{
		absorbTentativeOnBookingFn: func(_ context.Context, booking *model.Reservation) (*model.Reservation, error) {
			absorbCalled = true
			return booking, nil
		},
		promoteTentativeOnCheckinFn: func(_ context.Context, booking *model.Reservation) (*model.Reservation, error) {
			promoteCalled = true
			booking.ID = 777
			return booking, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error { return nil },
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, typeRepo, &mockTransactor{}, nil, nil, nil, &mockLineReservationSettingFinder{}, &mockClinicHolidayFinder{})

	got, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(15 * time.Minute),
		ReservationTypeID: 9,
		Status:            model.ReservationStatusCheckedIn,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, promoteCalled, "checked_in 作成は check-in promote 入口を通る")
	assert.False(t, absorbCalled, "checked_in 作成は booking absorb 入口を通らない")
}

func TestReservationService_CreateBatch_AbsorbsTentativePerPet(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	var absorbedPets []uint64

	repo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) { return 1, nil },
		countConflictsFn:     func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) { return 0, nil },
		findPetOwnerInClinicFn: func(_ context.Context, _ uint64, petID uint64) (uint64, error) {
			return petID * 10, nil
		},
		absorbTentativeOnBookingFn: func(_ context.Context, booking *model.Reservation) (*model.Reservation, error) {
			absorbedPets = append(absorbedPets, *booking.PetID)
			return booking, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Error("CreateBatch は repo.Create を直接呼ばない")
			return nil
		},
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, typeRepo, &mockTransactor{}, nil, nil, nil, &mockLineReservationSettingFinder{}, &mockClinicHolidayFinder{})

	got, err := svc.CreateBatch(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 9,
		Status:            model.ReservationStatusConfirmed,
	}, []ReservationBatchPet{
		{OwnerID: 10, PetID: 1},
		{OwnerID: 20, PetID: 2},
	})
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.ElementsMatch(t, []uint64{1, 2}, absorbedPets, "一括作成はペットごとに absorb 入口を通る")
}

func TestReservationAdminService_Create_UsesTentativeAbsorbEntrypoint(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	ownerID, petID := uint64(11), uint64(22)
	absorbCalled := false

	resRepo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) { return 1, nil },
		countConflictsFn:     func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) { return 0, nil },
		findPetOwnerInClinicFn: func(_ context.Context, _ uint64, _ uint64) (uint64, error) {
			return ownerID, nil
		},
		absorbTentativeOnBookingFn: func(_ context.Context, booking *model.Reservation) (*model.Reservation, error) {
			absorbCalled = true
			booking.ID = 888
			return booking, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Error("admin Create は repo.Create を直接呼ばない")
			return nil
		},
	}
	svc := NewReservationAdminServiceWithClinicHolidays(
		&mockReservationAdminRepository{}, resRepo, nil, &mockTransactor{}, nil, nil, nil, openDayHolidayFinder())

	got, err := svc.Create(context.Background(), 1, &CreateReservationAdminInput{
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		OwnerID:           &ownerID,
		PetID:             &petID,
		ReservationTypeID: 9,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, absorbCalled, "walk-in/admin Create は tentative absorb 入口を通る")
	assert.Equal(t, uint64(888), got.ID)
}

func TestLiffService_CreateReservation_PreresolvesOwnerPetForAbsorb(t *testing.T) {
	ownerID, petID := uint64(31), uint64(42)
	var captured *CreateReservationInput

	svc := newReservationTestSvc(func(s *liffService) {
		s.customerRepo = &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:      1,
					OwnerID: &ownerID,
					Owner: &model.Owner{
						ID:   ownerID,
						Pets: []model.Pet{{ID: petID, Status: model.PetStatusAlive}},
					},
				}, nil
			},
		}
		s.validators = &mockLiffValidators{
			validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
				captured = input
				return &model.Reservation{ID: 99, ClinicID: 3, OwnerID: input.OwnerID, PetID: input.PetID}, nil
			},
		}
	})

	got, err := svc.CreateReservation(context.Background(), 3, 1, reservationBaseInput())
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, captured)
	require.NotNil(t, captured.OwnerID, "LIFF は作成前に owner を解決して validators へ渡す")
	assert.Equal(t, ownerID, *captured.OwnerID)
	require.NotNil(t, captured.PetID, "LIFF は作成前に pet を解決して validators へ渡す（tentative absorb の照合キー）")
	assert.Equal(t, petID, *captured.PetID)
}

func TestLiffService_CreateReservation_UnlinkedCustomerCreatesWithoutPet(t *testing.T) {
	var captured *CreateReservationInput
	svc := newReservationTestSvc(func(s *liffService) {
		// 未リンク顧客（OwnerID nil）: tentative 照合不能のため従来どおり nil のまま作成。
		s.customerRepo = &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{ID: 1}, nil
			},
		}
		s.validators = &mockLiffValidators{
			validateAndCreateFn: func(_ context.Context, input *CreateReservationInput) (*model.Reservation, error) {
				captured = input
				return &model.Reservation{ID: 100, ClinicID: 3}, nil
			},
		}
	})

	_, err := svc.CreateReservation(context.Background(), 3, 1, reservationBaseInput())
	require.NoError(t, err)
	require.NotNil(t, captured)
	assert.Nil(t, captured.OwnerID)
	assert.Nil(t, captured.PetID)
}

// ---------------------------------------------------------------------------
// A2: 当日受付 PATCH — tentative → checked_in で sentinel 時刻を実受付時刻に置換
// ---------------------------------------------------------------------------

func TestReservationService_Update_TentativeCheckinFillsActualTime(t *testing.T) {
	sentinelStart := time.Date(2030, 1, 5, 0, 0, 0, 0, config.JST)
	sentinelEnd := sentinelStart.Add(nextVisitIntentDurationMinutes * time.Minute)
	petID := uint64(22)
	tentative := &model.Reservation{
		ID:                42,
		ClinicID:          1,
		PetID:             &petID,
		Status:            model.ReservationStatusTentative,
		StartTime:         sentinelStart,
		EndTime:           sentinelEnd,
		ReservationTypeID: 7,
		Source:            model.ReservationSourceManual,
	}

	var captured map[string]any
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.Reservation, error) {
			return tentative, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.Reservation, error) {
			return tentative, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			captured = fields
			return tentative, nil
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, nil, &mockTransactor{}, nil, nil, nil, &mockLineReservationSettingFinder{}, &mockClinicHolidayFinder{})

	checkedIn := model.ReservationStatusCheckedIn
	before := time.Now()
	_, err := svc.Update(context.Background(), 1, 42, &UpdateReservationInput{Status: &checkedIn})
	require.NoError(t, err)
	require.NotNil(t, captured)

	assert.Equal(t, model.ReservationStatusCheckedIn, captured["status"])
	require.Contains(t, captured, "checked_in_at")
	startVal, ok := captured["start_time"].(time.Time)
	require.True(t, ok, "tentative → checked_in は start_time を実受付時刻へ置き換える")
	assert.False(t, startVal.Before(before.Add(-time.Second)), "start_time は受付実行時刻前後")
	endVal, ok := captured["end_time"].(time.Time)
	require.True(t, ok)
	assert.Equal(t, nextVisitIntentDurationMinutes*time.Minute, endVal.Sub(startVal))
}

func TestReservationService_Update_NonTentativeCheckinKeepsSchedule(t *testing.T) {
	petID := uint64(22)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	confirmed := &model.Reservation{
		ID:                43,
		ClinicID:          1,
		PetID:             &petID,
		Status:            model.ReservationStatusConfirmed,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 7,
	}
	var captured map[string]any
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.Reservation, error) {
			return confirmed, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			captured = fields
			return confirmed, nil
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, nil, &mockTransactor{}, nil, nil, nil, &mockLineReservationSettingFinder{}, &mockClinicHolidayFinder{})

	checkedIn := model.ReservationStatusCheckedIn
	_, err := svc.Update(context.Background(), 1, 43, &UpdateReservationInput{Status: &checkedIn})
	require.NoError(t, err)
	require.NotNil(t, captured)
	assert.NotContains(t, captured, "start_time", "通常予約の受付は予約時刻を書き換えない")
	assert.NotContains(t, captured, "end_time")
	assert.Contains(t, captured, "checked_in_at")
}

// ---------------------------------------------------------------------------
// A3: キャンセルは tentative 由来行を自動復活させず、provenance を保持する
// ---------------------------------------------------------------------------

func TestReservationService_Update_CancelConvertedRowDoesNotRevive(t *testing.T) {
	petID := uint64(22)
	intended := time.Date(2030, 3, 10, 0, 0, 0, 0, config.JST)
	converted := &model.Reservation{
		ID:                50,
		ClinicID:          1,
		PetID:             &petID,
		Status:            model.ReservationStatusConfirmed,
		StartTime:         time.Date(2030, 3, 10, 10, 0, 0, 0, config.JST),
		EndTime:           time.Date(2030, 3, 10, 10, 30, 0, 0, config.JST),
		ReservationTypeID: 7,
		IntendedDate:      &intended,
	}
	intentWriteCalled := false
	var updateFields map[string]any
	deleteCalled := false

	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.Reservation, error) {
			return converted, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			updateFields = fields
			return converted, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
		setNextVisitIntentFn: func(_ context.Context, _, _ uint64, _ SetNextVisitIntentInput) (*model.Reservation, error) {
			intentWriteCalled = true
			return nil, nil
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, nil, &mockTransactor{}, nil, nil, nil, &mockLineReservationSettingFinder{}, &mockClinicHolidayFinder{})

	cancelled := model.ReservationStatusCancelled
	_, err := svc.Update(context.Background(), 1, 50, &UpdateReservationInput{Status: &cancelled})
	require.NoError(t, err)
	require.NotNil(t, updateFields)
	assert.Equal(t, model.ReservationStatusCancelled, updateFields["status"])
	assert.True(t, deleteCalled, "cancel は soft delete を伴う")
	assert.False(t, intentWriteCalled, "cancel は tentative を自動復活（SetNextVisitIntent）しない")
}

func TestToReservationResponse_HadIntendedDate(t *testing.T) {
	intended := time.Date(2030, 3, 10, 0, 0, 0, 0, config.JST)
	withIntent := toReservationResponse(&model.Reservation{
		ID:           1,
		ClinicID:     1,
		Status:       model.ReservationStatusCancelled,
		IntendedDate: &intended,
	})
	assert.True(t, withIntent.HadIntendedDate, "cancelled でも intended_date が残る行は true")
	require.NotNil(t, withIntent.IntendedDate)
	assert.Equal(t, "2030-03-10", *withIntent.IntendedDate)

	withoutIntent := toReservationResponse(&model.Reservation{
		ID:       2,
		ClinicID: 1,
		Status:   model.ReservationStatusPending,
	})
	assert.False(t, withoutIntent.HadIntendedDate)
	assert.Nil(t, withoutIntent.IntendedDate)
}
