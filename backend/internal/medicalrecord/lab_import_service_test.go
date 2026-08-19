package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ------------------------------------
// State machine tests
// ------------------------------------

// TestCanTransitionTo_ValidTransitions は仕様書に定義された全許可遷移を確認する。
func TestCanTransitionTo_ValidTransitions(t *testing.T) {
	cases := []struct {
		from model.LabImportJobStatus
		to   model.LabImportJobStatus
	}{
		{model.LabImportJobStatusReceived, model.LabImportJobStatusValidated},
		{model.LabImportJobStatusReceived, model.LabImportJobStatusFailed},
		{model.LabImportJobStatusValidated, model.LabImportJobStatusMapped},
		{model.LabImportJobStatusValidated, model.LabImportJobStatusNeedsReview},
		{model.LabImportJobStatusValidated, model.LabImportJobStatusFailed},
		{model.LabImportJobStatusMapped, model.LabImportJobStatusPersisted},
		{model.LabImportJobStatusMapped, model.LabImportJobStatusDuplicate},
		{model.LabImportJobStatusMapped, model.LabImportJobStatusNeedsReview},
		{model.LabImportJobStatusMapped, model.LabImportJobStatusFailed},
		// TASK-032: compensating terminal transition (owned by Revert endpoint, not TransitionStatus).
		{model.LabImportJobStatusPersisted, model.LabImportJobStatusReverted},
		{model.LabImportJobStatusNeedsReview, model.LabImportJobStatusValidated},
		{model.LabImportJobStatusNeedsReview, model.LabImportJobStatusFailed},
		{model.LabImportJobStatusFailed, model.LabImportJobStatusReceived},
	}
	for _, c := range cases {
		if !CanTransitionTo(c.from, c.to) {
			t.Errorf("expected transition %s → %s to be allowed", c.from, c.to)
		}
	}
}

// TestCanTransitionTo_InvalidTransitions は不正遷移が拒否されることを確認する。
func TestCanTransitionTo_InvalidTransitions(t *testing.T) {
	cases := []struct {
		from model.LabImportJobStatus
		to   model.LabImportJobStatus
	}{
		// terminal states cannot transition (except persisted → reverted via CanTransitionTo)
		{model.LabImportJobStatusPersisted, model.LabImportJobStatusValidated},
		{model.LabImportJobStatusPersisted, model.LabImportJobStatusFailed},
		{model.LabImportJobStatusReverted, model.LabImportJobStatusPersisted},
		{model.LabImportJobStatusReverted, model.LabImportJobStatusReceived},
		{model.LabImportJobStatusDuplicate, model.LabImportJobStatusValidated},
		{model.LabImportJobStatusDuplicate, model.LabImportJobStatusFailed},
		// backward skips
		{model.LabImportJobStatusValidated, model.LabImportJobStatusReceived},
		{model.LabImportJobStatusMapped, model.LabImportJobStatusReceived},
		{model.LabImportJobStatusMapped, model.LabImportJobStatusValidated},
		// direct skip from received to terminal
		{model.LabImportJobStatusReceived, model.LabImportJobStatusPersisted},
		{model.LabImportJobStatusReceived, model.LabImportJobStatusDuplicate},
		{model.LabImportJobStatusReceived, model.LabImportJobStatusNeedsReview},
		// self-transitions
		{model.LabImportJobStatusReceived, model.LabImportJobStatusReceived},
		{model.LabImportJobStatusFailed, model.LabImportJobStatusFailed},
	}
	for _, c := range cases {
		if CanTransitionTo(c.from, c.to) {
			t.Errorf("expected transition %s → %s to be DENIED", c.from, c.to)
		}
	}
}

// TestCanTransitionTo_UnknownFromStatus は未知の from 状態を false にする。
func TestCanTransitionTo_UnknownFromStatus(t *testing.T) {
	if CanTransitionTo("ghost_state", model.LabImportJobStatusValidated) {
		t.Error("unknown from_status should return false")
	}
}

// ------------------------------------
// Service unit tests (nil-repo, pre-DB guard)
// ------------------------------------

