package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type failingLiffTrimmingDetailRepository struct {
	db            *gorm.DB
	setOptionsErr error
}

func (r failingLiffTrimmingDetailRepository) Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	return persistence.DBOrTx(ctx, r.db).Create(detail).Error
}

func (r failingLiffTrimmingDetailRepository) SetOptions(context.Context, uint64, uint64, []uint64) error {
	return r.setOptionsErr
}

func TestReservationValidators_RealDBRejectsForeignStaffAndRollsBackTrimmingGraph(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	reservationRepo := NewReservationRepository(db)
	typeRepo := NewReservationTypeRepository(db)
	staffRepo := NewReservationStaffRepository(db, nil)
	courseRepo := &mockTrimmingCourseFinder{findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
		var course model.TrimmingCourse
		query := persistence.DBOrTx(ctx, db)
		if persistence.TxFromContext(ctx) != nil {
			query = query.Clauses(clause.Locking{Strength: "SHARE"})
		}
		if err := query.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&course).Error; err != nil {
			return nil, err
		}
		return &course, nil
	}}
	optionRepo := &mockTrimmingOptionFinder{findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
		var option model.TrimmingOption
		query := persistence.DBOrTx(ctx, db)
		if persistence.TxFromContext(ctx) != nil {
			query = query.Clauses(clause.Locking{Strength: "SHARE"})
		}
		if err := query.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&option).Error; err != nil {
			return nil, err
		}
		return &option, nil
	}}
	customer := makeLineCustomerForAdmin(t, db, clinicA, "validator-real-db-customer")
	foreignCustomer := makeLineCustomerForAdmin(t, db, clinicB, "validator-foreign-customer")
	validStaff := makeDoctor(t, db, clinicA, "医院A LIFF担当")
	foreignStaff := makeDoctor(t, db, clinicB, "医院B LIFF担当")
	for _, item := range []model.StaffClinicAssignment{
		{StaffID: validStaff.ID, ClinicID: clinicA, IsMain: true},
		{StaffID: foreignStaff.ID, ClinicID: clinicB, IsMain: true},
	} {
		require.NoError(t, db.Create(&item).Error)
	}
	reservationType := &model.ReservationType{
		ClinicID:           clinicA,
		Name:               "公開トリミング予約区分",
		IsActive:           true,
		ReservationVisible: true,
		Category:           model.ReservationTypeCategoryTrimming,
	}
	require.NoError(t, db.Create(reservationType).Error)
	require.NoError(t, db.Create(&model.StaffReservationCapability{
		ClinicID:          clinicA,
		StaffID:           validStaff.ID,
		ReservationTypeID: reservationType.ID,
	}).Error)
	course := &model.TrimmingCourse{ClinicID: clinicA, Name: "公開コース", IsActive: true}
	option := &model.TrimmingOption{ClinicID: clinicA, Name: "公開オプション", IsActive: true}
	require.NoError(t, db.Create(course).Error)
	require.NoError(t, db.Create(option).Error)

	baseInput := func(staffID uint64) *CreateReservationInput {
		courseID := course.ID
		return &CreateReservationInput{
			ClinicID:          clinicA,
			CustomerID:        customer.ID,
			ReservationTypeID: reservationType.ID,
			StaffID:           staffID,
			Date:              dateInDays(3),
			StartTime:         "1000",
			EndTime:           "1015",
			Settings:          newSettingForValidation(),
			TrimmingCourseID:  &courseID,
			TrimmingOptionIDs: []uint64{option.ID},
		}
	}

	t.Run("foreign staff is rejected before appointment insert", func(t *testing.T) {
		validators := NewReservationValidators(
			testNewTransactor(db), reservationRepo, typeRepo, staffRepo,
			courseRepo, optionRepo, failingLiffTrimmingDetailRepository{db: db},
		)
		out, err := validators.ValidateAndCreate(ctx, baseInput(foreignStaff.ID))
		require.Error(t, err)
		assert.Nil(t, out)

		var count int64
		require.NoError(t, db.Model(&model.Reservation{}).Where("line_customer_id = ?", customer.ID).Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("option failure rolls back appointment and detail", func(t *testing.T) {
		sentinel := errors.New("set options failed")
		validators := NewReservationValidators(
			testNewTransactor(db), reservationRepo, typeRepo, staffRepo,
			courseRepo, optionRepo, failingLiffTrimmingDetailRepository{db: db, setOptionsErr: sentinel},
		)
		input := baseInput(validStaff.ID)
		input.Date = time.Now().In(input.Date.Location()).AddDate(0, 0, 3)
		out, err := validators.ValidateAndCreate(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, sentinel)
		assert.Nil(t, out)

		var appointmentCount, detailCount int64
		require.NoError(t, db.Model(&model.Reservation{}).Where("line_customer_id = ?", customer.ID).Count(&appointmentCount).Error)
		require.NoError(t, db.Model(&model.AppointmentTrimmingDetail{}).Count(&detailCount).Error)
		assert.Zero(t, appointmentCount)
		assert.Zero(t, detailCount)
	})

	t.Run("foreign line customer is rejected before appointment insert", func(t *testing.T) {
		validators := NewReservationValidators(
			testNewTransactor(db), reservationRepo, typeRepo, staffRepo,
			courseRepo, optionRepo, failingLiffTrimmingDetailRepository{db: db},
		)
		input := baseInput(validStaff.ID)
		input.CustomerID = foreignCustomer.ID
		input.Date = dateInDays(4)

		out, err := validators.ValidateAndCreate(ctx, input)
		require.Error(t, err)
		assert.Nil(t, out)
		var count int64
		require.NoError(t, db.Model(&model.Reservation{}).
			Where("clinic_id = ? AND line_customer_id = ?", clinicA, foreignCustomer.ID).
			Count(&count).Error)
		assert.Zero(t, count)
	})
}

func TestReservationValidators_RealDBRejectsInactiveReservationTypeWithoutPartialWrites(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(23801)
	testdb.SeedClinicsForFK(t, db, clinicID)

	customer := makeLineCustomerForAdmin(t, db, clinicID, "issue-238-inactive-type-customer")
	reservationType := &model.ReservationType{
		ClinicID:           clinicID,
		Name:               "無効な公開予約区分",
		IsActive:           true,
		ReservationVisible: true,
		IsInternal:         false,
		Category:           model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, db.WithContext(ctx).Create(reservationType).Error)
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Where("clinic_id = ? AND id = ?", clinicID, reservationType.ID).
		Update("is_active", false).Error)

	var auditCountBefore int64
	require.NoError(t, db.WithContext(ctx).
		Model(&model.AuditLog{}).
		Where("clinic_id = ? AND resource = ?", clinicID, model.AuditResourceReservation).
		Count(&auditCountBefore).Error)

	validators := NewReservationValidators(
		testNewTransactor(db),
		NewReservationRepository(db),
		NewReservationTypeRepository(db),
		NewReservationStaffRepository(db, nil),
		nil,
		nil,
		nil,
	)
	out, err := validators.ValidateAndCreate(ctx, &CreateReservationInput{
		ClinicID:          clinicID,
		CustomerID:        customer.ID,
		ReservationTypeID: reservationType.ID,
		Date:              dateInDays(3),
		StartTime:         "1000",
		EndTime:           "1015",
		Settings:          newSettingForValidation(),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, out)

	var appointmentCount, auditCountAfter int64
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Reservation{}).
		Where(
			"clinic_id = ? AND line_customer_id = ? AND reservation_type_id = ?",
			clinicID,
			customer.ID,
			reservationType.ID,
		).
		Count(&appointmentCount).Error)
	require.NoError(t, db.WithContext(ctx).
		Model(&model.AuditLog{}).
		Where("clinic_id = ? AND resource = ?", clinicID, model.AuditResourceReservation).
		Count(&auditCountAfter).Error)
	assert.Zero(t, appointmentCount)
	assert.Equal(t, auditCountBefore, auditCountAfter)
}

