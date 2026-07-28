package reservation

// reservation_cross_tenant_master_fk_write_test.go — BE9-2C R③:
// service/cross_tenant_master_fk_write_test.go から reservationValidators/reservationService 節を同名移動。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestReservationValidators_ValidateAndCreate_RejectsCrossClinicReservationType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(50)
	const foreignTypeID = uint64(999)

	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
			if id != ownedTypeID {
				return nil, apperrors.WrapNotFound("reservation_type", "foreign")
			}
			return &model.ReservationType{ID: id, IsActive: true, ReservationVisible: true}, nil
		},
	}

	newSvc := func(created *bool) ReservationValidators {
		repo := &mockReservationRepository{
			createFn: func(_ context.Context, _ *model.Reservation) error {
				*created = true
				return nil
			},
		}
		return NewReservationValidators(&mockTransactor{}, repo, typeRepo, &mockReservationStaffRepositoryForCapability{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTrimmingDetailRepository{})
	}

	baseInput := func(typeID uint64) *CreateReservationInput {
		return &CreateReservationInput{
			ClinicID:          clinicID,
			CustomerID:        2,
			ReservationTypeID: typeID,
			StaffID:           10,
			Date:              dateInDays(3),
			StartTime:         "1000",
			EndTime:           "1015",
			Settings:          newSettingForValidation(),
		}
	}

	t.Run("rejects cross-clinic reservation_type_id and does not persist", func(t *testing.T) {
		created := false
		validators := newSvc(&created)
		out, err := validators.ValidateAndCreate(context.Background(), baseInput(foreignTypeID))
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "reservation must NOT be persisted referencing another clinic's reservation type")
	})

	t.Run("accepts same-clinic reservation_type_id (no false-reject)", func(t *testing.T) {
		created := false
		validators := newSvc(&created)
		out, err := validators.ValidateAndCreate(context.Background(), baseInput(ownedTypeID))
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestReservationValidators_ValidateAndCreate_RejectsCrossClinicTrimmingFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(50)
	const ownedCourseID = uint64(300)
	const foreignCourseID = uint64(999)
	const ownedOptionID = uint64(400)
	const foreignOptionID = uint64(998)

	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:                 id,
				Category:           model.ReservationTypeCategoryTrimming,
				IsActive:           true,
				ReservationVisible: true,
			}, nil
		},
	}

	newSvc := func(created *bool, courseRepo trimmingCourseFinder, optionRepo trimmingOptionFinder) ReservationValidators {
		repo := &mockReservationRepository{
			createFn: func(_ context.Context, _ *model.Reservation) error {
				*created = true
				return nil
			},
		}
		return NewReservationValidators(&mockTransactor{}, repo, typeRepo, &mockReservationStaffRepositoryForCapability{}, courseRepo, optionRepo, &mockTrimmingDetailRepository{})
	}

	baseInput := func() *CreateReservationInput {
		return &CreateReservationInput{
			ClinicID:          clinicID,
			CustomerID:        2,
			ReservationTypeID: ownedTypeID,
			StaffID:           10,
			Date:              dateInDays(3),
			StartTime:         "1000",
			EndTime:           "1015",
			Settings:          newSettingForValidation(),
		}
	}

	t.Run("rejects cross-clinic trimming_course_id and does not persist", func(t *testing.T) {
		created := false
		validators := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID), okTrimmingOptionRepo())
		foreign := foreignCourseID
		input := baseInput()
		input.TrimmingCourseID = &foreign
		out, err := validators.ValidateAndCreate(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "reservation must NOT be persisted referencing another clinic's trimming course")
	})

	t.Run("rejects cross-clinic trimming_option_ids and does not persist", func(t *testing.T) {
		created := false
		validators := newSvc(&created, okTrimmingCourseRepo(), rejectTrimmingOptionRepo(ownedOptionID))
		input := baseInput()
		input.TrimmingOptionIDs = []uint64{ownedOptionID, foreignOptionID}
		out, err := validators.ValidateAndCreate(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "reservation must NOT be persisted referencing another clinic's trimming option")
	})

	t.Run("accepts same-clinic trimming_course_id/trimming_option_ids (no false-reject)", func(t *testing.T) {
		created := false
		validators := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID), rejectTrimmingOptionRepo(ownedOptionID))
		owned := ownedCourseID
		input := baseInput()
		input.TrimmingCourseID = &owned
		input.TrimmingOptionIDs = []uint64{ownedOptionID}
		out, err := validators.ValidateAndCreate(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestReservationValidators_ValidateAndCreate_ValidatesPublicStaffInsideTransaction(t *testing.T) {
	const (
		clinicID          = uint64(1)
		reservationTypeID = uint64(50)
		staffID           = uint64(10)
	)
	type txMarkerKey struct{}
	baseInput := func() *CreateReservationInput {
		return &CreateReservationInput{
			ClinicID:          clinicID,
			CustomerID:        2,
			ReservationTypeID: reservationTypeID,
			StaffID:           staffID,
			Date:              dateInDays(3),
			StartTime:         "1000",
			EndTime:           "1015",
			Settings:          newSettingForValidation(),
		}
	}
	typeRepo := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{
			ID:                 id,
			ClinicID:           clinicID,
			IsActive:           true,
			ReservationVisible: true,
			Category:           model.ReservationTypeCategoryGeneral,
		}, nil
	}}

	t.Run("rejects a staff member not assigned to the clinic", func(t *testing.T) {
		created := false
		repo := &mockReservationRepository{createFn: func(_ context.Context, _ *model.Reservation) error {
			created = true
			return nil
		}}
		staffRepo := &mockReservationStaffRepositoryForCapability{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
				return nil, apperrors.WrapNotFound("reservation_staff", "foreign")
			},
		}
		validators := NewReservationValidators(
			&mockTransactor{}, repo, typeRepo, staffRepo,
			okTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTrimmingDetailRepository{},
		)

		out, err := validators.ValidateAndCreate(context.Background(), baseInput())

		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("checks assignment and capability in the write transaction", func(t *testing.T) {
		staffCheckedInTx := false
		capabilityCheckedInTx := false
		tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, txMarkerKey{}, true))
		}}
		staffRepo := &mockReservationStaffRepositoryForCapability{
			findByIDFn: func(ctx context.Context, _, id uint64) (*model.Staff, error) {
				staffCheckedInTx = ctx.Value(txMarkerKey{}) == true
				return &model.Staff{ID: id, IsActive: true, ReservationVisible: true}, nil
			},
			supportsReservationTypeFn: func(ctx context.Context, _, _, _ uint64) (bool, error) {
				capabilityCheckedInTx = ctx.Value(txMarkerKey{}) == true
				return true, nil
			},
		}
		validators := NewReservationValidators(
			tx, &mockReservationRepository{createFn: func(_ context.Context, _ *model.Reservation) error { return nil }}, typeRepo, staffRepo,
			okTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTrimmingDetailRepository{},
		)

		out, err := validators.ValidateAndCreate(context.Background(), baseInput())

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, staffCheckedInTx)
		assert.True(t, capabilityCheckedInTx)
	})
}

