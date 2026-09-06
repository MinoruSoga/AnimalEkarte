package medicalrecord

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// labImportTransitions は許可された状態遷移テーブル。
// 変更時は lab_import_service_test.go の TestCanTransitionTo_* も更新すること。
//
// needs_review は意図的に非終端状態: FinishedAt が nil のまま残る。
// オペレーターが validated または failed に遷移させるまでジョブはオープンのまま。
// Phase 1 以降でタイムアウトスイープを実装する場合は FinishedAt IS NULL を利用すること。
var labImportTransitions = map[model.LabImportJobStatus][]model.LabImportJobStatus{
	model.LabImportJobStatusReceived:  {model.LabImportJobStatusValidated, model.LabImportJobStatusFailed},
	model.LabImportJobStatusValidated: {model.LabImportJobStatusMapped, model.LabImportJobStatusNeedsReview, model.LabImportJobStatusFailed},
	model.LabImportJobStatusMapped:    {model.LabImportJobStatusPersisted, model.LabImportJobStatusDuplicate, model.LabImportJobStatusNeedsReview, model.LabImportJobStatusFailed},
	// TASK-032: persisted → reverted is the sole compensating terminal transition.
	// It is owned by RevertLabImport, not TransitionStatus (different permission/endpoint).
	model.LabImportJobStatusPersisted:   {model.LabImportJobStatusReverted},
	model.LabImportJobStatusDuplicate:   {},
	model.LabImportJobStatusNeedsReview: {model.LabImportJobStatusValidated, model.LabImportJobStatusFailed},
	model.LabImportJobStatusFailed:      {model.LabImportJobStatusReceived},
	model.LabImportJobStatusReverted:    {},
}

// CanTransitionTo は from → to の遷移が許可されているかを返す。
func CanTransitionTo(from, to model.LabImportJobStatus) bool {
	allowed, ok := labImportTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// LabImportJobService は lab_import_jobs のドメインサービス。
// Phase 0: フィクスチャ入力のみ。外部 MDB・機器接続・credential は BLOCKED。
type LabImportJobService interface {
	// CreateJob は新規ジョブを received 状態で作成する。
	CreateJob(ctx context.Context, clinicID uint64, batch model.LabInboundBatch) (*model.LabImportJob, error)
	// TransitionStatus はジョブ状態を遷移させ、監査イベントを記録する。
	// 不正遷移の場合は ErrInvalidInput を返す。
	TransitionStatus(ctx context.Context, clinicID uint64, jobID uuid.UUID, to model.LabImportJobStatus, counts TransitionCounts) (*model.LabImportJob, error)
	// GetJob はクリニックスコープでジョブを返す。
	GetJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) (*model.LabImportJob, error)
	// PreviewBatch はバッチ内容を検証し、プレビューサマリを返す（永続化しない）。
	PreviewBatch(ctx context.Context, clinicID uint64, batch model.LabInboundBatch) (*model.LabImportPreviewResponse, error)
	// ListEvents はクリニックスコープでジョブのイベント一覧を返す（作成昇順）。
	ListEvents(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error)
}

// TransitionCounts は状態遷移時に更新するカウンタ群。
type TransitionCounts struct {
	RowCount         int
	PersistedCount   int
	DuplicateCount   int
	NeedsReviewCount int
	FailedCount      int
	ErrorCode        *string
	ErrorMessage     *string
}

type labImportJobService struct {
	jobRepo   LabImportJobRepository
	eventRepo LabImportEventRepository
}

// NewLabImportJobService は LabImportJobService を初期化して返す。
func NewLabImportJobService(
	jobRepo LabImportJobRepository,
	eventRepo LabImportEventRepository,
) LabImportJobService {
	return &labImportJobService{
		jobRepo:   jobRepo,
		eventRepo: eventRepo,
	}
}

func (s *labImportJobService) CreateJob(ctx context.Context, clinicID uint64, batch model.LabInboundBatch) (*model.LabImportJob, error) {
	now := time.Now()
	job := &model.LabImportJob{
		ClinicID:          clinicID,
		SourceType:        batch.SourceType,
		SourceFingerprint: batch.SourceFingerprint,
		Status:            model.LabImportJobStatusReceived,
		RowCount:          len(batch.ResultRows),
		StartedAt:         &now,
	}
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, apperrors.Wrap(err, "failed to create lab import job")
	}
	event := &model.LabImportEvent{
		ClinicID:  clinicID,
		JobID:     job.ID,
		EventType: model.LabImportEventTypeStatusTransition,
		ToStatus:  ptr(model.LabImportJobStatusReceived),
		RowCount:  len(batch.ResultRows),
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, apperrors.Wrap(err, "failed to append lab import event")
	}
	return job, nil
}

