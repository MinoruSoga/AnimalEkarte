// Package repository provides data access implementations for Staff entity.
package staff

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ---- Staff ----

type StaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	FindByID(ctx context.Context, id uint64) (*model.Staff, error)
	FindByIDInClinic(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	LockActiveByIDForUpdate(ctx context.Context, id uint64) (*model.Staff, error)
	LockActiveByIDForUpdateInClinic(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	LockActiveByIDForShare(ctx context.Context, id uint64) (*model.Staff, error)
	FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
	// IsActiveSystemAdminStaff reports whether the staff can currently authenticate
	// as a system administrator (active staff + active system-admin account).
	IsActiveSystemAdminStaff(ctx context.Context, staffID uint64) (bool, error)
	// CountActiveSystemAdminStaff counts staff who can currently authenticate as
	// system administrators. Callers must hold the mutation transaction so the
	// count is consistent with concurrent deactivation attempts.
	CountActiveSystemAdminStaff(ctx context.Context) (int64, error)
	// Create はスタッフを作成する。
	Create(ctx context.Context, staff *model.Staff) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	UpdatePrimaryClinicID(ctx context.Context, id, clinicID uint64) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountBlockingReferencesByStaffID(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error)
	// --- 予約用途の staffs 書き込み（ADR-006 論点#1 案A: staffs テーブルの書き込みは
	// staff domain の exported メソッドへ一本化し、reservation 側は delegate 経由で呼ぶ）。
	// 既存 Create/Update/Reorder と意図的に別メソッド: エラーリソース名
	// ("reservation_staff")・スコープ機構（primary clinic_id vs assignment EXISTS）・
	// tx 構成（main assignment 同時作成 / 隣接 swap）が異なり、統合は挙動変更になる。
	CreateForReservation(ctx context.Context, staff *model.Staff, clinicID uint64) error
	UpdateForReservation(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	SwapSortOrderForReservation(ctx context.Context, clinicID, id uint64, direction string) error
}

type StaffDependencyCount struct {
	Label string
	Count int64
}

type staffRepository struct{ db *gorm.DB }

func NewStaffRepository(db *gorm.DB) StaffRepository { return &staffRepository{db: db} }

func paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((page - 1) * limit).Limit(limit)
	}
}

func (r *staffRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	staffs := make([]model.Staff, 0)
	var total int64

	buildBase := func() *gorm.DB {
		q := persistence.DBOrTx(ctx, r.db).Model(&model.Staff{}).
			Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id"+
				" AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", clinicID).
			Where("staffs.deleted_at IS NULL")
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "staff", "")
	}
	if err := buildBase().
		Preload("Account", "deleted_at IS NULL").
		Preload("Occupation", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(paginate(page, limit)).
		Order("staffs.sort_order ASC, staffs.name ASC").
		Distinct("staffs.*").
		Find(&staffs).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "staff", "")
	}
	return staffs, total, nil
}

func (r *staffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	var staff model.Staff
	db := persistence.DBOrTx(ctx, r.db)
	err := db.
		Where("deleted_at IS NULL").
		Preload("Account", "deleted_at IS NULL").
		First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	if staff.OccupationID == nil {
		return &staff, nil
	}
	var occupation model.Occupation
	err = db.
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *staff.OccupationID, staff.ClinicID).
		First(&occupation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		staff.OccupationID = nil
		return &staff, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "occupation", fmt.Sprintf("%d", *staff.OccupationID))
	}
	staff.Occupation = &occupation
	return &staff, nil
}

