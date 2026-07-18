// Package staffclinicassignment owns staff_clinic_assignments data access
// (BE8-4 batch19 — leaf domain: staff-clinic junction table).
package staffclinicassignment

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository はスタッフ-クリニック中間テーブルのインターフェース
type Repository interface {
	FindByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	CountByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (int64, error)
	Create(ctx context.Context, assignment *model.StaffClinicAssignment) error
	Delete(ctx context.Context, staffID uint64) error
}

// repository は Repository の実装
type repository struct {
	db *gorm.DB
}

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByStaffID はスタッフIDでクリニック所属を取得する（複数件）
// NOTE: model.StaffClinicAssignment は gorm.DeletedAt を持つため、GORM SoftDelete スコープにより
// deleted_at IS NULL フィルタは自動適用される。明示的な条件追加は不要。
func (r *repository) FindByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	var assignments []model.StaffClinicAssignment
	if err := repohelpers.DBOrTx(ctx, r.db).Where("staff_id = ?", staffID).Find(&assignments).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff_id=%d", staffID))
	}
	return assignments, nil
}

// CountByStaffAndClinic はスタッフが指定クリニックに所属しているレコード数を返す
// NOTE: GORM SoftDelete スコープにより deleted_at IS NULL は自動適用される。
func (r *repository) CountByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (int64, error) {
	var count int64
	if err := repohelpers.DBOrTx(ctx, r.db).Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ? AND clinic_id = ?", staffID, clinicID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID))
	}
	return count, nil
}

// Create は新規クリニック所属を作成する
func (r *repository) Create(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Create(assignment).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", "create")
	}
	return nil
}

// Delete はスタッフの全クリニック所属を削除する
func (r *repository) Delete(ctx context.Context, staffID uint64) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Where("staff_id = ?", staffID).Delete(&model.StaffClinicAssignment{}).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff_id=%d", staffID))
	}
	return nil
}
