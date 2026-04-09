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
	// ExcludedServiceTypes
	FindExcludedServiceTypes(ctx context.Context, staffID uint64) ([]model.StaffExcludedServiceType, error)
	ReplaceExcludedServiceTypes(ctx context.Context, staffID uint64, courseIDs []uint64) error
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
			return apperrors.Wrap(err, "create reservation staff")
		}
		assignment := &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: clinicID,
			IsMain:   true,
		}
		if err := tx.Create(assignment).Error; err != nil {
			return apperrors.Wrap(err, "create staff clinic assignment")
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
			return apperrors.Wrap(err, "swap sort_order target")
		}
		if err := tx.Model(&model.Staff{}).Where("id = ?", neighbor.ID).Update("sort_order", targetOrder).Error; err != nil {
			return apperrors.Wrap(err, "swap sort_order neighbor")
		}
		return nil
	})
}

func (r *reservationStaffRepository) FindExcludedServiceTypes(ctx context.Context, staffID uint64) ([]model.StaffExcludedServiceType, error) {
	var items []model.StaffExcludedServiceType
	err := r.db.WithContext(ctx).
		Preload("ServiceType").
		Where("staff_id = ?", staffID).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", "")
	}
	return items, nil
}

// ReplaceExcludedServiceTypes は staffID の除外コースを courseIDs で完全置換する（差分更新）
func (r *reservationStaffRepository) ReplaceExcludedServiceTypes(ctx context.Context, staffID uint64, courseIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存を全削除
		if err := tx.Where("staff_id = ?", staffID).Delete(&model.StaffExcludedServiceType{}).Error; err != nil {
			return apperrors.Wrap(err, "delete excluded service types")
		}
		// 新規挿入
		if len(courseIDs) == 0 {
			return nil
		}
		items := make([]model.StaffExcludedServiceType, 0, len(courseIDs))
		for _, cid := range courseIDs {
			items = append(items, model.StaffExcludedServiceType{
				StaffID:       staffID,
				ServiceTypeID: cid,
			})
		}
		if err := tx.Create(&items).Error; err != nil {
			return apperrors.Wrap(err, "insert excluded service types")
		}
		return nil
	})
}
