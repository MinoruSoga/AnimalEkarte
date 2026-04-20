package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TrimmingOptionRepository はトリミングオプションのデータアクセスインターフェース
type TrimmingOptionRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountRecordsByOptionID(ctx context.Context, clinicID, optionID uint64) (int64, error)
}

type trimmingOptionRepository struct{ db *gorm.DB }

// NewTrimmingOptionRepository は TrimmingOptionRepository を生成する
func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return &trimmingOptionRepository{db: db}
}

func (r *trimmingOptionRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	options := make([]model.TrimmingOption, 0)
	if err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&options).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", "")
	}
	return options, nil
}

func (r *trimmingOptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	var option model.TrimmingOption
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&option).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", id))
	}
	return &option, nil
}

func (r *trimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	if err := r.db.WithContext(ctx).Create(option).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "trimming_option", "")
	}
	return nil
}

func (r *trimmingOptionRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error) {
	result := r.db.WithContext(ctx).
		Model(&model.TrimmingOption{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "trimming_option", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("trimming_option", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingOptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.TrimmingOption{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "trimming_option", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_option", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *trimmingOptionRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.TrimmingOption{}, "trimming_option", clinicID, ids)
}

// CountRecordsByOptionID は指定オプションを使用しているトリミングオプション数を返す（BUG-201）
// appointment_trimming_options は直接 clinic_id を持たないため appointments を JOIN してテナント分離する
func (r *trimmingOptionRepository) CountRecordsByOptionID(ctx context.Context, clinicID, optionID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.AppointmentTrimmingOption{}).
		Joins("JOIN appointments ON appointments.id = appointment_trimming_options.appointment_id AND appointments.clinic_id = ? AND appointments.deleted_at IS NULL", clinicID).
		Where("appointment_trimming_options.option_id = ?", optionID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "appointment_trimming_option", "")
	}
	return count, nil
}
