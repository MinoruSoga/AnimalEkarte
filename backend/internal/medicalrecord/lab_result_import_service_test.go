package medicalrecord

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ------------------------------------
// Stubs for LabResultImportService tests
// ------------------------------------

// stubLabJobService is an in-memory LabImportJobService for orchestration tests.
type stubLabJobService struct {
	jobs        map[uuid.UUID]*model.LabImportJob
	events      []*model.LabImportEvent
	nextID      int
	createErr   error
	transErr    error
	transErrFor model.LabImportJobStatus // only fail on this target status (zero = all)
}

func newStubLabJobService() *stubLabJobService {
	return &stubLabJobService{jobs: make(map[uuid.UUID]*model.LabImportJob)}
}

func (s *stubLabJobService) CreateJob(_ context.Context, clinicID uint64, batch model.LabInboundBatch) (*model.LabImportJob, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	s.nextID++
	id := uuid.New()
	now := time.Now()
	job := &model.LabImportJob{
		ID:                id,
		ClinicID:          clinicID,
		SourceType:        batch.SourceType,
		SourceFingerprint: batch.SourceFingerprint,
		Status:            model.LabImportJobStatusReceived,
		RowCount:          len(batch.ResultRows),
		StartedAt:         &now,
	}
	s.jobs[id] = job
	return job, nil
}

func (s *stubLabJobService) TransitionStatus(_ context.Context, clinicID uint64, jobID uuid.UUID, to model.LabImportJobStatus, counts TransitionCounts) (*model.LabImportJob, error) {
	if s.transErr != nil && (s.transErrFor == "" || s.transErrFor == to) {
		return nil, s.transErr
	}
	job, ok := s.jobs[jobID]
	if !ok || job.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("lab_import_job", jobID.String())
	}
	job.Status = to
	job.PersistedCount = counts.PersistedCount
	job.DuplicateCount = counts.DuplicateCount
	job.NeedsReviewCount = counts.NeedsReviewCount
	job.FailedCount = counts.FailedCount
	return job, nil
}

func (s *stubLabJobService) GetJob(_ context.Context, clinicID uint64, jobID uuid.UUID) (*model.LabImportJob, error) {
	job, ok := s.jobs[jobID]
	if !ok || job.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("lab_import_job", jobID.String())
	}
	return job, nil
}

