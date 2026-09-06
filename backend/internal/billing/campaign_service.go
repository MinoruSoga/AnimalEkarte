package billing

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// CreateCampaignInput は割引キャンペーン作成の入力 (#81)
type CreateCampaignInput struct {
	Name             string
	StartDate        time.Time
	EndDate          time.Time
	DiscountType     model.CampaignDiscountType
	DiscountValue    float64
	IsActive         bool
	SortOrder        int
	TargetCategories []model.ItemCategory
	TargetItemIDs    []uint64
}

// UpdateCampaignInput は割引キャンペーン更新の入力（nil = 未指定）
type UpdateCampaignInput struct {
	Name             *string
	StartDate        *time.Time
	EndDate          *time.Time
	DiscountType     *model.CampaignDiscountType
	DiscountValue    *float64
	IsActive         *bool
	SortOrder        *int
	TargetCategories *[]model.ItemCategory
	TargetItemIDs    *[]uint64
}

const (
	colCampaignName          = "name"
	colCampaignStartDate     = "start_date"
	colCampaignEndDate       = "end_date"
	colCampaignDiscountType  = "discount_type"
	colCampaignDiscountValue = "discount_value"
	colCampaignIsActive      = "is_active"
	colCampaignSortOrder     = "sort_order"
)

func buildCampaignUpdate(input *UpdateCampaignInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colCampaignName] = *input.Name
	}
	if input.StartDate != nil {
		fields[colCampaignStartDate] = *input.StartDate
	}
	if input.EndDate != nil {
		fields[colCampaignEndDate] = *input.EndDate
	}
	if input.DiscountType != nil {
		fields[colCampaignDiscountType] = *input.DiscountType
	}
	if input.DiscountValue != nil {
		fields[colCampaignDiscountValue] = *input.DiscountValue
	}
	if input.IsActive != nil {
		fields[colCampaignIsActive] = *input.IsActive
	}
	if input.SortOrder != nil {
		fields[colCampaignSortOrder] = *input.SortOrder
	}
	return fields
}

func buildCampaignTargetCategories(categories []model.ItemCategory) []model.CampaignTargetCategory {
	targets := make([]model.CampaignTargetCategory, 0, len(categories))
	for _, category := range categories {
		targets = append(targets, model.CampaignTargetCategory{Category: category})
	}
	return targets
}

func buildCampaignTargetItems(itemIDs []uint64) []model.CampaignTargetItem {
	targets := make([]model.CampaignTargetItem, 0, len(itemIDs))
	for _, id := range itemIDs {
		targets = append(targets, model.CampaignTargetItem{MerchandiseItemID: id})
	}
	return targets
}

func campaignTargetCategoryValues(targets []model.CampaignTargetCategory) []model.ItemCategory {
	categories := make([]model.ItemCategory, 0, len(targets))
	for _, target := range targets {
		categories = append(categories, target.Category)
	}
	return categories
}

func campaignTargetItemIDValues(targets []model.CampaignTargetItem) []uint64 {
	itemIDs := make([]uint64, 0, len(targets))
	for _, target := range targets {
		itemIDs = append(itemIDs, target.MerchandiseItemID)
	}
	return itemIDs
}

func resolveCampaignPeriod(current *model.Campaign, input *UpdateCampaignInput) (start, end time.Time, hasPeriodUpdate bool) {
	start, end = current.StartDate, current.EndDate
	hasPeriodUpdate = input.StartDate != nil || input.EndDate != nil
	if input.StartDate != nil {
		start = *input.StartDate
	}
	if input.EndDate != nil {
		end = *input.EndDate
	}
	return start, end, hasPeriodUpdate
}

func resolveCampaignDiscount(current *model.Campaign, input *UpdateCampaignInput) (model.CampaignDiscountType, float64, bool) {
	discountType, discountValue := current.DiscountType, current.DiscountValue
	hasDiscountUpdate := input.DiscountType != nil || input.DiscountValue != nil
	if input.DiscountType != nil {
		discountType = *input.DiscountType
	}
	if input.DiscountValue != nil {
		discountValue = *input.DiscountValue
	}
	return discountType, discountValue, hasDiscountUpdate
}

