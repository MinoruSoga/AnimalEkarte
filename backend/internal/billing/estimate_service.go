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
	// ExistingItemID は既存行との一致ヒント。同一見積の item id と 1:1 で
	// 解決できない場合は更新 TX 内で invalid input として拒否される。
	// ヒントは保存 PK として扱わない（ReplaceItems は新規行として作り直す）。
	ExistingItemID        *uint64
	Name                  string
	Category              model.ItemCategory
	UnitPrice             int64
	Quantity              float64
	TaxType               model.TaxType
	TaxRate               *float64
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

const maxSafeEstimateAmount int64 = (1 << 53) - 1

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
		if input.UnitPrice > maxSafeEstimateAmount {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].unit_price exceeds the maximum safe amount", i))
		}
		if input.Quantity <= 0 || math.IsNaN(input.Quantity) || math.IsInf(input.Quantity, 0) {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].quantity must be a positive finite value", i))
		}
		if input.Quantity > 999999999.9 {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].quantity exceeds the maximum allowed value", i))
		}
		if math.Abs(input.Quantity-math.Round(input.Quantity*10)/10) > 1e-10 {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].quantity must be a multiple of 0.1", i))
		}
		if input.TaxType != "" {
			switch input.TaxType {
			case model.TaxTypeIncluded, model.TaxTypeExcluded, model.TaxTypeExempt:
			default:
				return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].tax_type is invalid", i))
			}
		}
		if input.TaxRate != nil {
			rate := *input.TaxRate
			if rate < 0 || rate > 1 || math.IsNaN(rate) || math.IsInf(rate, 0) {
				return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].tax_rate must be between 0 and 1", i))
			}
			if math.Abs(rate*100-math.Round(rate*100)) > 1e-9 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].tax_rate must have at most 2 decimal places", i))
			}
		}
		if input.DiscountRate < 0 || input.DiscountRate > 100 || math.IsNaN(input.DiscountRate) || math.IsInf(input.DiscountRate, 0) {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_rate must be between 0 and 100", i))
		}
		if input.DiscountAmount < 0 {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_amount must be 0 or greater", i))
		}
		if input.DiscountAmount > maxSafeEstimateAmount {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_amount exceeds the maximum safe amount", i))
		}
		rawProduct := float64(input.UnitPrice) * input.Quantity
		if math.IsNaN(rawProduct) || math.IsInf(rawProduct, 0) || rawProduct > float64(maxSafeEstimateAmount) {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].unit_price times quantity exceeds the maximum safe amount", i))
		}
		if float64(input.DiscountAmount) > rawProduct {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_amount exceeds the raw line amount", i))
		}
	}
	return validateEstimateItemTotals(estimateItemsFromInput(0, inputs))
}

// applyMatchedEstimateItemTaxDefaults は existing_item_id ヒントを同一見積の既存行と
// 突き合わせ、税フィールド未指定の行には保存済みの税区分・税率を引き継ぐ。
// 呼び出し元の入力は書き換えず、正規化済みのコピーを返す（後続の検証・Item 生成は
// このコピーを使う）。ヒントは一致ヒントであり保存 PK ではない。未知 ID・重複 ID・
// 自見積外の行は fail-closed で 400 とする。
func applyMatchedEstimateItemTaxDefaults(inputs []EstimateItemInput, existingItems []model.EstimateItem) ([]EstimateItemInput, error) {
	normalized := make([]EstimateItemInput, len(inputs))
	copy(normalized, inputs)
	byID := make(map[uint64]*model.EstimateItem, len(existingItems))
	for i := range existingItems {
		byID[existingItems[i].ID] = &existingItems[i]
	}
	seen := make(map[uint64]int, len(inputs))
	for i := range normalized {
		hint := normalized[i].ExistingItemID
		if hint == nil {
			continue
		}
		if prev, dup := seen[*hint]; dup {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].existing_item_id %d duplicates items[%d]", i, *hint, prev))
		}
		seen[*hint] = i
		old, ok := byID[*hint]
		if !ok {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].existing_item_id %d does not match an item of this estimate", i, *hint))
		}
		if normalized[i].TaxType == "" {
			normalized[i].TaxType = old.TaxType
		}
		if normalized[i].TaxRate == nil {
			rate := old.TaxRate
			normalized[i].TaxRate = &rate
		}
	}
	return normalized, nil
}