func (s *stubLabJobService) ListJobs(_ context.Context, clinicID uint64, _ int) ([]*model.LabImportJob, error) {
	var out []*model.LabImportJob
	for _, j := range s.jobs {
		if j.ClinicID == clinicID {
			cp := *j
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *stubLabJobService) ListEvents(_ context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
	var out []*model.LabImportEvent
	for _, e := range s.events {
		if e.ClinicID == clinicID && e.JobID == jobID {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *stubLabJobService) PreviewBatch(_ context.Context, _ uint64, batch model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
	resp := &model.LabImportPreviewResponse{
		RowCount:        len(batch.ResultRows),
		MappingWarnings: []string{},
		BlockedReasons:  []string{},
	}
	if batch.SourceType == model.LabImportSourceTypeDrWan {
		resp.BlockedReasons = append(resp.BlockedReasons, "drwan blocked")
	}
	if batch.SourceType == model.LabImportSourceTypeManual {
		resp.BlockedReasons = append(resp.BlockedReasons, "manual blocked")
	}
	return resp, nil
}

// stubLabExamService is an in-memory LabImportExaminationService for orchestration tests.
type stubLabExamService struct {
	persistErr error
	results    []*LabExamPersistResult
	callCount  int
	persistFn  func(input LabExamPersistInput) (*LabExamPersistResult, error)
}

func (s *stubLabExamService) PersistExam(_ context.Context, input LabExamPersistInput) (*LabExamPersistResult, error) {
	s.callCount++
	if s.persistFn != nil {
		return s.persistFn(input)
	}
	if s.persistErr != nil {
		return nil, s.persistErr
	}
	res := &LabExamPersistResult{
		ExamID:    uint64(s.callCount),
		ItemCount: len(input.Items),
		Duplicate: false,
		JobID:     input.JobID,
	}
	s.results = append(s.results, res)
	return res, nil
}

func (s *stubLabExamService) PersistBatch(ctx context.Context, inputs []LabExamPersistInput) ([]*LabExamPersistResult, error) {
	var out []*LabExamPersistResult
	for _, inp := range inputs {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		res, err := s.PersistExam(ctx, inp)
		if err != nil {
			out = append(out, &LabExamPersistResult{RowError: err, JobID: inp.JobID})
			continue
		}
		out = append(out, res)
	}
	return out, nil
}

// syntheticFixtureBatch builds a minimal synthetic fixture batch for tests.
// Only uses synthetic analyte names and no real patient data.
func syntheticFixtureBatch(rowCount int) model.LabInboundBatch {
	rows := make([]model.LabInboundResultRow, rowCount)
	for i := range rows {
		rows[i] = model.LabInboundResultRow{
			OldPetKey:    "SYNTHETIC_PET_001",
			ExamCode:     "BUN",
			ExamName:     "血液化学検査(合成)",
			ItemName:     "BUN",
			DisplayValue: "12.3",
		}
	}
	return model.LabInboundBatch{
		SourceType:        model.LabImportSourceTypeFixture,
		SourceFingerprint: "sha256:synthetic-test-fixture",
		ReceivedAt:        time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		ResultRows:        rows,
	}
}

// ------------------------------------
// Preview tests
// ------------------------------------

// TestLabResultImportService_Preview_Fixture verifies preview returns row_count and no blocked_reasons
// for a fixture batch, and does NOT create a job or any exam/result rows.
func TestLabResultImportService_Preview_Fixture(t *testing.T) {
	jobSvc := newStubLabJobService()
	examSvc := &stubLabExamService{}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(3)
	resp, err := svc.Preview(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("Preview: unexpected error: %v", err)
	}
	if resp.RowCount != 3 {
		t.Errorf("expected row_count=3, got %d", resp.RowCount)
	}
	if len(resp.BlockedReasons) != 0 {
		t.Errorf("fixture should have no blocked_reasons, got %v", resp.BlockedReasons)
	}
	// preview must not create any jobs or exams
	if len(jobSvc.jobs) != 0 {
		t.Errorf("Preview must not create a job, got %d jobs", len(jobSvc.jobs))
	}
	if examSvc.callCount != 0 {
		t.Errorf("Preview must not call PersistExam, got %d calls", examSvc.callCount)
	}
}

// TestLabResultImportService_Preview_DrWan verifies drwan source is blocked.
func TestLabResultImportService_Preview_DrWan(t *testing.T) {
	svc := NewLabResultImportService(newStubLabJobService(), &stubLabExamService{})

	batch := model.LabInboundBatch{
		SourceType: model.LabImportSourceTypeDrWan,
		ReceivedAt: time.Now(),
	}
	resp, err := svc.Preview(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BlockedReasons) == 0 {
		t.Error("drwan source must produce at least one blocked_reason")
	}
}

// TestLabResultImportService_Preview_Manual verifies manual source is blocked.
func TestLabResultImportService_Preview_Manual(t *testing.T) {
	svc := NewLabResultImportService(newStubLabJobService(), &stubLabExamService{})

	batch := model.LabInboundBatch{
		SourceType: model.LabImportSourceTypeManual,
		ReceivedAt: time.Now(),
	}
	resp, err := svc.Preview(context.Background(), 1, batch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BlockedReasons) == 0 {
		t.Error("manual source must produce at least one blocked_reason")
	}
}

// TestLabResultImportService_Preview_NoRows verifies zero-row fixture preview.
func TestLabResultImportService_Preview_NoRows(t *testing.T) {
	svc := NewLabResultImportService(newStubLabJobService(), &stubLabExamService{})

	resp, err := svc.Preview(context.Background(), 1, syntheticFixtureBatch(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RowCount != 0 {
		t.Errorf("expected row_count=0, got %d", resp.RowCount)
	}
}

// ------------------------------------
// Commit tests
// ------------------------------------

// TestLabResultImportService_Commit_NonFixtureBlocked verifies drwan source cannot be committed.
func TestLabResultImportService_Commit_NonFixtureBlocked_DrWan(t *testing.T) {
	svc := NewLabResultImportService(newStubLabJobService(), &stubLabExamService{})

	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now(), JobID: uuid.New()}}
	batch := model.LabInboundBatch{
		SourceType: model.LabImportSourceTypeDrWan,
		ReceivedAt: time.Now(),
	}
	_, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err == nil {
		t.Fatal("expected error when committing drwan source in Phase 2")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput for blocked source, got: %v", err)
	}
}

// TestLabResultImportService_Commit_NonFixtureBlocked_Manual verifies manual source cannot be committed.
func TestLabResultImportService_Commit_NonFixtureBlocked_Manual(t *testing.T) {
	svc := NewLabResultImportService(newStubLabJobService(), &stubLabExamService{})

	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now(), JobID: uuid.New()}}
	batch := model.LabInboundBatch{
		SourceType: model.LabImportSourceTypeManual,
		ReceivedAt: time.Now(),
	}
	_, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err == nil {
		t.Fatal("expected error when committing manual source in Phase 2")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput for blocked source, got: %v", err)
	}
}

func TestLabResultImportService_Commit_DeviceSourcesBlocked(t *testing.T) {
	svc := NewLabResultImportService(newStubLabJobService(), &stubLabExamService{})
	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now(), JobID: uuid.New()}}
	for _, source := range []model.LabImportSourceType{
		model.LabImportSourceTypeFujiNX600,
		model.LabImportSourceTypeFujiAU10V,
		model.LabImportSourceTypeArkrayPU4010,
	} {
		_, err := svc.Commit(context.Background(), 1, model.LabInboundBatch{SourceType: source, ReceivedAt: time.Now()}, inputs)
		if err == nil || !apperrors.IsInvalidInput(err) {
			t.Fatalf("%s must not commit via fixture allowlist, err=%v", source, err)
		}
	}
}

