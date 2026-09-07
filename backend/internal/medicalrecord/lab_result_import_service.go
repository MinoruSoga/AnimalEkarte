package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LabResultImportService は fixture-only の lab import 全体フローを調整するサービス。
//
// Phase 2 スコープ:
//   - fixture ソース (source_type=fixture) のみ commit を許可する。
//   - drwan / manual ソースは preview のみ可能で commit は BLOCKED。
//   - Preview: バッチ内容を検証し、row_count / warnings / blocked_reasons を返す（DB 書き込みなし）。
//   - Commit: job 作成 → received/validated/mapped/persisted/failed の状態遷移 → exam 永続化。
//   - 各 LabExamPersistInput は Commit 呼び出し前に clinic_id / exam_type_id / date の
//     基本バリデーションが済んでいる前提（PersistExam 内でも再検証される）。
//   - JobID は Commit が発行した job.ID で上書きする（呼び出し元が設定した値は無視）。
//
// Phase BLOCKED:
//   - Dr.Wan MDB 接続 / raw デバイスペイロード解析（外部スキーマ未確認）
//   - 外部 credentials / ネットワーク I/O
//   - API エンドポイント (Phase 3)
type LabResultImportService interface {
	// Preview はバッチ内容を検証し、保存前のサマリを返す（DB 書き込みなし）。
	// fixture / drwan / manual すべてのソースで呼び出し可能。
	// drwan / manual は blocked_reasons に列挙する。
	Preview(ctx context.Context, clinicID uint64, batch model.LabInboundBatch) (*model.LabImportPreviewResponse, error)

	// Commit は fixture ソースのバッチを受け取り、job を作成して exams + exam_results を永続化する。
	// inputs は呼び出し元がバッチ行から変換した LabExamPersistInput スライス。
	// Commit は inputs の JobID フィールドを発行した job.ID で上書きする。
	// drwan / manual ソースは ErrInvalidInput を返す。
	Commit(ctx context.Context, clinicID uint64, batch model.LabInboundBatch, inputs []LabExamPersistInput) (*model.LabImportCommitResponse, error)
}

type labResultImportService struct {
	jobSvc    LabImportJobService
	examSvc   LabImportExaminationService
	eventRepo LabImportEventRepository // optional: records usage_tracking_started on persisted (TASK-032)
	tx        Transactor               // optional: when set, terminal status + usage marker share one tx
}

// NewLabResultImportService は LabResultImportService を初期化して返す。
// eventRepo may be nil; when set, a successful persisted terminal transition also writes
// usage_tracking_started so compensating revert can measure downstream use.
// When tx is non-nil, terminal TransitionStatus and usage_tracking_started run in one WithTx
// so a marker failure cannot leave a committed persisted job without the marker (TASK-032).
func NewLabResultImportService(
	jobSvc LabImportJobService,
	examSvc LabImportExaminationService,
	eventRepo ...LabImportEventRepository,
) LabResultImportService {
	var events LabImportEventRepository
	if len(eventRepo) > 0 {
		events = eventRepo[0]
	}
	return &labResultImportService{jobSvc: jobSvc, examSvc: examSvc, eventRepo: events}
}

// WithTransactor attaches a Transactor for atomic terminal status + usage marker writes.
func (s *labResultImportService) WithTransactor(tx Transactor) *labResultImportService {
	s.tx = tx
	return s
}

// NewLabResultImportServiceWithTx is the composition-root constructor that binds eventRepo and Transactor.
func NewLabResultImportServiceWithTx(
	jobSvc LabImportJobService,
	examSvc LabImportExaminationService,
	eventRepo LabImportEventRepository,
	tx Transactor,
) LabResultImportService {
	return &labResultImportService{jobSvc: jobSvc, examSvc: examSvc, eventRepo: eventRepo, tx: tx}
}

func (s *labResultImportService) Preview(ctx context.Context, clinicID uint64, batch model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
	return s.jobSvc.PreviewBatch(ctx, clinicID, batch)
}

