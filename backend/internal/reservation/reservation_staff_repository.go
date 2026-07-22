package reservation

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
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

type reservationStaffRepository struct {
	db *gorm.DB
	// staff は staffs テーブルの唯一の書き込み者（staff domain・ADR-006 論点#1 案A）の
	// consumer-side view。read と junction (staff_reservation_exclusions/capabilities) は
	// reservation 所有のまま。具象は repository facade（staffRepository）が注入する。
	staff staffsWriter
}

// staffsWriter は staff domain の予約用途 staffs write の最小 view（ADR-006 論点#1 案A）。
type staffsWriter interface {
	CreateForReservation(ctx context.Context, staff *model.Staff, clinicID uint64) error
	UpdateForReservation(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	DeleteForReservation(ctx context.Context, clinicID, id uint64) error
	SwapSortOrderForReservation(ctx context.Context, clinicID, id uint64, direction string) error
}

func NewReservationStaffRepository(db *gorm.DB, staff staffsWriter) ReservationStaffRepository {
	return &reservationStaffRepository{db: db, staff: staff}
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
// BE-refactor.md H-7: repohelpers.DBOrTx(ctx, r.db) にすることで、reservationStaffService.Update が
// WithTx 閉包内で行う所有権確認がその ambient tx に参加する（確認〜更新の TOCTOU 窓を閉じる）。
func (r *reservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	var staff model.Staff
	db := repohelpers.DBOrTx(ctx, r.db)
	if repohelpers.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id AND sca.clinic_id = ? AND sca.deleted_at IS NULL", clinicID).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		First(&staff).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

// Create は staffRepository.CreateForReservation へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationStaffRepository) Create(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	return r.staff.CreateForReservation(ctx, staff, clinicID)
}

// Update は staffRepository.UpdateForReservation へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return r.staff.UpdateForReservation(ctx, clinicID, id, fields)
}

// Delete は staffRepository.DeleteForReservation へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return r.staff.DeleteForReservation(ctx, clinicID, id)
}

func (r *reservationStaffRepository) CountUsageByStaffID(ctx context.Context, clinicID, staffID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("doctor_id = ? AND deleted_at IS NULL", staffID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "appointment", fmt.Sprintf("staff_id=%d", staffID))
	}
	return count, nil
}

// UpdateSortOrder は staffRepository.SwapSortOrderForReservation へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationStaffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	return r.staff.SwapSortOrderForReservation(ctx, clinicID, id, direction)
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
	// BE-refactor.md X-8: repohelpers.DBOrTx(ctx, r.db) にすることで、この所有権検証読み取りも後続の書込と
	// 同一 ambient tx 内で行われる（ambient tx 内で先行コミットされた ReservationType の変更も見える）。
	if err := repohelpers.ValidateClinicScopedMasterIDs(ctx, repohelpers.DBOrTx(ctx, r.db), clinicID, courseIDs,
		&model.ReservationType{}, "reservation_type",
		"reservation_type_ids contains invalid reservation type"); err != nil {
		return err
	}
	// BE-refactor.md X-8: repohelpers.DBOrTx(ctx, r.db).Transaction(...) にすることで、ambient tx があれば
	// そのネスト tx（SAVEPOINT）として参加する。過去は r.db.WithContext(ctx).Transaction(...) で
	// 常に独立した新規 tx を開始しており、reservationStaffService.Create の外側 WithTx が rollback
	// しても本メソッドの DELETE→INSERT は既にコミット済みのため巻き戻らなかった。
	return repohelpers.ReplaceJunctionInTransaction(repohelpers.DBOrTx(ctx, r.db), func(tx *gorm.DB) error {
		// BE-refactor.md H-2: staff_reservation_exclusions は自前 clinic_id を持たないため、
		// staff_id のみで DELETE すると多施設所属スタッフ（staff_clinic_assignments）が
		// 他クリニックで正当に保持する除外設定まで削除してしまう（UpdateReservationCapabilities
		// は junction 自体に clinic_id 列を持ち対称に scope 済み — deleteJunctionByClinicAndStaff 参照）。
		// reservation_types.clinic_id のサブクエリで削除対象を clinicID にスコープする。
		if err := repohelpers.DeleteJunctionViaMasterClinicScope(tx, clinicID, staffID,
			&model.StaffReservationExclusion{}, &model.ReservationType{}, "reservation_type_id",
			"staff_reservation_exclusion", fmt.Sprintf("%d", staffID)); err != nil {
			return err
		}
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
		return repohelpers.InsertJunctionRowsInBatches(tx, items, "staff_reservation_exclusion", "")
	}, "failed to replace excluded reservation types")
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
	// BE-refactor.md X-8: repohelpers.DBOrTx(ctx, r.db) にすることで、この所有権検証読み取りも後続の書込と
	// 同一 ambient tx 内で行われる（UpdateExcludedReservationTypes と対称）。
	if err := repohelpers.ValidateClinicScopedMasterIDs(ctx, repohelpers.DBOrTx(ctx, r.db), clinicID, typeIDs,
		&model.ReservationType{}, "reservation_type",
		"reservation_type_ids contains invalid reservation type"); err != nil {
		return err
	}
	// BE-refactor.md X-8: repohelpers.DBOrTx(ctx, r.db).Transaction(...) で ambient tx 参加を統一する
	// （UpdateExcludedReservationTypes と対称）。
	return repohelpers.ReplaceJunctionInTransaction(repohelpers.DBOrTx(ctx, r.db), func(tx *gorm.DB) error {
		if err := repohelpers.DeleteJunctionByClinicAndStaff(tx, clinicID, staffID,
			&model.StaffReservationCapability{}, "staff_reservation_capability", fmt.Sprintf("%d", staffID)); err != nil {
			return err
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
		return repohelpers.InsertJunctionRowsInBatches(tx, items, "staff_reservation_capability", "")
	}, "failed to replace capable reservation types")
}

func (r *reservationStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	var capability model.StaffReservationCapability
	db := repohelpers.DBOrTx(ctx, r.db).Model(&model.StaffReservationCapability{})
	if repohelpers.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Select("id").
		Where("clinic_id = ? AND staff_id = ? AND reservation_type_id = ?", clinicID, staffID, reservationTypeID).
		Take(&capability).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, apperrors.FromGORM(err, "staff_reservation_capability", "")
	}
	return true, nil
}
