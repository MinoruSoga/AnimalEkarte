package billing

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ---- EstimateService ----

// CreateEstimateInput は見積書作成のサービス入力DTO
type CreateEstimateInput struct {
	MedicalRecordID *uint64
	Title           string
	OwnerID         *uint64
	PetID           *uint64
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
	Items           []EstimateItemInput
}

// EstimateItemInput は見積明細の作成/置換入力。
type EstimateItemInput struct {
	Name                  string
	Category              model.ItemCategory
	UnitPrice             int64
	Quantity              float64
	DiscountRate          float64
	DiscountAmount        int64
	IsInsuranceApplicable bool
	SortOrder             int
}

// CreateSuccessorInput は確定見積の後継ドラフト作成入力（TASK-012 FINAL B）。
// 原見積は不変。unlock は存在しない。
type CreateSuccessorInput struct {
	Reason  string  // required, min=1 max=500（handler binding と同契約）
	Title   *string // optional override
	Comment *string // optional override
	Notes   *string // optional override
	ActorID uint64  // required（handler staff id）
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
	Items           *[]EstimateItemInput
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

func validateEstimateItemInputs(inputs []EstimateItemInput) error {
	for i, input := range inputs {
		if strings.TrimSpace(input.Name) == "" {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].name is required", i))
		}
		if input.Category != "" {
			if err := validateItemCategory(string(input.Category)); err != nil {
				return apperrors.Wrap(err, fmt.Sprintf("items[%d].category is invalid", i))
			}
		}
		if input.UnitPrice < 0 {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].unit_price must be 0 or greater", i))
		}
		if input.Quantity <= 0 || math.IsNaN(input.Quantity) || math.IsInf(input.Quantity, 0) {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].quantity must be a positive finite value", i))
		}
		if input.DiscountRate < 0 || input.DiscountRate > 100 || math.IsNaN(input.DiscountRate) || math.IsInf(input.DiscountRate, 0) {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_rate must be between 0 and 100", i))
		}
		if input.DiscountAmount < 0 {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_amount must be 0 or greater", i))
		}
	}
	return nil
}

func estimateItemsFromInput(estimateID uint64, inputs []EstimateItemInput) []model.EstimateItem {
	items := make([]model.EstimateItem, 0, len(inputs))
	for i, in := range inputs {
		category := in.Category
		if category == "" {
			category = model.ItemCategoryOther
		}
		sortOrder := in.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		items = append(items, model.EstimateItem{
			EstimateID:            estimateID,
			Name:                  strings.TrimSpace(in.Name),
			Category:              category,
			UnitPrice:             in.UnitPrice,
			Quantity:              in.Quantity,
			TaxType:               model.TaxTypeExcluded,
			TaxRate:               0.10,
			DiscountRate:          in.DiscountRate,
			DiscountAmount:        in.DiscountAmount,
			IsInsuranceApplicable: in.IsInsuranceApplicable,
			SortOrder:             sortOrder,
		})
	}
	return items
}

// calculateEstimateTotals mirrors CalculateBillingTotals for EstimateItem.
// Header insurance_amount and discount_amount remain separate existing fields.
func calculateEstimateTotals(items []model.EstimateItem) (subtotal, taxTotal, totalAmount int64) {
	var excludedTax int64
	for i := range items {
		itemSubtotal := max(int64(math.Round(float64(items[i].UnitPrice)*items[i].Quantity))-items[i].DiscountAmount, 0)
		taxAmount := items[i].CalculateTaxAmount()
		subtotal += itemSubtotal
		taxTotal += taxAmount
		if items[i].TaxType == model.TaxTypeExcluded {
			excludedTax += taxAmount
		}
	}
	totalAmount = subtotal + excludedTax
	return
}

func cloneEstimateItemsForSuccessor(estimateID uint64, items []model.EstimateItem) []model.EstimateItem {
	out := make([]model.EstimateItem, 0, len(items))
	for _, it := range items {
		out = append(out, model.EstimateItem{
			EstimateID:            estimateID,
			Name:                  it.Name,
			Category:              it.Category,
			UnitPrice:             it.UnitPrice,
			Quantity:              it.Quantity,
			TaxType:               it.TaxType,
			TaxRate:               it.TaxRate,
			DiscountRate:          it.DiscountRate,
			DiscountAmount:        it.DiscountAmount,
			IsInsuranceApplicable: it.IsInsuranceApplicable,
			ConsultationID:        it.ConsultationID,
			ProcedureID:           it.ProcedureID,
			MedicineID:            it.MedicineID,
			MerchandiseItemID:     it.MerchandiseItemID,
			SortOrder:             it.SortOrder,
		})
	}
	return out
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
	// CreateSuccessor は承認/却下済み見積の後継ドラフトを新規作成する（原行は不変・unlock 無し）。
	CreateSuccessor(ctx context.Context, clinicID, originalID uint64, input *CreateSuccessorInput) (*model.Estimate, error)
}

