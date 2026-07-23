package trimming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func rejectTrimmingCourseRepo(ownedID uint64) TrimmingCourseRepository {
	return &mockTrimmingCourseRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_course", "foreign")
		}
		return &model.TrimmingCourse{ID: id, IsActive: true}, nil
	}}
}

func rejectTrimmingOptionRepo(ownedID uint64) TrimmingOptionRepository {
	return &mockTrimmingOptionRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_option", "foreign")
		}
		return &model.TrimmingOption{ID: id, IsActive: true}, nil
	}}
}

func TestTrimmingCourseService_Update_RejectsCrossClinicCourseTypeFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedCourseTypeID = uint64(10)
	const foreignCourseTypeID = uint64(999)

	newService := func(updated *bool) TrimmingCourseService {
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
		return NewTrimmingCourseService(repo, courseTypeRepo, &mockTransactor{})
	}

	t.Run("rejects cross-clinic course_type_id and does not persist", func(t *testing.T) {
		updated := false
		foreign := foreignCourseTypeID
		result, err := newService(&updated).Update(context.Background(), clinicID, entityID, &UpdateTrimmingCourseInput{CourseTypeID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.False(t, updated)
	})

	t.Run("accepts same-clinic course_type_id", func(t *testing.T) {
		updated := false
		owned := ownedCourseTypeID
		result, err := newService(&updated).Update(context.Background(), clinicID, entityID, &UpdateTrimmingCourseInput{CourseTypeID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, updated)
	})
}

func TestTrimmingService_Create_RejectsCrossClinicCourseFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedID = uint64(10)
	const foreignID = uint64(999)

	newService := func(created *bool, courseRepo TrimmingCourseRepository) TrimmingService {
		reservationRepo := &mockTrimmingReservationRepository{
			createFn: func(_ context.Context, appointment *model.Reservation) error {
				*created = true
				appointment.ID = 1
				return nil
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, ClinicID: clinicID}, nil
			},
		}
		return withTrimmingTestActor(NewTrimmingServiceWithAudit(reservationRepo, &mockTrimmingReservationTypeRepository{}, nil, &mockTrimmingUnavailableTimeRepository{},
			&mockTrimmingDetailRepository{}, courseRepo, okTrimmingOptionRepo(), &mockTransactor{}, noopTrimmingAuditTxLogger{}))
	}

	t.Run("rejects cross-clinic course_id", func(t *testing.T) {
		created := false
		foreign := foreignID
		result, err := newService(&created, rejectTrimmingCourseRepo(ownedID)).Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), CourseID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.False(t, created)
	})

	t.Run("accepts same-clinic course_id", func(t *testing.T) {
		created := false
		owned := ownedID
		result, err := newService(&created, rejectTrimmingCourseRepo(ownedID)).Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), CourseID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, created)
	})
}

func TestTrimmingService_Create_RejectsCrossClinicOptionFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedID = uint64(20)
	const foreignID = uint64(998)

	newService := func(created *bool, optionRepo TrimmingOptionRepository) TrimmingService {
		reservationRepo := &mockTrimmingReservationRepository{
			createFn: func(_ context.Context, appointment *model.Reservation) error {
				*created = true
				appointment.ID = 1
				return nil
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, ClinicID: clinicID}, nil
			},
		}
		return withTrimmingTestActor(NewTrimmingServiceWithAudit(reservationRepo, &mockTrimmingReservationTypeRepository{}, nil, &mockTrimmingUnavailableTimeRepository{},
			&mockTrimmingDetailRepository{}, okTrimmingCourseRepo(), optionRepo, &mockTransactor{}, noopTrimmingAuditTxLogger{}))
	}

	t.Run("rejects cross-clinic option_id", func(t *testing.T) {
		created := false
		result, err := newService(&created, rejectTrimmingOptionRepo(ownedID)).Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), OptionIDs: []uint64{foreignID},
		})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.False(t, created)
	})

	t.Run("accepts same-clinic option_id", func(t *testing.T) {
		created := false
		result, err := newService(&created, rejectTrimmingOptionRepo(ownedID)).Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), OptionIDs: []uint64{ownedID},
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, created)
	})
}

func TestTrimmingService_Update_RejectsCrossClinicCourseFK(t *testing.T) {
	const clinicID = uint64(1)
	const appointmentID = uint64(1)
	const ownedID = uint64(10)
	const foreignID = uint64(999)

	newService := func(updated *bool, courseRepo TrimmingCourseRepository) TrimmingService {
		detail := &mockTrimmingDetailRepository{updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			*updated = true
			return nil
		}}
		reservationRepo := &mockTrimmingReservationRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		}}
		return withTrimmingTestActor(NewTrimmingServiceWithAudit(reservationRepo, &mockTrimmingReservationTypeRepository{}, nil, nil,
			detail, courseRepo, okTrimmingOptionRepo(), &mockTransactor{}, noopTrimmingAuditTxLogger{}))
	}

	t.Run("rejects cross-clinic course_id", func(t *testing.T) {
		updated := false
		foreign := foreignID
		result, err := newService(&updated, rejectTrimmingCourseRepo(ownedID)).Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{CourseID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.False(t, updated)
	})

	t.Run("accepts same-clinic course_id", func(t *testing.T) {
		updated := false
		owned := ownedID
		result, err := newService(&updated, rejectTrimmingCourseRepo(ownedID)).Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{CourseID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, updated)
	})
}

func TestTrimmingService_Update_RejectsCrossClinicOptionFK(t *testing.T) {
	const clinicID = uint64(1)
	const appointmentID = uint64(1)
	const ownedID = uint64(20)
	const foreignID = uint64(998)

	newService := func(updated *bool, optionRepo TrimmingOptionRepository) TrimmingService {
		detail := &mockTrimmingDetailRepository{setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
			*updated = true
			return nil
		}}
		reservationRepo := &mockTrimmingReservationRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		}}
		return withTrimmingTestActor(NewTrimmingServiceWithAudit(reservationRepo, &mockTrimmingReservationTypeRepository{}, nil, nil,
			detail, okTrimmingCourseRepo(), optionRepo, &mockTransactor{}, noopTrimmingAuditTxLogger{}))
	}

	t.Run("rejects cross-clinic option_id", func(t *testing.T) {
		updated := false
		foreign := []uint64{foreignID}
		result, err := newService(&updated, rejectTrimmingOptionRepo(ownedID)).Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{OptionIDs: &foreign})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.False(t, updated)
	})

	t.Run("accepts same-clinic option_id", func(t *testing.T) {
		updated := false
		owned := []uint64{ownedID}
		result, err := newService(&updated, rejectTrimmingOptionRepo(ownedID)).Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{OptionIDs: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, updated)
	})
}