// summarizeLabBatchResults は persisted/duplicate/failed カウントと終端ステータスを算出する
// 純関数（BE-refactor.md E-9）。
//
// 終端状態の決定:
//   - inputs が 0 件: 空コミット → duplicate（何も書かなかった = 冪等）
//   - 1 件でも persisted があれば persisted
//   - 全行 duplicate（persisted=0, failed=0）→ duplicate
//   - それ以外（全行 failed、または duplicate+failed 混在でも persisted=0）→ failed
func summarizeLabBatchResults(inputCount int, batchResults []*LabExamPersistResult) (model.LabImportJobStatus, TransitionCounts) {
	persisted, duplicate, failed := 0, 0, 0
	for _, res := range batchResults {
		switch {
		case res.RowError != nil:
			failed++
		case res.Duplicate:
			duplicate++
		default:
			persisted++
		}
	}

	var termStatus model.LabImportJobStatus
	switch {
	case inputCount == 0:
		termStatus = model.LabImportJobStatusDuplicate
	case persisted > 0:
		termStatus = model.LabImportJobStatusPersisted
	case duplicate > 0 && failed == 0:
		termStatus = model.LabImportJobStatusDuplicate
	default:
		termStatus = model.LabImportJobStatusFailed
	}

	finalCounts := TransitionCounts{
		RowCount:       inputCount,
		PersistedCount: persisted,
		DuplicateCount: duplicate,
		FailedCount:    failed,
	}
	return termStatus, finalCounts
}

// Commit は fixture ソース専用の commit フロー。
//
// フロー:
//  1. drwan / manual ソースは即座に ErrInvalidInput を返す。
//  2. 全 inputs の clinic_id を clinicID と照合する（clinic scope 事前チェック）。
//  3. job を received 状態で作成する。
//  4. received → validated → mapped に遷移する（fixture バッチは mapping warning のみで進む）。
//  5. inputs に job.ID を注入して PersistBatch を呼び出す。
//  6. 結果を集計して job を終端状態（persisted / duplicate / failed）に遷移させる。
//  7. LabImportCommitResponse を返す。
//
// 個別行のエラーは failed_count に加算する（関数エラーとして返さない）。
// job 作成失敗は関数エラーとして返す（job が存在しないため追記不可）。
func (s *labResultImportService) Commit(ctx context.Context, clinicID uint64, batch model.LabInboundBatch, inputs []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
	// Phase 2: fixture ソース以外は commit を拒否する。
	if batch.SourceType != model.LabImportSourceTypeFixture {
		return nil, apperrors.WrapInvalidInput(
			fmt.Sprintf("commit is only allowed for source_type=fixture in Phase 2; source_type=%s is BLOCKED", batch.SourceType),
		)
	}

	// 全 inputs の clinic scope 事前チェック。
	for i, inp := range inputs {
		if inp.ClinicID != clinicID {
			return nil, apperrors.WrapInvalidInput(
				fmt.Sprintf("input[%d].clinic_id=%d does not match commit clinic_id=%d", i, inp.ClinicID, clinicID),
			)
		}
	}

	// job 作成 (received)
	job, err := s.jobSvc.CreateJob(ctx, clinicID, batch)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create lab import job")
	}
	jobID := job.ID

	// received → validated
	if _, err := s.jobSvc.TransitionStatus(ctx, clinicID, jobID, model.LabImportJobStatusValidated,
		TransitionCounts{RowCount: len(inputs)}); err != nil {
		return nil, apperrors.Wrap(err, "failed to transition job to validated")
	}

	// validated → mapped (fixture: mapping is trivial, no blocking warnings)
	if _, err := s.jobSvc.TransitionStatus(ctx, clinicID, jobID, model.LabImportJobStatusMapped,
		TransitionCounts{RowCount: len(inputs)}); err != nil {
		return nil, apperrors.Wrap(err, "failed to transition job to mapped")
	}

	// JobID を注入して PersistBatch を実行する。
	// 呼び出し元が設定した JobID は上書きする（job はこの Commit が発行したものに限定）。
	injected := make([]LabExamPersistInput, len(inputs))
	for i, inp := range inputs {
		inp.JobID = jobID
		injected[i] = inp
	}

	batchResults, err := s.examSvc.PersistBatch(ctx, injected)
	if err != nil {
		// context キャンセル等のシステムエラー: job を failed に遷移させてから返す。
		// ctx は既にキャンセル済みのため、補償トランザクションには新しいコンテキストを使う。
		errMsg := err.Error()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cleanupCancel()
		if _, compErr := s.jobSvc.TransitionStatus(cleanupCtx, clinicID, jobID, model.LabImportJobStatusFailed,
			TransitionCounts{RowCount: len(inputs), ErrorCode: ptr("context_cancelled"), ErrorMessage: &errMsg}); compErr != nil {
			slog.ErrorContext(cleanupCtx, "lab result import: failed to transition to failed (compensation)",
				"error", compErr, "job_id", jobID)
			// 主エラーを優先して返す(挙動不変)。job は非終端で残るため jobID から追跡する。
		}
		return nil, apperrors.Wrap(err, "lab import batch interrupted")
	}

	// 結果集計 + 終端遷移。persisted 時は usage_tracking_started を同一 transaction で書く（TASK-032）。
	termStatus, finalCounts := summarizeLabBatchResults(len(inputs), batchResults)
	if err := s.transitionTerminalWithUsageMarker(ctx, clinicID, jobID, termStatus, finalCounts); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "lab result import commit complete",
		slog.String("job_id", jobID.String()),
		slog.Int("persisted", finalCounts.PersistedCount),
		slog.Int("duplicate", finalCounts.DuplicateCount),
		slog.Int("failed", finalCounts.FailedCount),
	)

	return &model.LabImportCommitResponse{
		JobID:            jobID,
		PersistedCount:   finalCounts.PersistedCount,
		DuplicateCount:   finalCounts.DuplicateCount,
		NeedsReviewCount: 0,
		FailedCount:      finalCounts.FailedCount,
	}, nil
}

