package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	repo repository.LstepTriggerPriorityRepository
}

// NewLstepTriggerPriorityService は LstepTriggerPriorityService を初期化して返す。
func NewLstepTriggerPriorityService(repo repository.LstepTriggerPriorityRepository) LstepTriggerPriorityService {
	return &lstepTriggerPriorityService{repo: repo}
}

func (s *lstepTriggerPriorityService) GetByClinicID(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error) {
	items, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find trigger priorities", "clinic_id", clinicID, "error", err)
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
	items := make([]model.LstepTriggerPriority, 0, len(input.Items))
	for _, it := range input.Items {
		if it.Priority < 1 {
			return apperrors.WrapInvalidInput("priority must be >= 1")
		}
		items = append(items, model.LstepTriggerPriority{
			ClinicID:    clinicID,
			TriggerType: it.TriggerType,
			Priority:    it.Priority,
		})
	}
	if err := s.repo.UpsertBatch(ctx, clinicID, items); err != nil {
		slog.ErrorContext(ctx, "failed to upsert trigger priorities", "clinic_id", clinicID, "error", err)
		return apperrors.Wrap(err, "failed to upsert trigger priorities")
	}
	return nil
}

func (s *lstepTriggerPriorityService) GetPriorityFor(ctx context.Context, clinicID uint64, triggerType string) (int, error) {
	p, err := s.repo.GetPriority(ctx, clinicID, triggerType)
	if err == nil {
		return p, nil
	}
	if !apperrors.IsNotFound(err) {
		slog.ErrorContext(ctx, "failed to get trigger priority", "clinic_id", clinicID, "trigger_type", triggerType, "error", err)
		return 0, apperrors.Wrap(err, "failed to get trigger priority")
	}
	// DB 未設定: DefaultTriggerPriorities → DefaultPriorityFallback
	if def, ok := model.DefaultTriggerPriorities[triggerType]; ok {
		return def, nil
	}
	return model.DefaultPriorityFallback, nil
}
