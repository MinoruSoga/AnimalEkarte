// Package repository provides data access implementations for Staff entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Staff ----

type StaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	FindByID(ctx context.Context, id uint64) (*model.Staff, error)
	LockActiveByIDForUpdate(ctx context.Context, id uint64) (*model.Staff, error)
	LockActiveByIDForUpdateInClinic(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	LockActiveByIDForShare(ctx context.Context, id uint64) (*model.Staff, error)
	FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
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

func (r *staffRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	staffs := make([]model.Staff, 0)
	var total int64

	buildBase := func() *gorm.DB {
		q := dbOrTx(ctx, r.db).Model(&model.Staff{}).
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
	err := dbOrTx(ctx, r.db).
		Where("deleted_at IS NULL").
		Preload("Account", "deleted_at IS NULL").
		Preload("Occupation", "deleted_at IS NULL").
		First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
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
	if txFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("staff update lock requires an active transaction")
	}
	var staff model.Staff
	if err := dbOrTx(ctx, r.db).
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
	if txFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("scoped staff update lock requires an active transaction")
	}
	var staff model.Staff
	if err := dbOrTx(ctx, r.db).
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
	if txFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("staff share lock requires an active transaction")
	}
	var staff model.Staff
	if err := dbOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", id).
		First(&staff).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

func (r *staffRepository) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	var staff model.Staff
	err := dbOrTx(ctx, r.db).Where("deleted_at IS NULL").First(&staff, "account_id = ?", accountID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("account_id=%d", accountID))
	}
	return &staff, nil
}

func (r *staffRepository) Create(ctx context.Context, staff *model.Staff) error {
	if err := dbOrTx(ctx, r.db).Create(staff).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("staff", staff.Name)
		}
		return apperrors.FromGORM(err, "staff", "")
	}
	return nil
}

func (r *staffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := dbOrTx(ctx, r.db).
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
	result := dbOrTx(ctx, r.db).
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
	transactionErr := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Staff{}).
				Where("staffs.id = ?", id).
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
	// exams.entered_by カラムは存在しないため除外。
	checks := []struct {
		table   string
		column  string
		label   string
		softDel bool
	}{
		{table: "medical_records", column: "doctor_id", label: "カルテ", softDel: true},
		{table: "medical_records", column: "entered_by", label: "カルテ入力履歴", softDel: true},
		{table: "medical_record_addenda", column: "author_user_id", label: "カルテ追記", softDel: false},
		{table: "hospitalizations", column: "doctor_id", label: "入院記録", softDel: true},
		{table: "exams", column: "doctor_id", label: "検査", softDel: true},
		{table: "shift_entries", column: "staff_id", label: "シフト", softDel: false},
		{table: "billing_refunds", column: "refunded_by", label: "返金", softDel: false},
		{table: "cash_register_closes", column: "closed_by", label: "レジ締め", softDel: false},
		{table: "vital_records", column: "staff_id", label: "バイタル記録", softDel: true},
	}

	dependencies := make([]StaffDependencyCount, 0, len(checks)+2)
	for _, check := range checks {
		query := fmt.Sprintf("clinic_id = ? AND %s = ?", check.column)
		if check.softDel {
			query += " AND deleted_at IS NULL"
		}

		var count int64
		if err := dbOrTx(ctx, r.db).
			Table(check.table).
			Where(query, clinicID, staffID).
			Count(&count).Error; err != nil {
			return nil, apperrors.FromGORM(err, check.table, fmt.Sprintf("staff_id=%d", staffID))
		}
		if count > 0 {
			dependencies = append(dependencies, StaffDependencyCount{Label: check.label, Count: count})
		}
	}

	// payments は clinic_id を持たない。billings 経由で clinic_id をフィルタする。
	var paymentCount int64
	if err := dbOrTx(ctx, r.db).
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
// BE-refactor.md X-8: dbOrTx(ctx, r.db).Transaction(...) にすることで、ambient tx（例:
// reservationStaffService.Create の Transactor.WithTx）があればそのネスト tx（SAVEPOINT）
// として参加する。過去は r.db.WithContext(ctx).Transaction(...) で常に独立した新規 tx を
// 開始しており、ambient tx が UpdateExcludedReservationTypes の失敗で rollback しても
// 本メソッドの staff/assignment 作成は既にコミット済みのため巻き戻らなかった
// （除外コース未設定のまま孤児スタッフが残る部分コミットのバグ）。
func (r *staffRepository) CreateForReservation(ctx context.Context, staff *model.Staff, clinicID uint64) error {
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

// UpdateForReservation は予約用の staffs 更新（primary clinic_id スコープ・リソース名 reservation_staff）。
// BE-refactor.md X-8: dbOrTx(ctx, r.db) にすることで、reservationStaffService.Update が
// Transactor.WithTx で本メソッドと UpdateExcludedReservationTypes を括った場合に同一 tx へ参加する。
func (r *staffRepository) UpdateForReservation(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return updateScopedByID(ctx, dbOrTx(ctx, r.db), &model.Staff{}, "reservation_staff", clinicID, id, fields)
}

// SwapSortOrderForReservation は隣接スタッフと sort_order を入れ替える（予約画面の並び替え）。
// BE-refactor.md X-8: dbOrTx(ctx, r.db).Transaction(...) で ambient tx 参加を統一する
// （CreateForReservation/UpdateExcludedReservationTypes/UpdateReservationCapabilities と対称）。
func (r *staffRepository) SwapSortOrderForReservation(ctx context.Context, clinicID, id uint64, direction string) error {
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
