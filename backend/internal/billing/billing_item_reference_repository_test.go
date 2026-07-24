package billing

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type billingItemReferenceFixture struct {
	db              *gorm.DB
	repo            BillingItemRepository
	clinicID        uint64
	owner           *model.Owner
	pet             *model.Pet
	appointment     *model.Reservation
	medicalRecord   *model.MedicalRecord
	treatment       *model.Treatment
	billing         *model.Billing
	merchandiseItem *model.MerchandiseItem
	course          *model.TrimmingCourse
	option          *model.TrimmingOption
}

type createTrackingBillingItemRepository struct {
	BillingItemRepository
	createCalls int
}

type concurrentBillingItemRepository struct {
	BillingItemRepository
	validated       atomic.Int32
	secondValidated chan struct{}
}

func (r *createTrackingBillingItemRepository) Create(ctx context.Context, item *model.BillingItem) error {
	r.createCalls++
	return r.BillingItemRepository.Create(ctx, item)
}

func (r *concurrentBillingItemRepository) ValidateCreateReferences(
	ctx context.Context,
	clinicID, billingID uint64,
	merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64,
) error {
	if err := r.BillingItemRepository.ValidateCreateReferences(
		ctx,
		clinicID,
		billingID,
		merchandiseItemID,
		treatmentID,
		appointmentID,
		trimmingCourseID,
		trimmingOptionID,
	); err != nil {
		return err
	}

	switch r.validated.Add(1) {
	case 1:
		// With a shared parent lock, both validations can finish and the two
		// later billing-total updates form a lock-upgrade deadlock. With the
		// required UPDATE lock, the second validation waits for the first
		// transaction, so release the first after a short observation window.
		select {
		case <-r.secondValidated:
		case <-time.After(250 * time.Millisecond):
		}
	case 2:
		close(r.secondValidated)
	}
	return nil
}

func setupBillingItemReferenceFixture(t *testing.T) billingItemReferenceFixture {
	t.Helper()

	const clinicID = uint64(1)
	db := setupBillingItemTrimmingTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.MerchandiseItem{}, &model.Treatment{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE merchandise_items CASCADE").Error)

	owner := testdb.MakeTestOwner(t, db, clinicID, "billing-item-reference-owner")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "billing-item-reference-pet")
	reservationType := makeTrimmingReservationType(t, db, clinicID)
	appointment := makeTrimmingAppointment(t, db, clinicID, pet.ID, reservationType.ID, model.ReservationStatusAccounting)
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("id = ?", appointment.ID).
		Update("owner_id", owner.ID).Error)
	appointment.OwnerID = &owner.ID

	medicalRecord := &model.MedicalRecord{
		ClinicID:      clinicID,
		RecordNo:      "billing-item-reference-mr",
		Date:          time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		OwnerID:       &owner.ID,
		PetID:         &pet.ID,
		AppointmentID: &appointment.ID,
	}
	require.NoError(t, db.Create(medicalRecord).Error)

	treatment := &model.Treatment{
		MedicalRecordID: medicalRecord.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "valid treatment",
	}
	require.NoError(t, db.Create(treatment).Error)

	billing := &model.Billing{
		ClinicID:        clinicID,
		MedicalRecordID: &medicalRecord.ID,
		OwnerID:         &owner.ID,
		PetID:           &pet.ID,
		Status:          model.BillingStatusWaiting,
		ScheduledDate:   time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.Create(billing).Error)

	merchandiseItem := &model.MerchandiseItem{
		ClinicID: clinicID,
		Name:     "valid merchandise",
		Category: model.ItemCategoryGoods,
		IsActive: true,
	}
	require.NoError(t, db.Create(merchandiseItem).Error)

	course := makeTrimmingCourse(t, db, clinicID, "valid course", priceOf(1000))
	option := makeTrimmingOption(t, db, clinicID, "valid option", priceOf(500))
	attachTrimmingCourse(t, db, clinicID, appointment.ID, course.ID)
	attachTrimmingOption(t, db, appointment.ID, option.ID, 0)

	return billingItemReferenceFixture{
		db:              db,
		repo:            NewBillingItemRepository(db),
		clinicID:        clinicID,
		owner:           owner,
		pet:             pet,
		appointment:     appointment,
		medicalRecord:   medicalRecord,
		treatment:       treatment,
		billing:         billing,
		merchandiseItem: merchandiseItem,
		course:          course,
		option:          option,
	}
}

func (f billingItemReferenceFixture) validate(
	t *testing.T,
	merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64,
) error {
	t.Helper()
	return testNewTransactor(f.db).WithTx(context.Background(), func(txCtx context.Context) error {
		return f.repo.ValidateCreateReferences(
			txCtx,
			f.clinicID,
			f.billing.ID,
			merchandiseItemID,
			treatmentID,
			appointmentID,
			trimmingCourseID,
			trimmingOptionID,
		)
	})
}

