package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

type mappingCountsCall struct {
	clinicID uint64
	jobID    uuid.UUID
	count    int
}

type mappingCountsJobRepo struct {
	LabDeviceReceiveRepository
	job         *model.LabImportJob
	items       map[uuid.UUID][]model.LabImportJobItem
	saveErr     error
	countsErr   error
	callLog     *[]string
	savedItems  []model.LabImportJobItem
	countsCalls []mappingCountsCall
}

func (r *mappingCountsJobRepo) FindJobByID(_ context.Context, clinicID uint64, jobID uuid.UUID) (*model.LabImportJob, error) {
	if r.job == nil || r.job.ClinicID != clinicID || r.job.ID != jobID {
		return nil, errors.New("job not found")
	}
	return r.job, nil
}

func (r *mappingCountsJobRepo) ListJobItems(_ context.Context, _ uint64, jobIDs []uuid.UUID) (map[uuid.UUID][]model.LabImportJobItem, error) {
	out := make(map[uuid.UUID][]model.LabImportJobItem, len(jobIDs))
	for _, id := range jobIDs {
		out[id] = r.items[id]
	}
	return out, nil
}

func (r *mappingCountsJobRepo) SaveJobItems(_ context.Context, items []model.LabImportJobItem) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	*r.callLog = append(*r.callLog, "save_items")
	r.savedItems = append([]model.LabImportJobItem(nil), items...)
	return nil
}

func (r *mappingCountsJobRepo) UpdateJobMappingCounts(_ context.Context, clinicID uint64, jobID uuid.UUID, unmappedCount int) error {
	if r.countsErr != nil {
		return r.countsErr
	}
	*r.callLog = append(*r.callLog, "update_counts")
	r.countsCalls = append(r.countsCalls, mappingCountsCall{clinicID: clinicID, jobID: jobID, count: unmappedCount})
	return nil
}

type mappingCountsMasters struct {
	LabDeviceItemMasterService
	resolution *LabDeviceMasterResolution
	resolveErr error
}

func (m *mappingCountsMasters) ResolveItemsForSpecies(_ context.Context, _ uint64, _, _ string, _ []string) (*LabDeviceMasterResolution, error) {
	return m.resolution, m.resolveErr
}

type mappingCountsJobService struct {
	LabImportJobService
	transitionCalls int
}

func (s *mappingCountsJobService) TransitionStatus(_ context.Context, _ uint64, _ uuid.UUID, _ model.LabImportJobStatus, _ TransitionCounts) (*model.LabImportJob, error) {
	s.transitionCalls++
	return nil, nil
}

type mappingCountsExams struct {
	LabImportExaminationService
	persistCalls int
}

func (s *mappingCountsExams) PersistExam(_ context.Context, _ LabExamPersistInput) (*LabExamPersistResult, error) {
	s.persistCalls++
	return nil, nil
}

type mappingCountsEvents struct {
	LabImportEventRepository
	hasEventType bool
	createCalls  int
}

func (r *mappingCountsEvents) Create(_ context.Context, _ *model.LabImportEvent) error {
	r.createCalls++
	return nil
}

func (r *mappingCountsEvents) HasEventType(_ context.Context, _ uint64, _ uuid.UUID, _ model.LabImportEventType) (bool, error) {
	return r.hasEventType, nil
}

func mappingCountsItems(clinicID uint64, jobID uuid.UUID, count int) ([]model.LabImportJobItem, []string) {
	items := make([]model.LabImportJobItem, 0, count)
	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code := "C" + string(rune('A'+i))
		items = append(items, model.LabImportJobItem{ClinicID: clinicID, JobID: jobID, DeviceItemCode: code, SortOrder: i})
		codes = append(codes, code)
	}
	return items, codes
}

func mappingCountsResolution(codes []string, mappedCount int) *LabDeviceMasterResolution {
	resolution := &LabDeviceMasterResolution{}
	for i, code := range codes {
		if i < mappedCount {
			resolution.Mapped = append(resolution.Mapped, LabDeviceResolvedItem{
				DeviceItemCode:  code,
				ValueShape:      "numeric",
				Unit:            "mg/dL",
				FieldName:       "F" + string(rune('A'+i)),
				ExamTypeFieldID: uint64(100 + i),
				ExamTypeID:      42,
			})
			continue
		}
		resolution.UnmappedCodes = append(resolution.UnmappedCodes, code)
	}
	return resolution
}

