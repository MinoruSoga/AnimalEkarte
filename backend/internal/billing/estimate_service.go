package billing

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
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
	UnitPrice             model.ExactNumeric
	Quantity              model.ExactNumeric
	TaxType               model.TaxType
	TaxRate               *model.ExactNumeric
	DiscountRate          model.ExactNumeric
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

// maxEstimateQuantityTenths は数量の上限 999999999.9 を 0.1 単位で表した値。
const maxEstimateQuantityTenths int64 = 9999999999

// validateEstimateItemInputs は見積明細の入力を exact のまま検査する（float64 を経由しない）。
// 算定器は通さない。内税・非課税などの明細も保存でき、合計は読出し時に算定できなければ null になる。
func validateEstimateItemInputs(inputs []EstimateItemInput) error {
	for i := range inputs {
		input := &inputs[i]
		if strings.TrimSpace(input.Name) == "" {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].name is required", i))
		}
		if input.Category != "" {
			if err := validateItemCategory(string(input.Category)); err != nil {
				return apperrors.Wrap(err, fmt.Sprintf("items[%d].category is invalid", i))
			}
		}
		if err := validateEstimateItemPriceAndQuantity(i, input); err != nil {
			return err
		}
		if input.TaxType != "" {
			switch input.TaxType {
			case model.TaxTypeIncluded, model.TaxTypeExcluded, model.TaxTypeExempt:
			default:
				return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].tax_type is invalid", i))
			}
		}
		if err := validateEstimateItemRates(i, input); err != nil {
			return err
		}
		if err := validateEstimateItemDiscountAmount(i, input); err != nil {
			return err
		}
	}
	return nil
}

// validateEstimateItemPriceAndQuantity は単価（整数円）と数量（0.1 単位）を検査する。
func validateEstimateItemPriceAndQuantity(i int, input *EstimateItemInput) error {
	if cmp, ok := input.UnitPrice.Cmp(model.ExactNumericFromInt64(0)); !ok || cmp < 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].unit_price must be 0 or greater", i))
	}
	if cmp, ok := input.UnitPrice.Cmp(model.ExactNumericFromInt64(maxSafeEstimateAmount)); !ok || cmp > 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].unit_price exceeds the maximum safe amount", i))
	}
	if _, err := input.UnitPrice.Int64(); err != nil {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].unit_price must be a whole number", i))
	}
	if cmp, ok := input.Quantity.Cmp(model.ExactNumericFromInt64(0)); !ok || cmp <= 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].quantity must be a positive finite value", i))
	}
	if cmp, ok := input.Quantity.Cmp(model.ExactNumericFromScaledInt64(maxEstimateQuantityTenths, 1)); !ok || cmp > 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].quantity exceeds the maximum allowed value", i))
	}
	if _, err := input.Quantity.ScaledInt64(1); err != nil {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].quantity must be a multiple of 0.1", i))
	}
	return nil
}

// validateEstimateItemRates は税率（割合。小数 2 桁、0 以上 1 以下）と割引率（百分率。0 以上 100 以下）を検査する。
func validateEstimateItemRates(i int, input *EstimateItemInput) error {
	if input.TaxRate != nil {
		if !exactInRange(*input.TaxRate, 0, 1) {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].tax_rate must be between 0 and 1", i))
		}
		if _, err := input.TaxRate.ScaledInt64(2); err != nil {
			return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].tax_rate must have at most 2 decimal places", i))
		}
	}
	if !exactInRange(input.DiscountRate, 0, 100) {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_rate must be between 0 and 100", i))
	}
	return nil
}

// validateEstimateItemDiscountAmount は割引額（円）と、単価 × 数量（丸め前）に対する上限を検査する。
func validateEstimateItemDiscountAmount(i int, input *EstimateItemInput) error {
	if input.DiscountAmount < 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_amount must be 0 or greater", i))
	}
	if input.DiscountAmount > maxSafeEstimateAmount {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_amount exceeds the maximum safe amount", i))
	}
	unit, unitOK := calcRat(input.UnitPrice)
	quantity, quantityOK := calcRat(input.Quantity)
	if !unitOK || !quantityOK {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].unit_price times quantity cannot be evaluated", i))
	}
	rawProduct := new(big.Rat).Mul(unit, quantity)
	if rawProduct.Cmp(new(big.Rat).SetInt64(maxSafeEstimateAmount)) > 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].unit_price times quantity exceeds the maximum safe amount", i))
	}
	if new(big.Rat).SetInt64(input.DiscountAmount).Cmp(rawProduct) > 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].discount_amount exceeds the raw line amount", i))
	}
	return nil
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
		// 保存済みの税率が NULL（不明）なら引き継がず、明細生成の既定（外税 10%）に従う。
		// つまり NULL の税率は書き換え時に旧来の既定 0.10 になる（NULL は直接の DB 書込みでしか生じず、許容済み）。
		if normalized[i].TaxRate == nil && old.TaxRate != nil {
			rate := *old.TaxRate
			normalized[i].TaxRate = &rate
		}
	}
	return normalized, nil
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
		taxRate := model.ExactNumericFromScaledInt64(10, 2)
		if in.TaxRate != nil {
			taxRate = *in.TaxRate
		}
		discountRate := in.DiscountRate
		discountAmount := in.DiscountAmount
		isInsuranceApplicable := in.IsInsuranceApplicable
		items = append(items, model.EstimateItem{
			EstimateID:            estimateID,
			Name:                  strings.TrimSpace(in.Name),
			Category:              category,
			UnitPrice:             in.UnitPrice,
			Quantity:              in.Quantity,
			TaxType:               taxType,
			TaxRate:               &taxRate,
			DiscountRate:          &discountRate,
			DiscountAmount:        &discountAmount,
			IsInsuranceApplicable: &isInsuranceApplicable,
			SortOrder:             &sortOrder,
		})
	}
	return items
}

