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
	FindByClinicID(ctx context.Context, clinicID uint64) ([]model.StaffClinicAssignment, error)
	Create(ctx context.Context, assignment *model.StaffClinicAssignment) error
	Update(ctx context.Context, assignment *model.StaffClinicAssignment) error
	Delete(ctx context.Context, staffID, clinicID uint64) error
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
	if err := r.db.WithContext(ctx).Where("staff_id = ?", staffID).Find(&assignments).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff_id=%d", staffID))
	}
	return assignments, nil
}

// FindByClinicID はクリニックIDでスタッフ所属を取得する
func (r *staffClinicAssignmentRepository) FindByClinicID(ctx context.Context, clinicID uint64) ([]model.StaffClinicAssignment, error) {
	var assignments []model.StaffClinicAssignment
	if err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Find(&assignments).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return assignments, nil
}

// Create は新規クリニック所属を作成する
func (r *staffClinicAssignmentRepository) Create(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	if err := r.db.WithContext(ctx).Create(assignment).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", "create")
	}
	return nil
}

// Update はクリニック所属情報を更新する
func (r *staffClinicAssignmentRepository) Update(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	if err := r.db.WithContext(ctx).Save(assignment).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("%d", assignment.ID))
	}
	return nil
}

// Delete はクリニック所属を削除する
func (r *staffClinicAssignmentRepository) Delete(ctx context.Context, staffID, clinicID uint64) error {
	if err := r.db.WithContext(ctx).Where("staff_id = ? AND clinic_id = ?", staffID, clinicID).Delete(&model.StaffClinicAssignment{}).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff_id=%d, clinic_id=%d", staffID, clinicID))
	}
	return nil
}

// DeleteByStaffID はスタッフの全クリニック所属を削除する
func (r *staffClinicAssignmentRepository) DeleteByStaffID(ctx context.Context, staffID uint64) error {
	if err := r.db.WithContext(ctx).Where("staff_id = ?", staffID).Delete(&model.StaffClinicAssignment{}).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff_id=%d", staffID))
	}
	return nil
}