// transitionTerminalWithUsageMarker writes the terminal status transition and, when
// status is persisted, the usage_tracking_started marker. Both writes share one
// ambient transaction when s.tx is configured so a marker failure cannot leave a
// committed persisted job without the marker (and cannot reverse a committed success).
func (s *labResultImportService) transitionTerminalWithUsageMarker(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	termStatus model.LabImportJobStatus,
	finalCounts TransitionCounts,
) error {
	run := func(txCtx context.Context) error {
		if _, err := s.jobSvc.TransitionStatus(txCtx, clinicID, jobID, termStatus, finalCounts); err != nil {
			return apperrors.Wrap(err, "failed to transition job to terminal status")
		}
		if termStatus != model.LabImportJobStatusPersisted {
			return nil
		}
		// Production composition always wires eventRepo. Nil is only for pure unit stubs;
		// without a marker the job is usage_unknown and revert is refused (safe default).
		if s.eventRepo == nil {
			slog.WarnContext(txCtx, "lab result import: event repository nil; skipping usage_tracking_started",
				"job_id", jobID)
			return nil
		}
		if err := s.eventRepo.Create(txCtx, &model.LabImportEvent{
			ClinicID:  clinicID,
			JobID:     jobID,
			EventType: model.LabImportEventTypeUsageTrackingStarted,
			RowCount:  finalCounts.PersistedCount,
		}); err != nil {
			return apperrors.Wrap(err, "failed to record usage_tracking_started")
		}
		return nil
	}

	if s.tx == nil {
		// Test / legacy path without Transactor: still sequential fail-closed, but not atomic.
		return run(ctx)
	}
	return s.tx.WithTx(ctx, run)
}
