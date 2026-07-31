package medicalrecord

// lab_import_repository.go — moved from internal/repository (BE9-2D sub-batch③, lab_import/lab_report
// saga roll-up). lab is a leaf domain, so the flat internal/repository side keeps only a temporary
// facade (repository/lab_import_repository.go) until the lab service/handler also move; that facade is
// deleted in Batch B/C. Behavior is unchanged: the package-private clinicScope helper is replaced by the
// exported persistence.ClinicScope (identical predicate), and r.db is referenced directly (no ambient-tx
// DBOrTx conversion) exactly as before.

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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
	// (job.ID is a non-zero UUID here) and silently ignores the chained Scopes(persistence.ClinicScope(...))
	// condition — the clinic_id predicate would have zero effect and allow cross-tenant
	// overwrites. Use an explicit Model+Where+Updates(map) so clinic_id actually gates the
	// UPDATE's WHERE clause (P4: clinicScope on Update, MANDATORY).
	result := r.db.WithContext(ctx).
		Model(&model.LabImportJob{}).
		Where("id = ?", job.ID).
		Scopes(persistence.ClinicScope(job.ClinicID)).
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

// LabImportDuplicateCheckerDB は exams / exam_results を参照する DB-backed 重複チェッカー。
//
// Issue #249 R-3 / PO ruling:
//
//	日付粒度 (clinic_id, exam_type_id, date, pet_id) だけでは重複にしない。
//	候補を 4-col で絞った後、medical_record_id・machine・exam_results ペイロードが
//	完全一致する既存 exam がある場合のみ true（完全同一再インポートのスキップ）。
//	同日・同検査種別でも内容が異なれば false → 新規保存する。
//
// pet_id / medical_record_id が nil の場合は "IS NULL" として比較する（ISO SQL: NULL ≠ NULL）。
//
// Phase 3A 決定（更新 R-3）: サービスレベル完全同一検知を正式方針として採用。
// DB unique 制約は追加しない（同日再検査の正当な複数行を許容するため）。
//
// TOCTOU 注意: IsDuplicate と Create の間には競合ウィンドウがある。
// DB unique 制約がないため、並行リクエストによる重複行の作成を DB レベルでは防げない。
// PersistExam の AlreadyExists 安全ネットは DB unique 制約が存在しない限り発火しない。
// concurrent import は呼び出し元で直列化するか、この重複を運用上許容すること。
//
// date 値は UTC 日付部のみ（時刻成分なし）で渡すこと。
// IsDuplicate に渡す前に呼び出し元で time.Date(y, m, d, 0, 0, 0, 0, time.UTC) に正規化すること。
// GORM が time.Time を PostgreSQL date 型と比較する際、サーバータイムゾーンによっては
// 時刻成分が日付境界をまたいで誤判定することがある。
//
// exam_results は clinic_id を持たないため、親 exams を clinic_id で絞ってから
// association 経由で読む（tenant isolation）。
type LabImportDuplicateCheckerDB struct {
	db *gorm.DB
}

// NewLabImportDuplicateCheckerDB は DB-backed LabImportDuplicateChecker を返す。
func NewLabImportDuplicateCheckerDB(db *gorm.DB) *LabImportDuplicateCheckerDB {
	return &LabImportDuplicateCheckerDB{db: db}
}

// IsDuplicate は入力ペイロードと完全一致する既存 exam がある場合に true を返す。
//
// 候補フィルタ: clinic_id, exam_type_id, date (UTC date-only), pet_id (NULL-aware)
// 完全一致: medical_record_id (NULL-aware), machine, exam_results の
//
//	(name, inspection_value, unit, reference_value, ref_min, ref_max, sort_order)
//
// 除外: JobID, Status, IsAbnormal/Status 派生値, id, timestamps, soft-deleted 行
//
// date は UTC 日付部のみであること。内部でも再正規化するが、呼び出し元の正規化が責務。
// GORM の soft-delete スコープが自動で "deleted_at IS NULL" を付与する。
func (c *LabImportDuplicateCheckerDB) IsDuplicate(ctx context.Context, input LabExamPersistInput) (bool, error) {
	// UTC date-only に再正規化。write path も同様に正規化していること。
	normalised := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, time.UTC)

	// 4-col 候補フィルタ（idx_exams_clinic_exam_type_date を利用）。内容比較は後段。
	q := c.db.WithContext(ctx).
		Model(&model.Examination{}).
		Preload("Items").
		Where("clinic_id = ? AND exam_type_id = ? AND date = ?", input.ClinicID, input.ExamTypeID, normalised)
	if input.PetID == nil {
		q = q.Where("pet_id IS NULL")
	} else {
		q = q.Where("pet_id = ?", *input.PetID)
	}

	var candidates []model.Examination
	if err := q.Find(&candidates).Error; err != nil {
		return false, apperrors.FromGORM(err, "exam", "")
	}
	for i := range candidates {
		if labImportExamFullyMatches(&candidates[i], input) {
			return true, nil
		}
	}
	return false, nil
}

// labImportExamFullyMatches は候補 exam が import 入力と完全同一かを判定する。
// 候補は既に clinic_id / exam_type_id / date / pet_id で絞られている前提。
func labImportExamFullyMatches(exam *model.Examination, input LabExamPersistInput) bool {
	if !labImportNullableUint64Equal(exam.MedicalRecordID, input.MedicalRecordID) {
		return false
	}
	if exam.Machine != input.Machine {
		return false
	}
	return labImportExamResultsMatch(exam.Items, input.Items)
}

// labImportExamResultsMatch は stored exam_results と入力 items の payload 一致を判定する。
// 比較対象: name, inspection_value, unit, reference_value, ref_min, ref_max, sort_order
// （IsAbnormal / Status は write 時の派生値のため identity から除外）。
func labImportExamResultsMatch(stored []model.ExamResult, input []LabExamItemInput) bool {
	if len(stored) != len(input) {
		return false
	}
	type payload struct {
		name, inspection, unit, reference string
		refMin, refMax                    *float64
		sortOrder                         int
	}
	left := make([]payload, len(stored))
	right := make([]payload, len(input))
	for i := range stored {
		left[i] = payload{
			name: stored[i].Name, inspection: stored[i].InspectionValue,
			unit: stored[i].Unit, reference: stored[i].ReferenceValue,
			refMin: stored[i].RefMin, refMax: stored[i].RefMax, sortOrder: stored[i].SortOrder,
		}
	}
	for i := range input {
		right[i] = payload{
			name: input[i].Name, inspection: input[i].InspectionValue,
			unit: input[i].Unit, reference: input[i].ReferenceValue,
			refMin: input[i].RefMin, refMax: input[i].RefMax, sortOrder: input[i].SortOrder,
		}
	}
	less := func(a, b payload) bool {
		if a.sortOrder != b.sortOrder {
			return a.sortOrder < b.sortOrder
		}
		return a.name < b.name
	}
	sort.Slice(left, func(i, j int) bool { return less(left[i], left[j]) })
	sort.Slice(right, func(i, j int) bool { return less(right[i], right[j]) })
	for i := range left {
		if left[i].name != right[i].name ||
			left[i].inspection != right[i].inspection ||
			left[i].unit != right[i].unit ||
			left[i].reference != right[i].reference ||
			left[i].sortOrder != right[i].sortOrder ||
			!labImportNullableFloat64Equal(left[i].refMin, right[i].refMin) ||
			!labImportNullableFloat64Equal(left[i].refMax, right[i].refMax) {
			return false
		}
	}
	return true
}

func labImportNullableUint64Equal(a, b *uint64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func labImportNullableFloat64Equal(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