// FindByIDInClinic returns the HTTP-facing staff representation only when the
// staff has an active assignment to the authenticated clinic. Occupation is a
// clinic-owned master; when the global staff identity currently points at
// another clinic's occupation, both the association and FK are omitted.
func (r *staffRepository) FindByIDInClinic(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	var staff model.Staff
	err := persistence.DBOrTx(ctx, r.db).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		Where(
			"EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)",
			clinicID,
		).
		Preload("Account", "deleted_at IS NULL").
		Preload("Occupation", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		First(&staff).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	if staff.OccupationID != nil && staff.Occupation == nil {
		staff.OccupationID = nil
	}
	return &staff, nil
}

// LockActiveByIDForUpdate locks a non-deleted staff identity for a mutation.
// The caller must already own the transaction so the lock lifetime covers all
// following authorization, dependency checks, and writes.
func (r *staffRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	id uint64,
) (*model.Staff, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("staff update lock requires an active transaction")
	}
	var staff model.Staff
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		First(&staff).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

// LockActiveByIDForUpdateInClinic locks a non-deleted staff identity only when
// it has an active assignment to the authenticated clinic. This prevents a
// cross-clinic identifier from acquiring or disclosing another clinic's row.
func (r *staffRepository) LockActiveByIDForUpdateInClinic(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("scoped staff update lock requires an active transaction")
	}
	var staff model.Staff
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		Where(
			"EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)",
			clinicID,
		).
		First(&staff).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

// LockActiveByIDForShare locks a non-deleted staff identity for a dependent
// write without blocking other readers. Writers that delete or replace the
// staff's assignments must take the update lock first, preserving lock order.
func (r *staffRepository) LockActiveByIDForShare(
	ctx context.Context,
	id uint64,
) (*model.Staff, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("staff share lock requires an active transaction")
	}
	var staff model.Staff
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		First(&staff).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

func (r *staffRepository) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	var staff model.Staff
	err := persistence.DBOrTx(ctx, r.db).Where("deleted_at IS NULL").First(&staff, "account_id = ?", accountID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("account_id=%d", accountID))
	}
	return &staff, nil
}

// activeSystemAdminStaffQuery is the shared predicate for staff who can currently
// authenticate as system administrators.
func (r *staffRepository) activeSystemAdminStaffQuery(ctx context.Context) *gorm.DB {
	return persistence.DBOrTx(ctx, r.db).Model(&model.Staff{}).
		Joins("INNER JOIN accounts ON accounts.id = staffs.account_id AND accounts.deleted_at IS NULL").
		Where("staffs.deleted_at IS NULL").
		Where("staffs.is_active = TRUE").
		Where("accounts.is_active = TRUE").
		Where("accounts.is_system_admin = TRUE")
}

// IsActiveSystemAdminStaff reports whether the staff can currently authenticate
// as a system administrator.
func (r *staffRepository) IsActiveSystemAdminStaff(ctx context.Context, staffID uint64) (bool, error) {
	var count int64
	err := r.activeSystemAdminStaffQuery(ctx).
		Where("staffs.id = ?", staffID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", staffID))
	}
	return count > 0, nil
}

// CountActiveSystemAdminStaff counts staff who can currently authenticate as
// system administrators. When an ambient transaction is present, matching staff
// rows are locked FOR UPDATE in deterministic order so concurrent deactivations
// serialize on the last-admin check.
func (r *staffRepository) CountActiveSystemAdminStaff(ctx context.Context) (int64, error) {
	db := r.activeSystemAdminStaffQuery(ctx)
	if persistence.TxFromContext(ctx) != nil {
		var ids []uint64
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Order("staffs.id ASC").
			Pluck("staffs.id", &ids).Error; err != nil {
			return 0, apperrors.FromGORM(err, "staff", "active_system_admin_count")
		}
		return int64(len(ids)), nil
	}
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "staff", "active_system_admin_count")
	}
	return count, nil
}

func (r *staffRepository) Create(ctx context.Context, staff *model.Staff) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantVisible := staff.ReservationVisible
	if err := db.Create(staff).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("staff", staff.Name)
		}
		return apperrors.FromGORM(err, "staff", "")
	}
	if !wantVisible {
		if err := db.Model(staff).Update("reservation_visible", false).Error; err != nil {
			return apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", staff.ID))
		}
		staff.ReservationVisible = false
	}
	return nil
}