func TestReservationValidators_ValidateAndCreate_TrimmingWritesAreAtomicAndActive(t *testing.T) {
	const (
		clinicID          = uint64(1)
		reservationTypeID = uint64(50)
		courseID          = uint64(300)
		optionID          = uint64(400)
	)
	typeRepo := mockReservationTypeFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{
			ID:                 id,
			ClinicID:           clinicID,
			IsActive:           true,
			ReservationVisible: true,
			Category:           model.ReservationTypeCategoryTrimming,
		}, nil
	}}
	baseInput := func() *CreateReservationInput {
		course := courseID
		return &CreateReservationInput{
			ClinicID:          clinicID,
			CustomerID:        2,
			ReservationTypeID: reservationTypeID,
			Date:              dateInDays(3),
			StartTime:         "1000",
			EndTime:           "1015",
			Settings:          newSettingForValidation(),
			TrimmingCourseID:  &course,
			TrimmingOptionIDs: []uint64{optionID},
		}
	}

	t.Run("rejects inactive trimming master before appointment insert", func(t *testing.T) {
		created := false
		courseRepo := &mockTrimmingCourseFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
			return &model.TrimmingCourse{ID: id, ClinicID: clinicID, IsActive: false}, nil
		}}
		optionRepo := &mockTrimmingOptionFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
			return &model.TrimmingOption{ID: id, ClinicID: clinicID, IsActive: true}, nil
		}}
		validators := NewReservationValidators(
			&mockTransactor{},
			&mockReservationRepository{createFn: func(_ context.Context, _ *model.Reservation) error { created = true; return nil }},
			typeRepo, &mockReservationStaffRepositoryForCapability{}, courseRepo, optionRepo, &mockTrimmingDetailRepository{},
		)

		out, err := validators.ValidateAndCreate(context.Background(), baseInput())

		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("detail failure fails the reservation transaction", func(t *testing.T) {
		created := false
		detailErr := errors.New("detail insert failed")
		courseRepo := &mockTrimmingCourseFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
			return &model.TrimmingCourse{ID: id, ClinicID: clinicID, IsActive: true}, nil
		}}
		optionRepo := &mockTrimmingOptionFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
			return &model.TrimmingOption{ID: id, ClinicID: clinicID, IsActive: true}, nil
		}}
		detailRepo := &mockTrimmingDetailRepository{createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			return detailErr
		}}
		validators := NewReservationValidators(
			&mockTransactor{},
			&mockReservationRepository{createFn: func(_ context.Context, appointment *model.Reservation) error {
				created = true
				appointment.ID = 99
				return nil
			}},
			typeRepo, &mockReservationStaffRepositoryForCapability{}, courseRepo, optionRepo, detailRepo,
		)

		out, err := validators.ValidateAndCreate(context.Background(), baseInput())

		assert.ErrorIs(t, err, detailErr)
		assert.Nil(t, out)
		assert.True(t, created, "appointment insert is attempted inside the transaction before the injected detail failure")
	})
}

