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
	ActorID         *uint64 // 監査ログ用（永続化しない）。handler extractStaffID から渡す。
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

func isEstimateLocked(status model.EstimateStatus) bool {
	return status == model.EstimateStatusApproved || status == model.EstimateStatusRejected
}

type staffClinicMembershipCounter interface {
	CountByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (int64, error)
}

// EstimateService は見積書のビジネスロジックインターフェース
type EstimateService interface {
	List(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error)
	Create(ctx context.Context, clinicID uint64, input *CreateEstimateInput) (*model.Estimate, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateEstimateInput) (*model.Estimate, error)
	Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error
}

type estimateService struct {
	repo              repository.EstimateRepository
	medicalRecordRepo repository.MedicalRecordRepository
	reservationRepo   repository.ReservationRepository
	staffClinicRepo   staffClinicMembershipCounter
	auditService      AuditService
}

// medicalRecordRepo / reservationRepo は AUD-005 の関連 FK clinic 所有・相互整合検証用。
// staffClinicRepo は AUD-005 の created_by clinic 所属検証用。
// auditService は Create/Update/Delete の best-effort 監査（medical_record 同型。fail-closed にしない）。
func NewEstimateService(
	repo repository.EstimateRepository,
	medicalRecordRepo repository.MedicalRecordRepository,
	reservationRepo repository.ReservationRepository,
	staffClinicRepo staffClinicMembershipCounter,
	auditService AuditService,
) EstimateService {
	return &estimateService{
		repo:              repo,
		medicalRecordRepo: medicalRecordRepo,
		reservationRepo:   reservationRepo,
		staffClinicRepo:   staffClinicRepo,
		auditService:      auditService,
	}
}

// logEstimateChangeBestEffort は見積の監査ログを LogEntry で記録する（best-effort）。
// AuditService インターフェースは広げず resource="estimate" で medical_record と同型の動詞を使う。
func (s *estimateService) logEstimateChangeBestEffort(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action string,
	estimateID uint64,
	oldValue, newValue map[string]any,
) {
	if s.auditService == nil {
		return
	}
	if err := s.auditService.LogEntry(ctx, &AuditLogInput{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  auditActorTypeFor(actorID),
		Action:     action,
		Resource:   "estimate",
		ResourceID: &estimateID,
		OldValue:   oldValue,
		NewValue:   newValue,
	}); err != nil {
		slog.ErrorContext(ctx, "audit log failed for estimate "+action, "error", err, "estimate_id", estimateID)
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

func (s *estimateService) verifyCreatedByClinicMembership(ctx context.Context, clinicID, staffID uint64) error {
	if s.staffClinicRepo == nil {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
	}
	count, err := s.staffClinicRepo.CountByStaffAndClinic(ctx, staffID, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to verify staff clinic membership", "error", err, "id", staffID, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to verify staff clinic membership")
	}
	if count == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
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

	if input.CreatedBy == nil {
		return nil, apperrors.WrapInvalidInput("created_by is required")
	}
	if err := s.verifyCreatedByClinicMembership(ctx, clinicID, *input.CreatedBy); err != nil {
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
	if isEstimateLocked(estimate.Status) {
		return nil, apperrors.WrapConflict("承認済みまたは却下済みの見積書は作成できません")
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

	// 監査ログ: create（best-effort）。actor は CreatedBy。
	s.logEstimateChangeBestEffort(ctx, clinicID, input.CreatedBy, "create", created.ID, nil, extractEstimateImportantFields(created))
	return created, nil
}

func (s *estimateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateEstimateInput) (*model.Estimate, error) {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find estimate", "error", err)
		return nil, apperrors.Wrap(err, "failed to find estimate")
	}
	if isEstimateLocked(existing.Status) {
		return nil, apperrors.WrapConflict("承認済みまたは却下済みの見積書は編集できません")
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
	isBecomingApproved := input.Status != nil && *input.Status == model.EstimateStatusApproved &&
		existing.Status != model.EstimateStatusApproved
	isBecomingRejected := input.Status != nil && *input.Status == model.EstimateStatusRejected &&
		existing.Status != model.EstimateStatusRejected

	// UpdateIfNotLocked: FindByID→isEstimateLocked と Update の間に approved/rejected へ
	// 遷移されても、status NOT IN 述語で原子的に拒否する（0 行 → Conflict）。
	updated, err := s.repo.UpdateIfNotLocked(ctx, clinicID, id, fields)
	if err != nil {
		if !apperrors.IsConflict(err) {
			slog.ErrorContext(ctx, "failed to update estimate", "error", err)
		}
		return nil, apperrors.Wrap(err, "failed to update estimate")
	}
	slog.InfoContext(ctx, "estimate updated",
		slog.Uint64("estimate_id", id),
		slog.Uint64("clinic_id", clinicID))

	// 監査ログ: update / approve / reject（best-effort）。ロック拒否パスでは到達しない。
	oldDiff, newDiff := diffEstimateImportantFields(existing, updated)
	if oldDiff != nil {
		s.logEstimateChangeBestEffort(ctx, clinicID, input.ActorID, "update", id, oldDiff, newDiff)
	}
	if isBecomingApproved {
		s.logEstimateChangeBestEffort(ctx, clinicID, input.ActorID, "approve", id, nil, extractEstimateImportantFields(updated))
	}
	if isBecomingRejected {
		s.logEstimateChangeBestEffort(ctx, clinicID, input.ActorID, "reject", id, nil, extractEstimateImportantFields(updated))
	}
	return updated, nil
}

func (s *estimateService) Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find estimate")
	}
	if isEstimateLocked(existing.Status) {
		return apperrors.WrapConflict("承認済みまたは却下済みの見積書は削除できません")
	}
	// 早期 Count は UX 用。防御の本体は DeleteIfNotLocked の原子条件（status + active items=0）。
	count, err := s.repo.CountItemsByEstimateID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check estimate item dependencies", "error", err, "clinic_id", clinicID, "estimate_id", id)
		return apperrors.Wrap(err, "failed to check estimate item dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この見積書には明細が登録されているため削除できません")
	}
	oldValue := extractEstimateImportantFields(existing)
	if err := s.repo.DeleteIfNotLocked(ctx, clinicID, id); err != nil {
		if !apperrors.IsConflict(err) && !apperrors.IsNotFound(err) {
			slog.ErrorContext(ctx, "failed to delete estimate", "error", err, "clinic_id", clinicID, "estimate_id", id)
		}
		return apperrors.Wrap(err, "failed to delete estimate")
	}
	slog.InfoContext(ctx, "estimate deleted",
		slog.Uint64("estimate_id", id),
		slog.Uint64("clinic_id", clinicID))

	// 監査ログ: delete（best-effort）。actorID=nil は system actor。
	s.logEstimateChangeBestEffort(ctx, clinicID, actorID, "delete", id, oldValue, nil)
	return nil
}
