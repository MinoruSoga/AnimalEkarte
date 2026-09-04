package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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

type validationStartBillingItemRepository struct {
	BillingItemRepository
	started chan struct{}
}

func (m *mockBillingItemRepository) LockActiveStaffAssignment(
	_ context.Context,
	_, _ uint64,
) error {
	return nil
}

func (r *createTrackingBillingItemRepository) Create(ctx context.Context, item *model.BillingItem) error {
	r.createCalls++
	return r.BillingItemRepository.Create(ctx, item)
}

func (r *concurrentBillingItemRepository) ValidateCreateReferences(
	ctx context.Context,
	clinicID, billingID uint64,
	merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64,
) (model.ItemCategory, error) {
	category, err := r.BillingItemRepository.ValidateCreateReferences(
		ctx,
		clinicID,
		billingID,
		merchandiseItemID,
		treatmentID,
		appointmentID,
		trimmingCourseID,
		trimmingOptionID,
	)
	if err != nil {
		return "", err
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
	return category, nil
}

func (r *validationStartBillingItemRepository) ValidateCreateReferences(
	ctx context.Context,
	clinicID, billingID uint64,
	merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64,
) (model.ItemCategory, error) {
	close(r.started)
	return r.BillingItemRepository.ValidateCreateReferences(
		ctx,
		clinicID,
		billingID,
		merchandiseItemID,
		treatmentID,
		appointmentID,
		trimmingCourseID,
		trimmingOptionID,
	)
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

type splitAppointmentReferenceFixture struct {
	db                  *gorm.DB
	repo                BillingItemRepository
	clinicID            uint64
	owner               *model.Owner
	pet                 *model.Pet
	examAppointment     *model.Reservation
	trimmingAppointment *model.Reservation
	medicalRecord       *model.MedicalRecord
	treatment           *model.Treatment
	billing             *model.Billing
	course              *model.TrimmingCourse
	option              *model.TrimmingOption
}

// setupSplitAppointmentReferenceFixture builds the S11 graph: exam appointment A
// owns medical_records.appointment_id; trimming appointment B owns course/option.
func setupSplitAppointmentReferenceFixture(t *testing.T, withBilling bool) splitAppointmentReferenceFixture {
	t.Helper()

	const clinicID = uint64(1)
	db := setupBillingItemTrimmingTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.MerchandiseItem{}, &model.Treatment{}))

	owner := testdb.MakeTestOwner(t, db, clinicID, "s11-split-owner")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "s11-split-pet")

	generalType := makeReservationType(t, db, clinicID)
	examAppointment := makeTrimmingAppointment(t, db, clinicID, pet.ID, generalType.ID, model.ReservationStatusAccounting)
	setAppointmentTime(t, db, examAppointment, time.Date(2026, 6, 11, 4, 0, 0, 0, time.UTC))
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("id = ?", examAppointment.ID).
		Update("owner_id", owner.ID).Error)
	examAppointment.OwnerID = &owner.ID

	trimmingType := makeTrimmingReservationType(t, db, clinicID)
	trimmingAppointment := makeTrimmingAppointment(t, db, clinicID, pet.ID, trimmingType.ID, model.ReservationStatusAccounting)
	setAppointmentTime(t, db, trimmingAppointment, time.Date(2026, 6, 11, 6, 0, 0, 0, time.UTC))
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("id = ?", trimmingAppointment.ID).
		Update("owner_id", owner.ID).Error)
	trimmingAppointment.OwnerID = &owner.ID

	medicalRecord := &model.MedicalRecord{
		ClinicID:      clinicID,
		RecordNo:      "s11-split-exam-mr",
		Date:          time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
		OwnerID:       &owner.ID,
		PetID:         &pet.ID,
		AppointmentID: &examAppointment.ID,
	}
	require.NoError(t, db.Create(medicalRecord).Error)

	treatment := &model.Treatment{
		MedicalRecordID: medicalRecord.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "s11 exam treatment",
	}
	require.NoError(t, db.Create(treatment).Error)

	course := makeTrimmingCourse(t, db, clinicID, "s11 split course", priceOf(1000))
	option := makeTrimmingOption(t, db, clinicID, "s11 split option", priceOf(500))
	attachTrimmingCourse(t, db, clinicID, trimmingAppointment.ID, course.ID)
	attachTrimmingOption(t, db, trimmingAppointment.ID, option.ID, 0)

	f := splitAppointmentReferenceFixture{
		db:                  db,
		repo:                NewBillingItemRepository(db),
		clinicID:            clinicID,
		owner:               owner,
		pet:                 pet,
		examAppointment:     examAppointment,
		trimmingAppointment: trimmingAppointment,
		medicalRecord:       medicalRecord,
		treatment:           treatment,
		course:              course,
		option:              option,
	}
	if withBilling {
		billing := &model.Billing{
			ClinicID:        clinicID,
			MedicalRecordID: &medicalRecord.ID,
			OwnerID:         &owner.ID,
			PetID:           &pet.ID,
			Status:          model.BillingStatusWaiting,
			ScheduledDate:   time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, db.Create(billing).Error)
		f.billing = billing
	}
	return f
}