func newMappingCountsFixture() (*labDeviceExamPersister, *mappingCountsJobRepo, *mappingCountsJobService, *mappingCountsExams, *mappingCountsEvents, uuid.UUID, uint64) {
	_, codes := mappingCountsItems(910099, uuid.Nil, 16)
	return newMappingCountsFixtureWithResolution(
		&LabDeviceMasterResolution{UnmappedCodes: codes},
		model.LabImportJobStatusNeedsReview,
	)
}

func newMappingCountsFixtureWithResolution(
	resolution *LabDeviceMasterResolution,
	status model.LabImportJobStatus,
) (*labDeviceExamPersister, *mappingCountsJobRepo, *mappingCountsJobService, *mappingCountsExams, *mappingCountsEvents, uuid.UUID, uint64) {
	clinicID, petID := uint64(910099), uint64(910098)
	jobID := uuid.New()
	items, _ := mappingCountsItems(clinicID, jobID, 16)
	callLog := &[]string{}
	jobs := &mappingCountsJobRepo{
		job: &model.LabImportJob{
			ID:         jobID,
			ClinicID:   clinicID,
			SourceType: model.LabImportSourceTypeFujiNX600,
			Status:     status,
			PetID:      &petID,
			RowCount:   16,
		},
		items:   map[uuid.UUID][]model.LabImportJobItem{jobID: items},
		callLog: callLog,
	}
	jobSvc := &mappingCountsJobService{}
	exams := &mappingCountsExams{}
	events := &mappingCountsEvents{}
	s := &labDeviceExamPersister{
		jobs:    jobs,
		masters: &mappingCountsMasters{resolution: resolution},
		exams:   exams,
		jobSvc:  jobSvc,
		events:  events,
		now:     time.Now,
	}
	return s, jobs, jobSvc, exams, events, jobID, petID
}

func TestLabDeviceJobMappingCounts_StoredWithoutStatusTransition(t *testing.T) {
	s, jobs, jobSvc, exams, _, jobID, petID := newMappingCountsFixture()
	if err := s.persistLinkedJob(context.Background(), 910099, jobID, petID); err != nil {
		t.Fatalf("persistLinkedJob failed: %v", err)
	}
	if len(jobs.countsCalls) != 1 {
		t.Fatalf("UpdateJobMappingCalls=%d, want 1", len(jobs.countsCalls))
	}
	call := jobs.countsCalls[0]
	if call.clinicID != 910099 || call.jobID != jobID || call.count != 16 {
		t.Fatalf("counter write args=(%d,%s,%d), want (910099,<jobID>,16)", call.clinicID, call.jobID, call.count)
	}
	if len(jobs.savedItems) != 16 {
		t.Fatalf("saved items=%d, want 16", len(jobs.savedItems))
	}
	log := *jobs.callLog
	if len(log) != 2 || log[0] != "save_items" || log[1] != "update_counts" {
		t.Fatalf("call order=%v, want [save_items update_counts]", log)
	}
	if jobSvc.transitionCalls != 0 || exams.persistCalls != 0 {
		t.Fatalf("unexpected transitions=%d exams=%d on needs_review path", jobSvc.transitionCalls, exams.persistCalls)
	}
}

func TestLabDeviceJobMappingCounts_ErrorPropagatesBeforeExamAndEvent(t *testing.T) {
	s, jobs, jobSvc, exams, events, jobID, petID := newMappingCountsFixture()
	sentinel := errors.New("counter write failed")
	jobs.countsErr = sentinel
	err := s.persistLinkedJob(context.Background(), 910099, jobID, petID)
	if !errors.Is(err, sentinel) {
		t.Fatalf("persistLinkedJob err=%v, want sentinel", err)
	}
	if len(jobs.countsCalls) != 0 || exams.persistCalls != 0 || events.createCalls != 0 || jobSvc.transitionCalls != 0 {
		t.Fatalf("downstream calls after setter failure: counts=%d exams=%d events=%d transitions=%d",
			len(jobs.countsCalls), exams.persistCalls, events.createCalls, jobSvc.transitionCalls)
	}
}

func TestLabDeviceJobMappingCounts_ItemSaveFailureNeverCallsCounter(t *testing.T) {
	s, jobs, _, exams, events, jobID, petID := newMappingCountsFixture()
	sentinel := errors.New("item save failed")
	jobs.saveErr = sentinel
	err := s.persistLinkedJob(context.Background(), 910099, jobID, petID)
	if !errors.Is(err, sentinel) {
		t.Fatalf("persistLinkedJob err=%v, want sentinel", err)
	}
	if len(jobs.countsCalls) != 0 || exams.persistCalls != 0 || events.createCalls != 0 {
		t.Fatalf("counter/exam/event called after item save failure: counts=%d exams=%d events=%d",
			len(jobs.countsCalls), exams.persistCalls, events.createCalls)
	}
}