// TestLabResultImportService_Commit_Fixture_Happy verifies a fixture commit creates a job,
// persists exams, records events, and returns a summary with persisted_count.
func TestLabResultImportService_Commit_Fixture_Happy(t *testing.T) {
	jobSvc := newStubLabJobService()
	examSvc := &stubLabExamService{}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(2)
	petA := uint64(10)
	petB := uint64(20)
	inputs := []LabExamPersistInput{
		{ClinicID: 1, PetID: &petA, ExamTypeID: 5, Date: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), JobID: uuid.Nil},
		{ClinicID: 1, PetID: &petB, ExamTypeID: 5, Date: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), JobID: uuid.Nil},
	}

	resp, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err != nil {
		t.Fatalf("Commit: unexpected error: %v", err)
	}
	if resp.JobID == uuid.Nil {
		t.Error("Commit must return a non-nil JobID")
	}
	if resp.PersistedCount != 2 {
		t.Errorf("expected persisted_count=2, got %d", resp.PersistedCount)
	}
	if resp.DuplicateCount != 0 {
		t.Errorf("expected duplicate_count=0, got %d", resp.DuplicateCount)
	}
	if resp.FailedCount != 0 {
		t.Errorf("expected failed_count=0, got %d", resp.FailedCount)
	}

	// job must exist and reach persisted terminal
	job, err := jobSvc.GetJob(context.Background(), 1, resp.JobID)
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if job.Status != model.LabImportJobStatusPersisted {
		t.Errorf("expected job status=persisted, got %s", job.Status)
	}

	// exam service was called for each input
	if examSvc.callCount != 2 {
		t.Errorf("expected PersistExam called 2 times, got %d", examSvc.callCount)
	}
}

// TestLabResultImportService_Commit_JobID_Injected verifies that Commit injects the created
// job's ID into each LabExamPersistInput before calling PersistExam, so results are linked.
func TestLabResultImportService_Commit_JobID_Injected(t *testing.T) {
	jobSvc := newStubLabJobService()

	var capturedJobID uuid.UUID
	examSvc := &stubLabExamService{
		persistFn: func(input LabExamPersistInput) (*LabExamPersistResult, error) {
			capturedJobID = input.JobID
			return &LabExamPersistResult{ExamID: 1, JobID: input.JobID}, nil
		},
	}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{
		{ClinicID: 1, ExamTypeID: 5, Date: time.Now()},
	}

	resp, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if capturedJobID == uuid.Nil {
		t.Error("JobID must be injected into PersistExam input")
	}
	if capturedJobID != resp.JobID {
		t.Errorf("captured JobID=%s does not match resp.JobID=%s", capturedJobID, resp.JobID)
	}
}

