package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ConfirmBillingConfirmationInput は会計医師確認の入力DTO
type ConfirmBillingConfirmationInput struct {
	ConfirmedBy uint64
	Memo        string
}

// ReturnBillingConfirmationInput は会計差し戻しの入力DTO
type ReturnBillingConfirmationInput struct {
	ReturnedBy   uint64
	ReturnReason string
	Memo         string
}

// BillingConfirmationService は会計医師確認のビジネスロジックインターフェース
type BillingConfirmationService interface {
	GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error)
	Confirm(ctx context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error)
	Return(ctx context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error)
}

type billingConfirmationService struct {
	repo repository.BillingConfirmationRepository
}

// NewBillingConfirmationService はBillingConfirmationServiceを初期化して返す
func NewBillingConfirmationService(repo repository.BillingConfirmationRepository) BillingConfirmationService {
	return &billingConfirmationService{repo: repo}
}

func (s *billingConfirmationService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
	review, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return nil, apperrors.Wrap(err, "failed to get billing review")
		}
		// 存在しない場合はpendingで新規作成
		review = &model.BillingConfirmation{
			MedicalRecordID: medicalRecordID,
			Status:          model.ConfirmationStatusPending,
		}
		if err := s.repo.Create(ctx, review); err != nil {
			return nil, apperrors.Wrap(err, "failed to create billing review")
		}
		slog.InfoContext(ctx, "billing_confirmation created",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("billing_confirmation_id", review.ID),
			slog.Uint64("medical_record_id", medicalRecordID))
	}
	return review, nil
}

func (s *billingConfirmationService) Confirm(ctx context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
	review, err := s.GetOrCreate(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get or create billing review")
	}
	if review.Status == model.ConfirmationStatusConfirmed {
		return nil, apperrors.WrapInvalidInput("billing review is already confirmed")
	}

	now := time.Now()
	fields := map[string]any{
		"status":       model.ConfirmationStatusConfirmed,
		"confirmed_by": input.ConfirmedBy,
		"confirmed_at": now,
		"memo":         input.Memo,
	}
	if err := s.repo.Update(ctx, clinicID, review.ID, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update billing review")
	}
	slog.InfoContext(ctx, "billing_confirmation confirmed",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("billing_confirmation_id", review.ID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Uint64("confirmed_by", input.ConfirmedBy))
	confirmed, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get confirmed billing review")
	}
	return confirmed, nil
}

func (s *billingConfirmationService) Return(ctx context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
	review, err := s.GetOrCreate(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get or create billing review")
	}

	now := time.Now()
	fields := map[string]any{
		"status":        model.ConfirmationStatusReturned,
		"returned_by":   input.ReturnedBy,
		"returned_at":   now,
		"return_reason": input.ReturnReason,
		"memo":          input.Memo,
	}
	if err := s.repo.Update(ctx, clinicID, review.ID, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update billing review")
	}
	slog.InfoContext(ctx, "billing_confirmation returned",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("billing_confirmation_id", review.ID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Uint64("returned_by", input.ReturnedBy))
	returned, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get returned billing review")
	}
	return returned, nil
}