func (f splitAppointmentReferenceFixture) validate(
	t *testing.T,
	merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64,
) (model.ItemCategory, error) {
	t.Helper()
	require.NotNil(t, f.billing)
	var category model.ItemCategory
	err := testNewTransactor(f.db).WithTx(context.Background(), func(txCtx context.Context) error {
		var err error
		category, err = f.repo.ValidateCreateReferences(
			txCtx,
			f.clinicID,
			f.billing.ID,
			merchandiseItemID,
			treatmentID,
			appointmentID,
			trimmingCourseID,
			trimmingOptionID,
		)
		return err
	})
	return category, err
}

func (f billingItemReferenceFixture) validate(
	t *testing.T,
	merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64,
) (model.ItemCategory, error) {
	t.Helper()
	var category model.ItemCategory
	err := testNewTransactor(f.db).WithTx(context.Background(), func(txCtx context.Context) error {
		var err error
		category, err = f.repo.ValidateCreateReferences(
			txCtx,
			f.clinicID,
			f.billing.ID,
			merchandiseItemID,
			treatmentID,
			appointmentID,
			trimmingCourseID,
			trimmingOptionID,
		)
		return err
	})
	return category, err
}

func TestBillingItemRepository_ValidateCreateReferences(t *testing.T) {
	t.Run("valid same-clinic active and related graph is retained", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)

		_, err := f.validate(
			t,
			&f.merchandiseItem.ID,
			&f.treatment.ID,
			&f.appointment.ID,
			&f.course.ID,
			&f.option.ID,
		)

		require.NoError(t, err)
	})

	t.Run("returns category from locked active merchandise reference", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)

		category, err := f.validate(t, &f.merchandiseItem.ID, nil, nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, model.ItemCategoryGoods, category)
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

		_, err := f.validate(t, &otherMerchandise.ID, &otherTreatment.ID, nil, nil, nil)

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "cross-clinic IDs must be indistinguishable from absent IDs: %v", err)
	})

	t.Run("soft-deleted treatment and inactive master references fail closed", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		require.NoError(t, f.db.Delete(&model.Treatment{}, f.treatment.ID).Error)
		require.NoError(t, f.db.Model(&model.MerchandiseItem{}).
			Where("id = ?", f.merchandiseItem.ID).
			Update("is_active", false).Error)

		_, err := f.validate(t, nil, &f.treatment.ID, nil, nil, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		_, err = f.validate(t, &f.merchandiseItem.ID, nil, nil, nil, nil)
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

		_, err := f.validate(
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

		_, err := f.validate(t, nil, &otherTreatment.ID, nil, nil, nil)

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

		_, err := f.validate(t, nil, nil, &otherAppointment.ID, &wrongCourse.ID, &wrongOption.ID)

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "same-clinic IDs from an unrelated pet/appointment must be rejected: %v", err)

		_, err = f.validate(t, nil, nil, &f.appointment.ID, &wrongCourse.ID, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "course must be attached to the referenced appointment: %v", err)

		_, err = f.validate(t, nil, nil, &f.appointment.ID, nil, &wrongOption.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "option must be attached to the referenced appointment: %v", err)
	})

	t.Run("trimming master without resolvable appointment is rejected", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		// Course exists in clinic but is not attached to any accounting-status
		// appointment for the billing pet — BUG-506 must still fail closed.
		orphanCourse := makeTrimmingCourse(t, f.db, f.clinicID, "orphan course", priceOf(900))

		_, err := f.validate(t, nil, nil, nil, &orphanCourse.ID, nil)

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("BUG-506: attached trimming master without appointment_id resolves unique appointment", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)

		_, err := f.validate(t, nil, nil, nil, &f.course.ID, nil)
		require.NoError(t, err)

		_, err = f.validate(t, nil, nil, nil, nil, &f.option.ID)
		require.NoError(t, err)
	})

	t.Run("S11 split appointments: trimming items on appointment B succeed with exam medical_record appointment A", func(t *testing.T) {
		f := setupSplitAppointmentReferenceFixture(t, true)

		_, err := f.validate(t, nil, nil, &f.trimmingAppointment.ID, &f.course.ID, &f.option.ID)
		require.NoError(t, err)

		_, err = f.validate(t, nil, nil, &f.trimmingAppointment.ID, &f.course.ID, nil)
		require.NoError(t, err)

		_, err = f.validate(t, nil, nil, &f.trimmingAppointment.ID, nil, &f.option.ID)
		require.NoError(t, err)

		_, err = f.validate(t, nil, &f.treatment.ID, &f.examAppointment.ID, nil, nil)
		require.NoError(t, err)

		_, err = f.validate(t, nil, &f.treatment.ID, nil, nil, nil)
		require.NoError(t, err)
	})

	t.Run("S11 split appointments: treatment from exam MR with trimming appointment B remains InvalidInput", func(t *testing.T) {
		f := setupSplitAppointmentReferenceFixture(t, true)

		_, err := f.validate(t, nil, &f.treatment.ID, &f.trimmingAppointment.ID, nil, nil)

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "treatment must stay bound to the exam appointment: %v", err)
		assert.Contains(t, err.Error(), "参照先の組み合わせが正しくありません")

		_, err = f.validate(t, nil, nil, &f.trimmingAppointment.ID, nil, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "appointment-only trimming id without course/option must still match medical_record appointment: %v", err)

		_, err = f.validate(t, nil, &f.treatment.ID, &f.trimmingAppointment.ID, &f.course.ID, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "treatment plus trimming provenance must not skip medical_record appointment equality: %v", err)
	})

	// BUG-506 / UAT S11: complete clients may omit appointment_id while still sending
	// trimming_course_id / trimming_option_id from unbilled candidates. Resolve the
	// unique accounting-status appointment for the billing pet instead of 400.
	t.Run("BUG-506: trimming course/option without appointment_id resolves unique accounting appointment", func(t *testing.T) {
		f := setupSplitAppointmentReferenceFixture(t, true)

		_, err := f.validate(t, nil, nil, nil, &f.course.ID, &f.option.ID)
		require.NoError(t, err)

		_, err = f.validate(t, nil, nil, nil, &f.course.ID, nil)
		require.NoError(t, err)

		_, err = f.validate(t, nil, nil, nil, nil, &f.option.ID)
		require.NoError(t, err)
	})
}