func resolveCampaignTargets(current *model.Campaign, input *UpdateCampaignInput) ([]model.ItemCategory, []uint64, bool) {
	categories := campaignTargetCategoryValues(current.TargetCategories)
	itemIDs := campaignTargetItemIDValues(current.TargetItems)
	hasTargets := input.TargetCategories != nil || input.TargetItemIDs != nil
	if input.TargetCategories != nil {
		categories = *input.TargetCategories
	}
	if input.TargetItemIDs != nil {
		itemIDs = *input.TargetItemIDs
	}
	return categories, itemIDs, hasTargets
}

func validateCampaignDiscount(dt model.CampaignDiscountType, value float64) error {
	switch dt {
	case model.CampaignDiscountTypeRate:
		if value < 0 || value > 100 {
			return apperrors.WrapInvalidInput("割引率は0〜100の範囲で指定してください")
		}
	case model.CampaignDiscountTypeAmount:
		if value < 0 {
			return apperrors.WrapInvalidInput("割引額は0以上で指定してください")
		}
	default:
		return apperrors.WrapInvalidInput("割引種別が不正です（rate または amount）")
	}
	return nil
}

func validateCampaignPeriod(start, end time.Time) error {
	if end.Before(start) {
		return apperrors.WrapInvalidInput("終了日は開始日以降にしてください")
	}
	return nil
}

// CampaignService は割引キャンペーンマスタのビジネスロジックインターフェース (#81)
type CampaignService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Campaign, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error)
	Create(ctx context.Context, clinicID uint64, input *CreateCampaignInput) (*model.Campaign, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateCampaignInput) (*model.Campaign, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type campaignService struct {
	repo                CampaignRepository
	merchandiseItemRepo merchandiseItemFinder
	transactor          Transactor
}

// NewCampaignService は CampaignService を初期化して返す。
// transactor は Create/Update の対象商品検証（FOR SHARE）と write/reload を同一 transaction に
// 載せるために必須（BIL-03 / X-06 / BE-ACT-CAMPAIGN-TARGET-SERIALIZATION）。
func NewCampaignService(repo CampaignRepository, merchandiseItemRepo merchandiseItemFinder, transactor Transactor) CampaignService {
	return &campaignService{repo: repo, merchandiseItemRepo: merchandiseItemRepo, transactor: transactor}
}

// uniqueSortedUint64s deduplicates ids and returns them in ascending order so
// multi-row FOR SHARE acquisition is deadlock-safe and deterministic.
func uniqueSortedUint64s(ids []uint64) []uint64 {
	if len(ids) == 0 {
		return nil
	}
	sorted := append([]uint64(nil), ids...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	out := make([]uint64, 0, len(sorted))
	for i, id := range sorted {
		if i == 0 || id != sorted[i-1] {
			out = append(out, id)
		}
	}
	return out
}

// validateOwnedMerchandiseItemIDs は request 由来の TargetItemIDs (clinic-scoped マスタFK) の
// 所有権を検証する。別 clinic の商品IDを campaign_target_items に紐付けられると、割引マッチング
// (CalculateItemCampaignDiscount) がクロステナントで汚染される (#124/#125 同型)。
// 呼び出し側は ambient transaction を開いたうえで渡し、FindByID の FOR SHARE を commit まで保持する。
// itemIDs は uniqueSortedUint64s 済みであること（ロック順序の決定論）。
func (s *campaignService) validateOwnedMerchandiseItemIDs(ctx context.Context, clinicID uint64, itemIDs []uint64) error {
	return sharedkernel.ValidateOwnedMasterFKs(ctx, "merchandise item", clinicID, itemIDs,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.merchandiseItemRepo.FindByID(actx, cid, mid)
			return err
		})
}

func (s *campaignService) List(ctx context.Context, clinicID uint64) ([]model.Campaign, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list campaigns")
	}
	return items, nil
}

func (s *campaignService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get campaign")
	}
	return result, nil
}

