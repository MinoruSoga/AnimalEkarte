package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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

func resolveCampaignPeriod(current *model.Campaign, input *UpdateCampaignInput) (time.Time, time.Time, bool) {
	start, end := current.StartDate, current.EndDate
	hasPeriodUpdate := input.StartDate != nil || input.EndDate != nil
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
	repo repository.CampaignRepository
}

// NewCampaignService は CampaignService を初期化して返す
func NewCampaignService(repo repository.CampaignRepository) CampaignService {
	return &campaignService{repo: repo}
}

func (s *campaignService) List(ctx context.Context, clinicID uint64) ([]model.Campaign, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list campaigns", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list campaigns")
	}
	return items, nil
}

func (s *campaignService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get campaign", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get campaign")
	}
	return result, nil
}

func (s *campaignService) Create(ctx context.Context, clinicID uint64, input *CreateCampaignInput) (*model.Campaign, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateCampaignPeriod(input.StartDate, input.EndDate); err != nil {
		return nil, err
	}
	if err := validateCampaignDiscount(input.DiscountType, input.DiscountValue); err != nil {
		return nil, err
	}
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
		TargetItems:      buildCampaignTargetItems(input.TargetItemIDs),
	}
	result, err := s.repo.Create(ctx, m)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create campaign", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create campaign")
	}
	slog.InfoContext(ctx, "campaign created", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", result.ID))
	return result, nil
}

func (s *campaignService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCampaignInput) (*model.Campaign, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	current, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get campaign", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get campaign")
	}
	if err := validateOptionalName(input.Name); err != nil {
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
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	if len(fields) > 0 {
		if _, err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
			slog.ErrorContext(ctx, "failed to update campaign", "error", err, "id", id, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to update campaign")
		}
	}
	if hasTargets {
		if err := s.repo.ReplaceTargets(ctx, id, cats, itemIDs); err != nil {
			slog.ErrorContext(ctx, "failed to replace campaign targets", "error", err, "id", id, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to replace campaign targets")
		}
	}
	slog.InfoContext(ctx, "campaign updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", id))
	updated, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get updated campaign", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get updated campaign")
	}
	return updated, nil
}

func (s *campaignService) Delete(ctx context.Context, clinicID, id uint64) error {
	// campaign は末端マスタ（段階1では他から FK 参照されない）。依存チェック不要 (P10)。
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get campaign")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete campaign", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete campaign")
	}
	slog.InfoContext(ctx, "campaign deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", id))
	return nil
}

func (s *campaignService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder campaigns", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder campaigns")
	}
	slog.InfoContext(ctx, "campaigns reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
