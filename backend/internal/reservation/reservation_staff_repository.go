package reservation

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReservationStaffRepository は予約スタッフ（staffs の予約用ラッパー）のデータアクセスインターフェース
type ReservationStaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	LockForMutation(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	Create(ctx context.Context, staff *model.Staff, clinicID uint64) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
	// ExcludedReservationTypes
	FindAllExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationExclusion, error)
	FindAllExcludedReservationTypesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationExclusion, error)
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
	SwapSortOrderForReservation(ctx context.Context, clinicID, id uint64, direction string) error
}

func NewReservationStaffRepository(
	db *gorm.DB,
	staff staffsWriter,
) ReservationStaffRepository {
	return &reservationStaffRepository{db: db, staff: staff}
}

func (r *reservationStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	var staffs []model.Staff
	err := r.db.WithContext(ctx).
		Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id"+
			" AND sca.clinic_id = ? AND sca.deleted_at IS NULL", clinicID).
		Where("staffs.deleted_at IS NULL").
		Order("staffs.sort_order ASC, staffs.id ASC").
		Find(&staffs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", "")
	}
	return staffs, nil
}

// FindByID はクリニック所属チェック込みでスタッフ 1 件を取得する（マルチテナント安全）。
// ambient tx 内では、スタッフ identity → 有効な所属行の順に個別の SHARE lock を取得する。
// JOIN への FOR SHARE は PostgreSQL の実行計画にロック順を委ねるため、所属解除・スタッフ削除と
// 予約書き込みの lock order を決定的に揃えられない。個別 query にすることで、後続の
// SupportsReservationType（capability SHARE lock）まで
// staff → assignment → capability の順を明示する。
func (r *reservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	var staff model.Staff
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		First(&staff).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", id))
	}

	var assignment model.StaffClinicAssignment
	assignmentDB := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		assignmentDB = assignmentDB.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := assignmentDB.
		Where("staff_id = ? AND clinic_id = ?", id, clinicID).
		First(&assignment).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff=%d,clinic=%d", id, clinicID),
		)
	}
	return &staff, nil
}

func lockReservationStaffMutationScope(
	tx *gorm.DB,
	clinicID, staffID uint64,
) (*model.Staff, error) {
	var staff model.Staff
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", staffID).
		First(&staff).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"reservation_staff",
			strconv.FormatUint(staffID, 10),
		)
	}

	var assignment model.StaffClinicAssignment
	if err := tx.
		Select("id").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		First(&assignment).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID),
		)
	}
	return &staff, nil
}

// LockForMutation は reservation staff の mutation ownership を
// staff -> scoped assignment の正規順で取得する。
func (r *reservationStaffRepository) LockForMutation(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError(
			"reservation staff mutation lock requires an active transaction",
		)
	}
	return lockReservationStaffMutationScope(persistence.DBOrTx(ctx, r.db), clinicID, id)
}

// Create は staffRepository.CreateForReservation へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationStaffRepository) Create(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	return r.staff.CreateForReservation(ctx, staff, clinicID)
}

// Update は staffRepository.UpdateForReservation へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return r.staff.UpdateForReservation(ctx, clinicID, id, fields)
}

// UpdateSortOrder は staffRepository.SwapSortOrderForReservation へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationStaffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	return r.staff.SwapSortOrderForReservation(ctx, clinicID, id, direction)
}

// FindAllExcludedReservationTypes scopes the clinic-less junction
// through its reservation_types master. This is required for shared staff:
// staff_id alone spans every clinic assignment and would disclose another
// clinic's reservation type IDs and names.
func (r *reservationStaffRepository) FindAllExcludedReservationTypes(
	ctx context.Context,
	clinicID, staffID uint64,
) ([]model.StaffReservationExclusion, error) {
	var items []model.StaffReservationExclusion
	err := persistence.DBOrTx(ctx, r.db).
		Joins(
			"JOIN reservation_types ON reservation_types.id = staff_reservation_exclusions.reservation_type_id"+
				" AND reservation_types.clinic_id = ? AND reservation_types.deleted_at IS NULL",
			clinicID,
		).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("staff_reservation_exclusions.staff_id = ?", staffID).
		Where(
			"EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staff_reservation_exclusions.staff_id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)",
			clinicID,
		).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff_reservation_exclusion", "")
	}
	return items, nil
}

// FindAllExcludedReservationTypesByStaffIDs は複数スタッフの除外コースを医院スコープで一括取得する（N+1回避）
func (r *reservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(
	ctx context.Context,
	clinicID uint64,
	staffIDs []uint64,
) ([]model.StaffReservationExclusion, error) {
	if len(staffIDs) == 0 {
		return nil, nil
	}
	var items []model.StaffReservationExclusion
	err := persistence.DBOrTx(ctx, r.db).
		Joins(
			"JOIN reservation_types ON reservation_types.id = staff_reservation_exclusions.reservation_type_id"+
				" AND reservation_types.clinic_id = ? AND reservation_types.deleted_at IS NULL",
			clinicID,
		).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("staff_reservation_exclusions.staff_id IN ?", staffIDs).
		Where(
			"EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staff_reservation_exclusions.staff_id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)",
			clinicID,
		).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff_reservation_exclusion", "")
	}
	return items, nil
}

func lockReservationJunctionWriteScope(
	tx *gorm.DB,
	clinicID, staffID uint64,
	reservationTypeIDs []uint64,
) error {
	if _, err := lockReservationStaffMutationScope(tx, clinicID, staffID); err != nil {
		return err
	}
	return lockReservationTypesForShare(tx, clinicID, reservationTypeIDs)
}

func lockReservationTypesForShare(
	tx *gorm.DB,
	clinicID uint64,
	reservationTypeIDs []uint64,
) error {
	if len(reservationTypeIDs) == 0 {
		return nil
	}
	sortedIDs := append([]uint64(nil), reservationTypeIDs...)
	sort.Slice(sortedIDs, func(i, j int) bool { return sortedIDs[i] < sortedIDs[j] })
	var reservationTypes []model.ReservationType
	if err := tx.
		Select("id").
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, sortedIDs).
		Order("id ASC").
		Find(&reservationTypes).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type", "")
	}
	if len(reservationTypes) != len(sortedIDs) {
		return apperrors.WrapInvalidInput(
			"reservation_type_ids contains invalid reservation type",
		)
	}
	return nil
}

// UpdateExcludedReservationTypes は staffID の除外コースを courseIDs で完全置換する（差分更新）
func (r *reservationStaffRepository) UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := lockReservationJunctionWriteScope(tx, clinicID, staffID, courseIDs); err != nil {
			return err
		}
		if err := persistence.DeleteJunctionViaMasterClinicScope(tx, clinicID, staffID,
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
		return persistence.InsertJunctionRowsInBatches(tx, items, "staff_reservation_exclusion", "")
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
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := lockReservationJunctionWriteScope(tx, clinicID, staffID, typeIDs); err != nil {
			return err
		}
		if err := persistence.DeleteJunctionByClinicAndStaff(tx, clinicID, staffID,
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
		return persistence.InsertJunctionRowsInBatches(tx, items, "staff_reservation_capability", "")
	}); err != nil {
		return apperrors.Wrap(err, "failed to replace capable reservation types")
	}
	return nil
}

func (r *reservationStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	var capability model.StaffReservationCapability
	db := persistence.DBOrTx(ctx, r.db).Model(&model.StaffReservationCapability{})
	if persistence.TxFromContext(ctx) != nil {
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
