package medicalrecord

// lab_import_repository.go — lab import job/event persistence + TASK-032 compensation receipts.
// All methods use persistence.DBOrTx. Lock / CAS methods reject ambient-tx absence (no base DB fallback).

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LabImportJobRepository は lab_import_jobs の永続化インターフェース。
type LabImportJobRepository interface {
	Create(ctx context.Context, job *model.LabImportJob) error
	Update(ctx context.Context, job *model.LabImportJob) error
	FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error)
	// LockByIDForUpdate locks the job by (clinic_id, id) without filtering on status.
	// Requires an ambient transaction.
	LockByIDForUpdate(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error)
	// CompareAndSetStatus updates status only when current status equals from.
	// Returns RowsAffected. Requires an ambient transaction.
	CompareAndSetStatus(ctx context.Context, clinicID uint64, id uuid.UUID, from, to model.LabImportJobStatus, finishedAt *time.Time) (int64, error)
}

// LabImportEventRepository は lab_import_events の永続化インターフェース。
type LabImportEventRepository interface {
	Create(ctx context.Context, event *model.LabImportEvent) error
	FindByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error)
	// HasEventType reports whether the job has at least one event of the given type (clinic-scoped).
	HasEventType(ctx context.Context, clinicID uint64, jobID uuid.UUID, eventType model.LabImportEventType) (bool, error)
}

// LabImportUsageReceiptRepository は append-only usage receipts。
type LabImportUsageReceiptRepository interface {
	Create(ctx context.Context, receipt *model.LabImportUsageReceipt) error
	// LockByJobForUpdate locks all usage receipts for a job (clinic-scoped). Requires ambient tx.
	LockByJobForUpdate(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportUsageReceipt, error)
	// CountByJob returns total receipt count for the job (clinic-scoped).
	CountByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) (int64, error)
	// CountManualMutationByJob returns manual_mutation receipt count for the job.
	CountManualMutationByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) (int64, error)
}

// LabImportRevertReceiptRepository は idempotent revert receipts。
type LabImportRevertReceiptRepository interface {
	// FindByIdempotencyKey looks up a receipt by clinic + key (no lock).
	FindByIdempotencyKey(ctx context.Context, clinicID uint64, key uuid.UUID) (*model.LabImportRevertReceipt, error)
	// LockByIdempotencyKey locks the receipt row if present. Requires ambient tx.
	LockByIdempotencyKey(ctx context.Context, clinicID uint64, key uuid.UUID) (*model.LabImportRevertReceipt, error)
	Create(ctx context.Context, receipt *model.LabImportRevertReceipt) error
}

// LabImportRetractionRepository は retraction snapshots。
type LabImportRetractionRepository interface {
	CreateWithItems(ctx context.Context, retraction *model.LabImportExamRetraction, items []model.LabImportExamRetractionItem) error
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
	if err := persistence.DBOrTx(ctx, r.db).Create(job).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_job", "")
	}
	return nil
}

func (r *labImportJobRepository) Update(ctx context.Context, job *model.LabImportJob) error {
	result := persistence.DBOrTx(ctx, r.db).
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
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		First(&job).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_job", id.String())
	}
	return &job, nil
}

func (r *labImportJobRepository) LockByIDForUpdate(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("lab import job lock requires an ambient transaction")
	}
	var job model.LabImportJob
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		First(&job).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_job", id.String())
	}
	return &job, nil
}

func (r *labImportJobRepository) CompareAndSetStatus(
	ctx context.Context,
	clinicID uint64,
	id uuid.UUID,
	from, to model.LabImportJobStatus,
	finishedAt *time.Time,
) (int64, error) {
	if persistence.TxFromContext(ctx) == nil {
		return 0, apperrors.WrapInternalServerError("lab import job CAS requires an ambient transaction")
	}
	updates := map[string]any{"status": to}
	if finishedAt != nil {
		updates["finished_at"] = finishedAt
	}
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.LabImportJob{}).
		Where("clinic_id = ? AND id = ? AND status = ?", clinicID, id, from).
		Updates(updates)
	if result.Error != nil {
		return 0, apperrors.FromGORM(result.Error, "lab_import_job", id.String())
	}
	return result.RowsAffected, nil
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
	if err := persistence.DBOrTx(ctx, r.db).Create(event).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_event", "")
	}
	return nil
}

func (r *labImportEventRepository) FindByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
	var events []*model.LabImportEvent
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND job_id = ?", clinicID, jobID).
		Order("created_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_event", jobID.String())
	}
	return events, nil
}

func (r *labImportEventRepository) HasEventType(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	eventType model.LabImportEventType,
) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.LabImportEvent{}).
		Where("clinic_id = ? AND job_id = ? AND event_type = ?", clinicID, jobID, eventType).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "lab_import_event", jobID.String())
	}
	return count > 0, nil
}

// ------------------------------------
// LabImportUsageReceiptRepository 実装
// ------------------------------------

type labImportUsageReceiptRepository struct{ db *gorm.DB }

// NewLabImportUsageReceiptRepository は LabImportUsageReceiptRepository を返す。
func NewLabImportUsageReceiptRepository(db *gorm.DB) LabImportUsageReceiptRepository {
	return &labImportUsageReceiptRepository{db: db}
}

