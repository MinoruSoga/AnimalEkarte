package trimming

// #228: 無効化(is_active=false)されたトリミングコース/オプションを新規に紐付けさせない。
// 既にカルテに紐づいている無効 ID は維持を許可する（データを消さない）。
//
// cross_tenant_master_fk_write_test.go の dual regression パターン（reject-new /
// accept-existing）を踏襲する。okTrimmingCourseRepo()/okTrimmingOptionRepo() は
// IsActive: true を返す（cross_tenant テストの前提を壊さないための必須修正、同ファイル参照）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func inactiveTrimmingCourseRepo() TrimmingCourseRepository {
	return &mockTrimmingCourseRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		return &model.TrimmingCourse{ID: id, IsActive: false}, nil
	}}
}

func inactiveTrimmingOptionRepo() TrimmingOptionRepository {
	return &mockTrimmingOptionRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		return &model.TrimmingOption{ID: id, IsActive: false}, nil
	}}
}

func TestTrimmingService_Create_RejectsInactiveCourseFK(t *testing.T) {
	const clinicID = uint64(1)
	const inactiveCourseID = uint64(10)

	newSvc := func(created *bool) TrimmingService {
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
		return withTrimmingTestActor(NewTrimmingServiceWithAudit(reserv, &mockTrimmingReservationTypeRepository{}, nil, &mockTrimmingUnavailableTimeRepository{},
			&mockTrimmingDetailRepository{}, inactiveTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTransactor{}, noopTrimmingAuditTxLogger{}))
	}

	t.Run("rejects newly referenced inactive course_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		course := inactiveCourseID
		out, err := svc.Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1,
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(time.Hour),
			CourseID:          &course,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "trimming appointment must NOT be persisted when newly referencing an inactive course")
	})
}

func TestTrimmingService_Create_RejectsInactiveOptionFK(t *testing.T) {
	const clinicID = uint64(1)
	const inactiveOptionID = uint64(20)

	newSvc := func(created *bool) TrimmingService {
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
		return withTrimmingTestActor(NewTrimmingServiceWithAudit(reserv, &mockTrimmingReservationTypeRepository{}, nil, &mockTrimmingUnavailableTimeRepository{},
			&mockTrimmingDetailRepository{}, okTrimmingCourseRepo(), inactiveTrimmingOptionRepo(), &mockTransactor{}, noopTrimmingAuditTxLogger{}))
	}

	t.Run("rejects newly referenced inactive option_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateTrimmingInput{
			ReservationTypeID: 1,
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(time.Hour),
			OptionIDs:         []uint64{inactiveOptionID},
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "trimming appointment must NOT be persisted when newly referencing an inactive option")
	})
}

func TestTrimmingService_Update_InactiveCourseFKGuard(t *testing.T) {
	const clinicID = uint64(1)
	const appointmentID = uint64(1)
	const inactiveCourseID = uint64(10)

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
	}

	t.Run("rejects newly added inactive course_id on update and does not persist", func(t *testing.T) {
		updated := false
		detail := &mockTrimmingDetailRepository{
			// 既存カルテには course が紐づいていない（新規追加シナリオ）。
			updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
				updated = true
				return nil
			},
		}
		svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(reserv, &mockTrimmingReservationTypeRepository{}, nil, nil,
			detail, inactiveTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTransactor{}, noopTrimmingAuditTxLogger{}))

		course := inactiveCourseID
		out, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{CourseID: &course})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "trimming detail must NOT be updated to newly reference an inactive course")
	})

	t.Run("allows retaining an already-linked inactive course_id (no data loss)", func(t *testing.T) {
		updated := false
		detail := &mockTrimmingDetailRepository{
			findByAppointmentIDFn: func(_ context.Context, _, apptID uint64) (*model.AppointmentTrimmingDetail, error) {
				existing := inactiveCourseID
				return &model.AppointmentTrimmingDetail{AppointmentID: apptID, CourseID: &existing}, nil
			},
			updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
				updated = true
				return nil
			},
		}
		svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(reserv, &mockTrimmingReservationTypeRepository{}, nil, nil,
			detail, inactiveTrimmingCourseRepo(), okTrimmingOptionRepo(), &mockTransactor{}, noopTrimmingAuditTxLogger{}))

		// クライアントが既存カルテと同じ無効化済み course_id を再送（維持）するケース。
		sameCourse := inactiveCourseID
		out, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{CourseID: &sameCourse})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

func TestTrimmingService_Update_InactiveOptionFKGuard(t *testing.T) {
	const clinicID = uint64(1)
	const appointmentID = uint64(1)
	const inactiveOptionID = uint64(20)

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
	}

	t.Run("rejects newly added inactive option_id on update and does not persist", func(t *testing.T) {
		updated := false
		detail := &mockTrimmingDetailRepository{
			// 既存カルテにはオプションが紐づいていない（新規追加シナリオ）。
			setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
				updated = true
				return nil
			},
		}
		svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(reserv, &mockTrimmingReservationTypeRepository{}, nil, nil,
			detail, okTrimmingCourseRepo(), inactiveTrimmingOptionRepo(), &mockTransactor{}, noopTrimmingAuditTxLogger{}))

		newOptions := []uint64{inactiveOptionID}
		out, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{OptionIDs: &newOptions})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "trimming options must NOT be updated to newly reference an inactive option")
	})

	t.Run("allows retaining an already-linked inactive option_id (no data loss)", func(t *testing.T) {
		updated := false
		detail := &mockTrimmingDetailRepository{
			findByAppointmentIDFn: func(_ context.Context, _, apptID uint64) (*model.AppointmentTrimmingDetail, error) {
				return &model.AppointmentTrimmingDetail{
					AppointmentID: apptID,
					Options:       []model.TrimmingOption{{ID: inactiveOptionID, IsActive: false}},
				}, nil
			},
			setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
				updated = true
				return nil
			},
		}
		svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(reserv, &mockTrimmingReservationTypeRepository{}, nil, nil,
			detail, okTrimmingCourseRepo(), inactiveTrimmingOptionRepo(), &mockTransactor{}, noopTrimmingAuditTxLogger{}))

		// クライアントが既存カルテと同じ無効化済み option_id 集合を再送（維持）するケース。
		sameOptions := []uint64{inactiveOptionID}
		out, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{OptionIDs: &sameOptions})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}
