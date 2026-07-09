package repository

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// OwnerVisitSummary は飼い主のカルテ集計結果（Lステップタグ同期用）。
type OwnerVisitSummary struct {
	FirstVisitAt *time.Time
	LastVisitAt  *time.Time
	TotalCount   int64
	AnnualCount  int64
}

// DormantOwnerEntry はバッチ用の休眠飼い主エントリ。
type DormantOwnerEntry struct {
	OwnerID   uint64
	DaysSince int
}

// FindLatestByOwner は飼い主の最新カルテ（created_at DESC）を返す（BE-006 次回来院推奨日タグ同期用）。
// カルテが存在しない場合は nil, nil を返す。
func (r *medicalRecordRepository) FindLatestByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("owner=%d", ownerID))
	}
	return &record, nil
}

// FindOwnerVisitSummary は飼い主の診療集計（初回/最終日・合計数・年間数）を返す（Lステップタグ同期用）。
func (r *medicalRecordRepository) FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*OwnerVisitSummary, error) {
	type row struct {
		FirstVisitAt *time.Time
		LastVisitAt  *time.Time
		TotalCount   int64
		AnnualCount  int64
	}
	var result row
	oneYearAgo := time.Now().In(time.Local).AddDate(-1, 0, 0)
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Select(`MIN(date) AS first_visit_at,
			MAX(date) AS last_visit_at,
			COUNT(*) AS total_count,
			COUNT(CASE WHEN date >= ? THEN 1 END) AS annual_count`, oneYearAgo).
		Scan(&result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("owner=%d", ownerID))
	}
	return &OwnerVisitSummary{
		FirstVisitAt: result.FirstVisitAt,
		LastVisitAt:  result.LastVisitAt,
		TotalCount:   result.TotalCount,
		AnnualCount:  result.AnnualCount,
	}, nil
}

// FindOwnersByFirstVisitDate は初回来院日（MIN(date)）が targetDate と一致する飼い主IDリストを返す（FEAT-383）。
func (r *medicalRecordRepository) FindOwnersByFirstVisitDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	target := targetDate.In(time.Local).Format(time.DateOnly)
	type row struct{ OwnerID uint64 }
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("deleted_at IS NULL").
		Select("owner_id").
		Group("owner_id").
		Having("MIN(date::date) = ?::date", target).
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d first_visit_date=%s", clinicID, target))
	}
	ids := make([]uint64, len(rows))
	for i, r := range rows {
		ids[i] = r.OwnerID
	}
	return ids, nil
}

// FindOwnersByLastVisitDays は最終来院日が asOf から exactDays 日前の飼い主IDリストを返す（FEAT-383）。
func (r *medicalRecordRepository) FindOwnersByLastVisitDays(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error) {
	target := asOf.In(time.Local).AddDate(0, 0, -exactDays).Format(time.DateOnly)
	type row struct{ OwnerID uint64 }
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("deleted_at IS NULL").
		Select("owner_id").
		Group("owner_id").
		Having("MAX(date::date) = ?::date", target).
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d last_visit_days=%d", clinicID, exactDays))
	}
	ids := make([]uint64, len(rows))
	for i, r := range rows {
		ids[i] = r.OwnerID
	}
	return ids, nil
}

// FindOwnersByNextVisitRecommended は次回来院推奨日が targetDate の最新カルテを持つ飼い主IDリストを返す（FEAT-383）。
// NOTE: P4 規約逸脱 (GORM Scopes 未使用) だが clinic_id を WHERE 句に二重指定して横テナント漏洩を防ぐ。
// リファクタ時に clinic_id WHERE のいずれか一方を削除しないこと (M-5 / AUDIT-2026-05-06 参照)。
func (r *medicalRecordRepository) FindOwnersByNextVisitRecommended(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	target := targetDate.In(time.Local).Format(time.DateOnly)
	type row struct{ OwnerID uint64 }
	var rows []row
	// 飼い主ごとに最新カルテ（MAX(id)）を取得し、その next_visit_recommended_date が targetDate のものを抽出。
	err := r.db.WithContext(ctx).Raw(`
		SELECT owner_id
		FROM medical_records
		WHERE clinic_id = ? AND deleted_at IS NULL
		  AND id IN (
		      SELECT MAX(id)
		      FROM medical_records
		      WHERE clinic_id = ? AND deleted_at IS NULL
		      GROUP BY owner_id
		  )
		  AND next_visit_recommended_date::date = ?::date
	`, clinicID, clinicID, target).Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d next_visit_recommended=%s", clinicID, target))
	}
	ids := make([]uint64, len(rows))
	for i, r := range rows {
		ids[i] = r.OwnerID
	}
	return ids, nil
}

// FindDormantOwnerEntries は最終来院から minDaysSince 日以上経過した飼い主一覧を返す（バッチ処理用）。
func (r *medicalRecordRepository) FindDormantOwnerEntries(ctx context.Context, clinicID uint64, minDaysSince int) ([]DormantOwnerEntry, error) {
	cutoff := time.Now().In(time.Local).AddDate(0, 0, -minDaysSince)
	type row struct {
		OwnerID     uint64
		LastVisitAt time.Time
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("deleted_at IS NULL").
		Select("owner_id, MAX(date) AS last_visit_at").
		Group("owner_id").
		Having("MAX(date) < ?", cutoff).
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d dormant", clinicID))
	}
	now := time.Now().In(time.Local)
	entries := make([]DormantOwnerEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, DormantOwnerEntry{
			OwnerID:   r.OwnerID,
			DaysSince: int(now.Sub(r.LastVisitAt).Hours() / 24),
		})
	}
	return entries, nil
}