// TestLabImportJobService_PreviewBatch_Fixture は fixture ソースで blocked_reasons が空であることを確認する。
func TestLabImportJobService_PreviewBatch_Fixture(t *testing.T) {
	svc := &labImportJobService{jobRepo: nil, eventRepo: nil}
	batch := model.LabInboundBatch{
		SourceType:        model.LabImportSourceTypeFixture,
		SourceFingerprint: "sha256:abc123",
		ReceivedAt:        time.Now(),
		ResultRows: []model.LabInboundResultRow{
			{OldPetKey: "P001", ExamCode: "BUN", ExamName: "BUN", ItemName: "BUN", DisplayValue: "12.3"},
		},
	}
	resp, err := svc.PreviewBatch(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RowCount != 1 {
		t.Errorf("expected row_count=1, got %d", resp.RowCount)
	}
	if len(resp.BlockedReasons) != 0 {
		t.Errorf("fixture source should have no blocked_reasons, got %v", resp.BlockedReasons)
	}
	if len(resp.MappingWarnings) != 0 {
		t.Errorf("row with exam_code should have no mapping_warnings, got %v", resp.MappingWarnings)
	}
}

// TestLabImportJobService_PreviewBatch_DrWan は drwan ソースが blocked_reasons に列挙されることを確認する。
func TestLabImportJobService_PreviewBatch_DrWan(t *testing.T) {
	svc := &labImportJobService{jobRepo: nil, eventRepo: nil}
	batch := model.LabInboundBatch{
		SourceType: model.LabImportSourceTypeDrWan,
		ReceivedAt: time.Now(),
	}
	resp, err := svc.PreviewBatch(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BlockedReasons) == 0 {
		t.Error("drwan source must produce at least one blocked_reason")
	}
}

func TestLabImportJobService_PreviewBatch_DeviceSourcesBlocked(t *testing.T) {
	svc := &labImportJobService{}
	for _, source := range []model.LabImportSourceType{
		model.LabImportSourceTypeFujiNX600,
		model.LabImportSourceTypeFujiAU10V,
		model.LabImportSourceTypeArkrayPU4010,
	} {
		resp, err := svc.PreviewBatch(context.Background(), 1, model.LabInboundBatch{SourceType: source, ReceivedAt: time.Now()})
		if err != nil {
			t.Fatalf("%s: %v", source, err)
		}
		if len(resp.BlockedReasons) == 0 {
			t.Errorf("%s must be blocked on preview/commit", source)
		}
	}
}

// TestLabImportJobService_PreviewBatch_MissingExamCode は exam_code 空行が mapping_warnings に現れることを確認する。
func TestLabImportJobService_PreviewBatch_MissingExamCode(t *testing.T) {
	svc := &labImportJobService{jobRepo: nil, eventRepo: nil}
	batch := model.LabInboundBatch{
		SourceType: model.LabImportSourceTypeFixture,
		ReceivedAt: time.Now(),
		ResultRows: []model.LabInboundResultRow{
			{OldPetKey: "P002", ExamCode: "", ExamName: "unknown", ItemName: "unknown", DisplayValue: "0"},
		},
	}
	resp, err := svc.PreviewBatch(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.MappingWarnings) == 0 {
		t.Error("empty exam_code must produce at least one mapping_warning")
	}
}

// ------------------------------------
// TransitionStatus: invalid transition guard test (no DB required)
// ------------------------------------

// stubJobRepo はテスト用のインメモリ LabImportJobRepository。
// createErr/updateErr/findByIDErr はゼロ値(nil)の場合は通常動作となり、
// 既存テストの挙動には影響しない。エラー注入が必要な新規テストのみが設定する。
type stubJobRepo struct {
	jobs        map[uuid.UUID]*model.LabImportJob
	createErr   error
	updateErr   error
	findByIDErr error
}

func newStubJobRepo() *stubJobRepo {
	return &stubJobRepo{jobs: make(map[uuid.UUID]*model.LabImportJob)}
}

func (r *stubJobRepo) Create(_ context.Context, job *model.LabImportJob) error {
	if r.createErr != nil {
		return r.createErr
	}
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	cp := *job
	r.jobs[job.ID] = &cp
	return nil
}

func (r *stubJobRepo) Update(_ context.Context, job *model.LabImportJob) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	if _, ok := r.jobs[job.ID]; !ok {
		return apperrors.WrapNotFound("lab_import_job", job.ID.String())
	}
	cp := *job
	r.jobs[job.ID] = &cp
	return nil
}

func (r *stubJobRepo) FindByID(_ context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error) {
	if r.findByIDErr != nil {
		return nil, r.findByIDErr
	}
	j, ok := r.jobs[id]
	if !ok || j.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("lab_import_job", id.String())
	}
	cp := *j
	return &cp, nil
}

