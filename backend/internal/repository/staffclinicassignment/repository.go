// Package staffclinicassignment owns staff_clinic_assignments data access
// (BE8-4 batch19 — leaf domain: staff-clinic junction table).
package staffclinicassignment

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository はスタッフ-クリニック中間テーブルのインターフェース
type Repository interface {
	FindByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	CountByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (int64, error)
	LockActiveByStaff(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	LockActiveByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error)
	RestoreOrCreate(ctx context.Context, assignment *model.StaffClinicAssignment) error
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

// LockActiveByStaff はスタッフの有効な所属行を決定的な順序で FOR UPDATE 取得する。
// 所属全置換とスタッフ削除が同じ staff -> assignment のロック順に参加するために使う。
func (r *repository) LockActiveByStaff(
	ctx context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	if repohelpers.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError(
			"staff clinic assignment lock requires an active transaction",
		)
	}
	var assignments []model.StaffClinicAssignment
	if err := repohelpers.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staff_id = ?", staffID).
		Order("clinic_id ASC, id ASC").
		Find(&assignments).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff_id=%d", staffID),
		)
	}
	return assignments, nil
}

// LockActiveByStaffAndClinic は有効な所属行を FOR SHARE で取得する。
// ShiftEntry.Create のトランザクション内で使用し、作成完了まで並行する所属解除
// （deleted_at の更新）を待機させることで所属確認とシフト書き込みの TOCTOU を防ぐ。
func (r *repository) LockActiveByStaffAndClinic(
	ctx context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	if repohelpers.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError(
			"staff clinic assignment lock requires an active transaction",
		)
	}
	var assignment model.StaffClinicAssignment
	if err := repohelpers.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("staff_id = ? AND clinic_id = ?", staffID, clinicID).
		First(&assignment).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID),
		)
	}
	return &assignment, nil
}

// RestoreOrCreate は本番の FULL UNIQUE(staff_id, clinic_id) を利用し、
// 論理削除済みの所属行を復元するか、未登録の組み合わせを新規作成する。
func (r *repository) RestoreOrCreate(
	ctx context.Context,
	assignment *model.StaffClinicAssignment,
) error {
	if repohelpers.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError(
			"staff clinic assignment restore requires an active transaction",
		)
	}
	if assignment == nil {
		return apperrors.WrapInvalidInput("staff clinic assignment is required")
	}

	candidate := &model.StaffClinicAssignment{
		StaffID:  assignment.StaffID,
		ClinicID: assignment.ClinicID,
		IsMain:   assignment.IsMain,
	}
	if err := repohelpers.DBOrTx(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "staff_id"},
				{Name: "clinic_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"is_main",
				"updated_at",
				"deleted_at",
			}),
		}).
		Create(candidate).Error; err != nil {
		return apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff=%d,clinic=%d", assignment.StaffID, assignment.ClinicID),
		)
	}
	return nil
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
