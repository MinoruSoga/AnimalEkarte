// Package service provides business logic implementations for BillingItem entity.
package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- DB column constants ---

const (
	colBillingItemUnitPrice             = "unit_price"
	colBillingItemQuantity              = "quantity"
	colBillingItemDiscountRate          = "discount_rate"
	colBillingItemDiscountAmount        = "discount_amount"
	colBillingItemTaxType               = "tax_type"
	colBillingItemTaxRate               = "tax_rate"
	colBillingItemIsInsuranceApplicable = "is_insurance_applicable"
)

// --- Input DTOs ---

// CreateBillingItemInput は billing_item 作成の入力DTO
type CreateBillingItemInput struct {
	ClinicID              uint64
	BillingID             uint64
	Category              string // Service 内で model.ItemCategory に変換
	Name                  string
	UnitPrice             int64
	Quantity              float64
	DiscountRate          float64
	DiscountAmount        int64
	TaxType               string // "" = デフォルト "excluded"
	TaxRate               float64
	IsInsuranceApplicable bool
	Source                string // "" = デフォルト "manual"
	TreatmentID           *uint64
	AppointmentID         *uint64
	TrimmingCourseID      *uint64
	TrimmingOptionID      *uint64
	MerchandiseItemID     *uint64 // #81: 個別商品指定によるキャンペーンマッチング
	SortOrder             int
}

// UpdateBillingItemInput は billing_item 更新の入力DTO（nil = 未指定）
type UpdateBillingItemInput struct {
	UnitPrice             *int64
	Quantity              *float64
	DiscountRate          *float64
	DiscountAmount        *int64
	TaxType               *model.TaxType
	TaxRate               *float64
	IsInsuranceApplicable *bool
}

func buildBillingItemUpdate(input *UpdateBillingItemInput) map[string]any {
	fields := make(map[string]any)
	if input.UnitPrice != nil {
		fields[colBillingItemUnitPrice] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields[colBillingItemQuantity] = *input.Quantity
	}
	if input.DiscountRate != nil {
		fields[colBillingItemDiscountRate] = *input.DiscountRate
	}
	if input.DiscountAmount != nil {
		fields[colBillingItemDiscountAmount] = *input.DiscountAmount
	}
	if input.TaxType != nil {
		fields[colBillingItemTaxType] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields[colBillingItemTaxRate] = *input.TaxRate
	}
	if input.IsInsuranceApplicable != nil {
		fields[colBillingItemIsInsuranceApplicable] = *input.IsInsuranceApplicable
	}
	return fields
}

// ---- BillingItemService ----

// BillingItemService は billing_items の CRUD とトータル再計算を担うインターフェース
type BillingItemService interface {
	CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error)
	DeleteItem(ctx context.Context, clinicID, id uint64) error
	GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
	// GetUngroupedSameDaySummary は同日同ペットの未会計対象化項目(診察/トリミング)の件数を返す(#77 取り残し警告)。
	GetUngroupedSameDaySummary(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error)
	// GetDiscountSuggestions は指定明細に適用可能な割引候補を返す（#81 Q-I スタッフ選択）。
	// campaignRepo 未配線の場合は飼主割引のみ。
	GetDiscountSuggestions(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error)
}

// UngroupedSameDaySummary は #77 取り残し警告用の未会計対象化件数サマリ。
type UngroupedSameDaySummary struct {
	MedicalRecordCount int64
	TrimmingCount      int64
}

type billingItemService struct {
	repo               repository.BillingItemRepository
	billingRepo        repository.AccountingRepository
	treatmentRepo      repository.TreatmentRepository
	transactor         repository.Transactor
	trimmingCourseRepo repository.TrimmingCourseRepository // X-4: クロステナント write 防止用の所有権検証
	trimmingOptionRepo repository.TrimmingOptionRepository // X-4: クロステナント write 防止用の所有権検証
	campaignRepo       repository.CampaignRepository       // #81 段階2b: nil の場合は自動割引なし
	ownerRepo          repository.OwnerRepository          // #81 段階2b: 飼主割引取得用
}

type unbilledTrimmingItemFinder interface {
	FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
}

type ungroupedTrimmingCounter interface {
	CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
}

