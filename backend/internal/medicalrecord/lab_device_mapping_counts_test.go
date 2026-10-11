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
}

func (m *mappingCountsMasters) ResolveItemsForSpecies(_ context.Context, _ uint64, _, _ string, _ []string) (*LabDeviceMasterResolution, error) {
	return m.resolution, nil
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
	createCalls int
}

func (r *mappingCountsEvents) Create(_ context.Context, _ *model.LabImportEvent) error {
	r.createCalls++
	return nil
}

func newMappingCountsFixture() (*labDeviceExamPersister, *mappingCountsJobRepo, *mappingCountsJobService, *mappingCountsExams, *mappingCountsEvents, uuid.UUID, uint64) {
	clinicID, petID := uint64(910099), uint64(910098)
	jobID := uuid.New()
	items := make([]model.LabImportJobItem, 0, 16)
	unmapped := make([]string, 0, 16)
	for i := 0; i < 16; i++ {
		code := "C" + string(rune('A'+i))
		items = append(items, model.LabImportJobItem{ClinicID: clinicID, JobID: jobID, DeviceItemCode: code, SortOrder: i})
		unmapped = append(unmapped, code)
	}
	callLog := &[]string{}
	jobs := &mappingCountsJobRepo{
		job: &model.LabImportJob{
			ID:         jobID,
			ClinicID:   clinicID,
			SourceType: model.LabImportSourceTypeFujiNX600,
			Status:     model.LabImportJobStatusNeedsReview,
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
		masters: &mappingCountsMasters{resolution: &LabDeviceMasterResolution{UnmappedCodes: unmapped}},
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