func (r *staffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Staff{}).
		Where("staffs.id = ?", id).
		Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)", clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *staffRepository) UpdatePrimaryClinicID(ctx context.Context, id, clinicID uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Staff{}).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)", clinicID).
		Update("clinic_id", clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *staffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	staffID := fmt.Sprintf("%d", id)
	var operationErr error
	transactionErr := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Lock the scoped identity and its active assignments so the multi-clinic
		// guard and soft delete observe the same state.
		var staff model.Staff
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("staffs.id = ?", id).
			Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)", clinicID).
			First(&staff).Error; err != nil {
			operationErr = apperrors.FromGORM(err, "staff", staffID)
			return operationErr
		}

		var assignments []model.StaffClinicAssignment
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("staff_id = ?", id).
			Find(&assignments).Error; err != nil {
			operationErr = apperrors.FromGORM(err, "staff_clinic_assignment", "staff_id="+staffID)
			return operationErr
		}
		if len(assignments) > 1 {
			operationErr = apperrors.WrapConflict("複数のクリニックに所属しているスタッフは削除できません")
			return operationErr
		}

		result := tx.
			Model(&model.Staff{}).
			Where("id = ?", id).
			Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)", clinicID).
			Update("deleted_at", gorm.Expr("now()"))
		if result.Error != nil {
			operationErr = apperrors.FromGORM(result.Error, "staff", staffID)
			return operationErr
		}
		if result.RowsAffected == 0 {
			operationErr = apperrors.WrapNotFound("staff", staffID)
			return operationErr
		}
		return nil
	})
	if transactionErr != nil {
		if operationErr != nil {
			return operationErr
		}
		return apperrors.FromGORM(transactionErr, "staff", staffID)
	}
	return nil
}

func (r *staffRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Staff{}).
				Where("staffs.id = ? AND staffs.clinic_id = ?", id, clinicID).
				Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)", clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("staff id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to reorder staffs")
	}
	return nil
}