func TestBillingItemRepository_ValidateCreateReferences_HoldsBillingLockUntilAmbientTransactionCommits(
	t *testing.T,
) {
	f := setupBillingItemReferenceFixture(t)
	tx := f.db.Begin()
	require.NoError(t, tx.Error)
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback().Error
		}
	}()

	txCtx := persistence.WithTxValue(context.Background(), tx)
	_, err := f.repo.ValidateCreateReferences(
		txCtx,
		f.clinicID,
		f.billing.ID,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	competingTx := f.db.Begin()
	require.NoError(t, competingTx.Error)
	require.NoError(t, competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	err = competingTx.
		Model(&model.Billing{}).
		Where("id = ?", f.billing.ID).
		Update("status", model.BillingStatusCompleted).Error
	require.ErrorContains(
		t,
		err,
		"lock timeout",
		"competing billing update must time out while the ambient transaction holds its lock",
	)
	require.NoError(t, competingTx.Rollback().Error)

	require.NoError(t, tx.Commit().Error)
	committed = true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, f.db.WithContext(ctx).
		Model(&model.Billing{}).
		Where("id = ?", f.billing.ID).
		Update("status", model.BillingStatusCompleted).Error)
}

func newBillingItemReferenceService(
	f billingItemReferenceFixture,
	repo BillingItemRepository,
	opts ...billingItemServiceOption,
) BillingItemService {
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID || billingID != f.billing.ID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		return f.billing, nil
	}
	// BUG-440: vaccination claim 解放は auditTx 必須（fail-closed）。参照フィクスチャでも
	// production 同様に監査依存を配線する（既定は noop、検証時は WithBillingItemAuditTx で上書き）。
	defaultOpts := []billingItemServiceOption{WithBillingItemAuditTx(&noopAuditTxLogger{})}
	defaultOpts = append(defaultOpts, opts...)
	return NewBillingItemServiceWithCampaign(
		repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
		defaultOpts...,
	)
}

