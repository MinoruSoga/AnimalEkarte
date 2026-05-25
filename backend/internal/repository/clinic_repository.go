package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ClinicRepository interface {
	FindAll(ctx context.Context) ([]model.Clinic, error)
	FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	FindByID(ctx context.Context, id uint64) (*model.Clinic, error)
	GetCompany(ctx context.Context) (*model.Company, error)
	Create(ctx context.Context, clinic *model.Clinic) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	CountOwnersByClinicID(ctx context.Context, clinicID uint64) (int64, error)
	CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error)
	CountBlockingReferencesByClinicID(ctx context.Context, clinicID uint64) ([]ClinicDependencyCount, error)
}

type ClinicDependencyCount struct {
	Label string
	Count int64
}

type clinicRepository struct {
	db *gorm.DB
}

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepository{db: db}
}

func (r *clinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	clinics := make([]model.Clinic, 0)
	err := r.db.WithContext(ctx).Order("name ASC").Find(&clinics).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", "")
	}
	return clinics, nil
}

func (r *clinicRepository) FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error) {
	clinics := make([]model.Clinic, 0)
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.clinic_id = clinics.id"+
			" AND staff_clinic_assignments.deleted_at IS NULL").
		Where("staff_clinic_assignments.staff_id = ?", staffID).
		Order("clinics.name ASC").
		Find(&clinics).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("staff_id=%d", staffID))
	}
	return clinics, nil
}

func (r *clinicRepository) FindByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	var clinic model.Clinic
	err := r.db.WithContext(ctx).First(&clinic, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", id))
	}
	return &clinic, nil
}

func (r *clinicRepository) GetCompany(ctx context.Context) (*model.Company, error) {
	var company model.Company
	err := r.db.WithContext(ctx).First(&company).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "company", "singleton")
	}
	return &company, nil
}

func (r *clinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	err := r.db.WithContext(ctx).Create(clinic).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("clinic", clinic.Name)
		}
		return apperrors.FromGORM(err, "clinic", "")
	}
	return nil
}

func (r *clinicRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).Model(&model.Clinic{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinic", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *clinicRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Soft-deleted PG rows remain as physical rows and block the clinic FK.
		// Hard-delete only those rows (deleted_at IS NOT NULL) before removing the clinic.
		// Active PGs (deleted_at IS NULL) are still caught by CountBlockingReferencesByClinicID → 409.
		if err := tx.Unscoped().
			Where("clinic_id = ? AND deleted_at IS NOT NULL", id).
			Delete(&model.PermissionGroup{}).Error; err != nil {
			return apperrors.FromGORM(err, "permission_group", fmt.Sprintf("clinic_id=%d", id))
		}
		result := tx.Delete(&model.Clinic{}, "id = ?", id)
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "clinic", fmt.Sprintf("%d", id))
		}
		if result.RowsAffected == 0 {
			return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to delete clinic")
	}
	return nil
}

func (r *clinicRepository) CountOwnersByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Owner{}).
		Scopes(clinicScope(clinicID)).Where("deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "owner", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return count, nil
}

func (r *clinicRepository) CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", clinicID).
		Where("staffs.deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "staff", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return count, nil
}

func (r *clinicRepository) CountBlockingReferencesByClinicID(ctx context.Context, clinicID uint64) ([]ClinicDependencyCount, error) {
	checks := []struct {
		table   string
		label   string
		softDel bool
	}{
		{table: "appointments", label: "予約", softDel: true},
		{table: "medical_records", label: "カルテ", softDel: true},
		{table: "hospitalizations", label: "入院記録", softDel: true},
		{table: "exams", label: "検査", softDel: true},
		{table: "vaccinations", label: "ワクチン接種記録", softDel: true},
		{table: "checkups", label: "健診記録", softDel: true},
		{table: "billings", label: "会計", softDel: true},
		{table: "clinic_settings", label: "医院設定", softDel: false},
		{table: "clinic_integrations", label: "連携設定", softDel: false},
		{table: "lstep_settings", label: "Lステップ設定", softDel: false},
		{table: "permission_groups", label: "権限グループ", softDel: true},
	}

	dependencies := make([]ClinicDependencyCount, 0, len(checks))
	for _, check := range checks {
		query := "clinic_id = ?"
		if check.softDel {
			query += " AND deleted_at IS NULL"
		}

		var count int64
		if err := r.db.WithContext(ctx).
			Table(check.table).
			Where(query, clinicID).
			Count(&count).Error; err != nil {
			return nil, apperrors.FromGORM(err, check.table, fmt.Sprintf("clinic_id=%d", clinicID))
		}
		if count > 0 {
			dependencies = append(dependencies, ClinicDependencyCount{Label: check.label, Count: count})
		}
	}
	return dependencies, nil
}