// TestLabResultImportService_Commit_WithDuplicate verifies duplicate rows are counted
// correctly in the commit response and job final state.
func TestLabResultImportService_Commit_WithDuplicate(t *testing.T) {
	jobSvc := newStubLabJobService()
	callCount := 0
	examSvc := &stubLabExamService{
		persistFn: func(input LabExamPersistInput) (*LabExamPersistResult, error) {
			callCount++
			if callCount == 2 {
				return &LabExamPersistResult{Duplicate: true, JobID: input.JobID}, nil
			}
			return &LabExamPersistResult{ExamID: uint64(callCount), JobID: input.JobID}, nil
		},
	}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(3)
	inputs := []LabExamPersistInput{
		{ClinicID: 1, ExamTypeID: 1, Date: time.Now()},
		{ClinicID: 1, ExamTypeID: 2, Date: time.Now()},
		{ClinicID: 1, ExamTypeID: 3, Date: time.Now()},
	}

	resp, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if resp.PersistedCount != 2 {
		t.Errorf("expected persisted_count=2, got %d", resp.PersistedCount)
	}
	if resp.DuplicateCount != 1 {
		t.Errorf("expected duplicate_count=1, got %d", resp.DuplicateCount)
	}

	// final job status: rows were persisted, so status should be persisted (not duplicate)
	job, _ := jobSvc.GetJob(context.Background(), 1, resp.JobID)
	if job.Status != model.LabImportJobStatusPersisted {
		t.Errorf("expected job status=persisted (some rows persisted), got %s", job.Status)
	}
}

// TestLabResultImportService_Commit_AllDuplicate verifies all-duplicate batches reach
// the duplicate terminal state.
func TestLabResultImportService_Commit_AllDuplicate(t *testing.T) {
	jobSvc := newStubLabJobService()
	examSvc := &stubLabExamService{
		persistFn: func(input LabExamPersistInput) (*LabExamPersistResult, error) {
			return &LabExamPersistResult{Duplicate: true, JobID: input.JobID}, nil
		},
	}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(2)
	inputs := []LabExamPersistInput{
		{ClinicID: 1, ExamTypeID: 1, Date: time.Now()},
		{ClinicID: 1, ExamTypeID: 2, Date: time.Now()},
	}

	resp, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if resp.DuplicateCount != 2 {
		t.Errorf("expected duplicate_count=2, got %d", resp.DuplicateCount)
	}
	if resp.PersistedCount != 0 {
		t.Errorf("expected persisted_count=0, got %d", resp.PersistedCount)
	}

	job, _ := jobSvc.GetJob(context.Background(), 1, resp.JobID)
	if job.Status != model.LabImportJobStatusDuplicate {
		t.Errorf("expected job status=duplicate when all rows are duplicate, got %s", job.Status)
	}
}

// TestLabResultImportService_Commit_RowError verifies that per-row errors increment failed_count
// and are counted separately from persistence failures.
func TestLabResultImportService_Commit_RowError(t *testing.T) {
	jobSvc := newStubLabJobService()
	callCount := 0
	examSvc := &stubLabExamService{
		persistFn: func(input LabExamPersistInput) (*LabExamPersistResult, error) {
			callCount++
			if callCount == 2 {
				return nil, errors.New("synthetic db error on row 2")
			}
			return &LabExamPersistResult{ExamID: uint64(callCount), JobID: input.JobID}, nil
		},
	}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(3)
	inputs := []LabExamPersistInput{
		{ClinicID: 1, ExamTypeID: 1, Date: time.Now()},
		{ClinicID: 1, ExamTypeID: 2, Date: time.Now()},
		{ClinicID: 1, ExamTypeID: 3, Date: time.Now()},
	}

	resp, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err != nil {
		t.Fatalf("Commit must not return function error for per-row failure: %v", err)
	}
	if resp.PersistedCount != 2 {
		t.Errorf("expected persisted_count=2 (rows 1 and 3), got %d", resp.PersistedCount)
	}
	if resp.FailedCount != 1 {
		t.Errorf("expected failed_count=1 (row 2 error), got %d", resp.FailedCount)
	}

	job, _ := jobSvc.GetJob(context.Background(), 1, resp.JobID)
	// some rows persisted, so persisted status
	if job.Status != model.LabImportJobStatusPersisted {
		t.Errorf("expected job status=persisted (some rows persisted), got %s", job.Status)
	}
}