func TestBillingItemRepository_ValidateCreateReferences(t *testing.T) {
	t.Run("valid same-clinic active and related graph is retained", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)

		err := f.validate(
			t,
			&f.merchandiseItem.ID,
			&f.treatment.ID,
			&f.appointment.ID,
			&f.course.ID,
			&f.option.ID,
		)

		require.NoError(t, err)
	})

	t.Run("cross-clinic treatment and merchandise references fail closed", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		const otherClinicID = uint64(2)
		otherOwner := testdb.MakeTestOwner(t, f.db, otherClinicID, "foreign owner")
		otherPet := makeSpeciesAndPet(t, f.db, otherClinicID, otherOwner.ID, "foreign pet")
		otherMedicalRecord := &model.MedicalRecord{
			ClinicID: otherClinicID,
			RecordNo: "foreign-mr",
			Date:     time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			OwnerID:  &otherOwner.ID,
			PetID:    &otherPet.ID,
		}
		require.NoError(t, f.db.Create(otherMedicalRecord).Error)
		otherTreatment := &model.Treatment{
			MedicalRecordID: otherMedicalRecord.ID,
			ItemType:        model.TreatmentItemTypeOther,
			Content:         "foreign treatment",
		}
		require.NoError(t, f.db.Create(otherTreatment).Error)
		otherMerchandise := &model.MerchandiseItem{
			ClinicID: otherClinicID,
			Name:     "foreign merchandise",
			Category: model.ItemCategoryGoods,
			IsActive: true,
		}
		require.NoError(t, f.db.Create(otherMerchandise).Error)

		err := f.validate(t, &otherMerchandise.ID, &otherTreatment.ID, nil, nil, nil)

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "cross-clinic IDs must be indistinguishable from absent IDs: %v", err)
	})

	t.Run("soft-deleted treatment and inactive master references fail closed", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		require.NoError(t, f.db.Delete(&model.Treatment{}, f.treatment.ID).Error)
		require.NoError(t, f.db.Model(&model.MerchandiseItem{}).
			Where("id = ?", f.merchandiseItem.ID).
			Update("is_active", false).Error)

		err := f.validate(t, nil, &f.treatment.ID, nil, nil, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		err = f.validate(t, &f.merchandiseItem.ID, nil, nil, nil, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("inactive trimming masters remain valid when already attached to the appointment", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		require.NoError(t, f.db.Model(&model.TrimmingCourse{}).
			Where("id = ?", f.course.ID).
			Update("is_active", false).Error)
		require.NoError(t, f.db.Model(&model.TrimmingOption{}).
			Where("id = ?", f.option.ID).
			Update("is_active", false).Error)

		err := f.validate(
			t,
			nil,
			nil,
			&f.appointment.ID,
			&f.course.ID,
			&f.option.ID,
		)

		require.NoError(t, err)
	})

	t.Run("same-clinic treatment ID reused from another medical record is rejected", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		otherRecord := &model.MedicalRecord{
			ClinicID: f.clinicID,
			RecordNo: "same-clinic-other-mr",
			Date:     time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			OwnerID:  &f.owner.ID,
			PetID:    &f.pet.ID,
		}
		require.NoError(t, f.db.Create(otherRecord).Error)
		otherTreatment := &model.Treatment{
			MedicalRecordID: otherRecord.ID,
			ItemType:        model.TreatmentItemTypeOther,
			Content:         "wrong visit treatment",
		}
		require.NoError(t, f.db.Create(otherTreatment).Error)

		err := f.validate(t, nil, &otherTreatment.ID, nil, nil, nil)

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "same-tenant relation mismatch must be rejected: %v", err)
	})

	t.Run("appointment from another pet and masters attached to another appointment are rejected", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		otherPet := makeSpeciesAndPet(t, f.db, f.clinicID, f.owner.ID, "same-clinic-other-pet")
		reservationType := makeTrimmingReservationType(t, f.db, f.clinicID)
		otherAppointment := makeTrimmingAppointment(t, f.db, f.clinicID, otherPet.ID, reservationType.ID, model.ReservationStatusAccounting)
		require.NoError(t, f.db.Model(&model.Reservation{}).
			Where("id = ?", otherAppointment.ID).
			Update("owner_id", f.owner.ID).Error)
		wrongCourse := makeTrimmingCourse(t, f.db, f.clinicID, "course for other appointment", priceOf(2000))
		wrongOption := makeTrimmingOption(t, f.db, f.clinicID, "option for other appointment", priceOf(700))
		attachTrimmingCourse(t, f.db, f.clinicID, otherAppointment.ID, wrongCourse.ID)
		attachTrimmingOption(t, f.db, otherAppointment.ID, wrongOption.ID, 0)

		err := f.validate(t, nil, nil, &otherAppointment.ID, &wrongCourse.ID, &wrongOption.ID)

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "same-clinic IDs from an unrelated pet/appointment must be rejected: %v", err)

		err = f.validate(t, nil, nil, &f.appointment.ID, &wrongCourse.ID, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "course must be attached to the referenced appointment: %v", err)

		err = f.validate(t, nil, nil, &f.appointment.ID, nil, &wrongOption.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "option must be attached to the referenced appointment: %v", err)
	})

	t.Run("trimming master without appointment is rejected", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)

		err := f.validate(t, nil, nil, nil, &f.course.ID, nil)

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

