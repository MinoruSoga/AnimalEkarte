package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ExaminationService 検査サービスインターフェース
type ExaminationService interface {
	GetAllExaminations(ctx context.Context) ([]model.Examination, error)
	GetExaminationByID(ctx context.Context, id string) (*model.Examination, error)
	GetExaminationsByPetID(ctx context.Context, petID string) ([]model.Examination, error)
	GetExaminationsByOwnerID(ctx context.Context, ownerID string) ([]model.Examination, error)
	GetExaminationsByStatus(ctx context.Context, status string) ([]model.Examination, error)
	CreateExamination(ctx context.Context, req *model.CreateExaminationRequest) (*model.Examination, error)
	UpdateExamination(ctx context.Context, id string, req *model.UpdateExaminationRequest) (*model.Examination, error)
	DeleteExamination(ctx context.Context, id string) error
}

// GetAllExaminations 全ての検査を取得
func (s *Service) GetAllExaminations(ctx context.Context) ([]model.Examination, error) {
	return s.examinationRepo.GetAllExaminations(ctx)
}

// GetExaminationByID IDで検査を取得
func (s *Service) GetExaminationByID(ctx context.Context, id string) (*model.Examination, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid examination ID format")
	}

	exam, err := s.examinationRepo.GetExaminationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if exam == nil {
		return nil, apperrors.WrapNotFound("examination with id %s not found", id)
	}

	return exam, nil
}

// GetExaminationsByPetID ペットIDで検査を取得
func (s *Service) GetExaminationsByPetID(ctx context.Context, petID string) ([]model.Examination, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.examinationRepo.GetExaminationsByPetID(ctx, uid.String())
}

// GetExaminationsByOwnerID 飼い主IDで検査を取得
func (s *Service) GetExaminationsByOwnerID(ctx context.Context, ownerID string) ([]model.Examination, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.examinationRepo.GetExaminationsByOwnerID(ctx, uid.String())
}

// GetExaminationsByStatus ステータスで検査を取得
func (s *Service) GetExaminationsByStatus(ctx context.Context, status string) ([]model.Examination, error) {
	return s.examinationRepo.GetExaminationsByStatus(ctx, status)
}

// CreateExamination 検査を作成
func (s *Service) CreateExamination(ctx context.Context, req *model.CreateExaminationRequest) (*model.Examination, error) {
	exam := &model.Examination{
		ID:              uuid.New(),
		PetID:           req.PetID,
		OwnerID:         req.OwnerID,
		DoctorID:        req.DoctorID,
		MedicalRecordID: req.MedicalRecordID,
		ExaminationDate: req.ExaminationDate,
		TestType:        req.TestType,
		Machine:         req.Machine,
		Status:          "依頼中",
		ResultSummary:   req.ResultSummary,
		Items:           req.Items,
		Notes:           req.Notes,
	}

	if err := s.examinationRepo.CreateExamination(ctx, exam); err != nil {
		return nil, err
	}

	return exam, nil
}

// UpdateExamination 検査を更新
func (s *Service) UpdateExamination(ctx context.Context, id string, req *model.UpdateExaminationRequest) (*model.Examination, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid examination ID format")
	}

	exam, err := s.examinationRepo.GetExaminationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if exam == nil {
		return nil, apperrors.WrapNotFound("examination with id %s not found", id)
	}

	// Update fields
	if req.Status != "" {
		exam.Status = req.Status
	}
	if req.ResultSummary != "" {
		exam.ResultSummary = req.ResultSummary
	}
	if req.Items != "" {
		exam.Items = req.Items
	}
	if req.Notes != "" {
		exam.Notes = req.Notes
	}
	if req.ExaminationDate != nil {
		exam.ExaminationDate = *req.ExaminationDate
	}
	if req.TestType != "" {
		exam.TestType = req.TestType
	}
	if req.Machine != "" {
		exam.Machine = req.Machine
	}

	if err := s.examinationRepo.UpdateExamination(ctx, exam); err != nil {
		return nil, err
	}

	return exam, nil
}

// DeleteExamination 検査を削除
func (s *Service) DeleteExamination(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid examination ID format")
	}

	return s.examinationRepo.DeleteExamination(ctx, uid.String())
}
