package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lib/pq"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateCarePlanItemInput はケアプランアイテム作成のサービス入力DTO
type CreateCarePlanItemInput struct {
	Type                  string
	Name                  string
	Description           string
	Timing                []string
	Status                string
	Notes                 string
	MedicineID            *uint64
	ProcedureID           *uint64
	HospitalizationPlanID *uint64
	UnitPrice             int64
	Category              string
	SortOrder             int
}

// UpdateCarePlanItemInput はケアプランアイテム更新のサービス入力DTO
type UpdateCarePlanItemInput struct {
	Type                  *string
	Name                  *string
	Description           *string
	Timing                []string // nil = not changed, empty slice = clear
	Status                *string
	Notes                 *string
	MedicineID            *uint64
	ProcedureID           *uint64
	HospitalizationPlanID *uint64
	UnitPrice             *int64
	Category              *string
	SortOrder             *int
}

// CarePlanItemService はケアプランアイテムのビジネスロジックインターフェース
type CarePlanItemService interface {
	List(ctx context.Context, hospitalizationID uint64) ([]model.CarePlanItem, error)
	Create(ctx context.Context, hospitalizationID uint64, input *CreateCarePlanItemInput) (*model.CarePlanItem, error)
	Update(ctx context.Context, hospitalizationID, itemID uint64, input *UpdateCarePlanItemInput) (*model.CarePlanItem, error)
	Delete(ctx context.Context, hospitalizationID, itemID uint64) error
}

type carePlanItemService struct {
	repo repository.CarePlanItemRepository
}

// NewCarePlanItemService は CarePlanItemService を初期化して返す
func NewCarePlanItemService(repo repository.CarePlanItemRepository) CarePlanItemService {
	return &carePlanItemService{repo: repo}
}

func (s *carePlanItemService) List(ctx context.Context, hospitalizationID uint64) ([]model.CarePlanItem, error) {
	return s.repo.ListByHospitalizationID(ctx, hospitalizationID)
}

func (s *carePlanItemService) Create(ctx context.Context, hospitalizationID uint64, input *CreateCarePlanItemInput) (*model.CarePlanItem, error) {
	planType := model.CarePlanType(input.Type)
	if err := validateCarePlanType(planType); err != nil {
		return nil, err
	}

	status := model.CarePlanStatusActive
	if input.Status != "" {
		planStatus := model.CarePlanStatus(input.Status)
		if err := validateCarePlanStatus(planStatus); err != nil {
			return nil, err
		}
		status = planStatus
	}

	item := &model.CarePlanItem{
		HospitalizationID:     hospitalizationID,
		Type:                  planType,
		Name:                  input.Name,
		Description:           input.Description,
		Timing:                pq.StringArray(input.Timing),
		Status:                status,
		Notes:                 input.Notes,
		MedicineID:            input.MedicineID,
		ProcedureID:           input.ProcedureID,
		HospitalizationPlanID: input.HospitalizationPlanID,
		UnitPrice:             input.UnitPrice,
		Category:              input.Category,
		SortOrder:             input.SortOrder,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create care plan item: %w", err)
	}

	slog.InfoContext(ctx, "care plan item created",
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("care_plan_item_id", item.ID))

	return s.repo.FindByID(ctx, item.ID)
}

func (s *carePlanItemService) Update(ctx context.Context, hospitalizationID, itemID uint64, input *UpdateCarePlanItemInput) (*model.CarePlanItem, error) {
	// Verify item belongs to this hospitalization
	existing, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get care plan item: %w", err)
	}
	if existing.HospitalizationID != hospitalizationID {
		return nil, apperrors.WrapNotFound("care_plan_item", fmt.Sprintf("%d", itemID))
	}

	if input.Type != nil {
		if err := validateCarePlanType(model.CarePlanType(*input.Type)); err != nil {
			return nil, err
		}
	}
	if input.Status != nil {
		if err := validateCarePlanStatus(model.CarePlanStatus(*input.Status)); err != nil {
			return nil, err
		}
	}

	fields := buildCarePlanItemUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	if err := s.repo.Update(ctx, itemID, fields); err != nil {
		return nil, fmt.Errorf("failed to update care plan item: %w", err)
	}

	slog.InfoContext(ctx, "care plan item updated",
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("care_plan_item_id", itemID))

	return s.repo.FindByID(ctx, itemID)
}

func (s *carePlanItemService) Delete(ctx context.Context, hospitalizationID, itemID uint64) error {
	existing, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get care plan item: %w", err)
	}
	if existing.HospitalizationID != hospitalizationID {
		return apperrors.WrapNotFound("care_plan_item", fmt.Sprintf("%d", itemID))
	}

	if err := s.repo.Delete(ctx, itemID); err != nil {
		return fmt.Errorf("failed to delete care plan item: %w", err)
	}

	slog.InfoContext(ctx, "care plan item deleted",
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("care_plan_item_id", itemID))

	return nil
}

func validateCarePlanType(t model.CarePlanType) error {
	switch t {
	case model.CarePlanTypeFood, model.CarePlanTypeMedicine, model.CarePlanTypeTreatment,
		model.CarePlanTypeInstruction, model.CarePlanTypeItem:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid care plan type: %s", t))
	}
}

func validateCarePlanStatus(s model.CarePlanStatus) error {
	switch s {
	case model.CarePlanStatusActive, model.CarePlanStatusCompleted, model.CarePlanStatusDiscontinued:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid care plan status: %s", s))
	}
}

func buildCarePlanItemUpdateFields(input *UpdateCarePlanItemInput) map[string]any {
	fields := map[string]any{}
	if input.Type != nil {
		fields["type"] = *input.Type
	}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.Timing != nil {
		fields["timing"] = pq.StringArray(input.Timing)
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	if input.MedicineID != nil {
		fields["medicine_id"] = *input.MedicineID
	}
	if input.ProcedureID != nil {
		fields["procedure_id"] = *input.ProcedureID
	}
	if input.HospitalizationPlanID != nil {
		fields["hospitalization_plan_id"] = *input.HospitalizationPlanID
	}
	if input.UnitPrice != nil {
		fields["unit_price"] = *input.UnitPrice
	}
	if input.Category != nil {
		fields["category"] = *input.Category
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