func (r *stubJobRepo) LockByIDForUpdate(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error) {
	return r.FindByID(ctx, clinicID, id)
}

func (r *stubJobRepo) CompareAndSetStatus(
	_ context.Context,
	clinicID uint64,
	id uuid.UUID,
	from, to model.LabImportJobStatus,
	finishedAt *time.Time,
) (int64, error) {
	j, ok := r.jobs[id]
	if !ok || j.ClinicID != clinicID || j.Status != from {
		return 0, nil
	}
	j.Status = to
	if finishedAt != nil {
		j.FinishedAt = finishedAt
	}
	return 1, nil
}

// stubEventRepo はテスト用のインメモリ LabImportEventRepository。
// createErr/findByJobErr はゼロ値(nil)の場合は通常動作となり、既存テストの挙動には影響しない。
type stubEventRepo struct {
	events       []*model.LabImportEvent
	createErr    error
	findByJobErr error
}

func (r *stubEventRepo) Create(_ context.Context, event *model.LabImportEvent) error {
	if r.createErr != nil {
		return r.createErr
	}
	cp := *event
	r.events = append(r.events, &cp)
	return nil
}

func (r *stubEventRepo) FindByJob(_ context.Context, _ uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
	if r.findByJobErr != nil {
		return nil, r.findByJobErr
	}
	var out []*model.LabImportEvent
	for _, e := range r.events {
		if e.JobID == jobID {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *stubEventRepo) HasEventType(_ context.Context, _ uint64, jobID uuid.UUID, eventType model.LabImportEventType) (bool, error) {
	for _, e := range r.events {
		if e.JobID == jobID && e.EventType == eventType {
			return true, nil
		}
	}
	return false, nil
}

// TestTransitionStatus_ValidPath は received → validated → persisted の正常経路を確認する。
func TestTransitionStatus_ValidPath(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{
		SourceType: model.LabImportSourceTypeFixture,
		ReceivedAt: time.Now(),
		ResultRows: []model.LabInboundResultRow{
			{OldPetKey: "P001", ExamCode: "BUN", DisplayValue: "12.3"},
		},
	}

	job, err := svc.CreateJob(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.Status != model.LabImportJobStatusReceived {
		t.Errorf("expected status=received, got %s", job.Status)
	}

	job, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusValidated, TransitionCounts{RowCount: 1})
	if err != nil {
		t.Fatalf("TransitionStatus received→validated: %v", err)
	}
	if job.Status != model.LabImportJobStatusValidated {
		t.Errorf("expected status=validated, got %s", job.Status)
	}

	job, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusMapped, TransitionCounts{RowCount: 1})
	if err != nil {
		t.Fatalf("TransitionStatus validated→mapped: %v", err)
	}

	job, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusPersisted, TransitionCounts{RowCount: 1, PersistedCount: 1})
	if err != nil {
		t.Fatalf("TransitionStatus mapped→persisted: %v", err)
	}
	if job.Status != model.LabImportJobStatusPersisted {
		t.Errorf("expected status=persisted, got %s", job.Status)
	}
	if job.FinishedAt == nil {
		t.Error("persisted terminal status must set finished_at")
	}

	events, err := eventRepo.FindByJob(context.Background(), 1, job.ID)
	if err != nil {
		t.Fatalf("FindByJob: %v", err)
	}
	if len(events) != 4 {
		t.Errorf("expected 4 audit events (create + 3 transitions), got %d", len(events))
	}
}

// TestTransitionStatus_InvalidTransition は不正遷移が ErrInvalidInput を返すことを確認する。
func TestTransitionStatus_InvalidTransition(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	job, err := svc.CreateJob(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// received → persisted は不正
	_, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusPersisted, TransitionCounts{})
	if err == nil {
		t.Fatal("expected error for invalid transition received → persisted, got nil")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput error, got: %v", err)
	}
}

// TestTransitionStatus_TerminalCannotTransition は terminal 状態からの遷移が拒否されることを確認する。
func TestTransitionStatus_TerminalCannotTransition(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	job, err := svc.CreateJob(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// received → validated → mapped → persisted (terminal)
	for _, to := range []model.LabImportJobStatus{
		model.LabImportJobStatusValidated,
		model.LabImportJobStatusMapped,
		model.LabImportJobStatusPersisted,
	} {
		job, err = svc.TransitionStatus(context.Background(), 1, job.ID, to, TransitionCounts{RowCount: 1})
		if err != nil {
			t.Fatalf("TransitionStatus → %s: %v", to, err)
		}
	}

	// persisted → validated は不正
	_, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusValidated, TransitionCounts{})
	if err == nil {
		t.Fatal("expected error transitioning out of terminal persisted state, got nil")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput error, got: %v", err)
	}
}

