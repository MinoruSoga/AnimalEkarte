package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CheckupFilters はクリニック横断一覧のフィルタ条件
type CheckupFilters struct {
	StartDate     *string
	EndDate       *string
	NextStartDate *string
	NextEndDate   *string
}

type CheckupRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	FindByClinicID(ctx context.Context, clinicID uint64, filters CheckupFilters) ([]model.Checkup, error)
	// FindByOwnerID は飼い主に紐づく生存健診記録を全件返す（ISSUE-004 タグ再同期用）。
	// medical_records 経由で owner_id を解決する。
	FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Checkup, error)
	Create(ctx context.Context, checkup *model.Checkup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type checkupRepository struct {
	db *gorm.DB
}

func NewCheckupRepository(db *gorm.DB) CheckupRepository {
	return &checkupRepository{db: db}
}

func (r *checkupRepository) FindByClinicID(ctx context.Context, clinicID uint64, filters CheckupFilters) ([]model.Checkup, error) {
	checkups := make([]model.Checkup, 0)
	q := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("MedicalRecord", "deleted_at IS NULL").
		Preload("MedicalRecord.Pet", "deleted_at IS NULL").
		Preload("MedicalRecord.Pet.Owner", "deleted_at IS NULL")
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
	err := q.Order("date DESC").Find(&checkups).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", "")
	}
	return checkups, nil
}

// FindByOwnerID は飼い主に紐づく生存健診記録を返す（ISSUE-004 タグ再同期用）。
// medical_records 経由で owner_id を解決し、checkup と medical_record の両方が生存しているレコードのみ返す。
func (r *checkupRepository) FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error) {
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

func (r *checkupRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
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

func (r *checkupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Checkup, error) {
	var checkup model.Checkup
	err := r.db.WithContext(ctx).
		Preload("CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("MedicalRecord", "deleted_at IS NULL").
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		First(&checkup).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup", fmt.Sprintf("%d", id))
	}
	return &checkup, nil
}

func (r *checkupRepository) Create(ctx context.Context, checkup *model.Checkup) error {
	err := r.db.WithContext(ctx).Create(checkup).Error
	if err != nil {
		return apperrors.FromGORM(err, "checkup", "")
	}
	return nil
}

func (r *checkupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return updateScopedByID(ctx, r.db, &model.Checkup{}, "checkup", clinicID, id, fields)
}

func (r *checkupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.Checkup{}, "checkup", clinicID, id)
}
