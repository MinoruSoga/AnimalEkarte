// Package staff owns shift_entries / shift_entry_breaks data access.
package staff

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// staffAssignedToClinicCond keeps the shift entry visible while hiding a Staff
// association that does not have an active assignment to the requested clinic.
const staffAssignedToClinicCond = "deleted_at IS NULL AND EXISTS (SELECT 1 FROM staff_clinic_assignments sca WHERE sca.staff_id = staffs.id AND sca.clinic_id = ? AND sca.deleted_at IS NULL)"

// Filter はシフト一覧取得のフィルタ条件
type ShiftEntryFilter struct {
	YearMonth string // "YYYY-MM" 形式
	StaffID   *uint64
}

// Repository はシフト管理のデータアクセスインターフェース
type ShiftEntryRepository interface {
	FindAll(ctx context.Context, clinicID uint64, filter ShiftEntryFilter) ([]model.ShiftEntry, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error)
	LockActiveByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error)
	Create(ctx context.Context, entry *model.ShiftEntry) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error)
	FindClinicIDsByStaffID(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error)
	ReplaceBreaks(ctx context.Context, shiftEntryID uint64, breaks []model.ShiftEntryBreak) error
	// FindOnDutyStaffs は指定日にシフトが登録されているスタッフ一覧を返す (BUG-344)
	FindOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error)
	// --- 予約スケジュール用途の書き込み（ADR-006 論点#1 案A: shift_entries の書き込みは
	// staff domain（本package）の exported メソッドへ一本化し、reservation 側は delegate 経由で呼ぶ）。
	// 既存 Create/Update/Delete/ReplaceBreaks と意図的に別メソッド: (staff_id, date) キーの
	// upsert + breaks 全置換を単一 tx で行う予約画面固有の contract のため。
	SaveByStaffDate(
		ctx context.Context,
		clinicID uint64,
		entry *model.ShiftEntry,
		breaks []model.ShiftEntryBreak,
	) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error)
	DeleteByStaffDate(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

type shiftEntryRepository struct{ db *gorm.DB }

// New は Repository を初期化して返す
func NewShiftEntryRepository(db *gorm.DB) ShiftEntryRepository {
	return &shiftEntryRepository{db: db}
}

// isUniqueConstraintErr はPostgreSQLのユニーク制約違反（23505）を判定する
// （親 repository パッケージの同名ヘルパーの最小限の複製。repohelpers 未収載のため import cycle 回避のためローカル定義。BE8-4 batch13）。
func isUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *shiftEntryRepository) FindAll(ctx context.Context, clinicID uint64, filter ShiftEntryFilter) ([]model.ShiftEntry, error) {
	q := persistence.DBOrTx(ctx, r.db).
		Preload("Staff", staffAssignedToClinicCond, clinicID).
		Preload("Breaks").
		Scopes(persistence.ClinicScope(clinicID)).
		Order("date ASC, staff_id ASC")

	if filter.YearMonth != "" {
		// YYYY-MM → start/end dates
		t, err := time.ParseInLocation("2006-01", filter.YearMonth, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid year_month format: %s", filter.YearMonth))
		}
		start := t
		end := t.AddDate(0, 1, 0)
		q = q.Where("date >= ? AND date < ?", start.Format(time.DateOnly), end.Format(time.DateOnly))
	}
	if filter.StaffID != nil {
		q = q.Where("staff_id = ?", *filter.StaffID)
	}

	entries := make([]model.ShiftEntry, 0)
	if err := q.Find(&entries).Error; err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", "")
	}
	return entries, nil
}

func (r *shiftEntryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error) {
	var entry model.ShiftEntry
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Staff", staffAssignedToClinicCond, clinicID).
		Preload("Breaks").
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		First(&entry).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", strconv.FormatUint(id, 10))
	}
	return &entry, nil
}

func (r *shiftEntryRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.ShiftEntry, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError(
			"shift entry update lock requires an active transaction",
		)
	}
	var entry model.ShiftEntry
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&entry).Error; err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", strconv.FormatUint(id, 10))
	}
	return &entry, nil
}

func (r *shiftEntryRepository) Create(ctx context.Context, entry *model.ShiftEntry) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(entry).Error; err != nil {
		// PostgreSQL UNIQUE違反 (23505)
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("shift_entry",
				fmt.Sprintf("staff_id=%d date=%s", entry.StaffID, entry.Date.Format(time.DateOnly)))
		}
		return apperrors.FromGORM(err, "shift_entry", "")
	}
	return nil
}

func (r *shiftEntryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.ShiftEntry{}, "shift_entry", clinicID, id, fields)
}

