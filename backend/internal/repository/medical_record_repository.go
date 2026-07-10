package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordListFilters はカルテ一覧のフィルタ条件（B-1: server-side pagination/search 拡張）。
// Search は飼主名/カナ・ペット名/カナ・record_no・主訴を対象に ILIKE + NormalizeKana で部分一致検索する。
type MedicalRecordListFilters struct {
	PetID           *uint64
	OwnerID         *uint64
	StartDate       *string
	EndDate         *string
	Status          *model.MedicalRecordStatus
	DoctorID        *uint64
	AnimalSpeciesID *uint64
	Search          string
	// Sort/Order: B-1 follow-up（列ソート server 化）。Sort はハンドラ層で検証済みの許可キー
	// （medicalRecordSortColumns の key）のみが渡される想定。空文字は既定順
	// （date DESC, created_at DESC）を維持する。Order は "asc"/"desc"（既定 "desc"）。
	Sort  string
	Order string
}

// medicalRecordSortColumns は B-1 follow-up の列ソート許可リスト（FE sort key → 実 SQL カラム）。
// ハンドラ層 (resolveMedicalRecordSort) がこのキー集合で検証済みの値のみを Filters.Sort に渡すため、
// ここでの map lookup は防御的二重チェックとして機能する（SQL injection: 値は固定文字列のみ使用）。
var medicalRecordSortColumns = map[string]string{
	"date":       "medical_records.date",
	"owner_name": "owners.name",
	"pet_name":   "pets.name",
	"status":     "medical_records.status",
}

type MedicalRecordRepository interface {
	// FindAll は指定した複数医院 (#86 拠点横断) のカルテを検索する。clinicIDs はハンドラ層で所属検証済みであること。
	FindAll(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	// FindByIDForClinics は複数医院スコープでカルテを1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	// Update は clinicID+id+status=draft でカルテを更新する。expectedVersion が非 nil の場合、
	// WHERE に version 述語を追加し（BE-refactor.md X-10: 楽観ロックの原子化）、RowsAffected==0
	// を「他のユーザーが変更した」（version不一致）と「確定済みで編集不可」（not draft）に区別する。
	// expectedVersion が nil の場合は従来どおり version 述語なし（照合スキップ）。
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any, expectedVersion *int) (*model.MedicalRecord, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// DeleteDraftByAppointmentID は予約に紐づく draft カルテを論理削除する (#83 Q10)
	DeleteDraftByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) error
	CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error)
	// FindFirstVisitDateByPetID は指定ペットの初診日（最古の有効カルテ date）を返す（#158 飼主レポート）。
	// clinic スコープ + 論理削除除外。カルテが存在しない場合は nil, nil を返す。
	FindFirstVisitDateByPetID(ctx context.Context, clinicID, petID uint64) (*time.Time, error)
	// CountEstimatesByMedicalRecordID は BE-refactor.md R2-5 (D12) で clinic_id 述語を追加した
	// （呼び出し元 Delete が clinicID を既に保持・ownership 検証済みのため低リスクで追加可能）。
	CountEstimatesByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (int64, error)
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