func newBillingItemReferenceService(
	f billingItemReferenceFixture,
	repo BillingItemRepository,
) BillingItemService {
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID || billingID != f.billing.ID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		return f.billing, nil
	}
	return NewBillingItemServiceWithCampaign(
		repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
	)
}

func billingItemReferenceCreateInput(f billingItemReferenceFixture) *CreateBillingItemInput {
	return &CreateBillingItemInput{
		ClinicID:  f.clinicID,
		BillingID: f.billing.ID,
		Category:  string(model.ItemCategoryOther),
		Name:      "reference validation item",
		UnitPrice: 1000,
		Quantity:  1,
	}
}

func countBillingItems(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&model.BillingItem{}).Count(&count).Error)
	return count
}

func TestBillingItemService_CreateItem_RuntimeReferenceIsolation(t *testing.T) {
	t.Run("all valid same-clinic references are persisted", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		trackedRepo := &createTrackingBillingItemRepository{BillingItemRepository: f.repo}
		svc := newBillingItemReferenceService(f, trackedRepo)
		input := billingItemReferenceCreateInput(f)
		input.MerchandiseItemID = &f.merchandiseItem.ID
		input.TreatmentID = &f.treatment.ID
		input.AppointmentID = &f.appointment.ID
		input.TrimmingCourseID = &f.course.ID
		input.TrimmingOptionID = &f.option.ID
		before := countBillingItems(t, f.db)

		item, err := svc.CreateItem(context.Background(), input)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, 1, trackedRepo.createCalls)
		assert.Equal(t, before+1, countBillingItems(t, f.db))
	})

	t.Run("already-attached inactive trimming references remain billable", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		require.NoError(t, f.db.Model(&model.TrimmingCourse{}).
			Where("id = ?", f.course.ID).
			Update("is_active", false).Error)
		require.NoError(t, f.db.Model(&model.TrimmingOption{}).
			Where("id = ?", f.option.ID).
			Update("is_active", false).Error)
		trackedRepo := &createTrackingBillingItemRepository{BillingItemRepository: f.repo}
		svc := newBillingItemReferenceService(f, trackedRepo)
		input := billingItemReferenceCreateInput(f)
		input.AppointmentID = &f.appointment.ID
		input.TrimmingCourseID = &f.course.ID
		input.TrimmingOptionID = &f.option.ID
		before := countBillingItems(t, f.db)

		item, err := svc.CreateItem(context.Background(), input)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, 1, trackedRepo.createCalls)
		assert.Equal(t, before+1, countBillingItems(t, f.db))
	})

	tests := []struct {
		name   string
		mutate func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput)
	}{
		{
			name: "cross-clinic merchandise_item_id is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				foreign := &model.MerchandiseItem{
					ClinicID: 2,
					Name:     "foreign merchandise",
					Category: model.ItemCategoryGoods,
					IsActive: true,
				}
				require.NoError(t, f.db.Create(foreign).Error)
				input.MerchandiseItemID = &foreign.ID
			},
		},
		{
			name: "cross-clinic treatment_id is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				foreignRecord := &model.MedicalRecord{
					ClinicID: 2,
					RecordNo: "foreign-treatment-record",
					Date:     time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
				}
				require.NoError(t, f.db.Create(foreignRecord).Error)
				foreign := &model.Treatment{
					MedicalRecordID: foreignRecord.ID,
					ItemType:        model.TreatmentItemTypeOther,
					Content:         "foreign treatment",
				}
				require.NoError(t, f.db.Create(foreign).Error)
				input.TreatmentID = &foreign.ID
			},
		},
		{
			name: "cross-clinic appointment_id is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				foreignOwner := testdb.MakeTestOwner(t, f.db, 2, "foreign appointment owner")
				foreignPet := makeSpeciesAndPet(t, f.db, 2, foreignOwner.ID, "foreign appointment pet")
				foreignType := makeTrimmingReservationType(t, f.db, 2)
				foreign := makeTrimmingAppointment(t, f.db, 2, foreignPet.ID, foreignType.ID, model.ReservationStatusAccounting)
				require.NoError(t, f.db.Model(&model.Reservation{}).
					Where("id = ?", foreign.ID).
					Update("owner_id", foreignOwner.ID).Error)
				input.AppointmentID = &foreign.ID
			},
		},
		{
			name: "cross-clinic trimming_course_id attached to local appointment is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				foreign := makeTrimmingCourse(t, f.db, 2, "foreign course", priceOf(2000))
				require.NoError(t, f.db.Model(&model.AppointmentTrimmingDetail{}).
					Where("appointment_id = ?", f.appointment.ID).
					Update("course_id", foreign.ID).Error)
				input.AppointmentID = &f.appointment.ID
				input.TrimmingCourseID = &foreign.ID
			},
		},
		{
			name: "cross-clinic trimming_option_id attached to local appointment is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				foreign := makeTrimmingOption(t, f.db, 2, "foreign option", priceOf(800))
				attachTrimmingOption(t, f.db, f.appointment.ID, foreign.ID, 10)
				input.AppointmentID = &f.appointment.ID
				input.TrimmingOptionID = &foreign.ID
			},
		},
		{
			name: "soft-deleted billing is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, _ *CreateBillingItemInput) {
				require.NoError(t, f.db.Delete(&model.Billing{}, f.billing.ID).Error)
			},
		},
		{
			name: "soft-deleted merchandise_item_id is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				require.NoError(t, f.db.Delete(&model.MerchandiseItem{}, f.merchandiseItem.ID).Error)
				input.MerchandiseItemID = &f.merchandiseItem.ID
			},
		},
		{
			name: "soft-deleted treatment_id is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				require.NoError(t, f.db.Delete(&model.Treatment{}, f.treatment.ID).Error)
				input.TreatmentID = &f.treatment.ID
			},
		},
		{
			name: "soft-deleted appointment_id is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				require.NoError(t, f.db.Delete(&model.Reservation{}, f.appointment.ID).Error)
				input.AppointmentID = &f.appointment.ID
			},
		},
		{
			name: "soft-deleted trimming_course_id is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				require.NoError(t, f.db.Delete(&model.TrimmingCourse{}, f.course.ID).Error)
				input.AppointmentID = &f.appointment.ID
				input.TrimmingCourseID = &f.course.ID
			},
		},
		{
			name: "soft-deleted trimming_option_id is rejected before Create",
			mutate: func(t *testing.T, f billingItemReferenceFixture, input *CreateBillingItemInput) {
				require.NoError(t, f.db.Delete(&model.TrimmingOption{}, f.option.ID).Error)
				input.AppointmentID = &f.appointment.ID
				input.TrimmingOptionID = &f.option.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupBillingItemReferenceFixture(t)
			trackedRepo := &createTrackingBillingItemRepository{BillingItemRepository: f.repo}
			svc := newBillingItemReferenceService(f, trackedRepo)
			input := billingItemReferenceCreateInput(f)
			tt.mutate(t, f, input)
			before := countBillingItems(t, f.db)

			item, err := svc.CreateItem(context.Background(), input)

			require.Error(t, err)
			assert.Nil(t, item)
			assert.Zero(t, trackedRepo.createCalls, "invalid request-derived FK must stop before repository Create")
			assert.Equal(t, before, countBillingItems(t, f.db), "invalid request-derived FK must not persist a billing item")
		})
	}
}

