package service

// Cross-tenant master FK write isolation tests (#201 follow-up).
//
// These tests close the "clinic-scoped master FK write hole" class: a service
// Create/Update persisting a request-derived master FK (medicine/vaccine/
// exam_type/procedure/consultation/checkup_type/diagnosis) WITHOUT verifying the
// master belongs to the caller's clinic — the source of #124/#125-type mislinks.
//
// Each fix gets a dual regression test:
//   (i)  clinic A referencing clinic B's master  -> rejected, row NOT persisted
//   (ii) same-clinic master                      -> still succeeds (no false-reject)
//
// Shared permissive ("ok*") and rejecting ("reject*") master-repo mock builders
// live here so every service test reuses one tenancy model. The reject builders
// return NotFound for IDs not owned by the caller clinic, exactly mirroring the
// real *_repository.go FindByID(ctx, clinicID, id) clinic-scoped behavior.

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ── permissive builders: any FindByID succeeds (used to keep existing tests green) ──

func okMedicineRepo() repository.MedicineRepository {
	return &mockMedicineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
		return &model.Medicine{ID: id}, nil
	}}
}

func okProcedureRepo() repository.ProcedureRepository {
	return &mockProcedureRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
		return &model.Procedure{ID: id}, nil
	}}
}

func okConsultationRepo() repository.ConsultationRepository {
	return &mockConsultationRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Consultation, error) {
		return &model.Consultation{ID: id}, nil
	}}
}

func okExamTypeRepo() repository.ExamTypeRepository {
	return &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
		return &model.ExaminationType{ID: id}, nil
	}}
}

func okTrimmingCourseRepo() repository.TrimmingCourseRepository {
	return &mockTrimmingCourseRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		return &model.TrimmingCourse{ID: id, IsActive: true}, nil
	}}
}

func okTrimmingOptionRepo() repository.TrimmingOptionRepository {
	return &mockTrimmingOptionRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		return &model.TrimmingOption{ID: id, IsActive: true}, nil
	}}
}

func okDiagnosisTypeRepo() repository.DiagnosisTypeRepository {
	return &mockDiagnosisTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
		return &model.DiagnosisType{ID: id}, nil
	}}
}

func okDiagnosisNameRepo() repository.DiagnosisNameRepository {
	return &mockDiagnosisNameRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
		return &model.DiagnosisName{ID: id}, nil
	}}
}

func okChiefComplaintTypeRepo() repository.ChiefComplaintTypeRepository {
	return &mockChiefComplaintTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ChiefComplaintType, error) {
		return &model.ChiefComplaintType{ID: id}, nil
	}}
}

// okInventoryRepo also wires a no-op createFn: medicineService.Create unconditionally
// creates a linked InventoryItem (BUG-320) regardless of the ParentID/InventoryID guards
// under test here, and mockInventoryRepository.Create has no nil-createFn guard.
func okInventoryRepo() repository.InventoryRepository {
	return &mockInventoryRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.InventoryItem, error) {
			return &model.InventoryItem{ID: id}, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.InventoryItem) error { return nil },
	}
}

// rejectByForeignID returns a clinic-scoped FindByID stub: it returns the row for
// ownedID and NotFound for anything else, mimicking real FindByID(ctx, clinicID, id).
func rejectMedicineRepo(ownedID uint64) repository.MedicineRepository {
	return &mockMedicineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("medicine", "foreign")
		}
		return &model.Medicine{ID: id}, nil
	}}
}

func rejectProcedureRepo(ownedID uint64) repository.ProcedureRepository {
	return &mockProcedureRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("procedure", "foreign")
		}
		return &model.Procedure{ID: id}, nil
	}}
}

func rejectInsuranceRepo(ownedID uint64) repository.InsuranceRepository {
	return &mockInsuranceRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Insurance, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("insurance", "foreign")
		}
		return &model.Insurance{ID: id}, nil
	}}
}

// rejectInventoryRepo also wires a no-op createFn for the same BUG-320 reason as
// okInventoryRepo above — the "accepts same-clinic" sub-tests reach the auto-create step.
func rejectInventoryRepo(ownedID uint64) repository.InventoryRepository {
	return &mockInventoryRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.InventoryItem, error) {
			if id != ownedID {
				return nil, apperrors.WrapNotFound("inventory_item", "foreign")
			}
			return &model.InventoryItem{ID: id}, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.InventoryItem) error { return nil },
	}
}

