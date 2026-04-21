package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateExaminationInput は検査作成の入力DTO
type CreateExaminationInput struct {
	MedicalRecordID *uint64
	PetID           *uint64
	ExamTypeID      uint64
	DoctorID        *uint64
	Date            time.Time
	ResultSummary   string
	Machine         string
	Status          model.ExaminationStatus
}

// UpdateExaminationInput は検査更新のサービス入力 DTO
type UpdateExaminationInput struct {
	MedicalRecordID *uint64
	PetID           *uint64
	ExamTypeID      *uint64
	DoctorID        *uint64
	Date            *time.Time
	ResultSummary   *string
	Machine         *string
	Status          *model.ExaminationStatus
}

func buildExaminationUpdate(input UpdateExaminationInput) map[string]any {
	fields := make(map[string]any)
	if input.MedicalRecordID != nil {
		fields["medical_record_id"] = *input.MedicalRecordID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.ExamTypeID != nil {
		fields["exam_type_id"] = *input.ExamTypeID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Date != nil {
		fields["date"] = *input.Date
	}
	if input.ResultSummary != nil {
		fields["result_summary"] = *input.ResultSummary
	}
	if input.Machine != nil {
		fields["machine"] = *input.Machine
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	return fields
}
type ExaminationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	Create(ctx context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error)
	Update(ctx context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type examinationService struct {
	repo repository.ExaminationRepository
}

func NewExaminationService(repo repository.ExaminationRepository) ExaminationService {
	return &examinationService{repo: repo}
}

func (s *examinationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list examinations")
	}
	return items, total, nil
}

func (s *examinationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get examination")
	}
	return result, nil
}

func (s *examinationService) Create(ctx context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error) {
	status := input.Status
	if status == "" {
		status = model.ExaminationStatusPending
	}
	exam := &model.Examination{
		ClinicID:        clinicID,
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		ExamTypeID:      input.ExamTypeID,
		DoctorID:        input.DoctorID,
		Date:            input.Date,
		ResultSummary:   input.ResultSummary,
		Machine:         input.Machine,
		Status:          status,
	}
	if err := s.repo.Create(ctx, exam); err != nil {
		return nil, apperrors.Wrap(err, "failed to create examination")
	}
	slog.InfoContext(ctx, "examination created", slog.Uint64("clinic_id", clinicID), slog.Uint64("examination_id", exam.ID))
	return exam, nil
}

func (s *examinationService) Update(ctx context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error) {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get examination")
	}
	if existing.Status == model.ExaminationStatusConfirmed {
		return nil, apperrors.WrapInvalidInput("確定済みの検査は編集できません")
	}

	fields := buildExaminationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	exam, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update examination")
	}
	slog.InfoContext(ctx, "examination updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("examination_id", id))
	return exam, nil
}

func (s *examinationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find examination")
	}
	// FK依存チェック: 検査に紐付く検査明細が存在する場合は削除を拒否
	itemCount, err := s.repo.CountItemsByExamID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check examination item dependencies")
	}
	if itemCount > 0 {
		return apperrors.WrapConflict("検査結果が紐付いているため削除できません。先に検査結果を削除してください")
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete examination")
	}

	slog.InfoContext(ctx, "examination deleted",
		slog.Uint64("examination_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}

