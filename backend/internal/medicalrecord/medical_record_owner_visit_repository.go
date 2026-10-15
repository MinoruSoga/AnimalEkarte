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

// alivePetCandidateJoin は通知候補抽出（初診/最終来院/休眠ファインダ系）に使う
// 生存ペット JOIN 条件（EMR-330）。status と deceased_at の片側だけ立つ
// 不整合センチネルも両方除外し、論理削除ペットも除く。
// 履歴集計（FindOwnerVisitSummary）・CountByOwnerID・LTV/売上集計には適用しない。
const alivePetCandidateJoin = "current_owner_pet.status <> ? AND current_owner_pet.deceased_at IS NULL AND current_owner_pet.deleted_at IS NULL"

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
		Joins("JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id AND "+alivePetCandidateJoin, model.PetStatusDeceased).
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
		Joins("JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id AND "+alivePetCandidateJoin, model.PetStatusDeceased).
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
		Joins("INNER JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id AND "+alivePetCandidateJoin, model.PetStatusDeceased).
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
			DaysSince: dormantCalendarDaysSince(now, r.LastVisitAt),
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
		Joins("INNER JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id AND "+alivePetCandidateJoin, model.PetStatusDeceased).
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
			DaysSince: dormantCalendarDaysSince(evaluatedAt, r.LastVisitAt),
		})
	}
	return entries, nil
}

// FindDormantOwnerHistoryEntriesCursor uses all retained visit history for
// CPM/dormant tag classification, matching FindOwnerVisitSummary. Notification
// candidates continue to use the alive-filtered FindDormantOwnerEntries methods.
func (r *medicalRecordRepository) FindDormantOwnerHistoryEntriesCursor(ctx context.Context, clinicID uint64, minDaysSince int, afterOwnerID uint64, limit int) ([]DormantOwnerEntry, error) {
	return r.FindDormantOwnerHistoryEntriesCursorAt(ctx, clinicID, minDaysSince, afterOwnerID, limit, time.Now())
}

func (r *medicalRecordRepository) FindDormantOwnerHistoryEntriesCursorAt(ctx context.Context, clinicID uint64, minDaysSince int, afterOwnerID uint64, limit int, evaluatedAt time.Time) ([]DormantOwnerEntry, error) {
	evaluatedAt = evaluatedAt.In(time.Local)
	type row struct {
		OwnerID     uint64
		LastVisitAt time.Time
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("medical_records AS mr").
		Joins("JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id").
		Joins("JOIN owners AS o ON o.id = current_owner_pet.owner_id AND o.clinic_id = mr.clinic_id AND o.deleted_at IS NULL").
		Where("mr.clinic_id = ? AND mr.deleted_at IS NULL AND current_owner_pet.owner_id > ?", clinicID, afterOwnerID).
		Select("current_owner_pet.owner_id, MAX(mr.date) AS last_visit_at").Group("current_owner_pet.owner_id").
		Having("MAX(mr.date) < ?", evaluatedAt.AddDate(0, 0, -minDaysSince)).
		Order("current_owner_pet.owner_id ASC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d dormant history cursor", clinicID))
	}
	entries := make([]DormantOwnerEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, DormantOwnerEntry{OwnerID: row.OwnerID, DaysSince: dormantCalendarDaysSince(evaluatedAt, row.LastVisitAt)})
	}
	return entries, nil
}

// ownerHasAliveVisitBoundary は生存ペット由来カルテだけを対象に、飼い主の
// 初診日（first=true → MIN(date)）または最終来院日（first=false → MAX(date)）が
// targetDate の暦日と一致するかを owner 単位で返す。
// 通知トリガーの scheduled claim 前・delivery owner lock 内 write 直前に
// イベントソースを再読する EMR-330 適格性チェック用で、全クリニックスキャンを
// 再実行しない narrow な owner スコープの読み取り。
func (r *medicalRecordRepository) ownerHasAliveVisitBoundary(ctx context.Context, clinicID, ownerID uint64, targetDate time.Time, first bool) (bool, error) {
	target := targetDate.In(time.Local).Format(time.DateOnly)
	having := "MAX(mr.date::date) = ?::date"
	if first {
		having = "MIN(mr.date::date) = ?::date"
	}
	type row struct{ OwnerID uint64 }
	var rows []row
	// ambient tx が存在すれば同一接続で再読する（delivery-owner lock 内の
	// 最終適格性チェックは単一プールでも第 2 接続を要求してはならない）。
	err := persistence.DBOrTx(ctx, r.db).
		Table("medical_records AS mr").
		Joins("JOIN pets AS current_owner_pet ON current_owner_pet.id = mr.pet_id AND current_owner_pet.clinic_id = mr.clinic_id AND "+alivePetCandidateJoin, model.PetStatusDeceased).
		Joins("JOIN owners AS o ON o.id = current_owner_pet.owner_id AND o.clinic_id = mr.clinic_id AND o.deleted_at IS NULL").
		Where("mr.clinic_id = ? AND mr.deleted_at IS NULL AND current_owner_pet.owner_id = ?", clinicID, ownerID).
		Select("current_owner_pet.owner_id").
		Group("current_owner_pet.owner_id").
		Having(having, target).
		Scan(&rows).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d owner=%d alive_visit_boundary", clinicID, ownerID))
	}
	return len(rows) > 0, nil
}

// OwnerHasAliveFirstVisitOnDate は生存ペット由来カルテの初診日が targetDate と
// 一致するかを返す（EMR-330 初診フォローアップ再検証用）。
func (r *medicalRecordRepository) OwnerHasAliveFirstVisitOnDate(ctx context.Context, clinicID, ownerID uint64, targetDate time.Time) (bool, error) {
	return r.ownerHasAliveVisitBoundary(ctx, clinicID, ownerID, targetDate, true)
}

// OwnerHasAliveLastVisitOnDate は生存ペット由来カルテの最終来院日が targetDate と
// 一致するかを返す（EMR-330 休眠予防再検証用）。
func (r *medicalRecordRepository) OwnerHasAliveLastVisitOnDate(ctx context.Context, clinicID, ownerID uint64, targetDate time.Time) (bool, error) {
	return r.ownerHasAliveVisitBoundary(ctx, clinicID, ownerID, targetDate, false)
}

func dormantCalendarDaysSince(evaluatedAt, visitDate time.Time) int {
	year, month, day := evaluatedAt.Date()
	visitYear, visitMonth, visitDay := visitDate.Date()
	evaluatedDay := time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
	lastVisitDay := time.Date(visitYear, visitMonth, visitDay, 0, 0, 0, 0, time.UTC).Unix()
	return int((evaluatedDay - lastVisitDay) / (24 * 60 * 60))
}