// NewBillingItemService は BillingItemService を初期化して返す（キャンペーン自動割引なし）
func NewBillingItemService(repo repository.BillingItemRepository, billingRepo repository.AccountingRepository, treatmentRepo repository.TreatmentRepository, transactor repository.Transactor, trimmingCourseRepo repository.TrimmingCourseRepository, trimmingOptionRepo repository.TrimmingOptionRepository) BillingItemService {
	return &billingItemService{repo: repo, billingRepo: billingRepo, treatmentRepo: treatmentRepo, transactor: transactor, trimmingCourseRepo: trimmingCourseRepo, trimmingOptionRepo: trimmingOptionRepo}
}

// NewBillingItemServiceWithCampaign は #81 段階2b: キャンペーン/飼主割引の自動適用を有効にした BillingItemService を返す。
func NewBillingItemServiceWithCampaign(repo repository.BillingItemRepository, billingRepo repository.AccountingRepository, treatmentRepo repository.TreatmentRepository, transactor repository.Transactor, trimmingCourseRepo repository.TrimmingCourseRepository, trimmingOptionRepo repository.TrimmingOptionRepository, campaignRepo repository.CampaignRepository, ownerRepo repository.OwnerRepository) BillingItemService {
	return &billingItemService{repo: repo, billingRepo: billingRepo, treatmentRepo: treatmentRepo, transactor: transactor, trimmingCourseRepo: trimmingCourseRepo, trimmingOptionRepo: trimmingOptionRepo, campaignRepo: campaignRepo, ownerRepo: ownerRepo}
}

// resolveOwnerDiscountRate は会計に紐付く飼主の割引率を返す。ownerRepo 未配線または取得失敗は 0。
func (s *billingItemService) resolveOwnerDiscountRate(ctx context.Context, clinicID uint64, ownerID *uint64) float64 {
	if ownerID == nil || s.ownerRepo == nil {
		return 0
	}
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, *ownerID)
	if err != nil || owner == nil {
		return 0
	}
	return owner.DiscountRate
}

// resolveAutoDiscount は #81 段階2b: 明細に適用するキャンペーン/飼主割引額を算出する(best-effort)。
// campaignRepo 未配線時は 0。会計日(billing.ScheduledDate)・明細カテゴリ・個別商品IDで該当キャンペーンを検索し、
// 飼主割引と高い方を採用する(CalculateItemCampaignDiscount)。
func (s *billingItemService) resolveAutoDiscount(ctx context.Context, input *CreateBillingItemInput) int64 {
	if s.campaignRepo == nil {
		return 0
	}
	billing, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID)
	if err != nil || billing == nil {
		return 0
	}
	ownerRate := s.resolveOwnerDiscountRate(ctx, input.ClinicID, billing.OwnerID)
	campaign, cerr := s.campaignRepo.FindApplicableForItem(ctx, input.ClinicID, billing.ScheduledDate, model.ItemCategory(input.Category), input.MerchandiseItemID)
	if cerr != nil {
		campaign = nil // best-effort: キャンペーン検索失敗は割引なしで続行
	}
	itemSubtotal := int64(float64(input.UnitPrice) * input.Quantity)
	return CalculateItemCampaignDiscount(itemSubtotal, campaign, ownerRate)
}