// rejectOccupationRepo mirrors rejectInventoryRepo for staffService.OccupationID (X-14 U7).
// mockOccupationRepository is defined in occupation_service_test.go (same package).
func rejectOccupationRepo(ownedID uint64) repository.OccupationRepository {
	return &mockOccupationRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Occupation, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("occupation", "foreign")
		}
		return &model.Occupation{ID: id}, nil
	}}
}

// ── trimmingCourseService CourseTypeID (X-14b): Update now mirrors Create's
// pre-persist courseTypeRepo.FindByID guard (symmetric with Create). ──

func TestTrimmingCourseService_Update_RejectsCrossClinicCourseTypeFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedCourseTypeID = uint64(10)
	const foreignCourseTypeID = uint64(999)

	newSvc := func(updated *bool) TrimmingCourseService {
		repo := &mockTrimmingCourseRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
				return &model.TrimmingCourse{ID: id}, nil
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.TrimmingCourse, error) {
				*updated = true
				return &model.TrimmingCourse{ID: id}, nil
			},
		}
		courseTypeRepo := &mockTrimmingCourseTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourseType, error) {
				if id != ownedCourseTypeID {
					return nil, apperrors.WrapNotFound("trimming_course_type", "foreign")
				}
				return &model.TrimmingCourseType{ID: id}, nil
			},
		}
		return NewTrimmingCourseService(repo, courseTypeRepo)
	}

	t.Run("rejects cross-clinic course_type_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignCourseTypeID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateTrimmingCourseInput{CourseTypeID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "trimming course must NOT be updated to reference another clinic's course type")
	})

	t.Run("accepts same-clinic course_type_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedCourseTypeID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateTrimmingCourseInput{CourseTypeID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── vaccination (CRITICAL #125): vaccine_id ──

// ── billing_item (P1, X-4): trimming_course_id / trimming_option_id ──

func rejectTrimmingCourseRepo(ownedID uint64) repository.TrimmingCourseRepository {
	return &mockTrimmingCourseRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_course", "foreign")
		}
		return &model.TrimmingCourse{ID: id, IsActive: true}, nil
	}}
}

func rejectTrimmingOptionRepo(ownedID uint64) repository.TrimmingOptionRepository {
	return &mockTrimmingOptionRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_option", "foreign")
		}
		return &model.TrimmingOption{ID: id, IsActive: true}, nil
	}}
}

