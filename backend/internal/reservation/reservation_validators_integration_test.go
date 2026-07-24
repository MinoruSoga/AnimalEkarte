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

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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