func (r *staffRepository) CountBlockingReferencesByStaffID(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error) {
	// clinic_id カラムを直接持つテーブルのみ汎用ループで処理。
	// payments は clinic_id を持たないため後述の特殊クエリで対応。
	// medical_record_addenda / vital_records は子 clinic_id を持つが親業務行と
	// 食い違いうるため、親 clinic 相関を別クエリで行う（SEC-SWEEP-02-STAFF-B1）。
	// exams.entered_by カラムは存在しないため除外。
	checks := []struct {
		table   string
		column  string
		label   string
		softDel bool
	}{
		{table: "medical_records", column: "doctor_id", label: "カルテ", softDel: true},
		{table: "medical_records", column: "entered_by", label: "カルテ入力履歴", softDel: true},
		{table: "hospitalizations", column: "doctor_id", label: "入院記録", softDel: true},
		{table: "exams", column: "doctor_id", label: "検査", softDel: true},
		{table: "shift_entries", column: "staff_id", label: "シフト", softDel: false},
		{table: "billing_refunds", column: "refunded_by", label: "返金", softDel: false},
		{table: "cash_register_closes", column: "closed_by", label: "レジ締め", softDel: false},
	}

	dependencies := make([]StaffDependencyCount, 0, len(checks)+4)
	for _, check := range checks {
		query := fmt.Sprintf("clinic_id = ? AND %s = ?", check.column)
		if check.softDel {
			query += " AND deleted_at IS NULL"
		}

		var count int64
		if err := persistence.DBOrTx(ctx, r.db).
			Table(check.table).
			Where(query, clinicID, staffID).
			Count(&count).Error; err != nil {
			return nil, apperrors.FromGORM(err, check.table, fmt.Sprintf("staff_id=%d", staffID))
		}
		if count > 0 {
			dependencies = append(dependencies, StaffDependencyCount{Label: check.label, Count: count})
		}
	}

	// medical_record_addenda: 親 medical_records の clinic 相関のみ（deleted_at は付けない）。
	// 修飾列を使い JOIN 時の SQLSTATE 42702（ambiguous clinic_id）を避ける。
	var addendaCount int64
	if err := persistence.DBOrTx(ctx, r.db).
		Table("medical_record_addenda").
		Joins("JOIN medical_records ON medical_records.id = medical_record_addenda.medical_record_id AND medical_records.clinic_id = medical_record_addenda.clinic_id").
		Where("medical_record_addenda.clinic_id = ? AND medical_record_addenda.author_user_id = ?", clinicID, staffID).
		Count(&addendaCount).Error; err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_addenda", fmt.Sprintf("staff_id=%d", staffID))
	}
	if addendaCount > 0 {
		dependencies = append(dependencies, StaffDependencyCount{Label: "カルテ追記", Count: addendaCount})
	}

	// vital_records: 親 medical_records または daily_records の clinic 相関のみ。
	// 外来/入院のどちらか一方でも clinic が一致すれば有効な依存として数える。
	// parent deleted_at / pets.deceased_at は付けない（MR-B1/TRIM-B1 規約）。
	var vitalCount int64
	if err := persistence.DBOrTx(ctx, r.db).
		Table("vital_records").
		Where("vital_records.clinic_id = ? AND vital_records.staff_id = ? AND vital_records.deleted_at IS NULL", clinicID, staffID).
		Where(`(
			(vital_records.medical_record_id IS NOT NULL AND EXISTS (
				SELECT 1 FROM medical_records
				WHERE medical_records.id = vital_records.medical_record_id
				  AND medical_records.clinic_id = vital_records.clinic_id
			))
			OR
			(vital_records.daily_record_id IS NOT NULL AND EXISTS (
				SELECT 1 FROM daily_records
				WHERE daily_records.id = vital_records.daily_record_id
				  AND daily_records.clinic_id = vital_records.clinic_id
			))
		)`).
		Count(&vitalCount).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vital_records", fmt.Sprintf("staff_id=%d", staffID))
	}
	if vitalCount > 0 {
		dependencies = append(dependencies, StaffDependencyCount{Label: "バイタル記録", Count: vitalCount})
	}

	// payments は clinic_id を持たない。billings 経由で clinic_id をフィルタする。
	var paymentCount int64
	if err := persistence.DBOrTx(ctx, r.db).
		Table("payments").
		Joins("INNER JOIN billings ON billings.id = payments.billing_id AND billings.clinic_id = ?", clinicID).
		Where("payments.paid_by = ? AND payments.deleted_at IS NULL", staffID).
		Count(&paymentCount).Error; err != nil {
		return nil, apperrors.FromGORM(err, "payments", fmt.Sprintf("staff_id=%d", staffID))
	}
	if paymentCount > 0 {
		dependencies = append(dependencies, StaffDependencyCount{Label: "支払い", Count: paymentCount})
	}

	return dependencies, nil
}

// ---- 予約用途の staffs 書き込み（ADR-006 論点#1 案A で reservation_staff_repository.go から移動） ----