func billingItemReferenceCreateInput(f billingItemReferenceFixture) *CreateBillingItemInput {
	return &CreateBillingItemInput{
		ClinicID:  f.clinicID,
		BillingID: f.billing.ID,
		Category:  string(model.ItemCategoryExamination),
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

func TestBillingItemService_CreateItem_OtherMetadata(t *testing.T) {
	actorID := uint64(7)
	merchandiseItemID := uint64(50)
	tests := []struct {
		name             string
		input            *CreateBillingItemInput
		resolvedCategory model.ItemCategory
		wantErr          bool
		check            func(t *testing.T, item *model.BillingItem)
	}{
		{
			name: "rejects manual other without other_reason",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryOther),
				Name: "その他調整", UnitPrice: 500, Quantity: 1,
				Source: string(model.ItemSourceManual), CreatedBy: &actorID,
			},
			wantErr: true,
		},
		{
			name: "rejects manual other with blank other_reason",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryOther),
				Name: "その他調整", UnitPrice: 500, Quantity: 1,
				Source: string(model.ItemSourceManual), OtherReason: ptrString("   "), CreatedBy: &actorID,
			},
			wantErr: true,
		},
		{
			name: "rejects manual other with other_reason over 500 Unicode characters",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryOther),
				Name: "その他調整", UnitPrice: 500, Quantity: 1,
				Source: string(model.ItemSourceManual), OtherReason: ptrString(strings.Repeat("界", 501)), CreatedBy: &actorID,
			},
			wantErr: true,
		},
		{
			name: "rejects manual other without authenticated actor",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryOther),
				Name: "その他調整", UnitPrice: 500, Quantity: 1,
				Source: string(model.ItemSourceManual), OtherReason: ptrString("レジ締め分類確認"),
			},
			wantErr: true,
		},
		{
			name: "persists trimmed other_reason and authenticated actor",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryOther),
				Name: "その他調整", UnitPrice: 500, Quantity: 1,
				Source: string(model.ItemSourceManual), OtherReason: ptrString("  レジ締め分類確認  "), CreatedBy: &actorID,
			},
			check: func(t *testing.T, item *model.BillingItem) {
				require.NotNil(t, item.OtherReason)
				assert.Equal(t, "レジ締め分類確認", *item.OtherReason)
				require.NotNil(t, item.CreatedBy)
				assert.Equal(t, actorID, *item.CreatedBy)
			},
		},
		{
			name: "does not persist metadata for non-other manual category",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryTest),
				Name: "検査調整", UnitPrice: 500, Quantity: 1,
				Source: string(model.ItemSourceManual), OtherReason: ptrString("保存しない"), CreatedBy: &actorID,
			},
			check: func(t *testing.T, item *model.BillingItem) {
				assert.Nil(t, item.OtherReason)
				assert.Nil(t, item.CreatedBy)
			},
		},
		{
			name: "does not require metadata for non-manual other source",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryOther),
				Name: "カルテ由来その他", UnitPrice: 500, Quantity: 1,
				Source: string(model.ItemSourceMedicalRecord), CreatedBy: &actorID,
			},
			check: func(t *testing.T, item *model.BillingItem) {
				assert.Nil(t, item.OtherReason)
				assert.Nil(t, item.CreatedBy)
			},
		},
		{
			name: "does not require metadata for merchandise-derived other",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryTest),
				Name: "物販その他", UnitPrice: 500, Quantity: 1,
				Source: string(model.ItemSourceManual), OtherReason: ptrString("保存しない"),
				CreatedBy: &actorID, MerchandiseItemID: &merchandiseItemID,
			},
			resolvedCategory: model.ItemCategoryOther,
			check: func(t *testing.T, item *model.BillingItem) {
				assert.Equal(t, model.ItemCategoryOther, item.Category)
				assert.Nil(t, item.OtherReason)
				assert.Nil(t, item.CreatedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := defaultMockBillingItemRepo()
			repo.validateCreateReferencesFn = func(_ context.Context, _, _ uint64, _, _, _, _, _ *uint64) (model.ItemCategory, error) {
				return tt.resolvedCategory, nil
			}
			repo.createFn = func(_ context.Context, item *model.BillingItem) error {
				item.ID = 1
				return nil
			}
			svc := NewBillingItemServiceWithCampaign(
				repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{},
				okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil,
			)

			item, err := svc.CreateItem(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, item)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, item)
			if tt.check != nil {
				tt.check(t, item)
			}
		})
	}
}

