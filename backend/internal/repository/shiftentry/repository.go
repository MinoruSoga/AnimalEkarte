// Package shiftentry owns shift_entries / shift_entry_breaks data access (BE8-4 batch13 — leaf domain).
package shiftentry

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Filter はシフト一覧取得のフィルタ条件
type Filter struct {
	YearMonth string // "YYYY-MM" 形式
	StaffID   *uint64
}

// Repository はシフト管理のデータアクセスインターフェース
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64, filter Filter) ([]model.ShiftEntry, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error)
	Create(ctx context.Context, entry *model.ShiftEntry) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error)
	ReplaceBreaks(ctx context.Context, shiftEntryID uint64, breaks []model.ShiftEntryBreak) error
	// FindOnDutyStaffs は指定日にシフトが登録されているスタッフ一覧を返す (BUG-344)
	FindOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error)
}

type repository struct{ db *gorm.DB }

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// isUniqueConstraintErr はPostgreSQLのユニーク制約違反（23505）を判定する
// （親 repository パッケージの同名ヘルパーの最小限の複製。repohelpers 未収載のため import cycle 回避のためローカル定義。BE8-4 batch13）。
func isUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64, filter Filter) ([]model.ShiftEntry, error) {
	q := repohelpers.DBOrTx(ctx, r.db).
		Preload("Staff", "deleted_at IS NULL").
		Preload("Breaks").
		Scopes(repohelpers.ClinicScope(clinicID)).
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

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error) {
	var entry model.ShiftEntry
	err := repohelpers.DBOrTx(ctx, r.db).
		Preload("Staff", "deleted_at IS NULL").
		Preload("Breaks").
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).
		First(&entry).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", strconv.FormatUint(id, 10))
	}
	return &entry, nil
}

func (r *repository) Create(ctx context.Context, entry *model.ShiftEntry) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Create(entry).Error; err != nil {
		// PostgreSQL UNIQUE違反 (23505)
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("shift_entry",
				fmt.Sprintf("staff_id=%d date=%s", entry.StaffID, entry.Date.Format(time.DateOnly)))
		}
		return apperrors.FromGORM(err, "shift_entry", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return repohelpers.UpdateScopedByID(ctx, repohelpers.DBOrTx(ctx, r.db), &model.ShiftEntry{}, "shift_entry", clinicID, id, fields)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := repohelpers.DBOrTx(ctx, r.db).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.ShiftEntry{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "shift_entry", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_entry", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *repository) ReplaceBreaks(ctx context.Context, shiftEntryID uint64, breaks []model.ShiftEntryBreak) error {
	return repohelpers.ReplaceChildRowsByParentID(
		repohelpers.DBOrTx(ctx, r.db),
		shiftEntryID,
		breaks,
		&model.ShiftEntryBreak{},
		"shift_entry_id",
		"shift_entry_break",
		func(row *model.ShiftEntryBreak, id uint64) { row.ShiftEntryID = id },
		"failed to replace shift entry breaks",
	)
}

func (r *repository) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	var count int64
	err := repohelpers.DBOrTx(ctx, r.db).Model(&model.ShiftEntry{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("staff_id = ?", staffID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "shift_entry", "")
	}
	return count > 0, nil
}

// FindOnDutyStaffs は指定日にシフトが登録されているスタッフ一覧を返す (BUG-344)
func (r *repository) FindOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error) {
	var staffs []model.Staff
	dateStr := date.Format(time.DateOnly)
	// shift_entries テーブルは deleted_at カラムを持たない（論理削除なし）
	// 勤務日の絞り込みは JOIN 条件の shift_entries.clinic_id で担保する。
	err := repohelpers.DBOrTx(ctx, r.db).
		Joins("JOIN shift_entries ON shift_entries.staff_id = staffs.id"+
			" AND shift_entries.clinic_id = ?"+
			" AND DATE(shift_entries.date) = ?", clinicID, dateStr).
		Where("staffs.deleted_at IS NULL AND staffs.is_active = true").
		Distinct("staffs.*").
		Find(&staffs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "on_duty_staffs", dateStr)
	}
	return staffs, nil
}