func TestBillingItemService_CreateItem_SerializesConcurrentSameBillingWrites(
	t *testing.T,
) {
	f := setupBillingItemReferenceFixture(t)
	repo := &concurrentBillingItemRepository{
		BillingItemRepository: f.repo,
		secondValidated:       make(chan struct{}),
	}
	svc := newBillingItemReferenceService(f, repo)
	inputs := []*CreateBillingItemInput{
		billingItemReferenceCreateInput(f),
		billingItemReferenceCreateInput(f),
	}
	inputs[0].Name = "concurrent item one"
	inputs[0].UnitPrice = 1000
	inputs[1].Name = "concurrent item two"
	inputs[1].UnitPrice = 2000

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := make(chan struct{})
	errs := make(chan error, len(inputs))
	for i := range inputs {
		input := inputs[i]
		go func() {
			<-start
			_, err := svc.CreateItem(ctx, input)
			errs <- err
		}()
	}
	close(start)

	for range inputs {
		require.NoError(t, <-errs)
	}

	var items []model.BillingItem
	require.NoError(t, f.db.
		Where("billing_id = ? AND deleted_at IS NULL", f.billing.ID).
		Order("id ASC").
		Find(&items).Error)
	require.Len(t, items, len(inputs))

	var billing model.Billing
	require.NoError(t, f.db.First(&billing, f.billing.ID).Error)
	subtotal, taxTotal, totalAmount := CalculateBillingTotals(items)
	assert.Equal(t, subtotal, billing.Subtotal)
	assert.Equal(t, taxTotal, billing.TaxTotal)
	assert.Equal(t, totalAmount, billing.TotalAmount)
}
