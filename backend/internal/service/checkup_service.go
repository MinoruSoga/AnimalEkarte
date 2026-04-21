package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateCheckupInput は健診記録作成の入力DTO
type CreateCheckupInput struct {
	ClinicID      uint64
	CheckupTypeID uint64
	PetID         *uint64
	Date          time.Time
	NextDate      *time.Time
	DoctorID      *uint64
	Result        string
}

// UpdateCheckupInput は健診記録更新の入力DTO
type UpdateCheckupInput struct {
	CheckupTypeID *uint64
	PetID         *uint64
	Date          *time.Time
	NextDate      *time.Time
	DoctorID      *uint64
	Result        *string
}

// ListCheckupsByClinicInput はクリニック横断一覧取得の入力DTO
type ListCheckupsByClinicInput struct {
	ClinicID      uint64
	StartDate     *string
	EndDate       *string
	NextStartDate *string
	NextEndDate   *string
}

// CheckupService は健診記録のビジネスロジックを定義するインターフェース
type CheckupService interface {
	List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error)
	Update(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error)
	Delete(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) error
}

type checkupService struct {
	repo repository.CheckupRepository
}

// NewCheckupService は CheckupService の実装を返す
func NewCheckupService(repo repository.CheckupRepository) CheckupService {
	return &checkupService{repo: repo}
}

func (s *checkupService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	result, err := s.repo.ListByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list checkups")
	}
	return result, nil
}

func (s *checkupService) ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, error) {
	slog.InfoContext(ctx, "listing checkups by clinic", slog.Uint64("clinic_id", input.ClinicID))
	result, err := s.repo.ListByClinic(ctx, input.ClinicID, repository.CheckupFilters{
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		NextStartDate: input.NextStartDate,
		NextEndDate:   input.NextEndDate,
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list checkups by clinic")
	}
	return result, nil
}

func (s *checkupService) Create(ctx context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error) {
	checkup := &model.Checkup{
		ClinicID:        input.ClinicID,
		MedicalRecordID: medicalRecordID,
		CheckupTypeID:   input.CheckupTypeID,
		PetID:           input.PetID,
		Date:            input.Date,
		NextDate:        input.NextDate,
		DoctorID:        input.DoctorID,
		Result:          input.Result,
	}
	if err := s.repo.Create(ctx, checkup); err != nil {
		return nil, apperrors.Wrap(err, "failed to create checkup")
	}
	slog.InfoContext(ctx, "checkup created",
		slog.Uint64("clinic_id", input.ClinicID),
		slog.Uint64("checkup_id", checkup.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	created, err := s.repo.FindByID(ctx, input.ClinicID, checkup.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get checkup after create")
	}
	return created, nil
}

func (s *checkupService) Update(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error) {
	// 親カルテ所属確認（clinic_id スコープ済み）
	existing, err := s.repo.FindByID(ctx, clinicID, checkupID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get checkup")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
	}
	fields := buildCheckupUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, checkupID, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update checkup")
	}
	slog.InfoContext(ctx, "checkup updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("checkup_id", checkupID),
		slog.Uint64("medical_record_id", medicalRecordID))
	updated, err := s.repo.FindByID(ctx, clinicID, checkupID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get checkup after update")
	}
	return updated, nil
}

func (s *checkupService) Delete(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, checkupID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get checkup")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
	}
	if err := s.repo.Delete(ctx, clinicID, checkupID); err != nil {
		return apperrors.Wrap(err, "failed to delete checkup")
	}
	slog.InfoContext(ctx, "checkup deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("checkup_id", checkupID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return nil
}

func buildCheckupUpdate(input *UpdateCheckupInput) map[string]any {
	fields := map[string]any{}
	if input.CheckupTypeID != nil {
		fields["checkup_type_id"] = *input.CheckupTypeID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.Date != nil {
		fields["date"] = *input.Date
	}
	if input.NextDate != nil {
		fields["next_date"] = *input.NextDate
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Result != nil {
		fields["result"] = *input.Result
	}
	return fields
}
