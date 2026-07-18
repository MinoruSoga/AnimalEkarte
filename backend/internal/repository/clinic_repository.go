package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ClinicRepository interface {
	FindAll(ctx context.Context) ([]model.Clinic, error)
	FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	FindByID(ctx context.Context, id uint64) (*model.Clinic, error)
	FindCompany(ctx context.Context) (*model.Company, error)
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
	err := dbOrTx(ctx, r.db).First(&clinic, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", id))
	}
	return &clinic, nil
}

func (r *clinicRepository) FindCompany(ctx context.Context) (*model.Company, error) {
	var company model.Company
	err := dbOrTx(ctx, r.db).First(&company).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "company", "singleton")
	}
	return &company, nil
}

// BE-refactor.md X-7: dbOrTx で ambient tx に参加する。CreateClinic は clinic 作成 +
// デフォルト権限グループ2件の作成を transactor.WithTx で包むが、Create が r.db.WithContext(ctx)
// のまま tx 非参加だと、途中で失敗しても既にオートコミット済みの行は WithTx のロールバックで
// 巻き戻らず、デフォルト権限グループなしの孤児クリニックが生成しうるバグがあった。
func (r *clinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	err := dbOrTx(ctx, r.db).Create(clinic).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("clinic", clinic.Name)
		}
		return apperrors.FromGORM(err, "clinic", "")
	}
	return nil
}

func (r *clinicRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := dbOrTx(ctx, r.db).Model(&model.Clinic{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinic", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
	}
	return nil
}

// BE-refactor.md X-7: dbOrTx(ctx, r.db).Transaction(...) にすることで、ambient tx があれば
// その中のネスト tx（SAVEPOINT）として参加する（R1-1 と同一パターン、accounting_repository.go
// SavePaymentSplits 参照）。現状 Delete を ambient tx から呼ぶ呼び出し元は無く、
// 既存呼び出しに対する挙動は変わらない（ambient tx が無ければ dbOrTx は従来どおり新規 tx を開始する）。
func (r *clinicRepository) Delete(ctx context.Context, id uint64) error {
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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
