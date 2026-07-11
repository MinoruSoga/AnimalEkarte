package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationStaffRepository は予約スタッフ（staffs の予約用ラッパー）のデータアクセスインターフェース
type ReservationStaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	Create(ctx context.Context, staff *model.Staff, clinicID uint64) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	CountUsageByStaffID(ctx context.Context, clinicID, staffID uint64) (int64, error)
	UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
	// ExcludedReservationTypes
	FindAllExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error)
	FindAllExcludedReservationTypesByStaffIDs(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error)
	UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error
	// ReservationCapabilities
	FindAllReservationCapabilities(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error)
	FindAllReservationCapabilitiesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error)
	UpdateReservationCapabilities(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
	SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

type reservationStaffRepository struct{ db *gorm.DB }

func NewReservationStaffRepository(db *gorm.DB) ReservationStaffRepository {
	return &reservationStaffRepository{db: db}
}

func (r *reservationStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
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

// FindByID はクリニック所属チェック込みでスタッフ 1 件を取得する（マルチテナント安全）。
func (r *reservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.WithContext(ctx).
		Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id AND sca.clinic_id = ?", clinicID).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		First(&staff).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

// Create はスタッフ + StaffClinicAssignment をトランザクションで作成する。
// BE-refactor.md X-8: dbOrTx(ctx, r.db).Transaction(...) にすることで、ambient tx（例:
// reservationStaffService.Create の Transactor.WithTx）があればそのネスト tx（SAVEPOINT）
// として参加する。過去は r.db.WithContext(ctx).Transaction(...) で常に独立した新規 tx を
// 開始しており、ambient tx が UpdateExcludedReservationTypes の失敗で rollback しても
// 本メソッドの staff/assignment 作成は既にコミット済みのため巻き戻らなかった
// （除外コース未設定のまま孤児スタッフが残る部分コミットのバグ）。
func (r *reservationStaffRepository) Create(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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
	}); err != nil {
		return apperrors.Wrap(err, "failed to create reservation staff")
	}
	return nil
}

// BE-refactor.md X-8: dbOrTx(ctx, r.db) にすることで、reservationStaffService.Update が
// Transactor.WithTx で本メソッドと UpdateExcludedReservationTypes を括った場合に同一 tx へ参加する。
func (r *reservationStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return updateScopedByID(ctx, dbOrTx(ctx, r.db), &model.Staff{}, "reservation_staff", clinicID, id, fields)
}

// BE-refactor.md X-8: dbOrTx(ctx, r.db) で ambient tx 参加を統一する（他の write メソッドと対称）。
func (r *reservationStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := dbOrTx(ctx, r.db).
		Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id AND sca.clinic_id = ?", clinicID).
		Where("staffs.id = ?", id).
		Delete(&model.Staff{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationStaffRepository) CountUsageByStaffID(ctx context.Context, clinicID, staffID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Scopes(clinicScope(clinicID)).
		Where("doctor_id = ? AND deleted_at IS NULL", staffID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "appointment", fmt.Sprintf("staff_id=%d", staffID))
	}
	return count, nil
}

// BE-refactor.md X-8: dbOrTx(ctx, r.db).Transaction(...) で ambient tx 参加を統一する
// （Create/UpdateExcludedReservationTypes/UpdateReservationCapabilities と対称）。
func (r *reservationStaffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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
			wrapped := apperrors.FromGORM(err, "reservation_staff", "neighbor")
			if errors.Is(wrapped, apperrors.ErrNotFound) {
				// 隣接なし → 変更なし
				return nil
			}
			return wrapped
		}

		targetOrder := target.SortOrder
		neighborOrder := neighbor.SortOrder

		if err := tx.Scopes(clinicScope(clinicID)).Model(&model.Staff{}).Where("id = ?", target.ID).Update("sort_order", neighborOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", target.ID))
		}
		if err := tx.Scopes(clinicScope(clinicID)).Model(&model.Staff{}).Where("id = ?", neighbor.ID).Update("sort_order", targetOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", neighbor.ID))
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to swap sort order")
	}
	return nil
}

// BE-refactor.md R2-5 (D12) レビュー結果: clinic_id 述語なしを意図的に維持する
// (preload_clinic_scope_lint_test.go の site-exception と同一の判断)。
// staff_reservation_exclusions は clinic_id カラムを持たず、staffID にも clinicID 引数が無い。
// 貫通させるには reservation_staff_service (6箇所) + staff_service_permissions の呼び出し元まで
// シグネチャ変更が連鎖し、書込側は UpdateExcludedReservationTypes で既にクリニック所有権検証済み。
// 残存リスクは過去汚染データによる ReservationType 名の低severity漏洩のみ（P1 follow-up として記録済み）。
func (r *reservationStaffRepository) FindAllExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
	var items []model.StaffReservationExclusion
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "deleted_at IS NULL").
		Where("staff_id = ?", staffID).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", "")
	}
	return items, nil
}

// FindAllExcludedReservationTypesByStaffIDs は複数スタッフの除外コースを一括取得する（N+1回避）
func (r *reservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error) {
	if len(staffIDs) == 0 {
		return nil, nil
	}
	var items []model.StaffReservationExclusion
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "deleted_at IS NULL").
		Where("staff_id IN ?", staffIDs).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff_reservation_exclusion", "")
	}
	return items, nil
}