func TestReservationService_Create_RejectsCrossClinicReservationType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(50)
	const foreignTypeID = uint64(999)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationType, error) {
			if gotClinicID != clinicID || id != ownedTypeID {
				return nil, apperrors.WrapNotFound("reservation_type", "foreign")
			}
			return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
		},
	}

	newSvc := func(created *bool) ReservationService {
		repo := &mockReservationRepository{
			createFn: func(_ context.Context, _ *model.Reservation) error {
				*created = true
				return nil
			},
		}
		return NewReservationServiceWithAvailabilityAndType(repo, typeRepo, &mockTransactor{}, nil, nil)
	}

	baseInput := func(typeID uint64, route *string, status model.ReservationStatus) *CreateManualReservationInput {
		return &CreateManualReservationInput{
			ClinicID:          clinicID,
			StartTime:         start,
			EndTime:           start.Add(30 * time.Minute),
			ReservationTypeID: typeID,
			Status:            status,
			ReservationRoute:  route,
		}
	}

	// shortcut 経路(reception/exam_room/record_shortcut)は enforceBookingConstraints=false
	// となり容量チェック(FindByID)がスキップされる経路 — U6b で塞いだ真の穴。
	routes := []*string{ptrString("reception"), ptrString("exam_room"), ptrString("record_shortcut")}
	for _, route := range routes {
		t.Run("rejects cross-clinic reservation_type_id via shortcut route "+*route, func(t *testing.T) {
			created := false
			svc := newSvc(&created)
			out, err := svc.Create(context.Background(), baseInput(foreignTypeID, route, model.ReservationStatusPending))
			assert.Error(t, err)
			assert.Nil(t, out)
			assert.False(t, created, "reservation must NOT be persisted via shortcut route referencing another clinic's reservation type")
		})

		t.Run("accepts same-clinic reservation_type_id via shortcut route "+*route+" (no false-reject)", func(t *testing.T) {
			created := false
			svc := newSvc(&created)
			out, err := svc.Create(context.Background(), baseInput(ownedTypeID, route, model.ReservationStatusPending))
			assert.NoError(t, err)
			assert.NotNil(t, out)
			assert.True(t, created)
		})
	}

	t.Run("rejects cross-clinic reservation_type_id via normal (enforced) route", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), baseInput(foreignTypeID, nil, model.ReservationStatusPending))
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created)
	})
}

