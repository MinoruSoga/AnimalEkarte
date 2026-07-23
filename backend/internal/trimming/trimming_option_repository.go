// Trimming option persistence belongs to package trimming.
package trimming

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// TrimmingOptionRepository is the trimming option persistence contract.
type TrimmingOptionRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error)
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
	if err := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&options).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", "")
	}
	return options, nil
}

func (r *trimmingOptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	var option model.TrimmingOption
	db := repohelpers.DBOrTx(ctx, r.db)
	if repohelpers.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).First(&option).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", id))
	}
	return &option, nil
}

func (r *trimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	if err := r.db.WithContext(ctx).Create(option).Error; err != nil {
		return apperrors.FromGORM(err, "trimming_option", "")
	}
	return nil
}

func (r *trimmingOptionRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.TrimmingOption{}, "trimming_option", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingOptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, repohelpers.DBOrTx(ctx, r.db), &model.TrimmingOption{}, "trimming_option", clinicID, id)
}

func (r *trimmingOptionRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.TrimmingOption{}, "trimming_option", clinicID, ids, "sort_order")
}

// CountUsageByTrimmingOptionID は指定オプションを使用しているトリミングオプション数を返す（BUG-201）
// appointment_trimming_options は直接 clinic_id を持たないため appointments を JOIN してテナント分離する
func (r *trimmingOptionRepository) CountUsageByTrimmingOptionID(ctx context.Context, clinicID, optionID uint64) (int64, error) {
	var count int64
	if err := repohelpers.DBOrTx(ctx, r.db).
		Model(&model.AppointmentTrimmingOption{}).
		Joins("JOIN appointments ON appointments.id = appointment_trimming_options.appointment_id AND appointments.clinic_id = ? AND appointments.deleted_at IS NULL", clinicID).
		Where("appointment_trimming_options.option_id = ?", optionID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "appointment_trimming_option", "")
	}
	return count, nil
}