func (r *shiftEntryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.ShiftEntry{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "shift_entry", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_entry", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *shiftEntryRepository) ReplaceBreaks(ctx context.Context, shiftEntryID uint64, breaks []model.ShiftEntryBreak) error {
	return persistence.ReplaceChildRowsByParentID(
		persistence.DBOrTx(ctx, r.db),
		shiftEntryID,
		breaks,
		&model.ShiftEntryBreak{},
		"shift_entry_id",
		"shift_entry_break",
		func(row *model.ShiftEntryBreak, id uint64) { row.ShiftEntryID = id },
		"failed to replace shift entry breaks",
	)
}

func (r *shiftEntryRepository) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.ShiftEntry{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("staff_id = ?", staffID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "shift_entry", "")
	}
	return count > 0, nil
}

func (r *shiftEntryRepository) FindClinicIDsByStaffID(
	ctx context.Context,
	clinicIDs []uint64,
	staffID uint64,
) ([]uint64, error) {
	result := make([]uint64, 0)
	if len(clinicIDs) == 0 {
		return result, nil
	}

	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.ShiftEntry{}).
		Scopes(persistence.ClinicScopeIn(clinicIDs)).
		Where("staff_id = ?", staffID).
		Distinct().
		Order("clinic_id ASC").
		Pluck("clinic_id", &result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", "")
	}
	return result, nil
}

// FindOnDutyStaffs は指定日にシフトが登録されているスタッフ一覧を返す (BUG-344)
func (r *shiftEntryRepository) FindOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error) {
	var staffs []model.Staff
	dateStr := date.Format(time.DateOnly)
	// shift_entries テーブルは deleted_at カラムを持たない（論理削除なし）
	// 勤務日の絞り込みは JOIN 条件の shift_entries.clinic_id で担保する。
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN shift_entries ON shift_entries.staff_id = staffs.id"+
			" AND shift_entries.clinic_id = ?"+
			" AND DATE(shift_entries.date) = ?", clinicID, dateStr).
		Where("staffs.deleted_at IS NULL AND staffs.is_active = true").
		Where(
			"EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL)",
			clinicID,
		).
		Distinct("staffs.*").
		Find(&staffs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "on_duty_staffs", dateStr)
	}
	return staffs, nil
}

// ---- 予約スケジュール用途の書き込み（ADR-006 論点#1 案A で reservation_schedule_repository.go から移動） ----

func lockShiftScheduleMutationScope(
	tx *gorm.DB,
	clinicID, staffID uint64,
) error {
	var staff model.Staff
	if err := tx.
		Select("id").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", staffID).
		First(&staff).Error; err != nil {
		return apperrors.FromGORM(err, "staff", strconv.FormatUint(staffID, 10))
	}

	var assignment model.StaffClinicAssignment
	if err := tx.
		Select("id").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		First(&assignment).Error; err != nil {
		return apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID),
		)
	}
	return nil
}

func lockShiftEntryByStaffDateForUpdate(
	tx *gorm.DB,
	clinicID, staffID uint64,
	date time.Time,
) (*model.ShiftEntry, error) {
	var existing model.ShiftEntry
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("staff_id = ? AND date = ?", staffID, date.Format(time.DateOnly)).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(
			err,
			"shift_entry",
			fmt.Sprintf("staff=%d date=%s", staffID, date.Format(time.DateOnly)),
		)
	}
	return &existing, nil
}

func replaceScheduleBreaks(
	tx *gorm.DB,
	clinicID, shiftEntryID uint64,
	breaks []model.ShiftEntryBreak,
) ([]model.ShiftEntryBreak, error) {
	// RSV-08: parent clinic correlation for break writes/re-reads (string literals for grandchild lint).
	if err := tx.
		Where("shift_entry_id = ?", shiftEntryID).
		Where(
			"EXISTS (SELECT 1 FROM shift_entries WHERE shift_entries.id = shift_entry_breaks.shift_entry_id AND shift_entries.clinic_id = ?)",
			clinicID,
		).
		Delete(&model.ShiftEntryBreak{}).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"shift_entry_break",
			strconv.FormatUint(shiftEntryID, 10),
		)
	}
	if len(breaks) == 0 {
		return []model.ShiftEntryBreak{}, nil
	}

	replacements := make([]model.ShiftEntryBreak, len(breaks))
	for i := range breaks {
		replacements[i] = breaks[i]
		replacements[i].ID = 0
		replacements[i].ShiftEntryID = shiftEntryID
	}
	if err := tx.Create(&replacements).Error; err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry_break", "")
	}

	persisted := make([]model.ShiftEntryBreak, 0, len(replacements))
	if err := tx.
		Model(&model.ShiftEntryBreak{}).
		Joins("JOIN shift_entries ON shift_entries.id = shift_entry_breaks.shift_entry_id AND shift_entries.clinic_id = ?", clinicID).
		Where("shift_entry_breaks.shift_entry_id = ?", shiftEntryID).
		Order("shift_entry_breaks.id ASC").
		Find(&persisted).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"shift_entry_break",
			strconv.FormatUint(shiftEntryID, 10),
		)
	}
	return persisted, nil
}