func TestReservationService_Update_RejectsCrossClinicReservationType(t *testing.T) {
	const clinicID = uint64(1)
	const reservationID = uint64(7)
	const ownedTypeID = uint64(50)
	const foreignTypeID = uint64(999)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationType, error) {
			if gotClinicID != clinicID || id != ownedTypeID {
				return nil, apperrors.WrapNotFound("reservation_type", "foreign")
			}
			return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
		},
	}

	newSvc := func(updated *bool) ReservationService {
		repo := &mockReservationRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, ClinicID: gotClinicID, StartTime: start, EndTime: start.Add(30 * time.Minute), ReservationTypeID: ownedTypeID}, nil
			},
			lockAndFindByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, ClinicID: gotClinicID, StartTime: start, EndTime: start.Add(30 * time.Minute), ReservationTypeID: ownedTypeID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
				*updated = true
				return &model.Reservation{ID: reservationID, ClinicID: clinicID}, nil
			},
		}
		return NewReservationServiceWithAvailabilityAndType(repo, typeRepo, &mockTransactor{}, nil, nil)
	}

	t.Run("rejects cross-clinic reservation_type_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignTypeID
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{ReservationTypeID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "reservation must NOT be updated to reference another clinic's reservation type")
	})

	t.Run("accepts same-clinic reservation_type_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedTypeID
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{ReservationTypeID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

func TestReservationAdminService_Create_RejectsCrossClinicReservationType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(50)
	const foreignTypeID = uint64(999)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationType, error) {
			if gotClinicID != clinicID || id != ownedTypeID {
				return nil, apperrors.WrapNotFound("reservation_type", "foreign")
			}
			return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
		},
	}

	newSvc := func(created *bool) ReservationAdminService {
		resRepo := &mockReservationRepository{
			createFn: func(_ context.Context, _ *model.Reservation) error {
				*created = true
				return nil
			},
		}
		return NewReservationAdminServiceWithAvailabilityAndType(
			&mockReservationAdminRepository{}, resRepo, typeRepo, &mockTransactor{}, nil, nil,
		)
	}

	baseInput := func(typeID uint64) *CreateReservationAdminInput {
		return &CreateReservationAdminInput{
			StartTime:         start,
			EndTime:           start.Add(30 * time.Minute),
			ReservationTypeID: typeID,
		}
	}

	t.Run("rejects cross-clinic reservation_type_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, baseInput(foreignTypeID))
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "admin reservation must NOT be persisted referencing another clinic's reservation type")
	})

	t.Run("accepts same-clinic reservation_type_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, baseInput(ownedTypeID))
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

// TestLiffService_CreateReservation_RejectsCrossClinicTrimmingFK は liff 経由でも
// ValidateAndCreate の所有権ガードが効き、appointment が永続化されないことを検証する
// (U6a: liffService は validators に委譲するのみで、ガード本体は validators 側にある)。
func TestLiffService_CreateReservation_RejectsCrossClinicTrimmingFK(t *testing.T) {
	const clinicID = uint64(3)
	const customerID = uint64(1)
	const ownedCourseID = uint64(300)
	const foreignCourseID = uint64(999)

	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:                 id,
				Category:           model.ReservationTypeCategoryTrimming,
				IsActive:           true,
				ReservationVisible: true,
			}, nil
		},
	}

	newSvc := func(created *bool, courseRepo trimmingCourseFinder) *liffService {
		reservationRepo := &mockReservationRepository{
			createFn: func(_ context.Context, _ *model.Reservation) error {
				*created = true
				return nil
			},
		}
		return &liffService{
			settingRepo: &mockLiffSettingRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return newSettingForValidation(), nil
				},
			},
			customerRepo: &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					return &model.LineCustomer{ID: customerID}, nil
				},
			},
			ownerRepo:          nil,
			reservationRepo:    reservationRepo,
			trimmingDetailRepo: &mockTrimmingDetailRepository{},
			notifier:           nil,
			validators:         NewReservationValidators(&mockTransactor{}, reservationRepo, typeRepo, &mockReservationStaffRepositoryForCapability{}, courseRepo, okTrimmingOptionRepo(), &mockTrimmingDetailRepository{}),
		}
	}

	baseInput := func() *CreateReservationInput {
		return &CreateReservationInput{
			ReservationTypeID: 1,
			StaffID:           10,
			Date:              dateInDays(3),
			StartTime:         "1000",
			EndTime:           "1015",
		}
	}

	t.Run("rejects cross-clinic trimming course and does not create appointment", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID))
		foreign := foreignCourseID
		input := baseInput()
		input.TrimmingCourseID = &foreign

		out, err := svc.CreateReservation(context.Background(), clinicID, customerID, input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "appointment must NOT be persisted referencing another clinic's trimming course")
	})

	t.Run("accepts same-clinic trimming course (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID))
		owned := ownedCourseID
		input := baseInput()
		input.TrimmingCourseID = &owned

		out, err := svc.CreateReservation(context.Background(), clinicID, customerID, input)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}
