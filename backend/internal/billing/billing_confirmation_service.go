package billing

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
	repo       BillingConfirmationRepository
	medRec     billingMedicalRecordLocker
	transactor Transactor
}

// NewBillingConfirmationService はBillingConfirmationServiceを初期化して返す。transactor は
// BE-refactor.md X-11（確定と子書込の競合防止）のため、Confirm/Return の書込を LockByIDForUpdate の
// 行ロックと同一トランザクションに収める目的で注入する（SD-2 系ガード監査で発見された欠落）。
func NewBillingConfirmationService(repo BillingConfirmationRepository, medRec billingMedicalRecordLocker, transactor Transactor) BillingConfirmationService {
	return &billingConfirmationService{repo: repo, medRec: medRec, transactor: transactor}
}

func (s *billingConfirmationService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
	review, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			slog.ErrorContext(ctx, "failed to get billing review", "error", err)
			return nil, apperrors.Wrap(err, "failed to get billing review")
		}
		// テナント所有権検証（クロステナント write 防止）。
		// billing_confirmations は自前 clinic_id を持たず medical_records 経由で隔離するため、
		// 自動作成前に親カルテの所有権を明示検証する（NotFound 分岐は越境ケースと重なる）。
		if _, ownerErr := s.medRec.FindByID(ctx, clinicID, medicalRecordID); ownerErr != nil {
			if !apperrors.IsNotFound(ownerErr) {
				slog.ErrorContext(ctx, "failed to verify parent medical record", "error", ownerErr)
			}
			return nil, apperrors.Wrap(ownerErr, "failed to verify parent medical record")
		}
		// 存在しない場合はpendingで新規作成
		review = &model.BillingConfirmation{
			MedicalRecordID: medicalRecordID,
			Status:          model.ConfirmationStatusPending,
		}
		if err := s.repo.Create(ctx, review); err != nil {
			slog.ErrorContext(ctx, "failed to create billing review", "error", err)
			return nil, apperrors.Wrap(err, "failed to create billing review")
		}
	}
	return review, nil
}

func (s *billingConfirmationService) Confirm(ctx context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
	if input == nil || input.ConfirmedBy == 0 {
		return nil, apperrors.WrapUnauthorized("authenticated staff is required")
	}
	now := time.Now()
	fields := map[string]any{
		"status":       model.ConfirmationStatusConfirmed,
		"confirmed_by": input.ConfirmedBy,
		"confirmed_at": now,
		"memo":         input.Memo,
	}
	var review *model.BillingConfirmation
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// SD-2 + BE-refactor.md X-11: 親カルテが確定済みの場合は会計確認の変更を拒否。
		// LockByIDForUpdate の行ロックで finalize と直列化し、確定と同時の会計確認変更が
		// 確定済みカルテに混入する競合を防ぐ（examination_service.go Update 同型）。
		if err := sharedkernel.LockDraftMedicalRecord(txCtx, s.medRec, clinicID, medicalRecordID,
			"failed to find medical record", "確定済みカルテの会計確認は変更できません"); err != nil {
			return err
		}
		if err := s.repo.LockActiveStaffAssignment(txCtx, clinicID, input.ConfirmedBy); err != nil {
			return apperrors.Wrap(err, "failed to verify billing confirmation actor")
		}
		var err error
		review, err = s.GetOrCreate(txCtx, clinicID, medicalRecordID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get or create billing review", "error", err)
			return apperrors.Wrap(err, "failed to get or create billing review")
		}
		if review.Status == model.ConfirmationStatusConfirmed {
			return apperrors.WrapInvalidInput("billing review is already confirmed")
		}
		if err := s.repo.Update(txCtx, clinicID, review.ID, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update billing review", "error", err)
			return apperrors.Wrap(err, "failed to update billing review")
		}
		review, err = s.repo.FindByMedicalRecordID(
			txCtx,
			clinicID,
			medicalRecordID,
		)
		if err != nil {
			slog.ErrorContext(
				txCtx,
				"failed to get confirmed billing review",
				"error",
				err,
			)
			return apperrors.Wrap(
				err,
				"failed to get confirmed billing review",
			)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "billing_confirmation confirmed",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("billing_confirmation_id", review.ID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Uint64("confirmed_by", input.ConfirmedBy))
	return review, nil
}

func (s *billingConfirmationService) Return(ctx context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
	if input == nil || input.ReturnedBy == 0 {
		return nil, apperrors.WrapUnauthorized("authenticated staff is required")
	}
	now := time.Now()
	fields := map[string]any{
		"status":        model.ConfirmationStatusReturned,
		"returned_by":   input.ReturnedBy,
		"returned_at":   now,
		"return_reason": input.ReturnReason,
		"memo":          input.Memo,
	}
	var review *model.BillingConfirmation
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// SD-2 + BE-refactor.md X-11: 親カルテが確定済みの場合は会計確認の差し戻しを拒否。
		if err := sharedkernel.LockDraftMedicalRecord(txCtx, s.medRec, clinicID, medicalRecordID,
			"failed to find medical record", "確定済みカルテの会計確認は差し戻しできません"); err != nil {
			return err
		}
		if err := s.repo.LockActiveStaffAssignment(txCtx, clinicID, input.ReturnedBy); err != nil {
			return apperrors.Wrap(err, "failed to verify billing confirmation actor")
		}
		var err error
		review, err = s.GetOrCreate(txCtx, clinicID, medicalRecordID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get or create billing review", "error", err)
			return apperrors.Wrap(err, "failed to get or create billing review")
		}
		if review.Status != model.ConfirmationStatusConfirmed {
			return apperrors.WrapInvalidInput("差し戻しは確認済みの場合のみ可能です")
		}
		if err := s.repo.Update(txCtx, clinicID, review.ID, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update billing review", "error", err)
			return apperrors.Wrap(err, "failed to update billing review")
		}
		review, err = s.repo.FindByMedicalRecordID(
			txCtx,
			clinicID,
			medicalRecordID,
		)
		if err != nil {
			slog.ErrorContext(
				txCtx,
				"failed to get returned billing review",
				"error",
				err,
			)
			return apperrors.Wrap(
				err,
				"failed to get returned billing review",
			)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "billing_confirmation returned",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("billing_confirmation_id", review.ID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Uint64("returned_by", input.ReturnedBy))
	return review, nil
}
