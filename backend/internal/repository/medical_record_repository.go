package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

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

type MedicalRecordRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error)
	CountEstimatesByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (int64, error)
	// FindOwnerVisitSummary は飼い主の初回/最終診療日・年間来院数を集計して返す（Lステップ同期用）。
	FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*OwnerVisitSummary, error)
	// FindLatestByOwner は飼い主の最新カルテを返す（Lステップ次回来院推奨日タグ同期用）。
	// カルテが存在しない場合は nil, nil を返す。
	FindLatestByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error)
	// FindDormantOwnerEntries は最終来院から minDaysSince 日以上経過した飼い主一覧を返す（バッチ処理用）。
	FindDormantOwnerEntries(ctx context.Context, clinicID uint64, minDaysSince int) ([]DormantOwnerEntry, error)
	// FindOwnersByFirstVisitDate は初回来院日（MIN(date)）が targetDate と一致する飼い主IDリストを返す（FEAT-383）。
	FindOwnersByFirstVisitDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	// FindOwnersByLastVisitDays は最終来院日が asOf から exactDays 日前の飼い主IDリストを返す（FEAT-383）。
	FindOwnersByLastVisitDays(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error)
	// FindOwnersByNextVisitRecommended は次回来院推奨日が targetDate の飼い主IDリストを返す（FEAT-383）。
	FindOwnersByNextVisitRecommended(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	// CountByOwnerID は飼い主に紐付く有効カルテ数を返す（初診判定用: FEAT-383 Phase 2）。
	CountByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

type medicalRecordRepository struct {
	db *gorm.DB
}

func NewMedicalRecordRepository(db *gorm.DB) MedicalRecordRepository {
	return &medicalRecordRepository{db: db}
}

func (r *medicalRecordRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	records := make([]model.MedicalRecord, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.MedicalRecord{}).Scopes(clinicScope(clinicID))
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if startDate != nil {
		q = q.Where("date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("date <= ?", *endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	if err := q.Offset((page-1)*limit).Limit(limit).Order("date DESC, created_at DESC").
		Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("Pet.AnimalSpecies").Preload("Doctor", "deleted_at IS NULL").Preload("EnteredByStaff", "deleted_at IS NULL").Preload("Inquiry").Preload("Billing", "deleted_at IS NULL").
		Find(&records).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return records, total, nil
}

func (r *medicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	err := r.db.WithContext(ctx).
		Preload("Treatments", "deleted_at IS NULL").
		Preload("Vitals").
		Preload("Doctor", "deleted_at IS NULL").
		Preload("EnteredByStaff", "deleted_at IS NULL").
		Preload("Owner", "deleted_at IS NULL").
		Preload("Pet", "deleted_at IS NULL").
		Preload("Pet.AnimalSpecies").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", id))
	}
	return &record, nil
}

func (r *medicalRecordRepository) Create(ctx context.Context, record *model.MedicalRecord) error {
	err := r.db.WithContext(ctx).Create(record).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("medical_record", record.RecordNo)
		}
		return apperrors.FromGORM(err, "medical_record", "")
	}
	return nil
}

func (r *medicalRecordRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error) {
	result := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "medical_record", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *medicalRecordRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.MedicalRecord{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medical_record", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountByPetID は指定されたペットに関連するカルテ数を返す
func (r *medicalRecordRepository) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("pet_id = ? AND deleted_at IS NULL", petID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return count, nil
}

// CountByOwnerID は飼い主に紐付く有効カルテ数を返す（初診判定用: FEAT-383 Phase 2）。
func (r *medicalRecordRepository) CountByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return count, nil
}

// CountEstimatesByMedicalRecordID はカルテに紐付く見積書の件数を返す（BUG-201）
// estimates.medical_record_id は ON DELETE RESTRICT のため削除前チェックが必要。
func (r *medicalRecordRepository) CountEstimatesByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Estimate{}).
		Where("medical_record_id = ? AND deleted_at IS NULL", medicalRecordID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "estimate", "")
	}
	return count, nil
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
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
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
	target := targetDate.Format("2006-01-02")
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
	target := asOf.AddDate(0, 0, -exactDays).Format("2006-01-02")
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
	target := targetDate.Format("2006-01-02")
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
	cutoff := time.Now().AddDate(0, 0, -minDaysSince)
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
	now := time.Now()
	entries := make([]DormantOwnerEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, DormantOwnerEntry{
			OwnerID:   r.OwnerID,
			DaysSince: int(now.Sub(r.LastVisitAt).Hours() / 24),
		})
	}
	return entries, nil
}
