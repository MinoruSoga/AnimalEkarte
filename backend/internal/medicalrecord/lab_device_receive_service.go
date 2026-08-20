package medicalrecord

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

type labDeviceReceiveService struct {
	repo      LabDeviceReceiveRepository
	masters   LabDeviceItemMasterService
	pets      LabDevicePetFinder
	tx        Transactor
	persister LabDeviceExamPersister
	visits    LabDeviceTodayVisitFinder
	now       func() time.Time
}

// NewLabDeviceReceiveService initializes a LabDeviceReceiveService.
func NewLabDeviceReceiveService(
	repo LabDeviceReceiveRepository,
	masters LabDeviceItemMasterService,
	pets LabDevicePetFinder,
	tx Transactor,
	persister LabDeviceExamPersister,
	visits LabDeviceTodayVisitFinder,
) LabDeviceReceiveService {
	return &labDeviceReceiveService{
		repo:      repo,
		masters:   masters,
		pets:      pets,
		tx:        tx,
		persister: persister,
		visits:    visits,
		now:       time.Now,
	}
}

func (s *labDeviceReceiveService) withTx(ctx context.Context, fn func(context.Context) error) error {
	if s.tx == nil {
		return fn(ctx)
	}
	return s.tx.WithTx(ctx, fn)
}

func (s *labDeviceReceiveService) ReceiveFrames(
	ctx context.Context,
	clinicID uint64,
	payload []byte,
	hint string,
) (*LabDeviceReceiveResult, error) {
	frames, err := DecodeLabDeviceFrames(payload, hint)
	if err != nil {
		if errors.Is(err, ErrInvalidLabDevicePayload) {
			return nil, apperrors.WrapInvalidInput("invalid_payload")
		}
		return nil, err
	}
	out := &LabDeviceReceiveResult{Results: make([]LabDeviceFrameResult, 0, len(frames))}
	for i := range frames {
		result, recvErr := s.receiveOne(ctx, clinicID, frames[i])
		if recvErr != nil {
			return nil, recvErr
		}
		out.Results = append(out.Results, *result)
	}
	return out, nil
}

