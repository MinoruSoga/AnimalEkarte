package medicalrecord

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// dischargeWithBillingInTx is the movement side of discharge (EMR-499 U1). It reads the stay's
// effective state under the row lock and refuses everything it cannot prove; it never writes the
// removed hospitalizations.status / end_date and never infers a discharge from planned dates.
//
// The discharge movement event itself (stay_movement_events.event_kind = 'discharge') is NOT written
// here: the INSERT needs an operation binding + success receipt (001 operation_success_required) which
// the discharge route does not carry, and no admit writer exists yet, so no stay can currently be
// proven "admitted". The 422 below is the stable fail-closed answer until EMR-499 U2 lands that writer.
func (s *hospitalizationService) dischargeWithBillingInTx(
	txCtx context.Context,
	clinicID, id uint64,
	input DischargeWithBillingInput,
	result *DischargeWithBillingResult,
) error {
	locked, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to lock hospitalization for discharge")
	}
	if err := s.reservationRepo.AssertOwnerInClinic(txCtx, clinicID, locked.OwnerID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization owner ownership")
	}
	if _, err := s.reservationRepo.FindPetOwnerInClinic(txCtx, clinicID, locked.PetID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization pet ownership")
	}
	state, err := s.readDischargeStayState(txCtx, clinicID, locked)
	if err != nil {
		return err
	}
	switch state.State {
	case StayCageReleased:
		if state.ReleaseKind == StayCageReleasedByDischarge {
			return apperrors.WrapInvalidInput("hospitalization is already discharged")
		}
		return apperrors.WrapInvalidInput("hospitalization stay was cancelled and cannot be discharged")
	case StayCageAdmitted:
		if state.AdmittedAt == nil {
			return unknownDischargeStay()
		}
		// Calendar-day comparison in JST: discharge_date may carry no time of day.
		admittedDay := state.AdmittedAt.In(config.JST).Format(time.DateOnly)
		if input.DischargeDate.In(config.JST).Format(time.DateOnly) < admittedDay {
			return apperrors.WrapInvalidInput("discharge_date は入院開始日以降である必要があります")
		}
		return apperrors.WrapUnprocessableWithCode(codeStayDischargeWriterUnavailable,
			"退院の記録（滞在の移動履歴への書込）は、操作受付の仕組みの整備まで利用できません")
	case StayCageReserved:
		return apperrors.WrapInvalidInput("hospitalization has not been admitted")
	default:
		return unknownDischargeStay()
	}
}

func unknownDischargeStay() error {
	return apperrors.WrapUnprocessableWithCode(codeStayLifecycleUnknown,
		"この入院の滞在状態を確認できないため退院できません")
}

// readDischargeStayState reads the effective state inside the ambient transaction. A hospitalization
// without a stay case (legacy row) is unknown; there is no backfill and no guess.
func (s *hospitalizationService) readDischargeStayState(txCtx context.Context, clinicID uint64, locked *model.Hospitalization) (StayCageEffectiveState, error) {
	if locked.StayCaseID == nil || *locked.StayCaseID == 0 {
		return StayCageEffectiveState{}, unknownDischargeStay()
	}
	reader, ok := s.cageRepo.(StayCageEffectiveStateReader)
	if !ok {
		return StayCageEffectiveState{}, apperrors.WrapInternalServerError("stay effective-state reader dependency is required")
	}
	states, err := reader.ReadEffectiveStatesForCageAssignment(txCtx, clinicID, []uint64{*locked.StayCaseID})
	if err != nil {
		return StayCageEffectiveState{}, apperrors.Wrap(err, "failed to read stay effective state")
	}
	for _, st := range states {
		if st.ClinicID == clinicID && st.StayCaseID == *locked.StayCaseID {
			return st, nil
		}
	}
	return StayCageEffectiveState{}, unknownDischargeStay()
}
