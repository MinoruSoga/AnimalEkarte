package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ConfirmBillingReviewInput は会計医師確認の入力DTO
type ConfirmBillingReviewInput struct {
	ConfirmedBy uint64
	Memo        string
}

// ReturnBillingReviewInput は会計差し戻しの入力DTO
type ReturnBillingReviewInput struct {
	ReturnedBy   uint64
	ReturnReason string
	Memo         string
}

// BillingReviewService は会計医師確認のビジネスロジックインターフェース
type BillingReviewService interface {
	GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingReview, error)
	Confirm(ctx context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingReviewInput) (*model.BillingReview, error)
	Return(ctx context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingReviewInput) (*model.BillingReview, error)
}

type billingReviewService struct {
	repo repository.BillingReviewRepository
}

// NewBillingReviewService はBillingReviewServiceを初期化して返す
func NewBillingReviewService(repo repository.BillingReviewRepository) BillingReviewService {
	return &billingReviewService{repo: repo}
}

func (s *billingReviewService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingReview, error) {
	review, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return nil, apperrors.Wrap(err, "failed to get billing review")
		}
		// 存在しない場合はpendingで新規作成
		review = &model.BillingReview{
			MedicalRecordID: medicalRecordID,
			Status:          model.BillingReviewStatusPending,
		}
		if err := s.repo.Create(ctx, review); err != nil {
			return nil, apperrors.Wrap(err, "failed to create billing review")
		}
		slog.InfoContext(ctx, "billing_review created",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("billing_review_id", review.ID),
			slog.Uint64("medical_record_id", medicalRecordID))
	}
	return review, nil
}

func (s *billingReviewService) Confirm(ctx context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingReviewInput) (*model.BillingReview, error) {
	review, err := s.GetOrCreate(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get or create billing review")
	}
	if review.Status == model.BillingReviewStatusConfirmed {
		return nil, apperrors.WrapInvalidInput("billing review is already confirmed")
	}

	now := time.Now()
	fields := map[string]any{
		"status":       model.BillingReviewStatusConfirmed,
		"confirmed_by": input.ConfirmedBy,
		"confirmed_at": now,
		"memo":         input.Memo,
	}
	if err := s.repo.Update(ctx, clinicID, review.ID, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update billing review")
	}
	slog.InfoContext(ctx, "billing_review confirmed",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("billing_review_id", review.ID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Uint64("confirmed_by", input.ConfirmedBy))
	confirmed, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get confirmed billing review")
	}
	return confirmed, nil
}

func (s *billingReviewService) Return(ctx context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingReviewInput) (*model.BillingReview, error) {
	review, err := s.GetOrCreate(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get or create billing review")
	}

	now := time.Now()
	fields := map[string]any{
		"status":        model.BillingReviewStatusReturned,
		"returned_by":   input.ReturnedBy,
		"returned_at":   now,
		"return_reason": input.ReturnReason,
		"memo":          input.Memo,
	}
	if err := s.repo.Update(ctx, clinicID, review.ID, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update billing review")
	}
	slog.InfoContext(ctx, "billing_review returned",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("billing_review_id", review.ID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Uint64("returned_by", input.ReturnedBy))
	returned, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get returned billing review")
	}
	return returned, nil
}
