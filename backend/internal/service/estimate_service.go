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

// ---- EstimateService ----

// CreateEstimateInput は見積書作成のサービス入力DTO
type CreateEstimateInput struct {
	MedicalRecordID *uint64
	Title           string
	OwnerID         *uint64
	Status          model.EstimateStatus
	Subtotal        int64
	TaxTotal        int64
	TotalAmount     int64
	InsuranceAmount int64
	DiscountAmount  int64
	ValidUntil      *time.Time
	Comment         string
	Notes           string
	CreatedBy       *uint64
}

// UpdateEstimateInput は見積書更新のサービス入力DTO（nil = 未送信）
type UpdateEstimateInput struct {
	Title           *string
	Status          *model.EstimateStatus
	Subtotal        *int64
	TaxTotal        *int64
	TotalAmount     *int64
	InsuranceAmount *int64
	DiscountAmount  *int64
	ValidUntil      *time.Time
	ClearValidUntil bool
	Comment         *string
	Notes           *string
}

func buildEstimateUpdate(input *UpdateEstimateInput) map[string]any {
	fields := map[string]any{}
	if input.Title != nil {
		fields["title"] = *input.Title
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Subtotal != nil {
		fields["subtotal"] = *input.Subtotal
	}
	if input.TaxTotal != nil {
		fields["tax_total"] = *input.TaxTotal
	}
	if input.TotalAmount != nil {
		fields["total_amount"] = *input.TotalAmount
	}
	if input.InsuranceAmount != nil {
		fields["insurance_amount"] = *input.InsuranceAmount
	}
	if input.DiscountAmount != nil {
		fields["discount_amount"] = *input.DiscountAmount
	}
	if input.ClearValidUntil {
		fields["valid_until"] = nil
	} else if input.ValidUntil != nil {
		fields["valid_until"] = *input.ValidUntil
	}
	if input.Comment != nil {
		fields["comment"] = *input.Comment
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
}

// EstimateService は見積書のビジネスロジックインターフェース
type EstimateService interface {
	List(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error)
	Create(ctx context.Context, clinicID uint64, input *CreateEstimateInput) (*model.Estimate, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateEstimateInput) (*model.Estimate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type estimateService struct {
	repo              repository.EstimateRepository
	medicalRecordRepo repository.MedicalRecordRepository
	reservationRepo   repository.ReservationRepository
}

// medicalRecordRepo / reservationRepo は AUD-005 の関連 FK clinic 所有・相互整合検証用。
func NewEstimateService(
	repo repository.EstimateRepository,
	medicalRecordRepo repository.MedicalRecordRepository,
	reservationRepo repository.ReservationRepository,
) EstimateService {
	return &estimateService{
		repo:              repo,
		medicalRecordRepo: medicalRecordRepo,
		reservationRepo:   reservationRepo,
	}
}

// validateEstimateRelatedFKs は見積 Create の関連 FK（medical_record / owner）の
// clinic 所有と相互整合を検証する（AUD-005）。nil 関連は既存契約どおり許可する。
// Owner は validateReservationOwnerPetLinks（AUD-001）を再利用する。
func (s *estimateService) validateEstimateRelatedFKs(
	ctx context.Context,
	clinicID uint64,
	medicalRecordID, ownerID *uint64,
) error {
	var mr *model.MedicalRecord
	if medicalRecordID != nil {
		if s.medicalRecordRepo == nil {
			return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", *medicalRecordID))
		}
		record, err := s.medicalRecordRepo.FindByID(ctx, clinicID, *medicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to verify medical record ownership")
		}
		mr = record
	}

	if ownerID != nil {
		if s.reservationRepo == nil {
			return apperrors.WrapNotFound("owner", fmt.Sprintf("%d", *ownerID))
		}
		if err := validateReservationOwnerPetLinks(ctx, s.reservationRepo, clinicID, ownerID, nil); err != nil {
			return err
		}
	}

	if mr != nil {
		if err := assertBillingLinksMatchMedicalRecord(mr, ownerID, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *estimateService) List(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error) {
	result, total, err := s.repo.FindAll(ctx, clinicID, ownerID, medicalRecordID, status, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list estimate", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list estimate")
	}
	return result, total, nil
}

func (s *estimateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get estimate", "error", err)
		return nil, apperrors.Wrap(err, "failed to get estimate")
	}
	return result, nil
}

func (s *estimateService) Create(ctx context.Context, clinicID uint64, input *CreateEstimateInput) (*model.Estimate, error) {
	if input.Title == "" {
		return nil, apperrors.WrapInvalidInput("title is required")
	}
	if input.Subtotal < 0 {
		return nil, apperrors.WrapInvalidInput("subtotal must be 0 or greater")
	}
	if input.TaxTotal < 0 {
		return nil, apperrors.WrapInvalidInput("tax_total must be 0 or greater")
	}
	if input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput("total_amount must be 0 or greater")
	}
	if input.InsuranceAmount < 0 {
		return nil, apperrors.WrapInvalidInput("insurance_amount must be 0 or greater")
	}
	if input.DiscountAmount < 0 {
		return nil, apperrors.WrapInvalidInput("discount_amount must be 0 or greater")
	}

	if err := s.validateEstimateRelatedFKs(ctx, clinicID, input.MedicalRecordID, input.OwnerID); err != nil {
		return nil, err
	}

	estimate := &model.Estimate{
		ClinicID:        clinicID,
		MedicalRecordID: input.MedicalRecordID,
		Title:           input.Title,
		OwnerID:         input.OwnerID,
		Subtotal:        input.Subtotal,
		TaxTotal:        input.TaxTotal,
		TotalAmount:     input.TotalAmount,
		InsuranceAmount: input.InsuranceAmount,
		DiscountAmount:  input.DiscountAmount,
		ValidUntil:      input.ValidUntil,
		Comment:         input.Comment,
		Notes:           input.Notes,
		CreatedBy:       input.CreatedBy,
	}
	if input.Status != "" {
		estimate.Status = input.Status
	} else {
		estimate.Status = model.EstimateStatusDraft
	}

	if err := s.repo.Create(ctx, estimate); err != nil {
		slog.ErrorContext(ctx, "failed to create estimate", "error", err)
		return nil, apperrors.Wrap(err, "failed to create estimate")
	}
	slog.InfoContext(ctx, "estimate created",
		slog.Uint64("estimate_id", estimate.ID),
		slog.Uint64("clinic_id", clinicID))
	created, err := s.repo.FindByID(ctx, clinicID, estimate.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get estimate after create", "error", err)
		return nil, apperrors.Wrap(err, "failed to get estimate after create")
	}
	return created, nil
}

func (s *estimateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateEstimateInput) (*model.Estimate, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to find estimate", "error", err)
		return nil, apperrors.Wrap(err, "failed to find estimate")
	}
	if input.Subtotal != nil && *input.Subtotal < 0 {
		return nil, apperrors.WrapInvalidInput("subtotal must be 0 or greater")
	}
	if input.TaxTotal != nil && *input.TaxTotal < 0 {
		return nil, apperrors.WrapInvalidInput("tax_total must be 0 or greater")
	}
	if input.TotalAmount != nil && *input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput("total_amount must be 0 or greater")
	}
	if input.InsuranceAmount != nil && *input.InsuranceAmount < 0 {
		return nil, apperrors.WrapInvalidInput("insurance_amount must be 0 or greater")
	}
	if input.DiscountAmount != nil && *input.DiscountAmount < 0 {
		return nil, apperrors.WrapInvalidInput("discount_amount must be 0 or greater")
	}
	fields := buildEstimateUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update estimate", "error", err)
		return nil, apperrors.Wrap(err, "failed to update estimate")
	}
	slog.InfoContext(ctx, "estimate updated",
		slog.Uint64("estimate_id", id),
		slog.Uint64("clinic_id", clinicID))
	updated, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get estimate after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to get estimate after update")
	}
	return updated, nil
}

func (s *estimateService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find estimate")
	}
	count, err := s.repo.CountItemsByEstimateID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check estimate item dependencies", "error", err, "clinic_id", clinicID, "estimate_id", id)
		return apperrors.Wrap(err, "failed to check estimate item dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この見積書には明細が登録されているため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete estimate", "error", err, "clinic_id", clinicID, "estimate_id", id)
		return apperrors.Wrap(err, "failed to delete estimate")
	}
	slog.InfoContext(ctx, "estimate deleted",
		slog.Uint64("estimate_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
