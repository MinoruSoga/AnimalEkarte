package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LabImportJobRepository は lab_import_jobs の永続化インターフェース。
type LabImportJobRepository interface {
	// Create は新規インポートジョブを作成する。
	Create(ctx context.Context, job *model.LabImportJob) error
	// Update はジョブ状態を保存する（clinic_id スコープ）。
	Update(ctx context.Context, job *model.LabImportJob) error
	// FindByID はクリニックスコープで ID に一致するジョブを返す。
	FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error)
	// FindByClinic はクリニックスコープで最新順にジョブ一覧を返す。
	FindByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LabImportJob, error)
}

// LabImportEventRepository は lab_import_events の永続化インターフェース。
type LabImportEventRepository interface {
	// Create は監査イベントを追記する。
	Create(ctx context.Context, event *model.LabImportEvent) error
	// FindByJob はジョブ ID に紐づくイベントを時系列昇順で返す。
	FindByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error)
}

// ------------------------------------
// LabImportJobRepository 実装
// ------------------------------------

type labImportJobRepository struct{ db *gorm.DB }

// NewLabImportJobRepository は LabImportJobRepository を初期化して返す。
func NewLabImportJobRepository(db *gorm.DB) LabImportJobRepository {
	return &labImportJobRepository{db: db}
}

func (r *labImportJobRepository) Create(ctx context.Context, job *model.LabImportJob) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_job", "")
	}
	return nil
}

func (r *labImportJobRepository) Update(ctx context.Context, job *model.LabImportJob) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(job.ClinicID)).
		Save(job)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lab_import_job", job.ID.String())
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lab_import_job", job.ID.String())
	}
	return nil
}

func (r *labImportJobRepository) FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error) {
	var job model.LabImportJob
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		First(&job).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_job", id.String())
	}
	return &job, nil
}

func (r *labImportJobRepository) FindByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LabImportJob, error) {
	var jobs []*model.LabImportJob
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("created_at DESC").
		Limit(limit).
		Find(&jobs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_job", "")
	}
	return jobs, nil
}

// ------------------------------------
// LabImportEventRepository 実装
// ------------------------------------

type labImportEventRepository struct{ db *gorm.DB }

// NewLabImportEventRepository は LabImportEventRepository を初期化して返す。
func NewLabImportEventRepository(db *gorm.DB) LabImportEventRepository {
	return &labImportEventRepository{db: db}
}

func (r *labImportEventRepository) Create(ctx context.Context, event *model.LabImportEvent) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_event", "")
	}
	return nil
}

func (r *labImportEventRepository) FindByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
	var events []*model.LabImportEvent
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND job_id = ?", clinicID, jobID).
		Order("created_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_event", jobID.String())
	}
	return events, nil
}

// ------------------------------------
// LabImportDuplicateChecker DB 実装
// ------------------------------------

// LabImportDuplicateCheckerDB は exams テーブルを直接参照する DB-backed 重複チェッカー。
// (clinic_id, exam_type_id, date, pet_id) の複合キーで既存 exam を検索する。
// pet_id が nil の場合は "pet_id IS NULL" として検索する（ISO SQL: NULL ≠ NULL）。
//
// TOCTOU 注意: IsDuplicate と Create の間には競合ウィンドウがある。
// DB unique 制約がないため、並行リクエストによる重複行の作成を DB レベルでは防げない。
// PersistExam の AlreadyExists 安全ネットは DB unique 制約が存在しない限り発火しない。
// DB unique 制約は Phase 3 で追加予定（本番データの違反件数確認後）。
// それまでは concurrent import を直列化するか、この制約を運用上許容すること。
//
// date 値は UTC 日付部のみ（時刻成分なし）で渡すこと。
// IsDuplicate に渡す前に呼び出し元で time.Date(y, m, d, 0, 0, 0, 0, time.UTC) に正規化すること。
// GORM が time.Time を PostgreSQL date 型と比較する際、サーバータイムゾーンによっては
// 時刻成分が日付境界をまたいで誤判定することがある。
type LabImportDuplicateCheckerDB struct {
	db *gorm.DB
}

// NewLabImportDuplicateCheckerDB は DB-backed LabImportDuplicateChecker を返す。
func NewLabImportDuplicateCheckerDB(db *gorm.DB) *LabImportDuplicateCheckerDB {
	return &LabImportDuplicateCheckerDB{db: db}
}

// IsDuplicate は (clinic_id, exam_type_id, date, pet_id) に一致する exam が既存かを返す。
// date は UTC 日付部のみに正規化済みであること（呼び出し元責務）。
// GORM のソフトデリートスコープが deleted_at IS NULL を自動付与するが、
// 意図を明示するために明示的にも条件を加える。
func (c *LabImportDuplicateCheckerDB) IsDuplicate(ctx context.Context, clinicID, examTypeID uint64, date time.Time, petID *uint64) (bool, error) {
	// UTC date-only に正規化（呼び出し元が既に正規化している前提だが二重チェックとして）。
	normalised := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	q := c.db.WithContext(ctx).
		Model(&model.Examination{}).
		Where("clinic_id = ? AND exam_type_id = ? AND date = ?", clinicID, examTypeID, normalised)
	if petID == nil {
		q = q.Where("pet_id IS NULL")
	} else {
		q = q.Where("pet_id = ?", *petID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "exam", "")
	}
	return count > 0, nil
}
