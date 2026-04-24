package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// UpsertInquiryInput は問診 upsert の入力 DTO（nil = 未送信フィールド）
type UpsertInquiryInput struct {
	ClinicID             uint64
	MedicalRecordID      uint64
	ChiefComplaintTypeID *uint64
	ChiefComplaint       *string
	Notes                *string
}

// InquiryService は医療記録問診のビジネスロジックインターフェース
type InquiryService interface {
	Save(ctx context.Context, input UpsertInquiryInput) (*model.Inquiry, error)
}

type inquiryService struct {
	repo repository.InquiryRepository
}

// NewInquiryService は InquiryService を生成する
func NewInquiryService(repo repository.InquiryRepository) InquiryService {
	return &inquiryService{repo: repo}
}

// Save は medical_record_id に対応する問診を upsert する。
func (s *inquiryService) Save(ctx context.Context, input UpsertInquiryInput) (*model.Inquiry, error) {
	inquiry := &model.Inquiry{
		MedicalRecordID: input.MedicalRecordID,
	}
	if input.ChiefComplaintTypeID != nil {
		inquiry.ChiefComplaintTypeID = input.ChiefComplaintTypeID
	}
	if input.ChiefComplaint != nil {
		inquiry.ChiefComplaint = *input.ChiefComplaint
	}
	if input.Notes != nil {
		inquiry.Notes = *input.Notes
	}

	result, err := s.repo.SaveByMedicalRecordID(ctx, input.ClinicID, inquiry)
	if err != nil {
		slog.ErrorContext(ctx, "failed to upsert inquiry",
			slog.Uint64("medical_record_id", input.MedicalRecordID),
			slog.String("error", err.Error()))
		return nil, apperrors.Wrap(err, "failed to upsert inquiry")
	}
	slog.InfoContext(ctx, "inquiry upserted", slog.Uint64("clinic_id", input.ClinicID), slog.Uint64("medical_record_id", input.MedicalRecordID))
	return result, nil
}