// estimateItemCalcLine は見積明細 1 行を算定入力へ変換する（EMR-508）。U4-505 の請求明細と同じ対応で、
// 値引きは discount_amount（円）だけを定額として使い、discount_rate は再適用しない（見積の旧算定も同じ）。
// 税区分・税率は NULL のまま渡し、算定器が未設定として拒否する。discount_amount が NULL（不明）の明細は
// 0 とみなさず calculation_unsupported（422）で拒否する。
func estimateItemCalcLine(item *model.EstimateItem, index int) (CalcLine, error) {
	if item.DiscountAmount == nil {
		return CalcLine{}, calcLineError(index, "値引額が未設定の明細は算定できません")
	}
	return CalcLine{
		UnitPrice:   item.UnitPrice,
		Quantity:    item.Quantity,
		LinePercent: model.ExactNumericFromInt64(0),
		LineFixed:   *item.DiscountAmount,
		TaxType:     item.TaxType,
		TaxRate:     item.TaxRate,
	}, nil
}

// calculateEstimateTotals は見積明細から合計を算定する（保存せず、読出しのたびに求める。EMR-508）。
// 見積に会計全体の値引きはない（ヘッダー割引・保険は Q6 で 0 固定）。明細が 0 件の見積は合計 0 が事実なので
// 算定器を通さず 0 を返す。算定できない明細（内税・非課税・税率未設定など）は calculation_unsupported を返し、
// 呼出し側は 0 で埋めずに合計を null にする。
func calculateEstimateTotals(items []model.EstimateItem) (CalcResult, error) {
	if len(items) == 0 {
		return CalcResult{}, nil
	}
	lines := make([]CalcLine, len(items))
	for i := range items {
		line, err := estimateItemCalcLine(&items[i], i)
		if err != nil {
			return CalcResult{}, err
		}
		lines[i] = line
	}
	return CalculateBilling(CalcInput{
		Lines:       lines,
		BillPercent: model.ExactNumericFromInt64(0),
		BillFixed:   0,
	})
}

// rejectEstimateHeaderAdjustment は見積ヘッダーの割引額・保険適用額を 0 以外では受け付けない（EMR-508 / Q6）。
// estimates の合計列は廃止され、ヘッダー割引・保険の保存先（estimate_calculation_revisions の conditions /
// insurance_confirmation のキー）が未決定のため、保存できない値を受け付けて黙って捨てない。
// 未指定（nil）と 0 は許容する。応答の insurance_amount / discount_amount は常に 0。
func rejectEstimateHeaderAdjustment(discount, insurance *int64) error {
	if (discount != nil && *discount != 0) || (insurance != nil && *insurance != 0) {
		return apperrors.WrapEstimateHeaderAdjustmentPending("見積書のヘッダー割引・保険適用額は保存先が未決定のため、0 以外を指定できません")
	}
	return nil
}

// isEstimateLocked は見積が承認済み・却下済み、または status が NULL（不明）であることを返す。
// 不明な status は書込みでロック扱いにし（fail-closed、EMR-508 / Q12）、draft として扱わない。
// SQL 側の `status NOT IN (approved, rejected)` も NULL では一致しないため、Go と SQL の向きは一致する。
func isEstimateLocked(status *model.EstimateStatus) bool {
	if status == nil {
		return true
	}
	return *status == model.EstimateStatusApproved || *status == model.EstimateStatusRejected
}

// estimateStatusIs は status が NULL でなく want と等しいことを返す。
func estimateStatusIs(status *model.EstimateStatus, want model.EstimateStatus) bool {
	return status != nil && *status == want
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
	if err := rejectEstimateHeaderAdjustment(&input.DiscountAmount, &input.InsuranceAmount); err != nil {
		return nil, err
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
		ValidUntil:      input.ValidUntil,
		Comment:         input.Comment,
		Notes:           input.Notes,
		CreatedBy:       input.CreatedBy,
	}
	preparedItems := estimateItemsFromInput(0, input.Items)
	status := model.EstimateStatusDraft
	if input.Status != "" {
		status = input.Status
	}
	estimate.Status = &status
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
		// D2/495: 旧「親カルテ確定済みなら見積書削除を拒否」(SD-2 / X-11) の判定は削除した。
		// 見積書自身の承認/却下ロックと DeleteIfNotLocked の原子条件は維持する。
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
