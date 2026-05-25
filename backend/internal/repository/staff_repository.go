// Package repository provides data access implementations for Staff entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Staff ----

type StaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	FindByID(ctx context.Context, id uint64) (*model.Staff, error)
	FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
	// Create はスタッフを作成する。
	Create(ctx context.Context, staff *model.Staff) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountBlockingReferencesByStaffID(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error)
}

type StaffDependencyCount struct {
	Label string
	Count int64
}

type staffRepository struct{ db *gorm.DB }

func NewStaffRepository(db *gorm.DB) StaffRepository { return &staffRepository{db: db} }

func (r *staffRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	staffs := make([]model.Staff, 0)
	var total int64

	buildBase := func() *gorm.DB {
		// staffs テーブルに clinic_id は存在しない。
		// staff_clinic_assignments を経由して clinic_id でフィルタ
		q := dbOrTx(ctx, r.db).Model(&model.Staff{}).
			Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id"+
				" AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", clinicID).
			Where("staffs.deleted_at IS NULL")
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "staff", "")
	}
	if err := buildBase().
		Preload("Account", "deleted_at IS NULL").
		Preload("Occupation", "deleted_at IS NULL").
		Offset((page - 1) * limit).Limit(limit).
		Order("staffs.sort_order ASC, staffs.name ASC").
		Distinct("staffs.*").
		Find(&staffs).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "staff", "")
	}
	return staffs, total, nil
}

func (r *staffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	var staff model.Staff
	err := dbOrTx(ctx, r.db).
		Where("deleted_at IS NULL").
		Preload("Account", "deleted_at IS NULL").
		Preload("Occupation", "deleted_at IS NULL").
		First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

func (r *staffRepository) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	var staff model.Staff
	err := dbOrTx(ctx, r.db).Where("deleted_at IS NULL").First(&staff, "account_id = ?", accountID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("account_id=%d", accountID))
	}
	return &staff, nil
}

func (r *staffRepository) Create(ctx context.Context, staff *model.Staff) error {
	if err := dbOrTx(ctx, r.db).Create(staff).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("staff", staff.Name)
		}
		return apperrors.FromGORM(err, "staff", "")
	}
	return nil
}

func (r *staffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// staffs テーブルに clinic_id は存在しない。
	// staff_clinic_assignments を経由して clinic_id でフィルタ
	result := dbOrTx(ctx, r.db).
		Model(&model.Staff{}).
		Where("staffs.id = ?", id).
		Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)", clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *staffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// staffs テーブルに clinic_id は存在しない。
	// staff_clinic_assignments を経由して clinic_id でフィルタ
	result := dbOrTx(ctx, r.db).
		Model(&model.Staff{}).
		Where("id = ?", id).
		Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)", clinicID).
		Update("deleted_at", gorm.Expr("now()"))
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *staffRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	// staffs テーブルに clinic_id カラムは存在しない。
	// staff_clinic_assignments を経由して clinic_id でフィルタする。
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Staff{}).
				Where("staffs.id = ?", id).
				Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)", clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("staff id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to reorder staffs")
	}
	return nil
}

func (r *staffRepository) CountBlockingReferencesByStaffID(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error) {
	checks := []struct {
		table   string
		column  string
		label   string
		softDel bool
	}{
		{table: "medical_records", column: "doctor_id", label: "カルテ", softDel: true},
		{table: "medical_records", column: "entered_by", label: "カルテ入力履歴", softDel: true},
		{table: "medical_record_addenda", column: "author_user_id", label: "カルテ追記", softDel: false},
		{table: "payments", column: "paid_by", label: "支払い", softDel: true},
		{table: "hospitalizations", column: "doctor_id", label: "入院記録", softDel: true},
		{table: "exams", column: "doctor_id", label: "検査", softDel: true},
		{table: "exams", column: "entered_by", label: "検査入力履歴", softDel: true},
		{table: "vital_records", column: "doctor_id", label: "バイタル記録", softDel: true},
		{table: "shift_entries", column: "staff_id", label: "シフト", softDel: false},
		{table: "billing_refunds", column: "refunded_by", label: "返金", softDel: false},
		{table: "cash_register_closes", column: "closed_by", label: "レジ締め", softDel: false},
	}

	dependencies := make([]StaffDependencyCount, 0, len(checks))
	for _, check := range checks {
		query := fmt.Sprintf("clinic_id = ? AND %s = ?", check.column)
		if check.softDel {
			query += " AND deleted_at IS NULL"
		}

		var count int64
		if err := dbOrTx(ctx, r.db).
			Table(check.table).
			Where(query, clinicID, staffID).
			Count(&count).Error; err != nil {
			return nil, apperrors.FromGORM(err, check.table, fmt.Sprintf("staff_id=%d", staffID))
		}
		if count > 0 {
			dependencies = append(dependencies, StaffDependencyCount{Label: check.label, Count: count})
		}
	}
	return dependencies, nil
}
