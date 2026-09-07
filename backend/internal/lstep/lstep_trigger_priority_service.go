package lstep

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UpdateTriggerPrioritiesInput はトリガー優先順位の一括更新入力。
type UpdateTriggerPrioritiesInput struct {
	Items []TriggerPriorityItem
}

// TriggerPriorityItem は単一トリガー種別の優先順位。
type TriggerPriorityItem struct {
	TriggerType string
	Priority    int
}

// LstepTriggerPriorityService はトリガー優先順位のビジネスロジック (Q23)。
type LstepTriggerPriorityService interface {
	// GetByClinicID はクリニックの優先順位設定一覧を返す。DB 未設定項目はデフォルト値で補完する。
	GetByClinicID(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error)
	// UpdatePriorities は優先順位を一括 Upsert する。
	UpdatePriorities(ctx context.Context, clinicID uint64, input UpdateTriggerPrioritiesInput) error
	// GetPriorityFor はクリニック単位の優先度を返す。DB → DefaultTriggerPriorities → DefaultPriorityFallback の順で解決。
	GetPriorityFor(ctx context.Context, clinicID uint64, triggerType string) (int, error)
}

type lstepTriggerPriorityService struct {
	repo LstepTriggerPriorityRepository
}

// NewLstepTriggerPriorityService は LstepTriggerPriorityService を初期化して返す。
func NewLstepTriggerPriorityService(repo LstepTriggerPriorityRepository) LstepTriggerPriorityService {
	return &lstepTriggerPriorityService{repo: repo}
}

func (s *lstepTriggerPriorityService) GetByClinicID(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error) {
	items, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find trigger priorities")
	}

	// DB 設定をマップ化
	dbMap := make(map[string]int, len(items))
	for _, it := range items {
		dbMap[it.TriggerType] = it.Priority
	}

	// 全トリガー種別を補完: DB 値 → default map → fallback
	allTypes := model.AllTriggerTypes()
	result := make([]model.LstepTriggerPriority, 0, len(allTypes))
	for _, tt := range allTypes {
		p, ok := dbMap[tt]
		if !ok {
			if def, hasDefault := model.DefaultTriggerPriorities[tt]; hasDefault {
				p = def
			} else {
				p = model.DefaultPriorityFallback
			}
		}
		result = append(result, model.LstepTriggerPriority{
			ClinicID:    clinicID,
			TriggerType: tt,
			Priority:    p,
		})
	}
	return result, nil
}

func (s *lstepTriggerPriorityService) UpdatePriorities(ctx context.Context, clinicID uint64, input UpdateTriggerPrioritiesInput) error {
	allowed := make(map[string]struct{}, len(model.AllTriggerTypes()))
	for _, tt := range model.AllTriggerTypes() {
		allowed[tt] = struct{}{}
	}
	items := make([]model.LstepTriggerPriority, 0, len(input.Items))
	for _, it := range input.Items {
		if it.Priority < 1 {
			return apperrors.WrapInvalidInput("priority must be >= 1")
		}
		if _, ok := allowed[it.TriggerType]; !ok {
			// LSA-13: unknown trigger_type would persist but never affect delivery.
			return apperrors.WrapInvalidInput(fmt.Sprintf("unknown trigger_type: %s", it.TriggerType))
		}
		items = append(items, model.LstepTriggerPriority{
			ClinicID:    clinicID,
			TriggerType: it.TriggerType,
			Priority:    it.Priority,
		})
	}
	if err := s.repo.UpsertBatch(ctx, clinicID, items); err != nil {
		return apperrors.Wrap(err, "failed to upsert trigger priorities")
	}
	return nil
}

func (s *lstepTriggerPriorityService) GetPriorityFor(ctx context.Context, clinicID uint64, triggerType string) (int, error) {
	p, err := s.repo.FindPriorityByTriggerType(ctx, clinicID, triggerType)
	if err == nil {
		return p, nil
	}
	if !apperrors.IsNotFound(err) {
		return 0, apperrors.Wrap(err, "failed to get trigger priority")
	}
	// DB 未設定: DefaultTriggerPriorities → DefaultPriorityFallback
	if def, ok := model.DefaultTriggerPriorities[triggerType]; ok {
		return def, nil
	}
	return model.DefaultPriorityFallback, nil
}
