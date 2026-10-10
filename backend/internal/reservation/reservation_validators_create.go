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
) (*model.Reservation, *uint64, error) {
	if err := v.repo.AcquireBookingLock(ctx, input.ClinicID); err != nil {
		return nil, nil, err
	}
	if err := v.repo.AssertLineCustomerInClinic(ctx, input.ClinicID, input.CustomerID); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to verify LINE customer ownership")
	}
	reservationType, err := v.validateReservationMasterOwnership(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if input.StaffID != 0 {
		if err := ValidateLineReservationStaffCapability(ctx, v.staffRepo, input.ClinicID, &input.StaffID, input.ReservationTypeID); err != nil {
			return nil, nil, err
		}
	}
	startDT, err := ToDateTime(input.Date, input.StartTime)
	if err != nil {
		return nil, nil, passthroughOrInvalidDateTime(err)
	}
	endDT, err := ToDateTime(input.Date, input.EndTime)
	if err != nil {
		return nil, nil, passthroughOrInvalidDateTime(err)
	}
	if err := validateTimeRange(startDT, endDT); err != nil {
		return nil, nil, err
	}
	if err := validateClinicHoliday(ctx, v.holidayFinder, input.ClinicID, startDT); err != nil {
		return nil, nil, err
	}

	var doctorIDPtr *uint64
	if input.StaffID != 0 {
		id := input.StaffID
		doctorIDPtr = &id
	}
	if err := CheckSlotConflict(ctx, v.repo, input.ClinicID, doctorIDPtr, startDT, endDT, nil); err != nil {
		if errors.Is(err, errNoDoctorsOnDuty) {
			return nil, nil, &ReservationLimitError{
				Code:         "SLOT_TAKEN",
				Message:      "本日は医師が出勤していません。別の日をお選びください。",
				RedirectStep: 4,
			}
		}
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, nil, &ReservationLimitError{
				Code:         "SLOT_TAKEN",
				Message:      "選択された時間枠は既に予約が入っています。別の時間をお選びください。",
				RedirectStep: 5,
			}
		}
		return nil, nil, err
	}
	if err := CheckReservationTypeCapacity(ctx, v.repo, v.typeRepo, input.ClinicID, input.ReservationTypeID, startDT, nil); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, nil, &ReservationLimitError{
				Code:         "SLOT_TAKEN",
				Message:      "選択された時間枠は満員です。別の時間をお選びください。",
				RedirectStep: 5,
			}
		}
		return nil, nil, err
	}

	dayStart := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, input.Date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	if err := checkCustomerReservationLimit(ctx, v.repo, input, dayStart, dayEnd,
		"DAILY_LIMIT", "1日内に予約できる件数を超えています。別の日をお選びください。",
		"failed to count daily reservations", settings.DailyLimit); err != nil {
		return nil, nil, err
	}
	monthStart := time.Date(input.Date.Year(), input.Date.Month(), 1, 0, 0, 0, 0, input.Date.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	if err := checkCustomerReservationLimit(ctx, v.repo, input, monthStart, monthEnd,
		"MONTHLY_LIMIT", "1ヶ月内に予約できる件数を超えています。別の月をお選びください。",
		"failed to count monthly reservations", settings.MonthlyLimit); err != nil {
		return nil, nil, err
	}

	confirmationNumber, err := generateConfirmationNumber(ctx, v.repo, input.ClinicID, input.Date)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to generate confirmation number")
	}
	appt := buildLineReservation(input, startDT, endDT, confirmationNumber)
	// Pre-resolved identity can change before this transaction starts. Retain
	// the owner/pet locks through insertion and reject a transferred patient.
	if err := ValidateReservationOwnerPetLinksWithRepo(ctx, v.repo, input.ClinicID, appt.OwnerID, appt.PetID); err != nil {
		return nil, nil, err
	}
	// EMR-308: LINE 予約でもペットの active tentative を同一行昇格する。
	// pet 未解決（未リンク顧客）のときは repo 側が通常 INSERT にフォールバックする。
	// Codex 指摘: 事前解決時の死亡チェック（resolveReservationOwnerPet）はロックを
	// 伴わない事前 read のため、昇格・INSERT 直前に tx 内で再検証する（tx 内の
	// FindPetByIDInClinic は SHARE lock を取り、検証〜書込の死亡遷移を閉じる）。
	if err := ValidateReservationPetNotDeceased(ctx, v.repo, input.ClinicID, appt.PetID); err != nil {
		return nil, nil, err
	}
	promoted, previousOwnerID, err := createOrPromoteTentative(ctx, v.repo, appt)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to create reservation")
	}
	appt = promoted
	if reservationType.Category == model.ReservationTypeCategoryTrimming && hasLineTrimmingDetailInput(input) {
		if v.trimmingDetailRepo == nil {
			return nil, nil, apperrors.WrapInternalServerError("trimming detail repository is required for a LINE trimming reservation")
		}
		detail := &model.AppointmentTrimmingDetail{
			ClinicID:      input.ClinicID,
			AppointmentID: appt.ID,
			CourseID:      input.TrimmingCourseID,
			StyleRequest:  input.TrimmingStyleRequest,
		}
		if err := v.trimmingDetailRepo.Create(ctx, detail); err != nil {
			return nil, nil, apperrors.Wrap(err, "failed to create LINE trimming detail")
		}
		if len(input.TrimmingOptionIDs) == 0 {
			return appt, previousOwnerID, nil
		}
		if err := v.trimmingDetailRepo.SetOptions(ctx, input.ClinicID, appt.ID, input.TrimmingOptionIDs); err != nil {
			return nil, nil, apperrors.Wrap(err, "failed to set LINE trimming options")
		}
	}
	return appt, previousOwnerID, nil
}
