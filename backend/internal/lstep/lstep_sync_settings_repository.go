package lstep

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// LstepSyncSettingsRepository はクリニックごとのLステップ同期設定を管理する。
type LstepSyncSettingsRepository interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.LstepSettings, error)
	Upsert(ctx context.Context, settings *model.LstepSettings) (*model.LstepSettings, error)
}

type lstepSyncSettingsRepository struct {
	db *gorm.DB
}

func NewLstepSyncSettingsRepository(db *gorm.DB) LstepSyncSettingsRepository {
	return &lstepSyncSettingsRepository{db: db}
}

func (r *lstepSyncSettingsRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LstepSettings, error) {
	var s model.LstepSettings
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		First(&s).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_settings", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return &s, nil
}

func (r *lstepSyncSettingsRepository) Upsert(ctx context.Context, settings *model.LstepSettings) (*model.LstepSettings, error) {
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(settings.ClinicID)).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"is_sync_enabled", "sync_enabled_at", "updated_at"}),
		}).
		Create(settings).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_settings", fmt.Sprintf("clinic_id=%d", settings.ClinicID))
	}
	return r.FindByClinicID(ctx, settings.ClinicID)
}
