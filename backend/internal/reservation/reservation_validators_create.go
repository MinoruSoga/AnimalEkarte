package reservation

import (
	"context"
	"errors"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (v *reservationValidators) createLineReservationInTx(
	ctx context.Context,
	input *CreateReservationInput,
	settings *model.LineReservationSetting,
) (*model.Reservation, error) {
	if err := v.repo.AcquireBookingLock(ctx, input.ClinicID); err != nil {
		return nil, err
	}
	if err := v.repo.AssertLineCustomerInClinic(ctx, input.ClinicID, input.CustomerID); err != nil {
		return nil, apperrors.Wrap(err, "failed to verify LINE customer ownership")
	}
	reservationType, err := v.validateReservationMasterOwnership(ctx, input)
	if err != nil {
		return nil, err
	}
	if input.StaffID != 0 {
		if err := ValidateLineReservationStaffCapability(ctx, v.staffRepo, input.ClinicID, &input.StaffID, input.ReservationTypeID); err != nil {
			return nil, err
		}
	}
	startDT, err := ToDateTime(input.Date, input.StartTime)
	if err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	endDT, err := ToDateTime(input.Date, input.EndTime)
	if err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if err := validateTimeRange(startDT, endDT); err != nil {
		return nil, err
	}
	if err := validateClinicHoliday(ctx, v.holidayFinder, input.ClinicID, startDT); err != nil {
		return nil, err
	}

	var doctorIDPtr *uint64
	if input.StaffID != 0 {
		id := input.StaffID
		doctorIDPtr = &id
	}
	if err := CheckSlotConflict(ctx, v.repo, input.ClinicID, doctorIDPtr, startDT, endDT, nil); err != nil {
		if errors.Is(err, errNoDoctorsOnDuty) {
			return nil, &ReservationLimitError{
				Code:         "SLOT_TAKEN",
				Message:      "本日は医師が出勤していません。別の日をお選びください。",
				RedirectStep: 4,
			}
		}
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, &ReservationLimitError{
				Code:         "SLOT_TAKEN",
				Message:      "選択された時間枠は既に予約が入っています。別の時間をお選びください。",
				RedirectStep: 5,
			}
		}
		return nil, err
	}
	if err := CheckReservationTypeCapacity(ctx, v.repo, v.typeRepo, input.ClinicID, input.ReservationTypeID, startDT, nil); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, &ReservationLimitError{
				Code:         "SLOT_TAKEN",
				Message:      "選択された時間枠は満員です。別の時間をお選びください。",
				RedirectStep: 5,
			}
		}
		return nil, err
	}

	dayStart := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, input.Date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	if err := checkCustomerReservationLimit(ctx, v.repo, input, dayStart, dayEnd,
		"DAILY_LIMIT", "1日内に予約できる件数を超えています。別の日をお選びください。",
		"failed to count daily reservations", settings.DailyLimit); err != nil {
		return nil, err
	}
	monthStart := time.Date(input.Date.Year(), input.Date.Month(), 1, 0, 0, 0, 0, input.Date.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	if err := checkCustomerReservationLimit(ctx, v.repo, input, monthStart, monthEnd,
		"MONTHLY_LIMIT", "1ヶ月内に予約できる件数を超えています。別の月をお選びください。",
		"failed to count monthly reservations", settings.MonthlyLimit); err != nil {
		return nil, err
	}

	confirmationNumber, err := generateConfirmationNumber(ctx, v.repo, input.ClinicID, input.Date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to generate confirmation number")
	}
	appt := buildLineReservation(input, startDT, endDT, confirmationNumber)
	if err := v.repo.Create(ctx, appt); err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation")
	}
	if reservationType.Category == model.ReservationTypeCategoryTrimming && hasLineTrimmingDetailInput(input) {
		if v.trimmingDetailRepo == nil {
			return nil, apperrors.WrapInternalServerError("trimming detail repository is required for a LINE trimming reservation")
		}
		detail := &model.AppointmentTrimmingDetail{
			ClinicID:      input.ClinicID,
			AppointmentID: appt.ID,
			CourseID:      input.TrimmingCourseID,
			StyleRequest:  input.TrimmingStyleRequest,
		}
		if err := v.trimmingDetailRepo.Create(ctx, detail); err != nil {
			return nil, apperrors.Wrap(err, "failed to create LINE trimming detail")
		}
		if len(input.TrimmingOptionIDs) == 0 {
			return appt, nil
		}
		if err := v.trimmingDetailRepo.SetOptions(ctx, input.ClinicID, appt.ID, input.TrimmingOptionIDs); err != nil {
			return nil, apperrors.Wrap(err, "failed to set LINE trimming options")
		}
	}
	return appt, nil
}