// CreateForReservation はスタッフ + StaffClinicAssignment をトランザクションで作成する。
// BE-refactor.md X-8: persistence.DBOrTx(ctx, r.db).Transaction(...) にすることで、ambient tx（例:
// reservationStaffService.Create の Transactor.WithTx）があればそのネスト tx（SAVEPOINT）
// として参加する。過去は r.db.WithContext(ctx).Transaction(...) で常に独立した新規 tx を
// 開始しており、ambient tx が UpdateExcludedReservationTypes の失敗で rollback しても
// 本メソッドの staff/assignment 作成は既にコミット済みのため巻き戻らなかった
// （除外コース未設定のまま孤児スタッフが残る部分コミットのバグ）。
func (r *staffRepository) CreateForReservation(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	if staff == nil {
		return apperrors.WrapInvalidInput("reservation staff is required")
	}
	if staff.ClinicID != clinicID {
		return apperrors.WrapInvalidInput("reservation staff clinic_id mismatch")
	}
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantVisible := staff.ReservationVisible
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(staff).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_staff", "")
		}
		if !wantVisible {
			if err := tx.Model(staff).Update("reservation_visible", false).Error; err != nil {
				return apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", staff.ID))
			}
			staff.ReservationVisible = false
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

// UpdateForReservation は予約用の staffs 更新（primary clinic_id スコープ・リソース名 reservation_staff）。
// BE-refactor.md X-8: persistence.DBOrTx(ctx, r.db) にすることで、reservationStaffService.Update が
// Transactor.WithTx で本メソッドと UpdateExcludedReservationTypes を括った場合に同一 tx へ参加する。
func (r *staffRepository) UpdateForReservation(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.Staff{}, "reservation_staff", clinicID, id, fields)
}

// SwapSortOrderForReservation は隣接スタッフと sort_order を入れ替える（予約画面の並び替え）。
// BE-refactor.md X-8: persistence.DBOrTx(ctx, r.db).Transaction(...) で ambient tx 参加を統一する
// （CreateForReservation/UpdateExcludedReservationTypes/UpdateReservationCapabilities と対称）。
func (r *staffRepository) SwapSortOrderForReservation(ctx context.Context, clinicID, id uint64, direction string) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		locked, err := lockReservationStaffOrder(tx, clinicID)
		if err != nil {
			return err
		}
		target, neighbor, err := selectReservationStaffSwap(locked, id, direction)
		if err != nil || neighbor == nil {
			return err
		}
		if err := updateReservationStaffSortOrder(tx, clinicID, target.ID, neighbor.SortOrder); err != nil {
			return err
		}
		if err := updateReservationStaffSortOrder(tx, clinicID, neighbor.ID, target.SortOrder); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to swap sort order")
	}
	return nil
}

func lockReservationStaffOrder(tx *gorm.DB, clinicID uint64) ([]model.Staff, error) {
	var locked []model.Staff
	err := tx.
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).
		Where("staffs.clinic_id = ? AND staffs.deleted_at IS NULL", clinicID).
		Where(
			"EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)",
			clinicID,
		).
		Order("staffs.id ASC").
		Find(&locked).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_staff", "sort_order_lock")
	}
	return locked, nil
}

func selectReservationStaffSwap(
	locked []model.Staff,
	id uint64,
	direction string,
) (model.Staff, *model.Staff, error) {
	if direction != "up" && direction != "down" {
		return model.Staff{}, nil, apperrors.WrapInvalidInput("direction must be 'up' or 'down'")
	}
	ordered := append([]model.Staff(nil), locked...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].SortOrder == ordered[j].SortOrder {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].SortOrder < ordered[j].SortOrder
	})
	for index := range ordered {
		if ordered[index].ID != id {
			continue
		}
		neighborIndex := index + 1
		if direction == "up" {
			neighborIndex = index - 1
		}
		if neighborIndex < 0 || neighborIndex >= len(ordered) {
			return ordered[index], nil, nil
		}
		neighbor := ordered[neighborIndex]
		return ordered[index], &neighbor, nil
	}
	return model.Staff{}, nil, apperrors.WrapNotFound("reservation_staff", fmt.Sprintf("%d", id))
}

func updateReservationStaffSortOrder(
	tx *gorm.DB,
	clinicID, staffID uint64,
	sortOrder int,
) error {
	result := tx.Model(&model.Staff{}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		Where(
			"EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)",
			clinicID,
		).
		Update("sort_order", sortOrder)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_staff", fmt.Sprintf("%d", staffID))
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapNotFound("reservation_staff", fmt.Sprintf("%d", staffID))
	}
	return nil
}