func (s *labImportJobService) TransitionStatus(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	to model.LabImportJobStatus,
	counts TransitionCounts,
) (*model.LabImportJob, error) {
	job, err := s.jobRepo.FindByID(ctx, clinicID, jobID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find lab import job")
	}
	from := job.Status
	// TASK-032: compensating revert is a dedicated endpoint/state machine, not TransitionStatus.
	if to == model.LabImportJobStatusReverted {
		return nil, apperrors.WrapInvalidInput(
			"persisted → reverted must use POST /lab-imports/:id/revert, not TransitionStatus",
		)
	}
	if !CanTransitionTo(from, to) {
		return nil, apperrors.WrapInvalidInput(
			fmt.Sprintf("invalid lab import job transition: %s → %s", from, to),
		)
	}

	event := &model.LabImportEvent{
		ClinicID:         clinicID,
		JobID:            job.ID,
		EventType:        model.LabImportEventTypeStatusTransition,
		FromStatus:       ptr(from),
		ToStatus:         ptr(to),
		RowCount:         counts.RowCount,
		PersistedCount:   counts.PersistedCount,
		DuplicateCount:   counts.DuplicateCount,
		NeedsReviewCount: counts.NeedsReviewCount,
		ErrorCode:        counts.ErrorCode,
	}

	job.Status = to
	job.RowCount = counts.RowCount
	job.PersistedCount = counts.PersistedCount
	job.DuplicateCount = counts.DuplicateCount
	job.NeedsReviewCount = counts.NeedsReviewCount
	job.FailedCount = counts.FailedCount
	job.ErrorCode = counts.ErrorCode
	job.ErrorMessage = counts.ErrorMessage

	switch to {
	case model.LabImportJobStatusPersisted, model.LabImportJobStatusDuplicate, model.LabImportJobStatusFailed, model.LabImportJobStatusReverted:
		now := time.Now()
		job.FinishedAt = &now
	case model.LabImportJobStatusReceived:
		// retry: clear stale FinishedAt from the previous failed terminal
		job.FinishedAt = nil
	}

	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, apperrors.Wrap(err, "failed to update lab import job")
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, apperrors.Wrap(err, "failed to append lab import event")
	}
	return job, nil
}

func (s *labImportJobService) GetJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) (*model.LabImportJob, error) {
	job, err := s.jobRepo.FindByID(ctx, clinicID, jobID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get lab import job")
	}
	if job.SourceType == model.LabImportSourceTypeDrWan {
		return nil, apperrors.WrapInvalidInput("source_type=drwan cannot be opened")
	}
	return job, nil
}

// PreviewBatch は入力バッチを検証してサマリを返す（DB 書き込みなし）。
// Phase 0 では fixture ソース以外は blocked_reasons に列挙する。
func (s *labImportJobService) PreviewBatch(_ context.Context, _ uint64, batch model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
	resp := &model.LabImportPreviewResponse{
		RowCount:        len(batch.ResultRows),
		MappingWarnings: []string{},
		BlockedReasons:  []string{},
	}

	switch batch.SourceType {
	case model.LabImportSourceTypeDrWan:
		resp.BlockedReasons = append(resp.BlockedReasons,
			"source_type=drwan is BLOCKED: Dr.Wan MDB schema not confirmed; external device access not permitted in Phase 0",
		)
	case model.LabImportSourceTypeManual:
		resp.BlockedReasons = append(resp.BlockedReasons,
			"source_type=manual is not yet implemented; available from Phase 2",
		)
	case model.LabImportSourceTypeFujiNX600, model.LabImportSourceTypeFujiAU10V, model.LabImportSourceTypeArkrayPU4010:
		resp.BlockedReasons = append(resp.BlockedReasons,
			"device source_type cannot use preview/commit; use /lab-device/frames",
		)
	}

	for i := range batch.ResultRows {
		if batch.ResultRows[i].ExamCode == "" {
			resp.MappingWarnings = append(resp.MappingWarnings,
				fmt.Sprintf("row[%d]: exam_code is empty — crosswalk will use fallback exam type", i),
			)
		}
	}

	return resp, nil
}

func (s *labImportJobService) ListEvents(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
	if _, err := s.jobRepo.FindByID(ctx, clinicID, jobID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find lab import job")
	}
	events, err := s.eventRepo.FindByJob(ctx, clinicID, jobID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list lab import events")
	}
	return events, nil
}

func ptr[T any](v T) *T { return &v }