// TestTransitionStatus_FailedRetryPath は failed → received → validated の retry 経路を確認する。
func TestTransitionStatus_FailedRetryPath(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	job, err := svc.CreateJob(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// received → failed
	job, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusFailed, TransitionCounts{})
	if err != nil {
		t.Fatalf("received→failed: %v", err)
	}
	if job.FinishedAt == nil {
		t.Error("failed terminal status must set finished_at")
	}

	// failed → received (retry): FinishedAt must be cleared
	job, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusReceived, TransitionCounts{})
	if err != nil {
		t.Fatalf("failed→received (retry): %v", err)
	}
	if job.Status != model.LabImportJobStatusReceived {
		t.Errorf("expected status=received after retry, got %s", job.Status)
	}
	if job.FinishedAt != nil {
		t.Error("retry to received must clear FinishedAt (stale terminal timestamp must not persist)")
	}
}

// TestPreviewBatch_ManualSourceBlocked は manual ソースが blocked_reasons に列挙されることを確認する。
func TestPreviewBatch_ManualSourceBlocked(t *testing.T) {
	svc := &labImportJobService{jobRepo: nil, eventRepo: nil}
	batch := model.LabInboundBatch{
		SourceType: model.LabImportSourceTypeManual,
		ReceivedAt: time.Now(),
	}
	resp, err := svc.PreviewBatch(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BlockedReasons) == 0 {
		t.Error("manual source must produce at least one blocked_reason in Phase 0")
	}
}

// TestLabImportJobStatus_Constants は全ステータス定数が空でないことを確認する（リグレッション）。
func TestLabImportJobStatus_Constants(t *testing.T) {
	statuses := []model.LabImportJobStatus{
		model.LabImportJobStatusReceived,
		model.LabImportJobStatusValidated,
		model.LabImportJobStatusMapped,
		model.LabImportJobStatusPersisted,
		model.LabImportJobStatusDuplicate,
		model.LabImportJobStatusNeedsReview,
		model.LabImportJobStatusFailed,
	}
	for _, s := range statuses {
		if s == "" {
			t.Errorf("status constant must not be empty")
		}
	}
}

// TestLabImportSourceType_Constants はソース種別定数が空でないことを確認する。
func TestLabImportSourceType_Constants(t *testing.T) {
	sources := []model.LabImportSourceType{
		model.LabImportSourceTypeFixture,
		model.LabImportSourceTypeDrWan,
		model.LabImportSourceTypeManual,
	}
	for _, s := range sources {
		if s == "" {
			t.Errorf("source type constant must not be empty")
		}
	}
}

// ------------------------------------
// CreateJob
// ------------------------------------

func TestLabImportJobService_CreateJob_Success(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{
		SourceType:        model.LabImportSourceTypeFixture,
		SourceFingerprint: "sha256:abc123",
		ReceivedAt:        time.Now(),
		ResultRows: []model.LabInboundResultRow{
			{OldPetKey: "P001", ExamCode: "BUN", DisplayValue: "12.3"},
		},
	}

	job, err := svc.CreateJob(context.Background(), 7, batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ClinicID != 7 {
		t.Errorf("expected clinic_id=7, got %d", job.ClinicID)
	}
	if job.Status != model.LabImportJobStatusReceived {
		t.Errorf("expected status=received, got %s", job.Status)
	}
	if job.RowCount != 1 {
		t.Errorf("expected row_count=1, got %d", job.RowCount)
	}
	if job.StartedAt == nil {
		t.Error("expected started_at to be set")
	}
	events, err := eventRepo.FindByJob(context.Background(), 7, job.ID)
	if err != nil {
		t.Fatalf("FindByJob: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 creation event, got %d", len(events))
	}
}

func TestLabImportJobService_CreateJob_JobRepoError(t *testing.T) {
	jobRepo := newStubJobRepo()
	jobRepo.createErr = errors.New("db error")
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	job, err := svc.CreateJob(context.Background(), 1, batch)
	if err == nil {
		t.Fatal("expected error when job repo Create fails, got nil")
	}
	if job != nil {
		t.Error("expected nil job on error")
	}
	if len(eventRepo.events) != 0 {
		t.Error("expected no events to be created when job creation fails")
	}
}

func TestLabImportJobService_CreateJob_EventRepoError(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{createErr: errors.New("db error")}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	job, err := svc.CreateJob(context.Background(), 1, batch)
	if err == nil {
		t.Fatal("expected error when event repo Create fails, got nil")
	}
	if job != nil {
		t.Error("expected nil job on error")
	}
	if len(jobRepo.jobs) != 1 {
		t.Error("expected job to remain persisted even though the audit event append failed")
	}
}

// ------------------------------------
// TransitionStatus: additional error branches
// ------------------------------------

func TestLabImportJobService_TransitionStatus_JobNotFound(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	_, err := svc.TransitionStatus(context.Background(), 1, uuid.New(), model.LabImportJobStatusValidated, TransitionCounts{})
	if err == nil {
		t.Fatal("expected error for unknown job id, got nil")
	}
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected NotFound error, got: %v", err)
	}
}