// TestLabResultImportService_Commit_AllFailed verifies all-failed batches reach the failed terminal.
func TestLabResultImportService_Commit_AllFailed(t *testing.T) {
	jobSvc := newStubLabJobService()
	examSvc := &stubLabExamService{
		persistFn: func(input LabExamPersistInput) (*LabExamPersistResult, error) {
			return nil, errors.New("synthetic db error")
		},
	}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(2)
	inputs := []LabExamPersistInput{
		{ClinicID: 1, ExamTypeID: 1, Date: time.Now()},
		{ClinicID: 1, ExamTypeID: 2, Date: time.Now()},
	}

	resp, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err != nil {
		t.Fatalf("Commit must not return function error for per-row failures: %v", err)
	}
	if resp.FailedCount != 2 {
		t.Errorf("expected failed_count=2, got %d", resp.FailedCount)
	}
	if resp.PersistedCount != 0 {
		t.Errorf("expected persisted_count=0, got %d", resp.PersistedCount)
	}

	job, _ := jobSvc.GetJob(context.Background(), 1, resp.JobID)
	if job.Status != model.LabImportJobStatusFailed {
		t.Errorf("expected job status=failed when all rows fail, got %s", job.Status)
	}
}

// TestLabResultImportService_Commit_ClinicScopeEnforced verifies inputs with mismatched
// clinic_id (not matching the batch clinic) result in blocked behavior.
// Service enforces that all inputs must use the same clinicID as the batch.
func TestLabResultImportService_Commit_ClinicScopeEnforced(t *testing.T) {
	svc := NewLabResultImportService(newStubLabJobService(), &stubLabExamService{})

	batch := syntheticFixtureBatch(1)
	// input.ClinicID=999 does not match clinicID=1 passed to Commit
	inputs := []LabExamPersistInput{
		{ClinicID: 999, ExamTypeID: 1, Date: time.Now()},
	}

	_, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err == nil {
		t.Fatal("expected error when input.ClinicID does not match commit clinicID")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput for mismatched clinic scope, got: %v", err)
	}
}

// TestLabResultImportService_Commit_CreateJobError verifies that a job creation failure
// returns an error immediately without calling PersistExam.
func TestLabResultImportService_Commit_CreateJobError(t *testing.T) {
	jobSvc := newStubLabJobService()
	jobSvc.createErr = errors.New("db error creating job")
	examSvc := &stubLabExamService{}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now()}}

	_, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err == nil {
		t.Fatal("expected error when CreateJob fails")
	}
	if examSvc.callCount != 0 {
		t.Errorf("PersistExam must not be called when CreateJob fails, got %d calls", examSvc.callCount)
	}
}

// TestLabResultImportService_Commit_TransitionToValidatedError verifies that a failure
// transitioning the job to "validated" is wrapped and returned as an error (no PersistBatch call).
func TestLabResultImportService_Commit_TransitionToValidatedError(t *testing.T) {
	jobSvc := newStubLabJobService()
	jobSvc.transErr = errors.New("db error on validated transition")
	jobSvc.transErrFor = model.LabImportJobStatusValidated
	examSvc := &stubLabExamService{}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now()}}

	_, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err == nil {
		t.Fatal("expected error when transition to validated fails")
	}
	if examSvc.callCount != 0 {
		t.Errorf("PersistExam must not be called when validated transition fails, got %d calls", examSvc.callCount)
	}
}

// TestLabResultImportService_Commit_TransitionToMappedError verifies that a failure
// transitioning the job to "mapped" is wrapped and returned as an error (no PersistBatch call).
func TestLabResultImportService_Commit_TransitionToMappedError(t *testing.T) {
	jobSvc := newStubLabJobService()
	jobSvc.transErr = errors.New("db error on mapped transition")
	jobSvc.transErrFor = model.LabImportJobStatusMapped
	examSvc := &stubLabExamService{}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now()}}

	_, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err == nil {
		t.Fatal("expected error when transition to mapped fails")
	}
	if examSvc.callCount != 0 {
		t.Errorf("PersistExam must not be called when mapped transition fails, got %d calls", examSvc.callCount)
	}
}

// TestLabResultImportService_Commit_TerminalTransitionError verifies that a failure on the
// final terminal transition is returned as an error (TASK-032: do not report a successful
// persisted import when the job remains non-terminal; compensation gate depends on status).
// Exams may already be written; callers must treat the commit as failed and inspect job status.
func TestLabResultImportService_Commit_TerminalTransitionError(t *testing.T) {
	jobSvc := newStubLabJobService()
	jobSvc.transErr = errors.New("db error on terminal transition")
	jobSvc.transErrFor = model.LabImportJobStatusPersisted
	examSvc := &stubLabExamService{}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now()}}

	resp, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err == nil {
		t.Fatal("expected Commit to fail when terminal transition errors")
	}
	if resp != nil {
		t.Fatalf("expected nil response on terminal transition failure, got %+v", resp)
	}
	if examSvc.callCount != 1 {
		t.Errorf("expected PersistExam called once before terminal failure, got %d", examSvc.callCount)
	}
}