func (s *billingItemService) CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if input.BillingID == 0 {
		return nil, apperrors.WrapInvalidInput("請求IDは必須です")
	}
	if input.Name == "" {
		return nil, apperrors.WrapInvalidInput("商品名は必須です")
	}
	if input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	if input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput("数量は正の値である必要があります")
	}

	// テナント所有権確認: billing が同一クリニックに属することを確認
	if _, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID); err != nil {
		slog.ErrorContext(ctx, "billing not found or belongs to different clinic", "error", err)
		return nil, apperrors.Wrap(err, "billing not found or belongs to different clinic")
	}

	// X-4: クロステナント write 防止: trimming_course_id/trimming_option_id が
	// caller の clinic に属することを検証する(#124/#125 と同型の master FK 所有権チェック)。
	if input.TrimmingCourseID != nil {
		if _, err := s.trimmingCourseRepo.FindByID(ctx, input.ClinicID, *input.TrimmingCourseID); err != nil {
			slog.ErrorContext(ctx, "trimming course not found or belongs to different clinic", "error", err)
			return nil, apperrors.Wrap(err, "failed to verify trimming course ownership")
		}
	}
	if input.TrimmingOptionID != nil {
		if _, err := s.trimmingOptionRepo.FindByID(ctx, input.ClinicID, *input.TrimmingOptionID); err != nil {
			slog.ErrorContext(ctx, "trimming option not found or belongs to different clinic", "error", err)
			return nil, apperrors.Wrap(err, "failed to verify trimming option ownership")
		}
	}

	// カテゴリバリデーション
	if err := validateItemCategory(input.Category); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate item category")
	}

	// TaxType デフォルト設定とバリデーション
	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		if err := validateTaxType(input.TaxType); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate tax type")
		}
		taxType = model.TaxType(input.TaxType)
	}

	// TaxRate デフォルト設定
	taxRate := 0.10
	if input.TaxRate > 0 {
		taxRate = input.TaxRate
	}

	// Source デフォルト設定とバリデーション
	source := model.ItemSourceManual
	if input.Source != "" {
		if err := validateItemSource(input.Source); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate item source")
		}
		source = model.ItemSource(input.Source)
	}

	item := &model.BillingItem{
		BillingID:             input.BillingID,
		Category:              model.ItemCategory(input.Category),
		Name:                  input.Name,
		UnitPrice:             input.UnitPrice,
		Quantity:              input.Quantity,
		DiscountRate:          input.DiscountRate,
		DiscountAmount:        input.DiscountAmount,
		TaxType:               taxType,
		TaxRate:               taxRate,
		IsInsuranceApplicable: input.IsInsuranceApplicable,
		Source:                source,
		TreatmentID:           input.TreatmentID,
		AppointmentID:         input.AppointmentID,
		TrimmingCourseID:      input.TrimmingCourseID,
		TrimmingOptionID:      input.TrimmingOptionID,
		SortOrder:             input.SortOrder,
	}

	// #81 段階2b: 明示的な割引指定が無ければキャンペーン/飼主割引を自動適用(best-effort)
	if item.DiscountAmount == 0 {
		item.DiscountAmount = s.resolveAutoDiscount(ctx, input)
	}

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, item); err != nil {
			slog.ErrorContext(txCtx, "failed to create billing item", "error", err)
			return apperrors.Wrap(err, "failed to create billing item")
		}

		if err := s.recalculateTotals(txCtx, input.ClinicID, input.BillingID); err != nil {
			slog.ErrorContext(txCtx, "failed to recalculate billing totals after create",
				slog.Uint64("billing_id", input.BillingID),
				slog.String("error", err.Error()),
			)
			return apperrors.Wrap(err, "failed to recalculate billing totals")
		}

		slog.InfoContext(txCtx, "billing item created",
			slog.Uint64("clinic_id", input.ClinicID),
			slog.Uint64("billing_id", input.BillingID),
			slog.Uint64("item_id", item.ID),
		)
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create billing item in transaction")
	}

	return item, nil
}

func (s *billingItemService) UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error) {
	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get billing item", "error", err)
		return nil, apperrors.Wrap(err, "failed to get billing item")
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	if input.Quantity != nil && *input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput("数量は正の値である必要があります")
	}

	fields := buildBillingItemUpdate(input)
	if len(fields) == 0 {
		return item, nil
	}

	var updated *model.BillingItem
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Update(txCtx, clinicID, id, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update billing item", "error", err)
			return apperrors.Wrap(err, "failed to update billing item")
		}

		if err := s.recalculateTotals(txCtx, clinicID, item.BillingID); err != nil {
			slog.ErrorContext(txCtx, "failed to recalculate billing totals after update",
				slog.Uint64("billing_id", item.BillingID),
				slog.String("error", err.Error()),
			)
			return apperrors.Wrap(err, "failed to recalculate billing totals")
		}

		slog.InfoContext(txCtx, "billing item updated",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("item_id", id),
			slog.Uint64("billing_id", item.BillingID),
		)
		var err error
		updated, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get billing item after update", "error", err)
			return apperrors.Wrap(err, "failed to get billing item after update")
		}
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to update billing item in transaction")
	}

	return updated, nil
}

func (s *billingItemService) DeleteItem(ctx context.Context, clinicID, id uint64) error {
	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to get billing item")
	}
	billingID := item.BillingID

	return s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete billing item", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete billing item")
		}

		if err := s.recalculateTotals(txCtx, clinicID, billingID); err != nil {
			slog.ErrorContext(txCtx, "failed to recalculate billing totals after delete",
				slog.Uint64("billing_id", billingID),
				slog.String("error", err.Error()),
			)
			return apperrors.Wrap(err, "failed to recalculate billing totals")
		}

		slog.InfoContext(txCtx, "billing item deleted",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("item_id", id),
			slog.Uint64("billing_id", billingID),
		)
		return nil
	})
}