func TestToBillingItemResponse_OtherReasonRoundTripsWithoutCreatedBy(t *testing.T) {
	reason := "レジ締め分類確認"
	createdBy := uint64(7)

	got := ToBillingItemResponse(&model.BillingItem{
		ID: 1, BillingID: 10, Category: model.ItemCategoryOther, Name: "手入力",
		UnitPrice: 500, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
		Source: model.ItemSourceManual, OtherReason: &reason, CreatedBy: &createdBy,
	})

	require.NotNil(t, got.OtherReason)
	assert.Equal(t, reason, *got.OtherReason)
}

func TestBillingItemService_CreateItem_RejectsInvalidOtherActors(t *testing.T) {
	tests := []struct {
		name       string
		setupActor func(t *testing.T, f billingItemReferenceFixture) uint64
	}{
		{
			name: "cross-clinic actor",
			setupActor: func(t *testing.T, f billingItemReferenceFixture) uint64 {
				const foreignClinicID = uint64(2)
				actor := makeDoctor(t, f.db, foreignClinicID, "cross-clinic billing actor")
				assignBillingConfirmationStaff(t, f.db, actor.ID, foreignClinicID)
				return actor.ID
			},
		},
		{
			name: "inactive assigned actor",
			setupActor: func(t *testing.T, f billingItemReferenceFixture) uint64 {
				actor := makeDoctor(t, f.db, f.clinicID, "inactive billing actor")
				assignBillingConfirmationStaff(t, f.db, actor.ID, f.clinicID)
				require.NoError(t, f.db.Model(&model.Staff{}).
					Where("id = ?", actor.ID).
					Update("is_active", false).Error)
				return actor.ID
			},
		},
		{
			name: "revoked assignment actor",
			setupActor: func(t *testing.T, f billingItemReferenceFixture) uint64 {
				actor := makeDoctor(t, f.db, f.clinicID, "revoked billing actor")
				assignment := assignBillingConfirmationStaff(t, f.db, actor.ID, f.clinicID)
				require.NoError(t, f.db.Delete(assignment).Error)
				return actor.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupBillingItemReferenceFixture(t)
			actorID := tt.setupActor(t, f)
			before := countBillingItems(t, f.db)
			svc := newBillingItemReferenceService(f, f.repo)

			item, err := svc.CreateItem(context.Background(), &CreateBillingItemInput{
				ClinicID:  f.clinicID,
				BillingID: f.billing.ID,
				Category:  string(model.ItemCategoryOther),
				Name:      "actor-isolated other item",
				UnitPrice: 500,
				Quantity:  1,
				Source:    string(model.ItemSourceManual),
				OtherReason: ptrString(
					"actor assignment verification",
				),
				CreatedBy: &actorID,
			})

			require.Error(t, err)
			assert.True(t, errors.Is(err, apperrors.ErrForbidden))
			assert.Nil(t, item)
			assert.Equal(t, before, countBillingItems(t, f.db), "rejected actor must not persist a billing item")
		})
	}
}

