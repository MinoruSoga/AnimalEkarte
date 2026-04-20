package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicSettingsRepository は締め時間設定のデータアクセスインターフェース
type ClinicSettingsRepository interface {
	Get(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
	Upsert(ctx context.Context, s *model.ClinicSettings) (*model.ClinicSettings, error)
}

type clinicSettingsRepository struct{ db *gorm.DB }

// NewClinicSettingsRepository は ClinicSettingsRepository を初期化して返す
func NewClinicSettingsRepository(db *gorm.DB) ClinicSettingsRepository {
	return &clinicSettingsRepository{db: db}
}

func (r *clinicSettingsRepository) Get(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error) {
	var s model.ClinicSettings
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).First(&s).Error
	if err != nil {
		wrapped := apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
		if errors.Is(wrapped, apperrors.ErrNotFound) {
			// レコードがなければデフォルト値を返す
			return &model.ClinicSettings{
				ClinicID:            clinicID,
				ClosingAmPmBoundary: "14:00",
				ClosingWeekdayEnd:   "18:30",
				ClosingSundayEnd:    "17:30",
			}, nil
		}
		return nil, wrapped
	}
	return &s, nil
}

func (r *clinicSettingsRepository) Upsert(ctx context.Context, s *model.ClinicSettings) (*model.ClinicSettings, error) {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"closing_am_pm_boundary",
				"closing_weekday_end",
				"closing_sunday_end",
				"closed_weekdays",
				"updated_at",
			}),
		}).
		Create(s).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", s.ClinicID))
	}
	return s, nil
}
