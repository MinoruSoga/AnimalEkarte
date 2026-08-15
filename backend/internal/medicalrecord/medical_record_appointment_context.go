package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *medicalRecordService) applyAppointmentContextForCreate(
	ctx context.Context,
	clinicID uint64,
	input *CreateMedicalRecordInput,
) error {
	if input == nil || input.AppointmentID == nil {
		return nil
	}
	if s.reservationRepo == nil {
		return apperrors.WrapInternalServerError("reservation lifecycle dependency is required to create an appointment-linked medical record")
	}
	if input.Status != nil && *input.Status == model.MedicalRecordStatusFinalized {
		if err := s.reservationRepo.PrepareForMedicalRecordFinalization(ctx, clinicID, *input.AppointmentID); err != nil {
			return apperrors.Wrap(err, "failed to prepare appointment for medical record finalization")
		}
	}
	// Always take the reservation-owned appointment row lock, even when no context field is
	// missing. The lock is held by the medical-record transaction through duplicate detection and
	// INSERT, serializing create with trimming update/delete.
	if err := s.reservationRepo.BackfillForMedicalRecord(
		ctx,
		clinicID,
		*input.AppointmentID,
		input.OwnerID,
		input.PetID,
		input.DoctorID,
	); err != nil {
		return apperrors.Wrap(err, "failed to lock and validate appointment for medical record")
	}
	appt, err := s.reservationRepo.FindByID(ctx, clinicID, *input.AppointmentID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get appointment for medical record")
	}
	if appt == nil {
		return apperrors.WrapNotFound("appointment", "")
	}

	if err := resolveAppointmentUint64("pet_id", appt.PetID, &input.PetID); err != nil {
		return err
	}
	if err := resolveAppointmentUint64("owner_id", appt.OwnerID, &input.OwnerID); err != nil {
		return err
	}
	if err := resolveAppointmentUint64("doctor_id", appt.DoctorID, &input.DoctorID); err != nil {
		return err
	}
	appointmentDate := appt.StartTime.In(config.JST)
	input.Date = time.Date(
		appointmentDate.Year(), appointmentDate.Month(), appointmentDate.Day(),
		0, 0, 0, 0, config.JST,
	)
	if input.VisitType == "" {
		input.VisitType = appt.VisitType
	}

	return nil
}

func resolveAppointmentUint64(
	field string,
	appointmentValue *uint64,
	inputValue **uint64,
) error {
	if appointmentValue != nil {
		if *inputValue != nil && **inputValue != *appointmentValue {
			return apperrors.WrapInvalidInput(field + " does not match appointment")
		}
		if *inputValue == nil {
			*inputValue = appointmentValue
		}
		return nil
	}
	return nil
}

// validateMedicalRecordOwnerPetLinks はカルテ本体の Owner/Pet clinic 所有と整合を検証する（AUD-008）。
// reservationRepo 経由で AUD-001 の validateReservationOwnerPetLinks を再利用する。
func (s *medicalRecordService) validateMedicalRecordOwnerPetLinks(
	ctx context.Context,
	clinicID uint64,
	ownerID, petID *uint64,
) error {
	if ownerID == nil && petID == nil {
		return nil
	}
	if s.reservationRepo == nil {
		return apperrors.WrapInternalServerError("reservation ownership verifier is required")
	}
	return sharedkernel.ValidateReservationOwnerPetLinks(ctx, s.reservationRepo, clinicID, ownerID, petID)
}

// medicalRecordDeceasedPetMessage はカルテ Create が死亡ペットを拒否するときの安定メッセージ（BUG-002）。
const medicalRecordDeceasedPetMessage = "死亡したペットは新規カルテを作成できません"

// medicalRecordPetByIDAdapter は FindPetByIDInClinic を sharedkernel.PetByIDFinder へ適合する。
type medicalRecordPetByIDAdapter struct {
	repo mrReservationRepo
}

func (a medicalRecordPetByIDAdapter) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return a.repo.FindPetByIDInClinic(ctx, clinicID, id)
}

// assertMedicalRecordPetNotDeceased は petID 非 nil のとき死亡ペットへの新規カルテ作成を fail-closed で拒否する（BUG-002）。
// 入院登録・会計（BUG-001）と同様、FE 選択 UI だけでは API 直叩きを防げないため BE でも検証する。
func (s *medicalRecordService) assertMedicalRecordPetNotDeceased(ctx context.Context, clinicID uint64, petID *uint64) error {
	if petID == nil {
		return nil
	}
	if s.reservationRepo == nil {
		return apperrors.WrapInternalServerError("reservation ownership verifier is required")
	}
	return sharedkernel.ValidatePetNotDeceased(
		ctx,
		medicalRecordPetByIDAdapter{repo: s.reservationRepo},
		clinicID,
		*petID,
		medicalRecordDeceasedPetMessage,
	)
}

func (s *medicalRecordService) validateMedicalRecordSnapshotOwnerPetClinicRelations(
	ctx context.Context,
	clinicID uint64,
	ownerID, petID *uint64,
) error {
	if ownerID == nil && petID == nil {
		return nil
	}
	if s.reservationRepo == nil {
		return apperrors.WrapInternalServerError("reservation ownership verifier is required")
	}
	if ownerID != nil {
		if err := s.reservationRepo.AssertOwnerInClinic(ctx, clinicID, *ownerID); err != nil {
			return apperrors.Wrap(err, "failed to verify medical record owner ownership")
		}
	}
	if petID != nil {
		if _, err := s.reservationRepo.FindPetOwnerInClinic(ctx, clinicID, *petID); err != nil {
			return apperrors.Wrap(err, "failed to verify medical record pet ownership")
		}
	}
	return nil
}

func (s *medicalRecordService) validateMedicalRecordDoctor(
	ctx context.Context,
	clinicID uint64,
	doctorID *uint64,
) error {
	if doctorID == nil {
		return nil
	}
	if s.reservationRepo == nil {
		return apperrors.WrapInternalServerError("reservation doctor validation dependency is required")
	}
	if err := s.reservationRepo.AssertMedicalRecordDoctorInClinic(ctx, clinicID, *doctorID); err != nil {
		return apperrors.Wrap(err, "failed to verify medical record doctor ownership")
	}
	return nil
}

// resolveFinalMedicalRecordOwnerPet は PATCH 入力と現在値から最終 Owner/Pet を求める（AUD-008）。
func resolveFinalMedicalRecordOwnerPet(existing *model.MedicalRecord, input UpdateMedicalRecordInput) (ownerID, petID *uint64) {
	ownerID = existing.OwnerID
	petID = existing.PetID
	if input.OwnerID != nil {
		ownerID = input.OwnerID
	}
	if input.PetID != nil {
		petID = input.PetID
	}
	return ownerID, petID
}

func resolveFinalMedicalRecordDoctor(existing *model.MedicalRecord, input UpdateMedicalRecordInput) *uint64 {
	if input.DoctorID != nil {
		return input.DoctorID
	}
	return existing.DoctorID
}

func (s *medicalRecordService) findExistingRecordByAppointment(
	ctx context.Context,
	clinicID uint64,
	record *model.MedicalRecord,
) (*model.MedicalRecord, error) {
	if record == nil || record.AppointmentID == nil {
		return nil, nil
	}
	existing, err := s.repo.FindByAppointmentID(ctx, clinicID, *record.AppointmentID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check existing medical record by appointment")
	}
	if existing != nil {
		slog.InfoContext(ctx, "medical record create skipped because appointment already has a record",
			slog.Uint64("appointment_id", *record.AppointmentID),
			slog.Uint64("record_id", existing.ID))
	}
	return existing, nil
}