func validateEstimateItemTotals(items []model.EstimateItem) error {
	var subtotalTotal, taxTotal, totalAmount int64
	for i := range items {
		item := &items[i]
		itemSubtotal := max(int64(math.Round(float64(item.UnitPrice)*item.Quantity))-item.DiscountAmount, 0)
		taxAmount := item.CalculateTaxAmount()
		if taxAmount < 0 || taxAmount > maxSafeEstimateAmount {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].tax_amount exceeds the maximum safe amount", i))
		}
		rowTotal := itemSubtotal
		if item.TaxType == model.TaxTypeExcluded {
			if taxAmount > maxSafeEstimateAmount-itemSubtotal {
				return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].total_amount exceeds the maximum safe amount", i))
			}
			rowTotal += taxAmount
		}
		var err error
		subtotalTotal, err = addEstimateTotal(subtotalTotal, itemSubtotal, "subtotal", i)
		if err != nil {
			return err
		}
		taxTotal, err = addEstimateTotal(taxTotal, taxAmount, "tax_total", i)
		if err != nil {
			return err
		}
		totalAmount, err = addEstimateTotal(totalAmount, rowTotal, "total_amount", i)
		if err != nil {
			return err
		}
	}
	return nil
}

func addEstimateTotal(current, row int64, field string, itemIndex int) (int64, error) {
	if row < 0 || row > maxSafeEstimateAmount-current {
		return 0, apperrors.WrapInvalidInput(fmt.Sprintf("items[%d] would make %s exceed the maximum safe amount", itemIndex, field))
	}
	return current + row, nil
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
		// 省略時は従来 API と同じ外税 10%。新 UI は明示送信し、既存行との一致時は
		// applyMatchedEstimateItemTaxDefaults が保存済み値を引き継ぐ。
		taxType := in.TaxType
		if taxType == "" {
			taxType = model.TaxTypeExcluded
		}
		taxRate := 0.10
		if in.TaxRate != nil {
			taxRate = *in.TaxRate
		}
		items = append(items, model.EstimateItem{
			EstimateID:            estimateID,
			Name:                  strings.TrimSpace(in.Name),
			Category:              category,
			UnitPrice:             in.UnitPrice,
			Quantity:              in.Quantity,
			TaxType:               taxType,
			TaxRate:               taxRate,
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

// applyEstimateHeaderAdjustments は明細由来の総額から実効ヘッダー割引と保険適用額を
// 控除した合計を返す。subtotal / taxTotal は明細由来のままとし（行割引は明細内で処理済み
// のため二重控除しない）、割引・保険は totalAmount のみへ反映する。負の控除、明細総額を
// 超える割引、割引適用後の残額を超える保険は fail-closed の invalid input。
// 合計超過判定は discount+insurance の加算ではなく totalAmount-discount の減算で行い
// int64 オーバーフローを避ける。totalAmount が負になることはない。
func applyEstimateHeaderAdjustments(subtotal, taxTotal, totalAmount, discount, insurance int64) (int64, int64, int64, error) {
	if discount < 0 {
		return 0, 0, 0, apperrors.WrapInvalidInput("discount_amount must be 0 or greater")
	}
	if insurance < 0 {
		return 0, 0, 0, apperrors.WrapInvalidInput("insurance_amount must be 0 or greater")
	}
	if discount > totalAmount {
		return 0, 0, 0, apperrors.WrapInvalidInput("discount_amount exceeds the item total")
	}
	if insurance > totalAmount-discount {
		return 0, 0, 0, apperrors.WrapInvalidInput("insurance_amount exceeds the item total after the header discount")
	}
	return subtotal, taxTotal, totalAmount - discount - insurance, nil
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
		return nil, 0, apperrors.Wrap(err, "failed to list estimate")
	}
	return result, total, nil
}

func (s *estimateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
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
		// PR665: ヘッダー割引・保険適用額は数値契約として合計から控除する（ subtotal / tax は総額のまま）。
		subtotal, taxTotal, totalAmount := calculateEstimateTotals(preparedItems)
		subtotal, taxTotal, totalAmount, adjustErr := applyEstimateHeaderAdjustments(subtotal, taxTotal, totalAmount, input.DiscountAmount, input.InsuranceAmount)
		if adjustErr != nil {
			return nil, adjustErr
		}
		estimate.Subtotal, estimate.TaxTotal, estimate.TotalAmount = subtotal, taxTotal, totalAmount
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
		got, err := s.createEstimateInTx(txCtx, clinicID, input, estimate, preparedItems)
		if err != nil {
			return err
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
	var existing, updated *model.Estimate
	var isBecomingApproved, isBecomingRejected bool
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		result, err := s.updateEstimateInTx(txCtx, clinicID, id, input)
		if err != nil {
			return err
		}
		existing = result.existing
		updated = result.updated
		isBecomingApproved = result.isBecomingApproved
		isBecomingRejected = result.isBecomingRejected
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
			return apperrors.Wrap(err, "failed to check estimate item dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この見積書には明細が登録されているため削除できません")
		}
		if err := s.repo.DeleteIfNotLocked(txCtx, clinicID, id); err != nil {
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