func (s *labDeviceReceiveService) receiveOne(
	ctx context.Context,
	clinicID uint64,
	frame LabDeviceFrame,
) (*LabDeviceFrameResult, error) {
	var result LabDeviceFrameResult
	var persistPetID *uint64
	var persistJobID uuid.UUID
	err := s.withTx(ctx, func(txCtx context.Context) error {
		existing, findErr := s.repo.FindJobByFingerprint(txCtx, clinicID, string(frame.SourceType), frame.SourceFingerprint)
		if findErr != nil {
			return findErr
		}
		if existing != nil {
			card, cardErr := s.cardForJob(txCtx, clinicID, existing)
			if cardErr != nil {
				return cardErr
			}
			result = LabDeviceFrameResult{Duplicate: true, Job: *card}
			return nil
		}

		petID, waitErr := s.consumeWaitPet(txCtx, clinicID)
		if waitErr != nil {
			return waitErr
		}

		codes := make([]string, 0, len(frame.Items))
		for _, item := range frame.Items {
			codes = append(codes, item.Code)
		}
		resolution, resolveErr := s.masters.ResolveItems(txCtx, clinicID, string(frame.SourceType), codes)
		if resolveErr != nil {
			return resolveErr
		}
		mappedField := make(map[string]uint64, len(resolution.Mapped))
		for _, mapped := range resolution.Mapped {
			mappedField[mapped.DeviceItemCode] = mapped.ExamTypeFieldID
		}
		unmapped := make(map[string]struct{}, len(resolution.UnmappedCodes))
		for _, code := range resolution.UnmappedCodes {
			unmapped[code] = struct{}{}
		}

		now := s.now()
		job := &model.LabImportJob{
			ClinicID:          clinicID,
			SourceType:        frame.SourceType,
			SourceFingerprint: frame.SourceFingerprint,
			Status:            model.LabImportJobStatusReceived,
			RowCount:          len(frame.Items),
			NeedsReviewCount:  len(resolution.UnmappedCodes),
			PetID:             petID,
			MeasuredAt:        frame.MeasuredAt,
			ReceivedAt:        &now,
			DeviceHint:        frame.DeviceHint,
			SpecimenIDRaw:     frame.SpecimenIDRaw,
			UnmappedItemCount: len(resolution.UnmappedCodes),
		}
		if frame.NeedsReview && job.NeedsReviewCount == 0 {
			job.NeedsReviewCount = 1
		}
		items := make([]model.LabImportJobItem, 0, len(frame.Items))
		for i, item := range frame.Items {
			row := model.LabImportJobItem{
				DeviceItemCode: item.Code,
				ValueRaw:       item.ValueRaw,
				Unit:           item.Unit,
				Flag:           item.Flag,
				NeedsReview:    frame.NeedsReview,
				SortOrder:      i,
			}
			if fieldID, ok := mappedField[item.Code]; ok {
				row.ExamTypeFieldID = &fieldID
			}
			if _, ok := unmapped[item.Code]; ok {
				row.NeedsReview = true
			}
			items = append(items, row)
		}
		if createErr := s.repo.CreateJobWithItems(txCtx, job, items); createErr != nil {
			if apperrors.IsConflict(createErr) {
				dup, dupErr := s.repo.FindJobByFingerprint(txCtx, clinicID, string(frame.SourceType), frame.SourceFingerprint)
				if dupErr != nil {
					return dupErr
				}
				if dup != nil {
					card, cardErr := s.cardForJob(txCtx, clinicID, dup)
					if cardErr != nil {
						return cardErr
					}
					result = LabDeviceFrameResult{Duplicate: true, Job: *card}
					return nil
				}
			}
			return createErr
		}
		if petID != nil {
			persistPetID = petID
			persistJobID = job.ID
		}
		card, cardErr := s.cardForJob(txCtx, clinicID, job)
		if cardErr != nil {
			return cardErr
		}
		result = LabDeviceFrameResult{Duplicate: false, Job: *card}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if persistPetID != nil && s.persister != nil {
		result.Job = *s.persistLinkedOrUnlink(ctx, clinicID, persistJobID, *persistPetID)
	}
	return &result, nil
}

func (s *labDeviceReceiveService) persistLinkedOrUnlink(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	petID uint64,
) *LabDeviceJobCard {
	if persistErr := s.persister.PersistLinkedJob(ctx, clinicID, jobID, petID); persistErr != nil {
		_ = s.withTx(ctx, func(txCtx context.Context) error {
			_, err := s.repo.UpdateJobPetID(txCtx, clinicID, jobID, nil)
			return err
		})
	}
	job, err := s.repo.FindJobByID(ctx, clinicID, jobID)
	if err != nil {
		return &LabDeviceJobCard{JobID: jobID}
	}
	card, err := s.cardForJob(ctx, clinicID, job)
	if err != nil {
		return &LabDeviceJobCard{JobID: jobID}
	}
	return card
}

func (s *labDeviceReceiveService) consumeWaitPet(ctx context.Context, clinicID uint64) (*uint64, error) {
	wait, err := s.repo.LockActiveWait(ctx, clinicID)
	if err != nil || wait == nil {
		return nil, err
	}
	now := s.now()
	if !wait.ExpiresAt.After(now) {
		return nil, s.repo.ClearWaitByID(ctx, clinicID, wait.ID, now)
	}
	return &wait.PetID, nil
}

func (s *labDeviceReceiveService) PutWait(
	ctx context.Context,
	clinicID, staffID, petID uint64,
) (*LabDeviceWaitView, error) {
	if staffID == 0 {
		return nil, apperrors.WrapInvalidInput("staff_id is required")
	}
	pet, err := s.requireAlivePet(ctx, clinicID, petID)
	if err != nil {
		return nil, err
	}
	var view LabDeviceWaitView
	err = s.withTx(ctx, func(txCtx context.Context) error {
		station, stationErr := s.ensureStation(txCtx, clinicID)
		if stationErr != nil {
			return stationErr
		}
		now := s.now()
		active, lockErr := s.repo.LockActiveWait(txCtx, clinicID)
		if lockErr != nil {
			return lockErr
		}
		if active != nil {
			if clearErr := s.repo.ClearWaitByID(txCtx, clinicID, active.ID, now); clearErr != nil {
				return clearErr
			}
		}
		wait := &model.LabDeviceWait{
			ClinicID:  clinicID,
			PetID:     pet.ID,
			StaffID:   staffID,
			ExpiresAt: now.Add(time.Duration(station.WaitTTLSeconds) * time.Second),
		}
		if insertErr := s.repo.InsertWait(txCtx, wait); insertErr != nil {
			return insertErr
		}
		view = LabDeviceWaitView{
			PetID:     pet.ID,
			PetName:   pet.Name,
			StaffID:   staffID,
			ExpiresAt: wait.ExpiresAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (s *labDeviceReceiveService) ClearWait(ctx context.Context, clinicID uint64) error {
	return s.withTx(ctx, func(txCtx context.Context) error {
		wait, err := s.repo.LockActiveWait(txCtx, clinicID)
		if err != nil || wait == nil {
			return err
		}
		return s.repo.ClearWaitByID(txCtx, clinicID, wait.ID, s.now())
	})
}

func (s *labDeviceReceiveService) Board(ctx context.Context, clinicID uint64) (*LabDeviceBoard, error) {
	station, err := s.ensureStation(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	wait, err := s.activeWaitView(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	unlinked, err := s.Unlinked(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	savedJobs, err := s.repo.ListSavedJobs(ctx, clinicID, labDeviceSavedLimit)
	if err != nil {
		return nil, err
	}
	saved, err := s.cardsForJobs(ctx, clinicID, savedJobs)
	if err != nil {
		return nil, err
	}
	since := s.now().In(config.JST).AddDate(0, 0, -labDeviceReceivedLookbackDays)
	receivedJobs, err := s.repo.ListReceivedJobs(ctx, clinicID, since, labDeviceReceivedLimit)
	if err != nil {
		return nil, err
	}
	received, err := s.cardsForJobs(ctx, clinicID, receivedJobs)
	if err != nil {
		return nil, err
	}
	todayVisits, err := s.todayVisits(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return &LabDeviceBoard{
		Wait:        wait,
		Unlinked:    unlinked,
		Saved:       saved,
		Received:    received,
		TodayVisits: todayVisits,
		Station:     *station,
	}, nil
}

func (s *labDeviceReceiveService) todayVisits(ctx context.Context, clinicID uint64) ([]LabDeviceTodayVisit, error) {
	if s.visits == nil {
		return []LabDeviceTodayVisit{}, nil
	}
	date := s.now().In(config.JST).Format(time.DateOnly)
	visits, err := s.visits.ListTodayDraft(ctx, clinicID, date)
	if err != nil {
		return nil, err
	}
	if visits == nil {
		return []LabDeviceTodayVisit{}, nil
	}
	return visits, nil
}

func (s *labDeviceReceiveService) Unlinked(ctx context.Context, clinicID uint64) ([]LabDeviceJobCard, error) {
	jobs, err := s.repo.ListUnlinkedJobs(ctx, clinicID, labDeviceUnlinkedLimit)
	if err != nil {
		return nil, err
	}
	return s.cardsForJobs(ctx, clinicID, jobs)
}

func (s *labDeviceReceiveService) Attach(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	petID uint64,
) (*LabDeviceJobCard, error) {
	pet, err := s.requireAlivePet(ctx, clinicID, petID)
	if err != nil {
		return nil, err
	}
	var card *LabDeviceJobCard
	err = s.withTx(ctx, func(txCtx context.Context) error {
		job, findErr := s.requireDeviceJob(txCtx, clinicID, jobID)
		if findErr != nil {
			return findErr
		}
		if job.PetID != nil && *job.PetID != pet.ID {
			return apperrors.WrapConflict("この検査受信は別の患者に紐付いています")
		}
		updated, updErr := s.repo.UpdateJobPetID(txCtx, clinicID, jobID, &pet.ID)
		if updErr != nil {
			return updErr
		}
		built, cardErr := s.cardForJob(txCtx, clinicID, updated)
		if cardErr != nil {
			return cardErr
		}
		card = built
		return nil
	})
	if err == nil && s.persister != nil {
		card = s.persistLinkedOrUnlink(ctx, clinicID, jobID, pet.ID)
		return card, nil
	}
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (s *labDeviceReceiveService) Detach(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
) (*LabDeviceJobCard, error) {
	var card *LabDeviceJobCard
	err := s.withTx(ctx, func(txCtx context.Context) error {
		job, findErr := s.requireDeviceJob(txCtx, clinicID, jobID)
		if findErr != nil {
			return findErr
		}
		if job.Status == model.LabImportJobStatusPersisted {
			if s.persister == nil {
				return apperrors.WrapConflict("保存済みの検査は取り消せません。検査記録の訂正が必要です")
			}
			if retractErr := s.persister.RetractLinkedJob(txCtx, clinicID, jobID); retractErr != nil {
				return retractErr
			}
		}
		if job.Status == model.LabImportJobStatusReverted {
			return apperrors.WrapInvalidInput("reverted job cannot detach")
		}
		updated, updErr := s.repo.UpdateJobPetID(txCtx, clinicID, jobID, nil)
		if updErr != nil {
			return updErr
		}
		built, cardErr := s.cardForJob(txCtx, clinicID, updated)
		if cardErr != nil {
			return cardErr
		}
		card = built
		return nil
	})
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (s *labDeviceReceiveService) GetStation(ctx context.Context, clinicID uint64) (*LabDeviceStationView, error) {
	return s.ensureStation(ctx, clinicID)
}

func (s *labDeviceReceiveService) PutStation(
	ctx context.Context,
	clinicID uint64,
	slotsJSON string,
) (*LabDeviceStationView, error) {
	if err := validLabDeviceSlotsJSON(slotsJSON); err != nil {
		return nil, err
	}
	current, err := s.ensureStation(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	row := &model.LabDeviceStationSettings{
		ClinicID:       clinicID,
		WaitTTLSeconds: current.WaitTTLSeconds,
		SlotsJSON:      slotsJSON,
	}
	if err := s.repo.UpsertStation(ctx, row); err != nil {
		return nil, err
	}
	return &LabDeviceStationView{WaitTTLSeconds: row.WaitTTLSeconds, SlotsJSON: row.SlotsJSON}, nil
}

func (s *labDeviceReceiveService) ensureStation(ctx context.Context, clinicID uint64) (*LabDeviceStationView, error) {
	row, err := s.repo.GetStation(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		row = &model.LabDeviceStationSettings{
			ClinicID:       clinicID,
			WaitTTLSeconds: labDeviceDefaultWaitTTLSeconds,
			SlotsJSON:      labDeviceDefaultSlotsJSON,
		}
		if err := s.repo.UpsertStation(ctx, row); err != nil {
			return nil, err
		}
	}
	ttl := row.WaitTTLSeconds
	if ttl < labDeviceMinWaitTTLSeconds || ttl > labDeviceMaxWaitTTLSeconds {
		ttl = labDeviceDefaultWaitTTLSeconds
	}
	slots := row.SlotsJSON
	if slots == "" {
		slots = labDeviceDefaultSlotsJSON
	}
	return &LabDeviceStationView{WaitTTLSeconds: ttl, SlotsJSON: slots}, nil
}

func (s *labDeviceReceiveService) activeWaitView(ctx context.Context, clinicID uint64) (*LabDeviceWaitView, error) {
	wait, err := s.repo.FindActiveWait(ctx, clinicID)
	if err != nil || wait == nil {
		return nil, err
	}
	if !wait.ExpiresAt.After(s.now()) {
		if clearErr := s.repo.ClearWaitByID(ctx, clinicID, wait.ID, s.now()); clearErr != nil {
			return nil, clearErr
		}
		return nil, nil
	}
	name := ""
	if s.pets != nil {
		if pet, petErr := s.pets.FindByID(ctx, clinicID, wait.PetID); petErr == nil && pet != nil {
			name = pet.Name
		}
	}
	return &LabDeviceWaitView{
		PetID:     wait.PetID,
		PetName:   name,
		StaffID:   wait.StaffID,
		ExpiresAt: wait.ExpiresAt,
	}, nil
}

func (s *labDeviceReceiveService) requireAlivePet(ctx context.Context, clinicID, petID uint64) (*model.Pet, error) {
	if petID == 0 || s.pets == nil {
		return nil, apperrors.WrapInvalidInput("pet_id is required")
	}
	pet, err := s.pets.FindByID(ctx, clinicID, petID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.WrapInvalidInput("pet_id is not in this clinic")
		}
		return nil, err
	}
	if pet == nil || pet.ClinicID != clinicID {
		return nil, apperrors.WrapInvalidInput("pet_id is not in this clinic")
	}
	if pet.DeletedAt.Valid || pet.Status != model.PetStatusAlive {
		return nil, apperrors.WrapInvalidInput("pet is not alive")
	}
	return pet, nil
}

func (s *labDeviceReceiveService) requireDeviceJob(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
) (*model.LabImportJob, error) {
	job, err := s.repo.FindJobByID(ctx, clinicID, jobID)
	if err != nil {
		return nil, err
	}
	if !isLabDeviceSourceType(string(job.SourceType)) {
		return nil, apperrors.WrapInvalidInput("job is not a lab device source")
	}
	return job, nil
}

func (s *labDeviceReceiveService) cardsForJobs(
	ctx context.Context,
	clinicID uint64,
	jobs []model.LabImportJob,
) ([]LabDeviceJobCard, error) {
	cards := make([]LabDeviceJobCard, 0, len(jobs))
	ids := make([]uuid.UUID, 0, len(jobs))
	for i := range jobs {
		ids = append(ids, jobs[i].ID)
	}
	items, err := s.repo.ListJobItems(ctx, clinicID, ids)
	if err != nil {
		return nil, err
	}
	for i := range jobs {
		card := jobToCard(&jobs[i], items[jobs[i].ID])
		if jobs[i].PetID != nil && s.pets != nil {
			if pet, petErr := s.pets.FindByID(ctx, clinicID, *jobs[i].PetID); petErr == nil && pet != nil {
				card.PetName = pet.Name
			}
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (s *labDeviceReceiveService) cardForJob(
	ctx context.Context,
	clinicID uint64,
	job *model.LabImportJob,
) (*LabDeviceJobCard, error) {
	cards, err := s.cardsForJobs(ctx, clinicID, []model.LabImportJob{*job})
	if err != nil {
		return nil, err
	}
	return &cards[0], nil
}

func jobToCard(job *model.LabImportJob, items []model.LabImportJobItem) LabDeviceJobCard {
	views := make([]LabDeviceJobItemView, 0, len(items))
	for i := range items {
		views = append(views, LabDeviceJobItemView{
			DeviceItemCode:  items[i].DeviceItemCode,
			ValueRaw:        items[i].ValueRaw,
			Unit:            items[i].Unit,
			Flag:            items[i].Flag,
			ExamTypeFieldID: items[i].ExamTypeFieldID,
			NeedsReview:     items[i].NeedsReview,
			SortOrder:       items[i].SortOrder,
		})
	}
	return LabDeviceJobCard{
		JobID:             job.ID,
		SourceType:        job.SourceType,
		DeviceHint:        job.DeviceHint,
		Status:            job.Status,
		PetID:             job.PetID,
		MeasuredAt:        job.MeasuredAt,
		ReceivedAt:        job.ReceivedAt,
		SpecimenIDRaw:     job.SpecimenIDRaw,
		ItemCount:         job.RowCount,
		UnmappedItemCount: job.UnmappedItemCount,
		ClockSkew:         labDeviceHasClockSkew(job.MeasuredAt, job.ReceivedAt),
		Items:             views,
	}
}

func labDeviceHasClockSkew(measured, received *time.Time) bool {
	if measured == nil || received == nil {
		return false
	}
	delta := received.Sub(*measured)
	if delta < 0 {
		delta = -delta
	}
	return delta > 24*time.Hour
}
