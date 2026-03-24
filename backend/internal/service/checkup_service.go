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
	List(ctx context.Context, medicalRecordID uint64) ([]model.Checkup, error)
	ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error)
	Update(ctx context.Context, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error)
	Delete(ctx context.Context, medicalRecordID, checkupID uint64) error
}

type checkupService struct {
	repo repository.CheckupRepository
}

// NewCheckupService は CheckupService の実装を返す
func NewCheckupService(repo repository.CheckupRepository) CheckupService {
	return &checkupService{repo: repo}
}

func (s *checkupService) List(ctx context.Context, medicalRecordID uint64) ([]model.Checkup, error) {
	return s.repo.ListByMedicalRecordID(ctx, medicalRecordID)
}

func (s *checkupService) ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, error) {
	slog.InfoContext(ctx, "listing checkups by clinic", slog.Uint64("clinic_id", input.ClinicID))
	return s.repo.ListByClinic(ctx, input.ClinicID, repository.CheckupFilters{
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		NextStartDate: input.NextStartDate,
		NextEndDate:   input.NextEndDate,
	})
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
		return nil, err
	}
	slog.InfoContext(ctx, "checkup created",
		slog.Uint64("checkup_id", checkup.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return s.repo.FindByID(ctx, checkup.ID)
}

func (s *checkupService) Update(ctx context.Context, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error) {
	fields := buildCheckupUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	// 親カルテ所属確認
	existing, err := s.repo.FindByID(ctx, checkupID)
	if err != nil {
		return nil, err
	}
	if existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
	}
	if err := s.repo.Update(ctx, checkupID, fields); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "checkup updated",
		slog.Uint64("checkup_id", checkupID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return s.repo.FindByID(ctx, checkupID)
}

func (s *checkupService) Delete(ctx context.Context, medicalRecordID, checkupID uint64) error {
	existing, err := s.repo.FindByID(ctx, checkupID)
	if err != nil {
		return err
	}
	if existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
	}
	if err := s.repo.Delete(ctx, checkupID); err != nil {
		return err
	}
	slog.InfoContext(ctx, "checkup deleted",
		slog.Uint64("checkup_id", checkupID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return nil
}

func buildCheckupUpdateFields(input *UpdateCheckupInput) map[string]any {
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