func TestLabDeviceJobMappingCounts_RecalculatedPartialAndZero(t *testing.T) {
	cases := []struct {
		name      string
		mapped    int
		wantCount int
	}{
		{name: "one mapped leaves fifteen unmapped", mapped: 1, wantCount: 15},
		{name: "all mapped leaves zero unmapped", mapped: 16, wantCount: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, codes := mappingCountsItems(910099, uuid.Nil, 16)
			s, jobs, jobSvc, exams, events, jobID, petID := newMappingCountsFixtureWithResolution(
				mappingCountsResolution(codes, tc.mapped),
				model.LabImportJobStatusNeedsReview,
			)
			if err := s.persistLinkedJob(context.Background(), 910099, jobID, petID); err != nil {
				t.Fatalf("persistLinkedJob failed: %v", err)
			}
			if len(jobs.countsCalls) != 1 {
				t.Fatalf("UpdateJobMappingCounts calls=%d, want 1", len(jobs.countsCalls))
			}
			call := jobs.countsCalls[0]
			if call.clinicID != 910099 || call.jobID != jobID || call.count != tc.wantCount {
				t.Fatalf("counter write args=(%d,%s,%d), want (910099,<jobID>,%d)",
					call.clinicID, call.jobID, call.count, tc.wantCount)
			}
			if len(jobs.savedItems) != 16 {
				t.Fatalf("saved items=%d, want 16", len(jobs.savedItems))
			}
			mappedSet := make(map[string]struct{}, tc.mapped)
			for i := 0; i < tc.mapped; i++ {
				mappedSet[codes[i]] = struct{}{}
			}
			for _, item := range jobs.savedItems {
				if _, ok := mappedSet[item.DeviceItemCode]; ok {
					if item.ExamTypeFieldID == nil || *item.ExamTypeFieldID == 0 || item.NeedsReview {
						t.Fatalf("mapped item %s flags=(field=%v,review=%v), want resolved field and review=false",
							item.DeviceItemCode, item.ExamTypeFieldID, item.NeedsReview)
					}
					continue
				}
				if item.ExamTypeFieldID != nil || !item.NeedsReview {
					t.Fatalf("unmapped item %s flags=(field=%v,review=%v), want field=nil and review=true",
						item.DeviceItemCode, item.ExamTypeFieldID, item.NeedsReview)
				}
			}
			log := *jobs.callLog
			if len(log) != 2 || log[0] != "save_items" || log[1] != "update_counts" {
				t.Fatalf("call order=%v, want [save_items update_counts]", log)
			}
			if exams.persistCalls != 1 {
				t.Fatalf("exam persists=%d, want 1 (single exam type)", exams.persistCalls)
			}
			if jobSvc.transitionCalls != 3 {
				t.Fatalf("status transitions=%d, want 3 (validated->mapped->persisted)", jobSvc.transitionCalls)
			}
			if events.createCalls != 1 {
				t.Fatalf("usage tracking events=%d, want 1", events.createCalls)
			}
		})
	}
}

func TestLabDeviceJobMappingCounts_PersistedRetryDoesNotWrite(t *testing.T) {
	s, jobs, jobSvc, exams, events, jobID, petID := newMappingCountsFixtureWithResolution(
		nil, model.LabImportJobStatusPersisted,
	)
	s.masters = &mappingCountsMasters{resolveErr: errors.New("resolver must not be reached")}
	jobs.saveErr = errors.New("item save must not be reached")
	jobs.countsErr = errors.New("counter write must not be reached")
	if err := s.persistLinkedJob(context.Background(), 910099, jobID, petID); err != nil {
		t.Fatalf("persisted job retry should early-return nil, got %v", err)
	}
	if len(jobs.countsCalls) != 0 || len(jobs.savedItems) != 0 {
		t.Fatalf("writes on persisted retry: counts=%d items=%d, want 0/0",
			len(jobs.countsCalls), len(jobs.savedItems))
	}
	if jobSvc.transitionCalls != 0 || exams.persistCalls != 0 || events.createCalls != 0 {
		t.Fatalf("downstream calls on persisted retry: transitions=%d exams=%d events=%d, want 0/0/0",
			jobSvc.transitionCalls, exams.persistCalls, events.createCalls)
	}
}
