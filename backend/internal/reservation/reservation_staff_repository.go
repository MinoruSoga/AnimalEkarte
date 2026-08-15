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

// FindAllExcludedReservationTypes derives the exclusion compatibility facade
// from staff_reservation_capabilities (TASK-021 Stage B sole write SoT).
// excluded = clinic active universe \ capable. Legacy exclusion table rows are ignored.
func (r *reservationStaffRepository) FindAllExcludedReservationTypes(
	ctx context.Context,
	clinicID, staffID uint64,
) ([]model.StaffReservationExclusion, error) {
	universe, err := r.listActiveReservationTypeUniverse(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	if len(universe) == 0 {
		return nil, nil
	}
	// Shared-staff isolation: staff without an active assignment in this clinic
	// must not receive derived exclusions for this clinic's universe.
	assigned, err := r.hasActiveClinicAssignment(ctx, clinicID, staffID)
	if err != nil {
		return nil, err
	}
	if !assigned {
		return nil, nil
	}
	caps, err := r.FindAllReservationCapabilities(ctx, clinicID, staffID)
	if err != nil {
		return nil, err
	}
	capableIDs := make([]uint64, 0, len(caps))
	for _, c := range caps {
		capableIDs = append(capableIDs, c.ReservationTypeID)
	}
	return deriveExcludedFromUniverse(staffID, universe, capableIDs), nil
}

// FindAllExcludedReservationTypesByStaffIDs derives exclusion facade rows for
// many staff in one capability bulk read (TASK-021 Stage B).
func (r *reservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(
	ctx context.Context,
	clinicID uint64,
	staffIDs []uint64,
) ([]model.StaffReservationExclusion, error) {
	if len(staffIDs) == 0 {
		return nil, nil
	}
	universe, err := r.listActiveReservationTypeUniverse(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	if len(universe) == 0 {
		return nil, nil
	}
	// Match single-staff path: only staff with active assignment in this clinic.
	assignedIDs, err := r.filterStaffIDsWithActiveAssignment(ctx, clinicID, staffIDs)
	if err != nil {
		return nil, err
	}
	if len(assignedIDs) == 0 {
		return nil, nil
	}
	caps, err := r.FindAllReservationCapabilitiesByStaffIDs(ctx, clinicID, assignedIDs)
	if err != nil {
		return nil, err
	}
	byStaff := make(map[uint64][]uint64, len(assignedIDs))
	for _, c := range caps {
		byStaff[c.StaffID] = append(byStaff[c.StaffID], c.ReservationTypeID)
	}
	out := make([]model.StaffReservationExclusion, 0)
	for _, staffID := range assignedIDs {
		out = append(out, deriveExcludedFromUniverse(staffID, universe, byStaff[staffID])...)
	}
	return out, nil
}

func (r *reservationStaffRepository) filterStaffIDsWithActiveAssignment(
	ctx context.Context,
	clinicID uint64,
	staffIDs []uint64,
) ([]uint64, error) {
	if len(staffIDs) == 0 {
		return nil, nil
	}
	var rows []model.StaffClinicAssignment
	err := persistence.DBOrTx(ctx, r.db).
		Select("staff_id").
		Where("clinic_id = ? AND staff_id IN ? AND deleted_at IS NULL", clinicID, staffIDs).
		Find(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff_clinic_assignment", "")
	}
	out := make([]uint64, 0, len(rows))
	seen := make(map[uint64]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.StaffID]; ok {
			continue
		}
		seen[row.StaffID] = struct{}{}
		out = append(out, row.StaffID)
	}
	return out, nil
}

// listActiveReservationTypeUniverse returns clinic-scoped active non-deleted
// reservation types ordered by id (Stage B inverse-mapping universe).
func (r *reservationStaffRepository) listActiveReservationTypeUniverse(
	ctx context.Context,
	clinicID uint64,
) ([]model.ReservationType, error) {
	var types []model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).
		Select("id", "clinic_id", "name", "is_active").
		Where("clinic_id = ? AND deleted_at IS NULL AND is_active = ?", clinicID, true).
		Order("id ASC").
		Find(&types).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", "")
	}
	return types, nil
}

func (r *reservationStaffRepository) hasActiveClinicAssignment(
	ctx context.Context,
	clinicID, staffID uint64,
) (bool, error) {
	var n int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		Count(&n).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "staff_clinic_assignment", "")
	}
	return n > 0, nil
}

func deriveExcludedFromUniverse(
	staffID uint64,
	universe []model.ReservationType,
	capableIDs []uint64,
) []model.StaffReservationExclusion {
	universeIDs := make([]uint64, len(universe))
	byID := make(map[uint64]*model.ReservationType, len(universe))
	for i := range universe {
		universeIDs[i] = universe[i].ID
		byID[universe[i].ID] = &universe[i]
	}
	excludedIDs := excludedIDsFromCapable(universeIDs, capableIDs)
	out := make([]model.StaffReservationExclusion, 0, len(excludedIDs))
	for _, id := range excludedIDs {
		item := model.StaffReservationExclusion{
			StaffID:           staffID,
			ReservationTypeID: id,
		}
		if rt := byID[id]; rt != nil {
			copyRT := *rt
			item.ReservationType = &copyRT
		}
		out = append(out, item)
	}
	return out
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

// UpdateExcludedReservationTypes is a Stage B compatibility write facade.
// It converts excluded_type_ids into one atomic capability replacement
// (capable = active universe \ excluded) and does not write staff_reservation_exclusions.
func (r *reservationStaffRepository) UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Validate requested exclusion IDs belong to this clinic (non-deleted).
		if err := lockReservationJunctionWriteScope(tx, clinicID, staffID, courseIDs); err != nil {
			return err
		}
		// Load inverse-mapping universe under the same transaction.
		var universe []model.ReservationType
		if err := tx.
			Select("id").
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("clinic_id = ? AND deleted_at IS NULL AND is_active = ?", clinicID, true).
			Order("id ASC").
			Find(&universe).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type", "")
		}
		universeIDs := make([]uint64, len(universe))
		for i := range universe {
			universeIDs[i] = universe[i].ID
		}
		capableIDs := capableIDsFromExcluded(universeIDs, courseIDs)
		// Replace capabilities only — zero production writes to exclusions table.
		if err := replaceReservationCapabilitiesTx(tx, clinicID, staffID, capableIDs); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to replace excluded reservation types via capability facade")
	}
	return nil
}

func (r *reservationStaffRepository) FindAllReservationCapabilities(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error) {
	var items []model.StaffReservationCapability
	// DBOrTx: Stage B exclusion facade readback after capability replace must see
	// uncommitted rows in the same ambient transaction.
	err := persistence.DBOrTx(ctx, r.db).
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
	err := persistence.DBOrTx(ctx, r.db).
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
		return replaceReservationCapabilitiesTx(tx, clinicID, staffID, typeIDs)
	}); err != nil {
		return apperrors.Wrap(err, "failed to replace capable reservation types")
	}
	return nil
}

// replaceReservationCapabilitiesTx assumes ambient tx and that junction write scope
// (staff/assignment/types) is already locked by the caller when required.
func replaceReservationCapabilitiesTx(tx *gorm.DB, clinicID, staffID uint64, typeIDs []uint64) error {
	// When called from exclusion facade, typeIDs may differ from already-locked
	// exclusion IDs — lock capable IDs for share as well.
	if err := lockReservationTypesForShare(tx, clinicID, typeIDs); err != nil {
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
