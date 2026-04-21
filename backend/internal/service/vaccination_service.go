package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateVaccinationInput はワクチン接種作成の入力DTO
type CreateVaccinationInput struct {
	MedicalRecordID  *uint64
	PetID            *uint64
	VaccineID        uint64
	Date             time.Time
	DoctorID         *uint64
	NextDate         *time.Time
	NextScheduleType *model.NextScheduleType
	Supplemental     string
	Lot1             string
	Lot2             string
	Lot3             string
	Lot4             string
	Remarks          string
}

// UpdateVaccinationInput はワクチン接種更新のサービス入力 DTO
type UpdateVaccinationInput struct {
	MedicalRecordID  *uint64
	PetID            *uint64
	VaccineID        *uint64
	Date             *time.Time
	DoctorID         *uint64
	NextDate         *time.Time
	NextScheduleType *model.NextScheduleType
	Supplemental     *string
	Lot1             *string
	Lot2             *string
	Lot3             *string
	Lot4             *string
	Remarks          *string
}

func buildVaccinationUpdate(input *UpdateVaccinationInput) map[string]any {
	fields := make(map[string]any)
	if input.MedicalRecordID != nil {
		fields["medical_record_id"] = *input.MedicalRecordID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.VaccineID != nil {
		fields["vaccine_id"] = *input.VaccineID
	}
	if input.Date != nil {
		fields["date"] = *input.Date
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.NextDate != nil {
		fields["next_date"] = *input.NextDate
	}
	if input.NextScheduleType != nil {
		fields["next_schedule_type"] = *input.NextScheduleType
	}
	if input.Supplemental != nil {
		fields["supplemental"] = *input.Supplemental
	}
	if input.Lot1 != nil {
		fields["lot1"] = *input.Lot1
	}
	if input.Lot2 != nil {
		fields["lot2"] = *input.Lot2
	}
	if input.Lot3 != nil {
		fields["lot3"] = *input.Lot3
	}
	if input.Lot4 != nil {
		fields["lot4"] = *input.Lot4
	}
	if input.Remarks != nil {
		fields["remarks"] = *input.Remarks
	}
	return fields
}

type VaccinationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	Create(ctx context.Context, clinicID uint64, input *CreateVaccinationInput) (*model.Vaccination, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccinationInput) (*model.Vaccination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type vaccinationService struct {
	repo repository.VaccinationRepository
}

func NewVaccinationService(repo repository.VaccinationRepository) VaccinationService {
	return &vaccinationService{repo: repo}
}

func (s *vaccinationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list vaccinations")
	}
	return items, total, nil
}

func (s *vaccinationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get vaccination")
	}
	return result, nil
}

func (s *vaccinationService) Create(ctx context.Context, clinicID uint64, input *CreateVaccinationInput) (*model.Vaccination, error) {
	if input.VaccineID == 0 {
		return nil, apperrors.WrapInvalidInput("vaccine_id is required")
	}
	vaccination := &model.Vaccination{
		ClinicID:         clinicID,
		MedicalRecordID:  input.MedicalRecordID,
		PetID:            input.PetID,
		VaccineID:        input.VaccineID,
		Date:             input.Date,
		DoctorID:         input.DoctorID,
		NextDate:         input.NextDate,
		NextScheduleType: input.NextScheduleType,
		Supplemental:     input.Supplemental,
		Lot1:             input.Lot1,
		Lot2:             input.Lot2,
		Lot3:             input.Lot3,
		Lot4:             input.Lot4,
		Remarks:          input.Remarks,
	}
	if err := s.repo.Create(ctx, vaccination); err != nil {
		return nil, apperrors.Wrap(err, "failed to create vaccination")
	}
	slog.InfoContext(ctx, "vaccination created",
		slog.Uint64("vaccination_id", vaccination.ID),
		slog.Uint64("clinic_id", vaccination.ClinicID))
	return vaccination, nil
}

func (s *vaccinationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccinationInput) (*model.Vaccination, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find vaccination")
	}
	fields := buildVaccinationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	vaccination, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update vaccination")
	}
	slog.InfoContext(ctx, "vaccination updated",
		slog.Uint64("vaccination_id", id),
		slog.Uint64("clinic_id", clinicID))
	return vaccination, nil
}
func (s *vaccinationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find vaccination")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete vaccination")
	}
	slog.InfoContext(ctx, "vaccination deleted",
		slog.Uint64("vaccination_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
