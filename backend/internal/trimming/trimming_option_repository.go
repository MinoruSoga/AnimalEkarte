// Trimming option persistence belongs to package trimming.
package trimming

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// TrimmingOptionRepository is the trimming option persistence contract.
type TrimmingOptionRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingOptionInput) (*model.TrimmingOption, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByTrimmingOptionID(ctx context.Context, clinicID, optionID uint64) (int64, error)
}

type trimmingOptionRepository struct{ db *gorm.DB }

// NewTrimmingOptionRepository constructs a trimming option repository.
func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return &trimmingOptionRepository{db: db}
}

func (r *trimmingOptionRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	options := make([]model.TrimmingOption, 0)
	if err := persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&options).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", "")
	}
	return options, nil
}

func (r *trimmingOptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	var option model.TrimmingOption
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&option).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", id))
	}
	return &option, nil
}

func (r *trimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := option.IsActive
	wantCombinable := option.IsCombinable
	if err := db.Create(option).Error; err != nil {
		return apperrors.FromGORM(err, "trimming_option", "")
	}
	if !wantActive {
		if err := db.Model(option).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", option.ID))
		}
		option.IsActive = false
	}
	if !wantCombinable {
		if err := db.Model(option).Update("is_combinable", false).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", option.ID))
		}
		option.IsCombinable = false
	}
	return nil
}

func (r *trimmingOptionRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
	if err := r.update(ctx, clinicID, id, buildTrimmingOptionUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingOptionRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.TrimmingOption{}, "trimming_option", clinicID, id, fields)
}

func (r *trimmingOptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			First(&model.TrimmingOption{}).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", id))
		}
		result := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			Where(`NOT EXISTS (
				SELECT 1 FROM appointment_trimming_options
				JOIN appointments ON appointments.id = appointment_trimming_options.appointment_id
				  AND appointments.clinic_id = ?
				  AND appointments.deleted_at IS NULL
				WHERE appointment_trimming_options.option_id = trimming_options.id
			)`, clinicID).
			Delete(&model.TrimmingOption{})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "trimming_option", fmt.Sprintf("%d", id))
		}
		if result.RowsAffected == 0 {
			return r.normalizeTrimmingOptionDeleteMiss(persistence.WithTxValue(ctx, tx), clinicID, id)
		}
		return nil
	})
}

func (r *trimmingOptionRepository) normalizeTrimmingOptionDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("このトリミングオプションはトリミング記録で使用中のため削除できません")
}

func (r *trimmingOptionRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.TrimmingOption{}, "trimming_option", clinicID, ids, "sort_order")
}

// CountUsageByTrimmingOptionID は指定オプションを使用しているトリミングオプション数を返す（BUG-201）
// appointment_trimming_options は直接 clinic_id を持たないため appointments を JOIN してテナント分離する
func (r *trimmingOptionRepository) CountUsageByTrimmingOptionID(ctx context.Context, clinicID, optionID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.AppointmentTrimmingOption{}).
		Joins("JOIN appointments ON appointments.id = appointment_trimming_options.appointment_id AND appointments.clinic_id = ? AND appointments.deleted_at IS NULL", clinicID).
		Where("appointment_trimming_options.option_id = ?", optionID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "appointment_trimming_option", "")
	}
	return count, nil
}