func TestBillingItemService_CreateItem_RejectsCrossClinicTrimmingFK(t *testing.T) {
	const clinicID = uint64(1)
	const billingID = uint64(10)
	const ownedCourseID = uint64(300)
	const foreignCourseID = uint64(999)
	const ownedOptionID = uint64(400)
	const foreignOptionID = uint64(998)

	newSvc := func(created *bool, courseRepo repository.TrimmingCourseRepository, optionRepo repository.TrimmingOptionRepository) BillingItemService {
		repo := defaultMockBillingItemRepo()
		repo.createFn = func(_ context.Context, item *model.BillingItem) error { *created = true; item.ID = 1; return nil }
		billingRepo := &mockAccountingRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id}, nil
		}}
		return NewBillingItemServiceWithCampaign(repo, billingRepo, defaultMockTreatmentRepo(), &mockTransactor{}, courseRepo, optionRepo, nil, nil)
	}

	baseInput := func() *CreateBillingItemInput {
		return &CreateBillingItemInput{
			ClinicID:  clinicID,
			BillingID: billingID,
			Category:  string(model.ItemCategoryProcedure),
			Name:      "トリミング",
			UnitPrice: 5000,
			Quantity:  1,
		}
	}

	t.Run("rejects cross-clinic trimming_course_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID), okTrimmingOptionRepo())
		foreign := foreignCourseID
		input := baseInput()
		input.TrimmingCourseID = &foreign
		out, err := svc.CreateItem(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "billing item must NOT be persisted referencing another clinic's trimming course")
	})

	t.Run("rejects cross-clinic trimming_option_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, okTrimmingCourseRepo(), rejectTrimmingOptionRepo(ownedOptionID))
		foreign := foreignOptionID
		input := baseInput()
		input.TrimmingOptionID = &foreign
		out, err := svc.CreateItem(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "billing item must NOT be persisted referencing another clinic's trimming option")
	})

	t.Run("accepts same-clinic trimming_course_id/trimming_option_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID), rejectTrimmingOptionRepo(ownedOptionID))
		owned1, owned2 := ownedCourseID, ownedOptionID
		input := baseInput()
		input.TrimmingCourseID = &owned1
		input.TrimmingOptionID = &owned2
		out, err := svc.CreateItem(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

// ── reservationValidators.ValidateAndCreate / liffService.CreateReservation
//    (X-14/U6a): ReservationTypeID / TrimmingCourseID / TrimmingOptionIDs ──

func TestReservationValidators_ValidateAndCreate_RejectsCrossClinicReservationType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(50)
	const foreignTypeID = uint64(999)

	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
			if id != ownedTypeID {
				return nil, apperrors.WrapNotFound("reservation_type", "foreign")
			}
			return &model.ReservationType{ID: id}, nil
		},
	}

	newSvc := func(created *bool) ReservationValidators {
		repo := &mockReservationRepository{
			createFn: func(_ context.Context, _ *model.Reservation) error {
				*created = true
				return nil
			},
		}
		return NewReservationValidators(&mockTransactor{}, repo, typeRepo, okTrimmingCourseRepo(), okTrimmingOptionRepo())
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
			return &model.ReservationType{ID: id}, nil
		},
	}

	newSvc := func(created *bool, courseRepo repository.TrimmingCourseRepository, optionRepo repository.TrimmingOptionRepository) ReservationValidators {
		repo := &mockReservationRepository{
			createFn: func(_ context.Context, _ *model.Reservation) error {
				*created = true
				return nil
			},
		}
		return NewReservationValidators(&mockTransactor{}, repo, typeRepo, courseRepo, optionRepo)
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
			return &model.ReservationType{ID: id}, nil
		},
	}

	newSvc := func(created *bool, courseRepo repository.TrimmingCourseRepository) *liffService {
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
			validators:         NewReservationValidators(&mockTransactor{}, reservationRepo, typeRepo, courseRepo, okTrimmingOptionRepo()),
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

func rejectDiagnosisNameRepo(ownedID uint64) repository.DiagnosisNameRepository {
	return &mockDiagnosisNameRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("diagnosis_name", "foreign")
		}
		return &model.DiagnosisName{ID: id}, nil
	}}
}

// ── accounting (G14-1): payment_method_id ──
//
// payment_method_id is not guarded by a plain FindByID(clinicID, id) call like the
// other master FKs above — it is guarded by resolvePaymentMethodMasterID
// (accounting_service_builders.go), which resolves the request's method(ENUM) to
// the caller clinic's payment_methods master id and rejects any explicitly
// supplied id that does not match ("... or equivalent" per the lint's own
// guarded definition). This test proves that guard holds against a *real*
// foreign clinic's resolved master id, not just an arbitrary mismatched number.

func TestAccountingService_Update_RejectsForeignPaymentMethodID(t *testing.T) {
	const clinicA = uint64(1)
	const clinicB = uint64(2)
	const clinicACashID = uint64(101)
	const clinicBCashID = uint64(201)
	billingAmount := int64(5000)
	cashKey := "cash"

	payMethodRepo := &mockPaymentMethodMasterRepository{
		findAllFn: func(_ context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
			switch clinicID {
			case clinicA:
				return []model.PaymentMethodMaster{{ID: clinicACashID, ClinicID: clinicA, SystemKey: &cashKey}}, nil
			case clinicB:
				return []model.PaymentMethodMaster{{ID: clinicBCashID, ClinicID: clinicB, SystemKey: &cashKey}}, nil
			default:
				return nil, nil
			}
		},
	}

	newSvc := func(saved *bool) AccountingService {
		repo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
				return &model.Billing{ID: id, ClinicID: clinicID}, nil
			},
			updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
				return &model.Billing{ID: id, ClinicID: clinicID}, nil
			},
			savePaymentSplitsFn: func(_ context.Context, _ []model.PaymentSplit) error {
				*saved = true
				return nil
			},
		}
		return NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, payMethodRepo)
	}

	t.Run("rejects clinic B's payment_method_id on clinic A's billing update and does not persist", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		foreign := clinicBCashID // clinic B の cash master id（実在するが clinic A のものではない）
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      clinicA,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, PaymentMethodID: &foreign, Amount: 5000, ReceivedAmount: 5000},
			},
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, out)
		assert.False(t, saved, "billing must NOT be persisted with another clinic's payment_method_id")
	})

	t.Run("accepts clinic A's own payment_method_id (no false-reject)", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		owned := clinicACashID
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      clinicA,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, PaymentMethodID: &owned, Amount: 5000, ReceivedAmount: 5000},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, saved)
	})
}

