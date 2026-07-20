package medicalrecord

// lab_import_repository.go — moved from internal/repository (BE9-2D sub-batch③, lab_import/lab_report
// saga roll-up). lab is a leaf domain, so the flat internal/repository side keeps only a temporary
// facade (repository/lab_import_repository.go) until the lab service/handler also move; that facade is
// deleted in Batch B/C. Behavior is unchanged: the package-private clinicScope helper is replaced by the
// exported repohelpers.ClinicScope (identical predicate), and r.db is referenced directly (no ambient-tx
// DBOrTx conversion) exactly as before.

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// LabImportJobRepository は lab_import_jobs の永続化インターフェース。
type LabImportJobRepository interface {
	// Create は新規インポートジョブを作成する。
	Create(ctx context.Context, job *model.LabImportJob) error
	// Update はジョブ状態を保存する（clinic_id スコープ）。
	Update(ctx context.Context, job *model.LabImportJob) error
	// FindByID はクリニックスコープで ID に一致するジョブを返す。
	FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error)
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
	// NOTE: GORM's Save() keys UPDATE purely on the primary key when it is already populated
	// (job.ID is a non-zero UUID here) and silently ignores the chained Scopes(repohelpers.ClinicScope(...))
	// condition — the clinic_id predicate would have zero effect and allow cross-tenant
	// overwrites. Use an explicit Model+Where+Updates(map) so clinic_id actually gates the
	// UPDATE's WHERE clause (P4: clinicScope on Update, MANDATORY).
	result := r.db.WithContext(ctx).
		Model(&model.LabImportJob{}).
		Where("id = ?", job.ID).
		Scopes(repohelpers.ClinicScope(job.ClinicID)).
		Updates(map[string]any{
			"source_type":        job.SourceType,
			"source_fingerprint": job.SourceFingerprint,
			"status":             job.Status,
			"row_count":          job.RowCount,
			"persisted_count":    job.PersistedCount,
			"duplicate_count":    job.DuplicateCount,
			"needs_review_count": job.NeedsReviewCount,
			"failed_count":       job.FailedCount,
			"error_code":         job.ErrorCode,
			"error_message":      job.ErrorMessage,
			"started_at":         job.StartedAt,
			"finished_at":        job.FinishedAt,
		})
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
// Phase 3A 決定: サービスレベル重複防止を正式方針として採用。DB unique 制約は追加しない。
//
// 根拠（ローカルデータ調査 2026-06-26）:
//
//	4-col key (clinic_id, exam_type_id, date, pet_id) には 87 重複グループ（95 超過行）が存在する。
//	そのうち 84/85 非 null グループは distinct な medical_record_id を持つ（同日別カルテの正当な複数受診）。
//	これらに unique 制約を追加すると既存移行データを拒絶する。
//	5-col key (clinic_id, exam_type_id, date, pet_id, medical_record_id) でゼロ違反を確認済み。
//	lab import の重複判定意味論（同ペット同日同検査の再インポート検知）は 4-col key が正しく、
//	DB unique 制約ではなくサービスレベルで実装する。
//
// TOCTOU 注意: IsDuplicate と Create の間には競合ウィンドウがある。
// DB unique 制約がないため、並行リクエストによる重複行の作成を DB レベルでは防げない。
// PersistExam の AlreadyExists 安全ネットは DB unique 制約が存在しない限り発火しない。
// concurrent import は呼び出し元で直列化するか、この重複を運用上許容すること。
//
// 本番データに DB unique 制約を追加する前に以下の SQL 3 本で確認すること:
//
//	-- (1) non-null pet_id: 4-col 違反グループ（0 でなければ制約追加不可）
//	SELECT COUNT(*) FROM (
//	  SELECT clinic_id, exam_type_id, date, pet_id
//	  FROM exams WHERE deleted_at IS NULL AND pet_id IS NOT NULL
//	  GROUP BY clinic_id, exam_type_id, date, pet_id HAVING COUNT(*) > 1
//	) sub;
//
//	-- (2) null pet_id: 違反グループ
//	SELECT COUNT(*) FROM (
//	  SELECT clinic_id, exam_type_id, date
//	  FROM exams WHERE deleted_at IS NULL AND pet_id IS NULL
//	  GROUP BY clinic_id, exam_type_id, date HAVING COUNT(*) > 1
//	) sub;
//
//	-- (3) 真の重複グループ: 同一グループ内に 2 以上の distinct medical_record_id を持つ行
//	--     (1)(2) がゼロでも、これがゼロでなければ同一カルテの真の重複行が残っている
//	SELECT COUNT(*) FROM (
//	  SELECT clinic_id, exam_type_id, date, pet_id
//	  FROM exams WHERE deleted_at IS NULL AND pet_id IS NOT NULL
//	  GROUP BY clinic_id, exam_type_id, date, pet_id
//	  HAVING COUNT(DISTINCT medical_record_id) > 1
//	) sub;
//
// (1)(2)(3) の全クエリがゼロを返してから制約追加の migration を検討すること。
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
// date は UTC 日付部のみ (時刻成分なし) であること。呼び出し元は
// time.Date(y, m, d, 0, 0, 0, 0, time.UTC) で正規化してから渡すこと（呼び出し元責務）。
// 内部でも再正規化を行うが、これは write path との一致を保証しない安全ネットであり、
// 呼び出し元の正規化省略を許容する意図ではない。
// GORM の soft-delete スコープが自動で "deleted_at IS NULL" を付与する。
func (c *LabImportDuplicateCheckerDB) IsDuplicate(ctx context.Context, clinicID, examTypeID uint64, date time.Time, petID *uint64) (bool, error) {
	// UTC date-only に再正規化。write path も同様に正規化していること。
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
