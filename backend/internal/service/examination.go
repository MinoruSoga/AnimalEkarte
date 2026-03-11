package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ExaminationService 検査サービスインターフェース
type ExaminationService interface {
	GetAllExaminationRecords(ctx context.Context) ([]model.ExaminationRecord, error)
	GetExaminationRecordByID(ctx context.Context, id string) (*model.ExaminationRecord, error)
	GetExaminationRecordsByPetID(ctx context.Context, petID string) ([]model.ExaminationRecord, error)
	CreateExaminationRecord(ctx context.Context, req *model.CreateExaminationRecordRequest) (*model.ExaminationRecord, error)
	UpdateExaminationRecord(ctx context.Context, id string, req *model.UpdateExaminationRecordRequest) (*model.ExaminationRecord, error)
	DeleteExaminationRecord(ctx context.Context, id string) error
}

// GetAllExaminationRecords 全ての検査記録を取得
func (s *Service) GetAllExaminationRecords(ctx context.Context) ([]model.ExaminationRecord, error) {
	return s.examinationRepo.GetAllExaminationRecords(ctx)
}

// GetExaminationRecordByID IDで検査記録を取得
func (s *Service) GetExaminationRecordByID(ctx context.Context, id string) (*model.ExaminationRecord, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid examination record ID format")
	}

	exam, err := s.examinationRepo.GetExaminationRecordByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if exam == nil {
		return nil, apperrors.WrapNotFound("examination record with id %s not found", id)
	}

	return exam, nil
}

// GetExaminationRecordsByPetID ペットIDで検査記録を取得
func (s *Service) GetExaminationRecordsByPetID(ctx context.Context, petID string) ([]model.ExaminationRecord, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.examinationRepo.GetExaminationRecordsByPetID(ctx, uid.String())
}

// CreateExaminationRecord 検査記録を作成
func (s *Service) CreateExaminationRecord(ctx context.Context, req *model.CreateExaminationRecordRequest) (*model.ExaminationRecord, error) {
	medicalRecordID, err := uuid.Parse(req.MedicalRecordID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid medical record ID format")
	}

	petID, err := uuid.Parse(req.PetID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD")
	}

	exam := &model.ExaminationRecord{
		ID:              uuid.New(),
		MedicalRecordID: medicalRecordID,
		PetID:           petID,
		Date:            date,
		OwnerName:       req.OwnerName,
		PetName:         req.PetName,
		TestType:        req.TestType,
		Doctor:          req.Doctor,
		Status:          req.Status,
		ResultSummary:   req.ResultSummary,
		Machine:         req.Machine,
	}

	if exam.Status == "" {
		exam.Status = "依頼中"
	}

	if err := s.examinationRepo.CreateExaminationRecord(ctx, exam); err != nil {
		return nil, err
	}

	return exam, nil
}

// UpdateExaminationRecord 検査記録を更新
func (s *Service) UpdateExaminationRecord(ctx context.Context, id string, req *model.UpdateExaminationRecordRequest) (*model.ExaminationRecord, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid examination record ID format")
	}

	exam, err := s.examinationRepo.GetExaminationRecordByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if exam == nil {
		return nil, apperrors.WrapNotFound("examination record with id %s not found", id)
	}

	if req.Date != nil && *req.Date != "" {
		t, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD")
		}
		exam.Date = t
	}
	if req.TestType != nil {
		exam.TestType = *req.TestType
	}
	if req.Doctor != nil {
		exam.Doctor = *req.Doctor
	}
	if req.Status != nil {
		exam.Status = *req.Status
	}
	if req.ResultSummary != nil {
		exam.ResultSummary = *req.ResultSummary
	}
	if req.Machine != nil {
		exam.Machine = *req.Machine
	}

	if err := s.examinationRepo.UpdateExaminationRecord(ctx, exam); err != nil {
		return nil, err
	}

	return exam, nil
}

// DeleteExaminationRecord 検査記録を削除
func (s *Service) DeleteExaminationRecord(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid examination record ID format")
	}

	return s.examinationRepo.DeleteExaminationRecord(ctx, uid.String())
}