// UpdateExcludedReservationTypes は staffID の除外コースを courseIDs で完全置換する（差分更新）
func (r *reservationStaffRepository) UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error {
	// テナント越境 write 防止: 除外対象の予約区分IDが呼び出し元クリニックに属することを検証する。
	// staff_reservation_exclusions は自前 clinic_id を持たないため、ここで型IDの所有権を
	// 確認しなければ別クリニックの予約区分IDを書き込めてしまう（UpdateReservationCapabilities と対称）。
	// スタッフ所有権は呼び出し側（service/handler）で検証済み。検証は DELETE 前に行い部分書き込みを防ぐ。
	// BE-refactor.md X-8: dbOrTx(ctx, r.db) にすることで、この所有権検証読み取りも後続の書込と
	// 同一 ambient tx 内で行われる（ambient tx 内で先行コミットされた ReservationType の変更も見える）。
	if len(courseIDs) > 0 {
		var count int64
		if err := dbOrTx(ctx, r.db).
			Model(&model.ReservationType{}).
			Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, courseIDs).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type", "")
		}
		if count != int64(len(courseIDs)) {
			return apperrors.WrapInvalidInput("reservation_type_ids contains invalid reservation type")
		}
	}
	// BE-refactor.md X-8: dbOrTx(ctx, r.db).Transaction(...) にすることで、ambient tx があれば
	// そのネスト tx（SAVEPOINT）として参加する。過去は r.db.WithContext(ctx).Transaction(...) で
	// 常に独立した新規 tx を開始しており、reservationStaffService.Create の外側 WithTx が rollback
	// しても本メソッドの DELETE→INSERT は既にコミット済みのため巻き戻らなかった。
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// 既存を全削除
		if err := tx.Where("staff_id = ?", staffID).Delete(&model.StaffReservationExclusion{}).Error; err != nil {
			return apperrors.FromGORM(err, "staff_reservation_exclusion", fmt.Sprintf("%d", staffID))
		}
		// 新規挿入
		if len(courseIDs) == 0 {
			return nil
		}
		items := make([]model.StaffReservationExclusion, 0, len(courseIDs))
		for _, cid := range courseIDs {
			items = append(items, model.StaffReservationExclusion{
				StaffID:           staffID,
				ReservationTypeID: cid,
			})
		}
		if err := tx.Create(&items).Error; err != nil {
			return apperrors.FromGORM(err, "staff_reservation_exclusion", "")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to replace excluded reservation types")
	}
	return nil
}

func (r *reservationStaffRepository) FindAllReservationCapabilities(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error) {
	var items []model.StaffReservationCapability
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("clinic_id = ? AND staff_id = ?", clinicID, staffID).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff_reservation_capability", "")
	}
	return items, nil
}

func (r *reservationStaffRepository) FindAllReservationCapabilitiesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
	if len(staffIDs) == 0 {
		return nil, nil
	}
	var items []model.StaffReservationCapability
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("clinic_id = ? AND staff_id IN ?", clinicID, staffIDs).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff_reservation_capability", "")
	}
	return items, nil
}

func (r *reservationStaffRepository) UpdateReservationCapabilities(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if _, err := r.FindByID(ctx, clinicID, staffID); err != nil {
		return err
	}
	// BE-refactor.md X-8: dbOrTx(ctx, r.db) にすることで、この所有権検証読み取りも後続の書込と
	// 同一 ambient tx 内で行われる（UpdateExcludedReservationTypes と対称）。
	if len(typeIDs) > 0 {
		var count int64
		if err := dbOrTx(ctx, r.db).
			Model(&model.ReservationType{}).
			Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, typeIDs).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type", "")
		}
		if count != int64(len(typeIDs)) {
			return apperrors.WrapInvalidInput("reservation_type_ids contains invalid reservation type")
		}
	}
	// BE-refactor.md X-8: dbOrTx(ctx, r.db).Transaction(...) で ambient tx 参加を統一する
	// （UpdateExcludedReservationTypes と対称）。
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("clinic_id = ? AND staff_id = ?", clinicID, staffID).Delete(&model.StaffReservationCapability{}).Error; err != nil {
			return apperrors.FromGORM(err, "staff_reservation_capability", fmt.Sprintf("%d", staffID))
		}
		if len(typeIDs) == 0 {
			return nil
		}
		items := make([]model.StaffReservationCapability, 0, len(typeIDs))
		for _, typeID := range typeIDs {
			items = append(items, model.StaffReservationCapability{
				ClinicID:          clinicID,
				StaffID:           staffID,
				ReservationTypeID: typeID,
			})
		}
		if err := tx.Create(&items).Error; err != nil {
			return apperrors.FromGORM(err, "staff_reservation_capability", "")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to replace capable reservation types")
	}
	return nil
}

func (r *reservationStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.StaffReservationCapability{}).
		Where("clinic_id = ? AND staff_id = ? AND reservation_type_id = ?", clinicID, staffID, reservationTypeID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "staff_reservation_capability", "")
	}
	return count > 0, nil
}