func (s *campaignService) Create(ctx context.Context, clinicID uint64, input *CreateCampaignInput) (*model.Campaign, error) {
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateCampaignPeriod(input.StartDate, input.EndDate); err != nil {
		return nil, err
	}
	if err := validateCampaignDiscount(input.DiscountType, input.DiscountValue); err != nil {
		return nil, err
	}
	// Deduplicate+sort before the write so FOR SHARE order and persisted targets match.
	targetItemIDs := uniqueSortedUint64s(input.TargetItemIDs)
	m := &model.Campaign{
		ClinicID:         clinicID,
		Name:             input.Name,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
		DiscountType:     input.DiscountType,
		DiscountValue:    input.DiscountValue,
		IsActive:         input.IsActive,
		SortOrder:        input.SortOrder,
		TargetCategories: buildCampaignTargetCategories(input.TargetCategories),
		TargetItems:      buildCampaignTargetItems(targetItemIDs),
	}
	// Open the transaction BEFORE merchandise validation so share-locks cover the write.
	var result *model.Campaign
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.validateOwnedMerchandiseItemIDs(txCtx, clinicID, targetItemIDs); err != nil {
			return err
		}
		created, err := s.repo.Create(txCtx, m)
		if err != nil {
			return apperrors.Wrap(err, "failed to create campaign")
		}
		result = created
		slog.InfoContext(txCtx, "campaign created", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", created.ID))
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create campaign in transaction")
	}
	return result, nil
}

func (s *campaignService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCampaignInput) (*model.Campaign, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	current, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get campaign")
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	// 期間・割引のバリデーション（指定があれば現在値とマージして検証）
	start, end, hasPeriodUpdate := resolveCampaignPeriod(current, input)
	if hasPeriodUpdate {
		if err := validateCampaignPeriod(start, end); err != nil {
			return nil, err
		}
	}
	discountType, discountValue, hasDiscountUpdate := resolveCampaignDiscount(current, input)
	if hasDiscountUpdate {
		if err := validateCampaignDiscount(discountType, discountValue); err != nil {
			return nil, err
		}
	}

	fields := buildCampaignUpdate(input)
	cats, itemIDs, hasTargets := resolveCampaignTargets(current, input)
	if len(fields) == 0 && !hasTargets {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}
	// TargetItemIDs が指定されたときだけ商品 FK を検証する（カテゴリのみ差し替えは既存 item を再ロックしない）。
	// 検証・差し替えとも uniqueSorted 済み ID を使い、ロック順序と永続化内容を一致させる。
	var targetItemIDs []uint64
	if input.TargetItemIDs != nil {
		targetItemIDs = uniqueSortedUint64s(itemIDs)
		itemIDs = targetItemIDs
	}

	var updated *model.Campaign
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Merchandise validation runs inside the ambient tx so FOR SHARE covers ReplaceTargets.
		if input.TargetItemIDs != nil {
			if err := s.validateOwnedMerchandiseItemIDs(txCtx, clinicID, targetItemIDs); err != nil {
				return err
			}
		}
		if len(fields) > 0 {
			if _, err := s.repo.Update(txCtx, clinicID, id, *input); err != nil {
				return apperrors.Wrap(err, "failed to update campaign")
			}
		}
		if hasTargets {
			if err := s.repo.ReplaceTargets(txCtx, id, cats, itemIDs); err != nil {
				return apperrors.Wrap(err, "failed to replace campaign targets")
			}
		}
		// Reload inside the same tx so a post-commit Find failure cannot invert durable success (X-01).
		var err error
		updated, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get updated campaign")
		}
		slog.InfoContext(txCtx, "campaign updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", id))
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to update campaign in transaction")
	}
	return updated, nil
}

func (s *campaignService) Delete(ctx context.Context, clinicID, id uint64) error {
	// campaign は末端マスタ（段階1では他から FK 参照されない）。依存チェック不要 (P10)。
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get campaign")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete campaign")
	}
	slog.InfoContext(ctx, "campaign deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", id))
	return nil
}

func (s *campaignService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder campaigns")
	}
	slog.InfoContext(ctx, "campaigns reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