func rejectMerchandiseItemRepo(ownedID uint64) repository.MerchandiseItemRepository {
	return &mockMerchandiseItemRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("merchandise_item", "foreign")
		}
		return &model.MerchandiseItem{ID: id}, nil
	}}
}

// ── campaign (target-item FK): TargetItemIDs → campaign_target_items.merchandise_item_id ──

func TestCampaignService_Create_RejectsCrossClinicTargetItemFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedItemID = uint64(10)
	const foreignItemID = uint64(999)

	newSvc := func(created *bool) CampaignService {
		repo := &mockCampaignRepository{
			createFn: func(_ context.Context, m *model.Campaign) (*model.Campaign, error) {
				*created = true
				m.ID = 1
				return m, nil
			},
		}
		return NewCampaignService(repo, rejectMerchandiseItemRepo(ownedItemID))
	}

	baseInput := func(itemID uint64) *CreateCampaignInput {
		return &CreateCampaignInput{
			Name:          "Autumn Sale",
			StartDate:     time.Now(),
			EndDate:       time.Now().Add(24 * time.Hour),
			DiscountType:  model.CampaignDiscountTypeRate,
			DiscountValue: 10.0,
			TargetItemIDs: []uint64{itemID},
		}
	}

	t.Run("rejects cross-clinic target_item_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, baseInput(foreignItemID))
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "campaign must NOT be persisted referencing another clinic's merchandise item")
	})

	t.Run("accepts same-clinic target_item_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, baseInput(ownedItemID))
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCampaignService_Update_RejectsCrossClinicTargetItemFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedItemID = uint64(10)
	const foreignItemID = uint64(999)

	newSvc := func(replaced *bool) CampaignService {
		current := &model.Campaign{ID: 100, ClinicID: clinicID}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
			replaceTargetsFn: func(_ context.Context, _ uint64, _ []model.ItemCategory, _ []uint64) error {
				*replaced = true
				return nil
			},
		}
		return NewCampaignService(repo, rejectMerchandiseItemRepo(ownedItemID))
	}

	t.Run("rejects cross-clinic target_item_id on update and does not persist", func(t *testing.T) {
		replaced := false
		svc := newSvc(&replaced)
		ids := []uint64{foreignItemID}
		out, err := svc.Update(context.Background(), clinicID, 100, &UpdateCampaignInput{TargetItemIDs: &ids})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, replaced, "campaign targets must NOT be replaced referencing another clinic's merchandise item")
	})

	t.Run("accepts same-clinic target_item_id (no false-reject)", func(t *testing.T) {
		replaced := false
		svc := newSvc(&replaced)
		ids := []uint64{ownedItemID}
		out, err := svc.Update(context.Background(), clinicID, 100, &UpdateCampaignInput{TargetItemIDs: &ids})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, replaced)
	})
}

// ── X-14 batch: self-ref ParentID ownership guard ──
// (checkupType/consultation/examType/procedure/vaccine)
//
// Each of these five master-data services carries a self-referencing ParentID (a
// sub-category pointing at its own parent row in the same table). Prior to this batch,
// Create/Update persisted a request-supplied ParentID without verifying it belongs to
// the caller's clinic. Each service has a single repo dependency (self-ref), so the
// mock below wires findByIDFn (used both for the parent-ownership guard and, on Update,
// the pre-existing self-entity existence check) alongside createFn/updateFieldsFn.

func rejectChiefComplaintTypeRepo(ownedID uint64) repository.ChiefComplaintTypeRepository {
	return &mockChiefComplaintTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ChiefComplaintType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("chief_complaint_type", "foreign")
		}
		return &model.ChiefComplaintType{ID: id}, nil
	}}
}