func (r *labImportUsageReceiptRepository) Create(ctx context.Context, receipt *model.LabImportUsageReceipt) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(receipt).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_usage_receipt", "")
	}
	return nil
}

func (r *labImportUsageReceiptRepository) LockByJobForUpdate(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
) ([]*model.LabImportUsageReceipt, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("lab import usage receipt lock requires an ambient transaction")
	}
	var receipts []*model.LabImportUsageReceipt
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("clinic_id = ? AND job_id = ?", clinicID, jobID).
		Order("id ASC").
		Find(&receipts).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_usage_receipt", jobID.String())
	}
	return receipts, nil
}

func (r *labImportUsageReceiptRepository) CountByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.LabImportUsageReceipt{}).
		Where("clinic_id = ? AND job_id = ?", clinicID, jobID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "lab_import_usage_receipt", jobID.String())
	}
	return count, nil
}

func (r *labImportUsageReceiptRepository) CountManualMutationByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.LabImportUsageReceipt{}).
		Where("clinic_id = ? AND job_id = ? AND use_kind = ?", clinicID, jobID, model.LabImportUsageKindManualMutation).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "lab_import_usage_receipt", jobID.String())
	}
	return count, nil
}

// ------------------------------------
// LabImportRevertReceiptRepository 実装
// ------------------------------------

type labImportRevertReceiptRepository struct{ db *gorm.DB }

// NewLabImportRevertReceiptRepository は LabImportRevertReceiptRepository を返す。
func NewLabImportRevertReceiptRepository(db *gorm.DB) LabImportRevertReceiptRepository {
	return &labImportRevertReceiptRepository{db: db}
}

func (r *labImportRevertReceiptRepository) FindByIdempotencyKey(
	ctx context.Context,
	clinicID uint64,
	key uuid.UUID,
) (*model.LabImportRevertReceipt, error) {
	var receipt model.LabImportRevertReceipt
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND idempotency_key = ?", clinicID, key).
		First(&receipt).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_revert_receipt", key.String())
	}
	return &receipt, nil
}

func (r *labImportRevertReceiptRepository) LockByIdempotencyKey(
	ctx context.Context,
	clinicID uint64,
	key uuid.UUID,
) (*model.LabImportRevertReceipt, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("lab import revert receipt lock requires an ambient transaction")
	}
	var receipt model.LabImportRevertReceipt
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("clinic_id = ? AND idempotency_key = ?", clinicID, key).
		First(&receipt).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_revert_receipt", key.String())
	}
	return &receipt, nil
}

func (r *labImportRevertReceiptRepository) Create(ctx context.Context, receipt *model.LabImportRevertReceipt) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(receipt).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_revert_receipt", "")
	}
	return nil
}

// ------------------------------------
// LabImportRetractionRepository 実装
// ------------------------------------

type labImportRetractionRepository struct{ db *gorm.DB }

// NewLabImportRetractionRepository は LabImportRetractionRepository を返す。
func NewLabImportRetractionRepository(db *gorm.DB) LabImportRetractionRepository {
	return &labImportRetractionRepository{db: db}
}

func (r *labImportRetractionRepository) CreateWithItems(
	ctx context.Context,
	retraction *model.LabImportExamRetraction,
	items []model.LabImportExamRetractionItem,
) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("lab import retraction write requires an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Create(retraction).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_exam_retraction", "")
	}
	for i := range items {
		items[i].RetractionID = retraction.ID
		items[i].ClinicID = retraction.ClinicID
		items[i].JobID = retraction.JobID
		items[i].ExamID = retraction.ExamID
		if err := db.Create(&items[i]).Error; err != nil {
			return apperrors.FromGORM(err, "lab_import_exam_retraction_item", "")
		}
	}
	return nil
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
	q := persistence.DBOrTx(ctx, c.db).
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

// SoftDeleteExamByJob は clinic+exam+job で active exam を条件付き soft delete する。
// RowsAffected を返す。Requires ambient tx.
func SoftDeleteExamByJob(ctx context.Context, db *gorm.DB, clinicID, examID uint64, jobID uuid.UUID) (int64, error) {
	if persistence.TxFromContext(ctx) == nil {
		return 0, apperrors.WrapInternalServerError("lab import exam soft delete requires an ambient transaction")
	}
	result := persistence.DBOrTx(ctx, db).
		Model(&model.Examination{}).
		Where("clinic_id = ? AND id = ? AND job_id = ? AND deleted_at IS NULL", clinicID, examID, jobID).
		Delete(&model.Examination{})
	if result.Error != nil {
		return 0, apperrors.FromGORM(result.Error, "exam", "")
	}
	return result.RowsAffected, nil
}

// CountMedicalRecordImagesByExam counts durable image relations for an exam.
// medical_record_images has no clinic_id column; scope via medical_records join.
func CountMedicalRecordImagesByExam(ctx context.Context, db *gorm.DB, clinicID, examID uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, db).
		Table("medical_record_images").
		Joins("JOIN medical_records ON medical_records.id = medical_record_images.medical_record_id").
		Where("medical_record_images.exam_id = ? AND medical_records.clinic_id = ?", examID, clinicID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "medical_record_image", "")
	}
	return count, nil
}

// MustJSON is a small helper for retraction snapshots.
func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
