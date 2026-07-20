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

	"github.com/animal-ekarte/backend/internal/apperrors"
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
	// --- 予約スケジュール用途の書き込み（ADR-006 論点#1 案A: shift_entries の書き込みは
	// staff domain（本package）の exported メソッドへ一本化し、reservation 側は delegate 経由で呼ぶ）。
	// 既存 Create/Update/Delete/ReplaceBreaks と意図的に別メソッド: (staff_id, date) キーの
	// upsert + breaks 全置換を単一 tx で行う予約画面固有の contract のため。
	SaveByStaffDate(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error
	DeleteByStaffDate(ctx context.Context, clinicID, staffID uint64, date time.Time) error
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

// ---- 予約スケジュール用途の書き込み（ADR-006 論点#1 案A で reservation_schedule_repository.go から移動） ----

// SaveByStaffDate は ShiftEntry と ShiftEntryBreaks をトランザクションで upsert する。
// スコープは entry.ClinicID（呼び出し元 service が認証済み clinicID を設定する）。
func (r *repository) SaveByStaffDate(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error {
	// fail-closed: スコープ述語は entry.ClinicID を使うため、認証済み clinicID との
	// 不一致は書込前に拒否する（将来の呼び出し元が非認証由来の値を渡す事故を封じる）。
	if entry.ClinicID != clinicID {
		return apperrors.WrapInvalidInput("shift entry clinic_id mismatch")
	}
	if err := repohelpers.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// 既存エントリを検索
		var existing model.ShiftEntry
		err := tx.Scopes(repohelpers.ClinicScope(entry.ClinicID)).
			Where("staff_id = ? AND date = ?",
				entry.StaffID, entry.Date.Format(time.DateOnly)).
			First(&existing).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.FromGORM(err, "shift_entry", fmt.Sprintf("staff=%d", entry.StaffID))
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 新規作成
			if err2 := tx.Create(entry).Error; err2 != nil {
				return apperrors.FromGORM(err2, "shift_entry", "")
			}
		} else {
			// 更新
			entry.ID = existing.ID
			fields := map[string]any{
				"shift_type": entry.ShiftType,
				"start_time": entry.StartTime,
				"end_time":   entry.EndTime,
				"notes":      entry.Notes,
				"updated_at": gorm.Expr("NOW()"),
			}
			if err2 := tx.Scopes(repohelpers.ClinicScope(entry.ClinicID)).
				Model(&model.ShiftEntry{}).Where("id = ?", existing.ID).Updates(fields).Error; err2 != nil {
				return apperrors.FromGORM(err2, "shift_entry", fmt.Sprintf("%d", existing.ID))
			}
		}

		// 既存のbreaksを削除してから再作成
		if err2 := tx.Where("shift_entry_id = ?", entry.ID).Delete(&model.ShiftEntryBreak{}).Error; err2 != nil {
			return apperrors.FromGORM(err2, "shift_entry_break", fmt.Sprintf("%d", entry.ID))
		}
		if len(breaks) > 0 {
			for i := range breaks {
				breaks[i].ShiftEntryID = entry.ID
			}
			if err2 := tx.Create(&breaks).Error; err2 != nil {
				return apperrors.FromGORM(err2, "shift_entry_break", "")
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to upsert shift entry")
	}
	return nil
}

// DeleteByStaffDate は (staff_id, date) キーで ShiftEntry を削除する（clinic_id スコープ付き）。
func (r *repository) DeleteByStaffDate(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	result := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("staff_id = ? AND date = ?",
			staffID, date.Format(time.DateOnly)).
		Delete(&model.ShiftEntry{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "schedule_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format(time.DateOnly)))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format(time.DateOnly)))
	}
	return nil
}