func (r *medicalRecordRepository) FindAll(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
	records := make([]model.MedicalRecord, 0)
	var total int64

	// フェイルセーフ: 検証バグ等で空スライスが渡っても全件露出させない
	if len(clinicIDs) == 0 {
		return []model.MedicalRecord{}, 0, nil
	}

	// search / animal_species_id / 列ソート(pet_name・owner_name) は pets / owners / inquiries への JOIN が必要。
	// inquiries.medical_record_id と owners/pets の FK は 1 レコードにつき高々1件のため LEFT JOIN による重複行は発生しない。
	needsPetJoin := filters.AnimalSpeciesID != nil || filters.Search != "" || filters.Sort == "pet_name"
	needsOwnerJoin := filters.Search != "" || filters.Sort == "owner_name"
	needsInquiryJoin := filters.Search != ""

	buildBase := func() *gorm.DB {
		// clinicScopeIn は "clinic_id" を無修飾で参照するため、pets/owners
		// （いずれも clinic_id 列を持つ）を LEFT JOIN すると曖昧になる。
		// search/animal_species_id フィルタで JOIN が入るケースがあるため、
		// ここでは常に medical_records.clinic_id を明示指定する。
		q := r.db.WithContext(ctx).Model(&model.MedicalRecord{}).Where("medical_records.clinic_id IN ?", clinicIDs)
		if needsPetJoin {
			q = q.Joins("LEFT JOIN pets ON pets.id = medical_records.pet_id AND pets.deleted_at IS NULL")
		}
		if needsOwnerJoin {
			q = q.Joins("LEFT JOIN owners ON owners.id = medical_records.owner_id AND owners.deleted_at IS NULL")
		}
		if needsInquiryJoin {
			q = q.Joins("LEFT JOIN inquiries ON inquiries.medical_record_id = medical_records.id")
		}
		if filters.PetID != nil {
			q = q.Where("medical_records.pet_id = ?", *filters.PetID)
		}
		if filters.OwnerID != nil {
			q = q.Where("medical_records.owner_id = ?", *filters.OwnerID)
		}
		if filters.StartDate != nil {
			q = q.Where("medical_records.date >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			q = q.Where("medical_records.date <= ?", *filters.EndDate)
		}
		if filters.Status != nil {
			q = q.Where("medical_records.status = ?", *filters.Status)
		}
		if filters.DoctorID != nil {
			q = q.Where("medical_records.doctor_id = ?", *filters.DoctorID)
		}
		if filters.AnimalSpeciesID != nil {
			q = q.Where("pets.animal_species_id = ?", *filters.AnimalSpeciesID)
		}
		if filters.Search != "" {
			// NormalizeKana で検索語のカタカナをひらがなに正規化し、DB 列の translate() 正規化値と統一比較する（owner_repository.go と同型）。
			pattern := "%" + escapeLike(NormalizeKana(filters.Search)) + "%"
			q = q.Where(
				`(medical_records.record_no ILIKE ? ESCAPE '\'`+
					` OR owners.name ILIKE ? ESCAPE '\'`+
					` OR translate(owners.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
					` OR pets.name ILIKE ? ESCAPE '\'`+
					` OR translate(pets.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
					` OR inquiries.chief_complaint ILIKE ? ESCAPE '\')`,
				pattern,
				pattern,
				kanaSourceChars, kanaTargetChars, pattern,
				pattern,
				kanaSourceChars, kanaTargetChars, pattern,
				pattern,
			)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	if err := buildBase().
		Offset((page-1)*limit).Limit(limit).Order(medicalRecordOrderClause(filters.Sort, filters.Order)).
		Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("Pet.AnimalSpecies").Preload("Doctor", "deleted_at IS NULL").Preload("EnteredByStaff", "deleted_at IS NULL").Preload("Inquiry").Preload("Billing", "deleted_at IS NULL").
		Find(&records).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return records, total, nil
}

// medicalRecordOrderClause は B-1 follow-up の列ソート server 化用 ORDER BY 句を組み立てる。
// sort が medicalRecordSortColumns の許可キーでない場合は既存デフォルト順を維持する。
// created_at DESC を常に tie-breaker として付与し、同値時の並び順を安定させる
// （デフォルト順と同じ安定化ルールを踏襲）。
func medicalRecordOrderClause(sort, order string) string {
	col, ok := medicalRecordSortColumns[sort]
	if !ok {
		return "medical_records.date DESC, medical_records.created_at DESC"
	}
	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}
	return fmt.Sprintf("%s %s, medical_records.created_at DESC", col, dir)
}

func (r *medicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return r.findMedicalRecordByID(ctx, clinicScope(clinicID), id)
}

func (r *medicalRecordRepository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
	return r.findMedicalRecordByID(ctx, clinicScopeIn(clinicIDs), id)
}

// findMedicalRecordByID は scope（単一/複数クリニック）を受け取りカルテを1件取得する共通実装。
func (r *medicalRecordRepository) findMedicalRecordByID(ctx context.Context, scope func(*gorm.DB) *gorm.DB, id uint64) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	err := r.db.WithContext(ctx).
		Preload("Treatments", "deleted_at IS NULL").
		Preload("Vitals").
		Preload("Doctor", "deleted_at IS NULL").
		Preload("EnteredByStaff", "deleted_at IS NULL").
		Preload("Owner", "deleted_at IS NULL").
		Preload("Pet", "deleted_at IS NULL").
		Preload("Pet.AnimalSpecies").
		Scopes(scope).Where("id = ?", id).First(&record).Error
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

// Update は BE-refactor.md X-10 のとおり、expectedVersion が非 nil の場合に WHERE へ
// version 述語を追加して楽観ロックを UPDATE と原子化する。RowsAffected==0 になった場合、
// version 述語を追加していたときだけ現在の status を再読して理由を区別する
// （draft のまま = version不一致、draft でない = 従来どおり not-draft）。再読自体が
// 失敗した場合は情報を出し過ぎないよう従来の not-draft Conflict にフォールバックする。
func (r *medicalRecordRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any, expectedVersion *int) (*model.MedicalRecord, error) {
	q := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ? AND status = ?", id, model.MedicalRecordStatusDraft)
	if expectedVersion != nil {
		q = q.Where("version = ?", *expectedVersion)
	}
	result := q.Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "medical_record", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		if expectedVersion != nil {
			var current model.MedicalRecord
			err := r.db.WithContext(ctx).
				Model(&model.MedicalRecord{}).
				Scopes(clinicScope(clinicID)).
				Where("id = ?", id).
				First(&current).Error
			if err == nil && current.Status == model.MedicalRecordStatusDraft {
				return nil, apperrors.WrapConflict("他のユーザーがこのカルテを変更しました。再読み込みしてください")
			}
		}
		// status != draft（finalized or already being updated）
		return nil, apperrors.WrapConflict("medical_record is not in draft status")
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

// DeleteDraftByAppointmentID は予約(appointment_id)に紐づく draft カルテを論理削除する (#83 Q10)。
// draft 以外(診察開始済み等)は削除しない。削除対象なし(RowsAffected 0)は正常としエラーにしない。
func (r *medicalRecordRepository) DeleteDraftByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) error {
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("appointment_id = ? AND status = ?", appointmentID, model.MedicalRecordStatusDraft).
		Delete(&model.MedicalRecord{}).Error
	if err != nil {
		return apperrors.FromGORM(err, "medical_record", fmt.Sprintf("appointment:%d", appointmentID))
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

// FindFirstVisitDateByPetID は指定ペットの初診日（最古の有効カルテ date）を返す（#158 飼主レポート）。
// clinic スコープ + 論理削除除外で集計し、捏造せず実データの MIN(date) のみ返す。
// カルテが存在しない場合は MIN が NULL となり nil, nil を返す。
func (r *medicalRecordRepository) FindFirstVisitDateByPetID(ctx context.Context, clinicID, petID uint64) (*time.Time, error) {
	var result struct {
		MinDate *time.Time `gorm:"column:min_date"`
	}
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("pet_id = ? AND deleted_at IS NULL", petID).
		Select("MIN(date) AS min_date").
		Scan(&result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("pet:%d", petID))
	}
	return result.MinDate, nil
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
// BE-refactor.md R2-5 (D12): clinic_id 述語を追加（estimates は clinic_id カラムを直接持つ）。
func (r *medicalRecordRepository) CountEstimatesByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Estimate{}).
		Where("medical_record_id = ? AND clinic_id = ? AND deleted_at IS NULL", medicalRecordID, clinicID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "estimate", "")
	}
	return count, nil
}