func TestOwnerService_CreateWithPets_RejectsCrossClinicInsuranceID(t *testing.T) {
	const clinicID = uint64(1)
	const ownedInsuranceID = uint64(20)
	const foreignInsuranceID = uint64(888)

	newSvc := func(created *bool, insuranceRepo repository.InsuranceRepository) OwnerService {
		repo := &mockOwnerRepository{
			createWithPetsFn: func(_ context.Context, owner *model.Owner, _ []model.Pet) error {
				*created = true
				owner.ID = 1
				return nil
			},
		}
		return NewOwnerService(repo, insuranceRepo, &mockLstepTagSyncService{}, nil)
	}

	t.Run("rejects cross-clinic nested insurance_id and does not persist owner or pets", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectInsuranceRepo(ownedInsuranceID))
		foreign := foreignInsuranceID
		input := &CreateOwnerInput{
			OwnerName: "テスト 太郎",
			Pets: []CreatePetForOwnerInput{
				{Name: "ポチ", AnimalSpeciesID: 1, InsuranceID: &foreign},
			},
		}
		out, err := svc.CreateWithPets(context.Background(), clinicID, input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "owner/pets must NOT be persisted with a nested cross-clinic insurance_id")
	})

	t.Run("accepts same-clinic nested insurance_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectInsuranceRepo(ownedInsuranceID))
		owned := ownedInsuranceID
		input := &CreateOwnerInput{
			OwnerName: "テスト 太郎",
			Pets: []CreatePetForOwnerInput{
				{Name: "ポチ", AnimalSpeciesID: 1, InsuranceID: &owned},
			},
		}
		out, err := svc.CreateWithPets(context.Background(), clinicID, input)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestPetService_Create_RejectsCrossClinicInsuranceID(t *testing.T) {
	const clinicID = uint64(1)
	const ownedInsuranceID = uint64(20)
	const foreignInsuranceID = uint64(888)

	insuranceRepoFor := func(ownedID uint64) *mockInsuranceRepository {
		return &mockInsuranceRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Insurance, error) {
				if id != ownedID {
					return nil, apperrors.WrapNotFound("insurance", "foreign")
				}
				return &model.Insurance{ID: id}, nil
			},
		}
	}

	newSvc := func(created *bool) PetService {
		repo := &mockPetRepository{
			createFn: func(_ context.Context, pet *model.Pet) error {
				*created = true
				pet.ID = 1
				return nil
			},
		}
		return newPetSvc(repo, defaultOwnerRepo(), insuranceRepoFor(ownedInsuranceID), defaultMedicalRecordRepo())
	}

	t.Run("rejects cross-clinic insurance_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignInsuranceID
		out, err := svc.Create(context.Background(), clinicID, &CreatePetInput{Name: "ポチ", OwnerID: 5, InsuranceID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "pet must NOT be persisted referencing another clinic's insurance")
	})

	t.Run("accepts same-clinic insurance_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedInsuranceID
		out, err := svc.Create(context.Background(), clinicID, &CreatePetInput{Name: "ポチ", OwnerID: 5, InsuranceID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestPetService_Update_RejectsCrossClinicInsuranceID(t *testing.T) {
	const clinicID = uint64(1)
	const petID = uint64(1)
	const ownedInsuranceID = uint64(20)
	const foreignInsuranceID = uint64(888)

	insuranceRepoFor := func(ownedID uint64) *mockInsuranceRepository {
		return &mockInsuranceRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Insurance, error) {
				if id != ownedID {
					return nil, apperrors.WrapNotFound("insurance", "foreign")
				}
				return &model.Insurance{ID: id}, nil
			},
		}
	}

	newSvc := func(updated *bool) PetService {
		repo := &mockPetRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				*updated = true
				return nil
			},
		}
		return newPetSvc(repo, defaultOwnerRepo(), insuranceRepoFor(ownedInsuranceID), defaultMedicalRecordRepo())
	}

	t.Run("rejects cross-clinic insurance_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := ptrUint64(foreignInsuranceID)
		out, err := svc.Update(context.Background(), clinicID, petID, &UpdatePetInput{InsuranceID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "pet must NOT be updated to reference another clinic's insurance")
	})

	t.Run("accepts same-clinic insurance_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ptrUint64(ownedInsuranceID)
		out, err := svc.Update(context.Background(), clinicID, petID, &UpdatePetInput{InsuranceID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})

	t.Run("NULL clear (&nil) is not subject to the ownership guard", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		var nilInsurance *uint64
		out, err := svc.Update(context.Background(), clinicID, petID, &UpdatePetInput{InsuranceID: &nilInsurance})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── reservationAdminService.Create / reservationService.Create+Update /
//    reservationTypeService.Create+Update (X-14/U6b): ReservationTypeID / GroupID ──
//
// U6a closed the LINE-reservation path (reservationValidators/liffService). U6b closes
// the remaining reservation cluster: admin manual booking, electronic-karte booking
// (including the "shortcut" routes that skip capacity checks), and the reservation-type
// master itself (GroupID — ParentID was already guarded).

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

func TestReservationTypeService_Create_RejectsCrossClinicGroupID(t *testing.T) {
	const clinicID = uint64(1)
	const ownedGroupID = uint64(20)
	const foreignGroupID = uint64(999)

	groupRepoFor := func() repository.ReservationTypeGroupRepository {
		return &mockReservationTypeGroupRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationTypeGroup, error) {
				if gotClinicID != clinicID || id != ownedGroupID {
					return nil, apperrors.WrapNotFound("reservation_type_group", "foreign")
				}
				return &model.ReservationTypeGroup{ID: id, ClinicID: clinicID}, nil
			},
		}
	}

	newSvc := func(created *bool) ReservationTypeService {
		repo := &mockReservationTypeRepository{
			createFn: func(_ context.Context, _ *model.ReservationType) error {
				*created = true
				return nil
			},
		}
		return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, groupRepoFor())
	}

	t.Run("rejects cross-clinic group_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignGroupID
		out, err := svc.Create(context.Background(), clinicID, &CreateReservationTypeInput{Name: "診察", GroupID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "reservation type must NOT be persisted referencing another clinic's group")
	})

	t.Run("accepts same-clinic group_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedGroupID
		out, err := svc.Create(context.Background(), clinicID, &CreateReservationTypeInput{Name: "診察", GroupID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestReservationTypeService_Update_RejectsCrossClinicGroupID(t *testing.T) {
	const clinicID = uint64(1)
	const reservationTypeID = uint64(5)
	const ownedGroupID = uint64(20)
	const foreignGroupID = uint64(999)

	groupRepoFor := func() repository.ReservationTypeGroupRepository {
		return &mockReservationTypeGroupRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationTypeGroup, error) {
				if gotClinicID != clinicID || id != ownedGroupID {
					return nil, apperrors.WrapNotFound("reservation_type_group", "foreign")
				}
				return &model.ReservationTypeGroup{ID: id, ClinicID: clinicID}, nil
			},
		}
	}

	newSvc := func(updated *bool) ReservationTypeService {
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
			},
			updateFn: func(_ context.Context, gotClinicID, id uint64, _ map[string]any) (*model.ReservationType, error) {
				*updated = true
				return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
			},
		}
		return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, groupRepoFor())
	}

	t.Run("rejects cross-clinic group_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignGroupID
		out, err := svc.Update(context.Background(), clinicID, reservationTypeID, &UpdateReservationTypeInput{GroupID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "reservation type must NOT be updated to reference another clinic's group")
	})

	t.Run("accepts same-clinic group_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedGroupID
		out, err := svc.Update(context.Background(), clinicID, reservationTypeID, &UpdateReservationTypeInput{GroupID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// TestReservationTypeService_Update_RejectsCrossClinicParentID は ParentID の
// 所有権ガード(validateReservationTypeParent、既存実装)にクロステナント isolation の
// 実証テストを追加する(X-14: 「実装済みだが証拠不足」パターン。GroupID と対称)。
func TestReservationTypeService_Update_RejectsCrossClinicParentID(t *testing.T) {
	const clinicID = uint64(1)
	const reservationTypeID = uint64(5)
	const ownedParentID = uint64(30)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) ReservationTypeService {
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationType, error) {
				if id == reservationTypeID {
					return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
				}
				if gotClinicID != clinicID || id != ownedParentID {
					return nil, apperrors.WrapNotFound("reservation_type", "foreign")
				}
				return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
			},
			updateFn: func(_ context.Context, gotClinicID, id uint64, _ map[string]any) (*model.ReservationType, error) {
				*updated = true
				return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
			},
		}
		return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, reservationTypeID, &UpdateReservationTypeInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "reservation type must NOT be updated to reference another clinic's parent")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, reservationTypeID, &UpdateReservationTypeInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── staffService (X-14 batch U7): OccupationID ──
//
// staffService.Create/CreateWithAccount/Update persisted a request-derived OccupationID
// without verifying it belongs to the caller's clinic. Guard: occupationRepo.FindByID
// (ctx, clinicID, OccupationID) before persist, mirroring medicineService's
// validateInventoryOwnership (X-14 batch U2). occupationRepo is now a mandatory
// NewStaffService dependency (see staff_service_core.go validateOccupationOwnership).

func TestStaffService_Create_RejectsCrossClinicOccupationID(t *testing.T) {
	const clinicID = uint64(1)
	const ownedOccupationID = uint64(10)
	const foreignOccupationID = uint64(999)

	newSvc := func(created *bool, occupationRepo repository.OccupationRepository) StaffService {
		repo := &mockStaffRepository{
			createFn: func(_ context.Context, staff *model.Staff) error {
				*created = true
				staff.ID = 1
				return nil
			},
		}
		return NewStaffService(repo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, occupationRepo, noopTransactor{})
	}

	t.Run("rejects cross-clinic occupation_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectOccupationRepo(ownedOccupationID))
		foreign := foreignOccupationID
		out, err := svc.Create(context.Background(), &CreateStaffInput{ClinicID: clinicID, Name: "x", OccupationID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "staff must NOT be persisted referencing another clinic's occupation")
	})

	t.Run("accepts same-clinic occupation_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectOccupationRepo(ownedOccupationID))
		owned := ownedOccupationID
		out, err := svc.Create(context.Background(), &CreateStaffInput{ClinicID: clinicID, Name: "x", OccupationID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestStaffService_CreateWithAccount_RejectsCrossClinicOccupationID(t *testing.T) {
	const clinicID = uint64(1)
	const ownedOccupationID = uint64(10)
	const foreignOccupationID = uint64(999)

	newSvc := func(created *bool, occupationRepo repository.OccupationRepository) StaffService {
		repo := &mockStaffRepository{
			createFn: func(_ context.Context, staff *model.Staff) error {
				*created = true
				staff.ID = 1
				return nil
			},
		}
		accountRepo := &mockAccountForStaff{
			findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
				return nil, apperrors.WrapNotFound("account", "email")
			},
			createFn: func(_ context.Context, account *model.Account) error {
				account.ID = 1
				return nil
			},
		}
		return NewStaffService(repo, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, occupationRepo, noopTransactor{})
	}

	t.Run("rejects cross-clinic occupation_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectOccupationRepo(ownedOccupationID))
		foreign := foreignOccupationID
		out, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
			ClinicID: clinicID, Name: "x", Email: "u@example.com", Password: "Password1!", OccupationID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "staff must NOT be persisted referencing another clinic's occupation")
	})

	t.Run("accepts same-clinic occupation_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectOccupationRepo(ownedOccupationID))
		owned := ownedOccupationID
		out, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
			ClinicID: clinicID, Name: "x", Email: "u@example.com", Password: "Password1!", OccupationID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestStaffService_Update_RejectsCrossClinicOccupationID(t *testing.T) {
	const clinicID = uint64(1)
	const staffID = uint64(1)
	const ownedOccupationID = uint64(10)
	const foreignOccupationID = uint64(999)

	newSvc := func(updated *bool, occupationRepo repository.OccupationRepository) StaffService {
		repo := &mockStaffRepository{
			findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: clinicID}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				*updated = true
				return nil
			},
		}
		return NewStaffService(repo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, occupationRepo, noopTransactor{})
	}

	t.Run("rejects cross-clinic occupation_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated, rejectOccupationRepo(ownedOccupationID))
		foreign := foreignOccupationID
		out, err := svc.Update(context.Background(), clinicID, staffID, &UpdateStaffInput{OccupationID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "staff must NOT be updated to reference another clinic's occupation")
	})

	t.Run("accepts same-clinic occupation_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated, rejectOccupationRepo(ownedOccupationID))
		owned := ownedOccupationID
		out, err := svc.Update(context.Background(), clinicID, staffID, &UpdateStaffInput{OccupationID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── trimmingService CourseID/OptionIDs (X-14c): Create/Update now guard via
// trimmingCourseRepo/trimmingOptionRepo.FindByID(ctx, clinicID, id) before persist
// (DI added — reservation_validators.go:116-127 と同型の 2 repo ガード). ──

func TestTrimmingService_Create_RejectsCrossClinicCourseFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedCourseID = uint64(10)
	const foreignCourseID = uint64(999)

	newSvc := func(created *bool, courseRepo repository.TrimmingCourseRepository) TrimmingService {
		reserv := &mockTrimmingReservationRepository{
			createFn: func(_ context.Context, a *model.Reservation) error {
				*created = true
				a.ID = 1
				return nil
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, ClinicID: clinicID}, nil
			},
		}
		return NewTrimmingService(reserv, &mockTrimmingReservationTypeRepository{}, nil, nil, nil,
			&mockTrimmingDetailRepository{}, courseRepo, okTrimmingOptionRepo(), &mockTransactor{})
	}

	t.Run("rejects cross-clinic course_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID))
		foreign := foreignCourseID
		out, err := svc.Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1,
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(time.Hour),
			CourseID:          &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "trimming appointment must NOT be persisted when referencing another clinic's course")
	})

	t.Run("accepts same-clinic course_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID))
		owned := ownedCourseID
		out, err := svc.Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1,
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(time.Hour),
			CourseID:          &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestTrimmingService_Create_RejectsCrossClinicOptionFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedOptionID = uint64(20)
	const foreignOptionID = uint64(998)

	newSvc := func(created *bool, optionRepo repository.TrimmingOptionRepository) TrimmingService {
		reserv := &mockTrimmingReservationRepository{
			createFn: func(_ context.Context, a *model.Reservation) error {
				*created = true
				a.ID = 1
				return nil
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, ClinicID: clinicID}, nil
			},
		}
		return NewTrimmingService(reserv, &mockTrimmingReservationTypeRepository{}, nil, nil, nil,
			&mockTrimmingDetailRepository{}, okTrimmingCourseRepo(), optionRepo, &mockTransactor{})
	}

	t.Run("rejects cross-clinic option_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingOptionRepo(ownedOptionID))
		out, err := svc.Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1,
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(time.Hour),
			OptionIDs:         []uint64{foreignOptionID},
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "trimming appointment must NOT be persisted when referencing another clinic's option")
	})

	t.Run("accepts same-clinic option_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingOptionRepo(ownedOptionID))
		out, err := svc.Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1,
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(time.Hour),
			OptionIDs:         []uint64{ownedOptionID},
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestTrimmingService_Update_RejectsCrossClinicCourseFK(t *testing.T) {
	const clinicID = uint64(1)
	const appointmentID = uint64(1)
	const ownedCourseID = uint64(10)
	const foreignCourseID = uint64(999)

	newSvc := func(updated *bool, courseRepo repository.TrimmingCourseRepository) TrimmingService {
		// UpdateTrimmingInput{CourseID: ...} だけでは appointments 側の apptFields は空のまま
		// (course_id は appointment_trimming_detail 側のフィールド) なので、永続化の観測は
		// s.trimmingDetail.Update 呼び出しで行う。
		reserv := &mockTrimmingReservationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, ClinicID: clinicID}, nil
			},
		}
		detail := &mockTrimmingDetailRepository{
			updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
				*updated = true
				return nil
			},
		}
		return NewTrimmingService(reserv, &mockTrimmingReservationTypeRepository{}, nil, nil, nil,
			detail, courseRepo, okTrimmingOptionRepo(), &mockTransactor{})
	}

	t.Run("rejects cross-clinic course_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated, rejectTrimmingCourseRepo(ownedCourseID))
		foreign := foreignCourseID
		out, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{CourseID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "trimming appointment must NOT be updated to reference another clinic's course")
	})

	t.Run("accepts same-clinic course_id on update (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated, rejectTrimmingCourseRepo(ownedCourseID))
		owned := ownedCourseID
		out, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{CourseID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

func TestTrimmingService_Update_RejectsCrossClinicOptionFK(t *testing.T) {
	const clinicID = uint64(1)
	const appointmentID = uint64(1)
	const ownedOptionID = uint64(20)
	const foreignOptionID = uint64(998)

	newSvc := func(updated *bool, optionRepo repository.TrimmingOptionRepository) TrimmingService {
		reserv := &mockTrimmingReservationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, ClinicID: clinicID}, nil
			},
		}
		detail := &mockTrimmingDetailRepository{
			setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
				*updated = true
				return nil
			},
		}
		return NewTrimmingService(reserv, &mockTrimmingReservationTypeRepository{}, nil, nil, nil,
			detail, okTrimmingCourseRepo(), optionRepo, &mockTransactor{})
	}

	t.Run("rejects cross-clinic option_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated, rejectTrimmingOptionRepo(ownedOptionID))
		foreignIDs := []uint64{foreignOptionID}
		out, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{OptionIDs: &foreignIDs})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "trimming detail options must NOT be updated to reference another clinic's option")
	})

	t.Run("accepts same-clinic option_id on update (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated, rejectTrimmingOptionRepo(ownedOptionID))
		ownedIDs := []uint64{ownedOptionID}
		out, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{OptionIDs: &ownedIDs})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}
