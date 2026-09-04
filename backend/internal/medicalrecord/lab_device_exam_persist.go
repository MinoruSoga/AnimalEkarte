package medicalrecord

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type labDeviceExamPersister struct {
	jobs    LabDeviceReceiveRepository
	masters LabDeviceItemMasterService
	exams   LabImportExaminationService
	jobSvc  LabImportJobService
	events  LabImportEventRepository
	revert  LabImportRevertService
	tx      Transactor
	now     func() time.Time
}

// NewLabDeviceExamPersister writes one exam per mapped exam type for each linked measurement.
func NewLabDeviceExamPersister(
	jobs LabDeviceReceiveRepository,
	masters LabDeviceItemMasterService,
	exams LabImportExaminationService,
	jobSvc LabImportJobService,
	events LabImportEventRepository,
	revert LabImportRevertService,
	tx ...Transactor,
) LabDeviceExamPersister {
	return &labDeviceExamPersister{
		jobs:    jobs,
		masters: masters,
		exams:   exams,
		jobSvc:  jobSvc,
		events:  events,
		revert:  revert,
		tx:      firstTransactor(tx),
		now:     time.Now,
	}
}

func firstTransactor(values []Transactor) Transactor {
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func (s *labDeviceExamPersister) PersistLinkedJob(ctx context.Context, clinicID uint64, jobID uuid.UUID, petID uint64) error {
	if persistence.TxFromContext(ctx) != nil {
		return s.persistLinkedJob(ctx, clinicID, jobID, petID)
	}
	if s.tx == nil {
		return apperrors.WrapInternalServerError("lab device persist transaction dependency is required")
	}
	return s.tx.WithTx(ctx, func(txCtx context.Context) error {
		return s.persistLinkedJob(txCtx, clinicID, jobID, petID)
	})
}

func (s *labDeviceExamPersister) persistLinkedJob(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	petID uint64,
) error {
	job, err := s.jobs.FindJobByID(ctx, clinicID, jobID)
	if err != nil {
		return err
	}
	if !isLabDeviceSourceType(string(job.SourceType)) {
		return apperrors.WrapInvalidInput("job is not a lab device source")
	}
	if job.Status == model.LabImportJobStatusPersisted {
		return nil
	}
	if job.PetID == nil || *job.PetID != petID {
		return apperrors.WrapInvalidInput("pet_id is required")
	}

	itemMap, err := s.jobs.ListJobItems(ctx, clinicID, []uuid.UUID{jobID})
	if err != nil {
		return err
	}
	items := itemMap[jobID]
	codes := make([]string, 0, len(items))
	for i := range items {
		codes = append(codes, items[i].DeviceItemCode)
	}
	resolution, err := s.masters.ResolveItems(ctx, clinicID, string(job.SourceType), codes)
	if err != nil {
		return err
	}

	mappedField := applyLabJobItemResolution(items, *resolution)
	if err := s.jobs.SaveJobItems(ctx, items); err != nil {
		return err
	}
	job.UnmappedItemCount = len(resolution.UnmappedCodes)
	job.NeedsReviewCount = len(resolution.UnmappedCodes)
	counts := TransitionCounts{
		RowCount:         job.RowCount,
		NeedsReviewCount: job.NeedsReviewCount,
	}

	// T001: 複数 exam_type が混在する場合（VetLab 送信口など）も保存拒否しない。
	// 種別ごとに 1 exam を作る。1種別なら従来どおり 1 件。
	if len(resolution.Mapped) == 0 {
		return s.markNeedsReview(ctx, clinicID, jobID, job.Status, counts)
	}

	examTypeIDs := UniqueMappedExamTypeIDs(resolution.Mapped)
	date := s.now()
	if job.MeasuredAt != nil {
		date = *job.MeasuredAt
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	if s.exams == nil {
		return apperrors.WrapInternalServerError("lab device exam persister is not configured")
	}
	totalPersisted, totalDuplicate, persistErr := s.persistMappedLabExams(
		ctx, clinicID, petID, jobID, date, job.DeviceHint, examTypeIDs, items, mappedField,
	)
	if persistErr != nil {
		return persistErr
	}
	counts.PersistedCount = totalPersisted
	counts.DuplicateCount = totalDuplicate
	if err := s.advanceToPersisted(ctx, clinicID, jobID, job.Status, counts); err != nil {
		return err
	}
	return s.markUsageTracking(ctx, clinicID, jobID, counts.PersistedCount)
}

func (s *labDeviceExamPersister) RetractLinkedJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) error {
	if s.revert == nil {
		return apperrors.WrapConflict("保存済みの検査は取り消せません。検査記録の訂正が必要です")
	}
	return s.revert.DetachDeviceJob(ctx, clinicID, jobID)
}

func (s *labDeviceExamPersister) markNeedsReview(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	from model.LabImportJobStatus,
	counts TransitionCounts,
) error {
	steps := []model.LabImportJobStatus{model.LabImportJobStatusValidated, model.LabImportJobStatusNeedsReview}
	if from == model.LabImportJobStatusValidated {
		steps = []model.LabImportJobStatus{model.LabImportJobStatusNeedsReview}
	}
	if from == model.LabImportJobStatusNeedsReview {
		return nil
	}
	return s.walkStatus(ctx, clinicID, jobID, steps, counts)
}

func (s *labDeviceExamPersister) advanceToPersisted(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	from model.LabImportJobStatus,
	counts TransitionCounts,
) error {
	var steps []model.LabImportJobStatus
	switch from {
	case model.LabImportJobStatusReceived, model.LabImportJobStatusNeedsReview:
		steps = []model.LabImportJobStatus{
			model.LabImportJobStatusValidated,
			model.LabImportJobStatusMapped,
			model.LabImportJobStatusPersisted,
		}
	case model.LabImportJobStatusValidated:
		steps = []model.LabImportJobStatus{model.LabImportJobStatusMapped, model.LabImportJobStatusPersisted}
	case model.LabImportJobStatusMapped:
		steps = []model.LabImportJobStatus{model.LabImportJobStatusPersisted}
	case model.LabImportJobStatusPersisted:
		return nil
	default:
		return apperrors.WrapInvalidInput("job status cannot persist")
	}
	return s.walkStatus(ctx, clinicID, jobID, steps, counts)
}

func (s *labDeviceExamPersister) walkStatus(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	steps []model.LabImportJobStatus,
	counts TransitionCounts,
) error {
	if s.jobSvc == nil {
		return apperrors.WrapInternalServerError("lab import job service is not configured")
	}
	for _, to := range steps {
		if _, err := s.jobSvc.TransitionStatus(ctx, clinicID, jobID, to, counts); err != nil {
			return err
		}
	}
	return nil
}

func (s *labDeviceExamPersister) markUsageTracking(ctx context.Context, clinicID uint64, jobID uuid.UUID, persisted int) error {
	if s.events == nil {
		return apperrors.WrapInternalServerError("lab import event repository is not configured")
	}
	has, err := s.events.HasEventType(ctx, clinicID, jobID, model.LabImportEventTypeUsageTrackingStarted)
	if err != nil || has {
		return err
	}
	return s.events.Create(ctx, &model.LabImportEvent{
		ClinicID:  clinicID,
		JobID:     jobID,
		EventType: model.LabImportEventTypeUsageTrackingStarted,
		RowCount:  persisted,
	})
}

func applyLabJobItemResolution(
	items []model.LabImportJobItem,
	resolution LabDeviceMasterResolution,
) map[string]LabDeviceResolvedItem {
	mappedField := make(map[string]LabDeviceResolvedItem, len(resolution.Mapped))
	for _, mapped := range resolution.Mapped {
		mappedField[mapped.DeviceItemCode] = mapped
	}
	unmapped := make(map[string]struct{}, len(resolution.UnmappedCodes))
	for _, code := range resolution.UnmappedCodes {
		unmapped[code] = struct{}{}
	}
	for i := range items {
		if mapped, ok := mappedField[items[i].DeviceItemCode]; ok {
			fieldID := mapped.ExamTypeFieldID
			items[i].ExamTypeFieldID = &fieldID
			items[i].NeedsReview = false
			continue
		}
		items[i].ExamTypeFieldID = nil
		if _, ok := unmapped[items[i].DeviceItemCode]; ok {
			items[i].NeedsReview = true
		}
	}
	return mappedField
}

func (s *labDeviceExamPersister) persistMappedLabExams(
	ctx context.Context,
	clinicID, petID uint64,
	jobID uuid.UUID,
	date time.Time,
	deviceHint string,
	examTypeIDs []uint64,
	items []model.LabImportJobItem,
	mappedField map[string]LabDeviceResolvedItem,
) (totalPersisted, totalDuplicate int, err error) {
	itemsByExamType := make(map[uint64][]LabExamItemInput, len(examTypeIDs))
	for i := range items {
		mapped, ok := mappedField[items[i].DeviceItemCode]
		if !ok {
			continue
		}
		name := mapped.FieldName
		if name == "" {
			name = items[i].DeviceItemCode
		}
		unit := items[i].Unit
		if unit == "" {
			unit = mapped.Unit
		}
		fieldID := mapped.ExamTypeFieldID
		itemsByExamType[mapped.ExamTypeID] = append(itemsByExamType[mapped.ExamTypeID], LabExamItemInput{
			Name:            name,
			InspectionValue: items[i].ValueRaw,
			Unit:            unit,
			ExamTypeFieldID: &fieldID,
			SortOrder:       items[i].SortOrder,
		})
	}
	for _, examTypeID := range examTypeIDs {
		result, persistErr := s.exams.PersistExam(ctx, LabExamPersistInput{
			ClinicID:   clinicID,
			PetID:      &petID,
			ExamTypeID: examTypeID,
			Date:       date,
			Machine:    deviceHint,
			JobID:      jobID,
			Items:      itemsByExamType[examTypeID],
		})
		if persistErr != nil {
			return 0, 0, persistErr
		}
		if result != nil && result.RowError != nil {
			return 0, 0, result.RowError
		}
		if result != nil && result.Duplicate {
			totalDuplicate++
		} else {
			totalPersisted++
		}
	}
	return totalPersisted, totalDuplicate, nil
}