// SaveByStaffDate atomically upserts a shift graph and returns the persisted
// parent, persisted breaks, and whether this transaction created the parent.
// The exclusive staff -> assignment -> shift order also serializes absent keys.
func (r *shiftEntryRepository) SaveByStaffDate(
	ctx context.Context,
	clinicID uint64,
	entry *model.ShiftEntry,
	breaks []model.ShiftEntryBreak,
) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
	if entry == nil {
		return nil, nil, false, apperrors.WrapInvalidInput("shift entry is required")
	}
	// fail-closed: スコープ述語は entry.ClinicID を使うため、認証済み clinicID との
	// 不一致は書込前に拒否する（将来の呼び出し元が非認証由来の値を渡す事故を封じる）。
	if entry.ClinicID != clinicID {
		return nil, nil, false, apperrors.WrapInvalidInput("shift entry clinic_id mismatch")
	}

	var saved model.ShiftEntry
	var savedBreaks []model.ShiftEntryBreak
	created := false
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := lockShiftScheduleMutationScope(tx, clinicID, entry.StaffID); err != nil {
			return err
		}
		existing, err := lockShiftEntryByStaffDateForUpdate(
			tx,
			clinicID,
			entry.StaffID,
			entry.Date,
		)
		if err != nil {
			return err
		}

		if existing == nil {
			saved = model.ShiftEntry{
				ClinicID:  clinicID,
				StaffID:   entry.StaffID,
				Date:      entry.Date,
				ShiftType: entry.ShiftType,
				StartTime: entry.StartTime,
				EndTime:   entry.EndTime,
				Notes:     entry.Notes,
			}
			if err := tx.Omit("Staff", "Breaks").Create(&saved).Error; err != nil {
				return apperrors.FromGORM(err, "shift_entry", "")
			}
			created = true
		} else {
			fields := map[string]any{
				"shift_type": entry.ShiftType,
				"start_time": entry.StartTime,
				"end_time":   entry.EndTime,
				"notes":      entry.Notes,
				"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}
			result := tx.
				Scopes(persistence.ClinicScope(clinicID)).
				Model(&model.ShiftEntry{}).
				Where("id = ?", existing.ID).
				Updates(fields)
			if result.Error != nil {
				return apperrors.FromGORM(
					result.Error,
					"shift_entry",
					strconv.FormatUint(existing.ID, 10),
				)
			}
			if result.RowsAffected != 1 {
				return apperrors.WrapNotFound(
					"shift_entry",
					strconv.FormatUint(existing.ID, 10),
				)
			}
			saved.ID = existing.ID
		}
		if err := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", saved.ID).
			First(&saved).Error; err != nil {
			return apperrors.FromGORM(
				err,
				"shift_entry",
				strconv.FormatUint(saved.ID, 10),
			)
		}
		savedBreaks, err = replaceScheduleBreaks(tx, clinicID, saved.ID, breaks)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, false, apperrors.Wrap(err, "failed to upsert shift entry")
	}
	return &saved, savedBreaks, created, nil
}

// DeleteByStaffDate は (staff_id, date) キーで ShiftEntry を削除する（clinic_id スコープ付き）。
func (r *shiftEntryRepository) DeleteByStaffDate(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	resourceID := fmt.Sprintf("staff=%d date=%s", staffID, date.Format(time.DateOnly))
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := lockShiftScheduleMutationScope(tx, clinicID, staffID); err != nil {
			return err
		}
		existing, err := lockShiftEntryByStaffDateForUpdate(tx, clinicID, staffID, date)
		if err != nil {
			return err
		}
		if existing == nil {
			return apperrors.WrapNotFound("shift_entry", resourceID)
		}
		result := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", existing.ID).
			Delete(&model.ShiftEntry{})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "schedule_entry", resourceID)
		}
		if result.RowsAffected != 1 {
			return apperrors.WrapNotFound("shift_entry", resourceID)
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to delete shift entry")
	}
	return nil
}
