package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// StaffClinicAssignmentRepository はスタッフ-クリニック中間テーブルのインターフェース
type StaffClinicAssignmentRepository interface {
	FindByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	ExistsByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (bool, error)
	Create(ctx context.Context, assignment *model.StaffClinicAssignment) error
	DeleteByStaffID(ctx context.Context, staffID uint64) error
}

// staffClinicAssignmentRepository は StaffClinicAssignmentRepository の実装
type staffClinicAssignmentRepository struct {
	db *gorm.DB
}

// NewStaffClinicAssignmentRepository は StaffClinicAssignmentRepository を初期化して返す
func NewStaffClinicAssignmentRepository(db *gorm.DB) StaffClinicAssignmentRepository {
	return &staffClinicAssignmentRepository{db: db}
}

// FindByStaffID はスタッフIDでクリニック所属を取得する
func (r *staffClinicAssignmentRepository) FindByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	var assignments []model.StaffClinicAssignment
	if err := dbOrTx(ctx, r.db).Where("staff_id = ?", staffID).Find(&assignments).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff_id=%d", staffID))
	}
	return assignments, nil
}

// ExistsByStaffAndClinic はスタッフが指定クリニックに所属しているかを確認する
func (r *staffClinicAssignmentRepository) ExistsByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (bool, error) {
	var count int64
	if err := dbOrTx(ctx, r.db).Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ? AND clinic_id = ?", staffID, clinicID).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID))
	}
	return count > 0, nil
}

// Create は新規クリニック所属を作成する
func (r *staffClinicAssignmentRepository) Create(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	if err := dbOrTx(ctx, r.db).Create(assignment).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", "create")
	}
	return nil
}

// DeleteByStaffID はスタッフの全クリニック所属を削除する
func (r *staffClinicAssignmentRepository) DeleteByStaffID(ctx context.Context, staffID uint64) error {
	if err := dbOrTx(ctx, r.db).Where("staff_id = ?", staffID).Delete(&model.StaffClinicAssignment{}).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff_id=%d", staffID))
	}
	return nil
}