type estimateService struct {
	repo              EstimateRepository
	medicalRecordRepo billingMedicalRecordLocker
	reservationRepo   sharedkernel.OwnerPetLinkVerifier
	staffClinicRepo   staffClinicMembershipCounter
	auditService      billingAuditLogger
	auditTx           billingAuditTxLogger // fail-closed supersede 監査（TASK-012）。nil なら CreateSuccessor 拒否。
	transactor        Transactor
}

// estimateServiceOption は EstimateService 構築時の任意依存注入。
type estimateServiceOption func(*estimateService)

// WithEstimateAuditTx は後継ドラフト（supersede）の fail-closed 監査 logger を配線する（TASK-012）。
func WithEstimateAuditTx(auditTx billingAuditTxLogger) estimateServiceOption {
	return func(s *estimateService) {
		s.auditTx = auditTx
	}
}

// medicalRecordRepo / reservationRepo は AUD-005 の関連 FK clinic 所有・相互整合検証用。
// staffClinicRepo は AUD-005 の created_by clinic 所属検証用。
// auditService は Create/Update/Delete の best-effort 監査（medical_record 同型。fail-closed にしない）。
// auditTx は CreateSuccessor の fail-closed 監査（optional; WithEstimateAuditTx で注入）。
// transactor は BE-refactor.md X-11（確定と子書込の競合防止）のため、見積書（medical_record_id に
// 紐付く「カルテ配下データ」— docs/architecture/erd.md）の書込を LockByIDForUpdate の行ロックと
// 同一トランザクションに収める目的で注入する（SD-2 系ガード監査で発見された欠落）。
// 既存 call site 互換のため opts は可変長引数（未指定時 auditTx=nil）。
func NewEstimateService(
	repo EstimateRepository,
	medicalRecordRepo billingMedicalRecordLocker,
	reservationRepo sharedkernel.OwnerPetLinkVerifier,
	staffClinicRepo staffClinicMembershipCounter,
	auditService billingAuditLogger,
	transactor Transactor,
	opts ...estimateServiceOption,
) EstimateService {
	s := &estimateService{
		repo:              repo,
		medicalRecordRepo: medicalRecordRepo,
		reservationRepo:   reservationRepo,
		staffClinicRepo:   staffClinicRepo,
		auditService:      auditService,
		transactor:        transactor,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// logEstimateChangeBestEffort は見積の監査ログを LogEntry で記録する（best-effort）。
// billingAuditLogger インターフェースは広げず resource="estimate" で medical_record と同型の動詞を使う。
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
	if err := s.auditService.LogEntry(ctx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
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
// Owner は reservation.ValidateReservationOwnerPetLinksWithRepo（AUD-001）を再利用する。
func (s *estimateService) validateEstimateRelatedFKs(
	ctx context.Context,
	clinicID uint64,
	medicalRecordID, ownerID, petID *uint64,
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

	if ownerID != nil || petID != nil {
		if s.reservationRepo == nil {
			return notFoundOwnerOrPet(petID, ownerID)
		}
		if err := reservation.ValidateReservationOwnerPetLinksWithRepo(ctx, s.reservationRepo, clinicID, ownerID, petID); err != nil {
			return err
		}
	}

	if mr != nil {
		if err := AssertBillingLinksMatchMedicalRecord(mr, ownerID, petID); err != nil {
			return err
		}
	}
	return nil
}

func notFoundOwnerOrPet(petID, ownerID *uint64) error {
	if petID != nil {
		return apperrors.WrapNotFound("pet", fmt.Sprintf("%d", *petID))
	}
	return apperrors.WrapNotFound("owner", fmt.Sprintf("%d", *ownerID))
}

func lockDraftMedicalRecordIfPresent(
	ctx context.Context,
	repo sharedkernel.MedicalRecordLocker,
	clinicID uint64,
	medicalRecordID *uint64,
	findMsg, conflictMsg string,
) error {
	if medicalRecordID == nil {
		return nil
	}
	return sharedkernel.LockDraftMedicalRecord(ctx, repo, clinicID, *medicalRecordID, findMsg, conflictMsg)
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
	if err := validateEstimateItemInputs(input.Items); err != nil {
		return nil, err
	}

	if input.CreatedBy == nil {
		return nil, apperrors.WrapInvalidInput("created_by is required")
	}

	estimate := &model.Estimate{
		ClinicID:        clinicID,
		MedicalRecordID: input.MedicalRecordID,
		Title:           input.Title,
		OwnerID:         input.OwnerID,
		PetID:           input.PetID,
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
	preparedItems := estimateItemsFromInput(0, input.Items)
	if len(preparedItems) > 0 {
		estimate.Subtotal, estimate.TaxTotal, estimate.TotalAmount = calculateEstimateTotals(preparedItems)
	}
	if input.Status != "" {
		estimate.Status = input.Status
	} else {
		estimate.Status = model.EstimateStatusDraft
	}
	if isEstimateLocked(estimate.Status) {
		return nil, apperrors.WrapConflict("承認済みまたは却下済みの見積書は作成できません")
	}

	var created *model.Estimate
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// SD-2 + BE-refactor.md X-11: 親カルテが確定済みの場合は見積書追加を拒否。見積は
		// medical_record_id 任意（カルテに紐付かない独立見積も許容）のため、指定時のみガードする。
		if err := lockDraftMedicalRecordIfPresent(txCtx, s.medicalRecordRepo, clinicID, input.MedicalRecordID,
			"failed to find medical record", "確定済みカルテに見積書を追加できません"); err != nil {
			return err
		}
		if err := s.validateEstimateRelatedFKs(txCtx, clinicID, input.MedicalRecordID, input.OwnerID, input.PetID); err != nil {
			return err
		}
		if err := s.verifyCreatedByClinicMembership(txCtx, clinicID, *input.CreatedBy); err != nil {
			return err
		}
		// TASK-012: clinic スコープの EST-{N} を原子採番してから INSERT する。
		estimateNo, err := s.repo.AllocateNextEstimateNo(txCtx, clinicID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to allocate estimate number", "error", err)
			return apperrors.Wrap(err, "failed to allocate estimate number")
		}
		estimate.EstimateNo = estimateNo
		if err := s.repo.Create(txCtx, estimate); err != nil {
			slog.ErrorContext(txCtx, "failed to create estimate", "error", err)
			return apperrors.Wrap(err, "failed to create estimate")
		}
		if len(preparedItems) > 0 {
			if err := s.repo.ReplaceItems(txCtx, clinicID, estimate.ID, estimateItemsFromInput(estimate.ID, input.Items)); err != nil {
				return apperrors.Wrap(err, "failed to save estimate items")
			}
		}
		// 再取得は commit 前。失敗したら INSERT ごと rollback し、成功を失敗応答へ反転させない。
		got, err := s.repo.FindByID(txCtx, clinicID, estimate.ID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get estimate after create", "error", err)
			return apperrors.Wrap(err, "failed to get estimate after create")
		}
		created = got
		return nil
	}); err != nil {
		return nil, err
	}
	if created == nil {
		return nil, apperrors.WrapInternalServerError("estimate create returned empty record")
	}

	slog.InfoContext(ctx, "estimate created",
		slog.Uint64("estimate_id", created.ID),
		slog.Uint64("clinic_id", clinicID))

	// 監査ログ: create（best-effort）。actor は CreatedBy。
	s.logEstimateChangeBestEffort(ctx, clinicID, input.CreatedBy, "create", created.ID, nil, extractEstimateImportantFields(created))
	return created, nil
}

func (s *estimateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateEstimateInput) (*model.Estimate, error) {
	var fields map[string]any
	var preparedItems []model.EstimateItem

	var existing, updated *model.Estimate
	var isBecomingApproved, isBecomingRejected bool
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// All update paths first lock the editable parent. FindByID then loads active
		// items through this transaction, so header totals and item replacement share
		// one authoritative snapshot and serialization point.
		locked, err := s.repo.LockEditableByID(txCtx, clinicID, id)
		if err != nil && !apperrors.IsConflict(err) {
			slog.ErrorContext(txCtx, "failed to lock estimate for update", "error", err)
			return apperrors.Wrap(err, "failed to find estimate")
		}
		if err != nil {
			return err
		}
		existing = locked

		// Keep the prior error precedence: missing or locked estimates are rejected before
		// request validation. Validation still occurs before any write in this transaction.
		if input.Subtotal != nil && *input.Subtotal < 0 {
			return apperrors.WrapInvalidInput("subtotal must be 0 or greater")
		}
		if input.TaxTotal != nil && *input.TaxTotal < 0 {
			return apperrors.WrapInvalidInput("tax_total must be 0 or greater")
		}
		if input.TotalAmount != nil && *input.TotalAmount < 0 {
			return apperrors.WrapInvalidInput("total_amount must be 0 or greater")
		}
		if input.InsuranceAmount != nil && *input.InsuranceAmount < 0 {
			return apperrors.WrapInvalidInput("insurance_amount must be 0 or greater")
		}
		if input.DiscountAmount != nil && *input.DiscountAmount < 0 {
			return apperrors.WrapInvalidInput("discount_amount must be 0 or greater")
		}
		fields = buildEstimateUpdate(input)
		if len(fields) == 0 && input.Items == nil {
			return apperrors.WrapInvalidInput("at least one field must be provided")
		}
		if input.Items != nil {
			if err := validateEstimateItemInputs(*input.Items); err != nil {
				return err
			}
			preparedItems = estimateItemsFromInput(id, *input.Items)
			subtotal, taxTotal, totalAmount := calculateEstimateTotals(preparedItems)
			fields["subtotal"] = subtotal
			fields["tax_total"] = taxTotal
			fields["total_amount"] = totalAmount
		}

		if input.Items == nil && len(existing.Items) > 0 {
			// Active persisted items are the source of truth for estimate totals. Header-only
			// PATCHes must not allow client-supplied totals to drift from those items.
			subtotal, taxTotal, totalAmount := calculateEstimateTotals(existing.Items)
			fields["subtotal"] = subtotal
			fields["tax_total"] = taxTotal
			fields["total_amount"] = totalAmount
		}
		isBecomingApproved = input.Status != nil && *input.Status == model.EstimateStatusApproved &&
			existing.Status != model.EstimateStatusApproved
		isBecomingRejected = input.Status != nil && *input.Status == model.EstimateStatusRejected &&
			existing.Status != model.EstimateStatusRejected

		// SD-2 + BE-refactor.md X-11: 親カルテが確定済みの場合は見積書編集を拒否。
		if err := lockDraftMedicalRecordIfPresent(txCtx, s.medicalRecordRepo, clinicID, existing.MedicalRecordID,
			"failed to find medical record", "確定済みカルテの見積書は編集できません"); err != nil {
			return err
		}
		// UpdateIfNotLocked retains the status predicate as defense in depth. The parent
		// is already locked, so it cannot change between the authoritative read and write.
		got, err := s.repo.UpdateIfNotLocked(txCtx, clinicID, id, fields)
		if err != nil {
			if !apperrors.IsConflict(err) {
				slog.ErrorContext(txCtx, "failed to update estimate", "error", err)
			}
			return apperrors.Wrap(err, "failed to update estimate")
		}
		if input.Items != nil {
			if err := s.repo.ReplaceItems(txCtx, clinicID, id, preparedItems); err != nil {
				return apperrors.Wrap(err, "failed to save estimate items")
			}
			got, err = s.repo.FindByID(txCtx, clinicID, id)
			if err != nil {
				return apperrors.Wrap(err, "failed to reload estimate after item replace")
			}
		}
		updated = got
		return nil
	}); err != nil {
		return nil, err
	}
	if existing == nil || updated == nil {
		return nil, apperrors.WrapInternalServerError("estimate update returned empty record")
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
	oldValue := extractEstimateImportantFields(existing)
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// SD-2 + BE-refactor.md X-11: 親カルテが確定済みの場合は見積書削除を拒否。
		if err := lockDraftMedicalRecordIfPresent(txCtx, s.medicalRecordRepo, clinicID, existing.MedicalRecordID,
			"failed to find medical record", "確定済みカルテの見積書は削除できません"); err != nil {
			return err
		}
		// 早期 Count は UX 用。防御の本体は DeleteIfNotLocked の原子条件（status + active items=0）。
		count, err := s.repo.CountItemsByEstimateID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check estimate item dependencies", "error", err, "clinic_id", clinicID, "estimate_id", id)
			return apperrors.Wrap(err, "failed to check estimate item dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この見積書には明細が登録されているため削除できません")
		}
		if err := s.repo.DeleteIfNotLocked(txCtx, clinicID, id); err != nil {
			if !apperrors.IsConflict(err) && !apperrors.IsNotFound(err) {
				slog.ErrorContext(txCtx, "failed to delete estimate", "error", err, "clinic_id", clinicID, "estimate_id", id)
			}
			return apperrors.Wrap(err, "failed to delete estimate")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "estimate deleted",
		slog.Uint64("estimate_id", id),
		slog.Uint64("clinic_id", clinicID))

	// 監査ログ: delete（best-effort）。actorID=nil は system actor。
	s.logEstimateChangeBestEffort(ctx, clinicID, actorID, "delete", id, oldValue, nil)
	return nil
}
