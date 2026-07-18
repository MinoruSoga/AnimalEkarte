// Package checkup owns checkups (健診記録) data access.
package checkup

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Filters はクリニック横断一覧のフィルタ条件
type Filters struct {
	StartDate     *string
	EndDate       *string
	NextStartDate *string
	NextEndDate   *string
}

// Repository はCheckup管理のインターフェース
type Repository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	FindByClinicID(ctx context.Context, clinicID uint64, filters Filters, page, limit int) ([]model.Checkup, int64, error)
	// FindByOwnerID は飼い主に紐づく生存健診記録を全件返す（ISSUE-004 タグ再同期用）。
	// medical_records 経由で owner_id を解決する。
	FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Checkup, error)
	Create(ctx context.Context, checkup *model.Checkup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

// repository は Repository の実装
type repository struct {
	db *gorm.DB
}

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// paginate は 1-origin の page と limit を Offset/Limit に変換する共有 Scope。
// （親 repository パッケージの同名ヘルパーの最小限の複製。repohelpers 未収載のため import cycle 回避のためローカル定義。BE8-4 batch24）。
func paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

func (r *repository) FindByClinicID(ctx context.Context, clinicID uint64, filters Filters, page, limit int) ([]model.Checkup, int64, error) {
	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Checkup{}).
			Scopes(repohelpers.ClinicScope(clinicID))
		if filters.StartDate != nil {
			q = q.Where("date >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			q = q.Where("date <= ?", *filters.EndDate)
		}
		if filters.NextStartDate != nil {
			q = q.Where("next_date >= ?", *filters.NextStartDate)
		}
		if filters.NextEndDate != nil {
			q = q.Where("next_date <= ?", *filters.NextEndDate)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "checkup", "")
	}

	checkups := make([]model.Checkup, 0)
	err := buildBase().
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("MedicalRecord", "deleted_at IS NULL").
		Preload("MedicalRecord.Pet", "deleted_at IS NULL").
		Preload("MedicalRecord.Pet.Owner", "deleted_at IS NULL").
		Order("date DESC").
		Scopes(paginate(page, limit)).
		Find(&checkups).Error
	if err != nil {
		return nil, 0, apperrors.FromGORM(err, "checkup", "")
	}
	return checkups, total, nil
}

// FindByOwnerID は飼い主に紐づく生存健診記録を返す（ISSUE-004 タグ再同期用）。
// medical_records 経由で owner_id を解決し、checkup と medical_record の両方が生存しているレコードのみ返す。
func (r *repository) FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error) {
	checkups := make([]model.Checkup, 0)
	err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = checkups.medical_record_id"+
			" AND medical_records.clinic_id = ?"+
			" AND medical_records.deleted_at IS NULL", clinicID).
		Where("checkups.clinic_id = ? AND medical_records.owner_id = ?", clinicID, ownerID).
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("checkups.date DESC").
		Find(&checkups).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", fmt.Sprintf("owner=%d", ownerID))
	}
	return checkups, nil
}

func (r *repository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	checkups := make([]model.Checkup, 0)
	err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = checkups.medical_record_id"+
			" AND medical_records.clinic_id = ?"+
			" AND medical_records.deleted_at IS NULL", clinicID).
		Where("checkups.medical_record_id = ?", medicalRecordID).
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Order("checkups.date ASC").
		Find(&checkups).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", "")
	}
	return checkups, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Checkup, error) {
	var checkup model.Checkup
	err := r.db.WithContext(ctx).
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("MedicalRecord", "deleted_at IS NULL").
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&checkup).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", fmt.Sprintf("%d", id))
	}
	return &checkup, nil
}

func (r *repository) Create(ctx context.Context, checkup *model.Checkup) error {
	err := r.db.WithContext(ctx).Create(checkup).Error
	if err != nil {
		return apperrors.FromGORM(err, "checkup", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return repohelpers.UpdateScopedByID(ctx, r.db, &model.Checkup{}, "checkup", clinicID, id, fields)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.Checkup{}, "checkup", clinicID, id)
}