func (s *billingItemService) GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
	treatments, err := s.treatmentRepo.FindUnbilledByPetID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find unbilled treatments", "error", err)
		return nil, apperrors.Wrap(err, "failed to find unbilled treatments")
	}
	items := make([]model.BillingItem, 0, len(treatments))
	for i := range treatments {
		items = append(items, treatmentToUnbilledBillingItem(&treatments[i]))
	}

	if trimmingFinder, ok := s.repo.(unbilledTrimmingItemFinder); ok {
		trimmingItems, err := trimmingFinder.FindUnbilledTrimmingItemsByPetID(ctx, clinicID, petID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find unbilled trimming items", "error", err)
			return nil, apperrors.Wrap(err, "failed to find unbilled trimming items")
		}
		items = append(items, trimmingItems...)
	}
	return items, nil
}

func (s *billingItemService) GetUngroupedSameDaySummary(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error) {
	mrCount, err := s.treatmentRepo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicID, petID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count ungrouped medical records", "error", err)
		return UngroupedSameDaySummary{}, apperrors.Wrap(err, "failed to count ungrouped medical records")
	}
	var trimmingCount int64
	if counter, ok := s.repo.(ungroupedTrimmingCounter); ok {
		trimmingCount, err = counter.CountNonAccountingTrimmingByPetAndDate(ctx, clinicID, petID, date)
		if err != nil {
			slog.ErrorContext(ctx, "failed to count ungrouped trimming", "error", err)
			return UngroupedSameDaySummary{}, apperrors.Wrap(err, "failed to count ungrouped trimming")
		}
	}
	return UngroupedSameDaySummary{MedicalRecordCount: mrCount, TrimmingCount: trimmingCount}, nil
}

func treatmentToUnbilledBillingItem(t *model.Treatment) model.BillingItem {
	treatmentID := t.ID
	return model.BillingItem{
		ID:                    t.ID,
		BillingID:             0,
		Category:              treatmentTypeToItemCategory(t.ItemType),
		Name:                  t.Content,
		UnitPrice:             t.UnitPrice,
		Quantity:              t.Quantity,
		TaxType:               model.TaxTypeExcluded,
		TaxRate:               0.10,
		IsInsuranceApplicable: t.IsInsurance,
		Source:                model.ItemSourceMedicalRecord,
		TreatmentID:           &treatmentID,
		SortOrder:             t.SortOrder,
	}
}

func treatmentTypeToItemCategory(t model.TreatmentItemType) model.ItemCategory {
	switch t {
	case model.TreatmentItemTypeConsultation:
		return model.ItemCategoryExamination
	case model.TreatmentItemTypeProcedure:
		return model.ItemCategoryProcedure
	case model.TreatmentItemTypeMedicine:
		return model.ItemCategoryMedicine
	default:
		return model.ItemCategoryOther
	}
}

// GetDiscountSuggestions は指定明細に適用可能な割引候補を返す (#81 Q-I スタッフ選択)。
func (s *billingItemService) GetDiscountSuggestions(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error) {
	item, err := s.repo.FindByID(ctx, clinicID, itemID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing item")
	}
	billing, err := s.billingRepo.FindByID(ctx, clinicID, item.BillingID)
	if err != nil || billing == nil {
		return nil, apperrors.Wrap(err, "failed to find billing")
	}
	ownerRate := s.resolveOwnerDiscountRate(ctx, clinicID, billing.OwnerID)
	var campaigns []*model.Campaign
	if s.campaignRepo != nil {
		campaigns, err = s.campaignRepo.FindAllApplicableForItem(ctx, clinicID, billing.ScheduledDate, item.Category, nil)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find applicable campaigns for suggestions", "error", err)
			campaigns = nil // best-effort
		}
	}
	itemSubtotal := int64(float64(item.UnitPrice) * item.Quantity)
	return BuildDiscountSuggestions(itemSubtotal, campaigns, ownerRate), nil
}

// recalculateTotals は billing の全明細から subtotal/tax_total/total_amount を再計算して保存する
func (s *billingItemService) recalculateTotals(ctx context.Context, clinicID, billingID uint64) error {
	items, err := s.repo.FindByBillingID(ctx, clinicID, billingID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find billing items")
	}
	subtotal, taxTotal, totalAmount := CalculateBillingTotals(items)
	if err := s.repo.UpdateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount); err != nil {
		return apperrors.Wrap(err, "failed to update billing totals")
	}
	return nil
}
