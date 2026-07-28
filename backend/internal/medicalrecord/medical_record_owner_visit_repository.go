package medicalrecord

// Moved from internal/repository (BE9-2D ⑦ Batch A). 旧 package-private helper は repohelpers
// 同等物へ置換（同一述語/ambient-tx参加）。外部は internal/repository の facade alias 経由で不変。

import (
	"context"
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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
		Scopes(persistence.ClinicScope(clinicID)).
		Where(`
			EXISTS (
				SELECT 1
				FROM pets current_owner_pet
				JOIN owners current_owner
				  ON current_owner.id = current_owner_pet.owner_id
				 AND current_owner.clinic_id = current_owner_pet.clinic_id
				WHERE current_owner_pet.id = medical_records.pet_id
				  AND current_owner_pet.clinic_id = medical_records.clinic_id
				  AND current_owner.id = ?
			)
		`, ownerID).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		// #236 BUG#4: IsNotFound は生 gorm.ErrRecordNotFound では常に false。FromGORM でラップしてから判定する。
		wrapped := apperrors.FromGORM(err, "medical_record", fmt.Sprintf("owner=%d", ownerID))
		if apperrors.IsNotFound(wrapped) {
			return nil, nil
		}
		return nil, wrapped
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
		Scopes(persistence.ClinicScope(clinicID)).
		Where(`
			deleted_at IS NULL
			AND EXISTS (
				SELECT 1
				FROM pets current_owner_pet
				JOIN owners current_owner
				  ON current_owner.id = current_owner_pet.owner_id
				 AND current_owner.clinic_id = current_owner_pet.clinic_id
				WHERE current_owner_pet.id = medical_records.pet_id
				  AND current_owner_pet.clinic_id = medical_records.clinic_id
				  AND current_owner.id = ?
			)
		`, ownerID).
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
		Table("medical_records AS mr").
		Joins("JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id").
		Joins("JOIN owners AS o ON o.id = current_owner_pet.owner_id AND o.clinic_id = mr.clinic_id AND o.deleted_at IS NULL").
		Where("mr.clinic_id = ? AND mr.deleted_at IS NULL", clinicID).
		Select("current_owner_pet.owner_id").
		Group("current_owner_pet.owner_id").
		Having("MIN(mr.date::date) = ?::date", target).
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
		Table("medical_records AS mr").
		Joins("JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id").
		Joins("JOIN owners AS o ON o.id = current_owner_pet.owner_id AND o.clinic_id = mr.clinic_id AND o.deleted_at IS NULL").
		Where("mr.clinic_id = ? AND mr.deleted_at IS NULL", clinicID).
		Select("current_owner_pet.owner_id").
		Group("current_owner_pet.owner_id").
		Having("MAX(mr.date::date) = ?::date", target).
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
		SELECT current_owner_pet.owner_id
		FROM medical_records AS mr
		JOIN pets AS current_owner_pet
		  ON current_owner_pet.id = mr.pet_id
		 AND current_owner_pet.clinic_id = mr.clinic_id
		JOIN owners AS o
		  ON o.id = current_owner_pet.owner_id
		 AND o.clinic_id = mr.clinic_id
		 AND o.deleted_at IS NULL
		WHERE mr.clinic_id = ? AND mr.deleted_at IS NULL
		  AND mr.id IN (
		      SELECT MAX(latest_mr.id)
		      FROM medical_records AS latest_mr
		      JOIN pets AS latest_owner_pet
		        ON latest_owner_pet.id = latest_mr.pet_id
		       AND latest_owner_pet.clinic_id = latest_mr.clinic_id
		      WHERE latest_mr.clinic_id = ? AND latest_mr.deleted_at IS NULL
		      GROUP BY latest_owner_pet.owner_id
		  )
		  AND mr.next_visit_recommended_date::date = ?::date
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
		Table("medical_records AS mr").
		Joins("INNER JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id").
		Joins("INNER JOIN owners AS o ON o.id = current_owner_pet.owner_id AND o.clinic_id = mr.clinic_id AND o.deleted_at IS NULL").
		Where("mr.clinic_id = ? AND mr.deleted_at IS NULL", clinicID).
		Select("current_owner_pet.owner_id, MAX(mr.date) AS last_visit_at").
		Group("current_owner_pet.owner_id").
		Having("MAX(mr.date) < ?", cutoff).
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

// FindDormantOwnerEntriesCursor は最終来院から minDaysSince 日以上経過した飼い主一覧を
// owner_id カーソルページネーションで返す（PERF-FOLLOWUP-02）。owner_id 昇順で afterOwnerID より
// 大きいものを最大 limit 件返す。
func (r *medicalRecordRepository) FindDormantOwnerEntriesCursor(ctx context.Context, clinicID uint64, minDaysSince int, afterOwnerID uint64, limit int) ([]DormantOwnerEntry, error) {
	return r.FindDormantOwnerEntriesCursorAt(
		ctx,
		clinicID,
		minDaysSince,
		afterOwnerID,
		limit,
		time.Now(),
	)
}

// FindDormantOwnerEntriesCursorAt evaluates both the cutoff and DaysSince
// against the durable scheduler timestamp.
func (r *medicalRecordRepository) FindDormantOwnerEntriesCursorAt(
	ctx context.Context,
	clinicID uint64,
	minDaysSince int,
	afterOwnerID uint64,
	limit int,
	evaluatedAt time.Time,
) ([]DormantOwnerEntry, error) {
	evaluatedAt = evaluatedAt.In(time.Local)
	cutoff := evaluatedAt.AddDate(0, 0, -minDaysSince)
	type row struct {
		OwnerID     uint64
		LastVisitAt time.Time
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("medical_records AS mr").
		Joins("INNER JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id").
		Joins("INNER JOIN owners AS o ON o.id = current_owner_pet.owner_id AND o.clinic_id = mr.clinic_id AND o.deleted_at IS NULL").
		Where("mr.clinic_id = ? AND mr.deleted_at IS NULL AND current_owner_pet.owner_id > ?", clinicID, afterOwnerID).
		Select("current_owner_pet.owner_id, MAX(mr.date) AS last_visit_at").
		Group("current_owner_pet.owner_id").
		Having("MAX(mr.date) < ?", cutoff).
		Order("current_owner_pet.owner_id ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d dormant cursor", clinicID))
	}
	entries := make([]DormantOwnerEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, DormantOwnerEntry{
			OwnerID:   r.OwnerID,
			DaysSince: int(evaluatedAt.Sub(r.LastVisitAt).Hours() / 24),
		})
	}
	return entries, nil
}
