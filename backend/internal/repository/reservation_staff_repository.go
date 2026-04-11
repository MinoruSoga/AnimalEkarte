package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationStaffRepository は予約スタッフ（staffs の予約用ラッパー）のデータアクセスインターフェース
type ReservationStaffRepository interface {
	FindAllByClinicID(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	FindByID(ctx context.Context, id uint64) (*model.Staff, error)
	Create(ctx context.Context, staff *model.Staff, clinicID uint64) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	SoftDelete(ctx context.Context, id uint64) error
	SwapSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
	// ExcludedReservationCategories
	FindExcludedReservationCategories(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error)
	FindExcludedReservationCategoriesByStaffIDs(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error)
	ReplaceExcludedReservationCategories(ctx context.Context, staffID uint64, courseIDs []uint64) error
}

type reservationStaffRepository struct{ db *gorm.DB }

func NewReservationStaffRepository(db *gorm.DB) ReservationStaffRepository {
	return &reservationStaffRepository{db: db}
}

func (r *reservationStaffRepository) FindAllByClinicID(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	var staffs []model.Staff
	err := r.db.WithContext(ctx).
		Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id AND sca.clinic_id = ?", clinicID).
		Where("staffs.deleted_at IS NULL").
		Order("staffs.sort_order ASC, staffs.id ASC").
		Find(&staffs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", "")
	}
	return staffs, nil
}

func (r *reservationStaffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.WithContext(ctx).First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

// Create はスタッフ + StaffClinicAssignment をトランザクションで作成する
func (r *reservationStaffRepository) Create(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(staff).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_staff", "")
		}
		assignment := &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: clinicID,
			IsMain:   true,
		}
		if err := tx.Create(assignment).Error; err != nil {
			return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff=%d", staff.ID))
		}
		return nil
	})
}

func (r *reservationStaffRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationStaffRepository) SoftDelete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Staff{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationStaffRepository) SwapSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target model.Staff
		err := tx.
			Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id AND sca.clinic_id = ?", clinicID).
			Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
			First(&target).Error
		if err != nil {
			return apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", id))
		}

		var neighbor model.Staff
		q := tx.
			Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id AND sca.clinic_id = ?", clinicID).
			Where("staffs.deleted_at IS NULL")
		if direction == "up" {
			q = q.Where("staffs.sort_order < ?", target.SortOrder).Order("staffs.sort_order DESC")
		} else {
			q = q.Where("staffs.sort_order > ?", target.SortOrder).Order("staffs.sort_order ASC")
		}
		if err := q.First(&neighbor).Error; err != nil {
			// 隣接なし → 変更なし
			return nil
		}

		targetOrder := target.SortOrder
		neighborOrder := neighbor.SortOrder

		if err := tx.Model(&model.Staff{}).Where("id = ?", target.ID).Update("sort_order", neighborOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", target.ID))
		}
		if err := tx.Model(&model.Staff{}).Where("id = ?", neighbor.ID).Update("sort_order", targetOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", neighbor.ID))
		}
		return nil
	})
}

func (r *reservationStaffRepository) FindExcludedReservationCategories(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
	var items []model.StaffReservationExclusion
	err := r.db.WithContext(ctx).
		Preload("ReservationType").
		Where("staff_id = ?", staffID).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", "")
	}
	return items, nil
}

// FindExcludedReservationCategoriesByStaffIDs は複数スタッフの除外コースを一括取得する（N+1回避）
func (r *reservationStaffRepository) FindExcludedReservationCategoriesByStaffIDs(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error) {
	if len(staffIDs) == 0 {
		return nil, nil
	}
	var items []model.StaffReservationExclusion
	err := r.db.WithContext(ctx).
		Preload("ReservationType").
		Where("staff_id IN ?", staffIDs).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff_excluded_reservation_category", "")
	}
	return items, nil
}

// ReplaceExcludedReservationCategories は staffID の除外コースを courseIDs で完全置換する（差分更新）
func (r *reservationStaffRepository) ReplaceExcludedReservationCategories(ctx context.Context, staffID uint64, courseIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存を全削除
		if err := tx.Where("staff_id = ?", staffID).Delete(&model.StaffReservationExclusion{}).Error; err != nil {
			return apperrors.FromGORM(err, "staff_excluded_reservation_category", fmt.Sprintf("%d", staffID))
		}
		// 新規挿入
		if len(courseIDs) == 0 {
			return nil
		}
		items := make([]model.StaffReservationExclusion, 0, len(courseIDs))
		for _, cid := range courseIDs {
			items = append(items, model.StaffReservationExclusion{
				StaffID:       staffID,
				ReservationTypeID: cid,
			})
		}
		if err := tx.Create(&items).Error; err != nil {
			return apperrors.FromGORM(err, "staff_excluded_reservation_category", "")
		}
		return nil
	})
}
