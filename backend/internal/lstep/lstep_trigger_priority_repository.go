package lstep

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LstepTriggerPriorityRepository はクリニック単位のトリガー優先順位設定永続化インターフェース (Q23)。
type LstepTriggerPriorityRepository interface {
	// FindByClinicID はクリニックの優先順位設定一覧を返す。
	FindByClinicID(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error)
	// UpsertBatch はクリニックの優先順位設定を一括 Upsert する。
	UpsertBatch(ctx context.Context, clinicID uint64, items []model.LstepTriggerPriority) error
	// FindPriorityByTriggerType はクリニック・トリガー種別の優先度を返す。レコードなし時は model.ErrNotFound。
	FindPriorityByTriggerType(ctx context.Context, clinicID uint64, triggerType string) (int, error)
}

type lstepTriggerPriorityRepository struct{ db *gorm.DB }

// NewLstepTriggerPriorityRepository は LstepTriggerPriorityRepository を初期化して返す。
func NewLstepTriggerPriorityRepository(db *gorm.DB) LstepTriggerPriorityRepository {
	return &lstepTriggerPriorityRepository{db: db}
}

func (r *lstepTriggerPriorityRepository) FindByClinicID(ctx context.Context, clinicID uint64) ([]model.LstepTriggerPriority, error) {
	var items []model.LstepTriggerPriority
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("priority ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "lstep_trigger_priority", fmt.Sprintf("clinic:%d", clinicID))
	}
	return items, nil
}

func (r *lstepTriggerPriorityRepository) UpsertBatch(ctx context.Context, clinicID uint64, items []model.LstepTriggerPriority) error {
	if len(items) == 0 {
		return nil
	}
	persistedItems := make([]model.LstepTriggerPriority, len(items))
	copy(persistedItems, items)
	for i := range persistedItems {
		persistedItems[i].ClinicID = clinicID
	}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "trigger_type"}},
			DoUpdates: clause.AssignmentColumns([]string{"priority", "updated_at"}),
		}).
		Create(&persistedItems).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_trigger_priority", fmt.Sprintf("clinic:%d upsert", clinicID))
	}
	return nil
}

func (r *lstepTriggerPriorityRepository) FindPriorityByTriggerType(ctx context.Context, clinicID uint64, triggerType string) (int, error) {
	var item model.LstepTriggerPriority
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND trigger_type = ?", clinicID, triggerType).
		First(&item).Error; err != nil {
		return 0, apperrors.FromGORM(err, "lstep_trigger_priority", fmt.Sprintf("clinic:%d type:%s", clinicID, triggerType))
	}
	return item.Priority, nil
}