// TestLabResultImportService_Commit_ContextCancelledDuringPersist verifies that when
// PersistBatch is interrupted by a cancelled context, Commit transitions the job to
// "failed" via a compensating context (WithoutCancel) and returns the underlying error.
func TestLabResultImportService_Commit_ContextCancelledDuringPersist(t *testing.T) {
	jobSvc := newStubLabJobService()
	examSvc := &stubLabExamService{} // PersistBatch checks ctx.Err() before each row
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now()}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so the first PersistBatch loop iteration observes ctx.Err() != nil

	_, err := svc.Commit(ctx, 1, batch, inputs)
	if err == nil {
		t.Fatal("expected error when context is cancelled during PersistBatch")
	}

	// the job must have been created and transitioned to failed via the compensating context
	var job *model.LabImportJob
	for _, j := range jobSvc.jobs {
		job = j
	}
	if job == nil {
		t.Fatal("expected job to have been created before cancellation was observed")
		return
	}
	if job.Status != model.LabImportJobStatusFailed {
		t.Errorf("expected job status=failed after cancellation cleanup, got %s", job.Status)
	}
}

// TestLabResultImportService_Commit_ContextCancelledDuringPersist_CompensationTransitionAlsoFails
// verifies that when PersistBatch is interrupted by a cancelled context AND the compensating
// "failed" transition itself also fails (e.g. DB outage), Commit still returns the original
// primary error unchanged (behavior-preserving: the compensation-failure is only logged, never
// allowed to shadow or replace the primary error).
func TestLabResultImportService_Commit_ContextCancelledDuringPersist_CompensationTransitionAlsoFails(t *testing.T) {
	jobSvc := newStubLabJobService()
	jobSvc.transErr = errors.New("db outage on compensation transition")
	jobSvc.transErrFor = model.LabImportJobStatusFailed
	examSvc := &stubLabExamService{} // PersistBatch checks ctx.Err() before each row
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(1)
	inputs := []LabExamPersistInput{{ClinicID: 1, ExamTypeID: 1, Date: time.Now()}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so the first PersistBatch loop iteration observes ctx.Err() != nil

	_, err := svc.Commit(ctx, 1, batch, inputs)
	if err == nil {
		t.Fatal("expected error when context is cancelled during PersistBatch")
	}
	if !strings.Contains(err.Error(), "lab import batch interrupted") {
		t.Errorf("expected primary error wrapped as 'lab import batch interrupted', got: %v", err)
	}
	if errors.Is(err, jobSvc.transErr) {
		t.Errorf("returned error must not be (or wrap) the compensation-transition error, got: %v", err)
	}

	// the job must have been created but stays stuck in a non-terminal state (e.g. "mapped")
	// because the compensating transition to "failed" itself failed.
	var job *model.LabImportJob
	for _, j := range jobSvc.jobs {
		job = j
	}
	if job == nil {
		t.Fatal("expected job to have been created before cancellation was observed")
		return
	}
	if job.Status == model.LabImportJobStatusFailed {
		t.Errorf("expected job to remain stuck in a non-terminal status when compensation transition fails, got %s", job.Status)
	}
}

// TestLabResultImportService_Commit_EmptyInputs verifies an empty inputs slice
// creates a job, persists no exams, and reaches the duplicate terminal state
// (empty = nothing written = idempotent/duplicate semantics).
func TestLabResultImportService_Commit_EmptyInputs(t *testing.T) {
	jobSvc := newStubLabJobService()
	examSvc := &stubLabExamService{}
	svc := NewLabResultImportService(jobSvc, examSvc)

	batch := syntheticFixtureBatch(0)
	resp, err := svc.Commit(context.Background(), 1, batch, []LabExamPersistInput{})
	if err != nil {
		t.Fatalf("Commit with empty inputs: %v", err)
	}
	if resp.PersistedCount != 0 {
		t.Errorf("expected persisted_count=0, got %d", resp.PersistedCount)
	}
	if examSvc.callCount != 0 {
		t.Errorf("expected 0 PersistExam calls for empty inputs, got %d", examSvc.callCount)
	}
	// empty commit reaches duplicate terminal (nothing written = idempotent)
	job, _ := jobSvc.GetJob(context.Background(), 1, resp.JobID)
	if job == nil {
		t.Fatal("expected job to be created even for empty inputs")
		return
	}
	if job.Status != model.LabImportJobStatusDuplicate {
		t.Errorf("expected job status=duplicate for empty inputs, got %s", job.Status)
	}
}

// TestLabResultImportService_NoExternalIO documents that LabResultImportService
// has no external network or credential dependencies (compile-time structural guarantee).
func TestLabResultImportService_NoExternalIO(t *testing.T) {
	// LabResultImportService depends only on LabImportJobService and
	// LabImportExaminationService interfaces. No network, no credentials, no Dr.Wan access.
	// Phase BLOCKED: Dr.Wan MDB connection requires external schema confirmation.
	t.Log("no external IO: LabResultImportService depends only on injected service interfaces")
}

// TestLabResultImportService_Phase3A_ServiceLevelDuplicatePolicy は
// Phase 3A 決定: サービスレベル重複防止が正式方針であることを記録するコントラクトテスト。
//
// 決定根拠 (2026-06-26, Issue #249 R-3 で更新):
//   - DB unique 制約は同日再検査の正当な複数行を拒絶するため追加しない。
//   - 重複スキップは完全同一ペイロード（header + items）の再インポートのみ（PO ruling）。
//   - 同日・同検査種別で内容が異なれば新規 exam として保存する。
//   - TOCTOU リスク（IsDuplicate と Create の間の競合）は設計上許容する。
//
// このテストはサービスレベル重複チェック結果が Commit 集計へ正しく反映されることを verify する
// （stub が full-identical 相当の 2 行目 skip を返す）。
func TestLabResultImportService_Phase3A_ServiceLevelDuplicatePolicy(t *testing.T) {
	// サービスレベル重複チェック: 完全同一再インポートは DB 制約ではなく
	// LabImportDuplicateChecker によって検知・スキップされる。
	jobSvc := newStubLabJobService()
	callCount := 0
	examSvc := &stubLabExamService{
		persistFn: func(input LabExamPersistInput) (*LabExamPersistResult, error) {
			callCount++
			// 1 行目: 新規 → persisted
			// 2 行目: 完全同一ペイロード → サービスレベルで duplicate 検知
			if callCount == 2 {
				return &LabExamPersistResult{Duplicate: true, JobID: input.JobID}, nil
			}
			return &LabExamPersistResult{ExamID: uint64(callCount), JobID: input.JobID}, nil
		},
	}
	svc := NewLabResultImportService(jobSvc, examSvc)

	petID := uint64(42)
	examDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	batch := syntheticFixtureBatch(2)
	// orchestration 契約: 下位サービスが Duplicate=true を返した行は duplicate_count に集計される
	items := []LabExamItemInput{{Name: "BUN", InspectionValue: "12.0", SortOrder: 1}}
	inputs := []LabExamPersistInput{
		{ClinicID: 1, PetID: &petID, ExamTypeID: 5, Date: examDate, Machine: "Fuji", Items: items},
		{ClinicID: 1, PetID: &petID, ExamTypeID: 5, Date: examDate, Machine: "Fuji", Items: items}, // full-identical
	}

	resp, err := svc.Commit(context.Background(), 1, batch, inputs)
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	// 1 row persisted, 1 row blocked by service-level duplicate checker
	if resp.PersistedCount != 1 {
		t.Errorf("Phase3A: expected persisted_count=1, got %d", resp.PersistedCount)
	}
	if resp.DuplicateCount != 1 {
		t.Errorf("Phase3A: expected duplicate_count=1 (service-level), got %d", resp.DuplicateCount)
	}
	if resp.FailedCount != 0 {
		t.Errorf("Phase3A: expected failed_count=0, got %d", resp.FailedCount)
	}

	// job reaches persisted (some rows persisted)
	job, err := jobSvc.GetJob(context.Background(), 1, resp.JobID)
	if err != nil {
		t.Fatalf("Phase3A: GetJob: %v", err)
	}
	if job.Status != model.LabImportJobStatusPersisted {
		t.Errorf("Phase3A: expected job status=persisted, got %s", job.Status)
	}

	t.Log("Phase3A: service-level full-identical duplicate checker is the formal policy; no DB unique constraint")
}