func TestLabImportJobService_TransitionStatus_JobRepoUpdateError(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	job, err := svc.CreateJob(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	jobRepo.updateErr = errors.New("db error")
	_, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusValidated, TransitionCounts{})
	if err == nil {
		t.Fatal("expected error when job repo Update fails, got nil")
	}
}

func TestLabImportJobService_TransitionStatus_EventRepoError(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	job, err := svc.CreateJob(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	eventRepo.createErr = errors.New("db error")
	_, err = svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusValidated, TransitionCounts{})
	if err == nil {
		t.Fatal("expected error when event repo Create fails, got nil")
	}
}

// ------------------------------------
// GetJob
// ------------------------------------

func TestLabImportJobService_GetJob(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	created, err := svc.CreateJob(context.Background(), 3, batch)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	t.Run("returns job when found", func(t *testing.T) {
		job, err := svc.GetJob(context.Background(), 3, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if job.ID != created.ID {
			t.Errorf("expected job id %s, got %s", created.ID, job.ID)
		}
	})

	t.Run("returns not found for unknown job id", func(t *testing.T) {
		_, err := svc.GetJob(context.Background(), 3, uuid.New())
		if err == nil {
			t.Fatal("expected error for unknown job id, got nil")
		}
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected NotFound error, got: %v", err)
		}
	})

	t.Run("refuses to open drwan", func(t *testing.T) {
		drwan := &model.LabImportJob{
			ID:         uuid.New(),
			ClinicID:   3,
			SourceType: model.LabImportSourceTypeDrWan,
			Status:     model.LabImportJobStatusReceived,
		}
		jobRepo.jobs[drwan.ID] = drwan
		_, err := svc.GetJob(context.Background(), 3, drwan.ID)
		if err == nil {
			t.Fatal("expected error opening drwan job")
		}
		if !apperrors.IsInvalidInput(err) {
			t.Errorf("expected InvalidInput, got %v", err)
		}
	})

	t.Run("returns not found for wrong clinic scope", func(t *testing.T) {
		_, err := svc.GetJob(context.Background(), 999, created.ID)
		if err == nil {
			t.Fatal("expected error for cross-clinic access, got nil")
		}
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected NotFound error, got: %v", err)
		}
	})
}

// ------------------------------------
// ListEvents
// ------------------------------------

func TestLabImportJobService_ListEvents(t *testing.T) {
	jobRepo := newStubJobRepo()
	eventRepo := &stubEventRepo{}
	svc := NewLabImportJobService(jobRepo, eventRepo)

	batch := model.LabInboundBatch{SourceType: model.LabImportSourceTypeFixture, ReceivedAt: time.Now()}
	created, err := svc.CreateJob(context.Background(), 4, batch)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	t.Run("returns events for existing job", func(t *testing.T) {
		events, err := svc.ListEvents(context.Background(), 4, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) == 0 {
			t.Error("expected at least the creation event")
		}
	})

	t.Run("returns not found when job does not exist", func(t *testing.T) {
		_, err := svc.ListEvents(context.Background(), 4, uuid.New())
		if err == nil {
			t.Fatal("expected error for unknown job id, got nil")
		}
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected NotFound error, got: %v", err)
		}
	})

	t.Run("propagates event repository error", func(t *testing.T) {
		eventRepo.findByJobErr = errors.New("db error")
		defer func() { eventRepo.findByJobErr = nil }()
		_, err := svc.ListEvents(context.Background(), 4, created.ID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