func TestCreateBillingItem_AuthenticatedActorAndPublicResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reason := "レジ締め分類確認"
	svc := &mockBillingItemService{
		createItemFn: func(_ context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
			require.NotNil(t, input.CreatedBy)
			assert.Equal(t, uint64(7), *input.CreatedBy)
			return &model.BillingItem{
				ID:          5,
				BillingID:   input.BillingID,
				Category:    model.ItemCategoryOther,
				Name:        input.Name,
				UnitPrice:   input.UnitPrice,
				Quantity:    input.Quantity,
				TaxType:     model.TaxTypeExcluded,
				TaxRate:     0.10,
				Source:      model.ItemSourceManual,
				OtherReason: &reason,
				CreatedBy:   input.CreatedBy,
			}, nil
		},
	}
	handler := newHandlerWithBillingItemSvc(svc)
	body, err := json.Marshal(map[string]any{
		"billing_id":   10,
		"category":     string(model.ItemCategoryOther),
		"name":         "その他調整",
		"unit_price":   500,
		"quantity":     1,
		"source":       string(model.ItemSourceManual),
		"other_reason": reason,
	})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("clinic_id", "1")
	c.Set("user_id", "7")

	handler.CreateBillingItem(c)

	require.Equal(t, http.StatusCreated, recorder.Code)
	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, reason, response["other_reason"])
	_, exposesCreatedBy := response["created_by"]
	assert.False(t, exposesCreatedBy, "created_by is internal audit metadata")
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
		assert.Equal(t, model.ItemCategoryGoods, item.Category)

		var stored model.BillingItem
		require.NoError(t, f.db.First(&stored, item.ID).Error)
		assert.Equal(t, model.ItemCategoryGoods, stored.Category)
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

func TestBillingItemService_CreateItem_UsesCategoryCommittedBeforeShareLock(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	updateTx := f.db.Begin()
	require.NoError(t, updateTx.Error)
	committed := false
	defer func() {
		if !committed {
			_ = updateTx.Rollback().Error
		}
	}()
	require.NoError(t, updateTx.Model(&model.MerchandiseItem{}).
		Where("id = ?", f.merchandiseItem.ID).
		Update("category", model.ItemCategoryFood).Error)

	repo := &validationStartBillingItemRepository{
		BillingItemRepository: f.repo,
		started:               make(chan struct{}),
	}
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.Billing, error) {
		return f.billing, nil
	}
	campaignCategory := make(chan model.ItemCategory, 1)
	campaignRepo := &mockCampaignRepository{
		findApplicableForItemFn: func(_ context.Context, _ uint64, _ time.Time, category model.ItemCategory, _ *uint64) (*model.Campaign, error) {
			campaignCategory <- category
			return nil, nil
		},
	}
	svc := NewBillingItemServiceWithCampaign(
		repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		campaignRepo,
		nil,
	)
	input := billingItemReferenceCreateInput(f)
	input.Category = string(model.ItemCategoryMedicine)
	input.MerchandiseItemID = &f.merchandiseItem.ID

	type createResult struct {
		item *model.BillingItem
		err  error
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := make(chan createResult, 1)
	go func() {
		item, err := svc.CreateItem(ctx, input)
		result <- createResult{item: item, err: err}
	}()

	select {
	case <-repo.started:
	case <-ctx.Done():
		t.Fatal("CreateItem did not reach transactional reference validation")
	}
	select {
	case got := <-result:
		t.Fatalf("CreateItem completed before the pending merchandise category update committed: %v", got.err)
	case <-time.After(100 * time.Millisecond):
	}

	require.NoError(t, updateTx.Commit().Error)
	committed = true

	var got createResult
	select {
	case got = <-result:
	case <-ctx.Done():
		t.Fatal("CreateItem did not resume after the merchandise category update committed")
	}
	require.NoError(t, got.err)
	require.NotNil(t, got.item)
	assert.Equal(t, model.ItemCategoryFood, got.item.Category)
	select {
	case category := <-campaignCategory:
		assert.Equal(t, model.ItemCategoryFood, category)
	case <-ctx.Done():
		t.Fatal("campaign lookup did not receive the resolved merchandise category")
	}

	var stored model.BillingItem
	require.NoError(t, f.db.First(&stored, got.item.ID).Error)
	assert.Equal(t, model.ItemCategoryFood, stored.Category)
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