func TestReservationAdminRepository_FindAllByCustomerID_KeepsInactiveReservationTypeHistory(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(23802)
	testdb.SeedClinicsForFK(t, db, clinicID)

	customer := makeLineCustomerForAdmin(t, db, clinicID, "issue-238-historical-customer")
	reservationType := &model.ReservationType{
		ClinicID:           clinicID,
		Name:               "履歴に残す予約区分",
		IsActive:           true,
		ReservationVisible: true,
		Category:           model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, db.WithContext(ctx).Create(reservationType).Error)
	appointment := makeReservationForReservationRepoTest(
		t,
		db,
		clinicID,
		reservationType.ID,
		time.Now().UTC().Truncate(time.Minute),
		model.ReservationStatusConfirmed,
		model.ReservationSourceLine,
		&customer.ID,
		nil,
	)
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Where("clinic_id = ? AND id = ?", clinicID, reservationType.ID).
		Update("is_active", false).Error)

	items, err := NewReservationAdminRepository(db).FindAllByCustomerID(ctx, clinicID, customer.ID)

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, appointment.ID, items[0].ID)
	require.NotNil(t, items[0].ReservationType)
	assert.Equal(t, reservationType.ID, items[0].ReservationType.ID)
	assert.Equal(t, reservationType.Name, items[0].ReservationType.Name)
	assert.False(t, items[0].ReservationType.IsActive)
}
